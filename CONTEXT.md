# OpenPadel

A mobile-first PWA for organizing padel games and tracking scores courtside.

## Language

**Session**:
A single padel event with a lifecycle (`lobby` → `playing` → `done`), a game mode, and a set of players.
_Avoid_: Tournament, game, event

**Player**:
A participant in a Session. May be a **Guest** (name only, no account) or linked to a registered **User** via `UserID`. Guest join is a deliberate simplicity — anyone can join a Session with just a name, no signup required.
_Avoid_: Participant, member

**Guest**:
A Player with no `UserID` — joined by name only, no account behind them.
_Avoid_: Anonymous player, unregistered player

**User**:
A registered account holder (email + password). Distinct from Player — a User only becomes a Player by joining a specific Session.
_Avoid_: Account, member

**Rating**:
A player's self-assessed skill level on a 1–5 scale — Beginner, Improver, Intermediate, Advanced, Expert — each anchored to a concrete padel milestone, selected by label but stored and displayed as the number. Lives on the **Player** (per-session, admin-editable in the lobby), seeded from a `self_rating` on the **User** (set at registration, editable in settings). Used to make matches more competitive by sending pairs of similar strength to face each other — it influences **match-ups, not partnerships** (partner rotation is untouched). Unrated players count as the median (3). See `docs/adr/0006-rating-balances-matchups-not-partnerships.md`.
_Avoid_: Skill, level, handicap, seed

**Admin**:
Whoever holds a Session's `AdminToken` (a secret bearer token generated at creation). Admin status is the actual authorization gate for privileged actions (start, cancel, close, score entry, kick players) — checked via `isAdmin()`, independent of login state. A Session's join link without the token is read-only.
_Avoid_: Owner, moderator

**Creator**:
The User or Player who made the Session (`CreatorUserID` / `CreatorPlayerID`). Identity metadata only — used for "this is your session" UI (`IsCreator`) — not an authorization check. A Creator without the admin token cannot perform admin actions.
_Avoid_: Admin (these are related but distinct — see **Admin**)

**Game Mode**:
The pairing algorithm for a Session: `americano` (rotating partners, individual scoring) or `mexicano` (partners adapt based on standings, requires 2+ courts). A third mode, Timed Americano, was built and later removed entirely.
_Avoid_: Format, variant

**Round**:
One cycle of a Session — a set of Matches across all courts, plus the list of Players benched that cycle.

**Unlimited Rounds**:
A Session config where `rounds_total` is left unset — Rounds keep generating until the Admin ends the Session, instead of stopping at a fixed count. Chosen at setup, or entered mid-Session by tapping "Keep Playing" once a fixed count is reached (one-way: fixed → unlimited, no incremental add-N-rounds).
_Avoid_: Freeplay, Open-ended, Timed (a different, removed concept — see `docs/adr/0001-remove-timed-americano.md`)

**Bench**:
Players sitting out a given Round. Rotation guarantees a Player benched in Round N plays in Round N+1.

**Match**:
One court's game within a Round — two teams of two Players, an optional final Score, and an optional in-progress Live score.

**Standing / Leaderboard**:
The ranked, live-updating view of a Session's Players by points — public, no auth required to view.
_Avoid_: Ranking, scoreboard

**Contact**:
A saved connection between two Users (added via search), used to streamline sending Invites.

**Invite**:
A request from one User to another to join a specific Session, with a `pending | accepted | declined` status. Distinct from the public join link — Invites target a known User, not an anonymous link.

---

## Components & Design System

**Phase 1** components follow a consistent pattern: `tailwind-variants` (tv) for styling, comprehensive JSDoc, TypeScript Props interfaces, and Bits UI primitives for accessibility.

### Core Form Components

| Component | Variants / Options | Use Case |
|-----------|-------------------|----------|
| **Button** | `variant`: default, outline, secondary, ghost, destructive, destructive-solid, link `size`: xs, sm, default, lg, cta, icon, icon-xs, icon-sm, icon-lg | Primary CTAs, secondary actions, destructive actions, icon-only actions. `size="cta"` = full-width primary CTA; `destructive-solid` = filled danger. Pass `href` to render as link. |
| **Input** | `type`: text, password, email, number, search, tel, url, date, time, file | Text-based form inputs. Pair with Label for accessibility. Two-way binding: `value` for text-based, `files` for file inputs. |
| **Label** | — | Form label for inputs. Always pair with Input using `htmlFor` attribute. Supports required indicator via child `<span>`. |
| **Switch** | `size`: sm, default | Binary on/off state. Use in forms for boolean fields (vs. Toggle for state switching). Accessible: bound to Label via `htmlFor`. **Important**: Styling uses `data-[state=checked]` and `data-[state=unchecked]` Tailwind selectors (not `data-checked:`/`data-unchecked:`); bind to the `checked` prop and the component will handle state attributes automatically. |
| **Toggle** | `variant`: default, outline `size`: sm, default, lg | State switching (e.g., filters, display modes). Renders as button with `aria-pressed`. Supports icon content. |

