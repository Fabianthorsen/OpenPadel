---
name: ui-redesign-initiative
description: How the per-page UI redesign is structured — umbrella issue #167, per-page specs, shared rubric, anti-divergence rule, and the already-built shared layer
metadata:
  type: project
---

OpenPadel is running a per-page UI redesign (umbrella issue #167 "UI improvement"). Each screen has its own spec under `docs/specs/*.md` and a redesign ticket. Known sibling tickets share one IA: Active round #170/#179, Score entry #171, Live leaderboard #172, Lobby & join #173, Session complete #174.

Governing docs (read before planning any UI ticket):
- `docs/specs/redesign-rubric.md` — the four-lens checklist (design system / UX flow / visual polish / a11y) + Definition of Done every page must pass.
- `DESIGN.md` §4 — north-star: "calm by default, bold when we celebrate." Working screens = Nordic restraint; only leaderboard + session-complete may be expressive.
- `docs/research/ui-audit.md` — systemic issues catalog.

**Why:** Design-system debt is fixed per-page with no prerequisite pass. That only works via the **anti-divergence rule**: when a page needs a component/variant/token that doesn't exist, add it to the SHARED layer (`ui/button/`, `app.css @theme` + `design-tokens.ts`), never inline or duplicate. First page builds it; later pages consume.

**How to apply:**
- The shared design-system foundation already landed (commit b8179ab, PRs #189/#190). As of that work the shared layer ALREADY contains: `Button` `cta` size, `--radius-lg`(1rem)/`--radius-xl`(1.5rem), `--color-warning`/`--color-warning-muted`, `Spinner`, `Badge`, `Section`/`SectionLabel`. Verify with grep before assuming a token/variant must be built — most per-page tickets are now CONSUMPTION tasks (rewrite the page to use primitives + drop raw hex/tabs), not build tasks.
- The rubric's "Standing fixes" table Status column is STALE (shows ☐ pending for fixes already implemented in code). Trust the code, not that table; updating the table is a nice-to-have doc chore when a ticket confirms a fix.
- Every page spec names which shared additions it introduces so reviewers check they landed shared, not inline. Honor that in plans.
