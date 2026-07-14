# Live leaderboard — redesign spec

**Ticket:** [Redesign spec: Live leaderboard](https://github.com/Fabianthorsen/OpenPadel/issues/172) (#172)
**Map:** [UI improvement](https://github.com/Fabianthorsen/OpenPadel/issues/167) (#167) · **North-star:** [DESIGN.md](../../DESIGN.md) §6 · **Rubric:** [redesign-rubric.md](redesign-rubric.md) · **Audit:** [ui-audit.md](../research/ui-audit.md)
**Surfaces:** `web/src/lib/components/Leaderboard.svelte` (the `!complete` branch), rendered standalone and inside the round-first **Standings peek** (#170).

> **Scope:** this spec covers the **live standings** view only. The **final podium / celebration** (the `complete` branch — trophy, podium bars, add-contact, close) belongs to **Session complete (#174)**. `Leaderboard.svelte` stays one component branching on `complete`; the two specs coordinate on the shared standings-row markup.

---

## The decisions (grilling)

Live standings are **calm & glanceable**, not celebratory (the celebration is the finale, #174 — north-star updated to say so):

1. **Density: rank · name · points (lean).** No Games column, no W-D-L in the default row (DESIGN.md §6 "read the standings in under 2 seconds"). W-D-L/games detail is a nice-to-have on tap (optional), not a default column.
2. **Emphasis: subtle.** **No leader hero card, no colored top-3 rows, no court-line motif.** Rank 1 gets `--color-primary` on its rank number (and name) — the single subtle distinction. Everyone else neutral.

---

## Layout & hierarchy

Single column, `max-w-[480px]`, `pt-safe-page`:

1. **Header** — `SectionLabel` "Standings" (or session name) + a round/status line: `"Round {n} of {total} · Live"` (`text-text-secondary text-xs`); "Live" carries a small pulsing `primary` dot.
2. **Standings list** — rows, top → bottom by rank:
   - **Rank** — `tabular-nums`; rank 1 in `text-primary font-bold`, others `text-text-disabled`.
   - **Avatar** (`sm`) + **name** (`shortName`, truncate). Rank 1's name may also be `text-primary`.
   - **Points** — right-aligned, `text-base font-[800] tabular-nums`; the ranking metric, most prominent number in the row.
   - Rows are calm: `bg-surface`/subtle alternating or hairline dividers — **no** podium-color fills.
   - Optional: a row is tappable to reveal W-D-L / games detail (progressive disclosure); not required for v1.
3. No podium, no add-contact, no close button here — those are the finale (#174).

### Standings peek (from the round-first IA, #170)

The same lean list rendered in a **bottom sheet** opened by the round view's floating Standings pill: header shows `"Round n · Live"`, the standings list, swipe-down/backdrop/Esc to dismiss back to the round. Compact — the calm/no-hero treatment makes it fit a sheet cleanly.

---

## Components & tokens

- Use `SectionLabel`, `Avatar`; rows via `Card` or a simple tokenised row (no hand-rolled `bg-surface-raised rounded-2xl` string).
- **Tokenise the ad-hoc hex the audit flagged** — the live view drops the podium greens entirely (no colored rows); losses (only shown if W-D-L detail is surfaced) use `text-destructive`, never `text-[#c0392b]`. `text-primary-foreground` over `text-white`.
- No new shared component required beyond what #170/#171 introduced; reuse the standings row for the peek and (shared markup) the #174 finale.

---

## Interaction & flow

- **Live refresh** — reload on the `round_updated` SSE event via `stream.onEvent` (keep; no in-component polling — the page-level 30s fallback covers buffering proxies, per ARCHITECTURE.md / CLAUDE.md).
- **Re-sort** — when standings change, rows animate to new positions (optional 150ms ease-out; nice-to-have, not required).
- **Ties** — display the rank the backend assigns; if two players share points, show them at the same rank number (or a subtle `T`), rather than arbitrary insertion order. (Resolves an open question from the old spec.)
- Read-only for everyone (no admin controls here).

---

## States

- **Loading** — shared `<Spinner>` (not bare "Loading…").
- **Live** — standings with the "Live" indicator; updates via SSE.
- **Empty / round 0** — standings before any games: rank by seed/zero points, calm.
- **Standalone vs peek** — same list; peek is the sheet-hosted compact form.

---

## A11y & responsiveness

- Rank-1 distinction is not colour-only — it's also weight (colour-blind safe); the "Live" state has a text label, not just the dot.
- Standings updates announced via `aria-live="polite"`.
- Tap targets ≥48px if rows are tappable (detail disclosure); `focus-visible`.
- Safe-area; single column ≤480px; all copy via `$_()`.

---

## Before → after

| Before (shipped live view) | After (this spec) |
|---|---|
| Bold leader **hero card** (`bg-primary` + court-line SVG) | **Removed** — no hero (drama moves to the finale #174) |
| Full-color green/silver/bronze **top-3 rows** (`#4a7856`/`#a8c5b0`) | **Removed** — calm neutral rows; rank 1 subtle `primary` |
| 5-column table (# / Player / Games / W-D-L / Pts) | **Lean** rank · name · points (W-D-L optional on tap) |
| `text-[#c0392b]`, `text-white`, hex greens | tokens (`text-destructive`, `text-primary-foreground`, no hex) |

---

## Out of scope

- **Final podium / celebration** (trophy, podium bars, add-contact, close) → Session complete (#174).
- Ranking/tiebreak *logic* (backend); new modes; dark mode.
