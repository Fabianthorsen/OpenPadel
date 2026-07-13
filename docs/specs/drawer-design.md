# Custom Drawer Component Design

**Status:** Design Phase  
**Target:** Phase 1 Implementation (Ticket #159)  
**Scope:** Full ownership, Bits UI foundation, OpenPadel-specific styling and variants

---

## Problem Statement

OpenPadel needs a side-panel drawer for:
- Settings/preferences panels
- Navigation drawers (mobile)
- Filters and search panels
- Action sheets (create session, etc.)

Existing options:
- **Dialog (centered modal)** — wrong UX for side panels
- **Sheet (bottom drawer)** — only vertical positioning
- **External package** — dependency bottleneck, limited customization

**Solution:** Build a custom Drawer on Bits UI primitives, positioned left/right, with full control over animation, sizing, and accessibility.

---

## Design Goals

1. **Ownership** — Full control; no external drawer library dependency
2. **Consistency** — Aligns with design system patterns (tv(), JSDoc, accessibility)
3. **Flexibility** — Position (left/right), size (sm/md/lg), animation timing
4. **Accessibility** — Focus trapping, escape to close, ARIA roles
5. **Performance** — CSS transitions only (no animation libraries)
6. **Mobile-first** — Works seamlessly on small screens

---

## Component Architecture

### Structure

Drawer is a compound component family following Bits UI pattern:

```
<Drawer>
  <DrawerTrigger>Open Drawer</DrawerTrigger>
  <DrawerContent>
    <DrawerHeader>
      <DrawerTitle>Panel Title</DrawerTitle>
      <DrawerDescription>Optional subtitle</DrawerDescription>
    </DrawerHeader>
    <DrawerBody>
      {/* Main content */}
    </DrawerBody>
    <DrawerFooter>
      <DrawerClose>Cancel</DrawerClose>
      <Button>Save</Button>
    </DrawerFooter>
  </DrawerContent>
</Drawer>
```

### Sub-Components

| Component | Purpose | HTML | Bits UI Base |
|-----------|---------|------|--------------|
| **Drawer** | Root container, state management | none (context provider) | Dialog.Root |
| **DrawerTrigger** | Opens drawer | `<button>` | Dialog.Trigger |
| **DrawerContent** | Container for all drawer content | `<div>` | Dialog.Content |
| **DrawerHeader** | Top section with title/close | `<div>` | — |
| **DrawerTitle** | Panel title (semantic heading) | `<h2>` | Dialog.Title |
| **DrawerDescription** | Subtitle or description | `<p>` | Dialog.Description |
| **DrawerBody** | Main scrollable content area | `<div>` | — |
| **DrawerFooter** | Bottom section with actions | `<div>` | — |
| **DrawerClose** | Close button | `<button>` | Dialog.Close |
| **DrawerPortal** | Portal for overlay/content | none | Dialog.Portal |
| **DrawerOverlay** | Backdrop/scrim | `<div>` | Dialog.Overlay |

---

## Variants & Props

### DrawerContent Variants

**Position (where drawer slides from):**
- `left` — slides in from left edge (default)
- `right` — slides in from right edge

**Size (width):**
- `sm` — 320px (mobile-friendly, narrow)
- `md` — 480px (standard, comfortable for content)
- `lg` — 640px (wide, for complex forms/lists)

**Example:**
```typescript
export const drawerContentVariants = tv({
  base: '...',
  variants: {
    position: {
      left: 'inset-y-0 left-0 slide-in-from-left ...',
      right: 'inset-y-0 right-0 slide-in-from-right ...'
    },
    size: {
      sm: 'w-[320px]',
      md: 'w-[480px]',
      lg: 'w-[640px]'
    }
  },
  defaultVariants: { position: 'left', size: 'md' }
});
```

### Overlay Variants

**Backdrop opacity** (configurable via prop):
- `default` — 40% opacity (light scrim)
- `medium` — 60% opacity (standard)
- `strong` — 80% opacity (dark scrim)

---

## Animation Strategy

### CSS Transitions (No Animation Library)

**Drawer slide-in:**
```css
@keyframes slide-in-from-left {
  from { transform: translateX(-100%); }
  to { transform: translateX(0); }
}

@keyframes slide-in-from-right {
  from { transform: translateX(100%); }
  to { transform: translateX(0); }
}

.drawer-content[data-position="left"] {
  animation: slide-in-from-left 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}
```

**Exit animation** (when closing):
- Same transition, reversed (handled by Bits UI Dialog close behavior)

**Duration:** 300ms (snappy, responsive to user)

**Easing:** `cubic-bezier(0.4, 0, 0.2, 1)` (material design standard decelerate)

**Backdrop fade:**
```css
.drawer-overlay {
  animation: fade-in 0.2s ease-out;
}

@keyframes fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}
```

---

## Accessibility

### Focus Management

- **On open:** Focus moves to close button (top-right) or first focusable element in header
- **On close:** Focus returns to trigger button (Dialog.Root handles this)
- **Trap:** Focus cycles within drawer (Bits UI Dialog.Root provides this)

### ARIA Attributes

- `role="dialog"` on DrawerContent (handled by Bits UI Dialog)
- `aria-labelledby="drawer-title"` (links to DrawerTitle)
- `aria-describedby="drawer-description"` (optional, for DrawerDescription)
- `aria-modal="true"` (indicates modal behavior)

### Keyboard Interaction

- **Escape** → Close drawer (Bits UI Dialog handles)
- **Tab** → Cycle through focusable elements within drawer
- **Enter** → Activate buttons/links within drawer

### Screen Reader Announcements

- Title is announced when drawer opens (via aria-labelledby)
- Description provides context (via aria-describedby)
- DrawerClose button label: "Close" (concise, clear)

---

## Props Interface

### Drawer (Root)

```typescript
export interface DrawerProps {
  /** Controlled open state. If undefined, drawer is uncontrolled. */
  open?: boolean;

  /** Callback when open state changes. */
  onOpenChange?: (open: boolean) => void;

  /** If true, drawer cannot be closed (rare, use cautiously). */
  modal?: boolean;
}
```

### DrawerContent

```typescript
export interface DrawerContentProps {
  /** Position: left | right. Determines which edge drawer slides from. */
  position?: 'left' | 'right';

  /** Size: sm | md | lg. Controls drawer width (320px, 480px, 640px). */
  size?: 'sm' | 'md' | 'lg';

  /** Custom CSS class (composed with tailwind-variants). */
  class?: string;
}
```

### DrawerTrigger

```typescript
export interface DrawerTriggerProps {
  /** If true, trigger button is disabled. */
  disabled?: boolean;

  /** Custom CSS class. */
  class?: string;
}
```

### DrawerOverlay (Backdrop)

```typescript
export interface DrawerOverlayProps {
  /** Backdrop opacity: 'light' | 'medium' | 'dark' (40%, 60%, 80%). */
  opacity?: 'light' | 'medium' | 'dark';

  /** Custom CSS class. */
  class?: string;
}
```

---

## JSDoc & Type Discoverability

### DrawerContent JSDoc Example

```typescript
/**
 * Drawer side panel container. Slides in from left or right edge.
 * Positioned absolutely over page content with backdrop overlay.
 *
 * @example
 * <Drawer>
 *   <DrawerTrigger>Open Drawer</DrawerTrigger>
 *   <DrawerContent position="right" size="md">
 *     <DrawerHeader>
 *       <DrawerTitle>Settings</DrawerTitle>
 *     </DrawerHeader>
 *     <DrawerBody>Content here</DrawerBody>
 *   </DrawerContent>
 * </Drawer>
 *
 * @example
 * <DrawerContent position="left" size="lg">
 *   Mobile navigation with complex filters
 * </DrawerContent>
 */
export interface DrawerContentProps {
  /**
   * position: left | right. Determines which edge drawer slides from.
   * - left: slides from left edge (common for navigation)
   * - right: slides from right edge (common for settings/filters)
   */
  position?: 'left' | 'right';

  /**
   * size: sm | md | lg. Controls drawer width.
   * - sm: 320px (mobile, tight content)
   * - md: 480px (standard, balanced)
   * - lg: 640px (wide, complex forms/lists)
   */
  size?: 'sm' | 'md' | 'lg';
}
```

---

## Responsive Behavior

### Mobile (< 768px)

- Drawer sizes snap to viewport width (max-width: 90vw)
- Position: `left` on small screens (more natural thumb reach)
- Backdrop opacity: `medium` (strong visual separation on small screens)

### Tablet & Desktop (>= 768px)

- Drawer respects size variants exactly (320px, 480px, 640px)
- Position: flexible (left or right per use case)
- Backdrop opacity: configurable

### RTL Support

- Position: `left` → RTL-aware (right edge in RTL context)
- Position: `right` → RTL-aware (left edge in RTL context)
- Slide animations adapt via `rtl:translate-x-[calc(...)]` Tailwind directive

---

## Usage Patterns

### Basic Drawer

```svelte
<Drawer>
  <DrawerTrigger>Open Settings</DrawerTrigger>
  <DrawerContent>
    <DrawerHeader>
      <DrawerTitle>Settings</DrawerTitle>
    </DrawerHeader>
    <DrawerBody>
      {/* Form fields, toggles, etc. */}
    </DrawerBody>
    <DrawerFooter>
      <DrawerClose>Close</DrawerClose>
      <Button variant="primary">Save</Button>
    </DrawerFooter>
  </DrawerContent>
</Drawer>
```

### Controlled Drawer (Svelte 5)

```svelte
<script>
  let drawerOpen = $state(false);
</script>

<Drawer open={drawerOpen} onOpenChange={(open) => drawerOpen = open}>
  <DrawerTrigger>Filters</DrawerTrigger>
  <DrawerContent position="right" size="md">
    <DrawerHeader>
      <DrawerTitle>Filter Results</DrawerTitle>
      <DrawerDescription>Narrow down matches</DrawerDescription>
    </DrawerHeader>
    <DrawerBody>
      {/* Filter controls */}
    </DrawerBody>
  </DrawerContent>
</Drawer>
```

### Mobile Navigation Drawer

```svelte
<Drawer>
  <DrawerTrigger variant="ghost" size="icon">
    <Icon name="menu" />
  </DrawerTrigger>
  <DrawerContent position="left" size="sm">
    <DrawerHeader>
      <DrawerTitle>Menu</DrawerTitle>
    </DrawerHeader>
    <DrawerBody>
      <nav class="flex flex-col gap-2">
        <a href="/">Home</a>
        <a href="/sessions">Sessions</a>
        <a href="/profile">Profile</a>
      </nav>
    </DrawerBody>
  </DrawerContent>
</Drawer>
```

---

## Implementation Checklist

### Files to Create

- [ ] `web/src/lib/components/ui/drawer/drawer.svelte` — Root component
- [ ] `web/src/lib/components/ui/drawer/drawer-trigger.svelte` — Trigger button
- [ ] `web/src/lib/components/ui/drawer/drawer-content.svelte` — Main content container
- [ ] `web/src/lib/components/ui/drawer/drawer-header.svelte` — Header section
- [ ] `web/src/lib/components/ui/drawer/drawer-body.svelte` — Scrollable content area
- [ ] `web/src/lib/components/ui/drawer/drawer-footer.svelte` — Footer section
- [ ] `web/src/lib/components/ui/drawer/drawer-title.svelte` — Semantic title
- [ ] `web/src/lib/components/ui/drawer/drawer-description.svelte` — Optional subtitle
- [ ] `web/src/lib/components/ui/drawer/drawer-close.svelte` — Close button
- [ ] `web/src/lib/components/ui/drawer/drawer-overlay.svelte` — Backdrop overlay
- [ ] `web/src/lib/components/ui/drawer/drawer-portal.svelte` — Portal wrapper
- [ ] `web/src/lib/components/ui/drawer/index.ts` — Public exports

### CSS to Define (in `app.css`)

```css
/* Slide-in animations */
@keyframes slide-in-from-left {
  from { transform: translateX(-100%); }
  to { transform: translateX(0); }
}

@keyframes slide-in-from-right {
  from { transform: translateX(100%); }
  to { transform: translateX(0); }
}

@keyframes fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

/* Drawer animations (data-state="open") */
[data-state="open"][data-position="left"] {
  animation: slide-in-from-left 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

[data-state="open"][data-position="right"] {
  animation: slide-in-from-right 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.drawer-overlay[data-state="open"] {
  animation: fade-in 0.2s ease-out;
}
```

### Types to Define (in each component file)

- `DrawerProps` — Drawer.Root
- `DrawerTriggerProps` — Drawer.Trigger
- `DrawerContentProps` — Drawer.Content
- `DrawerOverlayProps` — Drawer.Overlay
- `DrawerHeaderProps` — Drawer.Header
- `DrawerFooterProps` — Drawer.Footer
- `DrawerBodyProps` — Drawer.Body

---

## Testing Strategy (Phase 2, Ticket #161)

### Component Tests

1. **Render and state:**
   - Drawer renders with trigger button
   - Clicking trigger opens drawer
   - Clicking close button closes drawer
   - Escape key closes drawer

2. **Variants:**
   - All position variants (left, right) render correctly
   - All size variants (sm, md, lg) apply correct widths
   - Backdrop opacity variants work

3. **Accessibility:**
   - Focus trap works (tab cycles within drawer)
   - Focus returns to trigger on close
   - ARIA attributes present (role, aria-labelledby, aria-modal)

4. **Animation:**
   - Drawer slides in from correct edge
   - Backdrop fades in
   - Animations respect data-state="open" attribute

---

## Out of Scope

- Custom animation libraries (Framer Motion, etc.)
- Swipe-to-dismiss gestures (mobile UX enhancement, Phase 2+)
- Nested drawers (edge case, not needed)
- Drawer stacking (only one drawer open at a time)
- Complex composition patterns (keep simple)

---

## References

- **Bits UI Dialog:** https://bits-ui.com/docs/components/dialog
- **shadcn/ui Drawer:** https://ui.shadcn.com/docs/components/drawer
- **Design tokens:** `docs/guides/component-patterns.md`
- **Accessibility:** WCAG 2.1 Level AA, ARIA Authoring Practices Guide (APG)

---

## Next Steps

- ✅ **#158 (This ticket):** Design spec — DONE
- **#159:** Implement Drawer component using this spec
- **#160:** Add Drawer to CONTEXT.md components table
- **#161:** Write integration tests for all Phase 1 components
