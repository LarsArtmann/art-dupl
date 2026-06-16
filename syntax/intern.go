package syntax

import "sync"

// filenamePool deduplicates filename strings across all AST nodes.
// In large codebases, the same filename appears thousands of times —
// each *Node stores a copy. Interning reduces this to one pointer per
// unique filename, cutting memory significantly.
var (
	filenameMu   sync.RWMutex
	filenamePool = make(map[string]string)
)

// InternFilename returns a canonicalized, deduplicated copy of the filename.
// Subsequent calls with equal strings return the same underlying string data,
// reducing memory usage for AST trees with repeated filenames.
func InternFilename(filename string) string {
	filenameMu.RLock()

	if s, ok := filenamePool[filename]; ok {
		filenameMu.RUnlock()

		return s
	}

	filenameMu.RUnlock()

	canonical := string([]byte(filename))

	filenameMu.Lock()
	defer filenameMu.Unlock()

	// Double-check: another goroutine may have stored while we waited
	if s, ok := filenamePool[filename]; ok {
		return s
	}

	filenamePool[filename] = canonical

	return canonical
}
