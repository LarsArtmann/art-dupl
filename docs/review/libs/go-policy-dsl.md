# Library Review: go-policy-dsl

| Field            | Value                                                                                        |
| ---------------- | -------------------------------------------------------------------------------------------- |
| Library          | `github.com/larsartmann/go-policy-dsl`                                                       |
| Version reviewed | `v0.4.0-1-gee88a84` (checkout `ee88a84a108e2b64eca5e461f89022e4350498c7`)                    |
| Consumer         | `github.com/LarsArtmann/art-dupl` @ `cee67b4a47a8e759f5af05669cad97dca8d1f22d`               |
| Date             | 2026-11-22                                                                                   |
| Auditor          | Group B fleet run                                                                            |
| Verdict          | **NOT-ADOPTED** — score **0/100** (0 of 8 applicable capabilities adopted; band: unrealized) |

**Repo state note:** the library checkout has 1 dirty file — `git -C ~/projects/go-policy-dsl status --short` reports `M .github/workflows/ci.yml`, an uncommitted `actions/setup-go` SHA bump (v7 → v7.1.0, `d8cbb804...`), cosmetic CI plumbing with no API impact. The single commit past the tag (`ee88a84`) is docs-only: it adds the `WhenAnyPresent` environment-gate documentation missed from the v0.4.0 changelog.

## What the library is

A fluent, compile-time-checked Go DSL for declaring _library governance policies_ — bans ("don't use gorm"), required companions ("if you depend on samber/do you must also depend on samber-do-auditlog"), and version-bounded or content-gated bans. Policies are plain Go values built via a `Builder` chain and terminated by `Spec()` into an immutable `PolicySpec` (`go-policy-dsl/builder.go:26`, `go-policy-dsl/policy.go:168`); there is no YAML, no codegen, no runtime parsing, and — critically — **no execution model**: `Detection` patterns (`ImportPatterns`, `GoModPatterns`, `ExcludeIfContains`, `RequireIfContains`, `ExcludeIfTransitiveFrom`, `WhenAnyPresent`, `go-policy-dsl/policy.go:86-118`) are opaque `[]string` whose matching semantics each consumer implements (`go-policy-dsl/README.md`, Design Decisions: "The DSL declares _what_ a policy is, not _how_ to detect violations"). Zero dependencies (stdlib only), panic-free by an in-repo regression guard, deny-by-default `Mode` so a typo can never silently disable a ban. The package doc names its canonical consumer: `library-policy` (`go-policy-dsl/policy.go:7`).

## Module map

- Single module `github.com/larsartmann/go-policy-dsl`, `go 1.27`, **zero require entries** (`go-policy-dsl/go.mod`) — a dependency would add nothing transitively, which is the DSL's stated adoption pitch.
- Packages: root `policydsl` only (`policy.go`, `builder.go`, `cve.go`); docs under `docs/DOMAIN_LANGUAGE.md`.
- art-dupl pinning: a plain `require github.com/larsartmann/go-policy-dsl v0.4.0` in `art-dupl/go.mod`. Being public and tagged, it follows art-dupl's public-dep upgrade path (`go get` → nix `vendorHash` refresh per `AGENTS.md:241`) — simpler than the gogenfilter private-dep pattern (`AGENTS.md:223-229`), with no module fan-out subtleties (unlike go-finding's independently-tagged submodules).

## Capability inventory

A = 8 applicable rows. Score = round(100 × Σ/A) with every applicable row at 0 (nothing adopted) → **0**.

