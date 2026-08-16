# Status Report: Public Presence Overhaul — Complete Session

**Date:** 2026-07-13 22:14
**Session Scope:** Making art-dupl fully public-ready: README rewrite, Astro wiki website, GitHub metadata, DNS/Firebase hosting configuration
**Branch:** fork

> **✅ FULLY RESOLVED (updated 2026-07-16):** All work from this session was **committed** and the website deployed. Commits: `2d3bc35` (Astro website + README rewrite + GitHub metadata), `23f7203` (CI/CD pipeline overhaul + security headers + domain migration to `art-dupl.lars.software`). The "NOTHING IS COMMITTED" critical risk is RESOLVED. DNS CNAME was added to the domains repo. The `FIREBASE_TOKEN` → `GOOGLE_APPLICATION_CREDENTIALS` migration in the deploy workflow is complete. Manual Firebase console domain setup + Terraform apply still needed for the custom domain SSL cert.

---

## Executive Summary

Over three work blocks this session, I built a complete Astro + Starlight documentation website (15 pages), rewrote the README, updated GitHub metadata, and configured DNS + Firebase hosting for the custom domain `art-dupl.lars.software`. The website builds successfully and the deploy workflow is ready. However, **nothing is committed**, no visual QA was ever performed, and several infrastructure steps require manual action before the site goes live.

---

## a) FULLY DONE

