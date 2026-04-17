package proxy

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/config"
	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/domainstore"
	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/filter"
	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/jwt"
	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/runtime"
	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/signing"
)

type hostConfig struct {
	upstream *url.URL
	clientID string
}

type Proxy struct {
	verifiedHeader string
	clientIDHeader string
	store          *domainstore.Store
	requestCount   *uint64
	filter         *filter.Engine
	jwt            *jwt.Service
	signer         *signing.Signer
	logWriter      *requestLogger
	l2             *redisDomainProvider
	l3             *postgresDomainProvider
	subCancel      context.CancelFunc
}

type requestLog struct {
	Time     time.Time `json:"time"`
	Host     string    `json:"host"`
	Method   string    `json:"method"`
	Path     string    `json:"path"`
	RemoteIP string    `json:"remoteIp"`
}

func New(cfg *config.Config, deps *runtime.Dependencies, requestCount *uint64) *Proxy {
	if deps == nil {
		deps = &runtime.Dependencies{}
	}
	l2 := newRedisDomainProvider(deps.Redis)
	l3 := newPostgresDomainProvider(deps.Postgres)
	store := domainstore.New(cfg.Proxy.Domains, l2, l3)
	p := &Proxy{
		verifiedHeader: cfg.Proxy.VerifiedHeader,
		clientIDHeader: cfg.Proxy.ClientIDHeader,
		store:          store,
		requestCount:   requestCount,
		filter:         filter.New(cfg),
		jwt:            jwt.New(cfg.JWT.Issuer, cfg.JWT.Audience, cfg.JWT.Secret, cfg.JWT.CookieName, cfg.JWT.TTLSeconds),
		signer:         signing.New(cfg.Proxy.Signing.Enabled, cfg.Proxy.Signing.Secret, cfg.Proxy.Signing.Header, cfg.Proxy.Signing.TimestampHeader),
		logWriter:      newRequestLogger(deps.Postgres),
		l2:             l2,
		l3:             l3,
	}
	p.warmFromPostgres()
	p.startRedisInvalidationSubscriber(deps.Redis)
	return p
}

func (p *Proxy) Close() {
	if p.subCancel != nil {
		p.subCancel()
	}
	if p.logWriter != nil {
		p.logWriter.Close()
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/sp-verify" {
		p.handleVerify(w, r)
		return
	}
	host := strings.ToLower(r.Host)
	if i := strings.IndexByte(host, ':'); i != -1 {
		host = host[:i]
	}
	domain, ok := p.store.Get(context.Background(), host)
	if !ok || !domain.Active {
		http.Error(w, "domain is not active", http.StatusNotFound)
		return
	}
	hc, ok := toHostConfig(domain)
	if !ok {
		http.Error(w, "invalid upstream", http.StatusBadGateway)
		return
	}
	if !p.validateToken(r) {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	decision := p.filter.Evaluate(r, host)
	if !decision.Allowed {
		for k, v := range decision.Headers {
			w.Header().Set(k, v)
		}
		if p.shouldRenderChallenge(decision) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(renderChallengePage(
				decision.Headers["X-Shield-Challenge-ID"],
				decision.Headers["X-Shield-Challenge-Prefix"],
				decision.Headers["X-Shield-Challenge-Difficulty"],
			)))
			p.logRequest(r, host, http.StatusForbidden)
			return
		}
		status := decision.Status
		if status == 0 {
			status = http.StatusForbidden
		}
		http.Error(w, decision.Body, status)
		p.logRequest(r, host, status)
		return
	}
	if decision.Cookie != nil {
		http.SetCookie(w, decision.Cookie)
	}

	atomic.AddUint64(p.requestCount, 1)
	proxy := httputil.NewSingleHostReverseProxy(hc.upstream)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Header.Set(p.verifiedHeader, "true")
		req.Header.Set(p.clientIDHeader, hc.clientID)
		if header, tsHeader, sig, ts, ok := p.signer.Sign(req.Method, req.URL.RequestURI(), host, hc.clientID, time.Now()); ok {
			req.Header.Set(header, sig)
			req.Header.Set(tsHeader, ts)
		}
	}
	rec := &statusCaptureWriter{ResponseWriter: w}
	proxy.ServeHTTP(rec, r)
	status := rec.status
	if status == 0 {
		status = http.StatusOK
	}
	p.logRequest(r, host, status)
}

