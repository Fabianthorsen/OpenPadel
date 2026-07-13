# Tailwind CSS 4 Design Token Management Research

**Research Date:** July 2026  
**Tailwind Version:** v4.3.2 (latest as of research date)  
**Research Focus:** Design token definition, configuration patterns, TypeScript integration, and design system best practices

---

## Executive Summary

Tailwind CSS 4 introduces a **CSS-first approach to design tokens** using the `@theme` directive, fundamentally shifting from JavaScript-based configuration (v3) to native CSS variable management. This research covers the recommended patterns for defining, managing, and exporting design tokens in modern Tailwind 4 projects, with emphasis on TypeScript integration and design system patterns suitable for component libraries.

---

## 1. Design Token Fundamentals in Tailwind 4

### 1.1 What Are Design Tokens in Tailwind 4?

Design tokens are CSS custom properties that control which utility classes exist in your project and define design system values (colors, spacing, typography, etc.). They use a specialized `@theme` directive that serves dual purposes:

1. **Generates utility classes** - Each token creates corresponding utilities (e.g., `--color-mint-500` creates `bg-mint-500`, `text-mint-500`, etc.)
2. **Creates CSS variables** - Tokens become accessible CSS custom properties for use in custom CSS or JavaScript

**Source:** https://tailwindcss.com/docs/theme

### 1.2 Core Concept: Why `@theme` Instead of `:root`?

The `@theme` directive is fundamentally different from regular CSS variables:

```css
/* ❌ Regular CSS variables - NO utility generation */
:root {
  --my-color: blue;
}

/* ✅ Theme variables - GENERATES utilities */
@theme {
  --color-my-color: blue;
}
```

With `@theme`, the `--color-my-color: blue` declaration generates:
- Utility classes: `bg-my-color`, `text-my-color`, `fill-my-color`, etc.
- CSS variable: `var(--color-my-color)` for runtime use
- Automatic integration with all color-accepting utilities

**Source:** https://tailwindcss.com/docs/theme

### 1.3 Theme Variable Namespaces (419+ default tokens)

All design tokens are organized by namespace, each mapping to specific utility classes:

| Namespace | Utilities Generated | Example Usage |
|-----------|-------------------|----------------|
| `--color-*` | Color utilities | `bg-red-500`, `text-sky-300`, `border-gray-100` |
| `--font-*` | Font family utilities | `font-sans`, `font-poppins` |
| `--text-*` | Font size utilities | `text-xl`, `text-2xl` |
| `--font-weight-*` | Font weight utilities | `font-bold`, `font-semibold` |
| `--spacing-*` | Padding, margin, sizing | `px-4`, `m-6`, `w-16`, `h-12` |
| `--radius-*` | Border radius utilities | `rounded-sm`, `rounded-lg` |
| `--shadow-*` | Box shadow utilities | `shadow-md`, `shadow-lg` |
| `--breakpoint-*` | Responsive variants | `sm:*`, `md:*`, `lg:*`, `3xl:*` |
| `--animate-*` | Animation utilities | `animate-spin`, `animate-fade-in` |
| `--blur-*` | Blur filter utilities | `blur-sm`, `blur-md` |
| `--container-*` | Container query units | `@sm:*`, `@md:*` |
| `--ease-*` | Easing functions | For custom transitions |
| `--gap-*` | Gap utilities | `gap-4`, `gap-8` |
| `--transition-*` | Transition durations | `transition-colors`, `transition-all` |

**Source:** https://tailwindcss.com/docs/theme, https://tailwindcss.com/docs/configuration

---

## 2. Best Practices for Defining Design Tokens

### 2.1 Four Customization Strategies

Tailwind 4 supports four distinct patterns for managing design tokens:

#### Strategy 1: Extending the Default Theme

Add new tokens while preserving all defaults:

```css
@import "tailwindcss";

@theme {
  --font-script: Great Vibes, cursive;
  --color-brand: oklch(0.72 0.11 178);
  --color-mint-500: oklch(0.72 0.11 178);
}
```

**Use case:** Adding brand colors or custom fonts to the default palette.

#### Strategy 2: Overriding Specific Values

