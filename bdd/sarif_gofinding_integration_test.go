package bdd

import (
	"context"
	"encoding/json/v2"
	"regexp"
	"sort"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	gofinding "github.com/larsartmann/go-finding"
)

// Issue #2 (go-finding integration): the committed tests only round-tripped
// SARIF bytes that printer/finding's own ToSARIF produced. This spec feeds
// REAL CLI --sarif bytes through go-finding's FindingsFromSARIF and asserts
// the exact clone groups come back — the same tree, re-run through --json,
// must reconstruct the identical group set (GroupID = clone-group content
// hash, deterministic across channels and runs).
// multiCloneSnippets returns three distinct clone families (cache, queue,
// formatter), each duplicated across three files, so the SARIF partition
// assertions exercise multiple groups with 3 occurrences each.
func multiCloneSnippets() map[string]string {
	return map[string]string{
		"multiCache": `package multi

import "sync"

type cacheStore struct {
	mu    sync.Mutex
	items map[string]int
}

func (c *cacheStore) load(key string) (int, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	value, ok := c.items[key]
	return value, ok
}
`,
		"multiQueue": `package multi

import "errors"

var errEmpty = errors.New("queue empty")

type queue struct {
	items []string
}

func (q *queue) pop() (string, error) {
	if len(q.items) == 0 {
		return "", errEmpty
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item, nil
}
`,
		"multiFormat": `package multi

import "fmt"

func render(label string, count int) string {
	header := fmt.Sprintf("%s: %d items", label, count)
	footer := fmt.Sprintf("%s: done", label)
	return header + "\n" + footer
}
`,
	}
}

// createMultiCloneTree writes the three clone families to the setup tree.
func createMultiCloneTree(setup *testutil.BDDTestSetup) {
	for prefix, code := range multiCloneSnippets() {
		err := setup.CreateDuplicateFiles(
			[]string{prefix + "1.go", prefix + "2.go", prefix + "3.go"},
			code,
		)
		Expect(err).NotTo(HaveOccurred())
	}
}

// sarifGroupPartition runs the CLI in SARIF mode (plus any extra args) and
// reconstructs the clone-group partition from go-finding: group hash →
// sorted occurrence file list.
func sarifGroupPartition(setup *testutil.BDDTestSetup, extraArgs ...string) map[string][]string {
	args := append([]string{"--sarif", "--threshold", "1", "--no-actionability"}, extraArgs...)

	sarifBytes, err := setup.RunArtDupl(args...)
	Expect(err).ToNot(HaveOccurred())

	findings, err := gofinding.FindingsFromSARIF(context.Background(), sarifBytes)
	Expect(err).ToNot(HaveOccurred())
	Expect(findings).NotTo(BeEmpty())

	report := gofinding.NewReportFromFindings(
		gofinding.ToolInfo{Name: "art-dupl"},
		findings,
	)

	partition := map[string][]string{}

	for id, groupFindings := range report.GroupFindings() {
		files := make([]string, 0, len(groupFindings))
		for _, f := range groupFindings {
			files = append(files, string(f.Position.File))
		}

		sort.Strings(files)
		partition[string(id)] = files
	}

	return partition
}

