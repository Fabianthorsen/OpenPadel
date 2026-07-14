import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import Fixture from './avatar.fixture.svelte';

describe('Avatar', () => {
	it('shows initials when no icon is set', () => {
		render(Fixture);
		expect(screen.getByText('AL')).toBeInTheDocument();
	});

	it('renders a corner badge when the badge snippet is provided', () => {
		render(Fixture, { withBadge: true });
		expect(screen.getByTestId('avatar-badge')).toBeInTheDocument();
		// avatar content still present alongside the badge
		expect(screen.getByText('AL')).toBeInTheDocument();
	});
});
