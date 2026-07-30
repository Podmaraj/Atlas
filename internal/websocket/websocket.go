package websocket

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"edgecore/internal/logger"
	"edgecore/internal/models"
)

// IsWebSocketRequest checks if request contains WebSocket upgrade headers
func IsWebSocketRequest(r *http.Request) bool {
	containsHeader := func(key, val string) bool {
		for _, v := range r.Header[key] {
			if strings.Contains(strings.ToLower(v), strings.ToLower(val)) {
				return true
			}
		}
		return false
	}
	return containsHeader("Connection", "upgrade") && containsHeader("Upgrade", "websocket")
}

// IsSSERequest checks if request is for Server-Sent Events stream
func IsSSERequest(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "text/event-stream")
}

// ProxyWebSocket performs bi-directional TCP hijacking for real-time WebSocket connection
func ProxyWebSocket(w http.ResponseWriter, r *http.Request, target *models.ServiceInstance) error {
	destConn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", target.Host, target.Port), 10*time.Second)
	if err != nil {
		http.Error(w, "Failed to connect to upstream WebSocket server", http.StatusBadGateway)
		return fmt.Errorf("failed to dial target websocket server: %w", err)
	}
	defer destConn.Close()

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Webserver doesn't support hijacking", http.StatusInternalServerError)
		return fmt.Errorf("responsewriter does not implement http.Hijacker")
	}

	clientConn, _, err := hj.Hijack()
	if err != nil {
		return fmt.Errorf("failed to hijack client connection: %w", err)
	}
	defer clientConn.Close()

	// Write original request header to destination connection
	if err := r.Write(destConn); err != nil {
		logger.Error("Failed to write request header to target websocket server", zap.Error(err))
		return err
	}

	errChan := make(chan error, 2)
	go func() {
		_, err := io.Copy(destConn, clientConn)
		errChan <- err
	}()

	go func() {
		_, err := io.Copy(clientConn, destConn)
		errChan <- err
	}()

	<-errChan
	return nil
}
