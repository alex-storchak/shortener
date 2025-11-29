package random

import (
	"net/url"
	"strings"
)

// URL generates a random valid HTTP URL with realistic structure.
// The URL includes a scheme, domain, and optional path segments.
//
// Returns:
//   - *url.URL: pointer to a parsed URL structure with random components
//
// Example:
//
//	randomURL := URL()
//	// Result: "http://example.com/path/segment", "http://test.net/abc", etc.
func URL() *url.URL {
	var res url.URL

	// generate SCHEME
	res.Scheme = "http"
	res.Host = Domain(5, 15)

	for i := 0; i < rnd.Intn(4); i++ {
		res.Path += "/" + strings.ToLower(ASCIIString(5, 15))
	}
	return &res
}

// Domain generates a random valid domain name with configurable parameters.
// The domain consists of a hostname and a TLD (top-level domain).
//
// Parameters:
//   - minLen: minimum length for the hostname part
//   - maxLen: maximum length for the hostname part
//   - zones: optional list of TLDs to choose from (default: com, ru, net, biz, yandex)
//
// Returns:
//   - string: random domain name in format "hostname.tld"
//
// Example:
//
//	domain := Domain(5, 10, "com", "net")
//	// Result: "example.com", "test.net", etc.
func Domain(minLen, maxLen int, zones ...string) string {
	if minLen == 0 {
		minLen = 5
	}
	if maxLen == 0 {
		maxLen = 15
	}

	// generate ZONE
	var zone string
	switch len(zones) {
	case 1:
		zone = zones[0]
	case 0:
		zones = []string{"com", "ru", "net", "biz", "yandex"}
		zone = zones[rnd.Intn(len(zones))]
	default:
		zone = zones[rnd.Intn(len(zones))]
	}

	// generate HOST
	host := strings.ToLower(ASCIIString(minLen, maxLen))
	return host + "." + strings.TrimLeft(zone, ".")
}
