package golang

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

const benchmarkSrc = `package test

type Config struct {
	Name    string
	Port    int
	Enabled bool
}

func (c *Config) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	if c.Port < 1 || c.Port > 65535 {
		return errors.New("invalid port")
	}
	return nil
}

func (c *Config) String() string {
	return fmt.Sprintf("%s:%d", c.Name, c.Port)
}

func process(items []Config) []Config {
	result := make([]Config, 0, len(items))
	for _, item := range items {
		if item.Enabled {
			result = append(result, item)
		}
	}
	return result
}
`

func BenchmarkSemanticVsExactVsStructural(b *testing.B) {
	modes := []struct {
		name string
		mode DetectionMode
	}{
		{"semantic", DetectionModeSemantic},
		{"exact", DetectionModeExact},
		{"structural", DetectionModeStructural},
	}

	tmpDir := b.TempDir()
	tmpFile := filepath.Join(tmpDir, "bench.go")
	if err := os.WriteFile(tmpFile, []byte(benchmarkSrc), 0o644); err != nil {
		b.Fatalf("Failed to write benchmark file: %v", err)
	}

	for _, bm := range modes {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				node, err := ParseWithConfig(tmpFile, ParseConfig{Mode: bm.mode})
				if err != nil {
					b.Fatalf("ParseWithConfig failed: %v", err)
				}
				if node == nil {
					b.Fatal("ParseWithConfig returned nil")
				}

				_ = syntax.Serialize(node)
			}
		})
	}
}