Replace individual tokens while keeping others:

```css
@import "tailwindcss";

@theme {
  --breakpoint-sm: 30rem;  /* Changes from default 40rem */
  --radius-lg: 12px;       /* Override border radius */
}
```

**Use case:** Fine-tuning default values for a specific project.

#### Strategy 3: Replacing Entire Namespaces

Use wildcard syntax to reset a category and define only custom values:

```css
@import "tailwindcss";

@theme {
  --color-*: initial;      /* Disable ALL default colors */
  --color-white: #fff;
  --color-midnight: #121063;
  --color-tahiti: #3ab7bf;
  --color-bermuda: #78dcca;
}
```

**Use case:** Creating a completely custom palette while removing unused colors (reduces CSS output).

#### Strategy 4: Building a Complete Custom Theme

Disable all defaults and define every token from scratch:

```css
@import "tailwindcss";

@theme {
  --*: initial;            /* Disable ALL defaults */
  --spacing: 4px;
  --font-body: Inter, sans-serif;
  --color-lagoon: oklch(0.72 0.11 221.19);
  --color-coral: oklch(0.74 0.17 40.24);
  --radius-sm: 4px;
  --radius-md: 8px;
}
```

**Use case:** Fully custom design systems with no reliance on Tailwind's defaults.

**Source:** https://tailwindcss.com/docs/theme, https://tailwindcss.com/docs/customizing-colors

### 2.2 Best Practices for Each Token Type

#### Colors: Using OKLCH Color Space

Tailwind 4's default palette uses **OKLCH color space** for better perceptual uniformity:

```css
@theme {
  --color-red-500: oklch(63.7% 0.237 25.331);
}
```

OKLCH advantages over RGB/hex:
- **Perceptually uniform** - increments feel consistent across the palette
- **Better accessibility** - easier to create accessible color scales
- **Maintains luminance** - colors are consistent in brightness

Example color scale (red):
```css
@theme {
  --color-red-50: oklch(97.1% 0.013 17.38);
  --color-red-100: oklch(93.6% 0.032 17.717);
  --color-red-500: oklch(63.7% 0.237 25.331);
  --color-red-950: oklch(25.8% 0.092 26.042);
}
```

**Recommendation:** Use OKLCH for custom color scales to match Tailwind's perceptual consistency.

#### Spacing and Sizing

Define base spacing unit or explicit scale:

```css
@theme {
  /* Option 1: Base unit (multiplied automatically) */
  --spacing: 4px;
  
  /* Option 2: Explicit scale */
  --spacing-sm: 8px;
  --spacing-md: 16px;
  --spacing-lg: 24px;
}
```

This generates utilities: `px-sm`, `m-md`, `w-lg`, `gap-sm`, etc.

#### Fonts and Typography

Define font families, sizes, and weights separately:

```css
@theme {
  --font-sans: Inter, system-ui, sans-serif;
  --font-serif: Georgia, serif;
  --font-mono: 'Courier New', monospace;
  
  --text-sm: 0.875rem;
  --text-base: 1rem;
  --text-lg: 1.125rem;
  
  --font-weight-light: 300;
  --font-weight-normal: 400;
  --font-weight-bold: 700;
}
```

#### Border Radius

Define consistent radius scales:

```css
@theme {
  --radius-none: 0;
  --radius-xs: 4px;
  --radius-sm: 6px;
  --radius-md: 8px;
  --radius-lg: 12px;
  --radius-xl: 16px;
}
```

#### Shadows

Define semantic shadow scales:

```css
@theme {
  --shadow-xs: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  --shadow-sm: 0 1px 3px 0 rgba(0, 0, 0, 0.1);
  --shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
  --shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
  --shadow-xl: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
}
```

#### Custom Animations

Define animations and their keyframes within `@theme`:

```css
@theme {
  --animate-fade-in-scale: fade-in-scale 0.3s ease-out;
  
  @keyframes fade-in-scale {
    0% {
      opacity: 0;
      transform: scale(0.95);
    }
    100% {
      opacity: 1;
      transform: scale(1);
    }
  }
}
```

