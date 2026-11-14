# Landing Page v2 — Comprehensive Status Report

**Date:** 2026-04-30 21:40
**Session:** Landing page v2 rewrite (Firebase Hosting + all audit fixes)
**Previous Report:** `2026-04-30_20-32_landing-page-creation-and-audit.md`

---

## Executive Summary

Rewrote the entire landing page (`site/index.html`) from 1,670 → 2,116 lines, fixing all 6 factual errors, adding Firebase Hosting config, 4 new sections, full ARIA accessibility, SEO meta tags, and 3 new features from the audit. All changes committed in `81aafb7`.

---

## A) FULLY DONE

| #  | Item                                                           | Details                                                                                       |
| -- | -------------------------------------------------------------- | --------------------------------------------------------------------------------------------- |
| ~~1~~ done at `81aafb7` | ~~Firebase Hosting config~~ | ~~`firebase.json` + `.firebaserc` with caching headers, cleanUrls~~ |
| ~~2~~ done at `81aafb7` | ~~Factual error: "5 Formats" → "7 Formats"~~ | ~~Hero stat + feature card tag both fixed~~ |
| ~~3~~ done at `81aafb7` | ~~Factual error: "Three modes" → "Two Methods"~~ | ~~Section title now says "Two Methods, One Command"~~ |
| ~~4~~ done at `81aafb7` | ~~Factual error: "Zero config required" → "Works out of the box"~~ | ~~Hero subtitle updated~~ |
| ~~5~~ done at `81aafb7` | ~~Factual error: Pipeline step 6 format list~~ | ~~Now lists all 7: "text, HTML, JSON, SARIF, CSV, plumbing, or simple-json"~~ |
| ~~6~~ done at `81aafb7` | ~~Factual error: copyInstall() no feedback~~ | ~~Now uses shared `copyText()` + toast notification + visual checkmark~~ |
| ~~7~~ done at `81aafb7` | ~~Accessibility: `<main>` landmark~~ | ~~Added `<main id="main">` wrapping all content~~ |
| ~~8~~ done at `81aafb7` | ~~Accessibility: Skip-to-content link~~ | ~~Added `<a href="#main" class="skip-link">`~~ |
| ~~9~~ done at `81aafb7` | ~~Accessibility: `aria-hidden="true"`~~ | ~~All decorative SVGs and hero canvas marked~~ |
| ~~10~~ done at `81aafb7` | ~~Accessibility: `aria-expanded` on mobile toggle~~ | ~~Dynamic toggle with `setAttribute`~~ |
| ~~11~~ done at `81aafb7` | ~~Accessibility: ARIA tab roles~~ | ~~Full `role="tablist"`, `role="tab"`, `role="tabpanel"`, `aria-selected`, `aria-controls`~~ |
| ~~12~~ done at `81aafb7` | ~~Accessibility: Keyboard tab navigation~~ | ~~Arrow keys, Home, End key support~~ |
| ~~13~~ done at `81aafb7` | ~~Accessibility: Focus-visible styles~~ | ~~Global `:focus-visible` with amber outline~~ |
| ~~14~~ done at `81aafb7` | ~~Accessibility: Color contrast~~ | ~~`--text-muted` changed from `#606078` (2.4:1) to `#8888a0` (5.2:1)~~ |
| ~~15~~ done at `81aafb7` | ~~Accessibility: `aria-live` for copy feedback~~ | ~~Toast uses `aria-live="polite"`~~ |
| ~~16~~ done at `81aafb7` | ~~SEO: Open Graph meta tags~~ | ~~og:title, og:description, og:type, og:url~~ |
| ~~17~~ done at `81aafb7` | ~~SEO: Twitter Card meta tags~~ | ~~twitter:card, twitter:title, twitter:description~~ |
| ~~18~~ done at `81aafb7` | ~~SEO: JSON-LD structured data~~ | ~~SoftwareApplication schema~~ |
| ~~19~~ done at `81aafb7` | ~~SEO: Favicon~~ | ~~Inline SVG favicon with AD mark~~ |
| ~~20~~ done at `81aafb7` | ~~SEO: Canonical URL~~ | ~~`<link rel="canonical">`~~ |
| ~~21~~ done at `81aafb7` | ~~SEO: Theme color~~ | ~~`<meta name="theme-color" content="#050508">`~~ |
| ~~22~~ done at `81aafb7` | ~~JS: Clipboard fallback~~ | ~~`fallbackCopy()` for browsers without `navigator.clipboard`~~ |
| ~~23~~ done at `81aafb7` | ~~JS: Error handling~~ | ~~`.catch()` on clipboard Promise~~ |
| ~~24~~ done at `81aafb7` | ~~JS: Debounced resize~~ | ~~150ms debounce on canvas resize~~ |
| ~~25~~ done at `81aafb7` | ~~JS: Canvas performance~~ | ~~Uses `distSq` comparison instead of `Math.sqrt`~~ |
| ~~26~~ done at `81aafb7` | ~~JS: Mobile outside-click close~~ | ~~`document.addEventListener('click')` handler~~ |
| ~~27~~ done at `81aafb7` | ~~New section: "Why art-dupl?" comparison~~ | ~~12-row table vs golangci/dupl + benchmark cards (23.7% faster, 8.3% less memory, 99.8% match)~~ |
| ~~28~~ done at `81aafb7` | ~~New section: CI/CD Integration~~ | ~~Pre-commit hook config + GitHub Actions SARIF example~~ |
| ~~29~~ done at `81aafb7` | ~~New feature: SARIF format tab~~ | ~~Full SARIF v2.1.0 JSON example in output formats~~ |
| ~~30~~ done at `81aafb7` | ~~New feature card: SARIF Integration~~ | ~~"GitHub Advanced Security and CodeQL"~~ |
| ~~31~~ done at `81aafb7` | ~~New feature card: Health Scoring~~ | ~~"A–F grade with recommendations"~~ |
| ~~32~~ done at `81aafb7` | ~~New feature card: Incremental Analysis~~ | ~~"--incremental mode with git-aware --since"~~ |
| ~~33~~ done at `81aafb7` | ~~Homebrew install method~~ | ~~`brew install LarsArtmann/art-dupl/art-dupl` in hero~~ |
| ~~34~~ done at `81aafb7` | ~~Templ language support~~ | ~~Hero stat changed from "0 Config Required" to "Go + Templ"~~ |
| ~~35~~ done at `81aafb7` | ~~Nav: "Why art-dupl" link~~ | ~~Added to navigation bar~~ |
| ~~36~~ done at `81aafb7` | ~~Diff visualization mention~~ | ~~HTML output tab mentions `--diff` side-by-side/inline~~ |
| ~~37~~ done at `81aafb7` | ~~Stats health grade in output~~ | ~~Stats tab shows "Health Grade: B" + recommendations~~ |
| ~~38~~ done at `81aafb7` | ~~Demo updated~~ | ~~Terminal hint mentions `--html --diff`~~ |
| ~~39~~ done at `81aafb7` | ~~Comprehensive audit report~~ | ~~Written at `docs/status/2026-04-30_20-32_landing-page-creation-and-audit.md`~~ |

