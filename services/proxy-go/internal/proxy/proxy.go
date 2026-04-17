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
}

type requestLog struct {
	Time     time.Time `json:"time"`
	Host     string    `json:"host"`
	Method   string    `json:"method"`
	Path     string    `json:"path"`
	RemoteIP string    `json:"remoteIp"`
}

func New(cfg *config.Config, requestCount *uint64) *Proxy {
	store := domainstore.New(cfg.Proxy.Domains, nil, nil)
	return &Proxy{
		verifiedHeader: cfg.Proxy.VerifiedHeader,
		clientIDHeader: cfg.Proxy.ClientIDHeader,
		store:          store,
		requestCount:   requestCount,
		filter:         filter.New(cfg),
		jwt:            jwt.New(cfg.JWT.Issuer, cfg.JWT.Audience, cfg.JWT.Secret, cfg.JWT.CookieName, cfg.JWT.TTLSeconds),
		signer:         signing.New(cfg.Proxy.Signing.Enabled, cfg.Proxy.Signing.Secret, cfg.Proxy.Signing.Header, cfg.Proxy.Signing.TimestampHeader),
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		status := decision.Status
		if status == 0 {
			status = http.StatusForbidden
		}
		http.Error(w, decision.Body, status)
		return
	}
	if decision.Cookie != nil {
		http.SetCookie(w, decision.Cookie)
	}

	atomic.AddUint64(p.requestCount, 1)
	p.logRequest(r, host)

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
	proxy.ServeHTTP(w, r)
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

func (p *Proxy) logRequest(r *http.Request, host string) {
	l := requestLog{
		Time:     time.Now().UTC(),
		Host:     host,
		Method:   r.Method,
		Path:     r.URL.Path,
		RemoteIP: r.RemoteAddr,
	}
	b, _ := json.Marshal(l)
	log.Printf("request=%s", b)
}
