package service

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// TestHttpsProxyConnectResponseHeaders pins the fix for the Transfer-Encoding
// regression: a 2xx response to CONNECT must not advertise Transfer-Encoding or
// Content-Length (RFC 7230 §3.3 / RFC 9112 §6.1). If net/http auto-frames the
// response as chunked, Go HTTP clients decode the tunnel as chunked data and
// the TLS handshake silently breaks.
func TestHttpsProxyConnectResponseHeaders(t *testing.T) {
	backend, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen backend: %v", err)
	}
	defer func() { _ = backend.Close() }()

	go func() {
		for {
			c, err := backend.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer func() { _ = c.Close() }()
				_, _ = io.Copy(c, c) // echo
			}(c)
		}
	}()

	host, _, err := net.SplitHostPort(backend.Addr().String())
	if err != nil {
		t.Fatalf("split host: %v", err)
	}

	handler := &CNProxyHandler{
		Logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
		AllowedFQDNs: []string{host},
	}
	proxy := httptest.NewServer(handler)
	defer proxy.Close()

	proxyURL, err := url.Parse(proxy.URL)
	if err != nil {
		t.Fatalf("parse proxy url: %v", err)
	}

	proxyConn, err := net.Dial("tcp", proxyURL.Host)
	if err != nil {
		t.Fatalf("dial proxy: %v", err)
	}
	defer func() { _ = proxyConn.Close() }()

	target := backend.Addr().String()
	if _, err := fmt.Fprintf(proxyConn,
		"CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", target, target); err != nil {
		t.Fatalf("write CONNECT: %v", err)
	}

	if err := proxyConn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("set deadline: %v", err)
	}
	br := bufio.NewReader(proxyConn)
	resp, err := http.ReadResponse(br, &http.Request{Method: http.MethodConnect})
	if err != nil {
		t.Fatalf("read CONNECT response: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: want 200, got %d", resp.StatusCode)
	}
	if len(resp.TransferEncoding) != 0 {
		t.Errorf("Transfer-Encoding: want none (RFC 7230 §3.3), got %v", resp.TransferEncoding)
	}
	if got := resp.Header.Get("Transfer-Encoding"); got != "" {
		t.Errorf("Transfer-Encoding header: want empty, got %q", got)
	}
	if got := resp.Header.Get("Content-Length"); got != "" {
		t.Errorf("Content-Length header: want empty (RFC 7230 §3.3), got %q", got)
	}

	// Verify the tunnel itself works once the response is past.
	const payload = "ping-through-tunnel\n"
	if _, err := proxyConn.Write([]byte(payload)); err != nil {
		t.Fatalf("write to tunnel: %v", err)
	}
	got, err := br.ReadString('\n')
	if err != nil {
		t.Fatalf("read echo: %v", err)
	}
	if got != payload {
		t.Errorf("echo: want %q, got %q", payload, got)
	}
}

// TestHttpsProxyForwardsPipelinedClientBytes guards against dropping bytes the
// client may have pipelined immediately after the CONNECT request line. Such
// bytes land in net/http's bufio.Reader and would be lost if the proxy copied
// from the raw net.Conn instead of the hijacked bufio.Reader.
func TestHttpsProxyForwardsPipelinedClientBytes(t *testing.T) {
	backend, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen backend: %v", err)
	}
	defer func() { _ = backend.Close() }()

	received := make(chan []byte, 1)
	go func() {
		c, err := backend.Accept()
		if err != nil {
			return
		}
		defer func() { _ = c.Close() }()
		buf := make([]byte, 64)
		n, err := c.Read(buf)
		if err != nil {
			received <- nil
			return
		}
		received <- buf[:n]
	}()

	host, _, _ := net.SplitHostPort(backend.Addr().String())
	handler := &CNProxyHandler{
		Logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
		AllowedFQDNs: []string{host},
	}
	proxy := httptest.NewServer(handler)
	defer proxy.Close()

	proxyURL, _ := url.Parse(proxy.URL)
	proxyConn, err := net.Dial("tcp", proxyURL.Host)
	if err != nil {
		t.Fatalf("dial proxy: %v", err)
	}
	defer func() { _ = proxyConn.Close() }()

	target := backend.Addr().String()
	const pipelined = "PIPELINED-BYTES"
	// Send CONNECT and the pipelined bytes back-to-back so net/http reads them
	// into its bufio.Reader along with the request.
	req := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n%s",
		target, target, pipelined)
	if _, err := proxyConn.Write([]byte(req)); err != nil {
		t.Fatalf("write CONNECT+payload: %v", err)
	}

	if err := proxyConn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("set deadline: %v", err)
	}
	br := bufio.NewReader(proxyConn)
	resp, err := http.ReadResponse(br, &http.Request{Method: http.MethodConnect})
	if err != nil {
		t.Fatalf("read CONNECT response: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: want 200, got %d", resp.StatusCode)
	}

	select {
	case got := <-received:
		if !strings.HasPrefix(string(got), pipelined) {
			t.Errorf("backend received %q, want prefix %q", got, pipelined)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("backend did not receive pipelined bytes")
	}
}
