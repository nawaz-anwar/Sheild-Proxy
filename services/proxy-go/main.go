package main

import (
"log"
"net/http"
"os"
"sync/atomic"

"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/config"
"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/proxy"
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

var requestCount uint64
p := proxy.New(cfg, &requestCount)

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
