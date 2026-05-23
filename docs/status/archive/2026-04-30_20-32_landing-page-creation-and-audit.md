# Landing Page — Comprehensive Status Report

**Date:** 2026-04-30 20:32
**Session:** Initial landing page creation + audit
**File:** `site/index.html`

---

## Executive Summary

Created a single-file landing page for art-dupl at `site/index.html`. The page has a working structure with 8 sections, responsive design, and interactive features. However, a deep audit revealed **significant gaps** in content accuracy, accessibility, SEO, and design originality. This report catalogues everything.

---

## A) FULLY DONE

| Item                     | Status | Details                                                            |
| ------------------------ | ------ | ------------------------------------------------------------------ |
| Basic page structure     | DONE   | Hero, Pipeline, Features, Methods, Demo, Output, CTA, Footer       |
| Responsive CSS           | DONE   | Desktop, tablet, mobile breakpoints at 1024/768/480px              |
| Navigation               | DONE   | Sticky glassmorphism nav, mobile hamburger menu                    |
| Copy-to-clipboard        | DONE   | Hero install command + CTA button (partially — CTA lacks feedback) |
| Format tab switching     | DONE   | 5 tabs with JS switching                                           |
| Scroll reveal animations | DONE   | IntersectionObserver-based with prefers-reduced-motion             |
| Hero canvas animation    | DONE   | Node graph network (amber particles + connections)                 |
| Color system             | DONE   | CSS custom properties with amber/gold accent on dark slate         |
| Typography               | DONE   | Syne (display), DM Sans (body), IBM Plex Mono (code)               |
| Terminal demo section    | DONE   | Styled terminal with realistic art-dupl output                     |
| GitHub link              | DONE   | In nav and CTA                                                     |
| Basic meta tags          | DONE   | title, description, viewport, charset                              |

## B) PARTIALLY DONE

| Item                   | Status | What's Missing                                                               |
| ---------------------- | ------ | ---------------------------------------------------------------------------- |
| Feature coverage       | ~40%   | 6 features shown, ~15+ features from FEATURES.md not mentioned               |
| Output format showcase | ~70%   | Shows 5 formats; missing SARIF, CSV, simple-json                             |
| Accessibility          | ~30%   | Missing main landmark, skip link, ARIA roles, focus styles, tab keyboard nav |
| SEO                    | ~20%   | Missing OG tags, Twitter cards, canonical URL, JSON-LD, favicon              |
| Design originality     | ~50%   | Particle canvas is common; no unique visual hook                             |
| CTA copy feedback      | ~50%   | Hero copy works; CTA copyInstall() has no visual feedback                    |
| Error handling in JS   | ~30%   | No .catch() on clipboard promises; no fallback for older browsers            |
| Mobile UX              | ~60%   | Menu doesn't close on outside click; no swipe gestures                       |

## C) NOT STARTED

| Item                                     | Priority | Notes                                                 |
| ---------------------------------------- | -------- | ----------------------------------------------------- |
| Open Graph / Twitter Card meta tags      | HIGH     | Critical for link sharing previews                    |
| `<main>` landmark + skip-to-content link | HIGH     | WCAG 2.1 Level A requirement                          |
| ARIA roles on format tabs                | HIGH     | role="tablist/tab/tabpanel", aria-selected            |
| Focus-visible styles                     | HIGH     | Tab navigation currently invisible on dark bg         |
| Color contrast fix (--text-muted)        | HIGH     | #606078 on #050508 = 2.4:1 (needs 4.5:1)              |
| Favicon                                  | MEDIUM   | No favicon defined                                    |
| JSON-LD structured data                  | MEDIUM   | SoftwareApplication schema for SEO                    |
| Canonical URL                            | MEDIUM   | Prevents duplicate content                            |
| CSS extraction to external file          | MEDIUM   | 1049 lines inline — hurts caching                     |
| SARIF format showcase                    | MEDIUM   | Enterprise differentiator                             |
| CI/CD integration section                | HIGH     | No examples of GitHub Actions, GitLab CI              |
| Pre-commit hook section                  | MEDIUM   | examples/pre-commit/ exists                           |
| Homebrew install method                  | MEDIUM   | brew install LarsArtmann/art-dupl/art-dupl            |
| Diff visualization feature               | MEDIUM   | --diff side-by-side/inline not mentioned              |
| Incremental analysis feature             | MEDIUM   | --incremental flag not mentioned                      |
| Stats health score (A-F grading)         | MEDIUM   | Unique differentiator from stats_health.go            |
| Stats recommendations                    | MEDIUM   | Smart actionable advice from stats_recommendations.go |
| SDK / API section                        | MEDIUM   | pkg/artdupl/ is a complete Go SDK                     |
| Configuration file section               | MEDIUM   | -config dupl.json for team consistency                |
| Comparison vs original dupl              | HIGH     | Must justify the fork                                 |
| Benchmarks section                       | MEDIUM   | 23.7% faster than golangci/dupl on Prometheus         |
| Typewriter terminal animation            | LOW      | Demo section is static text                           |
| Animated hero stats counter              | LOW      | Numbers just appear, no count-up                      |
| Pipeline step-by-step scroll animation   | LOW      | Pipeline is currently static                          |
| Before/after code comparison visual      | LOW      | No visual demonstration                               |
| Social proof (badges, stars)             | LOW      | Go Report Card, Codecov badges exist                  |
| Templ support callout                    | MEDIUM   | Full .templ support is unique                         |
| `aria-expanded` on mobile toggle         | HIGH     | WCAG requirement                                      |
| Canvas aria-hidden                       | MEDIUM   | Decorative canvas needs aria-hidden="true"            |
| Debounced resize handler                 | LOW      | Canvas recreates nodes on every resize                |
| SVG sprite for repeated icons            | LOW      | Checkmark SVG repeated 9 times                        |

