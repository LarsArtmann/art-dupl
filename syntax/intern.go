package syntax

import "sync"

// filenameInterner deduplicates filename strings across all AST nodes.
// In large codebases, the same filename appears thousands of times —
// each *Node stores a copy. Interning reduces this to one pointer per
// unique filename, cutting memory significantly.
var (
	filenamePool sync.Map
)

// InternFilename returns a canonicalized, deduplicated copy of the filename.
// Subsequent calls with equal strings return the same underlying string data,
// reducing memory usage for AST trees with repeated filenames.
func InternFilename(filename string) string {
	if v, ok := filenamePool.Load(filename); ok {
		return v.(string) //nolint:forcetypeassert // stored as string by us
	}

	// Store a new copy — the caller's string may be a slice of a larger buffer
	canonical := string([]byte(filename))
	actual, _ := filenamePool.LoadOrStore(filename, canonical)

	return actual.(string) //nolint:forcetypeassert // stored as string by us
}
