package server

import (
	"net/http"

	"github.com/guni1192/cnproxy/pkg/service"
)

func Serve() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", service.Healthcheck)

	return http.ListenAndServe(":8080", mux)
}
