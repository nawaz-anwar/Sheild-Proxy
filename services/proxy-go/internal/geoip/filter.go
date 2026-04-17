package geoip

import (
	"net"
	"strings"

	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/config"
)

type Result struct {
	Country string
	ASN     uint32
}

type Lookup interface {
	Lookup(ip net.IP) (Result, bool)
}

type StaticLookup struct {
	data map[string]Result
}

func NewStaticLookup(mock map[string]config.GeoHit) *StaticLookup {
	data := make(map[string]Result, len(mock))
	for ip, hit := range mock {
		data[ip] = Result{Country: strings.ToUpper(hit.Country), ASN: hit.ASN}
	}
	return &StaticLookup{data: data}
}

func (s *StaticLookup) Lookup(ip net.IP) (Result, bool) {
	if ip == nil {
		return Result{}, false
	}
	r, ok := s.data[ip.String()]
	return r, ok
}

type Filter struct {
	enabled      bool
	defaultAllow bool
	allowCountry map[string]struct{}
	blockCountry map[string]struct{}
	blockASN     map[uint32]struct{}
	lookup       Lookup
}

func NewFilter(cfg config.GeoIPConfig, lookup Lookup) *Filter {
	allowCountry := map[string]struct{}{}
	for _, c := range cfg.AllowedCountries {
		allowCountry[strings.ToUpper(strings.TrimSpace(c))] = struct{}{}
	}
	blockCountry := map[string]struct{}{}
	for _, c := range cfg.BlockedCountries {
		blockCountry[strings.ToUpper(strings.TrimSpace(c))] = struct{}{}
	}
	blockASN := map[uint32]struct{}{}
	for _, asn := range cfg.BlockedASNs {
		blockASN[asn] = struct{}{}
	}
	return &Filter{
		enabled:      cfg.Enabled,
		defaultAllow: cfg.DefaultAllow,
		allowCountry: allowCountry,
		blockCountry: blockCountry,
		blockASN:     blockASN,
		lookup:       lookup,
	}
}

func (f *Filter) Allow(ip net.IP) bool {
	if !f.enabled {
		return true
	}
	if f.lookup == nil {
		return f.defaultAllow
	}
	hit, ok := f.lookup.Lookup(ip)
	if !ok {
		return f.defaultAllow
	}
	if _, blocked := f.blockASN[hit.ASN]; blocked {
		return false
	}
	if _, blocked := f.blockCountry[strings.ToUpper(hit.Country)]; blocked {
		return false
	}
	if len(f.allowCountry) > 0 {
		_, allowed := f.allowCountry[strings.ToUpper(hit.Country)]
		return allowed
	}
	return true
}
