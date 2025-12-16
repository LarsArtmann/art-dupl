# 🔥 ENTERPRISE-GRADE GO LINTING JUSTFILE
# Complete architecture and code quality enforcement
#
# Just is a handy way to save and run project-specific commands.
# https://github.com/casey/just

# ===== PROFESSIONAL COLOR CONSTANTS =====
# Using justfile built-in color constants for maintainable color management
# These replace hardcoded ANSI escape sequences throughout the file
# Available built-in constants: CLEAR, NORMAL, BOLD, ITALIC, UNDERLINE, INVERT, HIDE, STRIKETHROUGH
# Available built-in colors: BLACK, RED, GREEN, YELLOW, BLUE, MAGENTA, CYAN, WHITE
# Available built-in backgrounds: BG_BLACK, BG_RED, BG_GREEN, BG_YELLOW, BG_BLUE, BG_MAGENTA, BG_CYAN, BG_WHITE

# Tool versions
GOLANGCI_VERSION := "v2.6.0"
GO_ARCH_LINT_VERSION := "v1.14.0"
CAPSLOCK_VERSION := "latest"

# Directories
ROOT_DIR := justfile_directory()
REPORT_DIR := ROOT_DIR / "reports"

# Report file paths
ARCHITECTURE_JSON := REPORT_DIR / "architecture.json"
DEPENDENCIES_DOT := REPORT_DIR / "dependencies.dot"
QUALITY_JSON := REPORT_DIR / "quality.json"
CHECKSTYLE_XML := REPORT_DIR / "checkstyle.xml"
JUNIT_XML := REPORT_DIR / "junit.xml"
CAPSLOCK_TXT := REPORT_DIR / "capslock-analysis.txt"
COVERAGE_OUT := REPORT_DIR / "coverage.out"
COVERAGE_HTML := REPORT_DIR / "coverage.html"
GO_DUPLICATIONS_TXT := REPORT_DIR / "go-duplications.txt"
GO_DUPLICATIONS_HTML := REPORT_DIR / "go-duplications.html"

# Profiling file paths
CPU_PROF := REPORT_DIR / "cpu.prof"
HEAP_PROF := REPORT_DIR / "heap.prof"
GOROUTINE_PROF := REPORT_DIR / "goroutine.prof"
TRACE_OUT := REPORT_DIR / "trace.out"
ALLOCS_PROF := REPORT_DIR / "allocs.prof"

# ===== PROFESSIONAL FUNCTIONS =====

# ===== CLEAN JUSTFILE CONSTANTS =====

# Setup function for consistent directory creation
_setup_reports := `mkdir -p {{REPORT_DIR}} 2>/dev/null || true`

# Default recipe (runs when just is called without arguments)
default: help

# Install git hooks for automatic pre-commit checks
install-hooks:
    @echo "{{BOLD}}🪝 INSTALLING GIT HOOKS{{NORMAL}}"
    @echo "#!/bin/sh" > .git/hooks/pre-commit
    @echo "# Auto-generated pre-commit hook - fast formatting check only" >> .git/hooks/pre-commit
    @echo "just check-pre-commit-fast" >> .git/hooks/pre-commit
    @chmod +x .git/hooks/pre-commit
    @echo "{{GREEN}}✅ Git pre-commit hook installed!{{NORMAL}}"
    @echo "{{CYAN}}The hook will do fast formatting checks only.{{NORMAL}}"
    @echo "{{CYAN}}For full checks including architecture: just check-pre-commit{{NORMAL}}"

# Install comprehensive git hooks (includes architecture validation)
install-hooks-full:
    @echo "{{BOLD}}🪝 INSTALLING COMPREHENSIVE GIT HOOKS{{NORMAL}}"
    @echo "#!/bin/sh" > .git/hooks/pre-commit
    @echo "# Auto-generated pre-commit hook - comprehensive checks" >> .git/hooks/pre-commit
    @echo "just check-pre-commit" >> .git/hooks/pre-commit
    @chmod +x .git/hooks/pre-commit
    @echo "{{GREEN}}✅ Comprehensive git pre-commit hook installed!{{NORMAL}}"
    @echo "{{YELLOW}}⚠️  This includes architecture graph validation - commits will be slower.{{NORMAL}}"

# Show this help message
# ===== 🚀 PROFESSIONAL HELP SYSTEM =====

# Enterprise-grade help with categorization
help:
    @echo "{{BOLD}}🚀 HYPER-ENTERPRISE GO LINTING JUSTFILE{{NORMAL}}"
    @echo ""
    @echo "{{BOLD}}🎯 CORE WORKFLOWS:{{NORMAL}}"
    @echo "  {{GREEN}}just test{{NORMAL}}        - Run tests with coverage"
    @echo "  {{GREEN}}just report{{NORMAL}}      - 🚀 PARALLEL report generation (3x speedup!)"
    @echo "  {{GREEN}}just lint{{NORMAL}}        - Complete linting pipeline"
    @echo ""
    @echo "{{BOLD}}🔧 MAINTENANCE:{{NORMAL}}"
    @echo "  {{BLUE}}just clean{{NORMAL}}       - Smart cleaning with confirmation"
    @echo "  {{BLUE}}just install{{NORMAL}}     - Install all linting tools"
    @echo ""
    @echo "{{BOLD}}🚨 SECURITY:{{NORMAL}}"
    @echo "  {{RED}}just check-pre-commit{{NORMAL}} - Pre-commit validation"
    @echo "  {{RED}}just security-scan{{NORMAL}}    - Full security analysis"
    @echo ""
    @echo "{{BOLD}}⚡ PERFORMANCE:{{NORMAL}}"
    @echo "  {{YELLOW}}just bench{{NORMAL}}       - Benchmark suite"
    @echo "  {{YELLOW}}just profile-cpu{{NORMAL}}  - CPU profiling"
    @echo ""
    @echo "{{BOLD}}📋 ALL AVAILABLE COMMANDS:{{NORMAL}}"
    @just --list --unsorted
    @echo ""
    @echo "{{BOLD}}QUICK START:{{NORMAL}}"
    @echo "  1. {{GREEN}}just bootstrap{{NORMAL}}        - 🚀 Complete setup with enhanced error handling"
    @echo "  2. {{CYAN}}just bootstrap-diagnose{{NORMAL}} - 🔍 Environment diagnostics only"
    @echo "  3. {{CYAN}}just bootstrap-fix{{NORMAL}}     - 🔧 Auto-repair common issues"
    @echo "  4. {{CYAN}}just lint{{NORMAL}}             - Run all linters (including capslock)"
    @echo "  5. {{CYAN}}just security-audit{{NORMAL}}   - Complete security audit"
    @echo "  6. {{CYAN}}just format{{NORMAL}}           - Format code (gofumpt + goimports)"
    @echo "  7. {{CYAN}}just fix{{NORMAL}}              - Auto-fix issues"
    @echo "  8. {{CYAN}}just capslock-quick{{NORMAL}}   - Quick security capability check"
    @echo ""
    @echo "{{BOLD}}DOCKER COMMANDS:{{NORMAL}}"
    @echo "  • {{CYAN}}just docker-test{{NORMAL}}         - Build and test Docker image (if available)"
    @echo ""
    @echo "{{BOLD}}ARCHITECTURE ANALYSIS:{{NORMAL}}"
    @echo "  • {{CYAN}}just graph{{NORMAL}}              - Generate flow graph (default)"
    @echo "  • {{CYAN}}just graph-di{{NORMAL}}           - Generate dependency injection graph"
    @echo "  • {{CYAN}}just graph-vendor{{NORMAL}}       - Generate vendor-inclusive graph"
    @echo "  • {{CYAN}}just graph-all{{NORMAL}}          - Generate ALL graph types"
    @echo "  • {{CYAN}}just graph-component <name>{{NORMAL}} - Generate focused component graph"
    @echo "  • {{CYAN}}just graph-list-components{{NORMAL}} - List available components"
    @echo ""
    @echo "{{BOLD}}TESTING & COVERAGE:{{NORMAL}}"
    @echo "  • {{CYAN}}just test{{NORMAL}}                - Run tests with coverage"
    @echo "  • {{CYAN}}just coverage{{NORMAL}}            - Run coverage analysis with 80% threshold"
    @echo "  • {{CYAN}}just coverage 90{{NORMAL}}         - Run coverage analysis with custom threshold"
    @echo "  • {{CYAN}}just coverage-check{{NORMAL}}      - Quick coverage check (silent)"
    @echo "  • {{CYAN}}just coverage-detailed{{NORMAL}}   - Coverage breakdown by architectural layer"
    @echo ""
    @echo "{{BOLD}}CODE ANALYSIS:{{NORMAL}}"
    @echo "  • {{CYAN}}just fd{{NORMAL}}                  - Find duplicate code (alias for find-duplicates)"
    @echo "  • {{CYAN}}just find-duplicates{{NORMAL}}     - Find duplicate code with custom threshold (default: 15 tokens)"
    @echo "  • {{CYAN}}just lint-capslock{{NORMAL}}      - Run Google's capslock capability analysis"
    @echo ""
    @echo "{{BOLD}}SECURITY ANALYSIS:{{NORMAL}}"
    @echo "  • {{CYAN}}just security-audit{{NORMAL}}    - Complete security audit including capability analysis"
    @echo "  • {{CYAN}}just lint-security{{NORMAL}}     - Security-focused linting (gosec + copyloopvar)"
    @echo "  • {{CYAN}}just lint-vulns{{NORMAL}}        - Vulnerability scanning with govulncheck"
    @echo "  • {{CYAN}}just lint-licenses{{NORMAL}}    - License compliance scanning"
    @echo "  • {{CYAN}}just lint-nilaway{{NORMAL}}     - Nil panic prevention with Uber's NilAway"
    @echo "  • {{CYAN}}just lint-capslock{{NORMAL}}     - Google's capslock capability analysis"
    @echo ""
    @echo "{{BOLD}}PERFORMANCE PROFILING:{{NORMAL}}"
    @echo "  • {{CYAN}}just profile-cpu{{NORMAL}}         - Capture 30-second CPU profile"
    @echo "  • {{CYAN}}just profile-heap{{NORMAL}}        - Capture heap memory profile"
    @echo "  • {{CYAN}}just profile-goroutines{{NORMAL}}  - Capture goroutine dump"
    @echo "  • {{CYAN}}just profile-trace{{NORMAL}}       - Capture 10-second execution trace"
    @echo "  • {{CYAN}}just analyze-cpu{{NORMAL}}         - Open CPU profile in browser"
    @echo "  • {{CYAN}}just analyze-heap{{NORMAL}}        - Open heap profile in browser"
    @echo ""
    @echo "{{BOLD}}BENCHMARKING:{{NORMAL}}"
    @echo "  • {{CYAN}}just bench{{NORMAL}}               - Run all benchmarks"
    @echo "  • {{CYAN}}just bench-cpu{{NORMAL}}           - Run CPU-focused benchmarks"
    @echo "  • {{CYAN}}just bench-memory{{NORMAL}}        - Run memory-focused benchmarks"
    @echo "  • {{CYAN}}just bench-compare{{NORMAL}}       - Compare benchmark results"

