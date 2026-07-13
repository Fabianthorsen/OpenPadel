import { describe, it, expect } from 'vitest';
import { labelVariants } from './label.svelte';

describe('Label Component', () => {
	describe('labelVariants', () => {
		it('applies base label styles', () => {
			const classes = labelVariants();
			expect(classes).toContain('flex');
			expect(classes).toContain('items-center');
			expect(classes).toContain('text-sm');
			expect(classes).toContain('font-medium');
		});

		it('dims and disables pointer events when the associated control is disabled', () => {
			const classes = labelVariants();
			expect(classes).toContain('group-data-[disabled=true]:pointer-events-none');
			expect(classes).toContain('group-data-[disabled=true]:opacity-50');
			expect(classes).toContain('peer-disabled:cursor-not-allowed');
		});

		it('prevents text selection so tapping the label toggles the control', () => {
			expect(labelVariants()).toContain('select-none');
		});
	});
});
