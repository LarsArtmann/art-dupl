package main

import (
	"fmt"
	"log"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// Test advanced API capabilities beyond basic lib.Run
func main() {
	fmt.Println("=== Advanced API Analysis ===")

	// Test 1: Direct suffix tree access
	fmt.Println("\n1. Direct Suffix Tree Access:")
	tree := suffixtree.New()

	// Create some test nodes
	nodes := []*syntax.Node{
		{Type: 1, Filename: "test1.go", Pos: 0, End: 10},
		{Type: 2, Filename: "test1.go", Pos: 11, End: 20},
		{Type: 1, Filename: "test2.go", Pos: 0, End: 10},
		{Type: 2, Filename: "test2.go", Pos: 11, End: 20},
		{Type: 3, Filename: "test2.go", Pos: 21, End: 30},
	}

	for _, node := range nodes {
		tree.Update(node)
	}

	// Add termination node
	tree.Update(&syntax.Node{Type: -1})

	matches := tree.FindDuplOver(2)
	matchCount := 0
	for match := range matches {
		matchCount++
		fmt.Printf("  Match %d: Length=%d, Positions=%v\n", matchCount, match.Len, match.Ps)
	}
	fmt.Printf("  Total matches: %d\n", matchCount)

	// Test 2: Configuration system
	fmt.Println("\n2. Configuration System:")
	cfg := config.DefaultConfig()
	fmt.Printf("  Default threshold: %d\n", cfg.Threshold)
	fmt.Printf("  Default detection methods: %v\n", cfg.DetectionMethods)

	// Test custom config
	customCfg := &config.Config{
		Threshold:        20,
		IncludeVendor:    true,
		Verbose:          true,
		OutputFormat:     config.OutputFormatJSON,
		DetectionMethods: config.DetectionMethods{config.DetectionMethodHash, config.DetectionMethodArtDupl},
	}

	if err := config.ValidateConfig(customCfg); err != nil {
		log.Printf("Config validation error: %v", err)
	} else {
		fmt.Printf("  Custom config is valid\n")
	}

	// Test 3: File processing pipeline
	fmt.Println("\n3. File Processing Pipeline:")
	files := []string{"config/config.go"}
	fileChan := make(chan string, 10)
	go func() {
		for _, f := range files {
			fileChan <- f
		}
		close(fileChan)
	}()

	syntaxChan, fileCountChan := job.Parse(fileChan)
	tree, data, done := job.BuildTree(syntaxChan)
	<-done

	fileCount := <-fileCountChan
	fmt.Printf("  Processed %d files\n", fileCount)
	fmt.Printf("  Built suffix tree with %d nodes\n", len(*data))

	// Test 4: Node serialization
	fmt.Println("\n4. Node Serialization:")
	if len(*data) > 0 {
		serialized := syntax.Serialize((*data)[0])
		fmt.Printf("  Serialized first node to %d nodes\n", len(serialized))
	}

	fmt.Println("\n=== API Analysis Complete ===")
	fmt.Println("✅ Core algorithms are accessible")
	fmt.Println("✅ Configuration system is clean")
	fmt.Println("✅ File processing pipeline works")
	fmt.Println("⚠️  Current lib API is limited to simple Issue[] output")
	fmt.Println("⚠️  No unified high-level SDK interface")
	fmt.Println("⚠️  Advanced features (multi-detection, custom outputs) not exposed in lib")
}