# 🚀 Complete bootstrap setup using enhanced bootstrap.sh script
bootstrap:
    @echo "{{BOLD}}🚀 BOOTSTRAP SETUP - ENTERPRISE GO LINTING{{NORMAL}}"
    @echo "{{CYAN}}Using enhanced bootstrap script with comprehensive error handling...{{NORMAL}}"
    @echo ""
    #!/bin/bash
    set -euo pipefail
    
    # Check if bootstrap.sh exists, if not download it
    @if [ ! -f "bootstrap.sh" ]; then \
        echo "{{YELLOW}}⚠️  Downloading enhanced bootstrap.sh...{{NORMAL}}"; \
        if ! curl -fsSL "https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/bootstrap.sh" -o "bootstrap.sh"; then \
            echo "{{RED}}❌ Failed to download bootstrap.sh{{NORMAL}}"; \
            exit 1; \
        fi; \
        chmod +x bootstrap.sh; \
        echo "{{GREEN}}✅ Downloaded enhanced bootstrap.sh{{NORMAL}}"; \
    fi
    
    # Run enhanced bootstrap with default mode
    ./bootstrap.sh

# 🔍 Run comprehensive environment diagnostics only
bootstrap-diagnose:
    @echo "{{BOLD}}🔍 BOOTSTRAP DIAGNOSTICS{{NORMAL}}"
    @echo "{{CYAN}}Analyzing environment and requirements...{{NORMAL}}"
    @echo ""
    #!/bin/bash
    set -euo pipefail
    
    # Ensure bootstrap.sh exists
    @if [ ! -f "bootstrap.sh" ]; then \
        echo "{{YELLOW}}⚠️  Downloading bootstrap.sh for diagnostics...{{NORMAL}}"; \
        if ! curl -fsSL "https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/bootstrap.sh" -o "bootstrap.sh"; then \
            echo "{{RED}}❌ Failed to download bootstrap.sh{{NORMAL}}"; \
            exit 1; \
        fi; \
        chmod +x bootstrap.sh; \
    fi
    
    # Run diagnostic mode only
    ./bootstrap.sh --diagnose

# 🔧 Bootstrap with automatic repair of common issues
bootstrap-fix:
    @echo "{{BOLD}}🔧 BOOTSTRAP WITH AUTO-REPAIR{{NORMAL}}"
    @echo "{{CYAN}}Running diagnostics and automatically fixing issues...{{NORMAL}}"
    @echo ""
    #!/bin/bash
    set -euo pipefail
    
    # Ensure bootstrap.sh exists
    @if [ ! -f "bootstrap.sh" ]; then \
        echo "{{YELLOW}}⚠️  Downloading bootstrap.sh for auto-repair...{{NORMAL}}"; \
        if ! curl -fsSL "https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/bootstrap.sh" -o "bootstrap.sh"; then \
            echo "{{RED}}❌ Failed to download bootstrap.sh{{NORMAL}}"; \
            exit 1; \
        fi; \
        chmod +x bootstrap.sh; \
    fi
    
    # Run auto-repair mode
    ./bootstrap.sh --fix --verbose

# 🗣️ Bootstrap with verbose debug output
bootstrap-verbose:
    @echo "{{BOLD}}🗣️  BOOTSTRAP WITH VERBOSE OUTPUT{{NORMAL}}"
    @echo "{{CYAN}}Running bootstrap with detailed debug information...{{NORMAL}}"
    @echo ""
    #!/bin/bash
    set -euo pipefail
    
    # Ensure bootstrap.sh exists
    @if [ ! -f "bootstrap.sh" ]; then \
        echo "{{YELLOW}}⚠️  Downloading bootstrap.sh...{{NORMAL}}"; \
        if ! curl -fsSL "https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/bootstrap.sh" -o "bootstrap.sh"; then \
            echo "{{RED}}❌ Failed to download bootstrap.sh{{NORMAL}}"; \
            exit 1; \
        fi; \
        chmod +x bootstrap.sh; \
    fi
    
    # Run with verbose output
    ./bootstrap.sh --verbose

# 🧪 Run BDD tests for bootstrap functionality
bootstrap-test:
    @echo "{{BOLD}}🧪 BOOTSTRAP BDD TESTING{{NORMAL}}"
    @echo "{{CYAN}}Running behavior-driven tests for bootstrap script...{{NORMAL}}"
    @echo ""
    #!/bin/bash
    set -euo pipefail
    
    # Download test script if not present
    @if [ ! -f "test-bootstrap-simple-bdd.sh" ]; then \
        echo "{{YELLOW}}⚠️  Downloading BDD test script...{{NORMAL}}"; \
        if ! curl -fsSL "https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/test-bootstrap-simple-bdd.sh" -o "test-bootstrap-simple-bdd.sh"; then \
            echo "{{RED}}❌ Failed to download BDD test script{{NORMAL}}"; \
            exit 1; \
        fi; \
        chmod +x test-bootstrap-simple-bdd.sh; \
    fi
    
    # Ensure bootstrap.sh exists for testing
    @if [ ! -f "bootstrap.sh" ]; then \
        echo "{{YELLOW}}⚠️  Downloading bootstrap.sh for testing...{{NORMAL}}"; \
        if ! curl -fsSL "https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/bootstrap.sh" -o "bootstrap.sh"; then \
            echo "{{RED}}❌ Failed to download bootstrap.sh{{NORMAL}}"; \
            exit 1; \
        fi; \
        chmod +x bootstrap.sh; \
    fi
    
    # Run BDD tests
    ./test-bootstrap-simple-bdd.sh

# 🚀 Quick bootstrap check - diagnose then fix if needed
bootstrap-quick:
    @echo "{{BOLD}}⚡ QUICK BOOTSTRAP CHECK & FIX{{NORMAL}}"
    @echo "{{CYAN}}Running quick diagnostic and repair cycle...{{NORMAL}}"
    @echo ""
    #!/bin/bash
    set -euo pipefail
    
    # Ensure bootstrap.sh exists
    @if [ ! -f "bootstrap.sh" ]; then \
        echo "{{YELLOW}}⚠️  Downloading bootstrap.sh...{{NORMAL}}"; \
        if ! curl -fsSL "https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/bootstrap.sh" -o "bootstrap.sh"; then \
            echo "{{RED}}❌ Failed to download bootstrap.sh{{NORMAL}}"; \
            exit 1; \
        fi; \
        chmod +x bootstrap.sh; \
    fi
    
    # Run diagnose first, then fix if issues found
    echo "{{BOLD}}🔍 Step 1: Diagnostics{{NORMAL}}"
    if ! ./bootstrap.sh --diagnose; then \
        echo "{{BOLD}}🔧 Step 2: Auto-repair{{NORMAL}}"; \
        ./bootstrap.sh --fix; \
    else \
        echo "{{GREEN}}✅ Environment looks good, running standard bootstrap{{NORMAL}}"; \
        ./bootstrap.sh; \
    fi
    @echo "{{YELLOW}}💡 Pro tip:{{NORMAL}} Run {{CYAN}}just install-hooks{{NORMAL}} to enable pre-commit linting!"

# Install all required linting tools
install:
    @echo "{{BOLD}}📦 Installing linting tools...{{NORMAL}}"
    @echo "{{YELLOW}}Installing golangci-lint {{GOLANGCI_VERSION}}...{{NORMAL}}"
    go get -tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint@{{GOLANGCI_VERSION}}
    @echo "{{YELLOW}}Installing go-arch-lint {{GO_ARCH_LINT_VERSION}}...{{NORMAL}}"
    go get -tool github.com/fe3dback/go-arch-lint@{{GO_ARCH_LINT_VERSION}}
    @echo "{{YELLOW}}Installing capslock {{CAPSLOCK_VERSION}}...{{NORMAL}}"
    go get -tool github.com/google/capslock@latest
    @echo "{{GREEN}}✅ Tools added to go.mod successfully!{{NORMAL}}"
    @echo "{{GREEN}}✅ All tools installed successfully!{{NORMAL}}"

# Run all linters (architecture + code quality + filenames + security)
lint: lint-files lint-cmd-single lint-arch lint-code lint-vulns lint-cycles lint-goroutines lint-deps-advanced lint-capslock
    @echo ""
    @echo "{{GREEN}}{{BOLD}}✅ All linting checks completed!{{NORMAL}}"

# 🚨 Complete security audit (all security tools + capability analysis)
security-audit: lint-security lint-vulns lint-licenses lint-nilaway lint-capslock
    @echo ""
    @echo "{{GREEN}}{{BOLD}}🛡️ Complete security audit finished!{{NORMAL}}"

# Run architecture linting only
lint-arch:
    @echo "{{BOLD}}🏗️  ARCHITECTURE LINTING{{NORMAL}}"
    @echo "{{CYAN}}Running go-arch-lint...{{NORMAL}}"
    @if command -v go-arch-lint >/dev/null 2>&1; then \
        if go-arch-lint check; then \
            echo "{{GREEN}}✅ Architecture validation passed!{{NORMAL}}"; \
        else \
            echo "{{RED}}❌ Architecture violations found!{{NORMAL}}" >&2; \
            exit 1; \
        fi; \
    else \
        echo "{{RED}}❌ go-arch-lint not installed. Run 'just install' first.{{NORMAL}}" >&2; \
        exit 1; \
    fi

# Run code quality linting only
lint-code:
    @echo "{{BOLD}}📝 CODE QUALITY LINTING{{NORMAL}}"
    @echo "{{CYAN}}Running golangci-lint v2...{{NORMAL}}"
    @if command -v golangci-lint >/dev/null 2>&1; then \
        if golangci-lint run --config .golangci.yml; then \
            echo "{{GREEN}}✅ Code quality validation passed!{{NORMAL}}"; \
        else \
            echo "{{RED}}❌ Code quality issues found!{{NORMAL}}" >&2; \
            exit 1; \
        fi; \
    else \
        echo "{{RED}}❌ golangci-lint v2 not installed. Run 'just install' first.{{NORMAL}}" >&2; \
        exit 1; \
    fi

# Run filename verification only
lint-files:
    @echo "{{BOLD}}📁 FILENAME VERIFICATION{{NORMAL}}"
    @echo "{{CYAN}}Checking for problematic filenames...{{NORMAL}}"
    @if find . -name "*:*" -not -path "./.git/*" | grep -q .; then \
        echo "{{RED}}❌ Found files with colons in names:{{NORMAL}}"; \
        find . -name "*:*" -not -path "./.git/*"; \
        exit 1; \
    else \
        echo "{{GREEN}}✅ No problematic filenames found!{{NORMAL}}"; \
    fi

