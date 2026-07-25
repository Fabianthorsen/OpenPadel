import { describe, it, expect, beforeAll, afterEach, beforeEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { init, register, waitLocale } from 'svelte-i18n';
import CreateDrawer from './CreateDrawer.svelte';

/**
 * Render tests for the optional "attach to a club" picker on the generic New
 * Tournament drawer (second creation flow).
 *
 * When opened WITHOUT a club preset, a member of ≥1 club can attach the game to
 * one of their clubs, turning it into a club event. Picking a club reveals the
 * "whole club will be notified" banner + club CTA and sends club_id on create;
 * the default "None — personal game" keeps it a plain session.
 */

const clubsList = vi.hoisted(() =>
	vi.fn().mockResolvedValue([
		{
			id: 'club_1',
			name: 'Bouvet Padel',
			avatar_icon: 'Trophy',
			avatar_color: 'forest',
			my_role: 'admin',
			roster_count: 4
		}
	])
);
const sessionsCreate = vi.hoisted(() =>
	vi.fn().mockResolvedValue({ id: 'S1', admin_token: 'atok' })
);
const goto = vi.hoisted(() => vi.fn());

vi.mock('$app/navigation', () => ({ goto }));
vi.mock('$lib/playerSession', () => ({ savePlayerSession: vi.fn() }));
vi.mock('$lib/auth.svelte', () => ({
	auth: { token: 'user-token', user: { id: 'u1', display_name: 'Alice' } }
}));

vi.mock('$lib/api/client', async (importOriginal) => {
	const actual = await importOriginal<typeof import('$lib/api/client')>();
	return {
		ApiError: actual.ApiError,
		api: {
			clubs: { list: clubsList },
			sessions: { create: sessionsCreate },
			players: { join: vi.fn().mockResolvedValue({ id: 'p1' }) }
		}
	};
});

beforeAll(async () => {
	register('en', () => import('../i18n/en.json'));
	init({ fallbackLocale: 'en', initialLocale: 'en' });
	await waitLocale('en');
});

beforeEach(() => {
	clubsList.mockClear();
	sessionsCreate.mockClear();
	goto.mockClear();
});

afterEach(() => {
	document.body.style.pointerEvents = '';
	document.body.style.removeProperty('--scrollbar-width');
});

describe('CreateDrawer — optional club picker', () => {
	it('lists the caller’s clubs with a personal default', async () => {
		render(CreateDrawer, { open: true });

		expect(await screen.findByText('Bouvet Padel')).toBeInTheDocument();
		expect(screen.getByText(/none — personal game/i)).toBeInTheDocument();
		// No club is selected yet, so it stays the ordinary create flow.
		expect(screen.queryByText(/whole club will be notified/i)).not.toBeInTheDocument();
		expect(screen.getByRole('button', { name: /invite link/i })).toBeInTheDocument();
	});

	it('picking a club reveals the banner and switches the CTA', async () => {
		const user = userEvent.setup();
		render(CreateDrawer, { open: true });

		await user.click(await screen.findByRole('button', { name: /bouvet padel/i }));

		expect(await screen.findByText(/scheduling for bouvet padel/i)).toBeInTheDocument();
		expect(screen.getByRole('button', { name: /notify club/i })).toBeInTheDocument();
	});

	it('sends club_id when a club is picked, and none when personal', async () => {
		const user = userEvent.setup();
		render(CreateDrawer, { open: true });

		// Pick the club, then create.
		await user.click(await screen.findByRole('button', { name: /bouvet padel/i }));
		await user.click(screen.getByRole('button', { name: /notify club/i }));

		await waitFor(() => expect(sessionsCreate).toHaveBeenCalled());
		expect(sessionsCreate.mock.calls[0][0]).toMatchObject({ club_id: 'club_1' });
	});
});
