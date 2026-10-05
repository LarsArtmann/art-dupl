package bdd

import (
	"encoding/json"
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	gofinding "github.com/larsartmann/go-finding"
)

// Plan T05 (M24): the --lsp channel must carry the same clone groups as the
// other machine formats. Real CLI --lsp bytes are parsed as per-file
// publishDiagnostics-style documents; every diagnostic round-trips through
// go-finding's FromLSP keeping its GroupID, and the GroupID set must equal
// the --json channel's clone-group hash set for the same tree.
var _ = Describe("LSP output format", func() {
	It("emits per-file diagnostics whose GroupIDs match the JSON channel", func() {
		setup := CreateBDDTestSetup()

		err := setup.CreateDuplicateFiles([]string{"lsp1.go", "lsp2.go"}, duplicateCode)
		Expect(err).NotTo(HaveOccurred())

		lspBytes, err := setup.RunArtDupl("--lsp", "--threshold", "1", "--no-actionability")
		Expect(err).ToNot(HaveOccurred())

		var documents []struct {
			URI         string                    `json:"uri"`
			Diagnostics []gofinding.LSPDiagnostic `json:"diagnostics"`
		}
		Expect(json.Unmarshal(lspBytes, &documents)).To(Succeed())
		Expect(documents).NotTo(BeEmpty())

		groupIDPattern := regexp.MustCompile(`^[0-9a-f]{16}$`)
		lspHashes := map[string]bool{}

		for _, doc := range documents {
			Expect(doc.URI).NotTo(BeEmpty())
			Expect(doc.Diagnostics).NotTo(BeEmpty())

			for _, diag := range doc.Diagnostics {
				Expect(diag.Source).To(Equal("art-dupl"))
				Expect(diag.Code).To(Equal("art-dupl/duplicate-code"))

				Expect(diag.Data).NotTo(BeNil(),
					"every diagnostic must carry go-finding round-trip data")
				Expect(string(diag.Data.GroupID)).To(MatchRegexp(groupIDPattern.String()),
					"GroupID must be the 16-char lowercase hex clone-group hash")

				restored := gofinding.FromLSP(gofinding.FilePath(doc.URI), diag)
				Expect(restored.GroupID).To(Equal(diag.Data.GroupID),
					"FromLSP must preserve the GroupID (G6)")
				Expect(restored.Position.File).To(Equal(gofinding.FilePath(doc.URI)))
				Expect(restored.Position.Line).To(BeNumerically(">", 0))

				lspHashes[string(diag.Data.GroupID)] = true
			}
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

		Expect(lspHashes).To(Equal(jsonHashes),
			"LSP and JSON channels must surface the identical group set for the same tree")
	})
})
