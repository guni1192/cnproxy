package server

import (
	"log/slog"
	"net/http"

	"github.com/guni1192/cnproxy/pkg/middleware/logger"
	"github.com/guni1192/cnproxy/pkg/service"
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
		service.Healthcheck(w, r)
	default:
		service.HandleProxy(h.Logger)(w, r)
	}
}

type CNProxyServer struct{}

func (s *CNProxyServer) Serve() error {
	h := &CNProxyHandler{
		Logger: logger.New(),
	}

	h.Logger.Info("listen on :8080")
	return http.ListenAndServe(":8080", h)
}
