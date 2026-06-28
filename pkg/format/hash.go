// Package format provides formatting utilities for art-dupl.
package format

// Hash converts a uint64 hash to a hex string.
// This is faster than fmt.Sprintf or encoding/hex for fixed-size uint64.
func Hash(hash uint64) string {
	const hexchars = "0123456789abcdef"

	buf := make([]byte, 16) //nolint:mnd,makezero // 16-char hex buffer filled backward by index
	for i := 15; i >= 0; i-- {
		buf[i] = hexchars[hash&0xf]
		hash >>= 4
	}

	return string(buf)
}
