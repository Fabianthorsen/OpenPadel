# Design System Phase 1: Reusable Components & Tokens Architecture

**Status:** Ready for Implementation  
**Initiated:** 2026-07-12  
**Target Completion:** TBD

---

## Problem Statement

OpenPadel's component library has evolved inconsistently:
- **Inconsistent patterns** — some components use `tailwind-variants` (Button), others hardcode inline Tailwind classes (Input)
- **Duplication** — similar styling logic reimplemented across components with no shared source of truth
- **Rigid components** — limited dynamic/responsive behavior due to lack of proper variant systems
- **Undocumented for AI** — component APIs are not discoverable; AI assistants cannot reliably compose components or know what variants exist
- **No design token documentation** — tokens exist in CSS but aren't exported or referenced in TypeScript, making them invisible to tooling

This inconsistency creates friction: developers spend time tweaking inline styles, AI doesn't know component capabilities, and design changes require hunting through multiple files.

---

## Solution

Establish a foundational design system layer with:

1. **Unified component pattern** — all components use `tailwind-variants` for consistent, documented variant definitions
2. **Centralized token architecture** — design tokens defined in CSS (`@theme`), mirrored in TypeScript, exportable for programmatic use
3. **AI-discoverable documentation** — JSDoc on component Props types + CONTEXT.md component table
4. **Incremental refactoring** — Phase 1 focuses on core building blocks; remaining components migrate as touched

This creates a single source of truth for design decisions, reduces duplication, and enables AI to understand component APIs without guessing.

---

## User Stories

1. As a developer, I want all my components to follow the same pattern, so that I can predict and understand any component's structure
2. As a developer, I want component variants defined in one place, so that I don't duplicate styling logic across files
3. As a developer, I want to know all available props/variants for a component by reading its JSDoc, so that I don't have to dig through implementation
4. As Claude (AI assistant), I want to understand what variants and props each component accepts, so that I can generate correct component code without guessing
5. As Claude, I want access to design tokens in TypeScript, so that I can compose dynamic styles or reference values in code
6. As a developer, I want design tokens documented in CONTEXT.md, so that I have a central reference for color, spacing, radius, and other theme values
7. As a designer, I want tokens to be defined once in CSS, so that updates propagate automatically to all components
8. As a developer, I want to add new component variants without modifying multiple files, so that the work stays localized
9. As a developer, I want the custom drawer component to be fully owned by OpenPadel, so that we're not dependent on an external package
10. As a developer, I want the custom drawer to follow our design system patterns, so that it feels consistent with other components
11. As a developer, I want to migrate existing components incrementally, so that I can refactor high-pain components first without a big-bang risk
12. As a developer, I want components to be responsive and dynamic by default, so that they adapt to different screen sizes and use cases without inline adjustments

---

## Implementation Decisions

### 1. Design Token Architecture

**Decision:** Keep design tokens in CSS `app.css` using the `@theme` directive (Tailwind 4 native); mirror them in a TypeScript module for programmatic access.

**Why:** Tailwind 4's `@theme` directive is the recommended approach. CSS is the source of truth for styling; TypeScript mirrors enable tooling (JSDoc, AI) and component composition. No duplication — one definition, two consumption paths.

**What changes:**
- Tokens already defined in `app.css` via `@theme` — no changes needed here
- New file: `src/lib/design-tokens.ts` exports TypeScript constants mirroring CSS variable names
  - Structure: nested objects (colors, spacing, radius, fonts) mapping to `var(--token-name)` values
  - Type-safe via `as const`
  - Allows developers and AI to reference tokens in code

**Example:**
```typescript
export const tokens = {
  colors: {
    primary: 'var(--color-primary)',
    destructive: 'var(--color-destructive)',
  },
  spacing: {
    xs: 'var(--spacing-xs)',
    sm: 'var(--spacing-sm)',
  },
  radius: {
    md: 'var(--radius-md)',
  },
} as const;
```

### 2. Component Pattern: tailwind-variants for All

