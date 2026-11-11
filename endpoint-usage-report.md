# Layer Index endpoint usage and deprecation report

Source: public access logs of layers.openembedded.org, 126 daily files covering 2026-07-08 to 2026-11-10.
Code reference: `layerindex-web` at `08e58cd` (official master, identical to the GitHub fork).

## Summary

- 18.9M log lines in total; 14.0M successful (2xx) application requests are in the `logs` table.
- Traffic is dominated by crawlers. One page type, the recipe detail page, takes 84% of all successful requests, and about 60% of that is from self-identified bots.
- The REST API is real, steady, tool-driven usage: about 295k requests, on 125 of 126 days, led by BitBake's own layer index client.
- A group of logged-in "Tools" features has essentially no usage: Bulk Change, Recipe Dependencies, the comparison-selection workflow and the branch-comparison export formats.
- Recommended for deprecation: 24 routes with strong evidence, a further group worth discussing, plus the unused API format-suffix variants.

## How to read the numbers

| Limitation | Effect |
|---|---|
| No client IPs in the published logs | "Who" is inferred from user agent and referrer only. |
| The importer keeps only 2xx responses | Redirect-only and failing endpoints are invisible in the database. Those figures come from a separate scan of the raw files and are labelled "all statuses". |
| "Bot" is a user-agent pattern match | Crawlers with a browser user agent count as non-bot, so non-bot figures are upper bounds on human use. |
| 6.9% of raw lines (1.30M) did not parse | They are malformed request lines and are excluded everywhere. |
| The Tools menu is shown only to logged-in users | Low usage of those pages reflects a small logged-in population, not broken links. |
| 126 days is one window | It includes the October release period, but rare yearly tasks could still be missed. |

A useful baseline: between 2026-08-20 and 2026-10-04 daily volume fell from 100k+ to about 10k requests, apparently because crawlers were being blocked. That 46-day "quiet window" is the best available view of human and tooling traffic and is quoted below where it changes the picture.

## Endpoint usage by category

Counts are 2xx requests over the full period unless stated. "Non-bot" is the share not matching the bot pattern.

### 1. Browsing the index (public, read-only)

| Endpoint | Requests | Non-bot | Quiet window |
|---|---:|---:|---:|
| `/layerindex/recipe/<pk>/` | 11,797,747 | 40% | 215,519 |
| `/layerindex/branch/<branch>/recipes/` | 432,711 | 33% | 53,391 |
| `/layerindex/branch/<branch>/layer/<slug>/` | 384,639 | 45% | 43,140 |
| `/layerindex/branch/<branch>/layers/` | 220,067 | 76% | 69,657 |
| `/layerindex/branch/<branch>/machines/` | 9,254 | 72% | 2,639 |
| `/layerindex/branch/<branch>/classes/` | 5,863 | 62% | 1,579 |
| `/layerindex/branch/<branch>/distros/` | 4,549 | 59% | 1,023 |
| `/layerindex/about/` | 834 | 62% | 125 |
| `/layerindex/history/` | 2,035 | 43% | 127 |

Use cases:
- Find which layer provides a recipe, and for which release. 87% of recipe-search requests carry a `q=` search term.
- Inspect a recipe: version, license, sources, patches, dependencies.
- Find a layer and read its repository URL, maintainers, dependencies and Yocto Project compatibility.
- Find the BSP layer for a machine, or a distro or class definition.

Branch spread: `master` takes 487,903 of the branch-scoped requests, followed by `dunfell` (99,480), `scarthgap` (76,209), `wrynose` (44,898) and `kirkstone` (41,284). Every old release back to `danny` still gets 13k to 25k requests, but about three quarters of that is bots.

### 2. REST API (public, read-only)