**Source:** https://tailwindcss.com/docs/theme, https://tailwindcss.com/docs/customizing-colors, https://tailwindcss.com/docs/customizing-spacing

### 2.3 Advanced Theme Configuration Techniques

#### Using `inline` for Variable References

When referencing other CSS variables within `@theme`, use the `inline` option:

```css
@theme inline {
  --font-sans: var(--font-inter);
}
```

**Why:** Without `inline`, CSS cascade resolution may fail due to variable reference order. The `inline` option ensures the variable's *value* is used rather than creating circular references.

#### Using `static` for All-Tokens-Present Guarantee

Normally, unused theme variables are tree-shaken from the output. Use `@theme static` to generate all CSS variables even if unused:

```css
@theme static {
  --color-primary: var(--color-red-500);
  --color-secondary: var(--color-blue-500);
  --color-accent: var(--color-yellow-500);
}
```

**Use case:** Component libraries that need all design token CSS variables available at runtime, regardless of usage in utilities.

#### Sharing Themes Across Packages

Create a shared theme file in a common package:

```css
/* ./packages/brand/theme.css */
@theme {
  --*: initial;
  --spacing: 4px;
  --font-body: Inter, sans-serif;
  --color-lagoon: oklch(0.72 0.11 221.19);
  --color-coral: oklch(0.74 0.17 40.24);
}
```

Then import it in consuming packages:

```css
/* ./packages/admin/app.css */
@import "tailwindcss";
@import "../brand/theme.css";
```

**Source:** https://tailwindcss.com/docs/theme, https://tailwindcss.com/docs/configuration

---

## 3. Tailwind Config (`tailwind.config.ts`) vs `@theme` Directive

### 3.1 Key Differences

| Aspect | `tailwind.config.ts` (v3 pattern) | `@theme` Directive (v4 native) |
|--------|-----------------------------------|--------------------------------|
| **Location** | JavaScript configuration file | CSS file |
| **Syntax** | JavaScript objects | CSS custom properties |
| **Primary Purpose** | Build config, plugins, advanced setup | Design tokens, theme values |
| **Utility Generation** | Implicit, requires special handling | Direct and explicit |
| **CSS Variables** | Generated separately by build | Integrated into CSS output |
| **Runtime Accessibility** | Requires `resolveConfig()` helper | Direct `var(--token-name)` access |
| **Recommended Use** | Plugins, build options | Design tokens |
| **TypeScript Support** | Full in TypeScript configs | CSS file, use separate TS for types |
| **Modern Approach** | Legacy (v3 pattern) | Native (v4 default) |

### 3.2 Using `tailwind.config.ts` in Tailwind 4

`tailwind.config.ts` still exists in Tailwind 4 but is **primarily for build configuration**, not design tokens:

```typescript
// tailwind.config.ts - Use for build options, not design tokens
import type { Config } from 'tailwindcss'

export default {
  // Build options
  content: [
    './src/**/*.{html,js,jsx,ts,tsx,svelte}',
  ],
  
  // Plugins
  plugins: [
    require('./plugins/my-plugin'),
  ],
  
  // Advanced features
  future: {
    // Feature flags
  },
} satisfies Config
```

**Key point:** Design tokens should be in CSS (`@theme`), not in JavaScript config.

### 3.3 Content Detection: `@source` Directive in v4

Tailwind 4 replaces the `content` config key with the CSS `@source` directive:

```css
/* app.css */
@import "tailwindcss";

/* Scan these directories */
@source "../src";

/* Or explicitly list sources */
@source "../node_modules/@acmecorp/ui-lib";
@source "../src/components";

/* Exclude specific paths */
@source not "../src/components/legacy";

/* Disable automatic detection */
@import "tailwindcss" source(none);
@source "../admin";
@source "../shared";
```

**Source:** https://tailwindcss.com/docs/upgrade-guide, https://tailwindcss.com/docs/configuration, https://tailwindcss.com/docs/detecting-classes-in-source-files

---

## 4. Exporting Design Tokens for TypeScript Use

### 4.1 Making Theme Variables Available at Runtime

