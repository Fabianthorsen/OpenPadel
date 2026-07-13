/** Design tokens mirroring CSS @theme directive from app.css for programmatic access. */

export const tokens = {
	colors: {
		background: 'var(--color-background)',
		surface: 'var(--color-surface)',
		surfaceRaised: 'var(--color-surface-raised)',
		border: 'var(--color-border)',
		borderStrong: 'var(--color-border-strong)',

		primary: 'var(--color-primary)',
		primaryHover: 'var(--color-primary-hover)',
		primaryMuted: 'var(--color-primary-muted)',
		primaryForeground: 'var(--color-primary-foreground)',

		textPrimary: 'var(--color-text-primary)',
		textSecondary: 'var(--color-text-secondary)',
		textDisabled: 'var(--color-text-disabled)',

		positive: 'var(--color-positive)',
		destructive: 'var(--color-destructive)',
		destructiveForeground: 'var(--color-destructive-foreground)',

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
		ring: 'var(--color-ring)'
	},

	spacing: {
		// TODO: add spacing tokens to app.css @theme when defined
	},

	radius: {
		md: 'var(--radius)'
	},

	fonts: {
		sans: 'var(--font-sans)'
	},

	shadows: {
		// TODO: add shadow tokens to app.css @theme when defined
	},

	animations: {
		shake: 'var(--animate-shake)',
		ptrFade: 'var(--animate-ptr-fade)'
	}
} as const;

export type TokenColor = keyof typeof tokens.colors;
export type TokenSpacing = keyof typeof tokens.spacing;
export type TokenRadius = keyof typeof tokens.radius;
export type TokenFont = keyof typeof tokens.fonts;
export type TokenShadow = keyof typeof tokens.shadows;
export type TokenAnimation = keyof typeof tokens.animations;