## D) TOTALLY FUCKED UP

| Item                                   | Why                                                             |
| -------------------------------------- | --------------------------------------------------------------- |
| "5 Output Formats" hero stat           | Actually 7: text, html, json, csv, plumbing, simple-json, sarif |
| "Three modes" vs "2 Detection Methods" | Inconsistent — methods section says 3, hero says 2              |
| "Zero config required"                 | Misleading — config files are a FEATURE, not a deficit          |
| Pipeline step 6 output list            | Says "text, HTML, JSON, or plumbing" — missing 3 formats        |
| copyInstall() CTA feedback             | Clicking shows nothing — no checkmark, no toast                 |
| Google Fonts URL                       | Missing `&display=swap` — render-blocking                       |

## E) WHAT WE SHOULD IMPROVE

### Architecture & Type Models

1. **Extract site to a build system**: The 1049-line inline CSS + JS is unmaintainable. Consider:
   - **Vite** or **Astro** for a proper build pipeline
   - Or at minimum: external `style.css` + `app.js` files
   - The `flake.nix` already supports builds — add a site build target

2. **Use existing art-dupl output for demo content**: Instead of fake demo output, run `art-dupl --json .` on the project itself and use real data. The JSON output format is already structured for this.

3. **Type-safe content**: Consider generating some page content from `config/output_format.go`, `FEATURES.md`, etc. to keep docs in sync with code.