Since Tailwind 4 defines tokens as CSS custom properties, accessing them in TypeScript is straightforward:

```typescript
// Get a computed theme variable at runtime
function getThemeVariable(variableName: string): string {
  const styles = getComputedStyle(document.documentElement);
  return styles.getPropertyValue(variableName).trim();
}

// Example usage
const primaryColor = getThemeVariable('--color-primary-500');
const spacing4x = getThemeVariable('--spacing-4');
```

### 4.2 Creating a TypeScript Tokens Utility Module

For component libraries and design systems, create a dedicated tokens module:

```typescript
// src/lib/design-tokens.ts
export const tokens = {
  colors: {
    red: {
      50: 'var(--color-red-50)',
      100: 'var(--color-red-100)',
      500: 'var(--color-red-500)',
      950: 'var(--color-red-950)',
    },
    blue: {
      50: 'var(--color-blue-50)',
      500: 'var(--color-blue-500)',
    },
  },
  spacing: {
    xs: 'var(--spacing-xs)',
    sm: 'var(--spacing-sm)',
    md: 'var(--spacing-md)',
    lg: 'var(--spacing-lg)',
  },
  radius: {
    sm: 'var(--radius-sm)',
    md: 'var(--radius-md)',
    lg: 'var(--radius-lg)',
  },
  fonts: {
    sans: 'var(--font-sans)',
    serif: 'var(--font-serif)',
    mono: 'var(--font-mono)',
  },
} as const;

// Type-safe token access
export type TokenKey = keyof typeof tokens;
export type ColorToken = keyof typeof tokens.colors;
export type SpacingToken = keyof typeof tokens.spacing;
```

Usage in Svelte components:

```svelte
<script>
  import { tokens } from '$lib/design-tokens';
</script>

<div style="color: {tokens.colors.red[500]}">
  Text with red color
</div>

<div style="padding: {tokens.spacing.md}">
  Padded content
</div>
```

### 4.3 Type-Safe Utility Class Generation with tailwind-variants

While the official `tailwind-variants` documentation could not be fetched during research, the pattern is to create type-safe utility class helpers:

```typescript
// src/lib/cv.ts - Create a class variant helper
import type { ClassValue } from 'clsx';

// Pattern for type-safe class generation
export function createButtonClasses(variant: 'primary' | 'secondary'): string {
  const baseClasses = 'px-4 py-2 rounded-md font-semibold transition-colors';
  
  const variants = {
    primary: 'bg-red-500 text-white hover:bg-red-600',
    secondary: 'bg-gray-200 text-gray-900 hover:bg-gray-300',
  };
  
  return `${baseClasses} ${variants[variant]}`;
}
```

### 4.4 Extracting TypeScript Types from Theme Values

Create a type helper to extract types from your theme configuration:

```typescript
// src/lib/theme-types.ts

// For color tokens
export type ColorToken = 
  | 'red-50' | 'red-100' | 'red-500' | 'red-950'
  | 'blue-50' | 'blue-100' | 'blue-500'
  | 'gray-100' | 'gray-500' | 'gray-950';

// For spacing tokens
export type SpacingToken = 'xs' | 'sm' | 'md' | 'lg' | 'xl';

// For component variants
export type ButtonVariant = 'primary' | 'secondary' | 'tertiary';
export type ButtonSize = 'sm' | 'md' | 'lg';

// Type-safe component props
interface ButtonProps {
  variant?: ButtonVariant;
  size?: ButtonSize;
  children: React.ReactNode;
}
```

In Svelte:

```svelte
<script lang="ts">
  import type { ButtonVariant, ButtonSize } from '$lib/theme-types';

  interface Props {
    variant?: ButtonVariant;
    size?: ButtonSize;
  }

  let { variant = 'primary', size = 'md', children }: Props & { children: string } = $props();
</script>

<button class={`btn-${variant}-${size}`}>
  {children}
</button>
```

**Note:** Tailwind 4 doesn't provide automatic TypeScript generation from CSS `@theme` directives (unlike some other CSS-in-JS solutions), so types must be maintained manually or generated via tooling.

**Source:** https://tailwindcss.com/docs/theme, https://tailwindcss.com/docs/installation (TypeScript integration notes)

