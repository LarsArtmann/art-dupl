# Monorepo Migration Estimate

## Executive Summary

| Metric | Value |
|--------|-------|
| Total Go Files | 192 |
| Total Lines of Code | ~38,000 |
| Current Packages | 25+ |
| Estimated Modules | 6-8 |
| **Estimated Effort** | **20-40 hours** |
| Risk Level | Medium |

---

## Current State Analysis

### Package Size Distribution

| Package | Files | Lines | Category |
|---------|-------|-------|----------|
| bdd | 18 | 6,448 | Tests |
| domain | 22 | 3,961 | Core |
| pkg/ | 19 | 4,290 | SDK |
| cmd | 13 | 2,562 | CLI |
| syntax | 15 | 3,556 | Algorithm |
| internal | 22 | 3,146 | Internal |
| printer | 28 | 3,450 | Output |
| detection | 5 | 1,400 | Core |
| errors | 5 | 878 | Core |
| git | 2 | 977 | Utility |
| cache | 2 | 809 | Core |
| examples | 3 | 772 | Docs |
| suffixtree | 6 | 1,137 | Algorithm |
| config | 5 | 1,233 | Config |
| job | 7 | 693 | Core |
| hash | 3 | 467 | Algorithm |
| migration | 3 | 482 | Utility |
| cli | 6 | 589 | CLI |
| adapter | 3 | 573 | Adapter |
| lib | 2 | 179 | Legacy |

### Dependency Graph (Simplified)

```
                    ┌─────────┐
                    │ errors  │
                    └────┬────┘
                         │
         ┌───────────────┼───────────────┐
         │               │               │
    ┌────┴────┐    ┌────┴────┐    ┌────┴────┐
    │ domain  │    │ pkg/pos │    │ internal│
    └────┬────┘    └────┬────┘    └────┬────┘
         │               │               │
         └───────────────┼───────────────┘
                         │
              ┌──────────┼──────────┐
              │          │          │
         ┌────┴────┐ ┌───┴───┐ ┌────┴────┐
         │ syntax  │ │ cache │ │  utils  │
         └────┬────┘ └───┬───┘ └────┬────┘
              │          │          │
    ┌─────────┴──────────┴──────────┴─────────┐
    │                                          │
    │    suffixtree    hash    config         │
    │         │         │         │           │
    │         └─────────┼─────────┘           │
    │                   │                     │
    │              detection                  │
    │                   │                     │
    │         ┌─────────┴─────────┐           │
    │         │                   │           │
    │       job               printer         │
    │         │                   │           │
    │         └─────────┬─────────┘           │
    │                   │                     │
    │           pkg/artdupl                   │
    │                   │                     │
    │         ┌─────────┴─────────┐           │
    │         │                   │           │
    │       lib                  cmd          │
    │                           CLI           │
    └──────────────────────────────────────────┘
```

### Key Findings

1. **Clean Layering**: The project already has good separation between:
   - Core domain (`domain/`, `errors/`)
   - Algorithms (`suffixtree/`, `hash/`, `syntax/`)
   - Infrastructure (`config/`, `cache/`, `printer/`)
   - Application (`cmd/`, `cli/`)
   - SDK (`pkg/artdupl/`)

2. **Dependency Issues**:
   - `domain` → `syntax` (circular potential)
   - `syntax` → `suffixtree` (tight coupling)
   - `pkg/artdupl` imports many internal packages

3. **Good Patterns**:
   - `internal/` properly used for private utilities
   - `pkg/` for public SDK
   - Clear separation of concerns

---

## Proposed Module Structure

```
art-dupl/
├── go.work                          # Workspace definition
│
├── modules/
│   ├── core/                        # Core domain & errors
│   │   ├── go.mod
│   │   ├── domain/                  # Value objects, entities
│   │   ├── errors/                  # Error types
│   │   └── types/                   # Result[T], Option[T]
│   │
│   ├── algorithms/                  # Detection algorithms
│   │   ├── go.mod
│   │   ├── suffixtree/              # Suffix tree implementation
│   │   ├── hash/                    # Rolling hash detection
│   │   └── syntax/                  # AST processing
│   │
│   ├── infrastructure/              # Cross-cutting concerns
│   │   ├── go.mod
│   │   ├── config/                  # Configuration
│   │   ├── cache/                   # File caching
│   │   └── printer/                 # Output formatting
│   │
│   ├── sdk/                         # Public API
│   │   ├── go.mod
│   │   └── pkg/artdupl/             # Programmatic API
│   │
│   └── internal/                    # Private utilities
│       ├── go.mod
│       ├── enum/
│       ├── simd/
│       ├── testutil/
│       ├── treesitter/
│       └── utils/
│
├── apps/
│   └── cli/                         # CLI application
│       ├── go.mod
│       └── cmd/art-dupl/
│
├── tests/                           # Integration tests
│   ├── go.mod
│   ├── bdd/                         # BDD tests
│   ├── internal/configtest/
│   └── internal/filtertest/
│
└── examples/                        # Usage examples
    ├── go.mod
    └── examples/
```

