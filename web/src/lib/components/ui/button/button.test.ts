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

		it('gives each variant its signature class', () => {
			expect(buttonVariants({ variant: 'default' })).toContain('bg-primary');
			expect(buttonVariants({ variant: 'outline' })).toContain('bg-background');
			expect(buttonVariants({ variant: 'secondary' })).toContain('bg-secondary');
			expect(buttonVariants({ variant: 'ghost' })).toContain('hover:bg-muted');
			expect(buttonVariants({ variant: 'destructive' })).toContain('text-destructive');
			expect(buttonVariants({ variant: 'link' })).toContain('underline');
		});

		it('produces a distinct class string for each of the 6 variants', () => {
			const variants = ['default', 'outline', 'secondary', 'ghost', 'destructive', 'link'] as const;
			const outputs = variants.map((v) => buttonVariants({ variant: v }));
			expect(new Set(outputs).size).toBe(variants.length);
		});

		it('gives each size its signature dimension', () => {
			expect(buttonVariants({ size: 'xs' })).toContain('h-6');
			expect(buttonVariants({ size: 'sm' })).toContain('h-7');
			expect(buttonVariants({ size: 'default' })).toContain('h-8');
			expect(buttonVariants({ size: 'lg' })).toContain('h-9');
			expect(buttonVariants({ size: 'icon' })).toContain('size-8');
			expect(buttonVariants({ size: 'icon-xs' })).toContain('size-6');
			expect(buttonVariants({ size: 'icon-sm' })).toContain('size-7');
			expect(buttonVariants({ size: 'icon-lg' })).toContain('size-9');
		});

		it('produces a distinct class string for each of the 8 sizes', () => {
			const sizes = ['xs', 'sm', 'default', 'lg', 'icon', 'icon-xs', 'icon-sm', 'icon-lg'] as const;
			const outputs = sizes.map((s) => buttonVariants({ size: s }));
			expect(new Set(outputs).size).toBe(sizes.length);
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
