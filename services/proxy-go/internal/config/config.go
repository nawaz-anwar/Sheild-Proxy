package config

import (
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type DomainConfig struct {
	Host     string `yaml:"host"`
	Upstream string `yaml:"upstream"`
	ClientID string `yaml:"client_id"`
	Active   bool   `yaml:"active"`
}

type Config struct {
	Phase    string         `yaml:"phase"`
	Proxy    ProxyConfig    `yaml:"proxy"`
	Redis    RedisConfig    `yaml:"redis"`
	Postgres PostgresConfig `yaml:"postgres"`
	JWT      JWTConfig      `yaml:"jwt"`
	GeoIP    GeoIPConfig    `yaml:"geoip"`
}

type ProxyConfig struct {
	VerifiedHeader string               `yaml:"verified_header"`
	ClientIDHeader string               `yaml:"client_id_header"`
	Domains        []DomainConfig       `yaml:"domains"`
	Filter         ProxyFilterConfig    `yaml:"filter"`
	RateLimit      RateLimitConfig      `yaml:"rate_limit"`
	HeaderAnalysis HeaderAnalysisConfig `yaml:"header_analysis"`
	Challenge      ChallengeConfig      `yaml:"challenge"`
}

type ProxyFilterConfig struct {
	AllowlistIPs []string `yaml:"allowlist_ips"`
	BannedIPs    []string `yaml:"banned_ips"`
}

type RateLimitConfig struct {
	Enabled       bool `yaml:"enabled"`
	WindowSeconds int  `yaml:"window_seconds"`
	MaxRequests   int  `yaml:"max_requests"`
}

type HeaderAnalysisConfig struct {
	Enabled            bool     `yaml:"enabled"`
	BlockedBotKeywords []string `yaml:"blocked_bot_keywords"`
	SEOBotAllowlist    []string `yaml:"seo_bot_allowlist"`
}

type ChallengeConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Difficulty int    `yaml:"difficulty"`
	TTLSeconds int    `yaml:"ttl_seconds"`
	CookieName string `yaml:"cookie_name"`
}

type RedisConfig struct {
	Addr      string `yaml:"addr"`
	KeyPrefix string `yaml:"key_prefix"`
}

type PostgresConfig struct {
	DSN string `yaml:"dsn"`
}

type JWTConfig struct {
	Issuer     string `yaml:"issuer"`
	Audience   string `yaml:"audience"`
	Secret     string `yaml:"secret"`
	CookieName string `yaml:"cookie_name"`
	TTLSeconds int    `yaml:"ttl_seconds"`
}

type GeoIPConfig struct {
	Enabled          bool              `yaml:"enabled"`
	DefaultAllow     bool              `yaml:"default_allow"`
	AllowedCountries []string          `yaml:"allowed_countries"`
	BlockedCountries []string          `yaml:"blocked_countries"`
	BlockedASNs      []uint32          `yaml:"blocked_asns"`
	MockData         map[string]GeoHit `yaml:"mock_data"`
}

type GeoHit struct {
	Country string `yaml:"country"`
	ASN     uint32 `yaml:"asn"`
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	applyDefaults(&cfg)
	return &cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Proxy.VerifiedHeader == "" {
		cfg.Proxy.VerifiedHeader = "X-Shield-Verified"
	}
	if cfg.Proxy.ClientIDHeader == "" {
		cfg.Proxy.ClientIDHeader = "X-Shield-Client-ID"
	}
	if cfg.Proxy.RateLimit.WindowSeconds <= 0 {
		cfg.Proxy.RateLimit.WindowSeconds = 60
	}
	if cfg.Proxy.RateLimit.MaxRequests <= 0 {
		cfg.Proxy.RateLimit.MaxRequests = 120
	}
	if cfg.Proxy.Challenge.Difficulty <= 0 {
		cfg.Proxy.Challenge.Difficulty = 3
	}
	if cfg.Proxy.Challenge.TTLSeconds <= 0 {
		cfg.Proxy.Challenge.TTLSeconds = 300
	}
	if cfg.Proxy.Challenge.CookieName == "" {
		cfg.Proxy.Challenge.CookieName = "shield_challenge"
	}
	if cfg.JWT.CookieName == "" {
		cfg.JWT.CookieName = "shield_token"
	}
	if cfg.JWT.TTLSeconds <= 0 {
		cfg.JWT.TTLSeconds = 900
	}
	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = "shield-local-dev-secret"
		log.Printf("warning: using default JWT secret; override in production")
	}
	if cfg.JWT.Issuer == "" {
		cfg.JWT.Issuer = "shield-proxy"
	}
	if cfg.JWT.Audience == "" {
		cfg.JWT.Audience = "shield-origin"
	}
	if cfg.GeoIP.MockData == nil {
		cfg.GeoIP.MockData = map[string]GeoHit{}
	}
	if cfg.Proxy.HeaderAnalysis.BlockedBotKeywords == nil {
		cfg.Proxy.HeaderAnalysis.BlockedBotKeywords = []string{"curl", "python-requests", "scrapy"}
	}
	if cfg.Proxy.HeaderAnalysis.SEOBotAllowlist == nil {
		cfg.Proxy.HeaderAnalysis.SEOBotAllowlist = []string{"googlebot", "bingbot", "duckduckbot", "yandexbot"}
	}
	for i := range cfg.Proxy.Filter.AllowlistIPs {
		cfg.Proxy.Filter.AllowlistIPs[i] = strings.TrimSpace(cfg.Proxy.Filter.AllowlistIPs[i])
	}
	for i := range cfg.Proxy.Filter.BannedIPs {
		cfg.Proxy.Filter.BannedIPs[i] = strings.TrimSpace(cfg.Proxy.Filter.BannedIPs[i])
	}
}
