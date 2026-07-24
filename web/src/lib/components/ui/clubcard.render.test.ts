import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import { fireEvent } from '@testing-library/dom';
import Fixture from './clubcard.fixture.svelte';

describe('ClubCard (render)', () => {
	it('renders the club name', () => {
		render(Fixture);
		expect(screen.getByText('Bouvet Padel')).toBeInTheDocument();
	});

	it('renders the roster count and role', () => {
		render(Fixture);
		expect(screen.getByText('18 members • admin')).toBeInTheDocument();
	});

	it('falls back to initials when no avatar icon is set', () => {
		render(Fixture);
		expect(screen.getByText('BP')).toBeInTheDocument();
	});

	it('fires onclick when the card is pressed', async () => {
		const onclick = vi.fn();
		render(Fixture, { onclick });
		await fireEvent.click(screen.getByRole('button'));
		expect(onclick).toHaveBeenCalledOnce();
	});
});
