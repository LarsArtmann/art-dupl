# Research Index

One line per audit/evaluation: file, date, question, verdict. Full re-reviews
get a NEW dated file; snapshots are extended with dated ledger sections, never
rewritten (policy: `AGENTS.md` → HTML research-report tracking policy).

| File                                                      | Date                        | Question                                        | Verdict                                                                                                                                                                                                                                                                                                                                    |
| --------------------------------------------------------- | --------------------------- | ----------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `2026-06-05_go-finding-integration-evaluation-summary.md` | 2026-06-05                  | Should art-dupl emit go-finding findings?       | Yes — as a consumer-side output layer, never the internal representation (later ADR-0025)                                                                                                                                                                                                                                                  |
| `algorithmic-alternatives-for-clone-detection.md`         | 2026-05                     | Which detection algorithms beyond suffix trees? | Hybrid suffix-tree + AST-hash retained; multi-method dispatch shipped                                                                                                                                                                                                                                                                      |
| `github-actions-distribution-options.md`                  | 2026-05                     | How to distribute the CLI via GitHub Actions?   | goreleaser + multi-OS assets (v0.5.0+)                                                                                                                                                                                                                                                                                                     |
| `2026-10-05_go-finding-deep-dive.html`                    | 2026-10-05 (+ ledger 10-06) | Is art-dupl using go-finding to the max?        | 60% pre-remediation → options channel, Builder, LSP, suppression, tags, columns shipped (§05 ledger); baseline-reinvention finding retracted; coverage channel upstream-gated (T07b). Shipped as **v0.9.0** (2026-10-06, proxy + pkg.go.dev live); consumer #1 live: BuildFlow `tool_options` → `toolsdk.WithOptions` (master `6b3ccc667`) |

Historical note: `docs/research/SPLIT-BRAIN.html` is cited by ADR-0005 and
AGENTS.md but no longer exists on disk (superseded by ADR-0005 itself and
`docs/status/archived/2026-06-22_19-44_split-brain-resolution-complete.md`;
the dangling citations are known and tolerated as historical references).