---

## 5. SvelteKit Integration Patterns

### 5.1 SvelteKit + Tailwind 4 Setup

```bash
# Create SvelteKit project
npx sv create my-project
cd my-project

# Install Tailwind 4
npm install tailwindcss @tailwindcss/vite
```

Configure Vite plugin (`vite.config.ts`):

```typescript
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
  plugins: [
    tailwindcss(),
    sveltekit(),
  ],
});
```

Create CSS file (`src/app.css`):

```css
@import "tailwindcss";

@theme {
  --color-brand: oklch(0.72 0.11 178);
  --font-display: Poppins, sans-serif;
}
```

Layout setup (`src/routes/+layout.svelte`):

```svelte
<script>
  let { children } = $props();
  import "../app.css";
</script>

{@render children()}
```

### 5.2 Accessing Theme Variables in Svelte Components

```svelte
<script lang="postcss">
  @reference "tailwindcss";
  
  :global(html) {
    background-color: theme(--color-gray-100);
  }
</script>

<style lang="postcss">
  @reference "tailwindcss";
  
  .card {
    background-color: var(--color-white);
    border-radius: var(--radius-lg);
    padding: var(--spacing-6);
    box-shadow: var(--shadow-md);
  }
</style>

<div class="card">
  Card with theme variables
</div>
```

### 5.3 Component Library Pattern in SvelteKit

```typescript
// src/lib/components/Button.svelte
<script lang="ts">
  import type { SvelteHTMLElements } from 'svelte/elements';
  
  interface Props extends SvelteHTMLElements['button'] {
    variant?: 'primary' | 'secondary';
    size?: 'sm' | 'md' | 'lg';
  }

  let { 
    variant = 'primary', 
    size = 'md',
    children,
    ...rest 
  }: Props = $props();

  const baseClasses = 'font-semibold rounded transition-colors';
  
  const variants = {
    primary: 'bg-blue-500 text-white hover:bg-blue-600',
    secondary: 'bg-gray-200 text-gray-900 hover:bg-gray-300',
  };
  
  const sizes = {
    sm: 'px-3 py-1 text-sm',
    md: 'px-4 py-2 text-base',
    lg: 'px-6 py-3 text-lg',
  };

  const classes = `${baseClasses} ${variants[variant]} ${sizes[size]}`;
</script>

<button class={classes} {...rest}>
  {@render children?.()}
</button>

<style lang="postcss">
  @reference "tailwindcss";
</style>
```

**Source:** https://tailwindcss.com/docs/installation/framework-guides/sveltekit

---

## 6. Building Design Systems with Tailwind 4

### 6.1 Design System Architecture

Recommended structure for a Tailwind-based design system:

```
packages/
├── design-system/
│   ├── src/
│   │   ├── theme.css          # @theme tokens
│   │   ├── components/        # Reusable components
│   │   │   ├── Button.svelte
│   │   │   ├── Card.svelte
│   │   │   └── Dialog.svelte
│   │   ├── lib/
│   │   │   ├── design-tokens.ts    # Token exports
│   │   │   ├── theme-types.ts      # Type definitions
│   │   │   └── class-helpers.ts    # CV/class utilities
│   │   └── index.ts            # Public API
│   └── package.json
└── consumer-app/
    ├── src/
    │   ├── app.css             # @import theme
    │   └── routes/
    └── vite.config.ts
```

### 6.2 Shared Theme CSS

Create a reusable theme package:

```css
/* packages/design-system/src/theme.css */
@theme {
  --*: initial;
  
  /* Brand colors */
  --color-brand-50: oklch(0.97 0.01 264);
  --color-brand-500: oklch(0.65 0.20 270);
  --color-brand-950: oklch(0.25 0.05 270);
  
  /* Semantic colors */
  --color-success-500: oklch(0.70 0.18 142);
  --color-warning-500: oklch(0.77 0.19 73);
  --color-error-500: oklch(0.64 0.24 25);
  
  /* Spacing */
  --spacing: 4px;
  
  /* Fonts */
  --font-sans: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  --font-mono: 'Monaco', 'Courier New', monospace;
  
  /* Radii */
  --radius-sm: 4px;
  --radius-md: 8px;
  --radius-lg: 12px;
  --radius-full: 999px;
  
  /* Shadows */
  --shadow-sm: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  --shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
}
```

