# Session complete — redesign spec

**Ticket:** [Redesign spec: Session complete](https://github.com/Fabianthorsen/OpenPadel/issues/174) (#174)
**Map:** [UI improvement](https://github.com/Fabianthorsen/OpenPadel/issues/167) (#167) · **North-star:** [visual language](../../CONTEXT.md#visual-language) · **Rubric:** [redesign-rubric.md](redesign-rubric.md) · **Audit:** [ui-audit.md](../research/ui-audit.md)
**Surfaces:** the `complete` branch of `web/src/lib/components/Leaderboard.svelte` (rendered by `s/[id]` when `status === 'done'`).

> Register: **the celebratory surface.** This is the one screen where bold is sanctioned (the live leaderboard #172 handed its drama here). **Terminal** — the session is over, so there's no session nav / Standings peek. Shares the standings-row markup with the live leaderboard (#172); coordinate.

---

## The decisions (grilling)

1. **Full podium, with real medal colors.** Keep the podium (trophy, 2nd/1st/3rd, bars, ranked list), but recolour it **gold / silver / bronze** (1st gold, 2nd silver, 3rd bronze) instead of the shipped green tints. Add a **`--color-medal-gold` / `--color-medal-silver` / `--color-medal-bronze`** celebratory palette (this refines #169/#172's "podium ≈ green tint" — the podium is medals). Medal colours are **accents** (rings / bars / rank badges), not body text — ensure contrast.
2. **Add-contact becomes a badge on the avatar.** Today the inline "Add"/"✓" button sits *in the row*, so rows without it (your own row, guests) don't align — the list looks **staggered**. Move it to a **small circular badge on the avatar's bottom-right corner** (`+` → tap → `✓` when added). Every row keeps a uniform width. Shown only to a logged-in viewer, on other **registered** players (has `user_id`, not self, not already a contact).
3. **Finale actions: all four** — **Share results** (native share, text + link now; a generated podium image is a later enhancement), **New session** (plain link to the create flow — *not* a same-players "rematch", which is new backend capability = out of scope), **keep add-contacts** (now the avatar badge), **Close / home** (styled).

---

## Layout & hierarchy

Single column, `max-w-[480px]`, `pt-safe-page`:

1. **Heading** — `SectionLabel "Final"` + session name (H1/bold).
2. **Podium** — visual order 2nd (left) · 1st (centre, elevated) · 3rd (right):
   - Winner: `Trophy` + `xl` avatar with a **gold** ring; 2nd `silver` ring, 3rd `bronze` ring.
   - Rank badge under each avatar in its medal colour (`--color-medal-*`).
   - Name, points (winner's in a prominent weight), and W-D-L (losses via `text-destructive`, never `#c0392b`).
   - **Podium bars** in medal colours, heights by rank (`shadow` on the winner is fine — celebratory).
   - **Add-contact badge** on each avatar (see below).
3. **Rest of standings (4th+)** — uniform rows: rank · avatar (with badge) · name · W-D-L · points. Badge-on-avatar keeps every row the same width (fixes the stagger).
4. **Actions** — **Share results** + **New session** (secondary/`cta`) and a styled **Close** → home/profile.

---

## Add-contact badge (the fix)

- A small circular badge overlapping the avatar's bottom-right corner: `UserPlus` → tap → adds → `Check` (added state).
- Visibility: logged-in viewer **and** the target has a `user_id` (registered) **and** isn't the viewer **and** isn't already a contact. Guests (no `user_id`) and your own avatar: no badge.
- **Uniform rows** — because the affordance lives on the avatar, rows without it don't collapse/stagger.
- A11y: the badge needs an adequate hit area (pad the tap target to a comfortable size even though the glyph is small) and an `aria-label` ("Add {name} to contacts"); added state announced.
- **Shared (anti-divergence):** implement as an `Avatar` action/badge slot or an `AvatarWithContact` wrapper so the pattern is reusable, not hand-rolled per row.

---

## Components & tokens

- Reuse `Avatar` (+ contact badge), `Button` (`cta` for primary action, secondary/outline for Close), `SectionLabel`, `Trophy` (lucide).
- **New celebratory tokens:** `--color-medal-gold/silver/bronze` (+ optional `-muted` for bar fills) in `app.css` + `design-tokens.ts` — celebratory-surface only.
- Tokenise the flagged hex: losses → `text-destructive`; `text-primary-foreground` over `text-white`; drop `#4a7856`/`#a8c5b0` in favour of the medal tokens.

---

## Interaction & flow

- **Share results** — `navigator.share({ title, text, url })` with the session link; clipboard-copy fallback. (Image card = future.)
- **New session** — link to the create flow (in scope). *Rematch with the same players/config is out of scope* (new capability) — flagged, not built.
- **Close** — styled button → `/` or `/profile` (logged-in).
- **Add contact** — `api.contacts.add`; optimistic → badge flips to added.

---

## States

- **Loading** — shared `<Spinner>`.
- **Final (podium)** — the celebration.
- **Few players** — <3 finishers (e.g. edge single-court/small session): podium shows what's available (1–2 spots), rest empty gracefully.
- **Logged-out viewer** — no add-contact badges; share + close still available.

---

## A11y & responsiveness

- DOM order = rank order (1st first) even though the winner is visually centred; medal colour is **not** the sole rank signal (the rank number/badge carries it — colour-blind safe).
- Add-contact badge: labelled, adequate hit area.
- Share / New session / Close keyboard-reachable, `focus-visible`.
- Safe-area; single column ≤480px; all copy via `$_()` (kill hardcoded "Final", "Add", "Added", "Close").

---

## Before → after

| Before (shipped `complete`) | After (this spec) |
|---|---|
| Green podium tints (`#4a7856`/`#a8c5b0`) | **Gold / silver / bronze** medal palette (tokens) |
| Inline "Add"/"✓" button → **staggered rows** | Add-contact **badge on the avatar** → uniform rows |
| Bare `✕ Close` link only | **Share results** + **New session** + styled Close |
| `text-[#c0392b]`, `text-white` | `text-destructive`, `text-primary-foreground` |

---

## Out of scope

- **Live standings** (mid-session) → Live leaderboard (#172).
- **Rematch** (new session pre-filled with the same players/config) → new capability, not this effort.
- **Share as a generated image** → later enhancement (text + link now).
- Ranking/tiebreak logic; dark mode.