**Decision:** All Phase 1 components use `tailwind-variants` (tv) to define variants, sizes, and other props. No inline Tailwind class strings in JSX/templates.

**Why:** `tailwind-variants` creates a single source of truth for each component's styling. Variants are discoverable, documented, and reusable. Matches Button's existing pattern. Enables JSDoc documentation of available variants.

**What changes:**
- Input: currently hardcodes inline classes → refactor to use tv with variants
- Label, badge, toggle, switch: refactor to use tv if not already
- Custom drawer: build from scratch using tv
- Variants should include: base styles, variant modes (primary, secondary, etc.), sizes, states (disabled, aria-invalid, etc.)

**Example shape:**
```typescript
export const componentVariants = tv({
  base: '...',
  variants: {
    variant: { primary: '...', secondary: '...' },
    size: { sm: '...', md: '...', lg: '...' },
  },
  defaultVariants: { variant: 'primary', size: 'md' },
});
```

### 3. JSDoc Documentation for Components

**Decision:** Every component has comprehensive JSDoc on its Props type. Variants, sizes, defaults, and usage examples are documented in code.

**Why:** JSDoc travels with the code, stays in sync, and is readable by AI. Developers (and Claude) can understand a component's full API without reading implementation.

**What changes:**
- Add JSDoc blocks to all Phase 1 component Props types
- Document: available variants, sizes, disabled/invalid states, slot behavior, accessibility attrs
- Include `@example` tags showing common usage patterns

**Example:**
```typescript
/**
 * Button component for primary actions.
 * 
 * @example
 * <Button variant="primary" size="lg">Click me</Button>
 * 
 * @example
 * <Button variant="secondary" disabled>Disabled</Button>
 */
export interface ButtonProps {
  /** Visual style: default, outline, secondary, ghost, destructive, link */
  variant?: ButtonVariant;
  /** Size: xs, sm, default, lg, icon, icon-xs, icon-sm, icon-lg */
  size?: ButtonSize;
  /** Disabled state */
  disabled?: boolean;
}
```

### 4. Component Documentation in CONTEXT.md

**Decision:** Add a "## Components" section to CONTEXT.md listing all Phase 1 components with their primary variants and use cases.

**Why:** CONTEXT.md is already the project's domain glossary. A simple components table keeps design system knowledge in one discoverable place.

**What changes:**
- New section in CONTEXT.md: "## Components & Design System"
- Table format: Component Name | Primary Variants | Use Case
- Links to JSDoc in component files (optional, for deep dives)

**Example:**
```markdown
## Components & Design System

| Component | Variants | Use |
|-----------|----------|-----|
| Button | primary, outline, secondary, ghost, destructive, link | Primary CTAs, secondary actions, destructive actions |
| Input | — (text, password, email, etc.) | Form text inputs |
| Label | — | Form labels with optional required indicator |
| Badge | — | Status badges, tags, pills |
| Toggle | — | Single toggle switch |
| Switch | — | Binary on/off control |
| Drawer | — | Side panel, modal drawer |
```

### 5. Phase 1 Component List (Refactor + New)

**Decision:** Prioritize 6 existing components for refactoring + 1 new custom component.

**Refactor (apply tv pattern + JSDoc):**
- Button (already follows pattern, enhance JSDoc)
- Input (hardcoded classes → tv pattern)
- Label (apply tv if needed)
- Badge (apply tv if needed)
- Toggle (apply tv if needed)
- Switch (apply tv if needed)

**Build new (Bits UI base, custom wrapper):**
- Custom Drawer (inspired by shadcn, full ownership)

**Why these?** Core building blocks; heavy use across the app; high duplication; most friction point for AI composition.

### 6. Custom Drawer Scope

**Decision:** Build a custom Drawer component from scratch, inspired by shadcn's drawer architecture, built on Bits UI primitives. Full ownership, no external drawer dependency.

**Why:** Autonomy over design, full control of variants and behavior, no import bottleneck, aligns with design system direction.