| # | Capability                                                                                                               | Would apply to art-dupl                                                                                                                                         | Evidence (library)                                     | Evidence (art-dupl hand-roll / absence)                                                                                                                |
| - | ------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 1 | Fluent policy `Builder` (`Ban`/`Because`/`Spec()`)                                                                       | No — art-dupl declares no library-governance policies anywhere; nothing in its config, scripts, or CI bans a dependency                                         | `go-policy-dsl/builder.go:26-39`                       | Absence: `.golangci.yml` is linter curation, not library policy; no dep-scan tool exists in the repo                                                   |
| 2 | Branded domain types (`Severity`, `Category`, `Mode`, `CVE`, `Version`)                                                  | No — art-dupl's enum needs (DetectionMode, OutputFormat, SortCriteria) are its own domain, not governance enums                                                 | `go-policy-dsl/policy.go:20-84`                        | `art-dupl/domain/detection_mode.go`, `config/config.go` (own enums via `art-dupl/pkg/enum/enum.go`)                                                    |
| 3 | Detection declarations (import/go.mod patterns, content gates, transitive exclusions, `WhenAnyPresent` environment gate) | No — art-dupl has no matcher over imports or go.mod contents; its pattern matching is clone/markup generation filtering (gogenfilter), a different domain       | `go-policy-dsl/policy.go:86-118`                       | `cmd/util.go` (`generatorIncludes`, gogenfilter integration — file-generation filtering, not dependency policy)                                        |
| 4 | Structural validation (`PolicySpec.Validate()` → `*InvertedVersionRangeError`, `ErrInvertedVersionRange` sentinel)       | No — no version-range policies exist to validate                                                                                                                | `go-policy-dsl/policy.go:203,239`                      | art-dupl validates its own config domain: `config/config_validate.go:12,17` (`Validate`/`ValidateConfig` — threshold ranges, mode/method combinations) |
| 5 | Companion requirements (`RequiresCompanion`, `AsCompanionOnly`)                                                          | No — the one art-dupl "companion" rule (go-finding must accompany findings output) is a hard architectural boundary enforced by arch-lint, not a runtime policy | `go-policy-dsl/builder.go` (`RequiresCompanion`)       | `.go-arch-lint.yml:1-312` (component dependency boundaries: printer/finding mayDependOn, SDK zero-config-import rule)                                  |
| 6 | Version ranges (`VersionRange`, `ParseVersion`, `Version.Compare`)                                                       | No — art-dupl pins dependency versions in go.mod/flake.nix, not via policy objects                                                                              | `go-policy-dsl/policy.go` (`Version` type + `Compare`) | `art-dupl/go.mod`, `flake.nix` (rev-pinned gogenfilter, toolchain pin `go 1.27.1`)                                                                     |
| 7 | Values-not-config interchange (`PolicySpec` with no struct tags; consumer bridges at boundary)                           | Not applicable as a _consumer_; conceivable only if art-dupl ever _emits_ or _scans_ policies                                                                   | `go-policy-dsl/policy.go:168`                          | art-dupl's config merging is reflection-based on its own `Config` (`config/config_merge.go:6,17` — ADR-0003), a different interchange problem          |
| 8 | Deny-by-default `Mode` contract (zero value/garbage = ban-active)                                                        | No — but note the _pattern_ parity: art-dupl independently reached the same design in its accept-directive and baseline gates (strict refusal on unknown state) | `go-policy-dsl/README.md` (deny-by-default section)    | `cmd/baseline_identity.go:16-18,55-60` (`--baseline-version-must-match` refuses unknown analyzer identity rather than silently comparing)              |

N/A rows excluded: CVE tagging, replacement suggestions, and the fuzz/BDD test suite are consumer-facing conveniences downstream of rows 1-6 — with no policy author in art-dupl they have no surface to land on; the row-8 note documents pattern parity without implying adoption.

## Gap analysis: what art-dupl hand-rolls vs what the lib offers

The premise of a gap mostly does not hold: **art-dupl does not hand-roll a library-policy DSL, because it has no library-policy problem.** The task brief's candidate analogs decompose as follows:

**1. Baseline policy hashing ≠ library policy.** `cmd/baseline_identity.go:24-34` folds hash-affecting settings into an xxh3 fingerprint (`cache=%d;mode=%s;typeAware=%t;...`) that `check --baseline-version-must-match` compares against the recorded baseline:

```go
// skip-validate
// cmd/baseline_identity.go:26-33
return fmt.Sprintf(
    "cache=%d;mode=%s;typeAware=%t;suggestGenerics=%t;maxChildren=%d;methods=%s",
    cache.CacheVersion, cfg.DetectionMode.String(), cfg.TypeAware,
    cfg.SuggestGenerics, cfg.MaxChildrenSerial, detectionMethodsPolicyKey(...))
```

This is _self-identity_ hashing (what produced these clone hashes), not _governance_ (what libraries may be used). go-policy-dsl offers nothing here: no hashing, no compatibility gate, no cache-version concept. The semantic overlap is limited to the word "policy".

