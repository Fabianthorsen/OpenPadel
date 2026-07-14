import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import JoinCodeInput from './join-code-input.svelte';

describe('JoinCodeInput (render)', () => {
	it('renders four single-character boxes', () => {
		render(JoinCodeInput);
		expect(screen.getAllByRole('textbox')).toHaveLength(4);
	});

	it('fires onComplete with the uppercased code once all boxes are filled', async () => {
		const user = userEvent.setup();
		const onComplete = vi.fn();
		render(JoinCodeInput, { onComplete });

		const boxes = screen.getAllByRole('textbox');
		await user.type(boxes[0], 'a');
		await user.type(boxes[1], 'b');
		await user.type(boxes[2], 'c');
		await user.type(boxes[3], 'd');

		expect(onComplete).toHaveBeenLastCalledWith('ABCD');
	});

	it('does not fire onComplete before every box is filled', async () => {
		const user = userEvent.setup();
		const onComplete = vi.fn();
		render(JoinCodeInput, { onComplete });

		const boxes = screen.getAllByRole('textbox');
		await user.type(boxes[0], 'a');
		await user.type(boxes[1], 'b');

		expect(onComplete).not.toHaveBeenCalled();
	});
});
