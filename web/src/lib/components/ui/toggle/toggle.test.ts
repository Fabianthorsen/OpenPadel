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

		it('produces a distinct class string for each variant', () => {
			const variants = ['default', 'outline'] as const;
			const outputs = variants.map((v) => toggleVariants({ variant: v }));
			expect(new Set(outputs).size).toBe(variants.length);
		});

		it('produces a distinct class string for each size', () => {
			const sizes = ['sm', 'default', 'lg'] as const;
			const outputs = sizes.map((s) => toggleVariants({ size: s }));
			expect(new Set(outputs).size).toBe(sizes.length);
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