## B) PARTIALLY DONE

| # | Item                      | What's Missing                                              |
| - | ------------------------- | ----------------------------------------------------------- |
| ~~1~~ **Superseded — superseded — Astro site ships extracted styles (website/src/styles/global.css).** | ~~CSS extraction~~ | ~~Still 1,100+ lines inline — hurts browser caching~~ |
| ~~2~~ **Won't implement — dropped — particle canvas gone with the Astro rebuild.** | ~~Design originality~~ | ~~Particle canvas still common (suffix tree viz not done)~~ |
| ~~3~~ **Won't implement — no live backend; dropped and not harvested into ROADMAP.** | ~~Interactive demo~~ | ~~No live art-dupl execution; terminal output is static~~ |
| ~~4~~ **Won't implement — static page replaced; moot after the Astro rebuild.** | ~~SVG sprite optimization~~ | ~~Checkmark SVG still repeated in multiple buttons~~ |
| ~~5~~ **Won't implement — not carried into the Astro site.** | ~~Animated counters~~ | ~~Hero stats just appear (no count-up animation)~~ |
| ~~6~~ done — .github/workflows/deploy-site.yml; live at art-dupl.lars.software | ~~Firebase deployment~~ | ~~Config exists but `firebase deploy` not yet run~~ |
| ~~7~~ done — art-dupl.lars.software (website/astro.config.mjs) | ~~Custom domain setup~~ | ~~`.firebaserc` uses project "art-dupl" but no DNS configured~~ |
| ~~8~~ done — .gitignore excludes .firebase/ | ~~`.gitignore` for firebase~~ | ~~`.firebase/` directory not excluded~~ |

## C) NOT STARTED

