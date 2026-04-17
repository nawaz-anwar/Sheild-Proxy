package proxy

import (
"net/http"
"net/http/httptest"
"sync/atomic"
"testing"

"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/config"
)

func TestProxyInjectsHeadersAndForwards(t *testing.T) {
origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
if r.Header.Get("X-Shield-Verified") != "true" {
t.Fatalf("missing verified header")
}
if r.Header.Get("X-Shield-Client-ID") != "client-a" {
t.Fatalf("missing client header")
}
w.WriteHeader(http.StatusOK)
}))
defer origin.Close()

cfg := &config.Config{}
cfg.Proxy.VerifiedHeader = "X-Shield-Verified"
cfg.Proxy.ClientIDHeader = "X-Shield-Client-ID"
cfg.Proxy.Domains = append(cfg.Proxy.Domains, struct {
Host     string "yaml:\"host\""
Upstream string "yaml:\"upstream\""
ClientID string "yaml:\"client_id\""
Active   bool   "yaml:\"active\""
}{
Host:     "example.com",
Upstream: origin.URL,
ClientID: "client-a",
Active:   true,
})

var count uint64
p := New(cfg, &count)

req := httptest.NewRequest(http.MethodGet, "http://example.com/test", nil)
req.Host = "example.com"
rr := httptest.NewRecorder()
p.ServeHTTP(rr, req)

if rr.Code != http.StatusOK {
t.Fatalf("expected 200 got %d", rr.Code)
}
if atomic.LoadUint64(&count) != 1 {
t.Fatalf("expected count=1")
}
}
