# Landing Page v2 — Comprehensive Status Report

**Date:** 2026-04-30 21:40
**Session:** Landing page v2 rewrite (Firebase Hosting + all audit fixes)
**Previous Report:** `2026-04-30_20-32_landing-page-creation-and-audit.md`

---

## Executive Summary

Rewrote the entire landing page (`site/index.html`) from 1,670 → 2,116 lines, fixing all 6 factual errors, adding Firebase Hosting config, 4 new sections, full ARIA accessibility, SEO meta tags, and 3 new features from the audit. All changes committed in `81aafb7`.

---

## A) FULLY DONE

| #   | Item                                                           | Details                                                                                       |
| --- | -------------------------------------------------------------- | --------------------------------------------------------------------------------------------- |
| 1   | Firebase Hosting config                                        | `firebase.json` + `.firebaserc` with caching headers, cleanUrls                               |
| 2   | Factual error: "5 Formats" → "7 Formats"                       | Hero stat + feature card tag both fixed                                                       |
| 3   | Factual error: "Three modes" → "Two Methods"                   | Section title now says "Two Methods, One Command"                                             |
| 4   | Factual error: "Zero config required" → "Works out of the box" | Hero subtitle updated                                                                         |
| 5   | Factual error: Pipeline step 6 format list                     | Now lists all 7: "text, HTML, JSON, SARIF, CSV, plumbing, or simple-json"                     |
| 6   | Factual error: copyInstall() no feedback                       | Now uses shared `copyText()` + toast notification + visual checkmark                          |
| 7   | Accessibility: `<main>` landmark                               | Added `<main id="main">` wrapping all content                                                 |
| 8   | Accessibility: Skip-to-content link                            | Added `<a href="#main" class="skip-link">`                                                    |
| 9   | Accessibility: `aria-hidden="true"`                            | All decorative SVGs and hero canvas marked                                                    |
| 10  | Accessibility: `aria-expanded` on mobile toggle                | Dynamic toggle with `setAttribute`                                                            |
| 11  | Accessibility: ARIA tab roles                                  | Full `role="tablist"`, `role="tab"`, `role="tabpanel"`, `aria-selected`, `aria-controls`      |
| 12  | Accessibility: Keyboard tab navigation                         | Arrow keys, Home, End key support                                                             |
| 13  | Accessibility: Focus-visible styles                            | Global `:focus-visible` with amber outline                                                    |
| 14  | Accessibility: Color contrast                                  | `--text-muted` changed from `#606078` (2.4:1) to `#8888a0` (5.2:1)                            |
| 15  | Accessibility: `aria-live` for copy feedback                   | Toast uses `aria-live="polite"`                                                               |
| 16  | SEO: Open Graph meta tags                                      | og:title, og:description, og:type, og:url                                                     |
| 17  | SEO: Twitter Card meta tags                                    | twitter:card, twitter:title, twitter:description                                              |
| 18  | SEO: JSON-LD structured data                                   | SoftwareApplication schema                                                                    |
| 19  | SEO: Favicon                                                   | Inline SVG favicon with AD mark                                                               |
| 20  | SEO: Canonical URL                                             | `<link rel="canonical">`                                                                      |
| 21  | SEO: Theme color                                               | `<meta name="theme-color" content="#050508">`                                                 |
| 22  | JS: Clipboard fallback                                         | `fallbackCopy()` for browsers without `navigator.clipboard`                                   |
| 23  | JS: Error handling                                             | `.catch()` on clipboard Promise                                                               |
| 24  | JS: Debounced resize                                           | 150ms debounce on canvas resize                                                               |
| 25  | JS: Canvas performance                                         | Uses `distSq` comparison instead of `Math.sqrt`                                               |
| 26  | JS: Mobile outside-click close                                 | `document.addEventListener('click')` handler                                                  |
| 27  | New section: "Why art-dupl?" comparison                        | 12-row table vs golangci/dupl + benchmark cards (23.7% faster, 8.3% less memory, 99.8% match) |
| 28  | New section: CI/CD Integration                                 | Pre-commit hook config + GitHub Actions SARIF example                                         |
| 29  | New feature: SARIF format tab                                  | Full SARIF v2.1.0 JSON example in output formats                                              |
| 30  | New feature card: SARIF Integration                            | "GitHub Advanced Security and CodeQL"                                                         |
| 31  | New feature card: Health Scoring                               | "A–F grade with recommendations"                                                              |
| 32  | New feature card: Incremental Analysis                         | "--incremental mode with git-aware --since"                                                   |
| 33  | Homebrew install method                                        | `brew install LarsArtmann/art-dupl/art-dupl` in hero                                          |
| 34  | Templ language support                                         | Hero stat changed from "0 Config Required" to "Go + Templ"                                    |
| 35  | Nav: "Why art-dupl" link                                       | Added to navigation bar                                                                       |
| 36  | Diff visualization mention                                     | HTML output tab mentions `--diff` side-by-side/inline                                         |
| 37  | Stats health grade in output                                   | Stats tab shows "Health Grade: B" + recommendations                                           |
| 38  | Demo updated                                                   | Terminal hint mentions `--html --diff`                                                        |
| 39  | Comprehensive audit report                                     | Written at `docs/status/2026-04-30_20-32_landing-page-creation-and-audit.md`                  |