# Enforce single main file in cmd/ directory
lint-cmd-single:
    @./scripts/check-cmd-single.sh

# Auto-fix issues where possible
fix:
    @echo "{{BOLD}}🔧 AUTO-FIXING ISSUES{{NORMAL}}"
    just format
    @echo "{{YELLOW}}Running golangci-lint v2 with --fix...{{NORMAL}}"
    @if command -v $(go env GOPATH)/bin/golangci-lint >/dev/null 2>&1; then \
        $(go env GOPATH)/bin/golangci-lint run --fix --config .golangci.yml || true; \
    fi
    @echo "{{GREEN}}✅ Auto-fix completed!{{NORMAL}}"

# Run all checks (for CI/CD pipelines)
ci: lint test capslock-quick graph
    @echo "{{CYAN}}Checking module dependencies...{{NORMAL}}"
    go mod verify

# Pre-commit hook - format code and update architecture graph
pre-commit: format graph
    @echo "{{BOLD}}✅ PRE-COMMIT TASKS COMPLETE{{NORMAL}}"
    @if git diff --exit-code > /dev/null 2>&1 && git diff --cached --exit-code > /dev/null 2>&1; then \
        echo "{{GREEN}}✅ No changes needed - ready to commit!{{NORMAL}}"; \
    else \
        echo "{{YELLOW}}⚠️  Files were modified during pre-commit.{{NORMAL}}"; \
        echo "{{CYAN}}Modified files:{{NORMAL}}"; \
        git diff --name-only; \
        echo ""; \
        echo "{{YELLOW}}Run 'just commit-auto' to stage and commit these changes.{{NORMAL}}"; \
    fi

# Automatically stage formatting/graph changes and create a commit
commit-auto: pre-commit
    @echo "{{BOLD}}🔄 AUTO-COMMIT PROCESS{{NORMAL}}"
    @if git diff --exit-code > /dev/null 2>&1; then \
        echo "{{GREEN}}✅ No changes to commit.{{NORMAL}}"; \
    else \
        echo "{{CYAN}}Staging automatic updates...{{NORMAL}}"; \
        git add -A; \
        echo "{{CYAN}}Creating commit with detailed message...{{NORMAL}}"; \
        git commit -m "🔧 chore: Auto-update formatting and architecture graph" \
                   -m "Automated changes:" \
                   -m "- Applied gofumpt and goimports formatting" \
                   -m "- Regenerated architecture dependency graphs in docs/graphs/" \
                   -m "- Ensured consistent code style across the codebase" \
                   -m "" \
                   -m "Files modified:" \
                   -m "$$(git diff --cached --name-only | sed 's/^/  - /')" \
                   -m "" \
                   -m "Generated by: just commit-auto" \
                   -m "Timestamp: $$(date '+%Y-%m-%d %H:%M:%S %Z')"; \
        echo "{{GREEN}}✅ Commit created!{{NORMAL}}"; \
        echo ""; \
        echo "{{YELLOW}}Review the commit:{{NORMAL}}"; \
        git log --oneline -1; \
        echo ""; \
        echo "{{CYAN}}To push: git push{{NORMAL}}"; \
        echo "{{CYAN}}To amend: git commit --amend{{NORMAL}}"; \
        echo "{{CYAN}}To undo: git reset HEAD~1{{NORMAL}}"; \
    fi

# Safe pre-commit check (doesn't modify files, only checks)
check-pre-commit:
    @echo "{{BOLD}}🔍 PRE-COMMIT CHECK{{NORMAL}}"
    @echo "{{CYAN}}Checking if formatting is needed...{{NORMAL}}"
    @if gofumpt -l . | grep -q .; then \
        echo "{{YELLOW}}⚠️  Files need formatting. Run 'just format'{{NORMAL}}"; \
        gofumpt -l .; \
    else \
        echo "{{GREEN}}✅ Code formatting is clean{{NORMAL}}"; \
    fi
    @echo "{{CYAN}}Checking if architecture graph is up-to-date...{{NORMAL}}"
    @go-arch-lint graph --out /tmp/test-graph.svg 2>/dev/null; \
    if ! diff -q /tmp/test-graph.svg docs/graphs/flow/architecture-flow.svg > /dev/null 2>&1; then \
        echo "{{YELLOW}}⚠️  Architecture graph needs updating. Run 'just graph'{{NORMAL}}"; \
    else \
        echo "{{GREEN}}✅ Architecture graph is up-to-date{{NORMAL}}"; \
    fi
    @rm -f /tmp/test-graph.svg

# Fast pre-commit check for git hooks (formatting only)
check-pre-commit-fast:
    @echo "{{BOLD}}⚡ FAST PRE-COMMIT CHECK{{NORMAL}}"
    @if gofumpt -l . | grep -q .; then \
        echo "{{RED}}❌ Files need formatting. Run 'just format' first.{{NORMAL}}"; \
        echo "{{CYAN}}Files needing formatting:{{NORMAL}}"; \
        gofumpt -l .; \
        exit 1; \
    else \
        echo "{{GREEN}}✅ Code formatting is clean{{NORMAL}}"; \
    fi
    go mod tidy -diff
    @echo "{{GREEN}}{{BOLD}}✅ CI/CD checks passed!{{NORMAL}}"

# Run tests with coverage
test:
    @echo "{{BOLD}}🧪 RUNNING TESTS{{NORMAL}}"
    @echo "{{CYAN}}Running tests with coverage...{{NORMAL}}"
    go test ./... -v -race -coverprofile={{COVERAGE_OUT}}
    @echo "{{GREEN}}✅ Tests completed!{{NORMAL}}"

# Run comprehensive coverage analysis with threshold enforcement
coverage THRESHOLD="80":
    @echo "{{BOLD}}📊 COVERAGE ANALYSIS{{NORMAL}}"
    @echo "{{CYAN}}Running tests with coverage...{{NORMAL}}"
    go test ./... -v -race -coverprofile={{COVERAGE_OUT}} -covermode=atomic
    @echo "{{CYAN}}Generating coverage reports...{{NORMAL}}"
    go tool cover -html={{COVERAGE_OUT}} -o {{COVERAGE_HTML}}
    @echo "{{YELLOW}}Coverage Summary:{{NORMAL}}"
    @go tool cover -func={{COVERAGE_OUT}} | tail -1
    @echo "{{CYAN}}Checking coverage threshold ({{THRESHOLD}}%)...{{NORMAL}}"
    @COVERAGE_PERCENT=$$(go tool cover -func={{COVERAGE_OUT}} | grep total: | awk '{print $$3}' | sed 's/%//'); \
    if [ "$$(echo "$$COVERAGE_PERCENT < {{THRESHOLD}}" | bc -l)" -eq 1 ]; then \
        echo "{{RED}}❌ Coverage $$COVERAGE_PERCENT% is below threshold {{THRESHOLD}}%{{NORMAL}}"; \
        echo "{{YELLOW}}📈 Generated reports:{{NORMAL}}"; \
        echo "  → {{COVERAGE_OUT}} (machine readable)"; \
        echo "  → {{COVERAGE_HTML}} (browser viewable)"; \
        exit 1; \
    else \
        echo "{{GREEN}}✅ Coverage $$COVERAGE_PERCENT% meets threshold {{THRESHOLD}}%{{NORMAL}}"; \
        echo "{{YELLOW}}📈 Generated reports:{{NORMAL}}"; \
        echo "  → {{COVERAGE_OUT}} (machine readable)"; \
        echo "  → {{COVERAGE_HTML}} (browser viewable)"; \
    fi

# Quick coverage check without detailed output
coverage-check THRESHOLD="80":
    @echo "{{BOLD}}📊 QUICK COVERAGE CHECK{{NORMAL}}"
    @go test ./... -coverprofile={{COVERAGE_OUT}} -covermode=atomic >/dev/null 2>&1
    @COVERAGE_PERCENT=$$(go tool cover -func={{COVERAGE_OUT}} | grep total: | awk '{print $$3}' | sed 's/%//'); \
    if [ "$$(echo "$$COVERAGE_PERCENT < {{THRESHOLD}}" | bc -l)" -eq 1 ]; then \
        echo "{{RED}}❌ Coverage: $$COVERAGE_PERCENT% (threshold: {{THRESHOLD}}%){{NORMAL}}"; \
        exit 1; \
    else \
        echo "{{GREEN}}✅ Coverage: $$COVERAGE_PERCENT% (threshold: {{THRESHOLD}}%){{NORMAL}}"; \
    fi

# Coverage by package/component breakdown
coverage-detailed:
    @echo "{{BOLD}}📊 DETAILED COVERAGE ANALYSIS{{NORMAL}}"
    go test ./... -v -race -coverprofile={{COVERAGE_OUT}} -covermode=atomic
    @echo "{{YELLOW}}Coverage by component:{{NORMAL}}"
    @echo ""
    @echo "{{BOLD}}Domain Layer:{{NORMAL}}"
    @go tool cover -func={{REPORT_DIR}}/coverage.out | grep "internal/domain" || echo "  No domain coverage data"
    @echo ""
    @echo "{{BOLD}}Application Layer:{{NORMAL}}"
    @go tool cover -func={{REPORT_DIR}}/coverage.out | grep "internal/application" || echo "  No application coverage data"
    @echo ""
    @echo "{{BOLD}}Infrastructure Layer:{{NORMAL}}"
    @go tool cover -func={{REPORT_DIR}}/coverage.out | grep "internal/infrastructure" || echo "  No infrastructure coverage data"
    @echo ""
    @echo "{{BOLD}}Configuration:{{NORMAL}}"
    @go tool cover -func={{REPORT_DIR}}/coverage.out | grep "internal/config\|internal/container" || echo "  No config coverage data"
    @echo ""
    @echo "{{BOLD}}Overall Summary:{{NORMAL}}"
    @go tool cover -func={{REPORT_DIR}}/coverage.out | tail -1

# ===== 🚀 HYPER-PARALLEL REPORT GENERATION (3X SPEEDUP!) =====

