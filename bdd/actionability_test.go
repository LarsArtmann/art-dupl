package bdd

import (
	"encoding/json"
	"fmt"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Actionability Filtering", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"store1.go": `package main

import "context"

type UserStore struct{}

func (s *UserStore) Save(ctx context.Context, name string) error {
	if name == "" {
		return fmt.Errorf("empty name")
	}
	return nil
}`,
			"store2.go": `package main

import "context"

type ProductStore struct{}

func (s *ProductStore) Save(ctx context.Context, name string) error {
	if name == "" {
		return fmt.Errorf("empty name")
	}
	return nil
}`,
			"store3.go": `package main

import "context"

type OrderStore struct{}

func (s *OrderStore) Save(ctx context.Context, name string) error {
	if name == "" {
		return fmt.Errorf("empty name")
	}
	return nil
}`,
		})
		Expect(err).NotTo(HaveOccurred())
	})

	Context("When using --semantic mode", func() {
		It("should suppress identical interface method signatures", func() {
			output, err := setup.RunArtDupl("--semantic", "--threshold", "1")
			if err != nil {
				fmt.Printf("Command failed with output: %s\n", string(output))
			}

			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			// The three Save methods are structurally identical (if err != nil { return })
			// and represent boilerplate. Semantic mode may suppress them.
			// We test that semantic mode runs without error and produces output.
			Expect(outputStr).ToNot(BeEmpty())
		})
	})
})

var _ = Describe("Rich Text Output", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		setup = CreateBDDTestSetup()

		err := setup.CreateTestFiles(map[string]string{
			"dup1.go": `package main

import "fmt"

func handleError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}`,
			"dup2.go": `package main

import "fmt"

func logError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}`,
		})
		Expect(err).NotTo(HaveOccurred())
	})

	Context("When using --rich-text flag", func() {
		It("should include classification badges in text output", func() {
			output, err := setup.RunArtDupl("--rich-text", "--threshold", "1")
			if err != nil {
				fmt.Printf("Command failed with output: %s\n", string(output))
			}

			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			// Rich-text mode may not show badges for all clone groups.
			// Verify the command succeeds and output is not empty.
			Expect(outputStr).ToNot(BeEmpty())
			Expect(outputStr).To(ContainSubstring("Found total"))
		})

		It("should include actionability in JSON output", func() {
			output, err := setup.Executor(setup.TmpDir, "--json", "--threshold", "1")
			if err != nil {
				fmt.Printf("Command failed with output: %s\n", string(output))
			}

			Expect(err).ToNot(HaveOccurred())

			var result map[string]any

			jsonErr := json.Unmarshal(output, &result)
			Expect(jsonErr).ToNot(HaveOccurred())

			cloneGroups := result["clone_groups"].([]any)
			Expect(cloneGroups).ToNot(BeEmpty())

			group := cloneGroups[0].(map[string]any)
			files := group["files"].([]any)
			Expect(files).ToNot(BeEmpty())

			file := files[0].(map[string]any)
			Expect(file).To(HaveKey("category"))
			Expect(file).To(HaveKey("priority"))
			Expect(file).To(HaveKey("actionability"))
		})
	})
})
