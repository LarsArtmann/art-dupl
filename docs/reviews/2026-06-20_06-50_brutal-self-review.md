# Brutal Self-Review — 2026-06-20

## Summary

Found and fixed 10 categories of issues across the codebase: ghost systems, dead code, lying documentation, deprecated API, and type naming issues. 10 commits, 300+ lines of dead code removed.

## Self-Review Answers

### 1. What did you forget?

- **Lying documentation**: domain.go and go.mod referenced types/features that don't exist (BytePosition, Threshold, SIMD, CloneSeverity, SafeMarshalConfig, domain.NewThreshold)
- **Ghost branded types**: domain.Filepath and domain.LineNumber had full implementations but zero consumers
- **Dead Parse functions**: ParseClonePriority/Category/Actionability exported but never called
- **Dead error constructors**: 5 constructors only existed in test files
- **Deprecated API in interface**: FindClonesStream was deprecated but still required by Detector interface

### 2. What is something stupid?

- Keeping dead branded types that give false confidence about type safety
- Documentation claiming "SIMD-optimized" when we explicitly renamed hash_simd.go → hash_seq.go

### 3. What could you have done better?

- Should have caught these in the previous sprint's self-review
- The self-review between sprints was too focused on architecture and missed basic dead code

### 4. What could you still improve?

- Fragment type unification ([]byte vs string split brain)
- Clone type consolidation (5-7 parallel types)
- Data→View rename in printer
- Test coverage for detection (61.8%) and domain (63.5%)

### 5. Did you lie?

- The documentation was lying about types and features. Fixed.

### 6. How can we be less stupid?

- Run art-dupl on itself regularly to catch dead code
- Keep documentation honest — if a type doesn't exist, don't mention it
- Delete dead code immediately, don't leave it "for later"

### 7. Ghost systems?

- domain.Filepath/LineNumber — DELETED
- 5 dead error constructors — DELETED
- 3 dead Parse functions — DELETED
- ErrInvalidLineNumber — DELETED
- CloneClassification.NodeTypeName — DELETED
- ErrorType.String() — DELETED
- FindClonesStream — DELETED from interface

### 8. Scope creep?

- Correctly deferred: Fragment type unification, GeneratorFilter extraction, Clone type consolidation
- Correctly accepted: duplicate sentinels/noOpLogger (decoupling pattern)

### 9. Did we remove something useful?

- No. Every removed item was verified to have zero production callers.

### 10. Split brains?

- Fragment []byte vs string — real split brain, deferred (works at boundaries)
- Duplicate sentinels — ACCEPTED (expected decoupling pattern)
- Duplicate noOpLogger — ACCEPTED (expected decoupling pattern)

### 11. Tests?

- All 22 packages pass
- Fixed 12 unusedwrite + 3 infertypeargs diagnostics
- detection 61.8%, domain 63.5% — below 80% target, deferred

## Completed Work (10 commits)

1. Fix domain.go doc — remove 6 non-existent type references
2. Fix go.mod doc — remove SIMD lies, dead function references
3. Remove dead Parse functions + infertypeargs diagnostics
4. Remove dead ErrInvalidLineNumber + NodeTypeName field
5. Fix examples_test.go unusedwrite diagnostics
6. Remove dead error constructors + ErrorType constants + EnumValidationError
7. Remove deprecated FindClonesStream from Detector interface
8. Delete dead domain.Filepath/LineNumber branded types + helpers
9. Rename SARIFConfig → SARIFPrinterOptions
10. Update TODO_LIST, FEATURES, AGENTS documentation