**2. Arch-lint exclusions ≠ the DSL.** `.go-arch-lint.yml:1-312` declares component dependency boundaries (`pkg/artdupl` must not import `config/`, etc.) enforced by go-arch-lint in CI (`AGENTS.md:152`). In DSL terms this is _architecture_ policy — and go-policy-dsl's `Category` enum even has `CategoryArchitecture` (`go-policy-dsl/policy.go`) — but the DSL has no component/import-graph model; adopting it would mean _re-implementing_ arch-lint's engine behind `ImportPatterns`, strictly worse than the existing dedicated tool.

**3. `//art-dupl:accept` directives ≠ ban policies.** `internal/accept/accept.go:1-450` implements in-source suppression of _clone findings_ (hash-precise accept directives with 5-line lookbehind, HTML-comment forms). This is suppression semantics for art-dupl's own detector output — the same consumer-side role go-policy-dsl explicitly refuses to play (it ships no matching engine, `go-policy-dsl/README.md` "When NOT to use"). No replacement, no simplification.

**4. The lint-config ban list is the closest real analog — and still doesn't fit.** art-dupl's strongest "policy" is the banned-linter invariant (`exhaustruct`/`tagliatelle` must never appear in `.golangci.yml`; enforced by `scripts/check-disabled-linters.sh:22` + the nix `disabled-linters` check at `flake.nix:254-256` + CI `.github/workflows/lint-config-guard.yml:31`). Could one declare `policydsl.Ban("tagliatelle").Because("mixed snake_case/camelCase JSON conventions per ADR-0016")...Spec()`? Cosmetically yes — but the DSL's patterns are opaque strings and the _entire detection engine_ (parse `.golangci.yml`, walk the enable list, fail the build, auto-remove) would still be art-dupl's to write, i.e. the shell script would survive intact next to a decorative policy value. A compile-time-checked policy object guarding a file that foreign daemons rewrite at runtime solves the wrong half of that problem anyway: art-dupl's failure mode is _another writer violating the declared policy_, not _the policy being mistyped_.

**Where the DSL genuinely shines — and why it still doesn't transfer:** its compile-time guarantee ("make impossible policy states unrepresentable") pays off when policies are _authored frequently and consumed by engines_ — exactly the `library-policy` use case, its first consumer at `~/projects/library-policy/go.mod:27` (`go-policy-dsl v0.4.0`). art-dupl would be at best a _scanned_ repo in such a fleet check, and scanned repos don't import the DSL — the scanner does.

## Verdict + derived score

**NOT-ADOPTED.** Inputs: 8 applicable capability rows, 0 adopted (Σ=0) → **0/100**, band "unrealized" (0-24). This is not a judgment of quality — the library is well-designed, dependency-free, and already adopted by its intended consumer — it is a judgment of audience: art-dupl contains no policy author, no policy engine, and no governance surface for it to model.

## Version currency

- Local describe: `v0.4.0-1-gee88a84` (1 commit past tag, docs-only); tag `v0.4.0` dated 2026-11-04 in `go-policy-dsl/CHANGELOG.md`.
- Tag cadence: v0.1.0 2026-07-26 → v0.2.0 2026-07-26 → v0.3.0 2026-08-08 → v0.4.0 2026-11-04. Roughly quarterly; the ~3-month gap to v0.4.0 plus a docs-only follow-up commit and a **stale README** (its status table still says "v0.3.0" while `v0.4.0` shipped `WhenAnyPresent`, `go-policy-dsl/README.md` Status section vs `go-policy-dsl/CHANGELOG.md` `[0.4.0]`) suggest a slowing, feature-complete-for-now project. `go-policy-dsl/TODO_LIST.md` reports **no open items** ("stable at v0.3.0" — itself stale by one minor) with the matcher subpackage / LSP server deferred to ROADMAP.
- `[Unreleased]` section: **empty** (`go-policy-dsl/CHANGELOG.md`) — no in-flight features. Next-tag ETA: unpredictable; nothing signals an imminent v0.5.0.

## Fleet argument

