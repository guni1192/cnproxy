package service

import (
	"io"
	"net"
	"net/http"
	"time"
)

func (h *CNProxyHandler) HandleProxy(w http.ResponseWriter, r *http.Request) {
	if h.ProxyMetrics != nil {
		h.ProxyMetrics.TotalRequests.Add(r.Context(), 1)
	}
	if r.Method == http.MethodConnect {
		h.httpsProxy(w, r)
	} else {
		h.httpProxy(w, r)
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *CNProxyHandler) httpsProxy(w http.ResponseWriter, r *http.Request) {
	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		h.Logger.Warn("failed to connect", "error", err)
		return
	}
	defer destConn.Close()

	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijack not supported", http.StatusInternalServerError)
		h.Logger.Error("hijack not supported")
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		h.Logger.Error("failed to hijack", "error", err)
	}
	defer clientConn.Close()

	go io.Copy(destConn, clientConn)
	io.Copy(clientConn, destConn)
}

func (h *CNProxyHandler) httpProxy(w http.ResponseWriter, r *http.Request) {
	targetURL := "http" + "://" + r.Host + r.URL.Path

	req, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.Logger.Error("failed to create request", "error", err)
		return
	}
	req.Header = r.Header
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		h.Logger.Error("failed to send request", "error", err)
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
