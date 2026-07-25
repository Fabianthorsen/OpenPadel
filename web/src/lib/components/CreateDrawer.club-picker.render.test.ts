import { describe, it, expect, beforeAll, afterEach, beforeEach, vi } from 'vitest';
import { render, screen } from '@testing-library/svelte';
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
	// bits-ui's Select uses pointer capture the jsdom environment doesn't emulate,
	// so disable user-event's pointer-events guard for these interactions.
	const user = () => userEvent.setup({ pointerEventsCheck: 0 });

	it('shows a personal default in the collapsed trigger', async () => {
		render(CreateDrawer, { open: true });

		// The trigger shows the "none" default; clubs stay hidden until it opens.
		expect(await screen.findByText(/none — personal game/i)).toBeInTheDocument();
		expect(screen.queryByText('Bouvet Padel')).not.toBeInTheDocument();
		// Nothing selected yet, so it stays the ordinary create flow.
		expect(screen.queryByText(/whole club will be notified/i)).not.toBeInTheDocument();
		expect(screen.getByRole('button', { name: /invite link/i })).toBeInTheDocument();
	});

	it('survives a null club-list response without crashing', async () => {
		// A user with no clubs gets JSON `null` from the endpoint (Go nil slice).
		// Assigned raw, `myClubs.length` threw inside Svelte's reactive flush and
		// froze the whole page. That throw lands in an async effect after render,
		// so catch it on the window rather than via a failed DOM assertion.
		// Svelte reports a throw inside an async effect as an unhandled promise
		// rejection, which surfaces on `process` (not the jsdom window).
		const errors: unknown[] = [];
		const onRejection = (reason: unknown) => errors.push(reason);
		process.on('unhandledRejection', onRejection);

		try {
			clubsList.mockResolvedValueOnce(null as unknown as App.ClubListItem[]);
			render(CreateDrawer, { open: true });

			// Let the awaited clubs.list resolve, its reactive flush run, and Node
			// promote any still-unhandled rejection (deferred to a macrotask).
			await screen.findByRole('button', { name: /invite link/i });
			await new Promise((r) => setImmediate(r));

			expect(errors).toEqual([]);
			// No clubs → the attach-to-a-club picker stays hidden.
			expect(screen.queryByLabelText(/attach to a club/i)).not.toBeInTheDocument();
		} finally {
			process.off('unhandledRejection', onRejection);
		}
	});

	it('exposes a labelled listbox trigger that opens', async () => {
		const u = user();
		render(CreateDrawer, { open: true });

		const trigger = await screen.findByLabelText(/attach to a club/i);
		expect(trigger).toHaveAttribute('aria-haspopup', 'listbox');

		await u.click(trigger);
		expect(trigger).toHaveAttribute('aria-expanded', 'true');
		// bits-ui's floating Select.Content options don't mount under jsdom (no
		// layout engine), so the option-pick → club_id path is covered end-to-end
		// against a live server rather than here.
	});
});
