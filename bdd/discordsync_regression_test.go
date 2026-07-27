package bdd_test

import (
	"encoding/json/v2"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	. "github.com/LarsArtmann/art-dupl/bdd"
	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/printer"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// DiscordSync regression corpus: real-world false-positive patterns extracted
// from a DiscordSync feedback session (82 groups, 97% false positives at -t 1).
// The fixtures live in testdata/discordsync/ (ignored by go build). The test
// copies them to a temp dir so the testdata-pair actionability pattern does not
// suppress everything.
var _ = Describe("DiscordSync Regression Corpus", func() {
	var (
		setup  *testutil.BDDTestSetup
		tmpDir string
	)

	BeforeEach(func() {
		setup = CreateBDDTestSetup()

		_, filename, _, _ := runtime.Caller(0)
		fixturesDir := filepath.Join(filepath.Dir(filename), "..", "testdata", "discordsync")

		tmpDir = GinkgoT().TempDir()

		entries, err := os.ReadDir(fixturesDir)
		Expect(err).ToNot(HaveOccurred())

		for _, entry := range entries {
			if filepath.Ext(entry.Name()) != ".go" {
				continue
			}

			data, readErr := os.ReadFile(filepath.Join(fixturesDir, entry.Name()))
			Expect(readErr).ToNot(HaveOccurred())

			Expect(os.WriteFile(filepath.Join(tmpDir, entry.Name()), data, 0o644)).To(Succeed())
		}
	})

	runJSON := func(threshold int, actionability bool) printer.JSONOutput {
		args := []string{tmpDir, "--threshold", strconv.Itoa(threshold), "--json"}
		if !actionability {
			args = append(args, "--no-actionability")
		}

		output_bytes, err := setup.Executor(args...)
		Expect(err).ToNot(HaveOccurred(), "output: %s", output_bytes)

		var output printer.JSONOutput
		Expect(json.Unmarshal(output_bytes, &output)).To(Succeed())

		return output
	}

	groupHasFile := func(group printer.CloneGroup, filePart string) bool {
		for _, clone := range group.Clones {
			if strings.Contains(clone.Filename, filePart) {
				return true
			}
		}

		return false
	}

	Describe("raw clone detection (no actionability)", func() {
		It("should detect clone groups at threshold 1", func() {
			output := runJSON(1, false)
			Expect(output.Summary.TotalCloneGroups).To(BeNumerically(">", 0),
				"expected at least some raw clone groups in the fixture corpus")
		})

		It("should detect error guard patterns from HTTP handlers", func() {
			output := runJSON(1, false)
			found := false

			for _, group := range output.CloneGroups {
				if groupHasFile(group, "http_handlers") {
					found = true

					break
				}
			}

			Expect(found).To(BeTrue(), "expected HTTP handler error guard clones")
		})

		It("should detect queryError wrapping patterns from DB layer", func() {
			output := runJSON(1, false)
			found := false

			for _, group := range output.CloneGroups {
				if groupHasFile(group, "db_queries") {
					found = true

					break
				}
			}

			Expect(found).To(BeTrue(), "expected queryError wrapping clones")
		})

		It("should detect defer cleanup patterns", func() {
			output := runJSON(1, false)
			found := false

			for _, group := range output.CloneGroups {
				if groupHasFile(group, "cleanup") {
					found = true

					break
				}
			}

			Expect(found).To(BeTrue(), "expected defer cleanup clones")
		})
	})

	Describe("actionability suppression (semantic mode)", func() {
		It("should suppress HTTP error guards via error-propagation", func() {
			output := runJSON(1, true)

			for _, group := range output.CloneGroups {
				Expect(groupHasFile(group, "http_handlers")).To(BeFalse(),
					"HTTP error guard should be suppressed, found in group hash=%s", group.Hash)
			}
		})

		It("should suppress queryError wrapping via error-wrapping (structural)", func() {
			output := runJSON(1, true)

			for _, group := range output.CloneGroups {
				Expect(groupHasFile(group, "db_queries")).To(BeFalse(),
					"queryError wrapping should be suppressed, found in group hash=%s", group.Hash)
			}
		})

		It("should suppress FuncLit defer cleanup (rows.Close, tx.Rollback)", func() {
			output := runJSON(1, true)

			// The FuncLit defer variants (rows.Close, tx.Rollback) are suppressed
			// by raii-defer. The bare ctx+cancel variant may survive as a
			// 2-statement clone (known limitation).
			for _, group := range output.CloneGroups {
				for _, clone := range group.Clones {
					lineStart := clone.LineStart
					// FuncLit defer patterns start at lines 27+ in cleanup.go
					if strings.Contains(clone.Filename, "cleanup") && lineStart >= 27 {
						Fail(fmt.Sprintf("FuncLit defer cleanup should be suppressed: %s:%d",
							clone.Filename, lineStart))
					}
				}
			}
		})

		It("should suppress bool-to-string guard clauses", func() {
			output := runJSON(1, true)

			for _, group := range output.CloneGroups {
				Expect(groupHasFile(group, "bool_strings")).To(BeFalse(),
					"bool-to-string guard clauses should be suppressed, found in group hash=%s", group.Hash)
			}
		})

		It("should suppress logging one-liners via error-propagation", func() {
			output := runJSON(1, true)

			for _, group := range output.CloneGroups {
				Expect(groupHasFile(group, "logging")).To(BeFalse(),
					"logging one-liners should be suppressed, found in group hash=%s", group.Hash)
			}
		})
	})
})
