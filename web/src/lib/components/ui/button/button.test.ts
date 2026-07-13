import { describe, it, expect } from 'vitest';
import { buttonVariants } from './button.svelte';

describe('Button Component', () => {
	describe('buttonVariants', () => {
		it('applies base styles', () => {
			const classes = buttonVariants();
			expect(classes).toContain('inline-flex');
			expect(classes).toContain('items-center');
			expect(classes).toContain('justify-center');
		});

		it('all variants produce non-empty output', () => {
			const variants = ['primary', 'outline', 'secondary', 'ghost', 'destructive', 'link'] as const;
			variants.forEach((v) => {
				const classes = buttonVariants({ variant: v });
				expect(classes.length).toBeGreaterThan(0);
			});
		});

		it('all sizes produce non-empty output', () => {
			const sizes = ['xs', 'sm', 'default', 'lg', 'icon', 'icon-xs', 'icon-sm', 'icon-lg'] as const;
			sizes.forEach((s) => {
				const classes = buttonVariants({ size: s });
				expect(classes.length).toBeGreaterThan(0);
			});
		});

		it('applies disabled state styles', () => {
			const classes = buttonVariants();
			expect(classes).toContain('disabled:opacity-50');
			expect(classes).toContain('disabled:pointer-events-none');
		});

		it('applies focus and transition styles', () => {
			const classes = buttonVariants();
			expect(classes).toContain('focus-visible:ring-3');
			expect(classes).toContain('transition-all');
		});
	});
});
