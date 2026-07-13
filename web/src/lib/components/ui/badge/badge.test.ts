import { describe, it, expect } from 'vitest';
import { badgeVariants } from './badge.svelte';

describe('Badge Component', () => {
	describe('badgeVariants', () => {
		it('applies base badge styles', () => {
			const classes = badgeVariants();
			expect(classes).toContain('inline-flex');
			expect(classes).toContain('rounded-4xl');
			expect(classes).toContain('gap-1');
		});

		it('applies all 6 variants', () => {
			const variants = ['default', 'secondary', 'destructive', 'outline', 'ghost', 'link'] as const;
			variants.forEach((v) => {
				const classes = badgeVariants({ variant: v });
				expect(classes).toBeTruthy();
				expect(classes.length).toBeGreaterThan(0);
			});
		});

		it('applies default variant (status)', () => {
			const classes = badgeVariants({ variant: 'default' });
			expect(classes).toContain('bg-primary');
			expect(classes).toContain('text-primary-foreground');
		});

		it('applies secondary variant (tags)', () => {
			const classes = badgeVariants({ variant: 'secondary' });
			expect(classes).toContain('bg-secondary');
		});

		it('applies destructive variant (errors)', () => {
			const classes = badgeVariants({ variant: 'destructive' });
			expect(classes).toContain('bg-destructive');
		});

		it('applies outline variant (neutral)', () => {
			const classes = badgeVariants({ variant: 'outline' });
			expect(classes).toContain('border-border');
		});

		it('applies ghost variant (minimal)', () => {
			const classes = badgeVariants({ variant: 'ghost' });
			expect(classes).toContain('hover:bg-muted');
		});

		it('applies link variant (clickable)', () => {
			const classes = badgeVariants({ variant: 'link' });
			expect(classes).toContain('text-primary');
			expect(classes).toContain('underline');
		});

		it('applies default variant when none specified', () => {
			const explicit = badgeVariants({ variant: 'default' });
			const implicit = badgeVariants();
			expect(explicit).toBe(implicit);
		});
	});
});
