/**
 * Design tokens exported as TypeScript constants.
 * These mirror the CSS tokens defined in app.css @theme directive.
 *
 * Use these for:
 * - Programmatic access to token values in components
 * - Type-safe references in tailwind-variants
 * - JSDoc documentation of available design values
 *
 * @example
 * import { tokens } from '$lib/design-tokens';
 *
 * // In tailwind-variants
 * const buttonVariants = tv({
 *   variants: {
 *     variant: {
 *       primary: `bg-[${tokens.colors.primary}]`,
 *     },
 *   },
 * });
 *
 * @example
 * // In JSDoc
 * /// Available colors: primary, destructive, secondary, etc.
 * /// See tokens.colors for full list
 */

export const tokens = {
  colors: {
    // Surfaces
    background: 'var(--color-background)',
    surface: 'var(--color-surface)',
    surfaceRaised: 'var(--color-surface-raised)',
    border: 'var(--color-border)',
    borderStrong: 'var(--color-border-strong)',

    // Primary — dark forest green
    primary: 'var(--color-primary)',
    primaryHover: 'var(--color-primary-hover)',
    primaryMuted: 'var(--color-primary-muted)',
    primaryForeground: 'var(--color-primary-foreground)',

    // Text
    textPrimary: 'var(--color-text-primary)',
    textSecondary: 'var(--color-text-secondary)',
    textDisabled: 'var(--color-text-disabled)',

    // Semantic
    positive: 'var(--color-positive)',
    destructive: 'var(--color-destructive)',
    destructiveForeground: 'var(--color-destructive-foreground)',

    // shadcn-svelte tokens (resolved hex values)
    foreground: 'var(--color-foreground)',
    card: 'var(--color-card)',
    cardForeground: 'var(--color-card-foreground)',
    popover: 'var(--color-popover)',
    popoverForeground: 'var(--color-popover-foreground)',
    secondary: 'var(--color-secondary)',
    secondaryForeground: 'var(--color-secondary-foreground)',
    muted: 'var(--color-muted)',
    mutedForeground: 'var(--color-muted-foreground)',
    accent: 'var(--color-accent)',
    accentForeground: 'var(--color-accent-foreground)',
    input: 'var(--color-input)',
    ring: 'var(--color-ring)',
  },

  radius: {
    md: 'var(--radius)',
  },

  fonts: {
    sans: 'var(--font-sans)',
  },

  animations: {
    shake: 'var(--animate-shake)',
    ptrFade: 'var(--animate-ptr-fade)',
  },
} as const;

/**
 * Type-safe access to token keys for advanced use cases.
 */
export type TokenColor = keyof typeof tokens.colors;
export type TokenRadius = keyof typeof tokens.radius;
export type TokenFont = keyof typeof tokens.fonts;
export type TokenAnimation = keyof typeof tokens.animations;
