// art-dupl is a fast, type-safe code duplication detector for Go projects.
//
// FEATURES:
// - Fast suffix tree algorithm for syntax-level clone detection
// - Type-safe domain model with validation at construction
// - Multiple detection methods: syntax-level, hash-based, TODO comments, legacy patterns
// - Multiple output formats: text, HTML, JSON, plumbing
// - Comprehensive statistics: duplication ratio, health score, complexity metrics
// - SIMD-optimized performance for large codebases
//
// ARCHITECTURE:
// - domain/: Value objects and entities (LineNumber, Threshold, Clone, CloneGroup, Analysis)
// - syntax/: Unified AST representation for language-agnostic processing
// - suffixtree/: Suffix tree data structure for efficient duplicate searching
// - detection/: Multi-method detection coordination (art-dupl, hash, todos, legacy)
// - config/: Configuration management with type-safe enums
// - errors/: Rich error types with context and wrapping
// - printer/: Output formatting with multiple formats
//
// TYPE SAFETY:
// - Domain types prevent accidental type mismatches (e.g., LineNumber vs int)
// - Validation enforced at construction (domain.NewThreshold(), etc.)
// - Typed marshaling functions for JSON (SafeMarshalConfig, SafeMarshalClone, etc.)
// - Idiomatic Go error handling: (T, error) returns with rich error context
//
// PERFORMANCE:
// - O(n) suffix tree construction where n = sequence length
// - O(n) duplicate search with typical code
// - SIMD-optimized transition search for >8 transitions
// - Memory-optimized node layout (40B per node, 37.5% reduction)
// - String interning for duplicate strings (StringInternPool)
//
// USAGE:
//
//	// As CLI
//	$ art-dupl ./... --threshold 15 --json
//
//	// As SDK (Go API)
//	import "github.com/LarsArtmann/art-dupl/pkg/artdupl"
//
//	opts := &artdupl.Options{
//	    Threshold: 15,
//	    DetectionMethods: []artdupl.DetectionMethod{artdupl.MethodArtDupl},
//	}
//
//	detector, err := artdupl.NewDetector(opts)
//	if err != nil { ... }
//
//	ctx := context.Background()
//	result, err := detector.FindClones(ctx, []string{"./"})
//	if err != nil { ... }
//
//	fmt.Printf("Found %d clone groups\n", len(result.CloneGroups))
//
// GETTING STARTED:
// - See README.md for detailed documentation
// - See docs/ directory for architecture decisions
// - See examples/ directory for usage examples
// - See domain/domain_types.go for available domain types
//
// PACKAGE ORGANIZATION:
// - domain/: Core domain model (value objects, entities, enums)
// - syntax/: Unified AST representation and processing
// - suffixtree/: Suffix tree data structure and search
// - detection/: Multi-method detection coordination
// - config/: Configuration and validation
// - errors/: Error types and error handling utilities
// - printer/: Output formatting and statistics
// - cmd/: CLI application
// - pkg/artdupl/: SDK for programmatic use
//
// CONTRIBUTING:
// - See CONTRIBUTING.md for guidelines
// - Run tests with: go test ./...
// - Run benchmarks with: go test -bench ./...
//
// LICENSE: MIT
module github.com/LarsArtmann/art-dupl

go 1.26.1

require (
	github.com/a-h/templ v0.3.1001
	github.com/charmbracelet/fang v1.0.0
	github.com/charmbracelet/lipgloss v1.1.0
	github.com/charmbracelet/log v1.0.0
	github.com/charmbracelet/x/exp/golden v0.0.0-20260323091123-df7b1bcffcca
	github.com/go-faster/yaml v0.4.6
	github.com/onsi/ginkgo/v2 v2.28.1
	github.com/onsi/gomega v1.39.0
	github.com/sergi/go-diff v1.4.0
	github.com/spf13/cobra v1.10.2
	github.com/zeebo/xxh3 v1.1.0
)

require (
	charm.land/lipgloss/v2 v2.0.2 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/a-h/parse v0.0.0-20250122154542-74294addb73e // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/aymanbagabas/go-udiff v0.4.1 // indirect
	github.com/charmbracelet/colorprofile v0.4.3 // indirect
	github.com/charmbracelet/ultraviolet v0.0.0-20260316091819-b93f6a3b8502 // indirect
	github.com/charmbracelet/x/ansi v0.11.6 // indirect
	github.com/charmbracelet/x/cellbuf v0.0.15 // indirect
	github.com/charmbracelet/x/exp/charmtone v0.0.0-20260323091123-df7b1bcffcca // indirect
	github.com/charmbracelet/x/term v0.2.2 // indirect
	github.com/charmbracelet/x/termios v0.1.1 // indirect
	github.com/charmbracelet/x/windows v0.2.2 // indirect
	github.com/clipperhouse/displaywidth v0.11.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/go-faster/errors v0.7.1 // indirect
	github.com/go-faster/jx v1.2.0 // indirect
	github.com/go-logfmt/logfmt v0.6.1 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20260302011040-a15ffb7f9dcc // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.3.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.21 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/mango v0.2.0 // indirect
	github.com/muesli/mango-cobra v1.3.0 // indirect
	github.com/muesli/mango-pflag v0.2.0 // indirect
	github.com/muesli/roff v0.1.0 // indirect
	github.com/muesli/termenv v0.16.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/exp v0.0.0-20260312153236-7ab1446f8b90 // indirect
	golang.org/x/mod v0.34.0 // indirect
	golang.org/x/net v0.52.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.35.0 // indirect
	golang.org/x/tools v0.43.0 // indirect
)
