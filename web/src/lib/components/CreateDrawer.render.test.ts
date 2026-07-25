import { describe, it, expect, beforeAll, afterEach, vi } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import { init, register, waitLocale } from 'svelte-i18n';
import CreateDrawer from './CreateDrawer.svelte';

/**
 * Render tests for the CreateDrawer's club-event preset (#127).
 *
 * When opened from a Club home the drawer carries a `club` prop, which must
 * re-frame the flow: a "whole club will be notified" banner, a club title, and a
 * "notify club" CTA — distinct from the ordinary share-an-invite-link create
 * flow. Without the prop it stays the normal tournament-create drawer.
 */

vi.mock('$lib/api/client', async (importOriginal) => {
	const actual = await importOriginal<typeof import('$lib/api/client')>();
	return {
		ApiError: actual.ApiError,
		api: { sessions: { create: vi.fn() }, players: { join: vi.fn() } }
	};
});

beforeAll(async () => {
	register('en', () => import('../i18n/en.json'));
	init({ fallbackLocale: 'en', initialLocale: 'en' });
	await waitLocale('en');
});

afterEach(() => {
	// Bits UI's scroll-lock leaves styles on <body> when a Drawer unmounts; clear
	// them so they don't leak into the next test's document (shared across files).
	document.body.style.pointerEvents = '';
	document.body.style.removeProperty('--scrollbar-width');
});

describe('CreateDrawer — club preset', () => {
	it('re-frames title, banner, and CTA for a club event', () => {
		render(CreateDrawer, { open: true, club: { id: 'club1', name: 'Bouvet Padel' } });

		expect(screen.getByRole('heading', { name: /schedule a club game/i })).toBeInTheDocument();
		expect(screen.getByText(/scheduling for bouvet padel/i)).toBeInTheDocument();
		expect(screen.getByRole('button', { name: /notify club/i })).toBeInTheDocument();
		// The single-user "get an invite link" framing must not appear here.
		expect(screen.queryByRole('button', { name: /invite link/i })).not.toBeInTheDocument();
	});

	it('stays the ordinary tournament-create drawer with no club prop', () => {
		render(CreateDrawer, { open: true });

		expect(screen.getByRole('heading', { name: /create new tournament/i })).toBeInTheDocument();
		expect(screen.getByRole('button', { name: /invite link/i })).toBeInTheDocument();
		// No club banner / club CTA when the drawer isn't club-scoped.
		expect(screen.queryByText(/whole club will be notified/i)).not.toBeInTheDocument();
		expect(screen.queryByRole('button', { name: /notify club/i })).not.toBeInTheDocument();
	});
});
