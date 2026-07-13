# Status Report: Public Presence Overhaul (README + Website + GitHub)

**Date:** 2026-07-13 21:18  
**Session Scope:** Making art-dupl public-ready: README, wiki website, GitHub metadata  
**Branch:** fork

---

## Executive Summary

Built a complete Astro + Starlight documentation website, rewrote the README as a sales page, and updated GitHub metadata (description, topics, homepage URL). The website builds successfully (15 pages). However, several important details were missed or rushed.

---

## a) FULLY DONE

| #   | Item                             | Details                                                                                                                                                                    |
| --- | -------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **README.md rewritten**          | Sales-page style, accurate threshold (5 statements not 15), 5 badges, comparison table, all features covered, `just` references removed                                    |
| 2   | **Astro website infrastructure** | `package.json`, `tsconfig.json`, `astro.config.mjs`, `firebase.json`, `.firebaserc`, `.gitignore`, `.node-version`, `.htmlvalidate.json`, `flake.nix`, `content.config.ts` |
| 3   | **Landing page**                 | Hero with GitHub stars fetch, animated badge, terminal mockup, 6-feature grid, 4-step pipeline, comparison matrix, 7 output format cards, 5 use cases, CTA                 |
| 4   | **Starlight docs (13 pages)**    | Installation, Quick Start, Detection Methods, Output Formats, CI/CD, Filtering, Performance, SDK, Configuration, CLI Flags, Changelog, Contributing, Related Tools         |
| 5   | **Brand theming**                | Amber/gold (#e8a020) accent matching existing site, Syne + JetBrains Mono fonts, dark/light mode, starlight.css color mapping                                              |
| 6   | **Public assets**                | favicon.svg, manifest.json, robots.txt, theme-init.js, animations.js, copy-code.js, header.js                                                                              |
| 7   | **GitHub metadata**              | Description (200 chars), Homepage URL set to `https://art-dupl.web.app`, 12 topics added                                                                                   |
| 8   | **Deploy workflow**              | `.github/workflows/deploy-site.yml` updated for Astro build (`npm ci` + `npm run build` + Firebase deploy from `website/`)                                                 |
| 9   | **Old site removed**             | `site/` directory, root `firebase.json`, root `.firebaserc` trashed. Stale `.gitignore` entry (`!site/*.html`) removed                                                     |
| 10  | **Go build verified**            | `go build ./...` still passes after all changes                                                                                                                            |

---

## b) PARTIALLY DONE

| #   | Item                            | What's Done                                                         | What's Missing                                                                                                                  |
| --- | ------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **Website visual verification** | Builds to 15 pages successfully                                     | Never previewed in browser — no visual QA. Colors/layout could have rendering issues                                            |
| 2   | **README badge accuracy**       | 5 badges added (Go Report Card, CI, License, Go Reference, Website) | codecov badge removed — was it intentionally abandoned? Also Go Reference badge may 404 if pkg.go.dev hasn't indexed the module |
| 3   | **Deploy workflow CI**          | YAML written with `npm ci`, cache, build, deploy                    | Not tested end-to-end. `npm ci` requires `package-lock.json` to be committed — it exists but is untracked                       |
| 4   | **Starlight sidebar config**    | 13 docs pages + sidebar groups configured                           | Sidebar may have ordering issues or missing pages — not visually verified                                                       |
| 5   | **OG image**                    | Old `site/og-image.svg` was removed                                 | Website has NO OG image. Social sharing will show no preview image                                                              |
| 6   | **Website `.gitignore`**        | Written with `dist/`, `.astro/`, `node_modules/`                    | Root `.gitignore` still has `dist/` globally — fine for website but could be confusing                                          |

---

## c) NOT STARTED

| #   | Item                                                     | Impact                                                                                                                                        |
| --- | -------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **Commit the changes**                                   | Nothing is committed. All work is uncommitted in working tree                                                                                 |
| 2   | **Push and deploy**                                      | Website not deployed. Firebase CI hasn't run. Site is still old version at `art-dupl.web.app`                                                 |
| 3   | **package-lock.json committed**                          | Exists locally but untracked. CI (`npm ci`) will fail without it                                                                              |
| 4   | **HTML validation**                                      | `.htmlvalidate.json` exists but `npm run build` doesn't run validation. No HTML validation done                                               |
| 5   | **TypeScript check**                                     | `npm run typecheck` (astro check) never run                                                                                                   |
| 6   | **Lighthouse / performance audit**                       | Not run                                                                                                                                       |
| 7   | **Accessibility audit**                                  | Not run (though skip-link, ARIA labels, focus styles are present)                                                                             |
| 8   | **Mobile responsive testing**                            | Tailwind responsive classes used but never tested                                                                                             |
| 9   | **Link checking**                                        | No verification that all internal doc links resolve                                                                                           |
| 10  | **Doc content accuracy audit**                           | Docs written from AGENTS.md/FEATURES.md but some specifics (e.g., exact `--test-threshold` default formula) not verified against source code  |
| 11  | **Existing `site/` Firebase redirects**                  | Old Firebase hosting may have redirect rules. No migration of redirect/rewrite config from old `firebase.json` to new `website/firebase.json` |
| 12  | **og-image generation**                                  | No OG image for the new website                                                                                                               |
| 13  | **README `HOW_TO_USE.md` and `USAGE.md` reconciliation** | README links to CONTRIBUTING.md and TESTING.md but doesn't mention HOW_TO_USE.md, USAGE.md, SDK_DESIGN.md which still exist in repo           |

---

## d) TOTALLY FUCKED UP / RISK AREAS

| #   | Issue                                              | Severity | Details                                                                                                                                                                                                                                         |
| --- | -------------------------------------------------- | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **Website has NO commit**                          | CRITICAL | All website files are untracked. If session ends, work is only on disk — not in git history                                                                                                                                                     |
| 2   | **package-lock.json not committed**                | HIGH     | Deploy workflow uses `npm ci` which fails without lock file in git                                                                                                                                                                              |
| 3   | **Old Firebase hosting still serving old `site/`** | HIGH     | `art-dupl.web.app` still shows old hand-written HTML. Until `website/dist/` is deployed, the public URL shows stale content                                                                                                                     |
| 4   | **No visual QA whatsoever**                        | HIGH     | Built 15 pages and 15 components without ever looking at any of them in a browser. Could have broken layouts, wrong colors, invisible text, etc.                                                                                                |
| 5   | **Hero code is fake**                              | MEDIUM   | The hero terminal output is fabricated illustrative content, not actual art-dupl output. Format may differ from reality                                                                                                                         |
| 6   | **`website/` not in root `.gitignore` exceptions** | LOW      | Root `.gitignore` ignores `dist/` globally. This is fine (`website/.gitignore` handles it) but could be confusing                                                                                                                               |
| 7   | **Deleted root firebase.json/.firebaserc**         | MEDIUM   | The old `firebase.json` at root was the active deploy config for `site/`. New `website/firebase.json` points to `dist/` but the deploy workflow triggers on `website/**` path changes — if someone edits root firebase config, it won't trigger |

---

## e) WHAT WE SHOULD IMPROVE

| #   | Area                            | Improvement                                                                                                                                                                                |
| --- | ------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 1   | **Visual QA process**           | Should have run `npm run preview` and screenshot/verified at least the landing page before declaring done                                                                                  |
| 2   | **OG image**                    | Generate a proper OG image (`website/public/og-image.svg` or dynamic OG via API route like gogenfilter does)                                                                               |
| 3   | **Doc accuracy**                | Cross-reference every flag default in CLI Flags doc against actual `cmd/flags.go` source                                                                                                   |
| 4   | **README stale doc references** | Repository has HOW_TO_USE.md, USAGE.md, SDK_DESIGN.md, WHAT_THIS_PROJECT_IS_NOT.md, PERFORMANCE_OPTIMIZATION.md, PARTS.md, MIGRATION_QUICK_START.md — README doesn't mention most of these |
| 4   | **Search console / analytics**  | No Google Search Console or analytics setup mentioned                                                                                                                                      |
| 5   | **README CONTRIBUTING.md**      | README links to CONTRIBUTING.md but that file may be stale (mentions `just` commands?)                                                                                                     |
| 6   | **Doc cross-linking**           | Docs should link to each other more. E.g., "see Detection Methods" links within output-formats.mdx                                                                                         |
| 7   | **Favicon refinement**          | Current favicon is just "AD" text in a gold square. Could be more distinctive                                                                                                              |
| 8   | **Website performance**         | Should add `preloading` hints, verify font loading strategy                                                                                                                                |
| 9   | **Schema.org completeness**     | Landing page has SoftwareApplication schema but docs pages have none                                                                                                                       |
| 10  | **Stale repo docs cleanup**     | Many docs in repo root (USAGE.md, HOW_TO_USE.md, PARTS.md, etc.) are likely stale and confuse visitors                                                                                     |

---

## f) Up to 50 Things to Get Done Next

### Immediate (Blocks deployment)

1. Commit all website + README + metadata changes
2. Push to trigger deploy-site workflow
3. Verify `npm ci` succeeds in CI (needs committed `package-lock.json`)
4. Run `npm run typecheck` — fix any TypeScript errors
5. Preview website locally and do visual QA on landing page
6. Preview at least 2-3 doc pages to verify Starlight renders correctly
7. Fix any visual issues found during preview
8. Verify deploy workflow succeeds end-to-end
9. Verify `art-dupl.web.app` serves the new site after deploy

### Website Polish

10. Generate OG image for social sharing
11. Add OG image meta tags to LandingLayout
12. Run HTML validation (`html-validate dist/**/*.html`)
13. Fix any HTML validation errors
14. Run Lighthouse audit on landing page
15. Fix Lighthouse issues (performance, accessibility, SEO, best practices)
16. Test mobile responsive layout (DevTools device emulation)
17. Fix mobile layout issues if any
18. Add internal cross-links between doc pages (next/prev, inline links)
19. Verify all sidebar links resolve
20. Add "Edit this page" links to GitHub in Starlight config
21. Check favicon rendering across browsers

### README Polish

22. Audit all README claims against actual code behavior
23. Verify pkg.go.dev badge URL works
24. Verify Go Report Card still works for the repo
25. Reconcile stale docs: decide which of HOW_TO_USE.md, USAGE.md, SDK_DESIGN.md, etc. to keep/link
26. Add a "Contributing" call-to-action section
27. Consider adding a project logo/banner image

### Docs Accuracy

28. Verify `--test-threshold` default formula (`max(30, threshold)`) in CLI Flags doc
29. Verify all flag defaults in cli-flags.mdx against `cmd/flags.go`
30. Verify JSON output structure described in output-formats.mdx matches actual output
31. Verify SARIF output format matches GitHub's expected schema
32. Add code examples that can actually be run (test them)
33. Verify SDK code examples compile against actual `pkg/artdupl` API
34. Add a "Troubleshooting" doc page (there's a `docs/TROUBLESHOOTING.md` already)

### SEO & Discovery

35. Submit sitemap to Google Search Console
36. Add canonical URLs verification
37. Add structured data to docs pages (not just landing page)
38. Write a comparison blog post or "alternatives" page
39. Add the website to the GitHub repo "About" section (done) and verify it renders
40. Consider adding the site to Go discovery platforms (awesome-go, etc.)

### Stale Content Cleanup

41. Audit and consolidate USAGE.md, HOW_TO_USE.md — these now overlap with website docs
42. Audit WHAT_THIS_PROJECT_IS_NOT.md — useful or stale?
43. Audit PARTS.md — is this still relevant?
44. Audit MIGRATION_QUICK_START.md, MIGRATION_TO_NIX_FLAKES_PROPOSAL.md — still needed?
45. Audit PERFORMANCE_OPTIMIZATION.md — accurate or stale?
46. Remove or archive `docs/status/archive/` (hundreds of historical status files)

### Infrastructure

47. Add `website/` to the root `flake.nix` if cross-referencing is needed
48. Consider adding html-validate to CI as a separate job
49. Set up Firebase preview channels for PR previews
50. Add a `Makefile` target or `flake.nix` app for `nix run .#website-dev`

---

## g) Top 2 Questions I Cannot Answer Myself

### 1. Should I commit and push these changes now, or do you want to review first?

**Context:** All changes are uncommitted. The deploy workflow will only fire on push to `fork` branch. You may want to preview the website locally before it goes live, or you may be fine deploying blind and fixing forward.

### 2. What should happen to the stale root-level documentation files?

**Context:** The repo root has `HOW_TO_USE.md`, `USAGE.md`, `SDK_DESIGN.md`, `WHAT_THIS_PROJECT_IS_NOT.md`, `PARTS.md`, `PERFORMANCE_OPTIMIZATION.md`, `MIGRATION_QUICK_START.md`, and others. These now heavily overlap with the website docs. Three options:

- **A)** Keep them and link from README (redundant with website)
- **B)** Move to `docs/` or `docs/archive/` (clean root)
- **C)** Delete and let the website be the single source of truth

This is a project owner decision — I don't know if external users or CI reference these files.
