package random

// ASCIIString generates a random ASCII string consisting of alphanumeric characters.
// The generated string starts with a letter (not a digit) to ensure it's a valid identifier.
//
// Parameters:
//   - minLen: minimum length of the generated string (inclusive)
//   - maxLen: maximum length of the generated string (exclusive)
//
// Returns:
//   - string: random ASCII string containing letters and digits
//
// Example:
//
//	str := ASCIIString(5, 10)
//	// Result: "Abc123", "XyZ987", etc.
func ASCIIString(minLen, maxLen int) string {
	var letters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFJHIJKLMNOPQRSTUVWXYZ"

	slen := rnd.Intn(maxLen-minLen) + minLen

	s := make([]byte, 0, slen)
	i := 0
	for len(s) < slen {
		idx := rnd.Intn(len(letters) - 1)
		char := letters[idx]
		if i == 0 && '0' <= char && char <= '9' {
			continue
		}
		s = append(s, char)
		i++
	}

	return string(s)
}
