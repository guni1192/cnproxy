package server

import (
	"fmt"
	"net/http"

	"github.com/guni1192/cnproxy/pkg/middleware/logger"
	"github.com/guni1192/cnproxy/pkg/service"
)

type CNProxyServer struct {
	Port    uint
	Address string
}

func (s *CNProxyServer) Serve() error {
	h := &service.CNProxyHandler{
		Logger: logger.New(),
	}

	host := fmt.Sprintf("%s:%d", s.Address, s.Port)
	h.Logger.Info("listening", "address", s.Address, "port", s.Port)
	return http.ListenAndServe(host, h)
}
