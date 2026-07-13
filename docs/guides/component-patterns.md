# Component Patterns & Design System Guide

Guide for writing reusable, AI-discoverable components following the tailwind-variants pattern, JSDoc conventions, composition patterns, and extending vs. creating custom components.

## Quick Start

A reusable component follows this pattern:

```svelte
<script lang="ts">
  import { tv, type VariantProps } from 'tailwind-variants';
  import type { HTMLButtonAttributes } from 'svelte/elements';

  export const buttonVariants = tv({
    base: 'inline-flex items-center justify-center rounded-lg font-medium transition-colors',
    variants: {
      variant: {
        primary: 'bg-primary text-primary-foreground hover:bg-primary/90',
        secondary: 'bg-secondary text-secondary-foreground hover:bg-secondary/80',
      },
      size: {
        sm: 'h-8 px-3 text-sm',
        md: 'h-10 px-4 text-base',
      },
    },
    defaultVariants: {
      variant: 'primary',
      size: 'md',
    },
  });

  export type ButtonVariant = VariantProps<typeof buttonVariants>['variant'];
  export type ButtonSize = VariantProps<typeof buttonVariants>['size'];

  /**
   * Primary action button. See "JSDoc: Documenting Components for AI" section for full documentation pattern.
   */
  export interface ButtonProps extends HTMLButtonAttributes {
    /** variant: primary | secondary. */
    variant?: ButtonVariant;
    /** size: sm | md. */
    size?: ButtonSize;
  }

  let { variant, size, class: className, ...rest }: ButtonProps = $props();
</script>

<button class={buttonVariants({ variant, size, class: className })} {...rest}>
  {#if children}
    {@render children?.()}
  {/if}
</button>
```

---

## The Pattern: tailwind-variants (tv)

`tailwind-variants` (tv) creates type-safe, composable Tailwind classes as the single source of truth for component styling.

### Structure: base, variants, defaultVariants

```typescript
const componentVariants = tv({
  base: 'flex items-center justify-center rounded-lg transition-colors',
  variants: {
    variant: {
      primary: 'bg-primary text-white',
      secondary: 'bg-secondary text-gray-900',
    },
    size: {
      sm: 'h-8 px-3 text-sm',
      md: 'h-10 px-4 text-base',
      lg: 'h-12 px-6 text-lg',
    },
    state: {
      disabled: 'opacity-50 cursor-not-allowed',
      loading: 'cursor-wait',
    },
  },
  defaultVariants: {
    variant: 'primary',
    size: 'md',
  },
});
```

### Extracting Types

Always export variant types so JSDoc and consumers can reference them:

```typescript
export type ButtonVariant = VariantProps<typeof buttonVariants>['variant'];
export type ButtonSize = VariantProps<typeof buttonVariants>['size'];
```

---

## JSDoc: Documenting Components for AI

Good JSDoc lets Claude understand your component without reading implementation.

### Template

```typescript
/**
 * Component name — short purpose statement.
 * Longer explanation of when and why to use this component.
 * 
 * @example
 * <ComponentName variant="primary">Label</ComponentName>
 * 
 * @example
 * <ComponentName variant="secondary" disabled>Disabled</ComponentName>
 */
export interface ComponentProps {
  /** Category of variants. Explain each:
   * - variant1: use when...
   * - variant2: use when...
   */
  variant?: ComponentVariant;
  
  /** Size options:
   * - sm: compact contexts
   * - md: default, most use cases
   * - lg: prominent actions
   */
  size?: ComponentSize;
  
  /** Disabled state: component cannot interact. */
  disabled?: boolean;
}
```

### Key JSDoc Rules

1. **One line before examples**: Brief component purpose
2. **@example blocks**: Show common usage patterns (2-3 examples)
3. **Variant descriptions**: Explain *when* to use each, not just names
4. **State documentation**: disabled, aria-invalid, aria-expanded, aria-pressed — explain what each means
5. **Reference design tokens**: `See tokens.colors for available colors` (import `{ tokens } from '$lib/design-tokens'`)

### States: Accessibility Attributes

Always document these states in your Props interface:

| State | When Used | Example |
|-------|-----------|---------|
| `disabled` | Component cannot interact | Button form submission disabled, input readonly |
| `aria-invalid` | Invalid input/data | Form field with validation error |
| `aria-expanded` | Content is expanded/collapsed | Accordion, drawer, menu |
| `aria-pressed` | Toggle is on/off | Toggle button, switch |
| `aria-busy` | Operation in progress | Loading state, async operation |

Document like this:

```typescript
/** Disabled state: component is inactive and cannot be interacted with. */
disabled?: boolean;

/** Invalid state: indicates aria-invalid for screen readers and error styling. */
ariaInvalid?: boolean;

/** Expanded state: true when associated content is visible. */
ariaExpanded?: boolean;
```

### JSDoc Template: Copy-Paste Reference