| Endpoint | Requests | Main clients |
|---|---:|---|
| `api/recipesExtended/` | 61,329 | python-requests, Python-urllib |
| `api/recipes/` | 55,991 | python-requests, Python-urllib, curl, BitBake |
| `api/layerBranches/` | 36,095 | BitBake, python-requests |
| `api/layerItems/` | 34,379 | BitBake, python-requests |
| `api/branches/` | 32,733 | BitBake, python-requests |
| `api/layerDependencies/` | 20,779 | BitBake (90%) |
| `api/` (root) | 18,642 | BitBake (92%) |
| `api/layers/` | 15,044 | Python-urllib |
| `api/layerBranches/<pk>/` | 7,336 | Python-urllib, curl |
| `api/machines/` | 3,930 | BitBake |
| `api/distros/` | 2,650 | BitBake |
| `api/branches/<pk>/` | 2,616 | Python-urllib |
| `api/layerItems/<pk>/` | 2,123 | Python-urllib, curl |
| `api/layers/<pk>/` | 1,104 | python-requests |
| `api/layerMaintainers/` | 168 | one study script, browsers |
| `api/appends/` | 133 | mixed |
| `api/recipes/<pk>/` | 128 | curl |
| `api/classes/` | 90 | mixed |
| `api/recipesExtended/<pk>/` | 78 | browsers, curl |
| `api/incFiles/` | 59 | mixed |
| `api/layerNotes/` | 27 | mixed |
| `api/classes/<pk>/`, `api/machines/<pk>/` | 3, 2 | curl |

Use cases:
- `bitbake-layers layerindex-fetch` and `layerindex-show-depends` (user agent `bitbake/lib/layerindex`, 96,554 requests): resolve and fetch a layer and its dependencies. It uses exactly nine routes: the API root, `branches`, `layerItems`, `layerBranches`, `layerDependencies`, `recipes`, `machines`, `distros` and `layers`.
- License, SBOM and firmware tooling pulling recipe metadata in bulk (`recipesExtended`, `recipes`): agents named `dafa-collect-license`, `bazel-sbom-creator`, `yocto-license-resolver`, `viss-sbom-audit`, `packages.ecosyste.ms`.
- Dependency checkers resolving package names to recipes (`rosdep_repo_check`).
- Ad-hoc scripts and manual exploration with curl.

API usage grew from 45k requests in July to 95k in October and did not drop in the quiet window, which confirms it is tool traffic, not crawling.

### 3. Recipe Reporting System (`/rrs/`)

| Endpoint | Requests | Non-bot | Quiet window |
|---|---:|---:|---:|
| `rrs/recipes/<plan>/<release>/<milestone>/` | 129,121 | 53% | 10,610 |
| `rrs/recipedetail/<plan>/<pk>/` | 101,382 | 34% | 7,276 |
| `rrs/maintainers/...` | 5,553 | 48% | 745 |
| `rrs/stats/...` | 3,103 | 28% | 77 |
| `rrs/recipesreport/...` | 2,452 | 40% | 47 |
| `rrs/maintplan/<plan>/` (redirect) | 1,360 | | |
| `rrs/` (redirect) | 25 | | |

Use cases:
- Recipe maintainers checking upstream version status and what needs upgrading.
- Looking up who maintains a recipe.
- Per-milestone maintenance statistics and the CSV recipe report.

93% of the recipe-list traffic is for the `OE-Core` maintenance plan; `meta-openembedded`, `meta-intel` and `meta-virtualization` share the rest.

### 4. Layer submission and maintenance (logged-in)

| Endpoint | 2xx | Notes |
|---|---:|---|
| `/layerindex/submit/` | 1,850 | plus 9,685 POSTs rejected with 403 |
| `/layerindex/submit/thanks/` | 16 | roughly the number of completed submissions |
| `/layerindex/review/`, `review/<slug>/` | 26, 78 | reviewers approving new layers |
| `/layerindex/branch/<branch>/edit/<slug>/` | 40 | maintainers editing a layer |
| `/layerindex/layer/<name>/publish/` | 0 | 31 requests in all: 15 redirects, 16 server errors |
| `layer/<slug>/addnote`, `editnote`, `deletenote` | 1 each | |
| `/layerindex/layer/<slug>/delete/` | 0 | |
| `/layerindex/branch/<branch>/update/<slug>/` | 0 | 1 redirect |

Use cases: submitting a new layer, reviewing and publishing it, editing layer details, adding notes.

This workflow is low-volume by nature (about 16 submissions in four months) and is core to the site, so low numbers here are not a deprecation signal.

### 5. Update history and status

| Endpoint | 2xx | Notes |
|---|---:|---|
| `/layerindex/layerupdate/<pk>/` | 58,305 | 104,482 requests in all; 44% are 404; 569 in the quiet window |
| `/layerindex/updates/<pk>/` | 2,104 | 1,267 more are 404 |
| `/layerindex/updates/` | 151 | 6 in the quiet window |

