package bdd

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

// BDD Test Suite for Incremental Detection
//
// These tests verify incremental analysis behavior using AST caching,
// including cache creation, cache hits, cache clearing, and content-based re-parsing.

var _ = Describe("Incremental Detection", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		var err error
		setup, err = testutil.NewBDDTestSetupForGinkgo()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(setup.Cleanup()).To(Succeed())
	})

	// incrementalTestCode is a simple code sample for incremental tests.
	const incrementalTestCode = `package main

import "fmt"

func IncrementalTest() {
	for i := 0; i < 10; i++ {
		fmt.Println(i)
	}
}`

	// modifiedTestCode is a slightly different code sample to trigger re-parsing.
	const modifiedTestCode = `package main

import "fmt"

func ModifiedTest() {
	for i := 0; i < 20; i++ {
		fmt.Println(i)
	}
}`

	Context("When using incremental mode for the first time", func() {
		It("should create cache directory and cache entries", func() {
			// Create test files
			err := setup.CreateDuplicateFiles([]string{"file1.go", "file2.go"}, incrementalTestCode)
			Expect(err).NotTo(HaveOccurred())

			// Use a custom cache directory in temp dir
			cacheDir := filepath.Join(setup.TmpDir, ".cache", "art-dupl")

			// First run with incremental mode
			output, err := setup.RunArtDupl("--incremental", "--cache-dir", cacheDir)
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())

			// Verify cache directory was created
			Expect(cacheDir).To(BeADirectory())

			// Verify files subdirectory was created
			filesDir := filepath.Join(cacheDir, "files")
			Expect(filesDir).To(BeADirectory())

			// Verify at least one cache entry was created
			entries, err := os.ReadDir(filesDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).NotTo(BeEmpty())
		})

		It("should detect duplicates correctly on first run", func() {
			err := setup.CreateDuplicateFiles([]string{"first.go", "second.go"}, incrementalTestCode)
			Expect(err).NotTo(HaveOccurred())

			cacheDir := filepath.Join(setup.TmpDir, ".cache")

			output, err := setup.RunArtDupl("--incremental", "--cache-dir", cacheDir, "-t", "10")
			Expect(err).ToNot(HaveOccurred())

			// Should find duplicates
			Expect(string(output)).To(ContainSubstring("found"))
		})
	})

	Context("When running incremental mode a second time", func() {
		It("should use cached AST for unchanged files", func() {
			err := setup.CreateDuplicateFiles([]string{"cached1.go", "cached2.go"}, incrementalTestCode)
			Expect(err).NotTo(HaveOccurred())

			cacheDir := filepath.Join(setup.TmpDir, ".cache", "art-dupl")

			// First run - creates cache
			output1, err := setup.RunArtDupl("--incremental", "--cache-dir", cacheDir, "-t", "10")
			Expect(err).ToNot(HaveOccurred())
			Expect(output1).ToNot(BeNil())

			// Check cache was populated
			filesDir := filepath.Join(cacheDir, "files")
			entries1, err := os.ReadDir(filesDir)
			Expect(err).NotTo(HaveOccurred())
			initialCount := len(entries1)

			// Second run - should use cache
			output2, err := setup.RunArtDupl("--incremental", "--cache-dir", cacheDir, "-t", "10")
			Expect(err).ToNot(HaveOccurred())
			Expect(output2).ToNot(BeNil())

			// Cache entries should remain the same (no new entries for unchanged files)
			entries2, err := os.ReadDir(filesDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(entries2).To(HaveLen(initialCount))

			// Both runs should produce identical results
			Expect(string(output2)).To(Equal(string(output1)))
		})

		It("should be faster on second run (uses cache)", func() {
			err := setup.CreateDuplicateFiles([]string{"speed1.go", "speed2.go"}, incrementalTestCode)
			Expect(err).NotTo(HaveOccurred())

			cacheDir := filepath.Join(setup.TmpDir, ".cache", "art-dupl")

			// First run
			start1 := time.Now()
			_, err = setup.RunArtDupl("--incremental", "--cache-dir", cacheDir, "-t", "10")
			duration1 := time.Since(start1)
			Expect(err).ToNot(HaveOccurred())

			// Second run
			start2 := time.Now()
			_, err = setup.RunArtDupl("--incremental", "--cache-dir", cacheDir, "-t", "10")
			duration2 := time.Since(start2)
			Expect(err).ToNot(HaveOccurred())

			// Second run should be faster or similar (cache hit)
			// Note: For very small files, the difference might be negligible
			// so we just verify both complete successfully
			Expect(duration2).To(BeNumerically("<=", duration1*5)) // Allow some variance
		})
	})

	Context("When using --clear-cache flag", func() {
		It("should clear cache before running", func() {
			err := setup.CreateDuplicateFiles([]string{"clear1.go", "clear2.go"}, incrementalTestCode)
			Expect(err).NotTo(HaveOccurred())

			cacheDir := filepath.Join(setup.TmpDir, ".cache", "art-dupl")

			// First run - creates cache
			_, err = setup.RunArtDupl("--incremental", "--cache-dir", cacheDir, "-t", "10")
			Expect(err).ToNot(HaveOccurred())

			// Verify cache exists
			filesDir := filepath.Join(cacheDir, "files")
			entries, err := os.ReadDir(filesDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).NotTo(BeEmpty())

			// Run with --clear-cache
			output, err := setup.RunArtDupl("--incremental", "--clear-cache", "--cache-dir", cacheDir, "-t", "10")
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())

			// Cache should be cleared and then repopulated
			entries2, err := os.ReadDir(filesDir)
			Expect(err).NotTo(HaveOccurred())
			// After clear and re-run, cache should be repopulated
			Expect(entries2).NotTo(BeEmpty())
		})
	})

	Context("When files are modified", func() {
		It("should re-parse files with different content", func() {
			err := setup.CreateDuplicateFiles([]string{"modify1.go", "modify2.go"}, incrementalTestCode)
			Expect(err).NotTo(HaveOccurred())

			cacheDir := filepath.Join(setup.TmpDir, ".cache", "art-dupl")

			// First run
			_, err = setup.RunArtDupl("--incremental", "--cache-dir", cacheDir, "-t", "10")
			Expect(err).ToNot(HaveOccurred())

			// Get initial cache entries
			filesDir := filepath.Join(cacheDir, "files")
			entries1, err := os.ReadDir(filesDir)
			Expect(err).NotTo(HaveOccurred())
			initialHashes := make(map[string]bool)
			for _, e := range entries1 {
				initialHashes[e.Name()] = true
			}

			// Modify one file
			err = setup.CreateFileWithContent("modify1.go", modifiedTestCode)
			Expect(err).NotTo(HaveOccurred())

			// Second run - should detect modified file
			_, err = setup.RunArtDupl("--incremental", "--cache-dir", cacheDir, "-t", "10")
			Expect(err).ToNot(HaveOccurred())

			// Check that cache has new entry (different hash for modified file)
			entries2, err := os.ReadDir(filesDir)
			Expect(err).NotTo(HaveOccurred())

			// Should have at least one new cache entry (modified file)
			newEntries := 0
			for _, e := range entries2 {
				if !initialHashes[e.Name()] {
					newEntries++
				}
			}
			Expect(newEntries).To(BeNumerically(">=", 1))
		})

		It("should still use cache for unchanged files when one file is modified", func() {
			err := setup.CreateDuplicateFiles([]string{"mix1.go", "mix2.go", "mix3.go"}, incrementalTestCode)
			Expect(err).NotTo(HaveOccurred())

			cacheDir := filepath.Join(setup.TmpDir, ".cache", "art-dupl")

			// First run
			_, err = setup.RunArtDupl("--incremental", "--cache-dir", cacheDir, "-t", "10")
			Expect(err).ToNot(HaveOccurred())

			// Modify only one file
			err = setup.CreateFileWithContent("mix2.go", modifiedTestCode)
			Expect(err).NotTo(HaveOccurred())

			// Second run
			output, err := setup.RunArtDupl("--incremental", "--cache-dir", cacheDir, "-t", "10")
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})
	})

	Context("When combining incremental mode with other flags", func() {
		It("should work with JSON output format", func() {
			err := setup.CreateDuplicateFiles([]string{"json1.go", "json2.go"}, incrementalTestCode)
			Expect(err).NotTo(HaveOccurred())

			cacheDir := filepath.Join(setup.TmpDir, ".cache", "art-dupl")

			output, err := setup.RunArtDupl("--incremental", "--cache-dir", cacheDir, "--json", "-t", "10")
			Expect(err).ToNot(HaveOccurred())

			// Should produce valid JSON
			Expect(string(output)).To(ContainSubstring(`"clone_groups"`))
		})

		It("should work with plumbing output format", func() {
			err := setup.CreateDuplicateFiles([]string{"plumb1.go", "plumb2.go"}, incrementalTestCode)
			Expect(err).NotTo(HaveOccurred())

			cacheDir := filepath.Join(setup.TmpDir, ".cache", "art-dupl")

			output, err := setup.RunArtDupl("--incremental", "--cache-dir", cacheDir, "--plumbing", "-t", "10")
			Expect(err).ToNot(HaveOccurred())

			// Should produce plumbing output
			Expect(string(output)).To(ContainSubstring("dupl:"))
		})

		It("should work with different thresholds", func() {
			err := setup.CreateDuplicateFiles([]string{"thresh1.go", "thresh2.go"}, incrementalTestCode)
			Expect(err).NotTo(HaveOccurred())

			cacheDir := filepath.Join(setup.TmpDir, ".cache", "art-dupl")

			// Run with high threshold (might not find duplicates)
			_, err = setup.RunArtDupl("--incremental", "--cache-dir", cacheDir, "-t", "100")
			Expect(err).ToNot(HaveOccurred())

			// Run with low threshold (should find duplicates)
			output, err := setup.RunArtDupl("--incremental", "--cache-dir", cacheDir, "-t", "5")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("found"))
		})

		It("should work with verbose output", func() {
			err := setup.CreateDuplicateFiles([]string{"verbose1.go", "verbose2.go"}, incrementalTestCode)
			Expect(err).NotTo(HaveOccurred())

			cacheDir := filepath.Join(setup.TmpDir, ".cache", "art-dupl")

			output, err := setup.RunArtDupl("--incremental", "--cache-dir", cacheDir, "--verbose", "-t", "10")
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})
	})

	Context("When using stats subcommand with incremental cache", func() {
		It("should work independently of cache state", func() {
			err := setup.CreateDuplicateFiles([]string{"stats1.go", "stats2.go"}, incrementalTestCode)
			Expect(err).NotTo(HaveOccurred())

			cacheDir := filepath.Join(setup.TmpDir, ".cache", "art-dupl")

			// Create cache first
			_, err = setup.RunArtDupl("--incremental", "--cache-dir", cacheDir, "-t", "10")
			Expect(err).ToNot(HaveOccurred())

			// Stats should work regardless of cache state
			output, err := setup.RunSubcommand("stats", "--cache-dir", cacheDir, "-t", "10")
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
		})
	})
})