### Module Dependency Graph

```
          ┌─────────┐
          │  core   │
          └────┬────┘
               │
        ┌──────┴──────┐
        │             │
   ┌────┴────┐   ┌────┴─────┐
   │ internal│   │algorithms│
   └────┬────┘   └────┬─────┘
        │             │
        └──────┬──────┘
               │
        ┌──────┴──────┐
        │             │
   ┌────┴────────┐    │
   │infrastructure│   │
   └────┬────────┘    │
        │             │
        └──────┬──────┘
               │
          ┌────┴────┐
          │   sdk   │
          └────┬────┘
               │
          ┌────┴────┐
          │   cli   │
          └─────────┘
```

---

## Effort Breakdown

### Phase 1: Analysis & Planning (4-6 hours)

| Task | Hours | Complexity |
|------|-------|------------|
| Map all package dependencies | 1-2 | Medium |
| Identify circular dependencies | 1 | Low |
| Design module boundaries | 2-3 | High |
| Create migration plan | 1 | Medium |

### Phase 2: Module Setup (4-6 hours)

| Task | Hours | Complexity |
|------|-------|------------|
| Create go.work file | 0.5 | Low |
| Create module go.mod files | 1 | Low |
| Set up replace directives | 1 | Medium |
| Configure CI for modules | 2-3 | Medium |

### Phase 3: Code Migration (8-12 hours)

| Task | Hours | Complexity |
|------|-------|------------|
| Move packages to modules | 3-4 | Medium |
| Update import paths | 3-4 | High |
| Resolve circular dependencies | 2-4 | High |

### Phase 4: Testing & Validation (4-6 hours)

| Task | Hours | Complexity |
|------|-------|------------|
| Update test imports | 1-2 | Medium |
| Fix broken tests | 2-3 | Medium |
| Verify all builds | 1 | Low |

### Phase 5: Documentation & Polish (2-4 hours)

| Task | Hours | Complexity |
|------|-------|------------|
| Update README | 0.5 | Low |
| Update AGENTS.md | 0.5 | Low |
| Create migration guide | 1-2 | Medium |
| Update justfile/makefile | 0.5 | Low |
| Clean up old files | 0.5 | Low |

---

## Risk Assessment

### High Risk

| Risk | Impact | Mitigation |
|------|--------|------------|
| Circular dependencies | Blocks migration | Break dependency cycles first |
| Import path changes | Breaks all code | Use find/replace carefully |
| CI/CD changes | Deployment issues | Test CI thoroughly |

### Medium Risk

| Risk | Impact | Mitigation |
|------|--------|------------|
| Test breakage | Delayed completion | Run tests after each change |
| Version conflicts | Build failures | Pin versions explicitly |
| Documentation drift | User confusion | Update docs immediately |

### Low Risk

| Risk | Impact | Mitigation |
|------|--------|------------|
| IDE configuration | Developer experience | Update .vscode, etc. |
| Git history clarity | Code review | Use clear commit messages |

---

## Recommendations

### Option A: Full Monorepo Migration (20-40 hours)

**Pros:**
- Clean module boundaries
- Independent versioning
- Better code organization
- Reusable components

**Cons:**
- Significant effort
- Risk of breakage
- Learning curve for team

### Option B: Minimal Workspace (8-12 hours)

Create a simple go.work without moving packages:

**Pros:**
- Quick to implement
- Low risk
- Enables local development

**Cons:**
- Still single module
- Less architectural benefit

### Option C: Status Quo (0 hours)

Keep current structure, add go.work only when needed.

**Pros:**
- No effort required
- Zero risk

**Cons:**
- Technical debt remains
- Limited scalability

---

## Decision Matrix

| Criteria | Option A | Option B | Option C |
|----------|----------|----------|----------|
| Effort | High (20-40h) | Medium (8-12h) | None |
| Risk | Medium | Low | None |
| Benefit | High | Medium | Low |
| Scalability | Excellent | Good | Limited |
| Maintainability | Excellent | Good | Fair |
| Team Adoption | Requires training | Easy | No change |

---

## Recommendation

**For immediate needs**: Option B (Minimal Workspace) - 8-12 hours
**For long-term architecture**: Option A (Full Monorepo) - 20-40 hours

If the project is expected to grow significantly or be split into multiple tools, invest in Option A now. If you just need local development improvements, Option B is sufficient.

---

## Next Steps

1. **Decision**: Choose migration option
2. **Branch**: Create `feature/monorepo-migration` branch
3. **Backup**: Ensure clean git state
4. **Execute**: Follow phase-by-phase plan
5. **Validate**: Run full test suite
6. **Document**: Update all documentation
7. **Review**: Team code review
8. **Merge**: Merge to main branch