Use case: a maintainer checking why a layer's last index update failed. Almost all of the volume is crawlers walking old IDs.

### 6. Comparison, OE-Classic and analysis tools

| Endpoint | Requests (all statuses) |
|---|---:|
| `/layerindex/comparison/recipe/<pk>/` | 1,731 |
| `/layerindex/branch/<branch>/duplicates/` | 342 |
| `/layerindex/comparison/recipes/<branch>/` | 284 |
| `/layerindex/stats/` | 76 (8 non-bot) |
| `/layerindex/oe-classic/recipes/` (redirect) | 68 |
| `/layerindex/oe-classic/recipe/<pk>/` | 67 |
| `/layerindex/comparison/stats/<branch>/` | 57 |
| `/layerindex/branch_comparison/` | 44 (11 were server errors) |
| `/layerindex/oe-classic/` (redirect) | 7 |
| `branch_comparison_rst/`, `_csv/`, `_plain/` | 7, 6, 5 |
| `/layerindex/bulkchange/` | 3 |
| `comparison/search-csv/<branch>/`, `ajax/layerchecklist/<branch>/` | 2, 2 |
| `bulkchange/<pk>/search`, `edit`, `review`, `patches`, `delete` | 0 |
| `/layerindex/recipe_deps/` | 0 |
| `comparison/select/<pk>/`, `comparison/selectdetail/...` | 0 |
| `branch/<branch>/comparison_update/` | 0 |
| `/layerindex/oe-classic/stats/` | 0 |
| `ajax/classchecklist/<branch>/` | 0 |

Use cases: comparing recipes with other distributions or OE-Classic, diffing two release branches (for example for release notes), finding duplicate recipes across layers, and preparing bulk metadata patches.

### 7. Accounts, admin and housekeeping

| Endpoint | Requests | Notes |
|---|---:|---|
| `/accounts/login/` | 339,927 | 99.9% carry `?next=` pointing at a recipe, layer-update or RRS page: crawlers following the "Log in" link |
| `/accounts/logout/` | 3,345 | 3,335 are POSTs rejected with 403 |
| `/admin/login/` | 1,286 | 76% bots |
| `/accounts/register/` | 151 | 37 POSTs; 9 activations completed |
| `/captcha/image/<key>/` | 279 | |
| `/accounts/password_reset/` | 59 | |
| `/layerindex/profile/`, `/accounts/password_change/` | 5, 5 | |
| Django admin model pages | about 100 | one or two staff users |
| `/accounts/delete/`, `/accounts/reregister/`, `/accounts/lockout/`, `captcha/audio`, `captcha/refresh` | 0 | |
| `/layerindex/task/<id>/`, `tasklog/<id>/` | 5, 10 | all but one are 4xx |
| `/layerindex/stoptask/<id>/`, `/layerindex/email_test/` | 0 | |

### 8. Redirect shortcuts (never appear in the database)

| Endpoint | Redirects |
|---|---:|
| `/layerindex/` | 181,762 |
| `/layerindex/layer/<slug>/` | 720 |
| `/layerindex/layers/` | 281 |
| `/layerindex/recipes/` | 195 |
| `/layerindex/branch/<branch>/` | 168 |
| `/layerindex/machines/` | 42 |
| `/layerindex/classes/` | 1 |
| `/layerindex/distros/` | 0 |

## Deprecation candidates

### Tier 1: remove (no meaningful use in 126 days)