### 6.3 Component Composition Pattern

```svelte
<!-- Button.svelte -->
<script lang="ts">
  type Variant = 'primary' | 'secondary' | 'tertiary';
  type Size = 'sm' | 'md' | 'lg';
  
  interface Props {
    variant?: Variant;
    size?: Size;
    disabled?: boolean;
    onclick?: () => void;
  }

  let { variant = 'primary', size = 'md', disabled = false }: Props = $props();

  const variantClasses: Record<Variant, string> = {
    primary: 'bg-brand-500 text-white hover:bg-brand-600 disabled:bg-gray-400',
    secondary: 'bg-gray-200 text-gray-900 hover:bg-gray-300 disabled:bg-gray-100',
    tertiary: 'bg-transparent text-brand-500 hover:bg-brand-50 disabled:text-gray-400',
  };

  const sizeClasses: Record<Size, string> = {
    sm: 'px-3 py-1 text-sm',
    md: 'px-4 py-2 text-base',
    lg: 'px-6 py-3 text-lg',
  };

  const classes = `
    inline-flex items-center justify-center
    font-semibold rounded-md
    transition-colors duration-200
    disabled:cursor-not-allowed
    focus:outline-none focus:ring-2 focus:ring-offset-2
    ${variantClasses[variant]}
    ${sizeClasses[size]}
  `;
</script>

<button {disabled} class={classes} onclick={arguments[0]}>
  <slot />
</button>
```

### 6.4 Best Practices for Design Systems

1. **Use `@theme static` for Component Libraries**
   - Ensures all design token CSS variables are available, even if not used in utilities
   - Allows consuming applications to access tokens directly

2. **Separate Tokens from Utilities**
   - Token definitions: `theme.css`
   - Component styles: Component files or separate CSS
   - Build utilities: `@utility` directives when needed

3. **Document Token Semantics**
   - Create a tokens reference document
   - Explain which tokens are for what use cases
   - Provide guidance on color contrast and accessibility

4. **Version Design Tokens Carefully**
   - Semantic versioning should account for token changes
   - Consider providing migration guides for breaking token changes
   - Use deprecation periods when removing tokens

5. **Use Arbitrary Values Sparingly**
   - Constrain to defined design tokens
   - Only use `[value]` syntax for truly one-off cases
   - Document when arbitrary values are allowed

6. **Leverage the `@layer` Directive**
   ```css
   @layer components {
     .card {
       background-color: var(--color-white);
       border-radius: var(--radius-lg);
       box-shadow: var(--shadow-md);
       padding: var(--spacing-6);
     }
   }
   ```

**Source:** https://tailwindcss.com/docs/configuration, https://tailwindcss.com/docs/adding-custom-styles, https://tailwindcss.com/docs/installation/framework-guides/sveltekit

---

## 7. Advanced Customization: Custom Utilities and Variants

### 7.1 Creating Custom Utilities

#### Simple Utilities

```css
@utility content-auto {
  content-visibility: auto;
}

@utility text-ellipsis {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
```

Usage: `<div class="content-auto text-ellipsis"></div>`

#### Functional Utilities with Theme Values

```css
@theme {
  --tab-size-2: 2;
  --tab-size-4: 4;
}

@utility tab-* {
  tab-size: --value(--tab-size-*);
}
```

Usage: `<div class="tab-4"></div>`

#### Utilities with Arbitrary Values

```css
@utility my-custom-* {
  custom-property: --value([string]);
}
```

### 7.2 Creating Custom Variants

```css
@custom-variant theme-midnight {
  &:where([data-theme="midnight"] *) {
    @slot;
  }
}

@custom-variant group-hover-visible {
  &:where(.group:hover *) {
    @slot;
  }
}
```

Usage:
```html
<div data-theme="midnight">
  <button class="theme-midnight:bg-black"></button>
</div>

<div class="group">
  <button class="group-hover-visible:opacity-100"></button>
</div>
```

