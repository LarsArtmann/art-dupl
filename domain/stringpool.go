package domain

import (
	"sync"
)

// StringID represents an interned string identifier.
// Uses 32-bit indices for compact storage (4B vs 16B for strings).
// Supports up to 4 billion unique strings, sufficient for practical use.
type StringID uint32

// NewStringID creates a validated StringID from a uint32.
func NewStringID(id uint32) StringID {
	if id == 0 {
		return 0 // ID 0 reserved for "not interned"
	}
	return StringID(id)
}

// IsValid checks if StringID is valid (non-zero).
func (sid StringID) IsValid() bool {
	return sid != 0
}

// Uint32 returns the underlying uint32 value.
func (sid StringID) Uint32() uint32 {
	return uint32(sid)
}

// MarshalJSON implements json.Marshaler for StringID.
// Serializes as the underlying string value from the global pool.
func (sid StringID) MarshalJSON() ([]byte, error) {
	if !sid.IsValid() {
		return []byte("null"), nil
	}
	str := GlobalPool().Lookup(sid)
	if str == "" {
		return []byte("null"), nil
	}
	return []byte("\"" + str + "\""), nil
}

// UnmarshalJSON implements json.Unmarshaler for StringID.
// Deserializes by interning the string value.
func (sid *StringID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*sid = 0
		return nil
	}
	// Remove quotes
	str := string(data[1 : len(data)-1])
	if str == "" {
		*sid = 0
		return nil
	}
	*sid = GlobalPool().Intern(str)
	return nil
}

// StringInternPool provides thread-safe string interning.
// Eliminates duplicate string storage by sharing common strings.
// Uses sync.RWMutex for efficient concurrent read access.
//
// Memory Impact:
// - Without interning: Each Clone has 16B string header + heap allocation
// - With interning: Each Clone has 4B StringID + shared string storage
// - Typical savings: 20-40% for filenames with duplication.
type StringInternPool struct {
	mu      sync.RWMutex
	strings []string
	index   map[string]StringID
	nextID  StringID
}

// NewStringInternPool creates a new string interning pool.
// Pre-allocates space for expected string count to reduce allocations.
func NewStringInternPool(expectedStrings int) *StringInternPool {
	return &StringInternPool{
		strings: make([]string, 0, expectedStrings),
		index:   make(map[string]StringID, expectedStrings/4),
		nextID:  1, // ID 0 reserved for "not interned"
	}
}

// Intern adds a string to the pool and returns its ID.
// Thread-safe: Multiple goroutines can call concurrently.
// Returns existing ID if string already interned.
func (p *StringInternPool) Intern(s string) StringID {
	if s == "" {
		return 0 // Empty string not interned
	}

	// Fast path: read lock for existing strings
	p.mu.RLock()
	if id, exists := p.index[s]; exists {
		p.mu.RUnlock()
		return id
	}
	p.mu.RUnlock()

	// Slow path: write lock to add new string
	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check: string might have been added while we were upgrading
	if id, exists := p.index[s]; exists {
		return id
	}

	// Add new string to pool
	id := p.nextID
	p.strings = append(p.strings, s)
	p.index[s] = id
	p.nextID++

	return id
}

// Lookup retrieves a string by its ID.
// Returns empty string if ID is invalid or not found.
func (p *StringInternPool) Lookup(id StringID) string {
	if id == 0 || !id.IsValid() {
		return ""
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	idx := int(id) - 1 // Adjust for 1-based IDs
	if idx >= 0 && idx < len(p.strings) {
		return p.strings[idx]
	}
	return ""
}

// Len returns the number of interned strings.
func (p *StringInternPool) Len() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.strings)
}

// Stats returns pool statistics for monitoring.
func (p *StringInternPool) Stats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return PoolStats{
		TotalStrings:    len(p.strings),
		TotalIDs:        int(p.nextID) - 1,
		UniqueFilenames: 0, // Could track by prefix if needed
	}
}

// PoolStats represents string pool statistics.
type PoolStats struct {
	TotalStrings    int `json:"totalStrings"`
	TotalIDs        int `json:"totalIds"`
	UniqueFilenames int `json:"uniqueFilenames"`
}

// GlobalPool is the default shared pool for filename interning.
// Initialized lazily on first use.
//
//nolint:gochecknoglobals // Thread-safe singleton pattern for shared string pool
var (
	globalPool     *StringInternPool
	globalPoolOnce = &sync.Once{}
)

// GlobalPool returns the shared string interning pool.
// Thread-safe singleton pattern.
func GlobalPool() *StringInternPool {
	// Fast path: pool already initialized (including test injection)
	if globalPool != nil {
		return globalPool
	}
	globalPoolOnce.Do(func() {
		// Double-check in case of race
		if globalPool == nil {
			globalPool = NewStringInternPool(1000) // Expect up to 1000 unique filenames
		}
	})
	return globalPool
}

// SetGlobalPoolForTesting sets a custom global pool for testing purposes.
// Returns a cleanup function to restore the original state.
// Must be called from a non-concurrent test context.
func SetGlobalPoolForTesting(pool *StringInternPool) func() {
	originalPool := globalPool
	originalOnce := globalPoolOnce

	globalPool = pool
	globalPoolOnce = &sync.Once{}

	return func() {
		globalPool = originalPool
		globalPoolOnce = originalOnce
	}
}
