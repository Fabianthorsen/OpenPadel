# Custom Drawer Component Design

**Status:** Implemented (Ticket #159)
**Scope:** Full ownership, Bits UI foundation, OpenPadel-specific styling and variants

> **Note on orientation:** this component was originally designed as a left/right side panel.
> During implementation it was changed to a **bottom drawer** (slides up from the bottom edge)
> because that matches the app's actual UX — the create-session and score-entry flows are
> bottom sheets. This document describes the as-built bottom drawer.

---

## Problem Statement

OpenPadel needs a bottom sheet / drawer for:

- The create-session flow (`CreateDrawer`)
- Courtside score entry (numpad on the session page)
- Future settings / filter panels

Existing options and why they don't fit:

- **Dialog (centered modal)** — wrong UX; we want a sheet anchored to the bottom edge.
- **External package (`vaul-svelte`)** — dependency bottleneck, and it had already drifted (dead
  `data-[vaul-drawer-direction]` selectors in consumers). Replaced by this component.

**Solution:** a custom Drawer built on Bits UI `Dialog` primitives, anchored to the bottom, with
full control over animation, sizing, and accessibility.

---

## Design Goals

1. **Ownership** — no external drawer library dependency (Bits UI `Dialog` only).
2. **Consistency** — aligns with design-system patterns (`tv()`, JSDoc, exported Props types).
3. **Flexibility** — size (sm/md/lg) controls height; horizontal centering via consumer classes.
4. **Accessibility** — focus trap, escape to close, and dialog ARIA roles inherited from Bits UI.
5. **Performance** — CSS keyframe animations only, no animation libraries.
6. **Mobile-first** — full-width bottom sheet on phones; centered, width-capped card on desktop.

---

## Component Architecture

A compound component family. All sub-components wrap a Bits UI `Dialog` part and share a common
prop base (`DrawerPrimitiveProps` in `types.ts`, which narrows `id` to `string` for Bits UI).

| Component          | Element / Bits UI part      | Purpose                                        |
| ------------------ | --------------------------- | ---------------------------------------------- |
| `Drawer`           | `Dialog.Root`               | Root; owns `open` state (`bind:open`)          |
| `DrawerTrigger`    | `Dialog.Trigger`            | Opens the drawer                               |
| `DrawerContent`    | `Dialog.Portal` + `Overlay` + `Content` | Bottom sheet container + backdrop  |
| `DrawerHeader`     | `<div>`                     | Header section (title + close)                 |
| `DrawerTitle`      | `Dialog.Title`              | Accessible heading (labels the dialog)         |
| `DrawerDescription`| `Dialog.Description`        | Optional subtitle (describes the dialog)       |
| `DrawerBody`       | `<div>`                     | Scrollable main content area                   |
| `DrawerFooter`     | `<div>`                     | Action button row                              |
| `DrawerClose`      | `Dialog.Close`              | Closes the drawer                              |

The overlay and portal are rendered **internally by `DrawerContent`** — they are not separate
public components. (Bits UI requires `Overlay` and `Content` to live inside `Portal`.)

```
<Drawer bind:open>
  <DrawerTrigger>Open</DrawerTrigger>
  <DrawerContent size="md">
    <DrawerHeader>
      <DrawerTitle>Title</DrawerTitle>
      <DrawerClose>×</DrawerClose>
    </DrawerHeader>
    <DrawerBody>…</DrawerBody>
    <DrawerFooter>…</DrawerFooter>
  </DrawerContent>
</Drawer>
```

---

## Variants

`drawerContentVariants` (tailwind-variants) — the drawer is always anchored to the bottom
(`fixed inset-x-0 bottom-0`); the only variant is **size**, which caps height:

| size | mobile height | desktop height |
| ---- | ------------- | -------------- |
| `sm` | `max-h-[40vh]` | `max-h-[300px]` |
| `md` (default) | `max-h-[60vh]` | `max-h-[480px]` |
| `lg` | `max-h-[80vh]` | `max-h-[640px]` |

Horizontal sizing/centering on desktop is left to the consumer (e.g.
`class="mx-auto w-full max-w-[480px]"`), so a drawer can be a full-width sheet or a centered card
without baking one choice into the component.

`drawerCloseVariants` styles the close button (ghost-style icon/text button).

---

## Animation

CSS keyframes scoped to `drawer-content.svelte`, driven by the Bits UI `data-state` attribute:

- **Open** — `slide-up` (translateY 100% → 0), 300ms `cubic-bezier(0.4, 0, 0.2, 1)`.
- **Close** — `slide-down` (translateY 0 → 100%), same duration/easing.
- **Overlay** — opacity transition (200ms) on `data-[state=open|closed]`.

No animation library; no dependency on tokens outside the component.

---

## Accessibility

Inherited from Bits UI `Dialog`:

- Focus is trapped within the content while open and restored to the trigger on close.
- `Escape` closes the drawer; clicking the overlay closes it.
- `DrawerTitle`/`DrawerDescription` render `Dialog.Title`/`Dialog.Description`, which Bits UI wires
  to the dialog via `aria-labelledby`/`aria-describedby`.
- `DrawerContent` sets `aria-modal="true"`; the overlay is `aria-hidden`.

---

## Out of Scope

- Left/right/top orientations (bottom only).
- Configurable overlay opacity (fixed 40% scrim).
- Swipe-to-dismiss gestures, snap points, nested drawers.

---

## Consumers

- `CreateDrawer.svelte` — create-session sheet (`max-w-[480px]`, centered on desktop).
- `routes/s/[id]/+page.svelte` — courtside score-entry numpad.
