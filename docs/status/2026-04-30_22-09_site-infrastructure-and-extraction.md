# Landing Page v3 — Site Infrastructure + Extraction Status

**Date:** 2026-04-30 22:09
**Session:** CSS/JS extraction, Firebase deploy workflow, OG image, sitemap, robots, 404 page
**Previous Report:** `2026-04-30_21-40_landing-page-v2-firebase-and-fixes.md`

---

## Executive Summary

Completed site infrastructure: extracted CSS/JS to external files, added Firebase deploy workflow, OG image, sitemap.xml, robots.txt, and custom 404 page. The site is now a proper static site ready for Firebase deployment with all production infrastructure in place.

---

## A) FULLY DONE

| #   | Item                                        | Details                                                          |
| --- | ------------------------------------------- | ---------------------------------------------------------------- |
| 1   | `.firebase/` in `.gitignore`                | Prevents Firebase cache from being committed                     |
| 2   | CSS extracted to `site/style.css`           | 1,191 lines, cached with 31536000s immutable headers             |
| 3   | JS extracted to `site/app.js`               | 199 lines, cached with 31536000s immutable headers               |
| 4   | HTML reduced to 725 lines                   | From original 2,116 lines (66% reduction)                        |
| 5   | Firebase deploy GitHub Actions              | `.github/workflows/deploy-site.yml` with `FIREBASE_TOKEN` secret |
| 6   | OG image                                    | `site/og-image.svg` with branded design (1200x630)               |
| 7   | `og:image` meta tag + `summary_large_image` | Updated in HTML head                                             |
| 8   | `sitemap.xml`                               | Standard sitemap for art-dupl.web.app                            |
| 9   | `robots.txt`                                | Allows all crawlers, references sitemap                          |
| 10  | Custom 404 page                             | `site/404.html` with "No clones found here" theme                |

### Site File Inventory

| File                | Lines     | Size      | Purpose              |
| ------------------- | --------- | --------- | -------------------- |
| `site/index.html`   | 725       | 37KB      | Main page            |
| `site/style.css`    | 1,191     | 24KB      | All styles           |
| `site/app.js`       | 199       | 6KB       | All interactivity    |
| `site/404.html`     | 24        | 1.4KB     | Error page           |
| `site/og-image.svg` | 43        | 2.8KB     | Social sharing image |
| `site/robots.txt`   | 4         | 70B       | Crawler directives   |
| `site/sitemap.xml`  | 9         | 267B      | SEO sitemap          |
| **Total**           | **2,195** | **~72KB** | —                    |

### Firebase Hosting Stack

- **Config**: `firebase.json` with cache headers (1yr for CSS/JS/fonts, 1hr for HTML)
- **Deploy workflow**: GitHub Actions on push to `fork` branch when `site/**` changes
- **Clean URLs**: Enabled
- **404**: Custom page matching site design
- **Project**: `art-dupl` (needs `FIREBASE_TOKEN` secret in GitHub)

## B) PARTIALLY DONE

| #   | Item                    | What's Missing                                     |
| --- | ----------------------- | -------------------------------------------------- |
| 1   | Firebase deployment     | Config ready but `firebase login` + deploy not run |
| 2   | `FIREBASE_TOKEN` secret | Needs to be added to GitHub repo settings          |
| 3   | CSS/JS minification     | External files are unminified                      |

## C) NOT STARTED

| #   | Item                                | Priority | Effort |
| --- | ----------------------------------- | -------- | ------ |
| 1   | Suffix tree canvas visualization    | HIGH     | 90min  |
| 2   | Animated hero stat counters         | LOW      | 20min  |
| 3   | Typewriter terminal effect          | LOW      | 20min  |
| 4   | `simple-json` and `csv` format tabs | LOW      | 15min  |
| 5   | SDK/API section                     | MED      | 20min  |
| 6   | Config file section                 | LOW      | 10min  |
| 7   | Go Report Card + Codecov badges     | LOW      | 5min   |
| 8   | Before/after code comparison visual | MED      | 30min  |
| 9   | Dark/light theme toggle             | LOW      | 20min  |
| 10  | Self-host Google Fonts              | LOW      | 15min  |
| 11  | CSS/JS minification build step      | LOW      | 10min  |
| 12  | Interactive live demo (Cloud Run)   | HIGH     | 4hr    |
| 13  | Custom domain DNS                   | MED      | 30min  |

## D) TOTALLY FUCKED UP

