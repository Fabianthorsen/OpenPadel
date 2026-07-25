import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import type { Component, ComponentProps } from 'svelte';

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

/**
 * Formats a Club-leaderboard form value (a normalized point-margin in [-1, +1])
 * as a signed score out of ±100, so it reads like a rating: 0 = even, positive
 * = winning by a wider margin than you concede. "+32", "0", "-15".
 */
export function formScore(form: number): string {
	const n = Math.round(form * 100);
	return n > 0 ? `+${n}` : `${n}`;
}

/** Returns "Firstname L." for multi-word names, or just the name if single word. "Fabian Thorsen" → "Fabian T." */
export function shortName(name: string): string {
	const words = name.trim().split(/\s+/).filter(Boolean);
	if (words.length <= 1) return name.trim();
	return `${words[0]} ${words[words.length - 1][0].toUpperCase()}.`;
}

/**
 * Returns a display name for a session. An explicit name always wins; otherwise
 * a club event falls back to a name tied to its Club ("<Club> <Mode>"), and a
 * plain session to the generic default.
 */
export function sessionName(session: {
	name?: string;
	club_name?: string;
	game_mode?: string;
}): string {
	if (session.name) return session.name;
	if (session.club_name) {
		const mode = session.game_mode === 'mexicano' ? 'Mexicano' : 'Americano';
		return `${session.club_name} ${mode}`;
	}
	return 'OpenPadel Americano';
}

/** Returns up to 2 uppercase initials from a display name. "Fabian Thorsen" → "FT" */
export function initials(name: string): string {
	const words = name.trim().split(/\s+/).filter(Boolean);
	if (words.length === 0) return '?';
	if (words.length === 1) return words[0][0].toUpperCase();
	return (words[0][0] + words[words.length - 1][0]).toUpperCase();
}

export type WithElementRef<T, E extends Element = HTMLElement> = T & {
	ref?: E | null;
};

export type WithoutChildren<T> = Omit<T, 'children'>;
export type WithoutChildrenOrChild<T> = Omit<T, 'children' | 'child'>;
export type AsChild<T extends Component> = {
	asChild?: boolean;
	child?: Component<{ props: ComponentProps<T> }>;
};
