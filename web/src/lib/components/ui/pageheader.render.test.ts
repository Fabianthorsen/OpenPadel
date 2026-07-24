import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import Fixture from './pageheader.fixture.svelte';

describe('PageHeader (render)', () => {
	it('renders the title', () => {
		render(Fixture, { title: 'My Club' });
		expect(screen.getByRole('heading', { name: 'My Club' })).toBeInTheDocument();
	});

	it('renders the subtitle when provided', () => {
		render(Fixture, { subtitle: '5 members' });
		expect(screen.getByText('5 members')).toBeInTheDocument();
	});

	it('renders a back link when backHref is set', () => {
		render(Fixture, { backHref: '/profile' });
		const back = screen.getByLabelText('Back');
		expect(back).toBeInTheDocument();
		expect(back).toHaveAttribute('href', '/profile');
	});

	it('renders no back link when backHref is absent', () => {
		render(Fixture, {});
		expect(screen.queryByLabelText('Back')).not.toBeInTheDocument();
	});

	it('renders an avatar when provided', () => {
		render(Fixture, { avatar: { name: 'Ada Lovelace', color: 'forest' } });
		expect(screen.getByText('AL')).toBeInTheDocument();
	});

	it('renders the action snippet', () => {
		render(Fixture, { withAction: true });
		expect(screen.getByTestId('header-action')).toBeInTheDocument();
	});

	it('renders below-header children content', () => {
		render(Fixture, { withChildren: true });
		expect(screen.getByTestId('header-children')).toBeInTheDocument();
	});
});
