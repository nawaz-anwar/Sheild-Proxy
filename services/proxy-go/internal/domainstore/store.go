package domainstore

import (
	"context"
	"log"
	"strings"
	"sync"

	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/config"
)

type Domain struct {
	Host     string
	Upstream string
	ClientID string
	Active   bool
}

type Provider interface {
	Get(ctx context.Context, host string) (Domain, bool, error)
}

type Store struct {
	mu sync.RWMutex
	l1 map[string]Domain
	l2 Provider
	l3 Provider
}

func New(domains []config.DomainConfig, l2 Provider, l3 Provider) *Store {
	l1 := make(map[string]Domain, len(domains))
	for _, d := range domains {
		host := strings.ToLower(strings.TrimSpace(d.Host))
		if host == "" {
			continue
		}
		l1[host] = Domain{
			Host:     host,
			Upstream: d.Upstream,
			ClientID: d.ClientID,
			Active:   d.Active,
		}
	}
	return &Store{l1: l1, l2: l2, l3: l3}
}

func (s *Store) Get(ctx context.Context, host string) (Domain, bool) {
	h := strings.ToLower(strings.TrimSpace(host))
	s.mu.RLock()
	d, ok := s.l1[h]
	s.mu.RUnlock()
	if ok {
		return d, true
	}
	if s.l2 != nil {
		if d, ok, err := s.l2.Get(ctx, h); ok {
			s.putL1(h, d)
			return d, true
		} else if err != nil {
			log.Printf("domainstore l2 lookup failed host=%s err=%v", h, err)
		}
	}
	if s.l3 != nil {
		if d, ok, err := s.l3.Get(ctx, h); ok {
			s.putL1(h, d)
			return d, true
		} else if err != nil {
			log.Printf("domainstore l3 lookup failed host=%s err=%v", h, err)
		}
	}
	return Domain{}, false
}

func (s *Store) putL1(host string, d Domain) {
	s.mu.Lock()
	s.l1[host] = d
	s.mu.Unlock()
}