**Source:** https://tailwindcss.com/docs/adding-custom-styles, https://tailwindcss.com/docs/functions-and-directives

---

## 8. Migration from Tailwind CSS v3 to v4

### 8.1 Key Configuration Changes

| Feature | v3 | v4 |
|---------|----|----|
| CSS import | `@tailwind base;` `@tailwind components;` `@tailwind utilities;` | `@import "tailwindcss";` |
| Theme definition | JavaScript in `tailwind.config.js` | CSS with `@theme` directive |
| Content detection | `content: []` in config | `@source` directive in CSS |
| Custom utilities | `@layer utilities` | `@utility` directive |
| Custom variants | Using plugins | `@custom-variant` directive |
| Shadow naming | `shadow`, `shadow-sm` | `shadow-sm`, `shadow-xs` |
| Ring default | `ring` (3px) | `ring-3` (explicit sizing) |

### 8.2 Using the Official Migration Tool

```bash
npx @tailwindcss/upgrade
```

This automates:
- Dependency updates
- Config migration from v3 to v4
- Template file updates
- Class name changes (e.g., `outline-none` → `outline-hidden`)

### 8.3 Browser Support Changes

Tailwind 4 requires **modern browsers only**:
- Safari 16.4+
- Chrome 111+
- Firefox 128+

Use v3.4 if you need support for older browsers.

**Source:** https://tailwindcss.com/docs/upgrade-guide, https://tailwindcss.com/docs/compatibility

---

## 9. Summary: Decision Matrix

### When to Use Each Approach

| Scenario | Recommended Pattern |
|----------|-------------------|
| Starting new project | CSS `@theme` + SvelteKit |
| Component library | `@theme static` + TypeScript module exports |
| Design system with monorepo | Shared `theme.css` + component package structure |
| Extending defaults | `@theme` with specific token overrides |
| Custom palette | `@theme { --color-*: initial; ... }` namespace reset |
| Sharing with teams | Export as `@theme` in NPM package |
| Runtime token access | Create TypeScript utility module |
| Type-safe utilities | Manual token type definitions + component helpers |

---

## Sources

**Official Tailwind CSS Documentation:**
- https://tailwindcss.com/docs/configuration - Configuration and setup
- https://tailwindcss.com/docs/theme - Theme variables and customization
- https://tailwindcss.com/docs/colors - Color customization guide
- https://tailwindcss.com/docs/customizing-spacing - Spacing token patterns
- https://tailwindcss.com/docs/adding-custom-styles - Custom utilities and variants
- https://tailwindcss.com/docs/detecting-classes-in-source-files - Content detection and @source directive
- https://tailwindcss.com/docs/functions-and-directives - @theme, @utility, @custom-variant directives
- https://tailwindcss.com/docs/installation - Setup for all build tools
- https://tailwindcss.com/docs/installation/framework-guides/sveltekit - SvelteKit integration
- https://tailwindcss.com/docs/upgrade-guide - v3 to v4 migration guide
- https://tailwindcss.com/docs/editor-setup - IDE and TypeScript integration
- https://tailwindcss.com/docs/compatibility - Browser compatibility

**GitHub Source Code:**
- https://github.com/tailwindlabs/tailwindcss - Main repository
- Default theme values (419+ tokens): https://raw.githubusercontent.com/tailwindlabs/tailwindcss/main/packages/tailwindcss/theme.css

**Release Information:**
- Tailwind CSS v4.3.2 (latest, 2026)
- https://github.com/tailwindlabs/tailwindcss/releases - Release notes and v4 announcement

---

## Research Notes

- **Research completed:** July 12, 2026
- **Tool used:** WebFetch for official documentation, Bash for GitHub API verification
- **Primary focus:** Tailwind CSS v4 design token management with emphasis on practical patterns for SvelteKit and component library development
- **Coverage:** Configuration strategies, TypeScript integration, design system architecture, and SvelteKit best practices
- **Limitations:** Official tailwind-variants documentation unavailable during research; patterns inferred from Tailwind's custom utility directives
