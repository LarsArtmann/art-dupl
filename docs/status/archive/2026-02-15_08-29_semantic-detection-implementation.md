# Status Report: Semantic-Aware Duplicate Detection Implementation

**Date:** 2026-02-15 08:29
**Status:** In Progress
**Feature:** Content-aware (semantic) filtering for duplicate detection

---

## Overview

Implementing semantic-aware duplicate detection to eliminate false positives when structurally similar but semantically different code is flagged as duplicates. The classic example: Ginkgo test blocks testing completely different methods being flagged as duplicates due to identical AST structure.

### Problem Statement

Current matching uses only AST node types (e.g., `CallExpr`, `SelectorExpr`), ignoring identifiers and literals. This causes false positives:

```go
// These get flagged as duplicates (same structure):
Expect(user.Name).To(Equal("John"))
Expect(user.Age).To(Equal(30))
```

### Solution

Encode identifier names into the type hash using FNV-1a algorithm with 8/24 bit split:

- Lower 8 bits: base AST node type (0-52)
- Upper 24 bits: identifier name hash

---

## Completed Work

### 1. Core Implementation Files

| File                               | Purpose                                   | Status |
| ---------------------------------- | ----------------------------------------- | ------ |
| `syntax/golang/identifier_hash.go` | NEW: Hash functions for semantic encoding | DONE   |
| `syntax/golang/transform.go`       | Modified: SelectorExpr, Ident cases       | DONE   |
| `config/config.go`                 | Modified: Semantic field added            | DONE   |

### 2. Hash Function Design

**Algorithm:** FNV-1a (fast for short strings, ~15ns per call)

**Key Functions:**

```go
var SemanticHashEnabled bool  // Global toggle (default: false)

func hashIdentifierFast(name string) int32  // FNV-1a to 24 bits
func encodeSemanticType(baseType int32, identifierName string) int32
func DecodeBaseType(t int32) int32
func DecodeSemanticHash(t int32) int32
```

### 3. AST Node Modifications

| AST Node       | Line    | What's Captured         | Priority |
| -------------- | ------- | ----------------------- | -------- |
| `SelectorExpr` | 212-214 | `n.Sel.Name` encoded    | CRITICAL |
| `Ident`        | 158-159 | `n.Name` encoded        | CRITICAL |
| `CallExpr`     | 50-55   | Handled via child nodes | HIGH     |
| `BasicLit`     | 33-34   | Deferred (MEDIUM)       | MEDIUM   |

### 4. Configuration

Added to `config.Config`:

```go
// Semantic enables semantic-aware duplicate detection
// When true, identifier names are included in the type hash
Semantic bool `json:"semantic,omitempty"`
```

Default: `false` for backward compatibility.

---

## Remaining Work

### Phase 1: Core Wiring (~9 min)

| #   | Task                                             | Est | Status  |
| --- | ------------------------------------------------ | --- | ------- |
| 1   | Add `--semantic` flag to `cmd/flags.go`          | 3m  | PENDING |
| 2   | Read semantic flag in `run_flags.go`             | 2m  | PENDING |
| 3   | Wire `golang.SemanticHashEnabled = cfg.Semantic` | 3m  | PENDING |
| 4   | Add import for `golang` package                  | 1m  | PENDING |

**Location for wiring:** `cmd/run_flags.go:142` after `config.MergeConfigs()`

### Phase 2: Unit Tests (~23 min)

| #   | Task                                       | Est | Status  |
| --- | ------------------------------------------ | --- | ------- |
| 5   | Create `identifier_hash_test.go`           | 2m  | PENDING |
| 6   | Test `hashIdentifierFast` consistency      | 5m  | PENDING |
| 7   | Test `encodeSemanticType` bit manipulation | 5m  | PENDING |
| 8   | Test `DecodeBaseType`                      | 3m  | PENDING |
| 9   | Test `DecodeSemanticHash`                  | 3m  | PENDING |
| 10  | Test collision behavior                    | 5m  | PENDING |

### Phase 3: Integration Tests (~26 min)

| #   | Task                                     | Est | Status  |
| --- | ---------------------------------------- | --- | ------- |
| 11  | Create test fixture: Ginkgo patterns     | 5m  | PENDING |
| 12  | BDD test: semantic OFF (finds duplicate) | 8m  | PENDING |
| 13  | BDD test: semantic ON (no duplicate)     | 8m  | PENDING |
| 14  | Test config file with `semantic: true`   | 5m  | PENDING |

### Phase 4: Documentation (~21 min)

| #   | Task                              | Est | Status  |
| --- | --------------------------------- | --- | ------- |
| 15  | Update `cmd/root.go` with example | 5m  | PENDING |
| 16  | Update `AGENTS.md`                | 8m  | PENDING |
| 17  | Update `README.md`                | 8m  | PENDING |

### Phase 5: Validation (~16 min)

| #   | Task                          | Est | Status  |
| --- | ----------------------------- | --- | ------- |
| 18  | Run `just test`               | 5m  | PENDING |
| 19  | Run `just check`              | 3m  | PENDING |
| 20  | Manual test with `--semantic` | 3m  | PENDING |
| 21  | Final review and cleanup      | 5m  | PENDING |

---

## Architecture Decisions

| Decision           | Choice           | Rationale                                                |
| ------------------ | ---------------- | -------------------------------------------------------- |
| Two-stage approach | REJECTED         | Modify token stream directly - simpler, more efficient   |
| Hash algorithm     | FNV-1a over xxh3 | ~3x faster for short strings (<20 chars)                 |
| Toggle mechanism   | Global variable  | Zero-overhead when disabled (simple bool check)          |
| Bit layout         | 8/24 split       | 256 base types (using ~53), 16M unique identifier hashes |

---

## Key File Locations

| File                               | Lines            | Purpose                       |
| ---------------------------------- | ---------------- | ----------------------------- |
| `syntax/golang/identifier_hash.go` | Full             | NEW: Hash functions           |
| `syntax/golang/transform.go`       | 158-159, 212-214 | Modified: Ident, SelectorExpr |
| `config/config.go`                 | 118-126, 154     | Modified: Semantic field      |
| `cmd/flags.go`                     | TBD              | NEXT: Add --semantic flag     |
| `cmd/run_flags.go`                 | 142              | NEXT: Wire global variable    |

---

## Test Status

**Current State:**

- All `syntax/golang` tests pass
- Build succeeds
- One unrelated test failure in `job` package (profiler_test.go:93)

```bash
# Known issue (unrelated to semantic work):
--- FAIL: TestContextTimeoutExpired (profiler_test.go:93)
FAIL    github.com/LarsArtmann/art-dupl/job
```

---

## Next Actions

1. **Immediate:** Add `--semantic` flag to `cmd/flags.go`
2. **Then:** Wire flag to `golang.SemanticHashEnabled` in `run_flags.go`
3. **Then:** Write unit tests for hash functions
4. **Then:** Create BDD integration tests
5. **Finally:** Documentation and validation

---

## Estimated Completion

- **Total remaining:** ~95 minutes
- **Critical path:** Flag wiring → Tests → Documentation → Validation
- **Blocking issues:** None

---

## Usage Examples (Post-Implementation)

```bash
# Enable semantic detection
art-dupl --semantic ./src

# Combine with other flags
art-dupl --semantic -t 20 --json ./src

# Config file (dupl.json)
{
  "semantic": true,
  "threshold": 15
}
```

---

_Generated: 2026-02-15 08:29_