func (p *Proxy) validateToken(r *http.Request) bool {
	token := ""
	if c, err := r.Cookie(p.jwt.CookieName()); err == nil {
		token = strings.TrimSpace(c.Value)
	}
	if token == "" {
		if bearer, err := jwt.BearerToken(r.Header.Get("Authorization")); err == nil {
			token = strings.TrimSpace(bearer)
		}
	}
	if token == "" {
		return true
	}
	ip, ok := clientIP(r)
	if !ok {
		return false
	}
	return p.jwt.Validate(token, ip, r.UserAgent(), time.Now().UTC())
}

func clientIP(r *http.Request) (string, bool) {
	if xfwd := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xfwd != "" {
		parts := strings.Split(xfwd, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if net.ParseIP(ip) != nil {
				return ip, true
			}
		}
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && net.ParseIP(host) != nil {
		return host, true
	}
	if net.ParseIP(strings.TrimSpace(r.RemoteAddr)) != nil {
		return strings.TrimSpace(r.RemoteAddr), true
	}
	return "", false
}

func toHostConfig(d domainstore.Domain) (hostConfig, bool) {
	u, err := url.Parse(d.Upstream)
	if err != nil {
		log.Printf("invalid upstream for host %s: %v", d.Host, err)
		return hostConfig{}, false
	}
	return hostConfig{upstream: u, clientID: d.ClientID}, true
}

func (p *Proxy) logRequest(r *http.Request, host string, statusCode int) {
	l := requestLog{
		Time:     time.Now().UTC(),
		Host:     host,
		Method:   r.Method,
		Path:     r.URL.Path,
		RemoteIP: r.RemoteAddr,
	}
	b, _ := json.Marshal(l)
	log.Printf("request=%s", b)
	if p.logWriter != nil {
		p.logWriter.Log(accessLogEntry{
			Host:      host,
			Method:    r.Method,
			Path:      r.URL.Path,
			Status:    statusCode,
			RemoteIP:  r.RemoteAddr,
			UserAgent: r.UserAgent(),
			CreatedAt: time.Now().UTC(),
		})
	}
}

func (p *Proxy) shouldRenderChallenge(decision filter.Decision) bool {
	return decision.Body == "challenge required" &&
		decision.Headers["X-Shield-Challenge-ID"] != "" &&
		decision.Headers["X-Shield-Challenge-Prefix"] != "" &&
		decision.Headers["X-Shield-Challenge-Difficulty"] != ""
}

func (p *Proxy) handleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	host := strings.ToLower(r.Host)
	if i := strings.IndexByte(host, ':'); i != -1 {
		host = host[:i]
	}
	var req struct {
		ChallengeID string `json:"challengeId"`
		Answer      string `json:"answer"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.ChallengeID) == "" || strings.TrimSpace(req.Answer) == "" {
		http.Error(w, "missing challenge fields", http.StatusBadRequest)
		return
	}
	verifyReq := r.Clone(r.Context())
	verifyReq.Header.Set("X-Shield-Challenge-ID", req.ChallengeID)
	verifyReq.Header.Set("X-Shield-Challenge-Answer", req.Answer)
	decision := p.filter.Evaluate(verifyReq, host)
	if !decision.Allowed || decision.Cookie == nil {
		http.Error(w, "challenge verification failed", http.StatusForbidden)
		p.logRequest(r, host, http.StatusForbidden)
		return
	}
	http.SetCookie(w, decision.Cookie)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"ok":true}`))
	p.logRequest(r, host, http.StatusOK)
}

