import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { SegmentedControl, type SegmentedOption } from './index';

/**
 * The segmented control is a single-choice form control: exactly one option is
 * always selected. This is the behaviour that a toggle-button group gets wrong —
 * clicking the active button there clears the selection. These tests lock in
 * "one is always selected" and "clicking the active option can't deselect it".
 */

const setup = () => userEvent.setup({ pointerEventsCheck: 0 });

const OPTIONS: SegmentedOption[] = [
	{ value: 'americano', label: 'Americano' },
	{ value: 'mexicano', label: 'Mexicano' }
];

describe('SegmentedControl', () => {
	it('renders every option as a radio', () => {
		render(SegmentedControl, { options: OPTIONS, value: 'americano' });

		expect(screen.getByRole('radio', { name: 'Americano' })).toBeInTheDocument();
		expect(screen.getByRole('radio', { name: 'Mexicano' })).toBeInTheDocument();
	});

	it('marks the current value as checked', () => {
		render(SegmentedControl, { options: OPTIONS, value: 'americano' });

		expect(screen.getByRole('radio', { name: 'Americano' })).toBeChecked();
		expect(screen.getByRole('radio', { name: 'Mexicano' })).not.toBeChecked();
	});

	it('emits the new value when a different option is chosen', async () => {
		const onChange = vi.fn();
		const user = setup();
		render(SegmentedControl, { options: OPTIONS, value: 'americano', onChange });

		await user.click(screen.getByRole('radio', { name: 'Mexicano' }));

		expect(onChange).toHaveBeenCalledWith('mexicano');
	});

	it('cannot be deselected: clicking the active option keeps it selected and emits nothing', async () => {
		const onChange = vi.fn();
		const user = setup();
		render(SegmentedControl, { options: OPTIONS, value: 'americano', onChange });

		await user.click(screen.getByRole('radio', { name: 'Americano' }));

		// The regression this guards: a toggle-group would fire an empty value here.
		expect(onChange).not.toHaveBeenCalled();
		expect(screen.getByRole('radio', { name: 'Americano' })).toBeChecked();
	});

	it('does not emit for a disabled option', async () => {
		const onChange = vi.fn();
		const user = setup();
		const opts: SegmentedOption[] = [
			{ value: '1', label: '1', disabled: true },
			{ value: '2', label: '2' }
		];
		render(SegmentedControl, { options: opts, value: '2', onChange });

		await user.click(screen.getByRole('radio', { name: '1' }));

		expect(onChange).not.toHaveBeenCalled();
		expect(screen.getByRole('radio', { name: '2' })).toBeChecked();
	});
});
