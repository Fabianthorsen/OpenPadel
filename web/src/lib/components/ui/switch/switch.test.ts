import { describe, it, expect } from 'vitest';
import { switchVariants } from './switch.svelte';

describe('Switch Component', () => {
	describe('switchVariants', () => {
		it('applies base switch styles', () => {
			const classes = switchVariants();
			expect(classes).toContain('relative');
			expect(classes).toContain('inline-flex');
			expect(classes).toContain('rounded-full');
		});

		it('applies checked state background', () => {
			const classes = switchVariants();
			expect(classes).toContain('data-checked:bg-primary');
		});

		it('applies unchecked state background', () => {
			const classes = switchVariants();
			expect(classes).toContain('data-unchecked:bg-input');
		});

		it('applies focus styles', () => {
			const classes = switchVariants();
			expect(classes).toContain('focus-visible:border-ring');
			expect(classes).toContain('focus-visible:ring-ring/50');
		});

		it('applies disabled state styles', () => {
			const classes = switchVariants();
			expect(classes).toContain('data-disabled:cursor-not-allowed');
			expect(classes).toContain('data-disabled:opacity-50');
		});

		it('applies invalid state styles', () => {
			const classes = switchVariants();
			expect(classes).toContain('aria-invalid:border-destructive');
		});

		it('applies default size variant', () => {
			const classes = switchVariants({ size: 'default' });
			expect(classes).toContain('data-[size=default]:h-[18.4px]');
			expect(classes).toContain('data-[size=default]:w-[32px]');
		});

		it('applies sm size variant', () => {
			const classes = switchVariants({ size: 'sm' });
			expect(classes).toContain('data-[size=sm]:h-[14px]');
			expect(classes).toContain('data-[size=sm]:w-[24px]');
		});

		it('applies default variant when size not specified', () => {
			const explicit = switchVariants({ size: 'default' });
			const implicit = switchVariants();
			expect(explicit).toBe(implicit);
		});

		it('applies transition for smooth animation', () => {
			const classes = switchVariants();
			expect(classes).toContain('transition-all');
		});
	});
});
