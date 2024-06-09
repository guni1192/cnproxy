package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/guni1192/cnproxy/pkg/middleware/logger"
	"github.com/guni1192/cnproxy/pkg/service"
)

type CNProxyServer struct {
	Context context.Context
	Logger  *slog.Logger
}

func (s *CNProxyServer) Serve() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", service.Healthcheck)

	muxWithLogging := logger.LoggingMiddleware(mux, s.Logger)

	s.Logger.Info("listen on :8080")
	return http.ListenAndServe(":8080", muxWithLogging)
}
