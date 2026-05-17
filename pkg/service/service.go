package service

import (
	"log/slog"
	"net/http"

	"github.com/guni1192/cnproxy/pkg/config"
	"github.com/guni1192/cnproxy/pkg/middleware/opentelemetry"
)

type CNProxyHandler struct {
	Logger       *slog.Logger
	ProxyMetrics *opentelemetry.ProxyMetrics
	AllowedFQDNs []string
	HTTPFilters  []config.HTTPFilter
}

func (h *CNProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("request",
		"method", r.Method,
		"host", r.Host,
		"path", r.URL.Path,
		"protocol", r.Proto,
	)

	// Proxy requests are either CONNECT (RFC 9112 §3.2.3) or use absolute-form
	// request targets (§3.2.2). Anything else is directed at the proxy itself,
	// so route it on path.
	if r.Method == http.MethodConnect || r.URL.IsAbs() {
		h.HandleProxy(w, r)
		return
	}

	switch r.URL.Path {
	case "/health":
		h.Healthcheck(w, r)
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}
