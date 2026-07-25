// The device-local "resume this session" pointer that powers the home-page
// rejoin affordance. One place so the localStorage keys aren't hand-built
// across components.
//
// Resuming follows identity (#252). A guest joins by name only and legitimately
// rejoins from this device-local pointer — that is by design. A registered user,
// though, must only resume their own in-progress sessions while authenticated:
// when the pointer was set by a logged-in account we record the owner, and on
// logout that account-owned pointer (and its admin token) is forgotten so a
// logged-out device no longer carries a registered user's session context.

const RESUME_KEY = 'last_session_id';
const OWNER_KEY = 'last_session_user_id';

const adminTokenKey = (sessionId: string) => `admin_token_${sessionId}`;

/**
 * Point the resume affordance at this session. Pass the authenticated user's id
 * as `ownerUserId` when set by a logged-in account, or null for a guest join.
 */
export function setResumeSession(sessionId: string, ownerUserId: string | null): void {
	localStorage.setItem(RESUME_KEY, sessionId);
	if (ownerUserId) {
		localStorage.setItem(OWNER_KEY, ownerUserId);
	} else {
		localStorage.removeItem(OWNER_KEY);
	}
}

/** The session id the resume affordance should target, or null. */
export function getResumeSessionId(): string | null {
	if (typeof localStorage === 'undefined') return null;
	return localStorage.getItem(RESUME_KEY);
}

/** Forget the resume pointer (e.g. the session has ended). */
export function clearResumeSession(): void {
	if (typeof localStorage === 'undefined') return;
	localStorage.removeItem(RESUME_KEY);
	localStorage.removeItem(OWNER_KEY);
}

/**
 * On logout: a resume pointer that belongs to a registered account must not
 * survive as a device-local affordance. Drop the pointer and the account's
 * admin token for that session (login re-recovers it from the server). A guest
 * pointer is left untouched.
 *
 * A pointer is account-owned when we recorded an owner, or — covering pointers
 * written before the owner marker existed — when we still hold an admin token
 * for it: admin tokens are only ever minted by creating a session while logged
 * in, which is exactly the leaked resume path (#252).
 */
export function clearAccountResumeSession(): void {
	if (typeof localStorage === 'undefined') return;
	const sessionId = localStorage.getItem(RESUME_KEY);
	if (!sessionId) return;
	const owned =
		!!localStorage.getItem(OWNER_KEY) || !!localStorage.getItem(adminTokenKey(sessionId));
	if (!owned) return;
	localStorage.removeItem(adminTokenKey(sessionId));
	clearResumeSession();
}