4. **Leverage well-established libs for future iterations**:
   - **Astro** — Static site generator perfect for landing pages, zero JS by default
   - **Tailwind CSS** — If we want rapid iteration with utility classes
   - **Shiki** — For actual syntax highlighting in the HTML output showcase (from the `printer/html.go` output)
   - **Chart.js / Chart.scss** — For the stats health score visualization
   - **Inter** — Actually good for body text, despite the skill guidance (it's for landing pages, not app UIs)

5. **Existing code that could be reused**:
   - `printer/html.go` already generates HTML reports — could be adapted for an interactive demo
   - `printer/stats_visualization.go` has ASCII bar charts — could be rendered as actual charts
   - `printer/stats_health.go` has the A-F grading logic — could power an interactive demo
   - `BENCHMARK_COMPARISON.md` has real benchmark data — use it directly

### Design & Content

6. **Unique visual identity**: The particle canvas is the #1 most common dev landing page background. Alternatives:
   - Animated suffix tree visualization (actually draw a suffix tree growing)
   - Code diff animation showing before/after
   - Geometric pattern based on actual Go AST node types
   - ASCII art aesthetic (terminal-forward, matching the tool's identity)

7. **"Why art-dupl?" section**: Critical for a fork. Must explain:
   - 23.7% faster than golangci/dupl
   - Semantic awareness (identifier-based matching)
   - 7 output formats vs dupl's 1
   - Professional CLI with completions, man pages
   - Stats with health scoring
   - SARIF for GitHub Advanced Security
   - Templ support
   - Pre-commit hooks
   - Config files
   - SDK for programmatic use

8. **CI/CD section**: Show GitHub Actions workflow, pre-commit hook, JSON + jq pipeline

## F) TOP #25 THINGS TO DO NEXT

Sorted by **Impact × Ease** (highest first):

| #   | Task                                                          | Impact | Effort | Type     |
| --- | ------------------------------------------------------------- | ------ | ------ | -------- |
| 1   | Fix "5 Output Formats" → "7 Output Formats"                   | HIGH   | 5min   | Bug fix  |
| 2   | Fix "Three modes" vs "2 Detection Methods" inconsistency      | HIGH   | 5min   | Bug fix  |
| 3   | Fix "Zero config required" → "Works out of the box"           | MED    | 2min   | Copy fix |
| 4   | Fix pipeline step 6 output list (add SARIF, CSV, simple-json) | MED    | 5min   | Bug fix  |
| 5   | Add `<main>` landmark + skip-to-content link                  | HIGH   | 10min  | A11y     |
| 6   | Add focus-visible styles                                      | HIGH   | 10min  | A11y     |
| 7   | Fix --text-muted color contrast (2.4:1 → 4.5:1+)              | HIGH   | 10min  | A11y     |
| 8   | Add aria-expanded on mobile toggle + close on outside click   | MED    | 10min  | A11y     |
| 9   | Add aria-hidden="true" on hero canvas                         | MED    | 2min   | A11y     |
| 10  | Add Open Graph + Twitter Card meta tags                       | HIGH   | 15min  | SEO      |
| 11  | Add favicon (generate from "AD" mark)                         | MED    | 10min  | SEO      |
| 12  | Add canonical URL + theme-color meta                          | MED    | 5min   | SEO      |
| 13  | Fix Google Fonts `&display=swap`                              | MED    | 2min   | Perf     |
| 14  | Fix copyInstall() — add visual feedback + error handling      | MED    | 10min  | UX       |
| 15  | Add ARIA roles on format tabs (tablist/tab/tabpanel)          | HIGH   | 15min  | A11y     |
| 16  | Add "Why art-dupl?" comparison section vs original dupl       | HIGH   | 30min  | Content  |
| 17  | Add CI/CD integration section (GitHub Actions, pre-commit)    | HIGH   | 20min  | Content  |
| 18  | Add SARIF format to output tabs                               | MED    | 15min  | Content  |
| 19  | Add Homebrew install method alongside go install              | MED    | 5min   | Content  |
| 20  | Add benchmark section (23.7% faster on Prometheus)            | MED    | 20min  | Content  |
| 21  | Add stats health score showcase (A-F grading)                 | MED    | 15min  | Content  |
| 22  | Add diff visualization feature mention                        | MED    | 10min  | Content  |
| 23  | Extract CSS to external file for caching                      | MED    | 10min  | Perf     |
| 24  | Debounce canvas resize handler                                | LOW    | 5min   | Perf     |
| 25  | Replace particle canvas with suffix tree visualization        | HIGH   | 60min  | Design   |

## G) TOP #1 QUESTION

**What is the deployment target for this landing page?**

This fundamentally affects all decisions:

- **GitHub Pages** (most likely for an open-source Go tool) → means we need a static build, CNAME config, and the `site/` directory works as-is
- **Custom domain** (art-dupl.dev or artdupl.io) → needs DNS config, SSL
- **Embedded in Go binary** → serve from `//go:embed` for a self-hosted demo
- **Part of the GitHub README** → means we need a much simpler approach

The answer determines whether we invest in build tooling (Vite/Astro) or keep the single-file approach, and whether we need a GitHub Actions workflow for deployment.

---

## Metrics Summary

| Metric                          | Value                                                             |
| ------------------------------- | ----------------------------------------------------------------- |
| Lines of HTML/CSS/JS            | ~1,670                                                            |
| Inline CSS lines                | ~1,049                                                            |
| JS lines                        | ~130                                                              |
| Sections                        | 8 (hero, pipeline, features, methods, demo, formats, cta, footer) |
| Features covered                | 6 of ~21                                                          |
| Accessibility score (estimated) | ~45/100                                                           |
| SEO score (estimated)           | ~30/100                                                           |
| Mobile responsiveness           | ~75/100                                                           |
| Design originality              | ~50/100                                                           |
| Content accuracy                | ~65/100                                                           |
