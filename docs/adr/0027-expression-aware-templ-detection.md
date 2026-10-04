# ADR-0027: Expression-Aware templ Detection

- **Status:** Accepted (2026-10-04)
- **Deciders:** Lars Artmann
- **Supersedes:** the markup-only templ tokenization contract (implicit since
  the templ parser was added; the retired regex normalizer in
  `syntax/templ/normalize.go`)
- **Motivated by:** `docs/status/2026-10-04_08-28_templ-expression-aware-detection.md`

## Context

The templ integration tokenized markup STRUCTURE only. Every embedded Go
expression — conditions, range clauses, switch tags, attribute values,
`{{ goCode }}` statements, component-call arguments, top-level Go
declarations — collapsed to a single constant token or vanished entirely.
Eight defect classes were confirmed by CLI probes before any code changed:

1. **`--exact` mode dropped ALL names.** A bool squeeze
   (`mode == Semantic`) fed the transformers, so exact mode could not
   distinguish `<div>` from `<section>` or `Card` from `Other`.
2. **Invisible conditions produced Type-1 false positives.**
   `if user.IsAdmin` matched `if user.IsCompletelyUnrelatedCondition` as an
   EXACT clone — the tool's worst failure mode on templ files.
3. **Else flattening.** `if X { A } else { B }` equaled `if X { A; B }`.
4. **Switch degeneracy.** `default:` equaled `case:`, and `fallthrough` was
   typed as a case body.
5. **Conditional attributes were opaque.** Condition and both branches
   invisible.
6. **CSS property names invisible**, and CSS declarations produced duplicate
   clone groups in output.
7. **Declaration hashing** covered the whole signature (`Panel(user User)`)
   instead of the declared name, and the regex normalizer aliased package
   qualifiers (`display.Card` → `v0.Card` — the old false-positive class).
8. **Clone categories** for templ nodes decoded against the parallel golang
   enum and landed in `unknown` (or a colliding Go category).

The root cause: templ's embedded Go fragments never entered the AST pipeline
at all, so no detection mode could see them.

## Decision

1. **Embedded Go parses through the real Go transformer.**
   `golang.ParseSnippet(kind, src, filename, fileOffset, mode)`
   (`syntax/golang/snippet.go`) wraps a fragment in a minimal synthetic Go
   file (padded so positions land at the real templ offsets), transforms it
   with a FRESH per-snippet normalizer, and shifts positions back
   (clamped ≥ 0). The emitted child tokens reuse the golang node-type space
   inside templ trees. Snippet kinds cover expressions, if conditions,
   for clauses, switch tags, case lists, statement blocks, and whole Go
   files (`TemplateFileGoExpression`; imports are dropped).

2. **API surface hashes VERBATIM; locals alpha-normalize.** Component/CSS/
   script declarations hash the declared NAME (`declarationName()`), callee
   paths in component calls hash verbatim (`display.Card` ≠
   `widgets.Card`), CSS property names and struct-literal keys are encoded.
   Locals inside snippets are canonicalized through
   `normalizer.collectSnippetLocals` (free lowercase-initial idents in
   source order; called functions, selector fields, composite-literal keys,
   and predeclared identifiers stay verbatim), so renamed locals still
   match as Type-2.

3. **Conservative fallback for hybrid sources.** A fragment templ tolerates
   but `go/parser` rejects becomes ONE opaque token hashed from the raw
   source: identical source still matches, divergent source does not.
   Parse-success-but-empty is silently dropped. No silent false positives.

4. **New templ node types sit ABOVE the golang range.** The two enums are
   parallel iota sequences (0–53 for Go); `ComponentElseStatement = 100`
   and `ComponentFallthroughStatement = 101` are deliberately above it so
   `syntax.IsStatementContainer` pins cannot collide with any Go type.
   Any future value-keyed switch over node types MUST be scoped by file
   kind — the compiler rejects merged switches via duplicate cases (this is
   how the category-table split is enforced).

5. **Detection mode is threaded, not squeezed.** `semantic bool` became
   `golang.DetectionMode`: exact mode distinguishes tags/names, structural
   ignores them, semantic alpha-normalizes locals. The regex normalizer
   (`normalizeExprValue`) is deleted.

6. **Category classification is file-kind-scoped.**
   `nodeTypeToCategory` dispatches to `goNodeTypeToCategory` or
   `templNodeTypeToCategory` (`printer/actionability/clone_classify.go`),
   and `templSuggestion` tailors refactoring wording to components.
   Regression pins: every templ type → category, the
   `golang.BasicLit == templ.Element == 4` collision
   (`TestNodeTypeToCategoryFileKindCollision`), and the behavior suite in
   `syntax/templ/behavior_test.go`.

7. **Structure fixes ride along:** else branches are `ComponentElseStatement`
   containers (statement children emit via `IsStatementContainer`),
   `default:` gets `ComponentSwitchDefaultCase`, fallthrough gets its own
   type, CSS declarations are statement composites (property names encoded),
   declarations hash names not signatures, and GoCode/`{{ ... }}`/string
   expressions splice parsed children.

8. **Default-on, CacheVersion bumped.** The blind spots are correctness
   defects in the tool's core promise (same rationale as ADR-0023
   Decision 4); a flag would keep default behavior wrong. `CacheVersion`
   is 5 (v4 entries would reproduce expression-blind templ results).

9. **Architecture boundaries updated.** `syntax-templ` mayDependOn
   `+syntax-golang`; `actionability` mayDependOn `+syntax-templ`
   (`.go-arch-lint.yml`).

## Alternatives Considered

- **Keep the regex normalizer and add more regex rules** — rejected: it was
  the source of defect 7 (package-qualifier aliasing), regex cannot see
  AST structure (conditions, arguments, nested calls), and every rule was a
  new false-positive class.
- **Hash raw expression STRINGS as single tokens** — rejected: loses Type-2
  detection inside expressions (`user.Name` vs `account.Name` would not
  match), and whitespace/ordering churn would churn hashes.
- **Structural-only markup matching forever** — rejected: defects 2–6 are
  false positives/false negatives on the tool's core promise; "templ files
  get weaker detection than Go files" is not a limitation, it is a defect.
- **Wrap fragments in REAL temp files via go/packages** — rejected: two
  orders of magnitude slower than parsing a padded string; snippets need no
  type information (type-aware mode is an opt-in Go-file feature).

## Consequences

- **Corpus shifts (documented 2026-10-04):** go-sse 17 shown at `-t 1`
  (baseline 16); go-cqrs-lite 347 vs baseline 411 at `-t 2` (suppressions
  now see structured templ/Go statements); templ-components 108 groups =
  10 true `_sources/` vendored-copy pairs. Art-dupl self-scan `-t 5`: 0
  shown (clone-free invariant holds).
- **Remaining known limits:** unparseable hybrid fragments degrade to one
  opaque token (documented); FuncLit bodies and LabeledStmt interiors stay
  composite-only (ADR-0023 scope, unchanged); GenDecl spec-level sharing
  still undetected.
- **Verification gates:** `syntax/templ/behavior_test.go` (token-stream
  level), `syntax/golang/snippet_test.go` (bridge), `syntax/templ/
  mode_plumbing_test.go` (mode threading),
  `printer/actionability/clone_classify_test.go` (categories/suggestions),
  BDD regression guard for the `if X` false-positive class.
