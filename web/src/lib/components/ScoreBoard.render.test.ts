import { describe, it, expect, vi, beforeAll } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { init, register, waitLocale } from 'svelte-i18n';
import ScoreBoard from './ScoreBoard.svelte';

/**
 * Render/interaction tests for the single-court ScoreBoard: the +/- steppers and
 * tap-a-score affordance are only shown while editable (admin + not yet scored),
 * and the Finalize button gates on the scores summing to the points target.
 * These guard the wiring between the on-card controls and the callbacks the
 * parent (ActiveSession) uses to adjust / open the numpad / submit.
 */
beforeAll(async () => {
	register('en', () => import('../i18n/en.json'));
	init({ fallbackLocale: 'en', initialLocale: 'en' });
	await waitLocale('en');
});

const teamA = {
	players: [{ avatar_icon: 'racket', avatar_color: '#3d7a24', name: 'Ann', rating: 4 }],
	name: 'Ann & Bo',
	score: 0
};
const teamB = {
	players: [{ avatar_icon: 'racket', avatar_color: '#3d7a24', name: 'Cy', rating: 2 }],
	name: 'Cy & Di',
	score: 0
};

function baseProps(overrides = {}) {
	return {
		teamA,
		teamB,
		scored: false,
		live: false,
		pointsTarget: 24,
		isAdmin: true,
		submitting: false,
		onAdjust: vi.fn(),
		onScoreTap: vi.fn(),
		onFinalize: vi.fn(),
		...overrides
	};
}

describe('ScoreBoard', () => {
	it('renders each player with their name and Rating badge', () => {
		render(ScoreBoard, { props: baseProps() });
		expect(screen.getByText('Ann')).toBeInTheDocument();
		expect(screen.getByText('Cy')).toBeInTheDocument();
		// The score-entry view (admin taps here) surfaces both players' ratings.
		const badges = screen.getAllByLabelText('Rating');
		expect(badges).toHaveLength(2);
		expect(badges.map((b) => b.textContent?.trim())).toEqual(['4', '2']);
	});

	it('shows steppers and tappable scores when editable (admin + not scored)', () => {
		render(ScoreBoard, { props: baseProps() });
		expect(screen.getByLabelText(/increase team a/i)).toBeInTheDocument();
		expect(screen.getByLabelText(/decrease team a/i)).toBeInTheDocument();
		expect(screen.getByLabelText(/increase team b/i)).toBeInTheDocument();
		expect(screen.getByLabelText(/decrease team b/i)).toBeInTheDocument();
		expect(screen.getByLabelText(/set team a score/i)).toBeInTheDocument();
		expect(screen.getByLabelText(/set team b score/i)).toBeInTheDocument();
	});

	it('hides steppers and finalize once scored', () => {
		render(ScoreBoard, { props: baseProps({ scored: true, teamA: { ...teamA, score: 24 } }) });
		expect(screen.queryByLabelText(/increase team a/i)).not.toBeInTheDocument();
		expect(screen.queryByRole('button', { name: /finalize/i })).not.toBeInTheDocument();
		expect(screen.getByText('Final')).toBeInTheDocument();
	});

	it('hides controls entirely for non-admins', () => {
		render(ScoreBoard, { props: baseProps({ isAdmin: false }) });
		expect(screen.queryByLabelText(/increase team a/i)).not.toBeInTheDocument();
		expect(screen.queryByRole('button', { name: /finalize/i })).not.toBeInTheDocument();
	});

	it('disables Finalize until the scores sum to the target', () => {
		render(ScoreBoard, { props: baseProps() }); // 0 + 0 !== 24
		expect(screen.getByRole('button', { name: /finalize/i })).toBeDisabled();
	});

	it('enables Finalize when the scores sum to the target', () => {
		render(ScoreBoard, {
			props: baseProps({ teamA: { ...teamA, score: 16 }, teamB: { ...teamB, score: 8 } })
		});
		expect(screen.getByRole('button', { name: /finalize/i })).toBeEnabled();
	});

	it('calls onAdjust with the team and delta when a stepper is clicked', async () => {
		const user = userEvent.setup();
		const onAdjust = vi.fn();
		render(ScoreBoard, { props: baseProps({ teamA: { ...teamA, score: 3 }, onAdjust }) });

		await user.click(screen.getByLabelText(/increase team a/i));
		expect(onAdjust).toHaveBeenCalledWith('a', 1);

		await user.click(screen.getByLabelText(/decrease team a/i));
		expect(onAdjust).toHaveBeenCalledWith('a', -1);
	});

	it('calls onScoreTap when a score is tapped', async () => {
		const user = userEvent.setup();
		const onScoreTap = vi.fn();
		render(ScoreBoard, { props: baseProps({ onScoreTap }) });

		await user.click(screen.getByLabelText(/set team b score/i));
		expect(onScoreTap).toHaveBeenCalledWith('b');
	});

	it('calls onFinalize when the enabled Finalize button is clicked', async () => {
		const user = userEvent.setup();
		const onFinalize = vi.fn();
		render(ScoreBoard, {
			props: baseProps({
				teamA: { ...teamA, score: 12 },
				teamB: { ...teamB, score: 12 },
				onFinalize
			})
		});

		await user.click(screen.getByRole('button', { name: /finalize/i }));
		expect(onFinalize).toHaveBeenCalledTimes(1);
	});
});
