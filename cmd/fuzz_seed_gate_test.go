package cmd

import (
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
	fuzzFuncRe := regexp.MustCompile(`func (Fuzz[A-Za-z0-9_]+)\(`)

	targets := map[string]string{} // name -> declaring file (repo-relative)
	err := filepath.Walk("..", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "website" || name == "node_modules" || name == "vendor" || name == "testdata" {
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
		rel, relErr := filepath.Rel("..", path)
		if relErr != nil {
			return relErr
		}
		for _, m := range fuzzFuncRe.FindAllStringSubmatch(string(data), -1) {
			targets[m[1]] = rel
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walk repo: %v", err)
	}
	if len(targets) == 0 {
		t.Fatal("found zero fuzz targets; the gate is measuring the wrong thing")
	}

	bare := []string{}
	for name, file := range targets {
		if _, exempt := fuzzSeedExemptions[name]; exempt {
			continue
		}

		seedDir := filepath.Join("..", filepath.Dir(file), "testdata", "fuzz", name)
		entries, readErr := os.ReadDir(seedDir)
		if readErr != nil || len(entries) == 0 {
			bare = append(bare, name+" ("+file+")")
		}
	}
	sort.Strings(bare)

	if len(bare) > 0 {
		t.Fatalf("fuzz targets with zero committed seeds: %v; add a seed (run the target briefly, commit a testdata/fuzz/<name>/ entry) or record a reasoned exemption in fuzzSeedExemptions", bare)
	}
}
