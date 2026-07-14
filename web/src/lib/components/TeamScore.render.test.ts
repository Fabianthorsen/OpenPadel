import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import TeamScore from './TeamScore.svelte';

describe('TeamScore', () => {
	it('renders a static span (no button) when not interactive', () => {
		render(TeamScore, { props: { score: 12, opponentScore: 12, scored: true } });
		expect(screen.queryByRole('button')).not.toBeInTheDocument();
		expect(screen.getByText('12')).toBeInTheDocument();
	});

	it('renders a button with the given label when interactive, and fires onTap', async () => {
		const user = userEvent.setup();
		const onTap = vi.fn();
		render(TeamScore, {
			props: {
				score: 5,
				opponentScore: 0,
				scored: false,
				interactive: true,
				label: 'Set Team A score',
				onTap
			}
		});

		const btn = screen.getByRole('button', { name: 'Set Team A score' });
		await user.click(btn);
		expect(onTap).toHaveBeenCalledTimes(1);
	});

	it('emphasises the winner and dims the loser once scored', () => {
		const { unmount } = render(TeamScore, {
			props: { score: 16, opponentScore: 8, scored: true }
		});
		expect(screen.getByText('16')).toHaveClass('text-primary', 'font-bold');
		unmount();

		render(TeamScore, { props: { score: 8, opponentScore: 16, scored: true } });
		expect(screen.getByText('8')).toHaveClass('text-text-disabled');
	});

	it('shows the underline affordance only for an interactive, not-yet-scored score', () => {
		render(TeamScore, {
			props: {
				score: 3,
				opponentScore: 0,
				scored: false,
				interactive: true,
				underline: true,
				label: 'Set Team A score'
			}
		});
		expect(screen.getByRole('button')).toHaveClass('underline');
	});

	it('supports a large (hero) and default (compact) size', () => {
		const { unmount } = render(TeamScore, {
			props: { score: 1, opponentScore: 1, scored: false, size: 'lg' }
		});
		expect(screen.getByText('1')).toHaveClass('text-5xl');
		unmount();

		render(TeamScore, { props: { score: 2, opponentScore: 2, scored: false } });
		expect(screen.getByText('2')).toHaveClass('text-2xl');
	});
});
