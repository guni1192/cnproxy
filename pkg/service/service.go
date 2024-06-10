package service

import (
	"log/slog"
	"net/http"
)

type CNProxyHandler struct {
	Logger *slog.Logger
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
	default:
		h.HandleProxy(w, r)
	}
}
