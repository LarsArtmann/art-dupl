# ADR-0020: Algorithmic alternatives analysis — suffix tree vs suffix array vs fingerprinting

## Status

Accepted (research decision — no code change, documents the analysis for future migration)

## Context

After implementing parallel search and memory-compact storage (ADR-0019), the question arose: **are there fundamentally better algorithms than Ukkonen's suffix tree for clone detection?** The original ROADMAP items ("parallel suffix tree construction" and "streaming suffix tree") were reframed, but the algorithmic question remained open.

A thorough literature review was conducted covering:

- Suffix arrays (SA-IS, DC3/skew) and Enhanced Suffix Arrays (ESA)
- Parallel suffix array construction (Shun-Blelloch, Labeit et al., CaPS-SA)
- Winnowing/fingerprinting (Schleimer et al., MOSS)
- Rolling-hash bucketing (PMD CPD / SonarQube)
- AST + LSH (Deckard)
- Inverted index (SourcererCC)

Full analysis: `docs/research/algorithmic-alternatives-for-clone-detection.md`

## Decision

### Keep Ukkonen's suffix tree for now. Plan suffix array migration as medium-term.

**Rationale:**

1. **The current implementation works.** 200K-token codebases use ~5-7 MB, well within memory limits. Parallel search gives 1.5-3.4x speedup. No urgency to replace a tested, understood algorithm.

2. **Suffix arrays are scientifically superior** for this use case:
   - 30-44% less memory (contiguous `[]int32` arrays vs pointer-heavy tree states)
   - 2-5x faster search (cache locality)
   - **Parallelizable construction** (the original ROADMAP goal, impossible with Ukkonen's)
   - Simpler implementation (~150 lines vs ~250 lines)
   - SA-IS is O(n) linear time, same as Ukkonen's

3. **The `[]*syntax.Node` (8N bytes) is the true bottleneck**, not the tree vs array choice. This is the irreducible floor for all approaches that need position-to-AST mapping. Eliminating it (per-file position offset map) is a higher-priority architectural improvement.

4. **No free lunch exists.** Every approach trades something:
   - Suffix tree: lossless, O(n), but pointer-heavy and sequential construction
   - Suffix array: lossless, O(n), cache-friendly, parallelizable, but batch-only (no online construction)
   - Winnowing: massively scalable, but lossy (misses clones below threshold t)
   - Rolling hash: simple, but different output semantics (k-gram matches, not maximal repeats)

### Migration path when ready

Implement SA-IS + LCP + maximal-repeat scan as a **new `MethodDetector`** alongside the existing suffix tree. The `detection.MethodDetector` interface makes this a clean addition, not a replacement. Benchmark both on real codebases. Switch when SA proves superior.

The migration is low-risk because:

- SA positions index into the same `[]*syntax.Node` array
- `FindSyntaxUnits` works unchanged (it takes `suffixtree.Match{Ps, Len}` — the `Ps` values are just positions)
- The `MethodDetector` interface abstracts the implementation
- Both detectors can coexist for A/B comparison

### Alphabet remapping requirement

SA-IS and DC3 require compact integer alphabets `[0, sigma)`. art-dupl's `TokenValue` range spans full `int32` (node types 1-100, statement fingerprints are FNV hashes, sentinels from `MinInt32/2`). A single O(n) remapping pass is needed before SA construction.

### What was ruled out

- **Go's `index/suffixarray`**: Doesn't expose SA values or LCP arrays. Designed for point lookups, not maximal-repeat enumeration. Would need reflection hacks.
- **Immediate winnowing adoption**: Lossy detection is a different product. Better as a future pre-filter for extreme-scale codebases.
- **Prioritizing parallel construction over `[]*syntax.Node` elimination**: The 8N node array dominates memory. Parallel construction saves wall-clock time but not memory.

## Consequences

- No code change now. Research document and this ADR serve as the decision record.
- ROADMAP.md updated with specific suffix-array migration item and per-file position offset map item.
- The suffix tree codebase is treated as stable — not invested in further optimization, but not deprecated.
- When the migration happens, it will be a new package (e.g., `suffixarray/`) implementing `MethodDetector`, not a modification to `suffixtree/`.
