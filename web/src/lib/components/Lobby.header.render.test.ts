import { describe, it, expect, beforeAll, beforeEach, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import { init, register, waitLocale } from 'svelte-i18n';
import Lobby from './Lobby.svelte';

/**
 * Render tests for the redesigned lobby header (#182).
 *
 * Guards three things that were fiddly to get right:
 *  - Long tournament names must clip (truncate), not widen the page. This is a
 *    CSS-class contract: the name node carries `truncate` + `min-w-0` so the
 *    flex row can shrink it instead of overflowing the viewport.
 *  - The game-mode text is the rules affordance (a button), replacing the old
 *    standalone info icon.
 *  - Config editing is admin-only, reached via the Settings icon → drawer.
 */

// The admin branch calls api.invites.listForSession on mount; stub the client
// so render doesn't reach the network. ApiError is preserved for the catch path.
vi.mock('$lib/api/client', async (importOriginal) => {
	const actual = await importOriginal<typeof import('$lib/api/client')>();
	return {
		ApiError: actual.ApiError,
		api: {
			invites: { listForSession: vi.fn().mockResolvedValue([]) },
			sessions: { update: vi.fn().mockResolvedValue({}), start: vi.fn(), cancel: vi.fn() },
			players: { join: vi.fn(), remove: vi.fn() },
			contacts: { search: vi.fn().mockResolvedValue([]) }
		}
	};
});

beforeAll(async () => {
	register('en', () => import('../i18n/en.json'));
	init({ fallbackLocale: 'en', initialLocale: 'en' });
	await waitLocale('en');
});

beforeEach(() => {
	localStorage.clear();
});

const LONG_NAME =
	'The Great Annual Friday Night Padel Championship of the Whole Entire Neighbourhood';

function makePlayers(): App.Player[] {
	return [1, 2, 3, 4].map((n) => ({
		id: `p${n}`,
		session_id: 's1',
		name: `Player ${n}`,
		avatar_icon: 'racket',
		avatar_color: '#3d7a24',
		active: true,
		joined_at: '2026-07-14T20:00:00Z'
	}));
}

function makeSession(overrides?: Partial<App.Session>): App.Session {
	return {
		id: 'ABCD',
		status: 'lobby',
		game_mode: 'americano',
		courts: 2,
		points: 24,
		rounds_total: undefined,
		current_round: 0,
		players: makePlayers(),
		created_at: '2026-07-14T20:00:00Z',
		updated_at: '2026-07-14T20:00:00Z',
		...overrides
	};
}

function makeProps(session: App.Session, isAdmin: boolean) {
	return {
		session,
		isAdmin,
		onRefresh: () => {},
		onStarted: () => {}
	};
}

describe('Lobby header — admin', () => {
	it('renders the tournament name as an edit affordance and a settings button', () => {
		render(Lobby, makeProps(makeSession({ name: 'Friday Padel' }), true));

		// Click-to-edit name (accessible name is the tournament title).
		expect(screen.getByRole('button', { name: /friday padel/i })).toBeInTheDocument();
		// Settings icon opens the config drawer.
		expect(screen.getByRole('button', { name: /edit settings/i })).toBeInTheDocument();
	});

	it('long names carry the truncate/min-w-0 classes so they clip instead of widening the page', () => {
		render(Lobby, makeProps(makeSession({ name: LONG_NAME }), true));

		const nameEl = screen.getByText(LONG_NAME);
		expect(nameEl.className).toContain('truncate');
		expect(nameEl.className).toContain('min-w-0');
	});

	it('the container is responsive: w-full capped at the mobile max and centered', () => {
		// Root cause of the rename/long-name overflow: without `w-full`, the
		// `mx-auto` on a column-flex child disables align-stretch, so `main`
		// sizes to its content and a long name pushes it past the viewport.
		// `w-full` pins it to min(100%, 480px) so child truncation can work.
		render(Lobby, makeProps(makeSession({ name: LONG_NAME }), true));

		const main = screen.getByRole('main');
		expect(main.className).toContain('w-full');
		expect(main.className).toContain('max-w-[480px]');
		expect(main.className).toContain('mx-auto');
	});

	it('exposes the game mode as the rules trigger (a button), not plain text', () => {
		render(Lobby, makeProps(makeSession({ name: 'Friday Padel' }), true));

		expect(screen.getByRole('button', { name: /^americano$/i })).toBeInTheDocument();
	});

	it('shows a config summary with courts and points', () => {
		render(Lobby, makeProps(makeSession({ name: 'Friday Padel', courts: 2, points: 24 }), true));

		expect(screen.getByText(/2 courts/i)).toBeInTheDocument();
		expect(screen.getByText(/24/)).toBeInTheDocument();
	});

	it('the rename input shares the row (flex-1/min-w-0) instead of forcing width:100% and widening the page', async () => {
		render(Lobby, makeProps(makeSession({ name: 'Friday Padel' }), true));

		await fireEvent.click(screen.getByRole('button', { name: /friday padel/i }));

		// Entering edit mode swaps the title for a text input. It must shrink to
		// share the row with the action icons — `w-full` would overflow by their width.
		const input = screen.getByPlaceholderText(/tournament name/i);
		expect(input.className).toContain('flex-1');
		expect(input.className).toContain('min-w-0');
		expect(input.className).not.toContain('w-full');
	});

	it('opens the config drawer when the settings button is clicked', async () => {
		render(Lobby, makeProps(makeSession({ name: 'Friday Padel' }), true));

		await fireEvent.click(screen.getByRole('button', { name: /edit settings/i }));

		// Drawer title uses the same key; the drawer also surfaces the mode/courts controls.
		expect(await screen.findByRole('heading', { name: /edit settings/i })).toBeInTheDocument();
	});
});

describe('Lobby header — non-admin (joined)', () => {
	it('renders the name as a static heading with no settings control', () => {
		// Joined-as-guest is derived from localStorage player_id matching an active player.
		localStorage.setItem('player_id_ABCD', 'p1');
		render(Lobby, makeProps(makeSession({ name: 'Friday Padel' }), false));

		expect(screen.getByRole('heading', { name: /friday padel/i })).toBeInTheDocument();
		expect(screen.queryByRole('button', { name: /edit settings/i })).not.toBeInTheDocument();
	});

	it('a long name on the static heading still carries truncate', () => {
		localStorage.setItem('player_id_ABCD', 'p1');
		render(Lobby, makeProps(makeSession({ name: LONG_NAME }), false));

		expect(screen.getByRole('heading', { name: new RegExp(LONG_NAME, 'i') }).className).toContain(
			'truncate'
		);
	});
});
