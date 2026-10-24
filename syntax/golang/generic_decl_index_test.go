package golang

import (
	"context"
	"path/filepath"
	"testing"
)

// collectForSource loads type-aware data for a single-file source and returns
// its generic-decl index.
func collectForSource(t *testing.T, src string) map[string]int {
	t.Helper()

	file := filepath.Join(t.TempDir(), "a.go")
	writeFile(t, file, src)

	typeData, err := LoadTypeAwareData(context.Background(), []string{file}, true)
	if err != nil {
		t.Fatalf("LoadTypeAwareData failed: %v", err)
	}

	return CollectGenericDecls(typeData)
}

func TestCollectGenericDecls(t *testing.T) {
	t.Parallel()

	src := `package testpkg

type Box[T any] struct{ value T }

type Pair[K comparable, V any] struct{ key K; val V }

func Map[T, U any](xs []T, f func(T) U) []U { return nil }

type Plain struct{}
`
	index := collectForSource(t, src)

	// go/packages reports the file= pattern package path as the containing
	// directory for unnamed packages ("command-line-arguments" for stdin-like
	// loads); the exact prefix does not matter as long as the bare-name
	// fallback resolves, which the assertions below verify through the
	// qualified lookup path.
	want := map[string]int{
		"Box":  1,
		"Pair": 2,
		"Map":  2,
	}

	if len(index) != len(want) {
		t.Fatalf("index has %d entries, want %d: %v", len(index), len(want), index)
	}

	for name, count := range want {
		var found bool

		for key, got := range index {
			if len(key) > len(name) && key[len(key)-len(name)-1:] == "."+name {
				if got != count {
					t.Errorf("%s indexed with %d type params, want %d", key, got, count)
				}

				found = true

				break
			}
		}

		if !found {
			t.Errorf("%s missing from index %v", name, index)
		}
	}
}

func TestCollectGenericDeclsNilAndEmpty(t *testing.T) {
	t.Parallel()

	if got := CollectGenericDecls(nil); len(got) != 0 {
		t.Errorf("nil data must yield empty index, got %v", got)
	}

	if got := CollectGenericDecls(TypeAwareData{}); len(got) != 0 {
		t.Errorf("empty data must yield empty index, got %v", got)
	}
}

func TestLookupGenericDecl(t *testing.T) {
	t.Parallel()

	index := map[string]int{
		"github.com/x/pkg.Box":  1,
		"github.com/x/pkg.Pair": 2,
		"Local":                 1,
	}

	cases := []struct {
		typeStr string
		want    string
	}{
		{"github.com/x/pkg.Box[int]", "github.com/x/pkg.Box"},
		{"github.com/x/pkg.Box[[]string]", "github.com/x/pkg.Box"},
		{"github.com/x/pkg.Box[pkg2.Pair[int,string]]", "github.com/x/pkg.Box"},
		{"*github.com/x/pkg.Box[int]", "github.com/x/pkg.Box"},
		{"[]github.com/x/pkg.Box[int]", "github.com/x/pkg.Box"},
		{"[2]github.com/x/pkg.Box[int]", "github.com/x/pkg.Box"},
		{"map[string]github.com/x/pkg.Box[int]", "github.com/x/pkg.Box"},
		{"github.com/x/pkg.Pair[int,string]", "github.com/x/pkg.Pair"},
		{"Local[string]", "Local"},
		{"string", ""},
		{"github.com/x/pkg.Box", ""},
		{"", ""},
	}

	for _, tc := range cases {
		if got := LookupGenericDecl(tc.typeStr, index); got != tc.want {
			t.Errorf("LookupGenericDecl(%q) = %q, want %q", tc.typeStr, got, tc.want)
		}
	}
}

func TestFormatGenericParams(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		count int
		want  string
	}{
		{"Box", 1, "Box[T1]"},
		{"Pair", 2, "Pair[T1,T2]"},
		{"Plain", 0, "Plain"},
		{"Weird", -1, "Weird"},
	}

	for _, tc := range cases {
		if got := FormatGenericParams(tc.name, tc.count); got != tc.want {
			t.Errorf("FormatGenericParams(%q, %d) = %q, want %q", tc.name, tc.count, got, tc.want)
		}
	}
}
