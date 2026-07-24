import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import MemberRow from './MemberRow.svelte';

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

	it('renders the member role', () => {
		render(MemberRow, { member });
		expect(screen.getByText('admin')).toBeInTheDocument();
	});

	it('falls back to initials when no avatar icon is set', () => {
		render(MemberRow, { member });
		expect(screen.getByText('AL')).toBeInTheDocument();
	});
});
