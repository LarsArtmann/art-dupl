package bdd

import (
	"fmt"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// End-to-end specs for the two collector-idiom actionability patterns
// (append-collector-tail incl. the deep-copy shape, option-apply-loop).
// The unit tests in printer/actionability cover matcher logic on hand-built
// and ParseSnippet trees; these specs prove the suppression survives the
// full CLI pipeline and that --explain names the pattern.

var _ = Describe("append-collector-tail pattern (e2e)", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"collect_docs.go": `package main

func copyDocuments(docs []string) []string {
	out := make([]string, 0, len(docs))
	for _, doc := range docs {
		if doc != "" {
			out = append(out, doc)
		}
	}
	return out
}
`,
			"collect_names.go": `package main

func copyNames(names []string) []string {
	out := make([]string, 0, len(names))
	for _, name := range names {
		if name != "" {
			out = append(out, name)
		}
	}
	return out
}
`,
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should suppress collector-tail clones by default", func() {
		output, err := setup.RunArtDupl("--quiet", "--semantic", "--threshold", "2")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(output)).To(ContainSubstring("0 shown"),
			"collector-tail clones should be suppressed as non-actionable")
	})

	It("should surface them with --show-suppressed and name the pattern via --explain", func() {
		output, err := setup.RunArtDupl(
			"--quiet", "--semantic", "--threshold", "2", "--show-suppressed", "--explain",
		)
		if err != nil {
			fmt.Printf("Command failed with output: %s\n", string(output))
		}

		Expect(err).ToNot(HaveOccurred())
		Expect(string(output)).To(ContainSubstring("append-collector-tail"),
			"--explain should attribute the suppression to the collector-tail pattern")
	})
})

var _ = Describe("deep-copy collector loop (e2e)", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"copy_docs.go": `package main

import "slices"

type document struct{ Payload []byte }

func allDocuments(docs []document) []document {
	out := make([]document, 0, len(docs))
	for _, doc := range docs {
		copied := doc
		copied.Payload = slices.Clone(doc.Payload)
		out = append(out, copied)
	}
	return out
}
`,
			"copy_tasks.go": `package main

import "slices"

type task struct{ Payload []byte }

func allTasks(tasks []task) []task {
	out := make([]task, 0, len(tasks))
	for _, item := range tasks {
		copied := item
		copied.Payload = slices.Clone(item.Payload)
		out = append(out, copied)
	}
	return out
}
`,
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should suppress deep-copy collector loops", func() {
		output, err := setup.RunArtDupl("--quiet", "--semantic", "--threshold", "2")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(output)).To(ContainSubstring("0 shown"),
			"deep-copy loops are per-type idioms, not actionable duplication")
	})

	It("should attribute the suppression to append-collector-tail", func() {
		output, err := setup.RunArtDupl(
			"--quiet", "--semantic", "--threshold", "2", "--show-suppressed", "--explain",
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(output)).To(ContainSubstring("append-collector-tail"))
	})
})

var _ = Describe("option-apply-loop pattern (e2e)", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"server.go": `package main

type server struct{ name string }

type serverOption func(*server)

func newServer(opts ...serverOption) *server {
	s := &server{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}
`,
			"client.go": `package main

type client struct{ name string }

type clientOption func(*client)

func newClient(opts ...clientOption) *client {
	c := &client{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}
`,
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should suppress bare option-apply constructor clones", func() {
		output, err := setup.RunArtDupl("--quiet", "--semantic", "--threshold", "2")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(output)).To(ContainSubstring("0 shown"),
			"bare option-apply constructors are the standard options idiom")
	})

	It("should attribute the suppression to option-apply-loop", func() {
		output, err := setup.RunArtDupl(
			"--quiet", "--semantic", "--threshold", "2", "--show-suppressed", "--explain",
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(output)).To(ContainSubstring("option-apply-loop"))
	})
})

var _ = Describe("collector pattern negative controls (e2e)", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"transform_a.go": `package main

func normalizeAll(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, strings.ToUpper(value))
	}
	return out
}
`,
			"transform_b.go": `package main

func encodeAll(inputs []string) []string {
	out := make([]string, 0, len(inputs))
	for _, input := range inputs {
		out = append(out, encode(input))
	}
	return out
}
`,
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should NOT suppress loops with real transform calls", func() {
		output, err := setup.RunArtDupl(
			"--quiet", "--semantic", "--threshold", "2", "--show-suppressed", "--explain",
		)
		if err != nil {
			fmt.Printf("Command failed with output: %s\n", string(output))
		}

		Expect(err).ToNot(HaveOccurred())
		outputStr := string(output)
		Expect(outputStr).ToNot(ContainSubstring("append-collector-tail"),
			"transform loops contain real logic and must stay actionable")
		Expect(outputStr).ToNot(ContainSubstring("option-apply-loop"))
	})
})
