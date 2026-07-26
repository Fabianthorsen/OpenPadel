import { describe, it, expect, beforeAll } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { init, register, waitLocale } from 'svelte-i18n';
import ExpandableListTest from './expandable-list.fixture.svelte';

// The "more"/"less" buttons render through svelte-i18n, so the store must have a
// locale before these mount — otherwise `$_` emits the raw key and the
// name-based queries below miss.
beforeAll(async () => {
	register('en', () => import('../../../i18n/en.json'));
	init({ fallbackLocale: 'en', initialLocale: 'en' });
	await waitLocale('en');
});

describe('ExpandableList', () => {
	it('shows initial items and hides excess', () => {
		const { container } = render(ExpandableListTest, {
			props: { items: Array.from({ length: 10 }, (_, i) => i), showCount: 3 }
		});

		expect(screen.getAllByRole('button', { name: /item/i })).toHaveLength(3);
		expect(screen.getByRole('button', { name: /7 more/i })).toBeInTheDocument();
	});

	it('expands list on "show more" click', async () => {
		const user = userEvent.setup();
		const { container } = render(ExpandableListTest, {
			props: { items: Array.from({ length: 10 }, (_, i) => i), showCount: 3 }
		});

		await user.click(screen.getByRole('button', { name: /7 more/i }));

		expect(screen.getAllByRole('button', { name: /item/i })).toHaveLength(10);
		expect(screen.getByRole('button', { name: /show less/i })).toBeInTheDocument();
	});

	it('collapses list on "show less" click', async () => {
		const user = userEvent.setup();
		const { container } = render(ExpandableListTest, {
			props: { items: Array.from({ length: 10 }, (_, i) => i), showCount: 3 }
		});

		await user.click(screen.getByRole('button', { name: /7 more/i }));
		await user.click(screen.getByRole('button', { name: /show less/i }));

		expect(screen.getAllByRole('button', { name: /item/i })).toHaveLength(3);
	});

	it('does not show expand button for short lists', () => {
		render(ExpandableListTest, {
			props: { items: Array.from({ length: 3 }, (_, i) => i), showCount: 5 }
		});

		expect(screen.queryByRole('button', { name: /more/ })).not.toBeInTheDocument();
	});

	it('uses custom showCount prop', () => {
		render(ExpandableListTest, {
			props: { items: Array.from({ length: 10 }, (_, i) => i), showCount: 6 }
		});

		expect(screen.getAllByRole('button', { name: /item/i })).toHaveLength(6);
		expect(screen.getByRole('button', { name: /4 more/i })).toBeInTheDocument();
	});
});
