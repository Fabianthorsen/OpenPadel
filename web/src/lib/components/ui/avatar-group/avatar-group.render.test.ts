import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/svelte';
import AvatarGroup from './avatar-group.svelte';

const people = (n: number) =>
	Array.from({ length: n }, (_, i) => ({ name: `Player ${i}`, avatar_color: 'forest' }));

describe('AvatarGroup (render)', () => {
	it('shows no overflow chip when under the max', () => {
		const { container } = render(AvatarGroup, { people: people(3), max: 4 });
		expect(container.textContent).not.toContain('+');
	});

	it('caps at max and shows a +N overflow chip', () => {
		const { getByText } = render(AvatarGroup, { people: people(7), max: 4 });
		expect(getByText('+3')).toBeInTheDocument();
	});
});
