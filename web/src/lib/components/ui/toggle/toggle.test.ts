import { describe, it, expect } from 'vitest';
import { toggleVariants } from './toggle.svelte';

describe('Toggle Component', () => {
	describe('toggleVariants', () => {
		it('applies base toggle styles', () => {
			const classes = toggleVariants();
			expect(classes).toContain('inline-flex');
			expect(classes).toContain('items-center');
			expect(classes).toContain('justify-center');
		});

		it('applies pressed state styles', () => {
			const classes = toggleVariants();
			expect(classes).toContain('aria-pressed:bg-muted');
		});

		it('applies all variants', () => {
			const variants = ['default', 'outline'] as const;
			variants.forEach((v) => {
				const classes = toggleVariants({ variant: v });
				expect(classes).toBeTruthy();
			});
		});

		it('applies all sizes', () => {
			const sizes = ['sm', 'default', 'lg'] as const;
			sizes.forEach((s) => {
				const classes = toggleVariants({ size: s });
				expect(classes).toBeTruthy();
			});
		});

		it('applies disabled state styles', () => {
			const classes = toggleVariants();
			expect(classes).toContain('disabled:opacity-50');
			expect(classes).toContain('disabled:pointer-events-none');
		});

		it('applies invalid state styles', () => {
			const classes = toggleVariants();
			expect(classes).toContain('aria-invalid:border-destructive');
		});

		it('default variant has transparent background', () => {
			const classes = toggleVariants({ variant: 'default' });
			expect(classes).toContain('bg-transparent');
		});

		it('outline variant has border', () => {
			const classes = toggleVariants({ variant: 'outline' });
			expect(classes).toContain('border');
		});
	});
});
