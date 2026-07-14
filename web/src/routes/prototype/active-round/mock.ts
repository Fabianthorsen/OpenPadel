// PROTOTYPE — wipe me. Mock data for the Active-round view prototype (wayfinder #170).
// Not wired to the backend; hardcoded so the variants can be judged against realistic density.

export type Player = {
	id: string;
	name: string;
	avatar_color: string;
	avatar_icon: string | undefined;
	active: boolean;
};
export type Match = {
	id: string;
	court: number;
	team_a: [string, string];
	team_b: [string, string];
	score: { a: number; b: number } | null;
	live: { a: number; b: number; server: 'a' | 'b' } | null;
};

export const points = 24;

export const players: Player[] = [
	{ id: 'p1', name: 'Ana Bergström', avatar_color: 'forest', avatar_icon: undefined, active: true },
	{ id: 'p2', name: 'Bruno', avatar_color: 'sky', avatar_icon: undefined, active: true },
	{ id: 'p3', name: 'Carl', avatar_color: 'orange', avatar_icon: undefined, active: true },
	{ id: 'p4', name: 'Diana Moreau', avatar_color: 'coral', avatar_icon: undefined, active: true },
	{ id: 'p5', name: 'Erik', avatar_color: 'purple', avatar_icon: undefined, active: true },
	{ id: 'p6', name: 'Fiona', avatar_color: 'teal', avatar_icon: undefined, active: true },
	{ id: 'p7', name: 'Gio', avatar_color: 'gold', avatar_icon: undefined, active: true },
	{ id: 'p8', name: 'Hanna', avatar_color: 'rose', avatar_icon: undefined, active: true },
	{ id: 'p9', name: 'Ivan', avatar_color: 'slate', avatar_icon: undefined, active: true }
];

export const playerById: Record<string, Player> = Object.fromEntries(players.map((p) => [p.id, p]));

export const session = {
	id: 'proto',
	name: 'Thursday Padel',
	game_mode: 'americano',
	points,
	rounds_total: 9,
	creator_player_id: 'p1',
	players
};

export const round: { number: number; matches: Match[]; bench: string[] } = {
	number: 3,
	matches: [
		{
			id: 'm1',
			court: 1,
			team_a: ['p1', 'p2'],
			team_b: ['p3', 'p4'],
			score: { a: 16, b: 8 },
			live: null
		},
		{
			id: 'm2',
			court: 2,
			team_a: ['p5', 'p6'],
			team_b: ['p7', 'p8'],
			score: null,
			live: { a: 11, b: 7, server: 'a' }
		}
	],
	bench: ['p9']
};

export const timeLeft = '12:34';

export function shortName(name: string): string {
	const parts = name.trim().split(' ');
	if (parts.length === 1) return parts[0];
	return `${parts[0]} ${parts[1][0]}.`;
}

export function teamLabel(ids: [string, string]): string {
	return `${shortName(playerById[ids[0]]?.name ?? '?')} & ${shortName(playerById[ids[1]]?.name ?? '?')}`;
}

export function statusOf(m: Match): 'final' | 'live' | 'upcoming' {
	return m.score ? 'final' : m.live ? 'live' : 'upcoming';
}
