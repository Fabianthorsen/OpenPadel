import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import Fixture from './section.fixture.svelte';

describe('Section (render)', () => {
	it('renders the title and body when open', () => {
		render(Fixture, { title: 'Contacts', open: true });
		expect(screen.getByText('Contacts')).toBeInTheDocument();
		expect(screen.getByText('section body content')).toBeInTheDocument();
	});

	it('exposes a collapse toggle when collapsible', () => {
		render(Fixture, { collapsible: true });
		expect(screen.getByRole('button')).toBeInTheDocument();
	});

	it('renders a static block with no toggle when not collapsible', () => {
		render(Fixture, { collapsible: false });
		expect(screen.queryByRole('button')).not.toBeInTheDocument();
	});
});
