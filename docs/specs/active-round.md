# Active round view — redesign spec

**Ticket:** [Redesign spec: Active round view](https://github.com/Fabianthorsen/OpenPadel/issues/170) (#170)
**Map:** [UI improvement](https://github.com/Fabianthorsen/OpenPadel/issues/167) (#167) · **North-star:** [DESIGN.md](../../DESIGN.md) §4 · **Rubric:** [redesign-rubric.md](redesign-rubric.md) · **Audit:** [ui-audit.md](../research/ui-audit.md)
**Surfaces:** `web/src/lib/components/ActiveSession.svelte` (scoring/round portion), `RoundIndicator.svelte`, `web/src/routes/s/[id]/+page.svelte`
**Prototype:** `web/src/routes/prototype/active-round/` (throwaway; variants A/B/C — see bottom)

> Register: **working screen → calm** (per the north-star). Score-*entry* interaction (+/− cards, numpad, auto-complement) belongs to **Score entry (#171)** — this spec covers the round/court *display* and the affordance into scoring, not the entry UI. Leaderboard *content* belongs to **Live leaderboard (#172)** — this spec only defines how you reach it (the Standings peek).

---

## The decisions (from prototype + grilling)

1. **Layout is court-count adaptive** (same session never flips — court count is fixed at creation):
   - **`courts === 1`** → **focused single-court** hero card (prototype Variant B, minus tabs).
   - **`courts > 1`** → **glanceable court list**, all courts as calm cards at once (prototype Variant A).
2. **Register: calm + subtle live tint.** Neutral surface cards; the **live/in-progress court** gets a subtle `bg-primary-muted` fill or a `primary` left-border so it's spottable across the court. **Removed:** the immersive `#3d7a24` cards, the SVG court-line pattern, the `🎾` emoji.
3. **Session IA: round-first (shared shell decision).** No 3-tab bottom nav. The round is the whole screen; a floating **Standings** pill (top-right, sticky) opens the live leaderboard as a dismissible **bottom-sheet peek**. **Players is absorbed** into the round view (on-court shown in the court cards, bench shown as a row) — no separate Players destination. → **This IA is inherited by the other session-screen specs: Live leaderboard (#172), Lobby & join (#173), Session complete (#174).**

---

## Layout & hierarchy

### Shared frame (both layouts)

Top → bottom, single column, `max-w-[480px]`, `pt-safe-page`, `px-4`:

1. **Top bar** — session name (`text-primary text-sm font-semibold`, left) · **Standings pill** (right, sticky) · `×` back (far right, or fold `×` into an end-session affordance for admin).
   - **Standings pill:** compact pill, `bg-surface-raised text-text-primary`, Trophy/ChartBar icon + "Standings". `aria-haspopup="dialog"`. Tapping opens the leaderboard bottom sheet (peek). Never collides with the bottom CTA.
2. **Round header** — `SectionLabel`: `"{mode} · Round {n} of {total}"` (or `"Round {n}"` for open/unlimited). Below it a row: `RoundIndicator` (left) + timer chip (right, non-americano only): Clock icon + `mm:ss`, `text-text-secondary font-mono text-xs`. On expiry, a calm inline notice using the new `--color-warning` token (not raw `amber-*`).
3. **Court(s)** — the adaptive part (below).
4. **Bench** — `SectionLabel "Bench"` + quiet row (`bg-surface-raised rounded-xl`): avatar(s) + names. Hidden if empty.
5. **Round action zone** (bottom of content, admin only):
   - all courts scored → **Next round →** (full-width `cta` Button); final round → **Final results**.
   - some but not all scored → disabled Next round + helper "N courts still need a score".
   - **End session** → a quiet destructive text affordance opening the end-session bottom sheet (Save / Discard / Keep going).

### Layout — `courts === 1` (focused single-court)

One centered hero card (`bg-surface border-border rounded-3xl p-6`):
- Status chip centered (Live / Final / Not started).
- **Team A**: stacked overlapping avatars → team name (`text-[15px] font-semibold`) → big score (`text-5xl font-[800] tabular-nums`).
- Thin `border-border` divider.
- **Team B**: score → name → avatars (mirrored).
- Admin: full-width **Enter scores** / **Edit score** Button below the card → opens Score entry (#171).

### Layout — `courts > 1` (glanceable court list)

Vertical `space-y-3` list; one calm card per court (`bg-surface border-border rounded-2xl shadow-sm`):
- Header row: `"Court {n}"` micro-label (left) + status chip (right).
- **Team A row**: overlapping avatars + name (flex-1, truncate) + score (`text-2xl font-[800] tabular-nums`, right).
- Hairline divider (`border-border mx-4`).
- **Team B row**: same.
- Admin: a full-width bottom strip button on the card ("Enter score →" / "Edit score") → Score entry (#171). (Or whole-card tap; button is clearer for a11y.)

---

## Court card states (both layouts)

| State | Card | Score | Emphasis |
|-------|------|-------|----------|
| **Upcoming** (no score, no live) | neutral surface | `–` in `text-text-disabled` | none |
| **Live** (`match.live`, no final) | **subtle `bg-primary-muted` tint / `primary` left-border** | live score, `text-text-primary` | pulsing `primary` dot + "Live" chip |
| **Final** (`match.score`) | neutral surface | final score | winning row gets a subtle `bg-primary-muted` fill; winning score in `text-primary font-bold`; loser `text-text-disabled`; draw = both neutral |

Winner is emphasized by weight + `primary` colour + a muted-tint row — **not** a full-bleed green fill.

---

## Components & tokens

Use, don't hand-roll: `Avatar`, `RoundIndicator`, `SectionLabel`, `Card` (for the court cards), `Button`, `Badge` (status chips), `Sheet`/`Drawer` (Standings peek + end-session), `Toaster` (confirm feedback).

**Shared additions this page introduces (per the rubric's anti-divergence rule — land them in the shared layer):**
- `Button` **`cta` size** (full-width, `rounded-2xl px-4 py-4 text-[15px] font-bold`) — replaces the inline `bg-primary … text-white` CTA string used ~20× app-wide. **This is the first page to need it; build it in `ui/button/`.**
- `--radius-xl` (24px / `rounded-3xl`) + `--radius-lg` (16px / `rounded-2xl`) tokens.
- `--color-warning` semantic token (replaces raw `amber-*` for the time-expired notice) + the "live" signal via `--color-primary` (no `emerald-*`).
- Fold the score-card `#3d7a24` → `--color-primary`; `text-white` on primary → `text-primary-foreground`.

No literal hex, no raw Tailwind palette colours, no inline `SectionLabel` re-typing.

---

## Interaction & flow

- **Enter/edit score** (admin) → opens Score entry (#171) for that court. Non-admins see scores **read-only** (live + final).
- **Standings peek** → top-right pill opens the live leaderboard (#172 content) in a bottom sheet; swipe-down / backdrop / Esc dismisses, returning to the round. Available to everyone.
- **Advance round** (admin, all scored) → `POST /rounds/advance`; optimistic; button stays visible to retry on error.
- **Final round** → "Final results" transitions to Session complete (#174).
- **End session** (admin) → bottom sheet: Save & close / Discard / Keep going (reuse `ConfirmDialog`/`Sheet` pattern, `Button` variants — no raw buttons).
- **Live updates** — SSE-first with 30s poll fallback (unchanged; do not remove — see CLAUDE.md). Scores/rounds update in place; announce via a live region.

---

## States to implement

- **Waiting / not started** — courts render "Not started", scores `–`.
- **Live** — at least one court `live`; live tint + pulsing dot; scores update via SSE.
- **Between rounds** — all courts final; admin sees Next round; players see "Round complete" calm notice.
- **Final round complete** — "Final results" CTA → Complete.
- **Loading** — shared `<Spinner>` (not a blank screen / bare "Loading…"; fixes the `s/[id]` dispatch gap the audit flagged).
- **Error** (advance/close fail) — toast; controls stay actionable to retry.
- **Cancelling** — spinner + status (existing behaviour, restyled to tokens).

---

## A11y & responsiveness

- Court card score-entry button, Standings pill, nav, CTAs: min 48×48px; `focus-visible` rings; `aria-label`s.
- Status chips: don't rely on colour alone — the "Live"/"Final" text label carries the meaning (colour-blind safe).
- Live score/round changes announced via an `aria-live="polite"` region.
- Standings sheet: focus-trapped, Esc to close, returns focus to the pill.
- Safe-area insets top & bottom; single column ≤480px.
- All copy via `$_()` — kill the hardcoded English ("Active Scorekeeper", "Team A/B", "Official", "On court", "Bench", "Scoring", "Players").

---

## Before → after

| Before (shipped) | After (this spec) |
|---|---|
| One court at a time via `🎾 Court N` tabs, always | Court-count adaptive: 1 court = focused card, >1 = glanceable list of all courts |
| Immersive `#3d7a24` cards + SVG court pattern | Calm surface cards + subtle `primary-muted` tint on the live court only |
| 3-tab bottom nav (Round/Standings/Players) | Round-first: floating Standings pill (bottom-sheet peek); Players absorbed into the round view |
| Raw `<button class="bg-primary…text-white">` CTAs (×several) | `Button` `cta` variant (shared) |
| Raw `amber-*` expiry box, `text-white`, `#3d7a24` | `--color-warning`, `text-primary-foreground`, `--color-primary` tokens |
| Hardcoded English labels | i18n throughout |

---

## Out of scope for this spec

- **Score-entry interaction** (+/− cards, numpad, auto-complement) → Score entry (#171).
- **Leaderboard content** (columns, podium, medals) → Live leaderboard (#172); here we only host it in the peek sheet.
- Building dark mode; new game modes.

## Prototype reference

Throwaway route `web/src/routes/prototype/active-round/` — Variants **A** (glanceable list, won for >1 court), **B** (focused one-court, won for 1 court), **C** (dense scoreboard, not chosen). Flip via `?variant=` + floating bar; the placeholder 3-tab nav there is superseded by the round-first IA above. To be captured to a throwaway branch and dropped from main once the implement ticket folds the winners in.
