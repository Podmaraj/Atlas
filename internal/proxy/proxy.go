package proxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"

	"edgecore/internal/logger"
	"edgecore/internal/models"
)

// ProxyEngine manages high-performance HTTP reverse proxy forwarding
type ProxyEngine struct {
	transport *http.Transport
}

// NewProxyEngine creates a ProxyEngine with tuned connection pooling & timeouts
func NewProxyEngine() *ProxyEngine {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          2000,
		MaxIdleConnsPerHost:   200,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
	}

	return &ProxyEngine{
		transport: transport,
	}
}

// ProxyRequest forwards an incoming request to an upstream service target
func (pe *ProxyEngine) ProxyRequest(
	w http.ResponseWriter,
	r *http.Request,
	target *models.ServiceInstance,
	route *models.Route,
	matchedPath string,
) error {
	scheme := "http"
	targetURL, err := url.Parse(fmt.Sprintf("%s://%s:%d", scheme, target.Host, target.Port))
	if err != nil {
		return fmt.Errorf("invalid upstream target URL: %w", err)
	}

	proxy := &httputil.ReverseProxy{
		Transport: pe.transport,
		Director: func(outReq *http.Request) {
			outReq.URL.Scheme = targetURL.Scheme
			outReq.URL.Host = targetURL.Host

			// Handle path rewrites and stripping
			reqPath := r.URL.Path
			if route.StripPath && matchedPath != "" {
				trimmed := strings.TrimPrefix(reqPath, matchedPath)
				if !strings.HasPrefix(trimmed, "/") {
					trimmed = "/" + trimmed
				}
				reqPath = trimmed
			}

			outReq.URL.Path = singleJoiningSlash(targetURL.Path, reqPath)
			outReq.URL.RawQuery = r.URL.RawQuery

			// Handle Host header
			if route.PreserveHost {
				outReq.Host = r.Host
			} else {
				outReq.Host = targetURL.Host
			}

			// Pass through client IP and headers
			clientIP, _, _ := net.SplitHostPort(r.RemoteAddr)
			if clientIP != "" {
				if prior := outReq.Header.Get("X-Forwarded-For"); prior != "" {
					clientIP = prior + ", " + clientIP
				}
				outReq.Header.Set("X-Forwarded-For", clientIP)
			}
			outReq.Header.Set("X-Forwarded-Host", r.Host)
			if r.TLS != nil {
				outReq.Header.Set("X-Forwarded-Proto", "https")
			} else {
				outReq.Header.Set("X-Forwarded-Proto", "http")
			}

			// Add Gateway Request ID header
			if reqID, ok := r.Context().Value(logger.RequestIDKey).(string); ok && reqID != "" {
				outReq.Header.Set("X-Gateway-Request-ID", reqID)
			}
		},
		ErrorHandler: func(rw http.ResponseWriter, req *http.Request, proxyErr error) {
			logger.Error("Reverse proxy upstream connection error",
				zap.Error(proxyErr),
				zap.String("upstream", targetURL.String()),
				zap.String("path", req.URL.Path),
			)

			if req.Context().Err() == context.Canceled {
				rw.WriteHeader(499) // Client Closed Request
				return
			}

			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusBadGateway)
			_, _ = rw.Write([]byte(`{"error": "Bad Gateway", "message": "Upstream service unavailable"}`))
		},
	}

	proxy.ServeHTTP(w, r)
	return nil
}

func singleJoiningSlash(a, b string) string {
	asSlash := strings.HasSuffix(a, "/")
	bsSlash := strings.HasPrefix(b, "/")
	switch {
	case asSlash && bsSlash:
		return a + b[1:]
	case !asSlash && !bsSlash:
		return a + "/" + b
	}
	return a + b
}
