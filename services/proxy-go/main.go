package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/config"
	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/proxy"
	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/runtime"
)

func main() {
	cfgPath := os.Getenv("CONFIG_FILE")
	if cfgPath == "" {
		cfgPath = "../../config/platform.yaml"
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	if err := applyEnvOverrides(cfg); err != nil {
		log.Fatalf("failed to apply env overrides: %v", err)
	}
	deps, err := runtime.Init(cfg)
	if err != nil {
		log.Fatalf("failed to init runtime dependencies: %v", err)
	}
	defer deps.Close()

	var requestCount uint64
	p := proxy.New(cfg, deps, &requestCount)
	defer p.Close()

	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	httpMux.Handle("/", p)

	metricsMux := http.NewServeMux()
	metricsMux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	metricsMux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte("shield_proxy_requests_total "))
		_, _ = w.Write([]byte(fmtUint(atomic.LoadUint64(&requestCount))))
		_, _ = w.Write([]byte("\n"))
	})

	httpAddr := envOr("PROXY_HTTP_ADDR", ":8080")
	metricsAddr := envOr("PROXY_METRICS_ADDR", ":9090")

	go func() {
		log.Printf("proxy metrics listening on %s", metricsAddr)
		if err := http.ListenAndServe(metricsAddr, metricsMux); err != nil {
			log.Fatalf("metrics server failed: %v", err)
		}
	}()

	log.Printf("proxy listening on %s", httpAddr)
	if err := http.ListenAndServe(httpAddr, httpMux); err != nil {
		log.Fatalf("proxy failed: %v", err)
	}
}

func envOr(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func fmtUint(v uint64) string {
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

func applyEnvOverrides(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}
	if dsn := strings.TrimSpace(os.Getenv("DATABASE_DSN")); dsn != "" {
		cfg.Postgres.DSN = dsn
	}
	originSecretProvided := false
	if jwtSecret, err := envOrFile("JWT_SECRET"); err != nil {
		return err
	} else if jwtSecret != "" {
		cfg.JWT.Secret = jwtSecret
	}
	if originSecret, err := envOrFile("ORIGIN_SECRET"); err != nil {
		return err
	} else if originSecret != "" {
		originSecretProvided = true
		cfg.Proxy.Signing.Secret = originSecret
	}
	if cfg.Proxy.Signing.Secret == "" {
		cfg.Proxy.Signing.Secret = cfg.JWT.Secret
	}
	if isProductionMode() {
		if !originSecretProvided || cfg.Proxy.Signing.Secret == "" {
			return fmt.Errorf("ORIGIN_SECRET must be explicitly configured in production")
		}
		if cfg.Proxy.Signing.Secret == cfg.JWT.Secret {
			return fmt.Errorf("JWT secret and origin signing secret must be different in production")
		}
	} else if originSecretProvided && cfg.Proxy.Signing.Secret == cfg.JWT.Secret {
		log.Printf("warning: JWT secret and origin signing secret are identical; use separate secrets in production")
	}
	if redisPassword, err := envOrFile("REDIS_PASSWORD"); err != nil {
		return err
	} else if redisPassword != "" {
		addr := strings.TrimSpace(cfg.Redis.Addr)
		switch {
		case addr == "":
			// leave runtime validation to existing checks
		case strings.HasPrefix(addr, "redis://") && !strings.Contains(addr, "@"):
			cfg.Redis.Addr = "redis://:" + redisPassword + "@" + strings.TrimPrefix(addr, "redis://")
		case !strings.HasPrefix(addr, "redis://") && !strings.Contains(addr, "@"):
			cfg.Redis.Addr = "redis://:" + redisPassword + "@" + addr
		}
	}
	return nil
}

func envOrFile(key string) (string, error) {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v, nil
	}
	path := strings.TrimSpace(os.Getenv(key + "_FILE"))
	if path == "" {
		return "", nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s_FILE: %w", key, err)
	}
	return strings.TrimSpace(string(b)), nil
}

func isProductionMode() bool {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("ENV")))
	if mode == "" {
		mode = strings.ToLower(strings.TrimSpace(os.Getenv("MODE")))
	}
	return mode == "prod" || mode == "production"
}