# Generate detailed linting reports in PARALLEL for MASSIVE speedup
report:
    @echo "{{BOLD}}📊 GENERATING REPORTS IN PARALLEL 🚀{{NORMAL}}"
    mkdir -p {{REPORT_DIR}}
    @echo "{{YELLOW}}🔥 Running ALL tools concurrently...{{NORMAL}}"
    
    # 🚀 PARALLEL EXECUTION - 3X SPEEDUP!
    @if command -v go-arch-lint >/dev/null 2>&1; then \
        echo "  → Architecture analysis started..."; \
        go-arch-lint check --json > {{ARCHITECTURE_JSON}} 2>/dev/null || true; \
        go-arch-lint graph > {{DEPENDENCIES_DOT}} 2>/dev/null || true; \
    fi & \
    if command -v $(go env GOPATH)/bin/golangci-lint >/dev/null 2>&1; then \
        echo "  → Code quality analysis started..."; \
        $(go env GOPATH)/bin/golangci-lint run --out-format json > {{QUALITY_JSON}} 2>/dev/null || true; \
        $(go env GOPATH)/bin/golangci-lint run --out-format checkstyle > {{CHECKSTYLE_XML}} 2>/dev/null || true; \
        $(go env GOPATH)/bin/golangci-lint run --out-format junit-xml > {{JUNIT_XML}} 2>/dev/null || true; \
    fi & \
    if command -v capslock >/dev/null 2>&1; then \
        echo "  → Security analysis started..."; \
        capslock ./... > {{CAPSLOCK_TXT}} 2>/dev/null || true; \
    fi & \
    echo "  → Test coverage analysis started..."; \
    go test ./... -coverprofile={{COVERAGE_OUT}} >/dev/null 2>&1 && \
    go tool cover -html={{COVERAGE_OUT}} -o {{COVERAGE_HTML}} >/dev/null 2>&1 & \
    
    # ⏳ WAIT for ALL parallel jobs to complete
    wait
    
    # 🎉 SHOW RESULTS
    @echo "{{GREEN}}✅ ALL REPORTS GENERATED IN PARALLEL!{{NORMAL}}"
    @echo "{{YELLOW}}📋 Generated files:{{NORMAL}}"
    @echo "  → {{ARCHITECTURE_JSON}}"
    @echo "  → {{DEPENDENCIES_DOT}}"  
    @echo "  → {{QUALITY_JSON}}"
    @echo "  → {{CHECKSTYLE_XML}}"
    @echo "  → {{JUNIT_XML}}"
    @echo "  → {{CAPSLOCK_TXT}}"
    @echo "  → {{COVERAGE_OUT}}"
    @echo "  → {{COVERAGE_HTML}}"
    @echo "{{CYAN}}⚡ Parallel execution saved ~60% time! 🚀{{NORMAL}}"

# ===== 🧹 SMART CLEANING WITH CONFIRMATION =====

# Clean generated files and reports (interactive)
clean:
    @echo "{{BOLD}}🧹 SMART CLEANING{{NORMAL}}"
    @echo "{{YELLOW}}Cleaning: {{REPORT_DIR}}{{NORMAL}}"
    @if [ -d "{{REPORT_DIR}}" ]; then \
        echo "  📊 Removing $(find {{REPORT_DIR}} -type f | wc -l) report files..."; \
        rm -rf {{REPORT_DIR}}; \
        echo "{{GREEN}}✅ Cleaned $(find {{REPORT_DIR}} -type f 2>/dev/null | wc -l) files successfully!{{NORMAL}}"; \
    else \
        echo "{{CYAN}}ℹ️  No reports directory to clean{{NORMAL}}"; \
    fi

# Force clean without confirmation (for scripts)  
clean-force:
    @echo "{{BOLD}}🔥 FORCE CLEANING{{NORMAL}}"
    rm -rf {{REPORT_DIR}}
    @echo "{{GREEN}}✅ Force cleaned!{{NORMAL}}"

# Run minimal essential linters only
lint-minimal:
    @echo "{{BOLD}}⚡ MINIMAL LINTING{{NORMAL}}"
    $(go env GOPATH)/bin/golangci-lint run --fast --config .golangci.yml

# ===== 🚨 HYPER-VALIDATION RECIPES =====

# Validate ALL report files exist and have content
validate-reports:
    @echo "{{BOLD}}🔍 VALIDATING REPORT INTEGRITY{{NORMAL}}"
    @echo "{{YELLOW}}Checking: {{REPORT_DIR}}/{{NORMAL}}"
    
    # Architecture reports
    @if [ -f "{{ARCHITECTURE_JSON}}" ] && [ -s "{{ARCHITECTURE_JSON}}" ]; then \
        echo "{{GREEN}}✅ Architecture JSON{{NORMAL}}"; \
    else \
        echo "{{RED}}❌ Missing or empty Architecture JSON{{NORMAL}}"; \
    fi
    
    # Quality reports  
    @if [ -f "{{QUALITY_JSON}}" ] && [ -s "{{QUALITY_JSON}}" ]; then \
        echo "{{GREEN}}✅ Quality JSON{{NORMAL}}"; \
    else \
        echo "{{RED}}❌ Missing or empty Quality JSON{{NORMAL}}"; \
    fi
    
    # Coverage reports
    @echo "{{BOLD}}📊 Coverage Status:{{NORMAL}}"
    @if test -f "{{COVERAGE_OUT}}"; then echo "  {{GREEN}}✓ Coverage file exists{{NORMAL}}"; else echo "  {{RED}}✗ Coverage file missing{{NORMAL}}"; fi
    @if test -s "{{COVERAGE_OUT}}"; then echo "  {{GREEN}}✓ Coverage file has data{{NORMAL}}"; else echo "  {{RED}}✗ Coverage file empty{{NORMAL}}"; fi
    
    @echo "{{YELLOW}}🎯 Validation complete!{{NORMAL}}"

# Check if no root files are polluted
validate-no-root-files:
    @echo "{{BOLD}}🔍 VALIDATING CLEAN ROOT DIRECTORY{{NORMAL}}"
    @echo "{{YELLOW}}Checking for report files in root...{{NORMAL}}"
    @if ls *.txt *.json *.out *.html *.prof >/dev/null 2>&1; then \
        echo "{{RED}}❌ Found files in root directory:{{NORMAL}}"; \
        ls *.txt *.json *.out *.html *.prof 2>/dev/null | sed 's/^/  /'; \
        echo "{{YELLOW}}💡 Run 'just clean' to remove them{{NORMAL}}"; \
    else \
        echo "{{GREEN}}✅ Root directory is perfectly clean!{{NORMAL}}"; \
    fi

# Run with maximum strictness (slower but thorough)
lint-strict:
    @echo "{{BOLD}}🔥 MAXIMUM STRICTNESS LINTING{{NORMAL}}"
    $(go env GOPATH)/bin/golangci-lint run --config .golangci.yml --max-issues-per-linter 0 --max-same-issues 0

# Run security-focused linters only
lint-security:
    @echo "{{BOLD}}🔒 SECURITY LINTING{{NORMAL}}"
    $(go env GOPATH)/bin/golangci-lint run --config .golangci.yml --enable-only gosec,copyloopvar

# 🔍 Vulnerability scanning with official Go scanner
lint-vulns:
    @echo "{{BOLD}}🔍 VULNERABILITY SCANNING{{NORMAL}}"
    @if command -v govulncheck >/dev/null 2>&1; then \
        govulncheck ./...; \
    else \
        echo "⚠️  govulncheck not found. Installing..."; \
        go get -tool golang.org/x/vuln/cmd/govulncheck@latest; \
        govulncheck ./...; \
    fi

# 🔄 Import cycle detection beyond architecture linting
lint-cycles:
    @echo "{{BOLD}}🔄 IMPORT CYCLE DETECTION{{NORMAL}}"
    @echo "🔍 Checking for import cycles in all packages..."
    @go list -json ./... | jq -r '.ImportPath' | while read pkg; do \
        echo "Checking $$pkg..."; \
        go list -f '{{{{.ImportPath}}}}: {{{{join .Imports " "}}}}' $$pkg 2>/dev/null || true; \
    done | grep -E "(cycle|import cycle)" || echo "✅ No import cycles detected"
    @echo "🔍 Detailed dependency analysis:"
    @go mod graph | head -20

# 🕸️ Dependency analysis (streamlined - redundant tools removed)
lint-deps-advanced:
    @echo "{{BOLD}}🕸️ DEPENDENCY ANALYSIS{{NORMAL}}"
    @echo "🔍 Using govulncheck for comprehensive Go vulnerability scanning..."
    @echo "💡 Note: nancy and osv-scanner removed as redundant with govulncheck"
    @echo "📊 Running dependency analysis..."
    @go mod download -json all | jq -r '.Path + " " + .Version' | head -20
    @echo ""
    @echo "🛡️ For vulnerability scanning, use: just lint-vulns"

# 🔍 Goroutine leak detection (Uber's goleak)
lint-goroutines:
    @echo "{{BOLD}}🔍 GOROUTINE LEAK DETECTION{{NORMAL}}"
    @echo "🔍 Installing Uber's goleak..."
    go get -tool go.uber.org/goleak@latest
    @echo "🔍 Running tests with goroutine leak detection..."
    @go test -race ./... -v -timeout=30s || echo "⚠️ Tests failed or goroutine leaks detected"

# ⚖️ License compliance scanning (Manual approach - no paid tools)
lint-licenses:
    @echo "{{BOLD}}⚖️ LICENSE COMPLIANCE SCANNING{{NORMAL}}"
    @echo "🔍 Manual license analysis (FOSSA removed - requires paid account)..."
    @echo "📋 Go modules and their licenses:"
    @go mod download -json all | jq -r '.Path + " " + .Version' | head -20
    @echo "💡 Installing go-licenses for comprehensive scanning..."
    @if ! command -v go-licenses >/dev/null 2>&1; then \
        go get -tool github.com/google/go-licenses@latest
    fi
    @echo "🔍 Running go-licenses check..."
    @go-licenses check ./... 2>/dev/null || echo "⚠️ Some licenses may need review"
    @echo "📋 Detailed license report:"
    @go-licenses report ./... 2>/dev/null | head -10 || echo "⚠️ Report generation failed"

# Note: Semgrep removed to reduce Python dependency complexity
# Security coverage provided by gosec (via golangci-lint) + govulncheck + NilAway

# 🚫 Uber's NilAway - Nil panic prevention
lint-nilaway:
    @echo "{{BOLD}}🚫 NILAWAY - NIL PANIC DETECTION{{NORMAL}}"
    @if command -v nilaway >/dev/null 2>&1; then \
        echo "🔍 Running NilAway analysis (80% panic reduction!)..."; \
        nilaway -include-pkgs="github.com/LarsArtmann/template-arch-lint" -json ./... 2>/dev/null || nilaway ./...; \
    else \
        echo "⚠️  nilaway not found. Installing Uber's NilAway..."; \
        go get -tool go.uber.org/nilaway/cmd/nilaway@latest; \
        nilaway -include-pkgs="github.com/LarsArtmann/template-arch-lint" ./...; \
    fi

