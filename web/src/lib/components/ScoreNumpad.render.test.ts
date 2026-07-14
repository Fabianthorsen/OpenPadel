import { describe, it, expect, vi, afterEach } from 'vitest';
import { render, screen, within } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import ScoreNumpad from './ScoreNumpad.svelte';
import { numpad } from '$lib/stores/numpad';

/**
 * Render/interaction tests for the store-driven numpad drawer. The component
 * has no props — it opens whenever the numpad store holds a state and forwards
 * key presses to the callbacks that state carries. jsdom notes mirror the
 * Drawer tests: Bits UI's pointer-events lock leaks across the shared document,
 * so clicks disable the check and <body> is reset between tests.
 */
const setup = () => userEvent.setup({ pointerEventsCheck: 0 });

function openNumpad(overrides = {}) {
	const handlers = {
		onDigit: vi.fn(),
		onDelete: vi.fn(),
		onConfirm: vi.fn(),
		onClose: vi.fn()
	};
	numpad.open({
		value: '',
		fresh: true,
		targetPoints: 24,
		shaking: false,
		...handlers,
		...overrides
	});
	return handlers;
}

afterEach(() => {
	numpad.close();
	document.body.style.pointerEvents = '';
});

describe('ScoreNumpad', () => {
	it('renders nothing while the store is empty', () => {
		render(ScoreNumpad);
		expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
	});

	it('shows the keypad and target once the store is opened', async () => {
		openNumpad({ targetPoints: 24 });
		render(ScoreNumpad);

		const dialog = await screen.findByRole('dialog');
		for (const d of ['1', '2', '3', '4', '5', '6', '7', '8', '9', '0']) {
			expect(within(dialog).getByRole('button', { name: `Enter ${d}` })).toBeInTheDocument();
		}
		expect(within(dialog).getByRole('button', { name: 'Delete' })).toBeInTheDocument();
		expect(within(dialog).getByRole('button', { name: 'Confirm' })).toBeInTheDocument();
		expect(within(dialog).getByText(/Target:\s*24/)).toBeInTheDocument();
	});

	it('forwards a digit press to onDigit', async () => {
		const user = setup();
		const handlers = openNumpad();
		render(ScoreNumpad);

		await user.click(await screen.findByRole('button', { name: 'Enter 5' }));
		expect(handlers.onDigit).toHaveBeenCalledWith('5');
	});

	it('forwards delete and confirm presses', async () => {
		const user = setup();
		const handlers = openNumpad();
		render(ScoreNumpad);

		await user.click(await screen.findByRole('button', { name: 'Delete' }));
		expect(handlers.onDelete).toHaveBeenCalledTimes(1);

		await user.click(screen.getByRole('button', { name: 'Confirm' }));
		expect(handlers.onConfirm).toHaveBeenCalledTimes(1);
	});
});
