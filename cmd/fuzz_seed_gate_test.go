package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// Fuzz seed-corpus presence gate: every `func FuzzX` target must have at
// least one committed seed in testdata/fuzz/FuzzX/. A fuzz target with zero
// committed seeds is suspect — `go test ./...` exercises only the f.Add
// seed functions, so regressions that need inputs outside those hand-picked
// examples reach CI invisible, and a found crasher reproduces only on the
// machine that found it. Noted twice in status reports (2026-09-28/29) before
// being built. Exempt a target by listing it here with a reason.
var fuzzSeedExemptions = map[string]string{
	// none yet — snapshot-only/derived targets would go here with rationale
}

func TestFuzzTargetsHaveSeedCorpora(t *testing.T) {
	targets, err := collectFuzzTargets("..")
	if err != nil {
		t.Fatalf("walk repo: %v", err)
	}
	if len(targets) == 0 {
		t.Fatal("found zero fuzz targets; the gate is measuring the wrong thing")
	}

	bare := bareFuzzTargets("..", targets, fuzzSeedExemptions)

	if len(bare) > 0 {
		t.Fatalf("fuzz targets with zero committed seeds: %v; add a seed (run the target briefly, commit a testdata/fuzz/<name>/ entry) or record a reasoned exemption in fuzzSeedExemptions", bare)
	}
}

// collectFuzzTargets maps every fuzz target name in the tree under root to
// its declaring file (root-relative). Nested Go modules are skipped: the nix
// build materializes dependency sources inside the tree (gogenfilter-real/)
// and their fuzz targets belong upstream, not to this repo's seed gate.
func collectFuzzTargets(root string) (map[string]string, error) {
	fuzzFuncRe := regexp.MustCompile(`func (Fuzz[A-Za-z0-9_]+)\(`)

	targets := map[string]string{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "website" || name == "node_modules" || name == "vendor" || name == "testdata" {
				return filepath.SkipDir
			}
			if path != root && isNestedGoModule(path) {
				return filepath.SkipDir
			}

			return nil
		}
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		for _, m := range fuzzFuncRe.FindAllStringSubmatch(string(data), -1) {
			targets[m[1]] = rel
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", root, err)
	}

	return targets, nil
}

func isNestedGoModule(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, "go.mod"))

	return err == nil && !info.IsDir()
}

func bareFuzzTargets(root string, targets, exemptions map[string]string) []string {
	bare := []string{}
	for name, file := range targets {
		if _, exempt := exemptions[name]; exempt {
			continue
		}

		seedDir := filepath.Join(root, filepath.Dir(file), "testdata", "fuzz", name)
		entries, readErr := os.ReadDir(seedDir)
		if readErr != nil || len(entries) == 0 {
			bare = append(bare, name+" ("+file+")")
		}
	}
	sort.Strings(bare)

	return bare
}

// TestCollectFuzzTargetsSkipsNestedModules pins the sandbox failure class:
// the nix build copies dependency sources (gogenfilter-real/) into the repo
// root before tests run, and their fuzz targets must not be gated as ours.
func TestCollectFuzzTargetsSkipsNestedModules(t *testing.T) {
	root := t.TempDir()

	// fuzzDecl builds a fuzz-func declaration so this file's own source does
	// not match the seed gate's scanner.
	fuzzDecl := func(name string) string {
		return name + "(f *testing.F)"
	}

	write := func(rel, content string) {
		t.Helper()

		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}

	// The fuzz-func literals are split so this file's own scanner cannot see
	// them as real targets.
	write("pkg/ours_test.go", "package pkg\n\nfunc "+fuzzDecl("FuzzOurs")+" {}\n")
	write("pkg/testdata/fuzz/FuzzOurs/seed", "seed")
	write("dep/go.mod", "module example.com/dep\n\ngo 1.27\n")
	write("dep/theirs_test.go", "package dep\n\nfunc "+fuzzDecl("FuzzTheirs")+" {}\n")

	targets, err := collectFuzzTargets(root)
	if err != nil {
		t.Fatalf("collectFuzzTargets: %v", err)
	}

	if _, found := targets["FuzzTheirs"]; found {
		t.Errorf("nested-module fuzz target leaked into the gate: %v", targets)
	}
	if _, found := targets["FuzzOurs"]; !found {
		t.Errorf("own fuzz target missed: %v", targets)
	}

	bare := bareFuzzTargets(root, targets, map[string]string{})
	if len(bare) != 0 {
		t.Errorf("seeded own target reported bare: %v", bare)
	}
}