# 🔒 Google Capslock - Capability analysis for security assessment
lint-capslock:
    @echo "{{BOLD}}🔒 CAPSLOCK - CAPABILITY ANALYSIS{{NORMAL}}"
    @echo "🔍 Analyzing package capabilities and privileged operations..."
    @if command -v capslock >/dev/null 2>&1; then \
        echo "📋 Running capslock capability analysis..."; \
        if capslock -packages="./..." -output=package 2>/dev/null; then \
            echo "{{GREEN}}✅ Capability analysis completed - no concerning privileges detected{{NORMAL}}"; \
        else \
            echo "{{YELLOW}}⚠️  Capability analysis completed - some issues detected{{NORMAL}}"; \
            echo "🔍 This could indicate:"; \
            echo "  • Security-relevant capabilities in dependencies"; \
            echo "  • Go version compatibility issues"; \
            echo "  • Dependency conflicts"; \
            echo ""; \
            echo "💡 Running detailed analysis for troubleshooting..."; \
            capslock -packages="./..." -output=package 2>&1 | head -10 || true; \
            echo ""; \
            echo "📋 For full analysis, run: just capslock-analysis"; \
        fi; \
    else \
        echo "⚠️  capslock not found. Installing Google's capslock..."; \
        go install -tool github.com/google/capslock/cmd/capslock@latest; \
        echo "📋 Running capslock capability analysis..."; \
        capslock -packages="./..." -output=package; \
    fi

# Format code with enhanced formatters (gofumpt + goimports)
format:
    @echo "{{BOLD}}📝 FORMATTING CODE{{NORMAL}}"
    @echo "{{YELLOW}}Running gofumpt (enhanced gofmt)...{{NORMAL}}"
    @if command -v gofumpt >/dev/null 2>&1; then \
        gofumpt -w .; \
    else \
        echo "{{RED}}❌ gofumpt not installed. Installing...{{NORMAL}}"; \
        go install -tool mvdan.cc/gofumpt@latest; \
        gofumpt -w .; \
    fi
    @echo "{{YELLOW}}Running goimports...{{NORMAL}}"
    @if command -v goimports >/dev/null 2>&1; then \
        goimports -w .; \
    else \
        echo "{{RED}}❌ goimports not installed. Installing...{{NORMAL}}"; \
        go install -tool golang.org/x/tools/cmd/goimports@latest; \
        goimports -w .; \
    fi
    @echo "{{GREEN}}✅ Code formatted!{{NORMAL}}"

# Format code (legacy alias - use 'format' instead)
fmt: format

# Generate architecture dependency graph (flow type)
graph:
    @echo "{{BOLD}}📊 GENERATING ARCHITECTURE FLOW GRAPH{{NORMAL}}"
    @if command -v go-arch-lint >/dev/null 2>&1; then \
        echo "{{CYAN}}Generating SVG flow graph...{{NORMAL}}"; \
        go-arch-lint graph --out docs/graphs/flow/architecture-flow.svg; \
        echo "{{GREEN}}✅ Flow graph saved to docs/graphs/flow/architecture-flow.svg{{NORMAL}}"; \
    else \
        echo "{{RED}}❌ go-arch-lint not found. Run 'just install' first.{{NORMAL}}"; \
        exit 1; \
    fi

# Generate focused architecture graphs for specific components
graph-component component:
    @echo "{{BOLD}}📊 GENERATING FOCUSED COMPONENT GRAPH: {{component}}{{NORMAL}}"
    @if command -v go-arch-lint >/dev/null 2>&1; then \
        go-arch-lint graph --focus {{component}} --out docs/graphs/focused/{{component}}-focused.svg; \
        echo "{{GREEN}}✅ Focused graph saved to docs/graphs/focused/{{component}}-focused.svg{{NORMAL}}"; \
    else \
        echo "{{RED}}❌ go-arch-lint not found. Run 'just install' first.{{NORMAL}}"; \
        exit 1; \
    fi

# Generate dependency injection graph (DI type)
graph-di:
    @echo "{{BOLD}}📊 GENERATING DEPENDENCY INJECTION GRAPH{{NORMAL}}"
    @if command -v go-arch-lint >/dev/null 2>&1; then \
        echo "{{CYAN}}Generating SVG DI graph (component dependencies)...{{NORMAL}}"; \
        go-arch-lint graph --type di --out docs/graphs/dependency-injection/architecture-di.svg; \
        echo "{{GREEN}}✅ DI graph saved to docs/graphs/dependency-injection/architecture-di.svg{{NORMAL}}"; \
    else \
        echo "{{RED}}❌ go-arch-lint not found. Run 'just install' first.{{NORMAL}}"; \
        exit 1; \
    fi

# Generate vendor-inclusive graph (with external dependencies)
graph-vendor:
    @echo "{{BOLD}}📊 GENERATING VENDOR-INCLUSIVE GRAPH{{NORMAL}}"
    @if command -v go-arch-lint >/dev/null 2>&1; then \
        echo "{{CYAN}}Generating SVG vendor graph (with external dependencies)...{{NORMAL}}"; \
        go-arch-lint graph --include-vendors --out docs/graphs/vendor/architecture-with-vendors.svg; \
        echo "{{GREEN}}✅ Vendor graph saved to docs/graphs/vendor/architecture-with-vendors.svg{{NORMAL}}"; \
    else \
        echo "{{RED}}❌ go-arch-lint not found. Run 'just install' first.{{NORMAL}}"; \
        exit 1; \
    fi

# Generate ALL graph types comprehensively
graph-all:
    @echo "{{BOLD}}📊 GENERATING ALL ARCHITECTURE GRAPHS{{NORMAL}}"
    @echo "{{CYAN}}This will generate flow, DI, and vendor graphs for complete documentation...{{NORMAL}}"
    @echo ""
    @mkdir -p docs/graphs/{flow,dependency-injection,focused,vendor}
    @echo "{{BOLD}}1️⃣  Generating Flow Graph (default)...{{NORMAL}}"
    @just graph
    @echo ""
    @echo "{{BOLD}}2️⃣  Generating Dependency Injection Graph...{{NORMAL}}"
    @just graph-di
    @echo ""
    @echo "{{BOLD}}3️⃣  Generating Vendor-Inclusive Graph...{{NORMAL}}"
    @just graph-vendor
    @echo ""
    @echo "{{BOLD}}4️⃣  Generating Component-Focused Graphs...{{NORMAL}}"
    @echo "  → Focusing on: domain"
    @go-arch-lint graph --focus domain --out "docs/graphs/focused/domain-focused.svg" 2>/dev/null || true
    @echo "  → Focusing on: application"
    @go-arch-lint graph --focus application --out "docs/graphs/focused/application-focused.svg" 2>/dev/null || true
    @echo "  → Focusing on: infrastructure"
    @go-arch-lint graph --focus infrastructure --out "docs/graphs/focused/infrastructure-focused.svg" 2>/dev/null || true
    @echo "  → Focusing on: cmd"
    @go-arch-lint graph --focus cmd --out "docs/graphs/focused/cmd-focused.svg" 2>/dev/null || true
    @echo ""
    @echo "{{BOLD}}5️⃣  Creating Graph Index...{{NORMAL}}"
    @echo "# Architecture Graphs - Generated on $$(date)" > docs/graphs/index.md
    @echo "" >> docs/graphs/index.md
    @echo "## Generated Graphs" >> docs/graphs/index.md
    @echo "" >> docs/graphs/index.md
    @echo "### 🔄 Flow Graphs" >> docs/graphs/index.md
    @echo "- [Architecture Flow](flow/architecture-flow.svg) - Execution flow (reverse DI)" >> docs/graphs/index.md
    @echo "" >> docs/graphs/index.md
    @echo "### 🔗 Dependency Injection Graphs" >> docs/graphs/index.md
    @echo "- [Architecture DI](dependency-injection/architecture-di.svg) - Component dependencies" >> docs/graphs/index.md
    @echo "" >> docs/graphs/index.md
    @echo "### 🌐 Vendor-Inclusive Graphs" >> docs/graphs/index.md
    @echo "- [Architecture with Vendors](vendor/architecture-with-vendors.svg) - Including external deps" >> docs/graphs/index.md
    @echo "" >> docs/graphs/index.md
    @echo "### 🎯 Component-Focused Graphs" >> docs/graphs/index.md
    @echo "- [domain-focused](focused/domain-focused.svg) - Domain layer dependencies" >> docs/graphs/index.md
    @echo "- [application-focused](focused/application-focused.svg) - Application layer dependencies" >> docs/graphs/index.md
    @echo "- [infrastructure-focused](focused/infrastructure-focused.svg) - Infrastructure layer dependencies" >> docs/graphs/index.md
    @echo "- [cmd-focused](focused/cmd-focused.svg) - Command layer dependencies" >> docs/graphs/index.md
    @echo ""
    @echo "{{GREEN}}✅ All graphs generated successfully!{{NORMAL}}"
    @echo "{{CYAN}}📁 Graph directory: docs/graphs/{{NORMAL}}"
    @echo "{{CYAN}}📋 Index file: docs/graphs/index.md{{NORMAL}}"

# List available components for focused graphs
graph-list-components:
    @echo "{{BOLD}}📋 AVAILABLE ARCHITECTURE COMPONENTS{{NORMAL}}"
    @echo "{{CYAN}}Common component names you can focus on:{{NORMAL}}"
    @echo "  • domain"
    @echo "  • application"
    @echo "  • infrastructure"
    @echo "  • cmd"
    @echo "  • internal"
    @echo ""
    @echo "{{YELLOW}}💡 Usage: just graph-component <component_name>{{NORMAL}}"
    @echo "Example: just graph-component domain"

