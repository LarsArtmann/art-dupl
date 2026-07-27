# Operational Zero False-Positives Workflow

This guide describes the recommended workflow for achieving **operational zero false-positives**
with art-dupl — every reported clone is actionable; residual judgment calls are tagged once
via `//art-dupl:accept` directive and honored forever.

## Quick Start

```bash
# 1. Run with type-aware mode for best classification
art-dupl --type-aware --semantic -t 3 .

# 2. Review LowConfidence clones (ambiguous — your judgment needed)
art-dupl --type-aware --semantic -t 3 --explain .

# 3. Accept residual false-positives with inline directives
# Add //art-dupl:accept above any clone you've reviewed and decided is not harmful

# 4. Re-run — accepted clones are suppressed automatically
art-dupl --type-aware --semantic -t 3 .
```

## The Three Confidence Tiers

| Tier | Confidence | Meaning | Action |
| --- | --- | --- | --- |
| **Actionable** | ≥ 0.8 | Real duplication — extract it | Extract to shared helper |
| **LowConfidence** | 0.5–0.8 | Ambiguous — needs human judgment | Review, then accept or extract |
| **NonActionable** | < 0.5 | Go idiom/boilerplate | Automatically suppressed |

## The Accept Directive

```go
// Place on the line above or within the clone's line range:

//art-dupl:accept idiomatic boilerplate
func (h *Handler) serve(w http.ResponseWriter, r *http.Request) {
    // ...
}

// Or as a trailing inline comment:
result := process(data) //art-dupl:accept unique per-site
```

The directive suppresses the clone group permanently. It is recognized in both compact
(`//art-dupl:accept`) and gofmt-canonical (`// art-dupl:accept`) forms.

## Property-Based Classification

art-dupl evaluates 4 computable properties to determine if a clone is harmful:

1. **Mechanically extractable** — always true
2. **Control-flow extractable** — false when returns are forced by void signatures
3. **ROI positive** — false when extraction costs more than it saves
4. **Parameterizable** — false when clones differ only in domain-value string literals

Use `--explain` to see which property classified each clone.

## The -t 1 Escape Hatch

`-t 1` is an explicit "show me everything" mode. It will always be noisy (matching single
statements). Use `-t 3` to `-t 5` for production-quality reports with meaningful noise filtering.
