import { describe, it, expect } from 'vitest';
import { spinnerVariants } from './spinner.svelte';

describe('Spinner', () => {
	it('applies base spin styles', () => {
		const c = spinnerVariants();
		expect(c).toContain('animate-spin');
		expect(c).toContain('rounded-full');
	});

	it('gives each size its dimension', () => {
		expect(spinnerVariants({ size: 'sm' })).toContain('size-4');
		expect(spinnerVariants({ size: 'md' })).toContain('size-7');
		expect(spinnerVariants({ size: 'lg' })).toContain('size-9');
	});

	it('defaults to md', () => {
		expect(spinnerVariants()).toContain('size-7');
	});
});
