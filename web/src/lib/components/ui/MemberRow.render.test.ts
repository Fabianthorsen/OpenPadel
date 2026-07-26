import { describe, it, expect, beforeAll } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import { init, register, waitLocale } from 'svelte-i18n';
import MemberRow from './MemberRow.svelte';

// The role label renders through svelte-i18n, so the store needs a locale before
// the component mounts.
beforeAll(async () => {
	register('en', () => import('../../i18n/en.json'));
	init({ fallbackLocale: 'en', initialLocale: 'en' });
	await waitLocale('en');
});

const member: App.ClubMember = {
	user_id: 'u1',
	display_name: 'Ada Lovelace',
	role: 'admin',
	avatar_icon: '',
	avatar_color: 'forest',
	joined_at: '2026-01-01T00:00:00Z'
};

describe('MemberRow (render)', () => {
	it('renders the member display name', () => {
		render(MemberRow, { member });
		expect(screen.getByText('Ada Lovelace')).toBeInTheDocument();
	});

	it('renders the translated member role', () => {
		render(MemberRow, { member });
		expect(screen.getByText('Admin')).toBeInTheDocument();
	});

	it('falls back to initials when no avatar icon is set', () => {
		render(MemberRow, { member });
		expect(screen.getByText('AL')).toBeInTheDocument();
	});
});