func renderChallengePage(challengeID, prefix, difficulty string) string {
	challengeID = escapeJSString(challengeID)
	prefix = escapeJSString(prefix)
	difficulty = escapeJSString(difficulty)
	return `<!doctype html>
<html lang="en"><head><meta charset="utf-8"/><meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>Security verification</title><style>body{font-family:system-ui,-apple-system,sans-serif;margin:2rem;max-width:640px}code{background:#f4f4f5;padding:0.1rem 0.25rem;border-radius:4px}</style></head>
<body><h1>Security verification required</h1><p>Please wait while we verify your browser.</p><p id="status">Solving challenge…</p>
<script>
const challengeId="` + challengeID + `";
const challengePrefix="` + prefix + `";
const challengeDifficulty="` + difficulty + `";
async function sha256hex(input){const data=new TextEncoder().encode(input);const hash=await crypto.subtle.digest('SHA-256',data);return Array.from(new Uint8Array(hash)).map(b=>b.toString(16).padStart(2,'0')).join('');}
async function solve(prefix,difficulty){const target='0'.repeat(Number(difficulty));let n=0;while(true){const ans=String(n++);const digest=await sha256hex(prefix+ans);if(digest.startsWith(target)) return ans;if(n%500===0) await new Promise(r=>setTimeout(r,0));}}
async function run(){
const status=document.getElementById('status');
try{
const answer=await solve(challengePrefix,challengeDifficulty);
const resp=await fetch('/sp-verify',{method:'POST',headers:{'Content-Type':'application/json'},credentials:'include',body:JSON.stringify({challengeId,answer})});
if(!resp.ok){throw new Error('verification failed')}
status.textContent='Verification successful. Reloading...';
window.location.reload();
}catch(err){
status.textContent='Verification failed. Please retry.';
}
}
run();
</script>
</body></html>`
}

func escapeJSString(in string) string {
	in = strings.ReplaceAll(in, "\\", "\\\\")
	in = strings.ReplaceAll(in, "\"", "\\\"")
	in = strings.ReplaceAll(in, "'", "\\'")
	in = strings.ReplaceAll(in, "`", "\\u0060")
	in = strings.ReplaceAll(in, "<", "\\u003c")
	in = strings.ReplaceAll(in, ">", "\\u003e")
	in = strings.ReplaceAll(in, "&", "\\u0026")
	in = strings.ReplaceAll(in, "\n", "")
	in = strings.ReplaceAll(in, "\r", "")
	return in
}

func (p *Proxy) warmFromPostgres() {
	if p.l3 == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rows, err := p.l3.LoadActive(ctx)
	if err != nil {
		log.Printf("domain warm cache load failed: %v", err)
		return
	}
	for _, d := range rows {
		p.store.Set(d)
		if p.l2 != nil {
			p.l2.Set(ctx, d)
		}
	}
	log.Printf("domain warm cache loaded %d records", len(rows))
}

func (p *Proxy) startRedisInvalidationSubscriber(redisClient *runtime.RedisClient) {
	if redisClient == nil || redisClient.Client == nil {
		return
	}
	channel := strings.TrimSpace(redisClient.Prefix)
	if channel == "" {
		channel = "shield"
	}
	channel = channel + ":domain:sync"
	subCtx, cancel := context.WithCancel(context.Background())
	p.subCancel = cancel
	go func() {
		pubsub := redisClient.Client.Subscribe(subCtx, channel)
		defer pubsub.Close()
		ch := pubsub.Channel()
		for {
			select {
			case <-subCtx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				host := strings.ToLower(strings.TrimSpace(msg.Payload))
				if host == "" {
					continue
				}
				p.store.Delete(host)
				if p.l2 != nil {
					p.l2.Delete(context.Background(), host)
				}
			}
		}
	}()
}