### Semantic Components

| Component | Variants / Options | Use Case |
|-----------|-------------------|----------|
| **Badge** | `variant`: default (status), secondary (tags), destructive (errors), outline (neutral), ghost (minimal), link (clickable) | Status indicators (online, active, success), category tags (feature, type), error states. Renders as `<span>` by default, `<a>` if `href` provided. |
| **Drawer** | `size`: sm (40vh mobile, 300px desktop), md (60vh mobile, 480px), lg (80vh mobile, 640px) | Bottom drawer for side panels, settings, filters. Slides up from bottom. Use as compound component: `<Drawer>`, `<DrawerTrigger>`, `<DrawerContent>`, `<DrawerHeader>`, `<DrawerBody>`, `<DrawerFooter>`, `<DrawerClose>`. |

### Foundation Components (shared)

Built by the design-system foundation pass (#189); compose these rather than hand-rolling.

| Component | Variants / Options | Use Case |
|-----------|-------------------|----------|
| **Spinner** | `size`: sm, md, lg | Loading / in-progress states (`role="status"` + `label`). Replaces hand-rolled `animate-spin` divs and bare "Loading…". |
| **Section** | `collapsible` (default true), `open` (bindable), `maxHeight` (internal scroll), `trailing` snippet | Titled, optionally-collapsible/scrollable section block. Replaces hand-rolled Collapsible + label + chevron. |
| **PasswordInput** | — (wraps Input) | Password field with a show/hide toggle. |
| **AvatarGroup** | `max` | Stacked overlapping avatars + "+N" overflow (player previews). |
| **JoinCodeInput** | `onComplete(code)` | 4-char session join code (OTP boxes, auto-advance, paste). Shared by Home + Profile. |

`Avatar` also accepts an optional `badge` snippet (corner action, e.g. add-contact — see #174).

### Design Tokens

All components reference design tokens defined in `src/lib/design-tokens.ts` and `src/app.css` via `@theme` directive:

- **Colors**: primary, destructive, positive, warning (+ muted), secondary, muted, accent, with foreground variants; medal gold/silver/bronze (celebratory / Session complete only)
- **Spacing**: 0–4 (0, 0.25rem, 0.5rem, 0.75rem, 1rem)
- **Radius**: sm (0.25rem), md (0.5rem), default (0.75rem)
- **Shadows**: sm, md, lg
- **Fonts**: sans (Inter)
- **Animations**: shake, ptr-fade

### Accessibility

All Phase 1 components:
- ✓ Built on Bits UI primitives (Dialog, Label, etc.) for focus management and keyboard interaction
- ✓ Support ARIA attributes: `aria-invalid`, `aria-pressed`, `aria-expanded`, `aria-modal`
- ✓ Require explicit pairing: Label with Input, proper semantic elements (heading for titles, paragraph for descriptions)
- ✓ Keyboard: Tab navigates, Escape closes (Drawer), Enter activates (Button)

### Patterns

**Form Field with Validation:**
```svelte
<Label htmlFor="email">Email <span class="text-destructive">*</span></Label>
<Input id="email" type="email" ariaInvalid={hasError} />
```

**Drawer with Actions:**
```svelte
<Drawer>
  <DrawerTrigger>Open Settings</DrawerTrigger>
  <DrawerContent size="md">
    <DrawerHeader>
      <DrawerTitle>Settings</DrawerTitle>
    </DrawerHeader>
    <DrawerBody>{/* form fields */}</DrawerBody>
    <DrawerFooter>
      <DrawerClose>Cancel</DrawerClose>
      <Button>Save</Button>
    </DrawerFooter>
  </DrawerContent>
</Drawer>
```

### Extending Components

Before creating a new component, ask: **Can I extend an existing one?**

- **New visual variant** (color, size, state) → Add to `tv()` variants
- **Different props** → Consider composition with existing components
- **Fundamentally different behavior** → Create custom component (e.g., Drawer required custom implementation; Dialog is for centered modals)

See `docs/guides/component-patterns.md` for the full decision tree.
