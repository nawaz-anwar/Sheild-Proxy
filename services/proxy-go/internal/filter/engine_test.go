package filter

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/config"
)

func TestBanRuleBlocksBeforeOtherChecks(t *testing.T) {
	cfg := &config.Config{}
	cfg.Proxy.Filter.BannedIPs = []string{"203.0.113.7"}
	cfg.Proxy.RateLimit.Enabled = true
	cfg.Proxy.RateLimit.MaxRequests = 1
	cfg.Proxy.RateLimit.WindowSeconds = 60

	e := New(cfg)
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.RemoteAddr = "203.0.113.7:10"

	decision := e.Evaluate(req, "example.com")
	if decision.Allowed {
		t.Fatalf("expected blocked request")
	}
	if decision.Status != http.StatusForbidden {
		t.Fatalf("expected 403 got %d", decision.Status)
	}
}

func TestAllowlistBypassesRateLimit(t *testing.T) {
	cfg := &config.Config{}
	cfg.Proxy.Filter.AllowlistIPs = []string{"203.0.113.8"}
	cfg.Proxy.RateLimit.Enabled = true
	cfg.Proxy.RateLimit.MaxRequests = 1
	cfg.Proxy.RateLimit.WindowSeconds = 3600

	e := New(cfg)
	e.now = func() time.Time { return time.Unix(1700000000, 0) }
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
		req.RemoteAddr = "203.0.113.8:10"
		decision := e.Evaluate(req, "example.com")
		if !decision.Allowed {
			t.Fatalf("allowlisted request should pass")
		}
	}
}

func TestHeaderBotChallengeIsIssued(t *testing.T) {
	cfg := &config.Config{}
	cfg.Proxy.HeaderAnalysis.Enabled = true
	cfg.Proxy.HeaderAnalysis.BlockedBotKeywords = []string{"curl"}
	cfg.Proxy.Challenge.Enabled = true
	cfg.JWT.Secret = "abc123"

	e := New(cfg)
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	req.RemoteAddr = "203.0.113.9:10"
	req.Header.Set("User-Agent", "curl/8.0")

	decision := e.Evaluate(req, "example.com")
	if decision.Allowed {
		t.Fatalf("expected challenge")
	}
	if decision.Status != http.StatusForbidden {
		t.Fatalf("expected 403 got %d", decision.Status)
	}
	if decision.Headers["X-Shield-Challenge-ID"] == "" {
		t.Fatalf("expected challenge id header")
	}
}