Use this template for every component. Customize the variant/size options and examples:

```typescript
/**
 * Component name and brief purpose.
 * 
 * @example
 * <ComponentName>Default</ComponentName>
 * 
 * @example
 * <ComponentName variant="secondary" disabled>Disabled</ComponentName>
 */
export interface ComponentNameProps {
  /** variant: primary | secondary | tertiary. See tokens.colors for palette. */
  variant?: ComponentVariant;
  
  /** size: sm | md | lg. sm for compact, md for default, lg for prominent. */
  size?: ComponentSize;
  
  /** Disabled: component inactive, cannot interact. */
  disabled?: boolean;
  
  /** Aria-expanded: true when associated content is visible (accordion, drawer). */
  ariaExpanded?: boolean;
}
```

### Snippet: Variant Documentation

For each variant prop:

```typescript
/**
 * variant: [option1 | option2 | option3]. See tokens.colors for palette.
 * - option1: use when [context/purpose]
 * - option2: use when [context/purpose]
 * - option3: use when [context/purpose]
 */
variant?: ComponentVariant;
```

Example (Button):
```typescript
/**
 * variant: primary | secondary | ghost | destructive. See tokens.colors for palette.
 * - primary: main CTA, preferred action
 * - secondary: supporting action, alternative flow
 * - ghost: low-emphasis action, deemphasized
 * - destructive: destructive action, requires confirmation
 */
variant?: ButtonVariant;
```

### Snippet: State Documentation

For accessibility states:

```typescript
/** disabled: component inactive, cannot interact. */
disabled?: boolean;

/** ariaInvalid: true if input data is invalid; indicates error state. */
ariaInvalid?: boolean;

/** ariaExpanded: true when associated content is visible. */
ariaExpanded?: boolean;

/** ariaPressed: true when toggle is pressed/active. */
ariaPressed?: boolean;
```

### Snippet: Slot/Children Documentation

For components accepting children:

```typescript
/** Children content rendered inside component. Supports text, elements, icons. */
children?: Snippet;
```

### Full Example: Alert Component

Complete, production-ready JSDoc for a hypothetical Alert component:

```typescript
import type { Snippet } from 'svelte';
import type { HTMLDivAttributes } from 'svelte/elements';

export const alertVariants = tv({
  base: 'rounded-lg border px-4 py-3 flex gap-3',
  variants: {
    variant: {
      info: 'border-blue-200 bg-blue-50 text-blue-900',
      success: 'border-green-200 bg-green-50 text-green-900',
      warning: 'border-yellow-200 bg-yellow-50 text-yellow-900',
      error: 'border-red-200 bg-red-50 text-red-900',
    },
  },
  defaultVariants: { variant: 'info' },
});

export type AlertVariant = VariantProps<typeof alertVariants>['variant'];

/**
 * Alert message container for notifications and status updates.
 * 
 * @example
 * <Alert variant="success">Operation completed successfully.</Alert>
 * 
 * @example
 * <Alert variant="error">
 *   <Icon name="alert-circle" />
 *   <span>An error occurred. Please try again.</span>
 * </Alert>
 */
export interface AlertProps extends HTMLDivAttributes {
  /**
   * variant: info | success | warning | error. See tokens.colors for palette.
   * - info: neutral information
   * - success: positive outcome
   * - warning: caution or attention needed
   * - error: problem or failure state
   */
  variant?: AlertVariant;
  
  /** Children content: text, icons, or other elements. Typically text + icon. */
  children?: Snippet;
  
  /** role: aria-live="polite" for dynamic updates (recommended for alerts). */
  role?: 'status' | 'alert';
  
  /** ariaLive: 'polite' for non-urgent, 'assertive' for urgent notifications. */
  ariaLive?: 'polite' | 'assertive';
}

/**
 * Component implementation:
 */
// <script lang="ts">
//   let { variant, children, role = 'status', ariaLive = 'polite', ...rest }:
//     AlertProps = $props();
// </script>
//
// <div class={alertVariants({ variant })} {role} aria-live={ariaLive} {...rest}>
//   {#if children}
//     {@render children()}
//   {/if}
// </div>
```

---

## Composition Patterns

### Slots: Making Components Composable

Use slot rendering for flexible content:

```svelte
<script>
  let { children } = $props();
</script>

<button>
  {#if children}
    {@render children?.()}
  {/if}
</button>
```

Usage:
```svelte
<Button>Just text</Button>
<Button>
  <Icon name="check" />
  <span>With icon</span>
</Button>
```

### Composition: Label + Input

Document common pairings:

```svelte
<!-- Example: Label with Input -->
<Label htmlFor="email">Email address</Label>
<Input id="email" type="email" />
```

Document in JSDoc:
```typescript
/**
 * Label for form inputs. Always pair with an Input using htmlFor.
 * 
 * @example
 * <Label htmlFor="name">Name</Label>
 * <Input id="name" />
 */
```