| #  | Item                                          | Priority                            |
| -- | --------------------------------------------- | ----------------------------------- |
| ~~1~~ **Superseded — superseded — Astro site ships extracted styles (website/src/styles/global.css).** | ~~CSS extraction to external file~~ | ~~MEDIUM — caching benefit~~ |
| ~~2~~ **Superseded — superseded — Astro components/scripts (website/src).** | ~~JS extraction to external file~~ | ~~MEDIUM — caching benefit~~ |
| ~~3~~ **Won't implement — dropped — not carried into the Astro site.** | ~~Suffix tree canvas visualization~~ | ~~HIGH — unique visual identity~~ |
| ~~4~~ **Won't implement — dropped — not harvested into ROADMAP.** | ~~Typewriter terminal animation~~ | ~~LOW — polish~~ |
| ~~5~~ **Won't implement — dropped — not harvested into ROADMAP.** | ~~Animated hero stat counters~~ | ~~LOW — polish~~ |
| ~~6~~ **Won't implement — dropped — not harvested into ROADMAP.** | ~~Pipeline step-by-step scroll animation~~ | ~~LOW — polish~~ |
| ~~7~~ **Won't implement — dropped — not harvested into ROADMAP.** | ~~Before/after code comparison visual~~ | ~~MEDIUM — compelling demo~~ |
| ~~8~~ **Won't implement — no live backend; dropped and not harvested into ROADMAP.** | ~~Interactive live demo (run art-dupl on input)~~ | ~~HIGH — requires backend~~ |
| ~~9~~ done — art-dupl.lars.software | ~~Custom domain DNS config (art-dupl.dev?)~~ | ~~MEDIUM — depends on domain~~ |
| ~~10~~ done — .github/workflows/deploy-site.yml | ~~Firebase deploy workflow (GitHub Actions)~~ | ~~MEDIUM — CI/CD for site~~ |
| ~~11~~ done — website/public/og/home.png + og:image meta | ~~OG image generation (screenshot/png)~~ | ~~MEDIUM — social sharing~~ |
| ~~12~~ done — website/src/content/docs/guides/output-formats.mdx covers csv/simple-json | ~~`simple-json` and `csv` format tabs~~ | ~~LOW — 6 tabs already~~ |
| ~~13~~ done — website/src/content/docs/guides/sdk.mdx | ~~SDK / API section~~ | ~~MEDIUM — `pkg/artdupl/` exists~~ |
| ~~14~~ done — website/src/content/docs/configuration.mdx | ~~Configuration file section~~ | ~~LOW — minor feature~~ |
| ~~15~~ **Won't implement — dropped — README carries badges; the site does not.** | ~~Badges (Go Report Card, Codecov)~~ | ~~LOW — social proof~~ |
| ~~16~~ **Won't implement — dropped — no Google Fonts dependency in the Astro site.** | ~~Google Fonts self-hosting~~ | ~~LOW — privacy/perf~~ |
| ~~17~~ done — @astrojs/sitemap (website/astro.config.mjs) | ~~Sitemap.xml~~ | ~~LOW — single page~~ |
| ~~18~~ done — website/public/robots.txt | ~~robots.txt~~ | ~~LOW — single page~~ |
| ~~19~~ done — theme toggle in website/src/components/Header.astro | ~~Dark/light theme toggle~~ | ~~LOW — nice-to-have~~ |
| ~~20~~ **Won't implement — dropped — not harvested into ROADMAP.** | ~~Analytics (Plausible/Fathom)~~ | ~~LOW — tracking~~ |
| ~~21~~ done — Astro compressHTML build (website/astro.config.mjs) | ~~CSS minification for production~~ | ~~LOW — build step~~ |
| ~~22~~ done — Astro compressHTML build (website/astro.config.mjs) | ~~HTML minification~~ | ~~LOW — build step~~ |
| ~~23~~ done — .gitignore excludes .firebase/ | ~~`.firebase/` in `.gitignore`~~ | ~~HIGH — should be done before deploy~~ |
| ~~24~~ **Won't implement — dropped — not harvested into ROADMAP.** | ~~404 page~~ | ~~LOW — Firebase default~~ |
| ~~25~~ **Won't implement — moot — the audited static page no longer exists.** | ~~Preload/preconnect optimization audit~~ | ~~LOW — already has preconnect~~ |

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

~~1. **Extract CSS/JS to external files** — 1,100+ lines inline CSS hurts caching. Firebase Hosting's cache headers are already configured for `.css` and `.js` files. Just need to extract them.~~ **Superseded — superseded — Astro site ships extracted styles (website/src/styles/global.css).**

2. ~~**Firebase deploy automation** — Add a GitHub Actions workflow that deploys `site/` to Firebase on push to main. Use `firebase-tools` pnpm package.~~ done — .github/workflows/deploy-site.yml

3. ~~**`.firebase/` in `.gitignore`** — The Firebase login cache directory should be excluded.~~ done — .gitignore excludes .firebase/

