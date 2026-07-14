# UI Audit — design-system inventory & DESIGN.md staleness

**Ticket:** [Audit: design-system inventory & DESIGN.md staleness](https://github.com/Fabianthorsen/OpenPadel/issues/168) (#168)
**Map:** [UI improvement: redesign every page with the Phase-1 design system](https://github.com/Fabianthorsen/OpenPadel/issues/167) (#167)
**Date:** 2026-07-14 · **Repo state:** `main` @ `09bf2a8`

> Groundwork asset — **no decisions here**, just what exists and where it is inconsistent. The
> "Open questions" at the end feed the [Lock the design north-star & page-redesign rubric](https://github.com/Fabianthorsen/OpenPadel/issues/169) (#169) ticket. Every per-page spec ticket should skim the relevant page section here first.

---

## TL;DR — the five findings that shape everything downstream

1. **DESIGN.md is materially stale.** Every neutral hex in its Color System changed, the accent shifted from `#4A7856` forest green to a darker/more-saturated `#2d5a1a`, the "shadcn-svelte" component table describes a stack we don't use, and 5 of the 7 "Key Screens" mockups no longer match the code (Leaderboard/Active-round/Score-entry diverged the most).
2. **The app drifted from "Nordic restraint" to a bolder, graphic direction.** Shipped UI has saturated full-bleed green cards, hero/podium treatments, decorative SVG court-line backgrounds, 800-weight display type, and emojis (`🎾`, `⏱`, `ℹ`) — all things DESIGN.md explicitly forbids. **This is the core north-star question for #169.**
3. **Only 7 of 19 primitives are on the Phase-1 pattern** (tv + JSDoc): Button, Input, Label, Switch, Toggle, Badge, Drawer. The other 12 hardcode inline Tailwind and lack JSDoc.
4. **The biggest systemic debt is a missing full-width CTA button variant.** The exact string `bg-primary … rounded-2xl px-4 py-4 text-[15px] font-semibold text-white` appears 20+ times — often as `<Button class="…">` *overriding the primitive's own variant*, because the `Button` default (`h-8 px-2.5 rounded-lg`) matches nothing the app actually uses.
5. **Token scale is out of sync with real usage** — radius tops out at `0.75rem` but every card uses `rounded-2xl`/`rounded-3xl`; there is **no typography scale token at all**, so `text-[28px]`/`font-[800]` are scattered everywhere; and semantic states (warning/live) use raw Tailwind palette (`amber-*`, `red-50`, `emerald-*`) with no tokens.

---

## Part A — Design-system primitive inventory

`web/src/lib/components/ui/*`. Pattern = `tv` (tailwind-variants), `inline` (hardcoded class strings), or `mixed`.

| Component | Pattern | Variants | Sizes | JSDoc | Base | Notes |
|-----------|---------|----------|-------|-------|------|-------|
| **Button** | tv | default, outline, secondary, ghost, destructive, link | default, xs, sm, lg, icon, icon-xs/sm/lg | ✅ | hand-rolled | Reference impl. **But** default size fits nothing in-app → always overridden. No full-width/CTA size, no solid-destructive. |
| **Badge** | tv | default, secondary, destructive, outline, ghost, link | fixed `h-5` | ✅ | hand-rolled | Uses `rounded-4xl` (not a radius token). **Never imported by any page/composite.** |
| **Input** | tv (base only) | — | — | ✅ | native | `tv()` but **no variants/sizes defined**. Always overridden to `bg-surface-raised rounded-2xl border-0`. |
| **Label** | tv (base only) | — | — | ✅ | bits-ui | base only. Rarely used; pages hand-roll labels instead. |
| **Switch** | tv | — | default, sm | ✅ | bits-ui | Two tv blocks. Reimplemented from scratch in `profile` anyway. |
| **Toggle** | tv | default, outline | default, sm, lg | ✅ | bits-ui | On-pattern. Unused by composites. |
| **Drawer** | tv | — | sm, md, lg | ✅ | bits-ui Dialog | Phase-1 "new custom" component + `types.ts`. Bottom slide-up. Good. |
| **Avatar** | inline | 10 color map | sm, md, lg, xl | ❌ | hand-rolled | Hardcodes raw hex (`bg-[#2d5a1a]`…), bypasses tokens. Imports **legacy `lucide-svelte`** not `@lucide/svelte`. |
| **Calendar** | inline | — | — | ❌ | bits-ui | Long inline strings; loose `[key:string]:unknown` props. |
| **Card** (+6 subs) | inline | — | default, sm | ❌ | hand-rolled | Single `cn()` string. **Never imported** — yet the `bg-surface-raised rounded-2xl px-4 py-3` row is hand-rolled 30+ times. |
| **Collapsible** | passthrough | — | — | ❌ | bits-ui | Thin re-export. (Profile uses it well.) |
| **Dialog** | inline | — | `sm:max-w-sm` | ❌ | bits-ui | Centered modal. No JSDoc. |
| **Sheet** | inline | side: top/right/bottom/left | fixed | ❌ | bits-ui | Largest inline string in the folder. |
| **PillToggleGroup / Item** | inline | — | — | ❌ | bits-ui | `any` props, no `ref`/`data-slot`, no JSDoc. (Separate from `toggle-group`.) |
| **SectionLabel** | inline | — | — | ❌ | wraps DS Label | Uppercase caption. **Reimplemented inline ~25× across pages** instead of imported. |
| **Separator** | inline | orientation | — | ❌ | bits-ui | Single `cn()` string. |
| **Stepper** | inline | — | — | ❌ | hand-rolled | `+`/`−` numeric. **No `index.ts`**, PascalCase filename (odd one out). |
| **Tabs** (+ list) | mixed | list: default, line | — | ❌ | bits-ui | Root inline, only `tabs-list` uses tv. Dead-looking `cn-tabs-list-variant-*` classes. **Unused** (ActiveSession hand-rolls its tab bar). |
| **ToggleGroup** (+ item) | mixed | inherits toggle variant/size via context; `spacing`, `orientation` | default, sm, lg | ❌ | bits-ui | Root inline; item borrows `toggleVariants`. |

**Migration debt (NOT on Phase-1 pattern):** Avatar, Calendar, Card, Collapsible, Dialog, Sheet, PillToggleGroup, SectionLabel, Separator, Stepper, Tabs, ToggleGroup (12 of 19). Input/Label are nominally on-pattern but expose no real variants.

**Tooling inconsistency:** two lucide packages installed — `lucide-svelte@^1.0.1` (legacy, used by Avatar) and `@lucide/svelte@^1.8.0` (used by dialog/sheet). Consolidate.

---

## Part A2 — Design tokens

Source of truth: `web/src/app.css` `@theme`, mirrored **1:1 and faithfully** in `web/src/lib/design-tokens.ts` (`as const` + exported key types).

**Colors (actual):**
- Surfaces: `background #edeee8`, `surface #f4f4f0`, `surface-raised #e3e3dc`, `border #d2d2cb`, `border-strong #b2b2aa`
- Primary: `primary #2d5a1a`, `primary-hover #234d13`, `primary-muted #e2ede0`, `primary-foreground #fff`
- Text: `text-primary #1a1a16`, `text-secondary #68685e`, `text-disabled #aeaea4`
- Semantic: `positive #2d5a1a`, `destructive #c0392b`, `destructive-foreground #fff`
- **Undocumented shadcn-compat block:** `card`, `popover`, `secondary`, `muted`, `accent`, `input`, `ring`, `foreground` + `*-foreground` (app.css:33 comment "shadcn-svelte tokens").

**Spacing:** `0–4` = 0 / 4 / 8 / 12 / 16px. **Stops at 16px** (no 20/24/32/48).
**Radius:** `--radius 0.75rem` (12, "base"), `md 0.5rem` (8), `sm 0.25rem` (4). **No pill token**, no 6px.
**Shadows:** `sm 0 1px 2px/.05`, `md 0 4px 6px -1px/.1`, `lg 0 10px 15px -3px/.1`.
**Fonts:** `--font-sans: 'Inter', system-ui…` only. **No mono, no serif, NO font-size / type-scale tokens.**
**Animations:** `--animate-shake` (+keyframes), `--animate-ptr-fade` (+keyframes).

**Gaps vs DESIGN.md:** colors all stale (see Part C); spacing scale truncated (DESIGN.md declares to 48px); radius mismatch (DESIGN.md 8/6/4/99px vs real 12/8/4 + the app's actual `rounded-2xl/3xl`); no typography tokens despite DESIGN.md's full scale; shadow values don't match DESIGN.md's single permitted `0 1px 3px/.07`.

**CONTEXT.md `## Components & Design System` table:** accurate for the 7 Phase-1 components it lists (variants/sizes all correct), but documents only those 7 (omits the other 12 — by design per the incremental spec). **One factual error:** claims animations are "spin, pulse" — actual tokens are `shake` + `ptr-fade`.

---

## Part B — Composite components & per-page styling debt

### Composite components (`web/src/lib/components/*.svelte`)

- **`ActiveSession.svelte` (851 lines — heaviest offender):** almost entirely hand-rolled. **Off-palette green `#3d7a24`** on score cards & +/− buttons (L473/520/531/539/561/572) — not in the palette. Primary CTAs are raw `<button>` with the inline CTA string (Finalize L605, Next-round L660/672). **Bottom tab nav hand-rolled** (L781–814) — `Tabs` primitive unused. Close-"×" duplicated 3× (L291/705/719). Status pills hand-rolled (`Badge` unused). Amber "time expired" box uses raw `amber-500/600` + `dark:text-amber-400` (L328–331). Inline SVG court-line decoration duplicated (L475/540). Hardcoded English: "Active Scorekeeper", "Team A/B", "Official", "On court", "Bench", "Scoring", "Players".
- **`Lobby.svelte` (1047 lines):** mixes primitives with heavy hand-rolling. `Button`/`Input` used but **fully class-overridden** everywhere (L502/538/544/798/820/937/944/958). Raw `<input>` for name edit (L578). Selectable-card pattern duplicated (mode L637, schedule L720). `bg-red-50 text-red-900` raw warning ×2 (L495/925). **Inline section-label re-typed 7×** (L632/656/675/691/727/753/765). Two near-duplicate join forms (join-screen vs in-lobby).
- **`Leaderboard.svelte`:** no primitives beyond Avatar. **Raw hex podium colors** `#4a7856` (2nd), `#a8c5b0` (3rd), `text-[#c0392b]` (losses = the destructive token value as literal hex) at L121/146/171/201/338/340. Contact "Add/Added" pill hand-rolled ×2. Same court-line SVG as ActiveSession. Bespoke `grid-cols-[2rem_1fr_3rem_3.5rem_3rem]`. Hardcoded `W/D/L`, "Loading…", "Add".
- **`CreateDrawer.svelte`:** cleanest — uses `Drawer` + `PillToggleGroup`. But CTA still raw `<button>` (L77) and hand-rolled close-"×". **Overlaps** the home-page setup form (redundant create UX).
- **`RoundIndicator.svelte`:** clean, token-correct. Low debt.
- **`ConfirmDialog.svelte`:** wraps `Dialog` (good) but action buttons are raw `<button>` with the CTA/destructive strings (L35–49) instead of `Button variant="destructive"`. Default `cancelLabel = 'Cancel'` hardcoded English.
- **`Footer.svelte` / `LocaleSwitcher.svelte`:** token-correct, minor arbitrary values. Low debt. (LocaleSwitcher uses `text-border` as a text color for the "·" — odd.)
- **`PullToRefresh.svelte`:** **token bug** — L124 references `var(--primary)` / `var(--text-disabled)` which don't exist (real tokens are `--color-primary` / `--color-text-disabled`), so the indicator color silently never resolves. Also appears **not mounted anywhere** — confirm it's wired.

### Per-page styling debt (keyed to the 9 redesign pages)

- **Home / landing (`routes/+page.svelte`):** hosts BOTH a `home` state and a full `setup` create form → **three create UIs exist** (home setup vs `CreateDrawer` vs Lobby config). 6 raw CTAs (L184/403). Hand-rolled initials circle (L247/389 — `Avatar` does this). "or" divider `bg-border h-px flex-1` ×2. Scheduling via raw `<input type=range>` (L367) — but Lobby schedules with Steppers (inconsistent). Setup step *does* use SectionLabel/PillToggleGroup/Switch/Calendar/Input/Button correctly.
- **Session creation (`CreateDrawer.svelte`):** raw CTA; overlaps home setup.
- **Lobby & join (`Lobby.svelte`, `s/[id]/+page.svelte`):** see Lobby above. The `s/[id]` page is a **state dispatcher** (L115–131): `!session`→"Loading…", `lobby`→Lobby, `playing`→ActiveSession, `done`→Leaderboard. Loading states are bare centered `<p>` (no spinner) — inconsistent with the spinners inside Lobby/ActiveSession.
- **Active round (`ActiveSession.svelte`):** off-palette `#3d7a24`, raw CTAs, hand-rolled tabs, hardcoded English.
- **Score entry (score flow in ActiveSession + numpad):** +/− raw circular buttons (`text-[#3d7a24]`). **Numpad lives in `s/[id]/+page.svelte` L154–197** — 12 hand-rolled `<button>` keys in a `Drawer`; uses `animate-[shake_0.4s_ease-in-out]` arbitrary value (wrong easing) instead of the `animate-shake` token. Auto-complement fills the opponent score.
- **Live leaderboard (`Leaderboard.svelte`):** hero leader card + inline SVG; bespoke grid; raw hex podium; hardcoded `W/D/L`.
- **Session complete (`Leaderboard.svelte` complete branch):** podium with hex medal colors, trophy, add-contact pills; close link raw bordered `<a>`. No "Share results" button (DESIGN.md §7 claims one).
- **Auth (`auth`/`forgot`/`reset` `+page.svelte`):** three near-identical shells; **brand header block copy-pasted 4×** (incl. home). Field labels hand-rolled as `<p class="…uppercase">` (SectionLabel/Label bypassed). Inputs overridden to `border-0`. `reset` uses raw `e.message` not `translateApiError`.
- **Profile (`profile/+page.svelte`, 868 lines):** **three different modal patterns** — `ConfirmDialog`, `Dialog`, and a fully hand-rolled `fixed inset-0 bg-black/40` modal (L869). **10× inline section-labels.** **Hand-rolled push toggle** (L475 — reimplements `Switch`). Hand-rolled initials circles. Raw `emerald-500` "Live/playing" badge (different green from ActiveSession). Raw OTP join-code boxes (L435 — home uses a single `Input`). `text-[#c0392b]` losses. Stat/invite/history rows all raw `bg-surface-raised rounded-2xl` (Card unused). Good: uses `Collapsible` consistently.
- **Global chrome (`+layout.svelte`, Footer, LocaleSwitcher, PullToRefresh):** `+layout.svelte` is **bare** (i18n gate + `Toaster` only). No shared header/brand/nav → every page re-implements its own `pt-safe-page min-h-svh px-6` shell + brand wordmark + `×` back nav. Loading while `$isLoading` renders a **blank screen**.

---

## Part B2 — Cross-cutting systemic debt (the high-leverage fixes)

Ranked by how many pages they touch:

1. **Missing full-width CTA button variant.** The `bg-primary … rounded-2xl px-4 py-4 … text-white` string appears 20+ times. → add a `size="cta"` / full-width variant (and a solid `destructive` variant) to `button.svelte`; ban the inline string.
2. **Radius scale mismatch drives the overrides.** Tokens top at `0.75rem`; the app's real language is `rounded-2xl` (cards/CTAs) + `rounded-3xl` (score cards/modals). → reconcile the radius tokens to reality (this is *why* primitives get class-overridden).
3. **`SectionLabel` reimplemented inline ~25×** (`text-[11px] font-semibold tracking-[0.1em] uppercase` + a `text-[10px]` micro-variant). → standardize on `SectionLabel` with a `size`/`muted` variant.
4. **No typography scale tokens** → `text-[28px]`, `text-[80px]`, `font-[800]` everywhere. A redesign should add a type scale (DESIGN.md declares one; nothing implements it).
5. **`text-white` instead of `text-primary-foreground`** on every green button/badge (bypasses the token even though both = `#fff`).
6. **Literal hex / raw palette for semantics with (or needing) tokens:** `text-[#c0392b]` (=`destructive`, 3 files); `bg-red-50`/`amber-*`/`emerald-*` for warning/live/success states that have **no semantic token** (only `positive`/`destructive` exist); off-palette greens `#3d7a24`/`#4a7856`/`#a8c5b0`.
7. **Primitives that exist but are never imported:** `card` (+7 subs), `separator`, `tabs`, `badge`, `label`, `toggle` — these map directly onto the most-duplicated hand-rolled patterns (list rows, dividers, tab bars, status pills).
8. **No shared page shell / chrome** → duplicated `pt-safe-page` shell + brand header (4×) + `×` back nav; inconsistent loading states (blank / bare text / spinner). → a `<PageShell>` / `<AppHeader>` + shared `<Spinner>` / `<CloseButton>`.
9. **Hardcoded English despite full i18n** — leaks in ActiveSession, Leaderboard, Profile, Lobby. Relevant since redesign will revisit copy.
10. **Duplicated flows to consolidate:** two create UIs, two join-code inputs (single `Input` vs 4-box OTP), two schedule pickers (range slider vs Steppers), two Lobby join forms, three profile confirm-modals.

---

## Part C — DESIGN.md staleness report

| DESIGN.md claim | Reality | Verdict |
|---|---|---|
| Color System hex table (Background `#F7F6F3`, Surface `#FFFFFF`, all neutrals) | app.css: `#edeee8`, `#f4f4f0`, all neutrals changed; surface no longer pure white | **stale** |
| Primary `#4A7856` forest green (+ hover/muted) | `#2d5a1a` darker/more-saturated; `#4A7856` only survives hardcoded in Leaderboard + `app.html` `theme-color` | **stale** |
| Destructive `#C0392B` | `--color-destructive: #c0392b` | ✅ accurate |
| "Only one accent color" | ~4 greens ship (`#2d5a1a` token, `#3d7a24`, `#4a7856`, `#a8c5b0`) + raw `red`/`amber` | **stale** |
| "No shadows beyond `0 1px 3px/.07`" | `shadow-lg 0 10px 15px -3px/.1` used on bottom nav, drawer, winner avatar | **stale** |
| Radius 8/6/4/99px | tokens 12/8/4; app uses `rounded-2xl`(16)/`rounded-3xl`(24) pervasively | **stale** |
| Inter via `@fontsource/inter`, no Google Fonts | ✅ `@fontsource/inter ^5.2.8`, imported in app.css | ✅ accurate |
| Type scale (Display 48/H1 26/…, weights ≤700) | not followed — arbitrary `text-[80px]`/`[28px]`/`[11px]`; `font-[800]` ×41 | **stale** |
| Headings `-0.01em` / line-height `1.2` | app.css: `-0.02em` / `1.1` | partially stale |
| Component Notes → **shadcn-svelte** | no shadcn dep; components **vendored** under `ui/*` on **bits-ui ^2.16.3** + tailwind-variants | partially stale |
| Toast → **Sonner** | **svelte-sonner ^1.1.0** | partially stale |
| Score numpad via `inputmode="numeric"` native input | **custom numpad store** (`$lib/stores/numpad`) with auto-complement, not native | **stale** |
| §1 Home: "Start session" primary + "Join with a link →" text link | CTA is **"Sign in →"** (creation auth-gated); join is a **4-char code input**; also hosts inline setup | **stale** |
| §2/§3 Start "enabled at ≥ 5 players" | threshold is **`courts × 4`** (`Lobby.svelte:316`) | **stale** |
| §4 Active round: quiet stacked courts + "[Enter scores]" | **tabbed** bottom-nav, one court at a time via `ToggleGroup` "🎾 Court N", immersive green cards + `⏱` timer | **stale** |
| §5 Score entry: separate screen, "24/24" indicator | integrated Scoring tab; 80px score + circular +/−; auto-complement; no "24/24" | **stale** |
| §6 Leaderboard: "Rank, name, points. Nothing else… #1 gets primary color, no badge" | leader **hero card**, 5-col grid (#/Player/Games/W-L/Pts) w/ avatars, colored podium rows, "LEADER" pill; complete = full podium | **stale (most diverged)** |
| §7 Complete: quiet typographic + "[Share results]" | 3-person **podium** (trophy, bars, add-contact); **no Share-results button** | partially stale |
| Micro-interactions (row fade-in, border-flash-green on confirm, row re-sort animation) | not found; actual = `animate-pulse`/`animate-shake`/`active:scale`; confirm shows a **toast** | partially stale |
| "No emojis" | `🎾` court tabs, `⏱` timer, `ℹ` info note | **stale** |
| "No decorative icons beyond copy/share/chevron" | Trophy, Crown, Shield, Activity, ChartBar, Users, etc. + decorative SVG court patterns | **stale** |
| "No dark mode in V1" | no dark palette/toggle, **but** `dark:` utility classes ship throughout vendored UI + `ActiveSession:331` → activate via `prefers-color-scheme` in Tailwind v4 → inconsistent half-dark render on dark devices | partially stale |
| "No gradients" | none found | ✅ accurate |

**Context:** git log shows Skeleton UI was installed (`d803f58`) then **reverted** (`09bf2a8`) — the repo stayed on bits-ui + vendored-shadcn-style components. `app.html:9` `theme-color` meta still ships the *old* `#4A7856`.

---

## Open questions for the Rubric ticket (#169)

These are **not** auto-fixable — they are north-star decisions the [rubric ticket](https://github.com/Fabianthorsen/OpenPadel/issues/169) must settle before per-page specs proceed:

1. **North-star philosophy:** Is it still "Nordic restraint — muted, typographically led, colour used sparingly," or has the product intentionally moved to the bolder/graphic direction that shipped (saturated green cards, podiums, SVG court patterns, 800-weight type)? Every page spec depends on this answer.
2. **Accent + green fragmentation:** Was the shift to `#2d5a1a` deliberate? Unify the stray `#3d7a24` / `#4a7856` / `#a8c5b0` (and the `app.html` `#4A7856` theme-color) to one tokenized green, or document the podium tints as an intentional palette?
3. **Dark mode:** strip all `dark:` variants to honor "no dark mode in V1," or commit to a real V2 dark palette? (Today it's an inconsistent half-dark render on dark devices.)
4. **Emoji & icon policy:** keep "no emojis / functional-icons-only" (remove `🎾`/`⏱`/`ℹ`, trim lucide), or relax the rule to match what shipped?
5. **Score-entry model:** confirm the auto-complement numpad (enter one score, the other fills) is the intended north-star before re-documenting §5.
6. **Systemic-debt sequencing:** do the cross-cutting fixes in Part B2 (CTA button variant, radius reconciliation, type-scale tokens, SectionLabel standardization, page shell) get done as a small **prerequisite design-system pass**, or absorbed page-by-page? (This is the map's "extend the design system only where a page demands it" boundary — Part B2 shows the demand is systemic, not per-page.)
7. **Rewrite scope for DESIGN.md:** which of the stale claims above get rewritten now vs. redrawn as each page spec lands?

---

## File index (cited)

`web/src/app.css` · `web/src/app.html` · `web/src/lib/design-tokens.ts` · `web/package.json` · `CONTEXT.md` · `web/src/lib/components/ui/*` · `web/src/lib/components/{ActiveSession,Lobby,Leaderboard,CreateDrawer,Footer,RoundIndicator,ConfirmDialog,LocaleSwitcher,PullToRefresh}.svelte` · `web/src/routes/{+layout,+page,auth/+page,forgot/+page,reset/+page,profile/+page,s/[id]/+page}.svelte`
