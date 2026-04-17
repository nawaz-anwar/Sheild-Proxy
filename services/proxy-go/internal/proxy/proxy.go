package proxy

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/config"
	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/domainstore"
	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/filter"
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
	}
	proxy.ServeHTTP(w, r)
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
