import { describe, it, expect } from 'vitest';
import { inputVariants } from './input.svelte';

describe('Input Component', () => {
	describe('inputVariants', () => {
		it('applies base input styles', () => {
			const classes = inputVariants();
			expect(classes).toContain('rounded-lg');
			expect(classes).toContain('border');
			expect(classes).toContain('bg-transparent');
		});

		it('applies focus styles', () => {
			const classes = inputVariants();
			expect(classes).toContain('focus-visible:border-ring');
			expect(classes).toContain('focus-visible:ring-ring/50');
		});

		it('applies disabled state styles', () => {
			const classes = inputVariants();
			expect(classes).toContain('disabled:opacity-50');
			expect(classes).toContain('disabled:cursor-not-allowed');
		});

		it('applies invalid state styles', () => {
			const classes = inputVariants();
			expect(classes).toContain('aria-invalid:border-destructive');
			expect(classes).toContain('aria-invalid:ring-destructive');
		});

		it('handles placeholder text color', () => {
			const classes = inputVariants();
			expect(classes).toContain('placeholder:text-muted-foreground');
		});

		it('handles file input styles', () => {
			const classes = inputVariants();
			expect(classes).toContain('file:text-foreground');
			expect(classes).toContain('file:bg-transparent');
		});
	});
});