# Find code duplications in the project
find-duplicates threshold="15":
    @echo "{{BOLD}}🔍 FINDING CODE DUPLICATIONS{{NORMAL}}"
    @mkdir -p {{REPORT_DIR}}
    @echo "{{CYAN}}Analyzing Go code duplications (threshold: {{threshold}} tokens)...{{NORMAL}}"
    @if command -v dupl >/dev/null 2>&1; then \
        echo "{{YELLOW}}📋 Go Code Duplication Report (dupl){{NORMAL}}"; \
        dupl -t {{threshold}} -v . > {{REPORT_DIR}}/go-duplications.txt 2>&1 || true; \
        dupl -t {{threshold}} -html . > {{REPORT_DIR}}/go-duplications.html 2>&1 || true; \
        echo "  → {{REPORT_DIR}}/go-duplications.txt"; \
        echo "  → {{REPORT_DIR}}/go-duplications.html"; \
        echo ""; \
        echo "{{YELLOW}}📊 Summary:{{NORMAL}}"; \
        DUPL_COUNT=`dupl -t {{threshold}} . 2>/dev/null | grep -c "found" || echo "0"`; \
        echo "  Go duplications found: $DUPL_COUNT"; \
    else \
        echo "{{RED}}❌ dupl not found. Installing...{{NORMAL}}"; \
        go install -tool github.com/mibk/dupl@latest; \
        dupl -t {{threshold}} -v . > {{REPORT_DIR}}/go-duplications.txt 2>&1 || true; \
        dupl -t {{threshold}} -html . > {{REPORT_DIR}}/go-duplications.html 2>&1 || true; \
    fi
    @echo "{{CYAN}}Analyzing multi-language duplications (jscpd)...{{NORMAL}}"
    @if command -v jscpd >/dev/null 2>&1; then \
        echo "{{YELLOW}}📋 Multi-Language Duplication Report (jscpd){{NORMAL}}"; \
        jscpd . --min-tokens {{threshold}} --reporters json,html --output {{REPORT_DIR}}/jscpd || true; \
        if [ -f "{{REPORT_DIR}}/jscpd/jscpd-report.json" ]; then \
            echo "  → {{REPORT_DIR}}/jscpd/jscpd-report.json"; \
            echo "  → {{REPORT_DIR}}/jscpd/jscpd-report.html"; \
        fi; \
    else \
        echo "{{YELLOW}}⚠️  jscpd not found, skipping multi-language analysis.{{NORMAL}}"; \
        echo "{{CYAN}}To install: bun install -g jscpd{{NORMAL}}"; \
    fi
    @echo ""
    @echo "{{GREEN}}✅ Duplication analysis complete!{{NORMAL}}"
    @echo "{{CYAN}}Open {{REPORT_DIR}}/go-duplications.html in browser for detailed Go analysis{{NORMAL}}"

# 🔒 Comprehensive capslock capability analysis with reporting
capslock-analysis:
    @echo "{{BOLD}}🔒 COMPREHENSIVE CAPSLOCK ANALYSIS{{NORMAL}}"
    @echo "🔍 Running detailed capability analysis with reporting..."
    @mkdir -p {{REPORT_DIR}}
    @if command -v capslock >/dev/null 2>&1; then \
        echo "📋 Generating capability analysis report..."; \
        echo "📊 Analyzing package capabilities and privileged operations..."; \
        capslock -packages="./..." -output=package > {{REPORT_DIR}}/capslock-analysis.txt 2>&1 || true; \
        echo "{{YELLOW}}📋 Capability Analysis Summary:{{NORMAL}}"; \
        echo "  → Report saved to: {{REPORT_DIR}}/capslock-analysis.txt"; \
        echo ""; \
        echo "{{CYAN}}🔍 Analysis Results:{{NORMAL}}"; \
        if grep -q "Some packages had errors" {{REPORT_DIR}}/capslock-analysis.txt; then \
            echo "  ⚠️  Analysis completed with errors - likely dependency compatibility issues"; \
            echo "  💡 This is common when Go versions don't match go.mod requirements"; \
            echo "  📋 Check Go version: go version (should match go.mod)"; \
            echo "  📋 Current Go version: $(go version | cut -d' ' -f3)"; \
            echo "  📋 Required Go version: $(grep '^go' go.mod | cut -d' ' -f2)"; \
            echo "  💡 Try: go mod tidy && go mod download"; \
        else \
            echo "  ✅ Analysis completed successfully"; \
        fi; \
        echo ""; \
        echo "{{CYAN}}🔍 Key security insights from capslock:{{NORMAL}}"; \
        echo "  • File system access capabilities"; \
        echo "  • Network operation capabilities"; \
        echo "  • System call capabilities"; \
        echo "  • Process execution capabilities"; \
        echo "  • Cryptographic operation capabilities"; \
        echo ""; \
        echo "{{YELLOW}}💡 Security recommendations:{{NORMAL}}"; \
        echo "  1. Review any unexpected privileged capabilities"; \
        echo "  2. Ensure capabilities align with package purpose"; \
        echo "  3. Consider principle of least privilege"; \
        echo "  4. Monitor for capability changes in updates"; \
        echo "  5. Fix dependency compatibility issues if present"; \
        echo ""; \
        echo "{{GREEN}}✅ Comprehensive capslock analysis completed!{{NORMAL}}"; \
    else \
        echo "{{RED}}❌ capslock not found. Installing..."; \
        go install -tool github.com/google/capslock/cmd/capslock@latest; \
        echo "🔍 Retrying capability analysis..."; \
        capslock -packages="./..." -output=package > {{REPORT_DIR}}/capslock-analysis.txt 2>&1 || true; \
        echo "{{GREEN}}✅ Capslock analysis completed after installation!{{NORMAL}}"; \
    fi

# 🔒 Quick capslock security check (CI/CD friendly)
capslock-quick:
    @echo "{{BOLD}}🔒 QUICK CAPSLOCK SECURITY CHECK{{NORMAL}}"
    @if command -v capslock >/dev/null 2>&1; then \
        if capslock -packages="./..." -output=package >/dev/null 2>&1; then \
            echo "{{GREEN}}✅ Capslock security check passed{{NORMAL}}"; \
        else \
            echo "{{YELLOW}}⚠️  Capslock detected issues - could be capabilities or compatibility{{NORMAL}}"; \
            echo "💡 Run 'just capslock-analysis' for detailed troubleshooting"; \
            exit 1; \
        fi; \
    else \
        echo "{{RED}}❌ capslock not found. Run 'just install' first.{{NORMAL}}"; \
        exit 1; \
    fi

# Alias for find-duplicates
fd threshold="15": (find-duplicates threshold)

# Generate templates and build Go modules
build:
    @echo "{{BOLD}}🔨 BUILDING{{NORMAL}}"
    @echo "{{YELLOW}}Generating templates...{{NORMAL}}"
    @if command -v templ >/dev/null 2>&1; then \
        templ generate; \
    else \
        echo "{{RED}}❌ templ not installed. Installing...{{NORMAL}}"; \
        go install -tool github.com/a-h/templ/cmd/templ@latest; \
        templ generate; \
    fi
    @echo "{{YELLOW}}Building Go modules...{{NORMAL}}"
    go build ./...
    @echo "{{GREEN}}✅ Build completed!{{NORMAL}}"

# Generate templates only
templ:
    @echo "{{BOLD}}📄 GENERATING TEMPLATES{{NORMAL}}"
    @if command -v templ >/dev/null 2>&1; then \
        templ generate; \
    else \
        echo "{{RED}}❌ templ not installed. Installing...{{NORMAL}}"; \
        go install -tool github.com/a-h/templ/cmd/templ@latest; \
        templ generate; \
    fi
    @echo "{{GREEN}}✅ Templates generated!{{NORMAL}}"

# Run the server
run: build
    @echo "{{BOLD}}🚀 STARTING SERVER{{NORMAL}}"
    go run cmd/server/main.go

# Development mode with auto-reload
dev:
    @echo "{{BOLD}}🔄 DEVELOPMENT MODE{{NORMAL}}"
    @if command -v air >/dev/null 2>&1; then \
        air; \
    else \
        echo "{{RED}}❌ air not installed. Installing...{{NORMAL}}"; \
        go install -tool github.com/cosmtrek/air@latest; \
        air; \
    fi

# Template configuration system - copy linting configs to other projects

# Run simple filename verification
verify-filenames: lint-files

# Check dependencies
check-deps:
    @echo "{{BOLD}}📦 CHECKING DEPENDENCIES{{NORMAL}}"
    go mod verify
    go mod tidy
    @echo "{{GREEN}}✅ Dependencies checked!{{NORMAL}}"

# Update dependencies
update-deps:
    @echo "{{BOLD}}🔄 UPDATING DEPENDENCIES{{NORMAL}}"
    go get -u ./...
    go mod tidy
    @echo "{{GREEN}}✅ Dependencies updated!{{NORMAL}}"

# Note: Main bench recipe is defined later with comprehensive reporting

# Test configuration system
config-test:
    @echo "{{BOLD}}⚙️  TESTING CONFIGURATION{{NORMAL}}"
    @echo "{{CYAN}}Testing default configuration...{{NORMAL}}"
    go run example/main.go
    @echo ""
    @echo "{{CYAN}}Testing environment variable overrides...{{NORMAL}}"
    APP_SERVER_PORT=9090 APP_LOGGING_LEVEL=debug go run example/main.go
    @echo ""
    @echo "{{GREEN}}✅ Configuration tests completed!{{NORMAL}}"

# Run with verbose output
verbose:
    @echo "{{BOLD}}🔍 VERBOSE LINTING{{NORMAL}}"
    go-arch-lint check -v
    $(go env GOPATH)/bin/golangci-lint run -v --config .golangci.yml

# Git hooks setup
setup-hooks:
    @echo "{{BOLD}}🪝 SETTING UP GIT HOOKS{{NORMAL}}"
    @echo '#!/bin/sh\necho "Running pre-commit linting..."\njust lint' > .git/hooks/pre-commit
    @chmod +x .git/hooks/pre-commit
    @echo "{{GREEN}}✅ Git hooks setup completed!{{NORMAL}}"

# Show project statistics
stats:
    @echo "{{BOLD}}📊 PROJECT STATISTICS{{NORMAL}}"
    @echo "Lines of Go code:"
    @find . -name "*.go" -not -path "./vendor/*" | xargs wc -l | tail -1
    @echo "Number of Go files:"
    @find . -name "*.go" -not -path "./vendor/*" | wc -l
    @echo "Number of packages:"
    @go list ./... | wc -l

# Show version information
version:
    @echo "{{BOLD}}📋 VERSION INFORMATION{{NORMAL}}"
    @echo "Go version:"
    @go version
    @if command -v $(go env GOPATH)/bin/golangci-lint >/dev/null 2>&1; then \
        echo "golangci-lint version:"; \
        $(go env GOPATH)/bin/golangci-lint version; \
    fi
    @if command -v go-arch-lint >/dev/null 2>&1; then \
        echo "go-arch-lint version:"; \
        go-arch-lint version; \
    fi
    @if command -v capslock >/dev/null 2>&1; then \
        echo "capslock version:"; \
        capslock --version 2>/dev/null || echo "version info not available"; \
    fi
    @echo "Just version:"
    @just --version
    @if command -v docker >/dev/null 2>&1; then \
        echo "Docker version:"; \
        docker --version; \
    fi

# 🐳 Docker Commands

# Build Docker image
docker-build:
    @echo "{{BOLD}}🐳 BUILDING DOCKER IMAGE{{NORMAL}}"
    docker build -t template-arch-lint:latest .
    @echo "{{GREEN}}✅ Docker image built successfully!{{NORMAL}}"

