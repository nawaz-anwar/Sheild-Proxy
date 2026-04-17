package filter

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/challenge"
	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/config"
	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/geoip"
	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/headers"
	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/jwt"
	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/ratelimiter"
)

type Decision struct {
	Allowed bool
	Status  int
	Body    string
	Cookie  *http.Cookie
	Headers map[string]string
}

type Engine struct {
	allowlist map[string]struct{}
	banned    map[string]struct{}
	limiter   *ratelimiter.Limiter
	geo       *geoip.Filter
	headers   *headers.Analyzer
	challenge *challenge.Manager
	jwt       *jwt.Service
	now       func() time.Time
}

func New(cfg *config.Config) *Engine {
	allowlist := map[string]struct{}{}
	for _, ip := range cfg.Proxy.Filter.AllowlistIPs {
		if ip = strings.TrimSpace(ip); ip != "" {
			allowlist[ip] = struct{}{}
		}
	}
	banned := map[string]struct{}{}
	for _, ip := range cfg.Proxy.Filter.BannedIPs {
		if ip = strings.TrimSpace(ip); ip != "" {
			banned[ip] = struct{}{}
		}
	}
	lookup := geoip.NewStaticLookup(cfg.GeoIP.MockData)
	return &Engine{
		allowlist: allowlist,
		banned:    banned,
		limiter:   ratelimiter.New(cfg.Proxy.RateLimit.Enabled, cfg.Proxy.RateLimit.WindowSeconds, cfg.Proxy.RateLimit.MaxRequests),
		geo:       geoip.NewFilter(cfg.GeoIP, lookup),
		headers:   headers.NewAnalyzer(cfg.Proxy.HeaderAnalysis.Enabled, cfg.Proxy.HeaderAnalysis.BlockedBotKeywords, cfg.Proxy.HeaderAnalysis.SEOBotAllowlist),
		challenge: challenge.New(cfg.Proxy.Challenge.Enabled, cfg.Proxy.Challenge.TTLSeconds, cfg.Proxy.Challenge.Difficulty, cfg.Proxy.Challenge.CookieName),
		jwt:       jwt.New(cfg.JWT.Issuer, cfg.JWT.Audience, cfg.JWT.Secret, cfg.JWT.CookieName, cfg.JWT.TTLSeconds),
		now:       time.Now,
	}
}

func (e *Engine) Evaluate(r *http.Request, domain string) Decision {
	ip, ok := clientIP(r)
	if !ok {
		return Decision{Allowed: false, Status: http.StatusForbidden, Body: "unable to determine client ip"}
	}
	if _, ok := e.allowlist[ip]; ok {
		return Decision{Allowed: true}
	}
	if _, ok := e.banned[ip]; ok {
		return Decision{Allowed: false, Status: http.StatusForbidden, Body: "blocked by ban list"}
	}
	if ok, _ := e.limiter.Allow(domain, ip, e.now().UTC()); !ok {
		return Decision{Allowed: false, Status: http.StatusTooManyRequests, Body: "rate limit exceeded"}
	}
	if !e.geo.Allow(net.ParseIP(ip)) {
		return Decision{Allowed: false, Status: http.StatusForbidden, Body: "geo policy blocked request"}
	}

	ua := r.UserAgent()
	hd := e.headers.Analyze(ua)
	if hd.Allow {
		return Decision{Allowed: true}
	}
	if !hd.RequiresChallenge || !e.challenge.Enabled() {
		return Decision{Allowed: false, Status: http.StatusForbidden, Body: "blocked by header analysis"}
	}
	if c, err := r.Cookie(e.jwt.CookieName()); err == nil && e.jwt.Validate(c.Value, ip, ua, e.now().UTC()) {
		return Decision{Allowed: true}
	}
	challengeID := r.Header.Get("X-Shield-Challenge-ID")
	challengeAnswer := r.Header.Get("X-Shield-Challenge-Answer")
	if id, answer, err := jwt.ParseChallengeHeaders(challengeID, challengeAnswer); err == nil {
		if e.challenge.Verify(id, answer, domain, ip, ua, e.now().UTC()) {
			token, issueErr := e.jwt.Issue(domain, ip, ua, e.now().UTC())
			if issueErr == nil {
				return Decision{
					Allowed: true,
					Cookie: &http.Cookie{
						Name:     e.jwt.CookieName(),
						Value:    token,
						Path:     "/",
						HttpOnly: true,
						Secure:   true,
						SameSite: http.SameSiteLaxMode,
						MaxAge:   int(e.jwt.TTL().Seconds()),
					},
				}
			}
		}
	}
	ch, err := e.challenge.Create(domain, ip, ua, e.now().UTC())
	if err != nil {
		return Decision{Allowed: false, Status: http.StatusServiceUnavailable, Body: "challenge generation unavailable"}
	}
	return Decision{
		Allowed: false,
		Status:  http.StatusForbidden,
		Body:    "challenge required",
		Headers: map[string]string{
			"X-Shield-Challenge-ID":         ch.ID,
			"X-Shield-Challenge-Prefix":     ch.Prefix,
			"X-Shield-Challenge-Difficulty": intToString(ch.Difficulty),
			"X-Shield-Challenge-Expires-At": ch.ExpiresAt.UTC().Format(time.RFC3339),
		},
	}
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

func intToString(v int) string {
	if v == 0 {
		return "0"
	}
	buf := [20]byte{}
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + (v % 10))
		v /= 10
	}
	return string(buf[i:])
}
