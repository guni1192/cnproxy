package server

import (
	"net/http"

	"github.com/guni1192/cnproxy/pkg/middleware/logger"
	"github.com/guni1192/cnproxy/pkg/service"
)

type CNProxyServer struct{}

func (s *CNProxyServer) Serve() error {
	h := &service.CNProxyHandler{
		Logger: logger.New(),
	}

	h.Logger.Info("listen on :8080")
	return http.ListenAndServe(":8080", h)
}
