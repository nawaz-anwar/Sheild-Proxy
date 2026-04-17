package proxy

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/config"
	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/jwt"
)

func TestProxyInjectsHeadersAndForwards(t *testing.T) {
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Shield-Verified") != "true" {
			t.Fatalf("missing verified header")
		}
		if r.Header.Get("X-Shield-Client-ID") != "client-a" {
			t.Fatalf("missing client header")
		}
		if r.Header.Get("X-Shield-Signature") == "" {
			t.Fatalf("missing request signature header")
		}
		if r.Header.Get("X-Shield-Signature-Timestamp") == "" {
			t.Fatalf("missing request signature timestamp header")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer origin.Close()

	cfg := &config.Config{}
	cfg.Proxy.VerifiedHeader = "X-Shield-Verified"
	cfg.Proxy.ClientIDHeader = "X-Shield-Client-ID"
	cfg.Proxy.Signing.Enabled = true
	cfg.Proxy.Signing.Secret = "signing-test-secret"
	cfg.Proxy.Signing.Header = "X-Shield-Signature"
	cfg.Proxy.Signing.TimestampHeader = "X-Shield-Signature-Timestamp"
	cfg.Proxy.Domains = append(cfg.Proxy.Domains, config.DomainConfig{
		Host:     "example.com",
		Upstream: origin.URL,
		ClientID: "client-a",
		Active:   true,
	})

	var count uint64
	p := New(cfg, nil, &count)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	req.Host = "example.com"
	req.RemoteAddr = "203.0.113.10:12345"
	rr := httptest.NewRecorder()
	p.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
	if atomic.LoadUint64(&count) != 1 {
		t.Fatalf("expected count=1")
	}
}

func TestProxyRejectsInvalidTokenBeforeFilter(t *testing.T) {
	cfg := &config.Config{}
	cfg.Proxy.VerifiedHeader = "X-Shield-Verified"
	cfg.Proxy.ClientIDHeader = "X-Shield-Client-ID"
	cfg.Proxy.Domains = append(cfg.Proxy.Domains, config.DomainConfig{
		Host:     "example.com",
		Upstream: "http://127.0.0.1:9999",
		ClientID: "client-a",
		Active:   true,
	})
	cfg.JWT.Secret = "jwt-test-secret"
	cfg.JWT.Issuer = "shield-proxy"
	cfg.JWT.Audience = "shield-origin"
	cfg.JWT.CookieName = "shield_token"
	cfg.JWT.TTLSeconds = 900

	var count uint64
	p := New(cfg, nil, &count)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	req.Host = "example.com"
	req.RemoteAddr = "203.0.113.10:12345"
	req.Header.Set("User-Agent", "unit-test-agent")
	req.Header.Set("Authorization", "Bearer not-a-valid-token")
	rr := httptest.NewRecorder()
	p.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d", rr.Code)
	}
}

func TestProxyAcceptsValidToken(t *testing.T) {
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer origin.Close()

	cfg := &config.Config{}
	cfg.Proxy.VerifiedHeader = "X-Shield-Verified"
	cfg.Proxy.ClientIDHeader = "X-Shield-Client-ID"
	cfg.Proxy.Domains = append(cfg.Proxy.Domains, config.DomainConfig{
		Host:     "example.com",
		Upstream: origin.URL,
		ClientID: "client-a",
		Active:   true,
	})
	cfg.JWT.Secret = "jwt-test-secret"
	cfg.JWT.Issuer = "shield-proxy"
	cfg.JWT.Audience = "shield-origin"
	cfg.JWT.CookieName = "shield_token"
	cfg.JWT.TTLSeconds = 900

	issuer := jwt.New(cfg.JWT.Issuer, cfg.JWT.Audience, cfg.JWT.Secret, cfg.JWT.CookieName, cfg.JWT.TTLSeconds)
	token, err := issuer.Issue("example.com", "203.0.113.10", "unit-test-agent", time.Now().UTC())
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	var count uint64
	p := New(cfg, nil, &count)
	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	req.Host = "example.com"
	req.RemoteAddr = "203.0.113.10:12345"
	req.Header.Set("User-Agent", "unit-test-agent")
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	p.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr.Code, strings.TrimSpace(rr.Body.String()))
	}
}

func TestProxyChallengeVerifyEndpointSetsTokenCookie(t *testing.T) {
	cfg := &config.Config{}
	cfg.Proxy.Domains = append(cfg.Proxy.Domains, config.DomainConfig{
		Host:     "example.com",
		Upstream: "http://127.0.0.1:9999",
		ClientID: "client-a",
		Active:   true,
	})
	cfg.Proxy.HeaderAnalysis.Enabled = true
	cfg.Proxy.HeaderAnalysis.BlockedBotKeywords = []string{"curl"}
	cfg.Proxy.Challenge.Enabled = true
	cfg.Proxy.Challenge.Difficulty = 1
	cfg.JWT.Secret = "jwt-test-secret"
	cfg.JWT.Issuer = "shield-proxy"
	cfg.JWT.Audience = "shield-origin"
	cfg.JWT.CookieName = "shield_token"
	cfg.JWT.TTLSeconds = 900

	var count uint64
	p := New(cfg, nil, &count)
	req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
	req.Host = "example.com"
	req.RemoteAddr = "203.0.113.10:12345"
	req.Header.Set("User-Agent", "curl/8.0")
	rr := httptest.NewRecorder()
	p.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected challenge 403 got %d", rr.Code)
	}
	challengeID := rr.Header().Get("X-Shield-Challenge-ID")
	prefix := rr.Header().Get("X-Shield-Challenge-Prefix")
	if challengeID == "" || prefix == "" {
		t.Fatalf("missing challenge headers")
	}
	answer := solveChallenge(prefix, 1)

	body, _ := json.Marshal(map[string]string{"challengeId": challengeID, "answer": answer})
	verifyReq := httptest.NewRequest(http.MethodPost, "http://example.com/sp-verify", strings.NewReader(string(body)))
	verifyReq.Host = "example.com"
	verifyReq.RemoteAddr = "203.0.113.10:22222"
	verifyReq.Header.Set("User-Agent", "curl/8.0")
	verifyReq.Header.Set("Content-Type", "application/json")
	verifyRR := httptest.NewRecorder()
	p.ServeHTTP(verifyRR, verifyReq)
	if verifyRR.Code != http.StatusOK {
		t.Fatalf("expected 200 on verify got %d body=%s", verifyRR.Code, strings.TrimSpace(verifyRR.Body.String()))
	}
	if !strings.Contains(verifyRR.Header().Get("Set-Cookie"), cfg.JWT.CookieName+"=") {
		t.Fatalf("expected token cookie in response")
	}
}

func solveChallenge(prefix string, difficulty int) string {
	target := strings.Repeat("0", difficulty)
	for i := 0; i < 500000; i++ {
		answer := fmtInt(i)
		sum := sha256.Sum256([]byte(prefix + answer))
		if strings.HasPrefix(hex.EncodeToString(sum[:]), target) {
			return answer
		}
	}
	return "0"
}

func fmtInt(v int) string {
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
