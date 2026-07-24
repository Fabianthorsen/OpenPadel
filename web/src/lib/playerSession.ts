// Client-side record of "which player am I in this session", persisted at join.
// The player id identifies the row; the player_token is the per-player secret
// required to self-remove a guest (#241). Kept in one place so the localStorage
// keys aren't hand-built across components.

const idKey = (sessionId: string) => `player_id_${sessionId}`;
const tokenKey = (sessionId: string) => `player_token_${sessionId}`;

/** Persist the joined player's id and (if present) self-removal secret. */
export function savePlayerSession(sessionId: string, player: App.Player): void {
	localStorage.setItem(idKey(sessionId), player.id);
	if (player.player_token) {
		localStorage.setItem(tokenKey(sessionId), player.player_token);
	}
}

/** The stored player id for this session, or null. */
export function getPlayerId(sessionId: string): string | null {
	if (typeof localStorage === 'undefined') return null;
	return localStorage.getItem(idKey(sessionId));
}

/** The stored self-removal secret for this session, or '' if none. */
export function getPlayerToken(sessionId: string): string {
	return localStorage.getItem(tokenKey(sessionId)) ?? '';
}

/** Forget this session's player id and secret (on leave). */
export function clearPlayerSession(sessionId: string): void {
	localStorage.removeItem(idKey(sessionId));
	localStorage.removeItem(tokenKey(sessionId));
}