The fleet-centralization logic _does_ apply here, but with the consumer role inverted relative to linter-autoconfigure-sdk: go-policy-dsl centralizes **policy declaration** so that fleet-wide dependency governance ("gorm is banned", "this repo, when it has samber/do, must pair samber-do-auditlog") is declared once, in Go, and consumed by one scanner (`library-policy`, already on v0.4.0). If Lars wants that fleet gate, the correct integration for art-dupl is **to be scanned by library-policy, not to import the DSL** — art-dupl contributes nothing by adding the dependency, and gains coverage automatically once library-policy points its scanner at `~/projects`. The counterfactual cost is real, though small today: if every repo grew its own dep-policy script (the pattern art-dupl actually exhibits with `check-disabled-linters.sh`), bans would fork per-repo and a fleet-wide "ban X" would need N edits. The DSL + library-policy pair is the right centralization _for dependency policy_; it simply has nothing to say about art-dupl's actual pain class — **config invariant violations by BuildFlow-side rewriters** — which is a writer-collision problem, not a policy-declaration problem, and lives on the host side (see the sibling report: `/home/lars/projects/art-dupl/docs/review/libs/linter-autoconfigure-sdk.md`, "Which side the value lands on").

## Ranked opportunities (impact × ease)

1. **Wait for library-policy to scan art-dupl** (zero art-dupl work; potential future flip of this verdict's flip-trigger below). Impact: low-medium (art-dupl's deps are few and curated). Ease: none on art-dupl's side.
2. **Not recommended:** import the DSL to "declare" the banned-linter invariant. Impact: nil — detection engine remains hand-rolled, the DSL value would be decorative. Ease: easy, pointless.
3. **Speculative, host-side:** if a future toolsdk config-intent manifest (see sibling report, opportunity 2) wants a declaration vocabulary, go-policy-dsl's builder pattern is a plausible shape for it — a fleet-level synergy, but one decision upstream of art-dupl.

## RECONSIDER-WHEN

Flip to ADOPT-CANDIDATE when **any** of:

- **library-policy (or any fleet dependency gate) starts actively scanning art-dupl** and art-dupl needs to declare repo-local policies or exceptions — then policy values belong in the shared DSL rather than a private script.
- art-dupl grows a **dependency-audit feature of its own** (e.g. linting consumers' go.mods for banned clone-magnifying libs like `github.com/hectane/go-acl`-style choices, or banning `encoding/json/v2` imports fleet-style — note `internal/jsonv2gate/` already hand-rolls exactly such an import ban as a test; if that ban ever generalizes beyond one package into a policy surface, the DSL's declaration layer becomes relevant).
- The DSL ships the **matcher subpackage** (deferred to `go-policy-dsl/ROADMAP.md`) — a ready-made consumer engine would lower the integration cost that today makes adoption net-negative.

## How to verify (appendix — exact commands)

```bash
# versions / state
git -C /home/lars/projects/go-policy-dsl rev-parse HEAD                 # ee88a84a...
git -C /home/lars/projects/go-policy-dsl describe --tags                # v0.4.0-1-gee88a84
git -C /home/lars/projects/go-policy-dsl status --short                 # M .github/workflows/ci.yml (setup-go SHA bump)
git -C /home/lars/projects/go-policy-dsl diff .github/workflows/ci.yml  # actions/setup-go v7 -> v7.1.0
git -C /home/lars/projects/art-dupl rev-parse HEAD                      # cee67b4a... (pinned audit target)

# consumer evidence
grep -n "go-policy-dsl" /home/lars/projects/library-policy/go.mod       # :27  v0.4.0 — first consumer
cat /home/lars/projects/go-policy-dsl/go.mod                            # zero requires

# library capability evidence
grep -n "^func \|^type \|WhenAnyPresent\|RequireIfContains\|ErrInvertedVersionRange" \
  /home/lars/projects/go-policy-dsl/policy.go /home/lars/projects/go-policy-dsl/builder.go
cat /home/lars/projects/go-policy-dsl/CHANGELOG.md /home/lars/projects/go-policy-dsl/TODO_LIST.md /home/lars/projects/go-policy-dsl/README.md  # full read

# art-dupl hand-roll evidence
sed -n '24,34p;55,60p' /home/lars/projects/cmd/baseline_identity.go   # policy hashing
grep -n "func Validate\|func ValidateConfig" /home/lars/projects/config/config_validate.go
head -5 /home/lars/projects/.go-arch-lint.yml
grep -n "" /home/lars/projects/internal/accept/accept.go | sed -n '1,3p'
sed -n '22p' /home/lars/projects/scripts/check-disabled-linters.sh
sed -n '254,256p' /home/lars/projects/flake.nix
```