**What includes:**
- Bits UI Dialog/Drawer primitives as foundation
- Drawer container, trigger, content, header, footer, close button
- Slide-in animation (left/right configurable)
- Backdrop overlay with configurable opacity
- Accessibility (focus trapping, escape to close)
- Variants: drawer position (left, right), size (sm, md, lg)

**Not included:** Complex drawer composition patterns, animation library integration beyond CSS transitions.

### 7. Incremental Migration Strategy

**Decision:** Phase 1 components are refactored immediately. Remaining components (calendar, card, collapsible, dialog, pill-toggle-group, etc.) migrate to this pattern as they are touched.

**Why:** Reduces big-bang risk, allows pattern refinement on high-value components first, lets old code age until it needs updates.

**What changes:**
- Phase 1: 7 components done (6 refactored + 1 new)
- Phase 2+: Calendar, Card, Collapsible, Dialog, etc., migrated incrementally as work touches them

---

## Testing Decisions

### What Makes a Good Test

- Tests external behavior (props in, rendered output and behavior out), not implementation details
- Does not mock Bits UI primitives or `tailwind-variants` (test the actual composition)
- For components: visual regression tests or integration tests that exercise all variant combinations
- Prior art: Component tests in this repo likely follow existing patterns (check `/web/src/lib/components/`)

### Testing Scope

**Components to test (Phase 1):**
- Button: all variants, sizes, disabled state, href vs button element
- Input: text, password, email, file types; disabled, invalid states; focus behavior
- Label: with and without required indicator
- Badge: (if it has variants)
- Toggle: on/off state, disabled
- Switch: on/off state, disabled
- Custom Drawer: open/close, animation, escape to close, focus trap

**Seams:**
- Single seam: component integration tests (render component with props, assert output)
- No need for multiple seams; if tv or Bits UI changes, unit tests catch it

**Prior Art:**
- Look at existing component test patterns in `/web/src/lib/components/` (if any exist)
- Follow the same testing style/patterns

---

## Out of Scope

- Figma integration or design tokens export to Figma
- Storybook or component playground (can be added later)
- Typography component (font scales, heading styles) — this is Phase 2+
- Animation library integration beyond CSS transitions
- Custom form validation layer (use HTML5 validation + `aria-invalid`)
- Accessibility audit beyond standard WCAG patterns (a11y review is separate)
- Refactoring of non-Phase 1 components (calendar, card, collapsible, etc.)
- Design token versioning or semantic versioning strategy (can be added later)

---

## Further Notes

### Token Completeness

Verify that `app.css @theme` includes all necessary token categories before Phase 1 implementation:
- Colors: primary, destructive, semantic colors (positive, etc.)
- Spacing: scale (xs, sm, md, lg, etc.)
- Radius: scale (sm, md, lg, etc.)
- Fonts: sans, serif, mono families
- Shadows: scale (sm, md, lg, etc.)
- Animations: shake, ptr-fade, any others in use

If gaps exist, add them to `app.css` before component refactoring begins.

### AI Documentation Goal

The JSDoc + CONTEXT.md pattern is specifically designed so that Claude and other AI assistants can:
1. Discover components and their variants by reading JSDoc
2. Reference design tokens in `design-tokens.ts` module
3. Generate correct component compositions without guessing
4. Understand use cases from CONTEXT.md

### Drawer Custom Implementation

The custom drawer is a deliberate choice to own the component fully. Before building, consider:
- Shadcn drawer is a good reference, but adapt to OpenPadel's specific needs
- Ensure it composes well with other Phase 1 components
- Test keyboard navigation (Tab, Escape) and screen reader behavior

### Future Phases

Phase 2 could include:
- Refactor remaining components (calendar, card, dialog, etc.)
- Add form composition patterns (FormField, FormControl)
- Add typography system (headings, body text scales)
- Add layout utilities (stack, grid, etc.)
- Consider adding Storybook for component documentation and isolated development
