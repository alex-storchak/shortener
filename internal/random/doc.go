// Package random provides cryptographically secure random data generation utilities
// specifically designed for testing and development purposes.
//
// # Security
//
// The package uses crypto/rand for proper random seed generation, ensuring
// cryptographically secure randomness for all generated values. This makes it
// suitable for generating test data in security-sensitive applications.
//
// # Core Functions
//
// The package provides several random generation functions:
//   - ASCIIString: generates random alphanumeric strings
//   - Domain: generates random domain names with configurable TLDs
//   - URL: generates random HTTP URLs with realistic structure
//
// # String Generation
//
// ASCIIString creates valid identifiers that:
//   - Start with a letter (not a digit)
//   - Contain only alphanumeric characters
//   - Have configurable length ranges
//   - Are suitable for use as identifiers, slugs, or test data
//
// # URL and Domain Generation
//
// The URL and Domain functions generate realistic web addresses:
//   - URLs use HTTP scheme with random domains and paths
//   - Domains support custom TLD lists with sensible defaults
//   - Generated values are valid and properly formatted
//
// # Example
//
//	// Generate random test data
//	url := random.URL()
//	domain := random.Domain(5, 10)
//	id := random.ASCIIString(8, 12)
package random
