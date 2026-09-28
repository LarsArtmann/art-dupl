package bdd

import (
	"context"
	"encoding/json/v2"
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	gofinding "github.com/larsartmann/go-finding"
)

// Issue #2 (go-finding integration): the committed tests only round-tripped
// SARIF bytes that printer/finding's own ToSARIF produced. This spec feeds
// REAL CLI --sarif bytes through go-finding's FindingsFromSARIF and asserts
// the exact clone groups come back — the same tree, re-run through --json,
// must reconstruct the identical group set (GroupID = clone-group content
// hash, deterministic across channels and runs).
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
})