---

## Design Tokens in Components

### Reference Tokens in JSDoc

Document available token categories to help AI understand options:

```typescript
import { tokens } from '$lib/design-tokens';

/**
 * Button variant: primary, secondary, destructive, ghost.
 * See tokens.colors for full palette (tokens.colors.primary, tokens.colors.destructive, etc.).
 */
variant?: ButtonVariant;
```

Use Tailwind class names in `tailwind-variants` (e.g., `bg-primary`, `text-destructive`), not token interpolation. Tokens should already be available as CSS variables in your theme.

---

## Extending Components: The Decision Tree

Before creating a new component, ask: **Can I extend an existing one?**

### Decision Tree

```
Do I have a new visual variant (color, size, state)?
├─ YES: Add a variant to tv → Done
│       Example: Button needs a "pill" variant
│
├─ NO: Do I need different props?
│      ├─ YES: Consider composition first
│      │        Example: Combining Button + Icon
│      │
│      └─ NO: Do I need fundamentally different behavior?
│             ├─ YES: Create custom component
│             │        Example: Custom Drawer (different interaction model)
│             │
│             └─ NO: Reuse existing component + props
```

### Example: Adding a Pill Variant to Button

**Problem:** Button needs a pill-shaped variant (fully rounded).

**Solution:** Extend the existing Button component.

```typescript
// In button/button.svelte

export const buttonVariants = tv({
  base: '...',
  variants: {
    variant: {
      primary: '...',
      secondary: '...',
      pill: 'rounded-full px-6', // Add pill variant
    },
  },
});

/**
 * Button variant:
 * - primary: standard button with sharp corners
 * - secondary: outlined style
 * - pill: fully rounded, use for pills/badges
 */
export type ButtonVariant = VariantProps<typeof buttonVariants>['variant'];
```

✅ **Don't create a Pill component.** Extend Button.

### Example: When to Create Custom

**Problem:** Drawer needs side-panel interaction with focus trapping, backdrop, and slide animation.

**Situation:** 
- Existing Dialog handles modals (centered)
- Existing Sheet is for bottom sheets
- Neither fits side-panel with custom animation

**Decision:** Create custom Drawer component because:
- Fundamentally different positioning (side, not center)
- Unique interaction model (swipe to dismiss, focus trap)
- Cannot achieve with composition + existing components

✅ **Create custom Drawer.** Cannot extend existing components.

---

## Testing for Reusability

### What Makes a Component Reusable?

1. **Clear props** — Props interface is complete and type-safe
2. **Documented variants** — Each variant has clear use case
3. **Composable** — Accepts children/slots for flexibility
4. **Accessible** — Handles aria-* attributes correctly
5. **AI-discoverable** — JSDoc explains API without reading code

### Testing Checklist

- [ ] All variants can be rendered and styled correctly
- [ ] Props are type-safe and defaulted appropriately
- [ ] JSDoc is complete and explains each variant
- [ ] Component is composable (slots work)
- [ ] Disabled state works correctly
- [ ] Tests pass

---

## Common Patterns

### Button with Icon

```svelte
<script>
  import Icon from './Icon.svelte';
</script>

<Button variant="primary">
  <Icon name="check" />
  <span>Save</span>
</Button>
```

Use size="icon" for icon-only buttons:

```svelte
<Button variant="ghost" size="icon">
  <Icon name="menu" />
</Button>
```

### Input with Label

```svelte
<Label htmlFor="email">Email</Label>
<Input id="email" type="email" ariaInvalid={hasError} />
{#if hasError}
  <span class="text-destructive text-sm">{errorMessage}</span>
{/if}
```

### Toggle Group

```svelte
<ToggleGroup value="option1" on:change={(e) => selected = e.detail}>
  <Toggle value="option1">Option 1</Toggle>
  <Toggle value="option2">Option 2</Toggle>
  <Toggle value="option3">Option 3</Toggle>
</ToggleGroup>
```

---

## FAQ

**Q: When should I add a new prop vs. a new variant?**  
A: Use a variant if it changes *visual appearance*. Use a prop if it changes *behavior* or *structure*. Example: `size="lg"` is a variant (visual), `disabled={true}` is a prop (behavior).

**Q: Should I support arbitrary Tailwind classes?**  
A: No. Constrain to defined variants using `tv()`. This ensures consistency and makes design changes easier.

**Q: How do I know if a component is well-documented?**  
A: Claude should be able to use it correctly without reading the source code, based only on JSDoc and TypeScript types.

**Q: What if my component doesn't fit the tv pattern?**  
A: Use `tv` for styling components. Use regular props for behavioral components. Most UI components need `tv`.

---

## See Also

- [`design-tokens.ts`](../../web/src/lib/design-tokens.ts) — Available design tokens
- Button component — Reference implementation
- [CLAUDE.md](../../CLAUDE.md) — Project standards
