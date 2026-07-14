import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import StepButton from './StepButton.svelte';

describe('StepButton', () => {
	it('shows + for increase and − for decrease, with the given label', () => {
		const { unmount } = render(StepButton, {
			props: { direction: 'increase', label: 'Increase Team A score', onclick: () => {} }
		});
		expect(screen.getByRole('button', { name: 'Increase Team A score' })).toHaveTextContent('+');
		unmount();

		render(StepButton, {
			props: { direction: 'decrease', label: 'Decrease Team A score', onclick: () => {} }
		});
		expect(screen.getByRole('button', { name: 'Decrease Team A score' })).toHaveTextContent('−');
	});

	it('fires onclick when pressed', async () => {
		const user = userEvent.setup();
		const onclick = vi.fn();
		render(StepButton, { props: { direction: 'increase', label: 'Increase', onclick } });

		await user.click(screen.getByRole('button'));
		expect(onclick).toHaveBeenCalledTimes(1);
	});

	it('does not fire when disabled', async () => {
		const user = userEvent.setup();
		const onclick = vi.fn();
		render(StepButton, {
			props: { direction: 'decrease', label: 'Decrease', disabled: true, onclick }
		});

		const btn = screen.getByRole('button');
		expect(btn).toBeDisabled();
		await user.click(btn);
		expect(onclick).not.toHaveBeenCalled();
	});
});