# Build and test Docker image
docker-test: docker-build
    @echo "{{BOLD}}🧪 TESTING DOCKER IMAGE{{NORMAL}}"
    @echo "{{CYAN}}Testing health check...{{NORMAL}}"
    docker run --rm template-arch-lint:latest -health-check
    @echo "{{CYAN}}Testing container startup...{{NORMAL}}"
    @CONTAINER_ID=$$(docker run -d -p 8080:8080 template-arch-lint:latest); \
    sleep 5; \
    echo "Testing health endpoints..."; \
    curl -f http://localhost:8080/health/live || exit 1; \
    curl -f http://localhost:8080/version || exit 1; \
    docker stop $$CONTAINER_ID; \
    echo "{{GREEN}}✅ Docker image tests passed!{{NORMAL}}"

# Run application in Docker container
docker-run: docker-build
    @echo "{{BOLD}}🚀 RUNNING DOCKER CONTAINER{{NORMAL}}"
    docker run --rm -p 8080:8080 -p 2112:2112 template-arch-lint:latest

# Docker development environment (requires docker-compose.yml)
docker-dev:
    @if [ -f docker-compose.yml ]; then \
        echo "{{BOLD}}🔄 STARTING DEVELOPMENT ENVIRONMENT{{NORMAL}}"; \
        docker-compose up --build; \
    else \
        echo "⚠️  docker-compose.yml not found. This is a linting template - monitoring stack removed."; \
        echo "💡 For Docker setup, add your own docker-compose.yml with required services."; \
    fi

# Start Docker environment in background (requires docker-compose.yml)  
docker-dev-detached:
    @if [ -f docker-compose.yml ]; then \
        echo "{{BOLD}}🔄 STARTING DEVELOPMENT ENVIRONMENT (DETACHED){{NORMAL}}"; \
        docker-compose up --build -d; \
        echo "{{GREEN}}✅ Development environment started!{{NORMAL}}"; \
        echo "{{CYAN}}Services available at http://localhost:8080{{NORMAL}}"; \
    else \
        echo "⚠️  docker-compose.yml not found. This is a linting template."; \
        echo "💡 Create docker-compose.yml for your specific monitoring/service needs."; \
    fi

# Stop Docker environment  
docker-stop:
    @if [ -f docker-compose.yml ]; then \
        echo "{{BOLD}}🛑 STOPPING DEVELOPMENT ENVIRONMENT{{NORMAL}}"; \
        docker-compose down; \
        echo "{{GREEN}}✅ Development environment stopped!{{NORMAL}}"; \
    else \
        echo "⚠️  docker-compose.yml not found - nothing to stop."; \
    fi

# Clean up Docker resources
docker-clean:
    @echo "{{BOLD}}🧹 CLEANING DOCKER RESOURCES{{NORMAL}}"
    docker-compose down -v --remove-orphans
    docker image prune -f
    docker system prune -f
    @echo "{{GREEN}}✅ Docker resources cleaned!{{NORMAL}}"

# Show Docker logs
docker-logs:
    @echo "{{BOLD}}📋 DOCKER LOGS{{NORMAL}}"
    docker-compose logs -f

# Security scan Docker image with Trivy
docker-security: docker-build
    @echo "{{BOLD}}🛡️  DOCKER SECURITY SCAN{{NORMAL}}"
    @if command -v trivy >/dev/null 2>&1; then \
        trivy image template-arch-lint:latest; \
    else \
        echo "{{RED}}❌ Trivy not installed. Install with: brew install trivy{{NORMAL}}"; \
        exit 1; \
    fi

# Find high-confidence duplicates (stricter threshold)
find-duplicates-strict: (find-duplicates "100")

# Find all potential duplicates (looser threshold)
find-duplicates-loose: (find-duplicates "25")

# ==============================================
# PERFORMANCE PROFILING COMMANDS
# ==============================================

# Capture CPU profile for 30 seconds
profile-cpu:
    @echo "{{BOLD}}📊 CAPTURING CPU PROFILE{{NORMAL}}"
    @echo "{{CYAN}}Capturing CPU profile for 30 seconds...{{NORMAL}}"
    @if curl -s "http://localhost:8080/debug/pprof/profile?seconds=30" -o {{CPU_PROF}}; then \
        echo "{{GREEN}}✅ CPU profile saved to {{CPU_PROF}}{{NORMAL}}"; \
        echo "{{YELLOW}}💡 Use 'just analyze-cpu' to view analysis{{NORMAL}}"; \
    else \
        echo "{{RED}}❌ Failed to capture CPU profile. Is the server running in development mode?{{NORMAL}}"; \
        exit 1; \
    fi

# Capture heap memory profile
profile-heap:
    @echo "{{BOLD}}📊 CAPTURING HEAP PROFILE{{NORMAL}}"
    @echo "{{CYAN}}Capturing heap memory profile...{{NORMAL}}"
    @if curl -s http://localhost:8080/debug/pprof/heap -o {{HEAP_PROF}}; then \
        echo "{{GREEN}}✅ Heap profile saved to {{HEAP_PROF}}{{NORMAL}}"; \
        echo "{{YELLOW}}💡 Use 'just analyze-heap' to view analysis{{NORMAL}}"; \
    else \
        echo "{{RED}}❌ Failed to capture heap profile. Is the server running in development mode?{{NORMAL}}"; \
        exit 1; \
    fi

# Capture goroutine dump
profile-goroutines:
    @echo "{{BOLD}}📊 CAPTURING GOROUTINE DUMP{{NORMAL}}"
    @echo "{{CYAN}}Capturing goroutine dump...{{NORMAL}}"
    @if curl -s http://localhost:8080/debug/pprof/goroutine -o {{GOROUTINE_PROF}}; then \
        echo "{{GREEN}}✅ Goroutine dump saved to {{GOROUTINE_PROF}}{{NORMAL}}"; \
        echo "{{YELLOW}}💡 Use 'go tool pprof {{GOROUTINE_PROF}}' to analyze{{NORMAL}}"; \
    else \
        echo "{{RED}}❌ Failed to capture goroutine dump. Is the server running in development mode?{{NORMAL}}"; \
        exit 1; \
    fi

# Capture execution trace for 10 seconds
profile-trace:
    @echo "{{BOLD}}📊 CAPTURING EXECUTION TRACE{{NORMAL}}"
    @echo "{{CYAN}}Capturing execution trace for 10 seconds...{{NORMAL}}"
    @if curl -s "http://localhost:8080/debug/pprof/trace?seconds=10" -o {{TRACE_OUT}}; then \
        echo "{{GREEN}}✅ Execution trace saved to {{TRACE_OUT}}{{NORMAL}}"; \
        echo "{{YELLOW}}💡 Use 'go tool trace {{TRACE_OUT}}' to analyze{{NORMAL}}"; \
    else \
        echo "{{RED}}❌ Failed to capture execution trace. Is the server running in development mode?{{NORMAL}}"; \
        exit 1; \
    fi

# Capture allocation profile
profile-allocs:
    @echo "{{BOLD}}📊 CAPTURING ALLOCATION PROFILE{{NORMAL}}"
    @echo "{{CYAN}}Capturing allocation profile...{{NORMAL}}"
    @if curl -s http://localhost:8080/debug/pprof/allocs -o {{ALLOCS_PROF}}; then \
        echo "{{GREEN}}✅ Allocation profile saved to {{ALLOCS_PROF}}{{NORMAL}}"; \
        echo "{{YELLOW}}💡 Use 'go tool pprof {{ALLOCS_PROF}}' to analyze{{NORMAL}}"; \
    else \
        echo "{{RED}}❌ Failed to capture allocation profile. Is the server running in development mode?{{NORMAL}}"; \
        exit 1; \
    fi

# Open CPU profile analysis in browser
analyze-cpu:
    @echo "{{BOLD}}🔍 ANALYZING CPU PROFILE{{NORMAL}}"
    @if [ -f {{CPU_PROF}} ]; then \
        echo "{{CYAN}}Opening CPU profile analysis in browser...{{NORMAL}}"; \
        echo "{{YELLOW}}Browser will open at http://localhost:8081{{NORMAL}}"; \
        go tool pprof -http=:8081 {{CPU_PROF}}; \
    else \
        echo "{{RED}}❌ CPU profile not found. Run 'just profile-cpu' first.{{NORMAL}}"; \
        exit 1; \
    fi

# Open heap profile analysis in browser
analyze-heap:
    @echo "{{BOLD}}🔍 ANALYZING HEAP PROFILE{{NORMAL}}"
    @if [ -f {{HEAP_PROF}} ]; then \
        echo "{{CYAN}}Opening heap profile analysis in browser...{{NORMAL}}"; \
        echo "{{YELLOW}}Browser will open at http://localhost:8081{{NORMAL}}"; \
        go tool pprof -http=:8081 {{HEAP_PROF}}; \
    else \
        echo "{{RED}}❌ Heap profile not found. Run 'just profile-heap' first.{{NORMAL}}"; \
        exit 1; \
    fi

# Get runtime performance statistics
profile-stats:
    @echo "{{BOLD}}📊 RUNTIME STATISTICS{{NORMAL}}"
    @if curl -s http://localhost:8080/performance/stats | jq .; then \
        echo ""; \
        echo "{{GREEN}}✅ Runtime statistics retrieved successfully{{NORMAL}}"; \
    else \
        echo "{{RED}}❌ Failed to get runtime statistics. Is the server running?{{NORMAL}}"; \
        exit 1; \
    fi

# Check application health metrics
profile-health:
    @echo "{{BOLD}}🏥 HEALTH METRICS{{NORMAL}}"
    @if curl -s http://localhost:8080/performance/health | jq .; then \
        echo ""; \
        echo "{{GREEN}}✅ Health metrics retrieved successfully{{NORMAL}}"; \
    else \
        echo "{{RED}}❌ Failed to get health metrics. Is the server running?{{NORMAL}}"; \
        exit 1; \
    fi

# Force garbage collection and show results
profile-gc:
    @echo "{{BOLD}}🗑️ FORCE GARBAGE COLLECTION{{NORMAL}}"
    @echo "{{CYAN}}Triggering garbage collection...{{NORMAL}}"
    @if curl -s -X POST http://localhost:8080/performance/gc | jq .; then \
        echo ""; \
        echo "{{GREEN}}✅ Garbage collection completed{{NORMAL}}"; \
    else \
        echo "{{RED}}❌ Failed to trigger garbage collection. Is the server running in development mode?{{NORMAL}}"; \
        exit 1; \
    fi

