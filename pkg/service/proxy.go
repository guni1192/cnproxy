package service

import (
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"
)

func HandleProxy(l *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodConnect {
			httpsProxy(l, w, r)
		} else {
			httpProxy(l, w, r)
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func httpsProxy(l *slog.Logger, w http.ResponseWriter, r *http.Request) {
	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		l.Warn("failed to connect", "error", err)
		return
	}
	defer destConn.Close()

	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijack not supported", http.StatusInternalServerError)
		l.Error("hijack not supported")
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		l.Error("failed to hijack", "error", err)
	}
	defer clientConn.Close()

	go io.Copy(destConn, clientConn)
	io.Copy(clientConn, destConn)
}

func httpProxy(l *slog.Logger, w http.ResponseWriter, r *http.Request) {
	targetURL := "http" + "://" + r.Host + r.URL.Path

	req, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		l.Error("failed to create request", "error", err)
		return
	}
	req.Header = r.Header
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		l.Error("failed to send request", "error", err)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	io.Copy(w, resp.Body)
}