| #  | Item                             | Details                                                                                                                                                                                                                                                                                 |
| -- | -------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1  | **README.md rewritten**          | Sales-page style, accurate threshold (5 statements not 15), 5 badges (Go Report Card, CI, License, Go Reference, Website), comparison table, all features, `just` references removed, URLs point to `art-dupl.lars.software`                                                            |
| 2  | **Astro website infrastructure** | `package.json`, `tsconfig.json`, `astro.config.mjs`, `.firebaserc`, `.gitignore`, `.node-version`, `.htmlvalidate.json`, `flake.nix`, `content.config.ts` — all in `website/`                                                                                                           |
| 3  | **Landing page**                 | Hero with live GitHub stars fetch, animated badge, terminal mockup, 6-feature grid, 4-step pipeline, comparison matrix (vs dupl/jscpd), 7 output format cards, 5 use cases, CTA                                                                                                         |
| 4  | **Starlight docs (13 pages)**    | Installation, Quick Start, Detection Methods, Output Formats, CI/CD, Filtering, Performance, SDK, Configuration, CLI Flags, Changelog, Contributing, Related Tools                                                                                                                      |
| 5  | **Brand theming**                | Amber/gold (#e8a020) accent, Syne + JetBrains Mono fonts, dark/light mode with toggle, starlight.css color mapping                                                                                                                                                                      |
| 6  | **Public assets**                | `favicon.svg`, `manifest.json`, `robots.txt`, `theme-init.js`, `animations.js`, `copy-code.js`, `header.js`                                                                                                                                                                             |
| 7  | **GitHub metadata**              | Description (200 chars), Homepage URL set to `https://art-dupl.lars.software`, 12 topics added                                                                                                                                                                                          |
| 8  | **Firebase config upgraded**     | `website/firebase.json` now has full security headers (HSTS, X-Frame-Options DENY, Permissions-Policy, CORP, COOP, X-Content-Type-Options, X-XSS-Protection, Referrer-Policy), split caching strategy (immutable for assets, must-revalidate for HTML), `/docs/*` redirect, 404 caching |
| 9  | **Deploy workflow upgraded**     | Two-job pattern (build → deploy) with `pnpm install --frozen-lockfile`, `astro check`, HTML validation, artifact upload/download, `GOOGLE_APPLICATION_CREDENTIALS` service account auth, concurrency groups, PR builds (no deploy)                                                      |
| 10 | **DNS record added**             | CNAME `art-dupl.lars.software` → `art-dupl.web.app.` added to `/home/lars/projects/domains/lars.software.tf` matching the exact pattern of gogenfilter, atomicwrite, and all other subdomains                                                                                           |
| 11 | **URLs updated everywhere**      | All canonical URLs updated from `art-dupl.web.app` to `art-dupl.lars.software` (README, astro.config.mjs, config.ts, package.json, robots.txt)                                                                                                                                          |
| 12 | **Old static site removed**      | `site/` directory, root `firebase.json`, root `.firebaserc` trashed. Stale `.gitignore` entry removed                                                                                                                                                                                   |
| 13 | **Build verified**               | Website builds to 15 pages successfully. `go build ./...` still passes                                                                                                                                                                                                                  |
| 14 | **package-lock.json exists**     | Generated locally, ready for commit so CI `pnpm install --frozen-lockfile` works                                                                                                                                                                                                        |

---

## b) PARTIALLY DONE

| # | Item                            | What's Done                                                                                                         | What's Missing                                                                                                                  |
| - | ------------------------------- | ------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| 1 | **Website visual verification** | Builds to 15 HTML pages successfully                                                                                | Never previewed in browser. No visual QA. Colors/layout/icons/responsiveness could have rendering issues                        |
| 2 | **Deploy workflow CI**          | YAML written with two-job pattern, `pnpm install --frozen-lockfile`, astro check, HTML validation, artifact passing | Not tested end-to-end. Uses `FIREBASE_SERVICE_ACCOUNT` secret which may not exist yet (old workflow used `FIREBASE_TOKEN`)      |
| 3 | **DNS configuration**           | CNAME record added to Terraform `.tf` file                                                                          | Terraform not applied. Firebase custom domain not added in console. SSL cert not provisioned                                    |
| 4 | **OG image**                    | Old `site/og-image.svg` was removed                                                                                 | Website has NO OG image. Social sharing will show no preview. No `<meta og:image>` tags in LandingLayout                        |
| 5 | **HTML/TS validation**          | `.htmlvalidate.json` configured, `astro check` in CI workflow                                                       | `astro check` never run locally. HTML validation never run. Both set to `continue-on-error: true` in CI so errors won't block   |
| 6 | **Doc accuracy**                | Written from AGENTS.md/FEATURES.md/code                                                                             | Some specifics not verified against source (e.g., exact JSON output structure, SARIF schema, SDK code examples compile-ability) |

---

## c) NOT STARTED

| #  | Item                                      | Impact                                                                                                                                                                                                                                                                                                  |
| -- | ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1  | **Commit the changes**                    | Nothing is committed. All work is uncommitted in working tree across two repos (art-dupl + domains)                                                                                                                                                                                                     |
| 2  | **Push and deploy**                       | Website not deployed. Firebase CI hasn't run. DNS not applied                                                                                                                                                                                                                                           |
| 3  | **Add custom domain in Firebase console** | Required for `art-dupl.lars.software` to serve content. Must be done manually in Firebase Hosting console                                                                                                                                                                                               |
| 4  | **Terraform apply**                       | DNS CNAME exists in `.tf` file but not applied to Namecheap                                                                                                                                                                                                                                             |
| 5  | **Provision SSL certificate**             | Firebase auto-provisions after custom domain is added and DNS propagates. Requires the above two steps first                                                                                                                                                                                            |
| 6  | **`FIREBASE_SERVICE_ACCOUNT` secret**     | New deploy workflow uses service account JSON auth (matching gogenfilter pattern). Old workflow used `FIREBASE_TOKEN`. This secret must be added to GitHub                                                                                                                                              |
| 6  | **Visual QA**                             | Never ran `pnpm run preview` or opened any page in a browser                                                                                                                                                                                                                                            |
| 7  | **Mobile responsive testing**             | Tailwind responsive classes used but never tested                                                                                                                                                                                                                                                       |
| 8  | **Link checking**                         | No verification that all internal doc links resolve correctly                                                                                                                                                                                                                                           |
| 9  | **Lighthouse audit**                      | Not run                                                                                                                                                                                                                                                                                                 |
| 10 | **OG image generation**                   | No social sharing image for the website                                                                                                                                                                                                                                                                 |
| 11 | **Stale repo docs cleanup**               | Root has HOW_TO_USE.md, USAGE.md, SDK_DESIGN.md, WHAT_THIS_PROJECT_IS_NOT.md, PARTS.md, PERFORMANCE_OPTIMIZATION.md, MIGRATION_QUICK_START.md, MIGRATION_TO_NIX_FLAKES_PROPOSAL.md, branching-flow-analysis.md, branching-flow-findings-table.md — these overlap with website docs and confuse visitors |
| 12 | **CONTRIBUTING.md update**                | README links to it, but it may still reference `just` commands                                                                                                                                                                                                                                          |
| 13 | **Preview deployment for PRs**            | No Firebase preview channel setup for pull request previews                                                                                                                                                                                                                                             |

---

## d) TOTALLY FUCKED UP / RISK AREAS

| # | Issue                                     | Severity     | Details                                                                                                                                                                                                                    |
| - | ----------------------------------------- | ------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | **NOTHING IS COMMITTED**                  | **CRITICAL** | All work across both repos is uncommitted. If session ends or machine restarts, work is only on disk. Two repos need commits: art-dupl (website + README + workflow + .gitignore) and domains (DNS record)                 |
| 2 | **No visual QA at all**                   | **HIGH**     | Built 37 source files (15 components, 13 docs, 5 data files, 4 styles/layouts) that render to 15 pages without ever looking at any of them in a browser. Could have broken layouts, invisible text, wrong icon paths, etc. |
| 3 | **Secret mismatch**                       | **HIGH**     | Deploy workflow references `FIREBASE_SERVICE_ACCOUNT` but old workflow used `FIREBASE_TOKEN`. If the new secret isn't added before the old one is removed, deploy will fail. The CI may fail on first push                 |
| 4 | **DNS not applied**                       | **MEDIUM**   | Terraform CNAME added to `.tf` file but not `terraform apply`'d. Until applied, `art-dupl.lars.software` resolves to nothing                                                                                               |
| 5 | **Firebase custom domain not configured** | **MEDIUM**   | Even with DNS, Firebase won't serve the custom domain until it's added in the Firebase console. The `art-dupl.web.app` default URL would still work but the custom domain won't                                            |
| 6 | **Hero code is fabricated**               | **LOW**      | The hero terminal output is illustrative content, not actual art-dupl output. The format may differ from reality                                                                                                           |
| 7 | **Doc code examples untested**            | **LOW**      | SDK guide has Go code examples that were not compile-tested against the actual `pkg/artdupl` API                                                                                                                           |

---

## e) WHAT WE SHOULD IMPROVE

| #  | Area                                    | Improvement                                                                                                                                                                                                                                            |
| -- | --------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 1  | **Always commit incrementally**         | Should have committed after README, after website build, after Firebase config — not left everything uncommitted                                                                                                                                       |
| 2  | **Visual QA is non-negotiable**         | Should have run `pnpm run preview` and verified at least the landing page before declaring done                                                                                                                                                        |
| 3  | **Test code examples**                  | SDK guide code examples should be compile-tested                                                                                                                                                                                                       |
| 4  | **OG image**                            | Need a proper social sharing image — every other LarsArtmann project site has one                                                                                                                                                                      |
| 5  | **Doc accuracy audit**                  | Cross-reference every flag default and JSON output field against actual source code                                                                                                                                                                    |
| 6  | **Stale docs cleanup**                  | 10+ stale root-level docs overlap with website and create confusion                                                                                                                                                                                    |
| 7  | **`astro check` should be blocking**    | CI has `continue-on-error: true` — TypeScript errors won't block builds                                                                                                                                                                                |
| 8  | **CONTRIBUTING.md**                     | Likely still references `just` commands — needs updating to match README                                                                                                                                                                               |
| 9  | **Website `flake.nix` app integration** | art-dupl root `flake.nix` doesn't reference the website — no `nix run .#website-dev`                                                                                                                                                                   |
| 10 | **Firebase `site` vs `target`**         | go-atomic-write and gogenfilter use `"target": "..."` with `.firebaserc` targets pointing to the `lars-software` project. art-dupl uses the standalone `art-dupl` Firebase project with no target. This works but differs from the established pattern |

---

## f) Up to 50 Things to Get Done Next

### CRITICAL — Do These First

1. **Commit all art-dupl changes** (website/, README.md, .github/workflows/deploy-site.yml, .gitignore)
2. **Commit domains changes** (lars.software.tf DNS record)
3. **Add `FIREBASE_SERVICE_ACCOUNT` secret** to GitHub (service account JSON for `art-dupl` Firebase project)
4. **Remove old `FIREBASE_TOKEN` secret** from GitHub if it exists
5. **Push both repos** to trigger CI
6. **Verify deploy workflow succeeds** end-to-end
7. **Run `pnpm run preview` locally** and visually QA the landing page
8. **Fix any visual issues** found during preview

### DNS & Firebase Domain Setup

9. **Add custom domain `art-dupl.lars.software`** in Firebase Hosting console
10. **Apply Terraform** (`cd domains && terraform plan && terraform apply`) to create the CNAME
11. **Wait for SSL cert provisioning** (Firebase auto-provisions after DNS propagates, can take minutes to hours)
12. **Verify `art-dupl.lars.software`** serves the new website
13. **Set up Firebase redirects** from `art-dupl.web.app` to `art-dupl.lars.software` (or keep both)

### Website Polish

14. **Generate OG image** for social sharing (`website/public/og-image.svg` or dynamic OG API route)
15. **Add `<meta og:image>` tags** to LandingLayout.astro
16. **Run `pnpm dlx astro check`** locally and fix TypeScript errors
17. **Run HTML validation** (`pnpm dlx html-validate "dist/**/*.html"`) and fix errors
18. **Remove `continue-on-error: true`** from astro check and HTML validation in CI once errors are fixed
19. **Run Lighthouse audit** on the deployed site
20. **Fix Lighthouse issues** (performance, accessibility, SEO, best practices)
21. **Test mobile responsive** layout via browser DevTools
22. **Verify all sidebar links** resolve correctly in Starlight docs
23. **Add "Edit this page on GitHub"** links in Starlight config
24. **Verify favicon rendering** across browsers
25. **Add cross-links between doc pages** (inline "see also" references)

### README & Repo Cleanup

26. **Audit all README claims** against actual code behavior
27. **Verify pkg.go.dev badge** URL works and module is indexed
28. **Verify Go Report Card** still works
29. **Update CONTRIBUTING.md** — remove `just` references, match README's `go build`/`go test` commands
30. **Consolidate stale docs**: HOW_TO_USE.md, USAGE.md, SDK_DESIGN.md → archive or delete (website is now source of truth)
31. **Audit WHAT_THIS_PROJECT_IS_NOT.md, PARTS.md, PERFORMANCE_OPTIMIZATION.md** for relevance
32. **Archive `docs/status/archive/`** — hundreds of historical status files clutter the repo
33. **Add Go module path consistency** — README uses `github.com/LarsArtmann/art-dupl` (capital L, A) but actual go.mod may differ

### Docs Accuracy

34. **Verify `--test-threshold` default formula** (`max(30, threshold)`) against `cmd/flags.go`
35. **Verify all flag defaults** in cli-flags.mdx against actual source code
36. **Verify JSON output structure** described in output-formats.mdx — run `art-dupl --json` and compare
37. **Verify SARIF output** meets GitHub's expected schema
38. **Compile-test SDK code examples** in a temporary Go file
39. **Add a Troubleshooting doc page** (there's a `docs/TROUBLESHOOTING.md` already to reference)
40. **Add actual `art-dupl` output** to hero code instead of fabricated content

### SEO & Discovery

41. **Submit sitemap** to Google Search Console for `art-dupl.lars.software`
42. **Add structured data** to docs pages (not just landing page)
43. **Submit to awesome-go** or other Go discovery platforms
44. **Consider canonical URL** from `art-dupl.web.app` → `art-dupl.lars.software` in Firebase config

### Infrastructure

45. **Add `website` apps to root `flake.nix`** (`nix run .#website-dev`, `nix run .#website-build`)
46. **Set up Firebase preview channels** for PR previews
47. **Add lighthouse CI** (gogenfilter has `.github/workflows/lighthouse.yml`)
48. **Add Dependabot config** for website pnpm dependencies
49. **Consider shared Firebase project** — move art-dupl hosting to `lars-software` project like gogenfilter/atomicwrite for consistency
50. **Add `lighthouserc.json`** for automated performance regression checks

---

## g) Top 2 Questions I Cannot Answer Myself

### 1. Should art-dupl use the shared `lars-software` Firebase project (like gogenfilter and go-atomic-write) or keep its standalone `art-dupl` project?

**Context:** gogenfilter and go-atomic-write both deploy to the `lars-software` Firebase project using hosting targets (`.firebaserc` has `"target": "gogenfilter"` / `"target": "atomicwrite"`). art-dupl currently uses a standalone `art-dupl` Firebase project with no targets. Using the shared project would be consistent but requires migrating the Firebase project, updating `.firebaserc`, and changing the deploy command. I don't know if the standalone project has other resources (databases, auth, etc.) that would complicate migration.

### 2. Should I commit and push now, or do you want to do visual QA and manual Firebase/DNS steps first?

**Context:** All changes are uncommitted across two repos. The deploy workflow and DNS config require manual steps that I cannot perform (Firebase console domain setup, Terraform apply, GitHub secrets). You may want to: (a) commit + push everything now and fix forward, (b) preview locally first, or (c) do the Firebase/DNS setup first so the first deploy goes to the right place.
