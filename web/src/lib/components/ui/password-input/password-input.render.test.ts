import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import PasswordInput from './password-input.svelte';

describe('PasswordInput (render)', () => {
	it('renders a masked password input by default', () => {
		const { container } = render(PasswordInput);
		expect(container.querySelector('input')?.getAttribute('type')).toBe('password');
	});

	it('toggles to text and back via the show/hide button', async () => {
		const user = userEvent.setup();
		const { container } = render(PasswordInput);
		const type = () => container.querySelector('input')?.getAttribute('type');

		expect(type()).toBe('password');
		await user.click(screen.getByRole('button', { name: 'Show password' }));
		expect(type()).toBe('text');
		await user.click(screen.getByRole('button', { name: 'Hide password' }));
		expect(type()).toBe('password');
	});
});
