package middleware

import (
	"net"
	"net/http"

	"go.uber.org/zap"
)

// NewTrustedSubnet creates a middleware for checking if a client's IP address
// belongs to a trusted subnet. The middleware examines the X-Real-IP header
// of the request and compares the specified IP address with the given trusted
// subnet in CIDR notation from config.
//
// If the X-Real-IP header is missing, contains an invalid IP address, or
// the IP address is not within the trusted subnet, the middleware returns
// status 403 Forbidden.
//
// Parameters:
//   - l: structured logger for logging operations
//   - trustedSubnet: trusted subnet in CIDR notation (e.g., "127.0.0.1/32") from config
//
// Returns a middleware function that can be used with http.Handler.
func NewTrustedSubnet(l *zap.Logger, trustedSubnet string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !isTrustedSubnet(l, r, trustedSubnet) {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isTrustedSubnet checks whether the IP address from the X-Real-IP header
// belongs to the specified trusted subnet.
func isTrustedSubnet(l *zap.Logger, r *http.Request, trustedSubnet string) bool {
	if trustedSubnet == "" {
		l.Debug("trusted subnet is empty")
		return false
	}

	ipStr := r.Header.Get("X-Real-IP")
	if ipStr == "" {
		l.Debug("empty X-Real-IP header")
		return false
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		l.Debug("invalid ip in X-Real-IP header")
		return false
	}

	_, ipNet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		l.Debug("invalid trusted subnet in config")
		return false
	}

	return ipNet.Contains(ip)
}