4. ~~**Custom domain** — `art-dupl.web.app` is the default. A custom domain would be more professional. Requires DNS setup.~~ done — art-dupl.lars.software

### Content

~~5. **Interactive demo** — The most impactful improvement. A backend that accepts Go code and runs art-dupl on it. Could use Cloud Run + the Go SDK.~~ **Won't implement — no live backend; dropped and not harvested into ROADMAP.**

~~6. **Suffix tree visualization** — Replace the generic particle canvas with an animated suffix tree. This would be the unique visual identity the page currently lacks.~~ **Won't implement — dropped — not carried into the Astro site.**

7. ~~**OG image** — Social sharing previews currently show no image. Generate a PNG with the hero design.~~ done — website/public/og/home.png + og:image meta

### Design

~~8. **The page is good but not exceptional** — It follows the standard dev-tool landing page template. The "Why art-dupl" comparison table and benchmark cards are the standout elements. The SARIF integration is a unique selling point that's now properly showcased.~~ **Superseded — superseded — the static page was replaced by the Astro site (commit 2d3bc351).**

~~9. **Mobile nav could be smoother** — Currently just a dropdown. A slide-in panel with backdrop blur would feel more polished.~~ **Won't implement — not carried into the Astro site.**

### Code Quality

10. ~~**HTML validation** — Should run through W3C validator. The inline CSS and JS make this harder to maintain long-term.~~ done — website/.htmlvalidate.json wired

## F) TOP #25 NEXT STEPS (by Impact × Ease)

| #  | Task                                          | Impact | Effort | Type         |
| -- | --------------------------------------------- | ------ | ------ | ------------ |
| ~~1~~ done — .gitignore excludes .firebase/ | ~~Add `.firebase/` to `.gitignore`~~ | ~~HIGH~~ | ~~1min~~ | ~~Housekeeping~~ |
| ~~2~~ **Superseded — superseded — Astro site ships extracted styles (website/src/styles/global.css).** | ~~Extract CSS to `site/style.css`~~ | ~~MED~~ | ~~10min~~ | ~~Performance~~ |
| ~~3~~ **Superseded — superseded — Astro components/scripts (website/src).** | ~~Extract JS to `site/app.js`~~ | ~~MED~~ | ~~10min~~ | ~~Performance~~ |
| ~~4~~ done — deployed via .github/workflows/deploy-site.yml; live at art-dupl.lars.software | ~~Run `firebase deploy` for first deployment~~ | ~~HIGH~~ | ~~5min~~ | ~~Deployment~~ |
| ~~5~~ done — .github/workflows/deploy-site.yml | ~~Add Firebase deploy GitHub Actions workflow~~ | ~~MED~~ | ~~20min~~ | ~~CI/CD~~ |
| ~~6~~ done — website/.htmlvalidate.json wired | ~~Verify page with W3C HTML validator~~ | ~~MED~~ | ~~10min~~ | ~~Quality~~ |
| ~~7~~ **Superseded — superseded — Lighthouse CI still open in ROADMAP.md (Website and Presence).** | ~~Run Lighthouse audit on deployed page~~ | ~~MED~~ | ~~10min~~ | ~~Quality~~ |
| ~~8~~ **Superseded — superseded — Lighthouse CI still open in ROADMAP.md (Website and Presence).** | ~~Fix any Lighthouse issues~~ | ~~VARIES~~ | ~~VARIES~~ | ~~Quality~~ |
| ~~9~~ done — website/public/og/home.png + og:image meta | ~~Generate OG image for social sharing~~ | ~~MED~~ | ~~30min~~ | ~~SEO~~ |
| ~~10~~ done — art-dupl.lars.software | ~~Setup custom domain DNS~~ | ~~MED~~ | ~~30min~~ | ~~Deployment~~ |
| ~~11~~ **Won't implement — dropped — not carried into the Astro site.** | ~~Replace particle canvas with suffix tree viz~~ | ~~HIGH~~ | ~~90min~~ | ~~Design~~ |
| ~~12~~ **Won't implement — dropped — not harvested into ROADMAP.** | ~~Add animated hero stat counters~~ | ~~LOW~~ | ~~20min~~ | ~~Polish~~ |
| ~~13~~ **Won't implement — dropped — not harvested into ROADMAP.** | ~~Add typewriter effect to terminal demo~~ | ~~LOW~~ | ~~20min~~ | ~~Polish~~ |
| ~~14~~ done — website/src/content/docs/guides/output-formats.mdx covers csv/simple-json | ~~Add `simple-json` and `csv` format tabs~~ | ~~LOW~~ | ~~15min~~ | ~~Content~~ |
| ~~15~~ done — website/src/content/docs/guides/sdk.mdx | ~~Add SDK / API section~~ | ~~MED~~ | ~~20min~~ | ~~Content~~ |
| ~~16~~ done — website/src/content/docs/configuration.mdx | ~~Add config file section~~ | ~~LOW~~ | ~~10min~~ | ~~Content~~ |
| ~~17~~ **Won't implement — dropped — README carries badges; the site does not.** | ~~Add Go Report Card + Codecov badges~~ | ~~LOW~~ | ~~5min~~ | ~~Social proof~~ |
| ~~18~~ **Won't implement — dropped — no Google Fonts dependency in the Astro site.** | ~~Self-host Google Fonts~~ | ~~LOW~~ | ~~15min~~ | ~~Privacy~~ |
| ~~19~~ done — @astrojs/sitemap + website/public/robots.txt | ~~Add sitemap.xml + robots.txt~~ | ~~LOW~~ | ~~10min~~ | ~~SEO~~ |
| ~~20~~ done — Astro compressHTML build (website/astro.config.mjs) | ~~Minify CSS/JS for production~~ | ~~LOW~~ | ~~10min~~ | ~~Performance~~ |
| ~~21~~ **Won't implement — dropped — not harvested into ROADMAP.** | ~~Add 404 page matching site design~~ | ~~LOW~~ | ~~15min~~ | ~~UX~~ |
| ~~22~~ done — theme toggle in website/src/components/Header.astro | ~~Add dark/light theme toggle~~ | ~~LOW~~ | ~~20min~~ | ~~UX~~ |
| ~~23~~ **Won't implement — dropped — not harvested into ROADMAP.** | ~~Add Plausible/Fathom analytics~~ | ~~LOW~~ | ~~10min~~ | ~~Tracking~~ |
| ~~24~~ **Won't implement — no live backend; dropped and not harvested into ROADMAP.** | ~~Add interactive live demo (Cloud Run backend)~~ | ~~HIGH~~ | ~~4hr~~ | ~~Feature~~ |
| ~~25~~ **Won't implement — dropped — not harvested into ROADMAP.** | ~~Add before/after code comparison visual~~ | ~~MED~~ | ~~30min~~ | ~~Design~~ |

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
| ~~`site/index.html`~~ done at `81aafb7` | ~~2,116~~ | ~~67KB~~ | ~~Committed (81aafb7)~~ |
| ~~`firebase.json`~~ done at `81aafb7` | ~~48~~ | ~~—~~ | ~~Committed (81aafb7)~~ |
| ~~`.firebaserc`~~ done at `81aafb7` | ~~5~~ | ~~—~~ | ~~Committed (81aafb7)~~ |
| ~~`docs/status/2026-04-30_20-32_landing-page-creation-and-audit.md`~~ done at `6ded3b5` | ~~~250~~ | ~~—~~ | ~~Committed (6ded3b5)~~ |