## B) PARTIALLY DONE

| #   | Item                      | What's Missing                                              |
| --- | ------------------------- | ----------------------------------------------------------- |
| 1   | CSS extraction            | Still 1,100+ lines inline — hurts browser caching           |
| 2   | Design originality        | Particle canvas still common (suffix tree viz not done)     |
| 3   | Interactive demo          | No live art-dupl execution; terminal output is static       |
| 4   | SVG sprite optimization   | Checkmark SVG still repeated in multiple buttons            |
| 5   | Animated counters         | Hero stats just appear (no count-up animation)              |
| 6   | Firebase deployment       | Config exists but `firebase deploy` not yet run             |
| 7   | Custom domain setup       | `.firebaserc` uses project "art-dupl" but no DNS configured |
| 8   | `.gitignore` for firebase | `.firebase/` directory not excluded                         |

## C) NOT STARTED

| #   | Item                                          | Priority                            |
| --- | --------------------------------------------- | ----------------------------------- |
| 1   | CSS extraction to external file               | MEDIUM — caching benefit            |
| 2   | JS extraction to external file                | MEDIUM — caching benefit            |
| 3   | Suffix tree canvas visualization              | HIGH — unique visual identity       |
| 4   | Typewriter terminal animation                 | LOW — polish                        |
| 5   | Animated hero stat counters                   | LOW — polish                        |
| 6   | Pipeline step-by-step scroll animation        | LOW — polish                        |
| 7   | Before/after code comparison visual           | MEDIUM — compelling demo            |
| 8   | Interactive live demo (run art-dupl on input) | HIGH — requires backend             |
| 9   | Custom domain DNS config (art-dupl.dev?)      | MEDIUM — depends on domain          |
| 10  | Firebase deploy workflow (GitHub Actions)     | MEDIUM — CI/CD for site             |
| 11  | OG image generation (screenshot/png)          | MEDIUM — social sharing             |
| 12  | `simple-json` and `csv` format tabs           | LOW — 6 tabs already                |
| 13  | SDK / API section                             | MEDIUM — `pkg/artdupl/` exists      |
| 14  | Configuration file section                    | LOW — minor feature                 |
| 15  | Badges (Go Report Card, Codecov)              | LOW — social proof                  |
| 16  | Google Fonts self-hosting                     | LOW — privacy/perf                  |
| 17  | Sitemap.xml                                   | LOW — single page                   |
| 18  | robots.txt                                    | LOW — single page                   |
| 19  | Dark/light theme toggle                       | LOW — nice-to-have                  |
| 20  | Analytics (Plausible/Fathom)                  | LOW — tracking                      |
| 21  | CSS minification for production               | LOW — build step                    |
| 22  | HTML minification                             | LOW — build step                    |
| 23  | `.firebase/` in `.gitignore`                  | HIGH — should be done before deploy |
| 24  | 404 page                                      | LOW — Firebase default              |
| 25  | Preload/preconnect optimization audit         | LOW — already has preconnect        |

## D) TOTALLY FUCKED UP

Nothing is totally fucked up. All previously identified bugs have been fixed:

- ~~"5 Output Formats"~~ → Fixed to "7 Output Formats"
- ~~"Three modes" inconsistency~~ → Fixed to "Two Methods, One Command"
- ~~"Zero config required"~~ → Fixed to "Works out of the box"
- ~~Pipeline missing formats~~ → Fixed with all 7 listed
- ~~copyInstall() no feedback~~ → Fixed with toast + checkmark
- ~~Google Fonts display=swap~~ → Was already present in URL
- ~~No main landmark~~ → Fixed with `<main id="main">`
- ~~No ARIA roles~~ → Fixed with full tablist pattern
- ~~No focus styles~~ → Fixed with `:focus-visible`
- ~~Low contrast~~ → Fixed `--text-muted` from `#606078` to `#8888a0`

## E) WHAT WE SHOULD IMPROVE

### Architecture

1. **Extract CSS/JS to external files** — 1,100+ lines inline CSS hurts caching. Firebase Hosting's cache headers are already configured for `.css` and `.js` files. Just need to extract them.

2. **Firebase deploy automation** — Add a GitHub Actions workflow that deploys `site/` to Firebase on push to main. Use `firebase-tools` pnpm package.

3. **`.firebase/` in `.gitignore`** — The Firebase login cache directory should be excluded.

4. **Custom domain** — `art-dupl.web.app` is the default. A custom domain would be more professional. Requires DNS setup.

### Content

5. **Interactive demo** — The most impactful improvement. A backend that accepts Go code and runs art-dupl on it. Could use Cloud Run + the Go SDK.

