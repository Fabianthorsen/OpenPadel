import { render } from '@testing-library/svelte';
import { beforeAll, describe, expect, it } from 'vitest';
import { init, register, waitLocale } from 'svelte-i18n';
import ScoreEntrySheet from './ScoreEntrySheet.svelte';

describe('ScoreEntrySheet', () => {
	beforeAll(async () => {
		register('en', () => import('../i18n/en.json'));
		init({ fallbackLocale: 'en', initialLocale: 'en' });
		await waitLocale('en');
	});

	const defaultProps = {
		matchId: 'match-1',
		courtNumber: 1,
		roundNumber: 2,
		pointsTarget: 24,
		teamAName: 'Player 1 & Player 2',
		teamBName: 'Player 3 & Player 4',
		onSubmit: async () => {},
		onClose: () => {},
		onLiveSave: () => {}
	};

	it('renders the sheet with court header', () => {
		const { getByText } = render(ScoreEntrySheet, { props: defaultProps });
		expect(getByText(/Round 2/i)).toBeInTheDocument();
	});

	it('renders team A and B blocks with names', () => {
		const { getByText } = render(ScoreEntrySheet, { props: defaultProps });
		expect(getByText('Player 1 & Player 2')).toBeInTheDocument();
		expect(getByText('Player 3 & Player 4')).toBeInTheDocument();
	});

	it('shows validity readout with target', () => {
		const { getByText } = render(ScoreEntrySheet, { props: defaultProps });
		expect(getByText(/24/)).toBeInTheDocument(); // Target points
	});

	it('renders plus and minus buttons for team A', () => {
		const { getByLabelText } = render(ScoreEntrySheet, { props: defaultProps });
		expect(getByLabelText(/increase.*team a/i)).toBeInTheDocument();
		expect(getByLabelText(/decrease.*team a/i)).toBeInTheDocument();
	});

	it('renders plus and minus buttons for team B', () => {
		const { getByLabelText } = render(ScoreEntrySheet, { props: defaultProps });
		expect(getByLabelText(/increase.*team b/i)).toBeInTheDocument();
		expect(getByLabelText(/decrease.*team b/i)).toBeInTheDocument();
	});

	it('renders Finalize button', () => {
		const { getByRole } = render(ScoreEntrySheet, { props: defaultProps });
		expect(getByRole('button', { name: /finalize/i })).toBeInTheDocument();
	});

	it('disables Finalize button when scores do not equal target', () => {
		const { getByRole } = render(ScoreEntrySheet, { props: defaultProps });
		const finalizeBtn = getByRole('button', { name: /finalize/i });
		expect(finalizeBtn).toBeDisabled();
	});

	it('renders close button', () => {
		const { getByLabelText } = render(ScoreEntrySheet, { props: defaultProps });
		expect(getByLabelText(/close|back/i)).toBeInTheDocument();
	});
});
