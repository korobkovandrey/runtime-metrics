// Package msubnet provides middleware for checking if the IP address from the X-Real-IP header is within a trusted subnet.
package msubnet

import (
	"net"
	"net/http"

	"github.com/korobkovandrey/runtime-metrics/internal/server/handlers"
)

// Middleware checks that the IP address from the X-Real-IP header is within a trusted subnet.
// If the IP address is missing, invalid or not in the subnet, 403 Forbidden is returned.
func Middleware(subnet *net.IPNet) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if subnet == nil {
				next.ServeHTTP(w, r)
				return
			}
			ipStr := r.Header.Get("X-Real-IP")
			if ipStr == "" {
				handlers.RequestCtxWithLogMessage(r, "missing X-Real-IP header")
				w.WriteHeader(http.StatusForbidden)
				return
			}
			ip := net.ParseIP(ipStr)
			if ip == nil {
				handlers.RequestCtxWithLogMessage(r, "invalid X-Real-IP header")
				w.WriteHeader(http.StatusForbidden)
				return
			}
			if !subnet.Contains(ip) {
				handlers.RequestCtxWithLogMessage(r, "X-Real-IP is not in the subnet")
				w.WriteHeader(http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
