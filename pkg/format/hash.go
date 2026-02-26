// Package format provides formatting utilities for art-dupl.
package format

// Hash converts a uint64 hash to a hex string.
// This is faster than fmt.Sprintf or encoding/hex for fixed-size uint64.
func Hash(h uint64) string {
	const hexchars = "0123456789abcdef"

	buf := make([]byte, 16)
	for i := 15; i >= 0; i-- {
		buf[i] = hexchars[h&0xf]
		h >>= 4
	}

	return string(buf)
}
