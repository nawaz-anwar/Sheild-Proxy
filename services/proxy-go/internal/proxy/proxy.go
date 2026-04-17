package proxy

import (
"encoding/json"
"log"
"net/http"
"net/http/httputil"
"net/url"
"strings"
"sync/atomic"
"time"

"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/config"
)

type hostConfig struct {
upstream *url.URL
clientID string
}

type Proxy struct {
verifiedHeader string
clientIDHeader string
hosts          map[string]hostConfig
requestCount   *uint64
}

type requestLog struct {
Time     time.Time `json:"time"`
Host     string    `json:"host"`
Method   string    `json:"method"`
Path     string    `json:"path"`
RemoteIP string    `json:"remoteIp"`
}

func New(cfg *config.Config, requestCount *uint64) *Proxy {
hosts := make(map[string]hostConfig, len(cfg.Proxy.Domains))
for _, d := range cfg.Proxy.Domains {
if !d.Active {
continue
}
u, err := url.Parse(d.Upstream)
if err != nil {
log.Printf("invalid upstream for host %s: %v", d.Host, err)
continue
}
hosts[strings.ToLower(d.Host)] = hostConfig{upstream: u, clientID: d.ClientID}
}
return &Proxy{
verifiedHeader: cfg.Proxy.VerifiedHeader,
clientIDHeader: cfg.Proxy.ClientIDHeader,
hosts:          hosts,
requestCount:   requestCount,
}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
host := strings.ToLower(r.Host)
if i := strings.IndexByte(host, ':'); i != -1 {
host = host[:i]
}

hc, ok := p.hosts[host]
if !ok {
http.Error(w, "domain is not active", http.StatusNotFound)
return
}

atomic.AddUint64(p.requestCount, 1)
p.logRequest(r, host)

proxy := httputil.NewSingleHostReverseProxy(hc.upstream)
originalDirector := proxy.Director
proxy.Director = func(req *http.Request) {
originalDirector(req)
req.Header.Set(p.verifiedHeader, "true")
req.Header.Set(p.clientIDHeader, hc.clientID)
}
proxy.ServeHTTP(w, r)
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