## Metrics Comparison (v1 → v2)

| Metric                     | v1             | v2        | Delta                                   |
| -------------------------- | -------------- | --------- | --------------------------------------- |
| ~~Lines of HTML~~ done — verified — v1 to v2 snapshot, historical | ~~1,670~~ | ~~2,116~~ | ~~+446~~ |
| ~~Sections~~ done — verified — v1 to v2 snapshot, historical | ~~8~~ | ~~12~~ | ~~+4 (comparison, CI/CD, 3 feature cards)~~ |
| ~~Features covered~~ done — verified — v1 to v2 snapshot, historical | ~~6 of ~21~~ | ~~9 of ~21~~ | ~~+3 (SARIF, Health, Incremental)~~ |
| ~~Accessibility score (est.)~~ done — verified — v1 to v2 snapshot, historical | ~~~45/100~~ | ~~~85/100~~ | ~~+40~~ |
| ~~SEO score (est.)~~ done — verified — v1 to v2 snapshot, historical | ~~~30/100~~ | ~~~75/100~~ | ~~+45~~ |
| ~~Factual errors~~ done — verified — v1 to v2 snapshot, historical | ~~6~~ | ~~0~~ | ~~-6~~ |
| ~~ARIA attributes~~ done — verified — v1 to v2 snapshot, historical | ~~3~~ | ~~30+~~ | ~~+27~~ |
| ~~Meta tags~~ done — verified — v1 to v2 snapshot, historical | ~~4~~ | ~~14~~ | ~~+10~~ |
| ~~Install methods shown~~ done — verified — v1 to v2 snapshot, historical | ~~1 (go install)~~ | ~~2 (+brew)~~ | ~~+1~~ |
