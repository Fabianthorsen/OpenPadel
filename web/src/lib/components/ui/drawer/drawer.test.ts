import { describe, it, expect } from 'vitest';
import { drawerContentVariants } from './drawer-content.svelte';

describe('Drawer Component', () => {
	describe('drawerContentVariants', () => {
		it('applies base drawer styles', () => {
			const classes = drawerContentVariants();
			expect(classes).toContain('fixed');
			expect(classes).toContain('inset-x-0');
			expect(classes).toContain('bottom-0');
		});

		it('produces a distinct class string for each size', () => {
			const sizes = ['sm', 'md', 'lg'] as const;
			const outputs = sizes.map((s) => drawerContentVariants({ size: s }));
			expect(new Set(outputs).size).toBe(sizes.length);
		});

		it('applies sm size constraints', () => {
			const classes = drawerContentVariants({ size: 'sm' });
			expect(classes).toContain('max-h-[40vh]');
		});

		it('applies md size constraints (default)', () => {
			const classes = drawerContentVariants({ size: 'md' });
			expect(classes).toContain('max-h-[60vh]');
		});

		it('applies lg size constraints', () => {
			const classes = drawerContentVariants({ size: 'lg' });
			expect(classes).toContain('max-h-[80vh]');
		});

		it('applies default size when none specified', () => {
			const explicit = drawerContentVariants({ size: 'md' });
			const implicit = drawerContentVariants();
			expect(explicit).toBe(implicit);
		});

		it('includes z-index for stacking', () => {
			const classes = drawerContentVariants();
			expect(classes).toContain('z-50');
		});

		it('includes flex layout for content arrangement', () => {
			const classes = drawerContentVariants();
			expect(classes).toContain('flex');
			expect(classes).toContain('flex-col');
		});

		it('includes responsive padding', () => {
			const classes = drawerContentVariants();
			expect(classes).toContain('p-4');
		});
	});
});