| Endpoints | Evidence | Justification |
|---|---|---|
| Bulk Change: `bulkchange/` and its five sub-routes | 3 hits on the landing page, 0 on every sub-route | Nobody completed, or even started, a bulk change. It is the most complex logged-in feature and needed two security fixes this year (`40b2394`, `20e8409`). |
| Recipe Dependencies: `recipe_deps/` | 0 requests of any status | Linked from the Tools menu yet never opened. |
| Comparison selection: `comparison/select/<pk>/`, `comparison/selectdetail/...` | 0 | Manual recipe-matching workflow with no users. |
| `branch/<branch>/comparison_update/` plus `task/<id>/`, `tasklog/<id>/`, `stoptask/<id>/` | 0, 5, 10, 0 | The update button is off unless `COMPARISON_UPDATE` is set; the task pages exist only to serve it. Check that nothing else links to the task pages before removing them. |
| `ajax/layerchecklist/<branch>/`, `ajax/classchecklist/<branch>/`, `comparison/search-csv/<branch>/` | 2, 0, 2 | Helpers for the comparison search filters and its CSV export; unused. |
| OE-Classic aliases: `oe-classic/`, `oe-classic/recipes/`, `oe-classic/stats/`, `oe-classic/recipe/<pk>/` | 7, 68, 0, 67 | Pure aliases of the `comparison/` routes. Replace with permanent redirects, or drop. |
| Branch comparison exports: `branch_comparison_plain/`, `_rst/`, `_csv/` | 5, 7, 6 | 18 requests in total. |
| `email_test/` | 0 | Unused staff helper. Its permission check also looks wrong: `not is_authenticated and is_staff()` lets any logged-in user through. I read the line but did not exercise it. |

### Tier 2: discuss (low use, some residual value)

| Endpoints | Evidence | Consideration |
|---|---|---|
| `branch_comparison/` | 44 requests, 11 of them server errors | Barely used and failing a quarter of the time. Either fix and keep for release-notes work, or remove along with its exports. |
| `stats/` | 76 requests, 8 non-bot | Effectively unvisited. |
| `branch/<branch>/duplicates/` | 342 requests, 5 in the quiet window | Mostly crawlers. |
| `comparison/recipes/<branch>/`, `comparison/recipe/<pk>/`, `comparison/stats/<branch>/` | 284, 1,731, 57; 84, 80 and 15 in the quiet window | The whole distro-comparison feature sees about four human-scale visits a day. Removing it would also retire most of Tier 1. This is the largest possible simplification, and a product decision. |
| API: `layerNotes`, `incFiles`, `classes`, `appends`, `layerMaintainers` | 27 to 168 list requests each; detail routes 0 to 3 | No named tool depends on them. Cheap to keep, so only worth removing if the API surface is being trimmed anyway. |
| API detail routes for `layerDependencies`, `distros`, `machines` | 0, 0, 2 | The list routes are used by BitBake and must stay; the per-item routes are not used. |
| Un-branched shortcuts `classes/`, `distros/`, `machines/` | 1, 0, 42 redirects | Harmless; remove only for tidiness. |

### Also unused: API format suffixes

All 28 `.<format>` variants registered by the DRF router (for example `api/recipes.json`) received zero requests. Clients select JSON with `?format=json` (72,411 requests) or content negotiation. The suffix patterns can be switched off without affecting anyone seen in these logs.

### Keep despite low numbers

- Layer submit, review, publish, edit and notes: low volume is expected, and this is the site's core workflow.
- Account deletion, re-registration, lockout, password reset and the captcha audio and refresh routes: zero or near-zero use, but they are needed for account hygiene and accessibility.
- `layer/<slug>/` and the other busy redirect shortcuts: they catch inbound links.
- All nine API routes used by BitBake, plus `recipesExtended` and `layers`.
- The RRS pages.

## Other findings worth acting on

- **Publish is failing.** 16 of 31 requests to `layer/<name>/publish/` returned a server error. Commit `890f335` ("fix publishing layers with multiple branches") may already address it; I did not check which version is deployed.
- **Crawlers are the load.** About 60% of successful requests self-identify as bots, and the quiet window suggests the true share is far higher. The recipe detail page, the login page with `?next=`, the per-layer recipes CSV export (107,149 of 167,922 requests from a single `aiohttp` scraper) and `layerupdate/<pk>/` (44% 404) are the main targets. Marking the login link and the CSV export `nofollow`, and excluding them and `layerupdate` in `robots.txt`, would cut this cheaply.
- **Submission spam.** 9,685 POSTs to `submit/` and 3,335 to `logout/` were rejected with 403, consistent with bots posting without a CSRF token.
- **`urls.json` is behind the code.** `/sitemap.xml` (added in `66806a1`) is missing from it, so 198 sitemap requests are stored with an empty route.

## Reproducing the numbers

- 2xx figures: the `logs` table in the `layerindex` database (`route` column), reachable through Metabase at http://localhost:3000/.
- All-status figures: a one-off scan of the decompressed `layersweb_*_log.txt` files against `urls.json`. The script was a scratch file and is not part of this repository.