**Nothing.** All previous bugs remain fixed. No regressions introduced.

## E) WHAT WE SHOULD IMPROVE

1. **Suffix tree visualization** — The single highest-impact design improvement. Replace the generic particle canvas with an animated suffix tree growing from source code. This would be the unique visual identity.

2. **CSS/JS minification** — Add a simple build step (e.g., `cat style.css | cssnano > style.min.css`) before deploy. Could be in the GitHub Actions workflow.

3. **Interactive demo** — The most impactful feature addition. Would require a backend (Cloud Run) running art-dupl on user-submitted code.

4. **Self-host Google Fonts** — For privacy and faster loads (eliminates 2 DNS lookups + HTTP/2 connection to Google).

5. **The OG image is SVG** — Some social platforms don't render SVG. Consider generating a PNG version.

## F) TOP #25 NEXT STEPS (by Impact × Ease)

| #   | Task                                        | Impact | Effort | Type     |
| --- | ------------------------------------------- | ------ | ------ | -------- |
| 1   | Run `firebase login` + `firebase deploy`    | HIGH   | 5min   | Deploy   |
| 2   | Add `FIREBASE_TOKEN` to GitHub secrets      | HIGH   | 2min   | Deploy   |
| 3   | Generate PNG version of OG image            | MED    | 10min  | SEO      |
| 4   | Add Go Report Card + Codecov badges to page | LOW    | 5min   | Social   |
| 5   | Add config file section to page             | LOW    | 10min  | Content  |
| 6   | Add SDK/API section to page                 | MED    | 20min  | Content  |
| 7   | Add `simple-json` + `csv` format tabs       | LOW    | 15min  | Content  |
| 8   | Add before/after code comparison visual     | MED    | 30min  | Design   |
| 9   | Build suffix tree canvas visualization      | HIGH   | 90min  | Design   |
| 10  | Add animated hero stat counters             | LOW    | 20min  | Polish   |
| 11  | Add typewriter effect to terminal demo      | LOW    | 20min  | Polish   |
| 12  | Add dark/light theme toggle                 | LOW    | 20min  | UX       |
| 13  | Self-host Google Fonts                      | LOW    | 15min  | Perf     |
| 14  | Add CSS/JS minification to deploy workflow  | LOW    | 10min  | Perf     |
| 15  | Run Lighthouse audit on deployed site       | MED    | 10min  | Quality  |
| 16  | Setup custom domain DNS                     | MED    | 30min  | Deploy   |
| 17  | Add Plausible/Fathom analytics              | LOW    | 10min  | Tracking |
| 18  | Build interactive live demo (Cloud Run)     | HIGH   | 4hr    | Feature  |
| 19  | Add `sitemap.xml` to `robots.txt`           | DONE   | —      | —        |
| 20  | Verify W3C HTML validation                  | MED    | 10min  | Quality  |
| 21  | Add preconnect hints for fonts              | LOW    | 2min   | Perf     |
| 22  | Add `manifest.json` for PWA                 | LOW    | 10min  | Feature  |
| 23  | Add Apple Touch Icon                        | LOW    | 5min   | iOS      |
| 24  | Test on real mobile devices                 | MED    | 15min  | QA       |
| 25  | Run axe-core accessibility audit            | MED    | 15min  | A11y     |

## G) TOP #1 QUESTION

**Ready to deploy?** The site is production-ready. You need to:

1. `npm install -g firebase-tools`
2. `firebase login`
3. `firebase deploy --only hosting`

Or set up the `FIREBASE_TOKEN` secret in GitHub and the deploy workflow will handle it automatically on push.

---

## Metrics

| Metric           | v1    | v2    | v3   | Delta        |
| ---------------- | ----- | ----- | ---- | ------------ |
| HTML lines       | 1,670 | 2,116 | 725  | -57% from v2 |
| Total site files | 1     | 1     | 7    | +6           |
| Total site size  | 67KB  | 67KB  | 72KB | +5KB         |
| External CSS     | No    | No    | Yes  | Cached 1yr   |
| External JS      | No    | No    | Yes  | Cached 1yr   |
| OG image         | No    | No    | Yes  | Social ready |
| Sitemap          | No    | No    | Yes  | SEO          |
| Robots.txt       | No    | No    | Yes  | SEO          |
| 404 page         | No    | No    | Yes  | UX           |
| Deploy workflow  | No    | No    | Yes  | Auto-deploy  |
