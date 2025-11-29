package random

import (
	"crypto/rand"
	"encoding/binary"
	"io"
	mathrand "math/rand"
)

// rnd is a cryptographically secure random number generator instance.
// It is initialized once per binary execution with a true random seed
// from crypto/rand to ensure proper randomness for all generated values.
var rnd = func() *mathrand.Rand {
	buf := make([]byte, 8)
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		panic(err)
	}
	src := mathrand.NewSource(int64(binary.LittleEndian.Uint64(buf)))
	return mathrand.New(src)
}()