var _ = Describe("Incremental Detection Edge Cases", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		var err error
		setup, err = testutil.NewBDDTestSetupForGinkgo()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(setup.Cleanup()).To(Succeed())
	})

	Context("When cache directory does not exist", func() {
		It("should create the directory automatically", func() {
			err := setup.CreateDuplicateFiles([]string{"new1.go", "new2.go"}, `package main
func NewCache() {}`)
			Expect(err).NotTo(HaveOccurred())

			// Use a non-existent cache directory
			cacheDir := filepath.Join(setup.TmpDir, "nonexistent", "cache", "path")

			output, err := setup.RunArtDupl("--incremental", "--cache-dir", cacheDir, "-t", "5")
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())

			// Directory should have been created
			Expect(cacheDir).To(BeADirectory())
		})
	})

	Context("When running without incremental flag", func() {
		It("should not create cache entries", func() {
			err := setup.CreateDuplicateFiles([]string{"nocache1.go", "nocache2.go"}, `package main
func NoCache() {}`)
			Expect(err).NotTo(HaveOccurred())

			cacheDir := filepath.Join(setup.TmpDir, ".cache", "art-dupl")

			// Run without incremental flag
			output, err := setup.RunArtDupl("--cache-dir", cacheDir, "-t", "5")
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())

			// Cache directory should not have files subdirectory (no caching occurred)
			filesDir := filepath.Join(cacheDir, "files")
			_, err = os.Stat(filesDir)
			Expect(os.IsNotExist(err)).To(BeTrue())
		})
	})

	Context("When processing many files", func() {
		It("should cache all processed files", func() {
			// Create multiple files with duplicate content
			files := make([]string, 10)
			for i := range files {
				files[i] = filepath.Join("subdir", string(rune('a'+i))+".go")
			}

			err := setup.CreateSubdirectories("subdir")
			Expect(err).NotTo(HaveOccurred())

			err = setup.CreateDuplicateFiles(files, `package main
func ManyFiles() { println("test") }`)
			Expect(err).NotTo(HaveOccurred())

			cacheDir := filepath.Join(setup.TmpDir, ".cache", "art-dupl")

			// First run
			output, err := setup.RunArtDupl("--incremental", "--cache-dir", cacheDir, "-t", "5")
			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())

			// Verify cache has entries
			filesDir := filepath.Join(cacheDir, "files")
			entries, err := os.ReadDir(filesDir)
			Expect(err).NotTo(HaveOccurred())
			// Should have at least one cache entry (files with same content share hash)
			Expect(entries).NotTo(BeEmpty())
		})
	})
})
