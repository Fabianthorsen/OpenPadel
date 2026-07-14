# Page-redesign rubric

**Ticket:** [Lock the design north-star & page-redesign rubric](https://github.com/Fabianthorsen/OpenPadel/issues/169) (#169)
**Map:** [UI improvement: redesign every page](https://github.com/Fabianthorsen/OpenPadel/issues/167) (#167)

The shared checklist **every per-page spec and implementation applies.** Read this + the relevant
section of the [UI audit](../research/ui-audit.md) before speccing a page. North-star lives in
[`DESIGN.md`](../../DESIGN.md).

---

## North-star in one line

**Calm by default, bold when we celebrate.** Working screens (home, creation, lobby, active round,
score, auth, profile) = Nordic restraint. Celebratory surfaces (leaderboard, session complete) may be
expressive (hero cards, podium, motifs, display type). Expressive treatments are **celebratory-only**.

---

## ⚠️ The anti-divergence rule (read first)

Design-system debt is being fixed **per-page** (no prerequisite pass). That only works if fixes are
**shared, not copied**. Therefore:

> **When a page needs a component, variant, or token that doesn't exist yet, add it to the shared
> layer — never inline it and never duplicate it.** The first page that needs it builds it for all;
> every later page reuses it.

- New button style → add a variant to `web/src/lib/components/ui/button/` (e.g. a full-width `cta`
  size, a solid `destructive`). Do **not** write `<Button class="bg-primary … rounded-2xl …">`.
- New token → add to `web/src/app.css @theme` **and** `web/src/lib/design-tokens.ts` (keep them in
  sync). Do **not** use literal hex or raw Tailwind palette (`amber-500`, `#3d7a24`).
- Shared chrome (page shell, header, close button, spinner) → build once in
  `web/src/lib/components/` and reuse. Do **not** re-hand-roll the `pt-safe-page min-h-svh px-6` shell
  per page.
- Section label → use `SectionLabel`. Do **not** re-type `text-[11px] … tracking-[0.1em] uppercase`.

A page spec must **name** which shared additions it introduces, so reviewers can check they landed in
the shared layer.

---

## The four lenses (checklist)

### 1. Apply the design system
- [ ] No inline Tailwind where a Phase-1 primitive exists (`Button`, `Input`, `Label`, `Switch`,
      `Toggle`, `Badge`, `Drawer`, `Card`, `SectionLabel`, `Stepper`, `Separator`, `Tabs`).
- [ ] No `<Button class="…">` that overrides the primitive's variant/size — extend the primitive instead.
- [ ] Tokens over literals: no literal hex, no raw Tailwind palette colors. `text-primary-foreground`
      not `text-white`; `text-destructive` not `text-[#c0392b]`.
- [ ] Greens: only `--color-primary` on working screens; podium tokens on celebratory surfaces only.
- [ ] Radius/spacing via tokens (add `lg`/`xl` radius tokens if the page needs `rounded-2xl/3xl`).
- [ ] Any new shared component/variant/token follows the anti-divergence rule above.

### 2. Rethink UX flow
- [ ] Courtside-first: the primary action is obvious and reachable one-handed.
- [ ] "Score in 3 taps" preserved/improved where relevant.
- [ ] Consolidate duplicated flows the audit flagged (two create UIs, two join-code inputs, two
      schedule pickers, two lobby join forms, three profile confirm-modals) — don't reproduce them.
- [ ] Every state specified: **empty / loading / live / error** (loading must not be a blank screen or
      bare "Loading…" — use a shared spinner/skeleton).

### 3. Visual polish (per DESIGN.md)
- [ ] Correct register: working screen = calm; celebratory = expressive. No expressive treatment
      leaking onto a working screen.
- [ ] Typographic hierarchy from the scale; `tabular-nums` on all scores/points.
- [ ] **No emoji** (replace `🎾`/`⏱`/`ℹ` with `@lucide/svelte`); decorative motifs celebratory-only.
- [ ] Shadows subtle and reserved; no gradients.

### 4. A11y & responsiveness
- [ ] Min 48×48px tap targets.
- [ ] `focus-visible` styles + `aria-label`/labels on every interactive element; real `<label for>`
      on inputs (not a styled `<p>`).
- [ ] Safe-area insets honored (`pt-safe`/`pb-safe`).
- [ ] Live regions announce updates (players joining, leaderboard changes) for screen readers.
- [ ] Single column, max-width 480px; renders correctly in **light mode on a dark device** (i.e. the
      `dark:` leakage is stripped on this page).
- [ ] Copy goes through i18n (`$_()`), no hardcoded English strings.

---

## Standing fixes to apply as pages are touched

These are the systemic issues from the audit. Whichever page hits one first fixes it **in the shared
layer**; later pages just consume it. Track them here as they land:

| Fix | Where it lands | Status |
|-----|----------------|--------|
| Full-width `cta` button size + solid `destructive` variant | `ui/button/` | ☐ pending |
| Add `--radius-lg` (16px) / `--radius-xl` (24px) tokens | `app.css` + `design-tokens.ts` | ☐ pending |
| Type-scale tokens (Display/H1/H2/H3/Body/Small/Micro) | `app.css` + `design-tokens.ts` | ☐ pending |
| `SectionLabel` used everywhere (add `size`/`muted` variant) | `ui/section-label/` | ☐ pending |
| Shared `<PageShell>` / `<AppHeader>` / `<CloseButton>` / `<Spinner>` | `lib/components/` | ☐ pending |
| Semantic `--color-warning` (+ "live" decision) tokens | `app.css` + `design-tokens.ts` | ☐ pending |
| Podium tokens `--color-podium-silver/bronze` | `app.css` + `design-tokens.ts` | ☐ pending |
| Fold `#3d7a24` → primary; `app.html` theme-color → `#2d5a1a` | app code + `app.html` | ☐ pending |
| Strip `dark:` leakage | vendored `ui/*` + composites | ☐ pending |
| Consolidate lucide (`Avatar` → `@lucide/svelte`) | `ui/Avatar.svelte` | ☐ pending |

---

## Definition of done (per page)

A page is done when: it passes the four-lens checklist above; every shared addition it introduced
lives in the shared layer (not inline); `make lint && make test` and `bun run check` (in `web/`) are
green; and the change is `/verify`-ed in the running app (not just tests).
