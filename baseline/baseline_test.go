package baseline

import (
	"path/filepath"
	"testing"
)

func TestFile_AddAndHas(t *testing.T) {
	t.Parallel()

	bf := NewFile(15)
	bf.Add("hash1", []string{"a.go", "b.go"}, 42)
	bf.Add("hash2", []string{"c.go"}, 10)
	bf.Add("hash1", []string{"dup.go"}, 99) // duplicate hash ignored

	if !bf.Has("hash1") {
		t.Error("Has(hash1) should be true after Add")
	}

	if bf.Has("unknown") {
		t.Error("Has(unknown) should be false")
	}

	if bf.Len() != 2 {
		t.Errorf("Len() = %d, want 2 (duplicate Add ignored)", bf.Len())
	}
}

func TestFile_AddEmptyHashIgnored(t *testing.T) {
	t.Parallel()

	bf := NewFile(15)
	bf.Add("", []string{"a.go"}, 5)

	if bf.Len() != 0 {
		t.Errorf("empty hash should be ignored, Len() = %d", bf.Len())
	}
}

func TestFile_SaveLoadRoundTrip(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "baseline.json")

	original := NewFile(20)
	original.Add("abc", []string{"a.go", "b.go"}, 42)
	original.Add("def", []string{"c.go"}, 10)

	err := original.Save(path)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.Len() != 2 {
		t.Fatalf("loaded Len() = %d, want 2", loaded.Len())
	}

	if !loaded.Has("abc") || !loaded.Has("def") {
		t.Error("loaded baseline missing expected hashes")
	}

	if loaded.Threshold != 20 {
		t.Errorf("loaded Threshold = %d, want 20", loaded.Threshold)
	}
}

func TestFile_SaveIsSortedAndDeterministic(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	p1 := filepath.Join(dir, "b1.json")
	p2 := filepath.Join(dir, "b2.json")

	for _, p := range []string{p1, p2} {
		bf := NewFile(15)
		bf.Add("zebra", []string{"z.go"}, 1)
		bf.Add("alpha", []string{"a.go"}, 2)
		bf.Add("mango", []string{"m.go"}, 3)

		err := bf.Save(p)
		if err != nil {
			t.Fatalf("Save %s: %v", p, err)
		}
	}

	b1, _ := Load(p1)
	b2, _ := Load(p2)

	if b1.Entries[0].Hash != "alpha" {
		t.Errorf("entries should be sorted by hash; first = %q", b1.Entries[0].Hash)
	}

	for i := range b1.Entries {
		if b1.Entries[i].Hash != b2.Entries[i].Hash {
			t.Errorf("non-deterministic order at %d", i)
		}
	}
}

func TestExists(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	missing := filepath.Join(dir, "nope.json")
	present := filepath.Join(dir, "yes.json")

	bf := NewFile(15)
	_ = bf.Save(present)

	if Exists(missing) {
		t.Error("Exists should be false for missing file")
	}

	if !Exists(present) {
		t.Error("Exists should be true for present file")
	}
}

func TestAnalyzerIdentityRoundTrip(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "baseline.json")
	identity := AnalyzerIdentity{Version: "v1.2.3", CacheVersion: 8, Policy: "0123456789abcdef"}

	original := NewFile(5)
	original.Add("abc", []string{"a.go", "b.go"}, 42)
	original.SetAnalyzer(identity)

	if err := original.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.Analyzer == nil {
		t.Fatal("loaded baseline lost its analyzer identity")
	}

	if *loaded.Analyzer != identity {
		t.Errorf("loaded Analyzer = %+v, want %+v", *loaded.Analyzer, identity)
	}
}

func TestAnalyzerIdentityAbsentInLegacyBaseline(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "legacy.json")

	legacy := NewFile(5)
	if err := legacy.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.Analyzer != nil {
		t.Errorf("baseline saved without SetAnalyzer must load with nil Analyzer, got %+v", *loaded.Analyzer)
	}
}

func TestAnalyzerIdentityCompatibleWith(t *testing.T) {
	t.Parallel()

	running := AnalyzerIdentity{Version: "v2.0.0", CacheVersion: 8, Policy: "aaa"}

	tests := []struct {
		name     string
		recorded AnalyzerIdentity
		want     bool
	}{
		{
			name:     "same policy and revision",
			recorded: AnalyzerIdentity{Version: "v1.0.0", CacheVersion: 8, Policy: "aaa"},
			want:     true,
		},
		{
			name:     "different version but same hash space",
			recorded: AnalyzerIdentity{Version: "v1.0.0-dev", CacheVersion: 8, Policy: "aaa"},
			want:     true,
		},
		{
			name:     "different policy",
			recorded: AnalyzerIdentity{Version: "v2.0.0", CacheVersion: 8, Policy: "bbb"},
			want:     false,
		},
		{
			name:     "different algorithm revision",
			recorded: AnalyzerIdentity{Version: "v2.0.0", CacheVersion: 7, Policy: "aaa"},
			want:     false,
		},
		{
			name:     "unrecorded policy",
			recorded: AnalyzerIdentity{},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := running.CompatibleWith(tt.recorded); got != tt.want {
				t.Errorf("CompatibleWith(%+v) = %v, want %v", tt.recorded, got, tt.want)
			}
		})
	}
}
