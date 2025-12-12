package service

import (
	"log/slog"
	"net/http"

	"github.com/guni1192/cnproxy/pkg/middleware/opentelemetry"
)

type CNProxyHandler struct {
	Logger       *slog.Logger
	ProxyMetrics *opentelemetry.ProxyMetrics
	AllowedFQDNs []string
}

func (h *CNProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("request",
		"method", r.Method,
		"host", r.Host,
		"path", r.URL.Path,
		"protocol", r.Proto,
	)

	switch r.URL.Path {
	case "/health":
		h.Healthcheck(w, r)
	case "":
		h.HandleProxy(w, r)
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}