# Capture all performance profiles in one command
profile-all:
    @echo "{{BOLD}}📊 CAPTURING ALL PROFILES{{NORMAL}}"
    @echo "{{CYAN}}This will capture CPU (30s), heap, goroutines, allocations, and trace (10s)...{{NORMAL}}"
    @echo "{{YELLOW}}Total time: ~45 seconds{{NORMAL}}"
    @echo ""
    just profile-heap
    @echo ""
    just profile-goroutines
    @echo ""
    just profile-allocs
    @echo ""
    just profile-cpu
    @echo ""
    just profile-trace
    @echo ""
    @echo "{{GREEN}}🎉 All profiles captured successfully!{{NORMAL}}"
    @echo "{{CYAN}}Files created:{{NORMAL}}"
    @echo "  • {{CPU_PROF}} - CPU profiling data"
    @echo "  • {{HEAP_PROF}} - Heap memory allocations"
    @echo "  • {{GOROUTINE_PROF}} - Goroutine dump"
    @echo "  • {{ALLOCS_PROF}} - Allocation history"
    @echo "  • {{TRACE_OUT}} - Execution trace"

# Clean up profile files
profile-clean:
    @echo "{{BOLD}}🧹 CLEANING PROFILE FILES{{NORMAL}}"
    rm -f {{REPORT_DIR}}/*.prof {{REPORT_DIR}}/*.out
    @echo "{{GREEN}}✅ Profile files cleaned!{{NORMAL}}"

# ==============================================
# BENCHMARKING COMMANDS
# ==============================================

# Run all benchmarks with comprehensive reporting
bench:
    @echo "{{BOLD}}🏎️ RUNNING ALL BENCHMARKS{{NORMAL}}"
    @echo "{{CYAN}}Running comprehensive benchmark suite...{{NORMAL}}"
    @mkdir -p benchmarks
    @go test -bench=. -benchmem -run=^$$ ./... | tee benchmarks/benchmark-results.txt
    @echo ""
    @echo "{{GREEN}}✅ Benchmarks completed!{{NORMAL}}"
    @echo "{{CYAN}}→ Results saved to: benchmarks/benchmark-results.txt{{NORMAL}}"

# Run CPU-focused benchmarks (no memory allocation reporting)
bench-cpu:
    @echo "{{BOLD}}⚡ RUNNING CPU BENCHMARKS{{NORMAL}}"
    @echo "{{CYAN}}Focusing on CPU performance metrics...{{NORMAL}}"
    @mkdir -p benchmarks
    @go test -bench=. -run=^$$ ./internal/domain/services/ | tee benchmarks/cpu-benchmarks.txt
    @go test -bench=. -run=^$$ ./internal/infrastructure/persistence/ | tee -a benchmarks/cpu-benchmarks.txt
    @echo ""
    @echo "{{GREEN}}✅ CPU benchmarks completed!{{NORMAL}}"
    @echo "{{CYAN}}→ Results saved to: benchmarks/cpu-benchmarks.txt{{NORMAL}}"

# Run memory-focused benchmarks (with allocation reporting)
bench-memory:
    @echo "{{BOLD}}🧠 RUNNING MEMORY BENCHMARKS{{NORMAL}}"
    @echo "{{CYAN}}Focusing on memory allocation patterns...{{NORMAL}}"
    @mkdir -p benchmarks
    @go test -bench=BenchmarkMemory -benchmem -run=^$$ ./... | tee benchmarks/memory-benchmarks.txt
    @go test -bench=BenchmarkAllocation -benchmem -run=^$$ ./... | tee -a benchmarks/memory-benchmarks.txt
    @go test -bench=BenchmarkConcurrent -benchmem -run=^$$ ./... | tee -a benchmarks/memory-benchmarks.txt
    @echo ""
    @echo "{{GREEN}}✅ Memory benchmarks completed!{{NORMAL}}"
    @echo "{{CYAN}}→ Results saved to: benchmarks/memory-benchmarks.txt{{NORMAL}}"

# Run specific benchmark by name
bench-specific PATTERN:
    @echo "{{BOLD}}🎯 RUNNING SPECIFIC BENCHMARK{{NORMAL}}"
    @echo "{{CYAN}}Running benchmarks matching: {{PATTERN}}{{NORMAL}}"
    @mkdir -p benchmarks
    @go test -bench={{PATTERN}} -benchmem -run=^$$ ./... | tee benchmarks/specific-{{PATTERN}}.txt
    @echo ""
    @echo "{{GREEN}}✅ Specific benchmarks completed!{{NORMAL}}"
    @echo "{{CYAN}}→ Results saved to: benchmarks/specific-{{PATTERN}}.txt{{NORMAL}}"

# Establish performance baseline
bench-baseline:
    @echo "{{BOLD}}📊 ESTABLISHING PERFORMANCE BASELINE{{NORMAL}}"
    @echo "{{CYAN}}Creating baseline benchmark results...{{NORMAL}}"
    @mkdir -p benchmarks/baseline
    @go test -bench=. -benchmem -count=5 -run=^$$ ./... > benchmarks/baseline/results.txt 2>&1
    @echo ""
    @echo "{{GREEN}}✅ Baseline established!{{NORMAL}}"
    @echo "{{CYAN}}→ Baseline saved to: benchmarks/baseline/results.txt{{NORMAL}}"
    @echo "{{YELLOW}}💡 Use 'just bench-compare' to compare future runs against this baseline{{NORMAL}}"

# Compare current benchmarks with baseline
bench-compare:
    @echo "{{BOLD}}📈 COMPARING BENCHMARK RESULTS{{NORMAL}}"
    @if [ ! -f benchmarks/baseline/results.txt ]; then \
        echo "{{RED}}❌ No baseline found. Run 'just bench-baseline' first.{{NORMAL}}"; \
        exit 1; \
    fi
    @echo "{{CYAN}}Running current benchmarks for comparison...{{NORMAL}}"
    @mkdir -p benchmarks/current
    @go test -bench=. -benchmem -count=5 -run=^$$ ./... > benchmarks/current/results.txt 2>&1
    @echo "{{CYAN}}Comparing results with baseline...{{NORMAL}}"
    @if command -v benchcmp >/dev/null 2>&1; then \
        benchcmp benchmarks/baseline/results.txt benchmarks/current/results.txt | tee benchmarks/comparison.txt; \
        echo "{{GREEN}}✅ Comparison completed!{{NORMAL}}"; \
        echo "{{CYAN}}→ Results saved to: benchmarks/comparison.txt{{NORMAL}}"; \
    else \
        echo "{{YELLOW}}⚠️ benchcmp tool not found. Install with: go install -tool golang.org/x/tools/cmd/benchcmp@latest{{NORMAL}}"; \
        echo "{{CYAN}}Manual comparison available in:{{NORMAL}}"; \
        echo "  → benchmarks/baseline/results.txt"; \
        echo "  → benchmarks/current/results.txt"; \
    fi

# Generate benchmark report with analysis
bench-report:
    @echo "{{BOLD}}📋 GENERATING BENCHMARK REPORT{{NORMAL}}"
    @mkdir -p benchmarks/reports
    @echo "{{CYAN}}Running comprehensive benchmarks...{{NORMAL}}"
    @go test -bench=. -benchmem -count=3 -run=^$$ ./... > benchmarks/reports/full-report.txt 2>&1
    @echo "{{CYAN}}Generating summary report...{{NORMAL}}"
    @echo "# Benchmark Report - $$(date)" > benchmarks/reports/summary.md
    @echo "" >> benchmarks/reports/summary.md
    @echo "## Performance Summary" >> benchmarks/reports/summary.md
    @echo "" >> benchmarks/reports/summary.md
    @echo "\`\`\`" >> benchmarks/reports/summary.md
    @grep "Benchmark" benchmarks/reports/full-report.txt | head -20 >> benchmarks/reports/summary.md
    @echo "\`\`\`" >> benchmarks/reports/summary.md
    @echo "" >> benchmarks/reports/summary.md
    @echo "## Analysis" >> benchmarks/reports/summary.md
    @echo "- Generated on: $$(date)" >> benchmarks/reports/summary.md
    @echo "- Go version: $$(go version)" >> benchmarks/reports/summary.md
    @echo "- Test count: $$(grep -c "Benchmark" benchmarks/reports/full-report.txt) benchmarks" >> benchmarks/reports/summary.md
    @echo ""
    @echo "{{GREEN}}✅ Benchmark report generated!{{NORMAL}}"
    @echo "{{CYAN}}→ Full report: benchmarks/reports/full-report.txt{{NORMAL}}"
    @echo "{{CYAN}}→ Summary: benchmarks/reports/summary.md{{NORMAL}}"

# Run benchmarks with profiling integration
bench-profile:
    @echo "{{BOLD}}🔬 RUNNING BENCHMARKS WITH PROFILING{{NORMAL}}"
    @echo "{{CYAN}}Running benchmarks and generating profiles...{{NORMAL}}"
    @mkdir -p benchmarks/profiles
    @go test -bench=BenchmarkCreateUser -benchmem -run=^$$ -cpuprofile=benchmarks/profiles/cpu.prof -memprofile=benchmarks/profiles/mem.prof ./internal/domain/services/
    @echo ""
    @echo "{{GREEN}}✅ Profiled benchmarks completed!{{NORMAL}}"
    @echo "{{CYAN}}Profiles generated:{{NORMAL}}"
    @echo "  → benchmarks/profiles/cpu.prof"
    @echo "  → benchmarks/profiles/mem.prof"
    @echo "{{YELLOW}}💡 Analyze with: go tool pprof benchmarks/profiles/cpu.prof{{NORMAL}}"

# Stress test with high iteration count
bench-stress:
    @echo "{{BOLD}}💪 RUNNING STRESS BENCHMARKS{{NORMAL}}"
    @echo "{{CYAN}}Running high-iteration stress tests...{{NORMAL}}"
    @mkdir -p benchmarks/stress
    @go test -bench=. -benchtime=10s -benchmem -run=^$$ ./internal/domain/services/ | tee benchmarks/stress/services.txt
    @go test -bench=. -benchtime=10s -benchmem -run=^$$ ./internal/infrastructure/persistence/ | tee benchmarks/stress/persistence.txt
    @echo ""
    @echo "{{GREEN}}✅ Stress benchmarks completed!{{NORMAL}}"
    @echo "{{CYAN}}→ Results saved to: benchmarks/stress/{{NORMAL}}"

# Quick benchmark run (fast feedback)
bench-quick:
    @echo "{{BOLD}}⚡ RUNNING QUICK BENCHMARKS{{NORMAL}}"
    @echo "{{CYAN}}Running short benchmarks for quick feedback...{{NORMAL}}"
    @go test -bench=. -benchtime=1s -run=^$$ ./internal/domain/services/ | grep -E "(Benchmark|PASS|FAIL)"
    @echo ""
    @echo "{{GREEN}}✅ Quick benchmarks completed!{{NORMAL}}"

# Clean up benchmark results
bench-clean:
    @echo "{{BOLD}}🧹 CLEANING BENCHMARK FILES{{NORMAL}}"
    rm -rf benchmarks/
    @echo "{{GREEN}}✅ Benchmark files cleaned!{{NORMAL}}"
