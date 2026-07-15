/**
 * Number of rounds for a fair Americano tournament — a display-only mirror of
 * the backend's authoritative `americano.TotalRounds`
 * (internal/gamemode/americano/rounds.go). The backend recomputes this from the
 * active roster when the session starts; the frontend uses it to preview the
 * tournament length in the lobby. Keep the two in sync.
 *
 * - No bench (players === courts*4): N-1 rounds covers all unique pairs.
 * - With a bench: smallest multiple of N/gcd(N, benchSize) that is >= N-1, so
 *   everyone sits out equally AND there are enough rounds to be meaningful.
 *
 * Always pass the count of ACTIVE players (removed players are soft-deactivated
 * but remain in the roster), so this tracks joins and leaves.
 *
 * Examples: (10, 2) → 10, (8, 2) → 7, (6, 1) → 6, (9, 2) → 9, (4, 1) → 3.
 */
export function calculateAmericanoRounds(players: number, courts: number): number {
	const benchSize = players - courts * 4;
	if (benchSize <= 0) return players - 1;
	const cycle = players / gcd(players, benchSize);
	const target = players - 1;
	return Math.ceil(target / cycle) * cycle;
}

function gcd(a: number, b: number): number {
	return b === 0 ? a : gcd(b, a % b);
}
