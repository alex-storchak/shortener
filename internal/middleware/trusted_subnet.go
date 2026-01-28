package middleware

import (
	"errors"
	"fmt"
	"net"
	"net/http"

	"go.uber.org/zap"
)

var (
	ErrEmptyTrustedSubnet   = errors.New("trusted subnet is empty")
	ErrEmptyXRealIPHeader   = errors.New("empty X-Real-IP header")
	ErrInvalidXRealIPHeader = errors.New("invalid X-Real-IP header")
	ErrNotTrustedSubnet     = errors.New("subnet is not trusted")
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
			if err := checkTrustedSubnet(r, trustedSubnet); err != nil {
				l.Debug("trusted subnet check failed", zap.Error(err))
				w.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// checkTrustedSubnet checks whether the IP address from the X-Real-IP header
// belongs to the specified trusted subnet.
func checkTrustedSubnet(r *http.Request, trustedSubnet string) error {
	if trustedSubnet == "" {
		return ErrEmptyTrustedSubnet
	}

	ipStr := r.Header.Get("X-Real-IP")
	if ipStr == "" {
		return ErrEmptyXRealIPHeader
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return ErrInvalidXRealIPHeader
	}

	_, ipNet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		return fmt.Errorf("invalid trusted subnet in config: %w", err)
	}

	if ipNet.Contains(ip) {
		return ErrNotTrustedSubnet
	}
	return nil
}