var _ = Describe("SARIF to go-finding integration", func() {
	It("reconstructs exact clone groups from real CLI SARIF output", func() {
		setup := CreateBDDTestSetup()

		err := setup.CreateDuplicateFiles([]string{"gofinding1.go", "gofinding2.go"}, duplicateCode)
		Expect(err).NotTo(HaveOccurred())

		sarifBytes, err := setup.RunArtDupl("--sarif", "--threshold", "1", "--no-actionability")
		Expect(err).ToNot(HaveOccurred())

		findings, err := gofinding.FindingsFromSARIF(context.Background(), sarifBytes)
		Expect(err).ToNot(HaveOccurred())
		Expect(findings).NotTo(BeEmpty())

		report := gofinding.NewReportFromFindings(
			gofinding.ToolInfo{Name: "art-dupl"},
			findings,
		)
		byGroup := report.GroupFindings()
		Expect(byGroup).NotTo(BeEmpty())

		groupIDPattern := regexp.MustCompile(`^[0-9a-f]{16}$`)
		for id, groupFindings := range byGroup {
			Expect(string(id)).To(MatchRegexp(groupIDPattern.String()),
				"GroupID must be the 16-char lowercase hex clone-group hash")

			// Every SARIF result is one clone occurrence; a clone group by
			// definition has at least two.
			Expect(len(groupFindings)).To(BeNumerically(">=", 2),
				"group %s reconstructed with fewer occurrences than exist", id)
		}

		jsonBytes, err := setup.RunArtDupl("--json", "--threshold", "1", "--no-actionability")
		Expect(err).ToNot(HaveOccurred())

		var output struct {
			CloneGroups []struct {
				Hash string `json:"hash"`
			} `json:"clone_groups"`
		}
		Expect(json.Unmarshal(jsonBytes, &output)).To(Succeed())
		Expect(output.CloneGroups).NotTo(BeEmpty())

		jsonHashes := map[string]bool{}
		for _, group := range output.CloneGroups {
			jsonHashes[group.Hash] = true
		}

		sarifHashes := map[string]bool{}
		for id := range byGroup {
			sarifHashes[string(id)] = true
		}

		Expect(sarifHashes).To(Equal(jsonHashes),
			"SARIF and JSON channels must surface the identical group set for the same tree")
	})

	It("reconstructs the exact multi-clone partition with sibling links", func() {
		setup := CreateBDDTestSetup()

		createMultiCloneTree(setup)

		sarifBytes, err := setup.RunArtDupl("--sarif", "--threshold", "1", "--no-actionability")
		Expect(err).ToNot(HaveOccurred())

		findings, err := gofinding.FindingsFromSARIF(context.Background(), sarifBytes)
		Expect(err).ToNot(HaveOccurred())
		Expect(findings).NotTo(BeEmpty())

		report := gofinding.NewReportFromFindings(
			gofinding.ToolInfo{Name: "art-dupl"},
			findings,
		)
		byGroup := report.GroupFindings()
		Expect(byGroup).NotTo(BeEmpty())

		jsonBytes, err := setup.RunArtDupl("--json", "--threshold", "1", "--no-actionability")
		Expect(err).ToNot(HaveOccurred())

		var output struct {
			CloneGroups []struct {
				Hash  string `json:"hash"`
				Files []struct {
					Filename string `json:"filename"`
				} `json:"files"`
			} `json:"clone_groups"`
		}
		Expect(json.Unmarshal(jsonBytes, &output)).To(Succeed())
		Expect(output.CloneGroups).NotTo(BeEmpty())

		jsonGroups := map[string][]string{}
		for _, group := range output.CloneGroups {
			files := make([]string, 0, len(group.Files))
			for _, f := range group.Files {
				files = append(files, f.Filename)
			}
			jsonGroups[group.Hash] = files
		}

		sarifGroups := map[string]bool{}
		for id := range byGroup {
			sarifGroups[string(id)] = true
		}

		jsonHashes := map[string]bool{}
		for hash := range jsonGroups {
			jsonHashes[hash] = true
		}

		Expect(sarifGroups).To(Equal(jsonHashes),
			"the SARIF round-trip must reconstruct exactly the JSON group set — no split, no merge")

		for id, groupFindings := range byGroup {
			wantFiles, ok := jsonGroups[string(id)]
			Expect(ok).To(BeTrue(), "reconstructed group %s missing from JSON channel", id)

			gotFiles := make([]string, 0, len(groupFindings))
			for _, f := range groupFindings {
				gotFiles = append(gotFiles, string(f.Position.File))
			}

			Expect(gotFiles).To(ConsistOf(wantFiles),
				"group %s occurrences must match the JSON channel exactly", id)

			for _, f := range groupFindings {
				Expect(f.Related).To(HaveLen(len(wantFiles)-1),
					"group %s: every occurrence must link its %d siblings", id, len(wantFiles)-1)

				for _, rel := range f.Related {
					Expect(string(rel.Position.File)).To(BeElementOf(wantFiles),
						"group %s: related location must point at a sibling occurrence", id)
				}
			}
		}
	})

	It("reconstructs the identical partition under --search-workers 1 (determinism twin)", func() {
		setup := CreateBDDTestSetup()

		createMultiCloneTree(setup)

		parallel := sarifGroupPartition(setup)
		sequential := sarifGroupPartition(setup, "--search-workers", "1")

		Expect(parallel).NotTo(BeEmpty())
		Expect(sequential).To(Equal(parallel),
			"the reconstructed clone-group partition must be identical across parallel and sequential search")
	})
})
