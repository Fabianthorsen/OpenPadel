import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import StatTile from './stat-tile.svelte';

/**
 * StatTile is pure presentation: it renders a pre-formatted value over a
 * pre-translated label, with no catalog or i18n dependency. These lock in that
 * it shows exactly what it's given and that `accent` only tints the value.
 */
describe('StatTile (render)', () => {
	it('renders the value and label verbatim', () => {
		render(StatTile, { value: '12–3–5', label: 'Record' });
		expect(screen.getByText('12–3–5')).toBeInTheDocument();
		expect(screen.getByText('Record')).toBeInTheDocument();
	});

	it('tints the value in the brand colour when accent is set', () => {
		render(StatTile, { value: '56%', label: 'Point win %', accent: true });
		expect(screen.getByText('56%')).toHaveClass('text-primary');
	});

	it('does not tint the value by default', () => {
		render(StatTile, { value: '8', label: 'Games' });
		expect(screen.getByText('8')).not.toHaveClass('text-primary');
	});
});
