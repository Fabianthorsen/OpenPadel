# Auth (login · forgot · reset) — redesign spec

**Ticket:** [Redesign spec: Auth (login, forgot, reset)](https://github.com/Fabianthorsen/OpenPadel/issues/177) (#177)
**Map:** [UI improvement](https://github.com/Fabianthorsen/OpenPadel/issues/167) (#167) · **Rubric:** [redesign-rubric.md](redesign-rubric.md) · **Audit:** [ui-audit.md](../research/ui-audit.md)
**Surfaces:** `web/src/routes/auth/+page.svelte`, `web/src/routes/forgot/+page.svelte`, `web/src/routes/reset/+page.svelte`

> Register: **working screens → calm**, utilitarian. Accounts are optional (guests are first-class) — auth is a **side path**: Home routes here for sign-in, and a "back home" link is always present.

---

## The decisions

- **Baseline (all included):** a shared **`AuthShell`** (brand header + centred card + back-home link — kills the 4× copy-paste); **real `<label for>`** on every field (they're styled `<p>`s today — an a11y gap); the **`Input` primitive** (drop the `border-0` override); **`Button` `cta`**; **consistent `translateApiError`** (fixes `reset`'s raw `e.message`) with `aria-invalid` + inline error text.
- **+ Password show/hide toggle** (a shared **`PasswordInput`**) on login / register / reset.
- Keep the three routes, the login↔register toggle on `/auth`, forgot's sent-confirmation card, and reset's invalid-link + `≥8` gate.

---

## Shared components (anti-divergence rule)

- **`AuthShell`** — `OpenPadel` wordmark + a subtitle slot + centred `max-w-sm` column (`pt-safe-page`) + a "← back home" link. Used by all three; mirrors Home's brand block.
- **`PasswordInput`** — `Input type="password"` + a show/hide eye toggle (`aria-pressed`, swaps `type`). Reusable.

---

## Screens

### `/auth` — login ↔ register
`AuthShell` (subtitle per mode). Register adds first/last name (2-up). All: email, **`PasswordInput`**. `Button` `cta` ("Sign in" / "Create account"). "Forgot password?" link (login only). Mode toggle via `?register=1` and a "No account? Create one / Have an account? Sign in" switch. Redirect logged-in → `redirect || /profile`.

### `/forgot`
`AuthShell`. Email field + `Button` `cta` (send). On submit → the **sent-confirmation card** ("check your email"); **never reveal whether the email exists** (keep). Back to `/auth`.

### `/reset`
`AuthShell`. No token → invalid-link message. Else: new-password (`PasswordInput`, `≥8`) + `Button` `cta`. **Error via `translateApiError`** (fix the raw `e.message`). On success → `/auth?reset=1` (success toast).

---

## Components & tokens

- Reuse `Input`, `Button` (`cta`), `Label`; add shared `AuthShell` + `PasswordInput`.
- Kill: the 4× brand copy-paste, the `<p>` pseudo-labels, the `Input` `border-0` override, the raw `bg-primary … text-white` CTA, and `reset`'s raw `e.message`.

---

## Interaction & flow

- Submit on Enter; disabled while loading; buttons show a loading state.
- login → `auth.login`; register → `auth.register`; forgot → `api.auth.forgotPassword` (always success UI); reset → `api.auth.resetPassword` → `/auth?reset=1`.
- Keep `autocomplete` (`email`, `current-password`/`new-password`, `given-name`/`family-name`) and `type="email"`.

---

## States

- **Idle** · **Submitting** (button loading, inputs locked) · **Error** (toast + `aria-invalid` + inline text).
- Forgot: **sent** (confirmation card). Reset: **invalid-link**, **success** (redirect + toast).

---

## A11y & responsiveness

- Real `<label for>` on every field; `PasswordInput` toggle labelled ("Show/Hide password", `aria-pressed`).
- `aria-invalid` + associated error text; `autocomplete`/`inputmode`; tap targets ≥48px; `focus-visible`.
- Safe-area; single column ≤480px; all copy via `$_()`.

---

## Before → after

| Before (shipped) | After (this spec) |
|---|---|
| Brand header copy-pasted across auth/forgot/reset (+ home) | Shared `AuthShell` |
| Field labels as styled `<p>` (not associated) | Real `<label for>` (a11y) |
| `Input` `border-0` override; raw CTA `<button>` | `Input` primitive; `Button` `cta` |
| `reset` shows raw `e.message` | `translateApiError` (consistent) |
| Plain password fields | `PasswordInput` with show/hide |

---

## Out of scope

- New auth methods (magic link, OAuth, "keep me signed in") — backend / new capability.
- Inline as-you-type validation (chose the password toggle only; on-submit + disabled gates stay).
- The Home entry / guest join (#176); the profile/account page (#178).
