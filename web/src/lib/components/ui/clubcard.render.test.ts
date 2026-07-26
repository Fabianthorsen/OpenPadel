import { describe, it, expect, vi, beforeAll } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import { fireEvent } from '@testing-library/dom';
import { init, register, waitLocale } from 'svelte-i18n';
import Fixture from './clubcard.fixture.svelte';

// The roster line ("18 members · Admin") renders through svelte-i18n, so the
// store needs a locale before the component mounts.
beforeAll(async () => {
	register('en', () => import('../../i18n/en.json'));
	init({ fallbackLocale: 'en', initialLocale: 'en' });
	await waitLocale('en');
});

describe('ClubCard (render)', () => {
	it('renders the club name', () => {
		render(Fixture);
		expect(screen.getByText('Bouvet Padel')).toBeInTheDocument();
	});

	it('renders the roster count and translated role', () => {
		render(Fixture);
		// The line is stitched from several i18n fragments, so match the whole
		// element's normalised text rather than a single text node.
		expect(
			screen.getByText(
				(_, el) =>
					el?.tagName === 'P' &&
					el.textContent?.replace(/\s+/g, ' ').trim() === '18 members · Admin'
			)
		).toBeInTheDocument();
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