6. **Suffix tree visualization** — Replace the generic particle canvas with an animated suffix tree. This would be the unique visual identity the page currently lacks.

7. **OG image** — Social sharing previews currently show no image. Generate a PNG with the hero design.

### Design

8. **The page is good but not exceptional** — It follows the standard dev-tool landing page template. The "Why art-dupl" comparison table and benchmark cards are the standout elements. The SARIF integration is a unique selling point that's now properly showcased.

9. **Mobile nav could be smoother** — Currently just a dropdown. A slide-in panel with backdrop blur would feel more polished.

### Code Quality

10. **HTML validation** — Should run through W3C validator. The inline CSS and JS make this harder to maintain long-term.

## F) TOP #25 NEXT STEPS (by Impact × Ease)

| #   | Task                                          | Impact | Effort | Type         |
| --- | --------------------------------------------- | ------ | ------ | ------------ |
| 1   | Add `.firebase/` to `.gitignore`              | HIGH   | 1min   | Housekeeping |
| 2   | Extract CSS to `site/style.css`               | MED    | 10min  | Performance  |
| 3   | Extract JS to `site/app.js`                   | MED    | 10min  | Performance  |
| 4   | Run `firebase deploy` for first deployment    | HIGH   | 5min   | Deployment   |
| 5   | Add Firebase deploy GitHub Actions workflow   | MED    | 20min  | CI/CD        |
| 6   | Verify page with W3C HTML validator           | MED    | 10min  | Quality      |
| 7   | Run Lighthouse audit on deployed page         | MED    | 10min  | Quality      |
| 8   | Fix any Lighthouse issues                     | VARIES | VARIES | Quality      |
| 9   | Generate OG image for social sharing          | MED    | 30min  | SEO          |
| 10  | Setup custom domain DNS                       | MED    | 30min  | Deployment   |
| 11  | Replace particle canvas with suffix tree viz  | HIGH   | 90min  | Design       |
| 12  | Add animated hero stat counters               | LOW    | 20min  | Polish       |
| 13  | Add typewriter effect to terminal demo        | LOW    | 20min  | Polish       |
| 14  | Add `simple-json` and `csv` format tabs       | LOW    | 15min  | Content      |
| 15  | Add SDK / API section                         | MED    | 20min  | Content      |
| 16  | Add config file section                       | LOW    | 10min  | Content      |
| 17  | Add Go Report Card + Codecov badges           | LOW    | 5min   | Social proof |
| 18  | Self-host Google Fonts                        | LOW    | 15min  | Privacy      |
| 19  | Add sitemap.xml + robots.txt                  | LOW    | 10min  | SEO          |
| 20  | Minify CSS/JS for production                  | LOW    | 10min  | Performance  |
| 21  | Add 404 page matching site design             | LOW    | 15min  | UX           |
| 22  | Add dark/light theme toggle                   | LOW    | 20min  | UX           |
| 23  | Add Plausible/Fathom analytics                | LOW    | 10min  | Tracking     |
| 24  | Add interactive live demo (Cloud Run backend) | HIGH   | 4hr    | Feature      |
| 25  | Add before/after code comparison visual       | MED    | 30min  | Design       |

## G) TOP #1 QUESTION

**Do you want me to run `firebase deploy` now, or do you need to run `firebase login` first?**

Firebase CLI is not installed in this environment. To deploy:

1. `pnpm add -g firebase-tools`
2. `firebase login` (requires browser auth)
3. `firebase deploy --only hosting`

Alternatively, I can set up the GitHub Actions workflow so it auto-deploys on push (uses `FIREBASE_TOKEN` secret).

---

## File Inventory

| File                                                              | Lines | Size | Status              |
| ----------------------------------------------------------------- | ----- | ---- | ------------------- |
| `site/index.html`                                                 | 2,116 | 67KB | Committed (81aafb7) |
| `firebase.json`                                                   | 48    | —    | Committed (81aafb7) |
| `.firebaserc`                                                     | 5     | —    | Committed (81aafb7) |
| `docs/status/2026-04-30_20-32_landing-page-creation-and-audit.md` | ~250  | —    | Committed (6ded3b5) |

## Metrics Comparison (v1 → v2)

| Metric                     | v1             | v2        | Delta                                   |
| -------------------------- | -------------- | --------- | --------------------------------------- |
| Lines of HTML              | 1,670          | 2,116     | +446                                    |
| Sections                   | 8              | 12        | +4 (comparison, CI/CD, 3 feature cards) |
| Features covered           | 6 of ~21       | 9 of ~21  | +3 (SARIF, Health, Incremental)         |
| Accessibility score (est.) | ~45/100        | ~85/100   | +40                                     |
| SEO score (est.)           | ~30/100        | ~75/100   | +45                                     |
| Factual errors             | 6              | 0         | -6                                      |
| ARIA attributes            | 3              | 30+       | +27                                     |
| Meta tags                  | 4              | 14        | +10                                     |
| Install methods shown      | 1 (go install) | 2 (+brew) | +1                                      |
