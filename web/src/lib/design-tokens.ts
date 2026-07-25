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

		signal: 'var(--color-signal)',
		signalStrong: 'var(--color-signal-strong)',
		signalMuted: 'var(--color-signal-muted)',
		signalForeground: 'var(--color-signal-foreground)',

		tile: 'var(--color-tile)',
		tileForeground: 'var(--color-tile-foreground)',
		tileBorder: 'var(--color-tile-border)',
		tileCobalt: 'var(--color-tile-cobalt)',
		tileCobaltForeground: 'var(--color-tile-cobalt-foreground)',

		textPrimary: 'var(--color-text-primary)',
		textSecondary: 'var(--color-text-secondary)',
		textDisabled: 'var(--color-text-disabled)',

		positive: 'var(--color-positive)',
		destructive: 'var(--color-destructive)',
		destructiveForeground: 'var(--color-destructive-foreground)',

		warning: 'var(--color-warning)',
		warningMuted: 'var(--color-warning-muted)',

		medalGold: 'var(--color-medal-gold)',
		medalSilver: 'var(--color-medal-silver)',
		medalBronze: 'var(--color-medal-bronze)',

		avatarCobalt: 'var(--color-avatar-cobalt)',
		avatarAzure: 'var(--color-avatar-azure)',
		avatarNavy: 'var(--color-avatar-navy)',
		avatarSteel: 'var(--color-avatar-steel)',
		avatarSlate: 'var(--color-avatar-slate)',
		avatarMidnight: 'var(--color-avatar-midnight)',

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
		0: 'var(--spacing-0)',
		1: 'var(--spacing-1)',
		2: 'var(--spacing-2)',
		3: 'var(--spacing-3)',
		4: 'var(--spacing-4)'
	},

	radius: {
		base: 'var(--radius)',
		md: 'var(--radius-md)',
		sm: 'var(--radius-sm)',
		lg: 'var(--radius-lg)',
		xl: 'var(--radius-xl)'
	},

	fonts: {
		sans: 'var(--font-sans)',
		display: 'var(--font-display)'
	},

	shadows: {
		sm: 'var(--shadow-sm)',
		md: 'var(--shadow-md)',
		lg: 'var(--shadow-lg)'
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
