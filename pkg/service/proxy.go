package service

import (
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

func (h *CNProxyHandler) HandleProxy(w http.ResponseWriter, r *http.Request) {
	if h.ProxyMetrics != nil {
		h.ProxyMetrics.TotalRequests.Add(r.Context(), 1)
	}

	if !h.isFQDNAllowed(r.Host) {
		http.Error(w, "FQDN not allowed", http.StatusForbidden)
		h.Logger.Warn("FQDN not allowed", "host", r.Host)
		return
	}

	if r.Method == http.MethodConnect {
		h.httpsProxy(w, r)
	} else {
		h.httpProxy(w, r)
	}
}

func (h *CNProxyHandler) isFQDNAllowed(host string) bool {
	if len(h.AllowedFQDNs) == 0 {
		return true
	}

	hostname, _, err := net.SplitHostPort(host)
	if err != nil {
		hostname = host
	}

	for _, allowedFQDN := range h.AllowedFQDNs {
		if matchHost(allowedFQDN, hostname) {
			return true
		}
	}

	return false
}

// matchHost reports whether host matches pattern.
//
// A leading "*." in pattern is a wildcard for one or more DNS labels, so
// "*.example.com" matches "a.example.com" and "a.b.example.com" but not
// "example.com" itself. Any other pattern is an exact, case-insensitive
// match.
func matchHost(pattern, host string) bool {
	pattern = strings.ToLower(pattern)
	host = strings.ToLower(host)
	if suffix, ok := strings.CutPrefix(pattern, "*."); ok {
		// host must end in ".suffix" with at least one label before it.
		dotSuffix := "." + suffix
		return strings.HasSuffix(host, dotSuffix) && len(host) > len(dotSuffix)
	}
	return pattern == host
}

func (h *CNProxyHandler) httpsProxy(w http.ResponseWriter, r *http.Request) {
	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		h.Logger.Warn("failed to connect", "error", err)
		return
	}
	defer func() {
		if err := destConn.Close(); err != nil {
			h.Logger.Warn("failed to close destination connection", "error", err)
		}
	}()

	// Hijack BEFORE writing any response. If we let net/http synthesize the
	// 200 it adds Transfer-Encoding: chunked (no Content-Length was set), and
	// RFC 7230 §3.3 / RFC 9112 §6.1 forbid Transfer-Encoding and
	// Content-Length on a 2xx response to CONNECT. Conforming clients (Go's
	// net/http transport in particular) then read the tunnel stream through a
	// chunked decoder, corrupting the TLS handshake.
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijack not supported", http.StatusInternalServerError)
		h.Logger.Error("hijack not supported")
		return
	}

	clientConn, brw, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		h.Logger.Error("failed to hijack", "error", err)
		return
	}
	defer func() {
		if err := clientConn.Close(); err != nil {
			h.Logger.Warn("failed to close client connection", "error", err)
		}
	}()

	if _, err := brw.WriteString("HTTP/1.1 200 Connection Established\r\n\r\n"); err != nil {
		h.Logger.Warn("failed to write 200 response", "error", err)
		return
	}
	if err := brw.Flush(); err != nil {
		h.Logger.Warn("failed to flush 200 response", "error", err)
		return
	}

	// Read from brw.Reader, not clientConn, so any bytes the client pipelined
	// after the CONNECT request (e.g. the start of a TLS ClientHello) that
	// landed in net/http's buffered reader are forwarded instead of dropped.
	go func() {
		if _, e := io.Copy(destConn, brw.Reader); e != nil {
			h.Logger.Warn("failed to copy client to destination", "error", e)
		}
	}()

	if _, err := io.Copy(clientConn, destConn); err != nil {
		h.Logger.Warn("failed to copy destination to client", "error", err)
	}
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
	defer func() {
		if err := resp.Body.Close(); err != nil {
			h.Logger.Warn("failed to close response body", "error", err)
		}
	}()

	w.WriteHeader(resp.StatusCode)
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		h.Logger.Warn("failed to copy response to client", "error", err)
	}
}
