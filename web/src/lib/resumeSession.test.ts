import { afterEach, describe, expect, it } from 'vitest';
import {
	clearAccountResumeSession,
	clearResumeSession,
	getResumeSessionId,
	setResumeSession
} from './resumeSession';

afterEach(() => {
	localStorage.clear();
});

describe('resumeSession', () => {
	it('stores and reads the resume pointer', () => {
		setResumeSession('sess-1', null);
		expect(getResumeSessionId()).toBe('sess-1');
	});

	it('clears the pointer', () => {
		setResumeSession('sess-1', 'user-1');
		clearResumeSession();
		expect(getResumeSessionId()).toBeNull();
	});

	it('leaves a guest pointer resumable on logout', () => {
		// A real guest has no admin token — nothing marks the pointer as owned.
		setResumeSession('sess-guest', null);
		clearAccountResumeSession();
		expect(getResumeSessionId()).toBe('sess-guest');
	});

	it('drops an account-owned pointer and its admin token on logout', () => {
		setResumeSession('sess-acct', 'user-1');
		localStorage.setItem('admin_token_sess-acct', 'tok');
		clearAccountResumeSession();
		expect(getResumeSessionId()).toBeNull();
		expect(localStorage.getItem('admin_token_sess-acct')).toBeNull();
	});

	it('drops a pre-marker pointer that still holds an admin token on logout', () => {
		// Pointer written before the owner marker existed: no OWNER_KEY, but the
		// admin token proves it was an account-created session.
		localStorage.setItem('last_session_id', 'sess-legacy');
		localStorage.setItem('admin_token_sess-legacy', 'tok');
		clearAccountResumeSession();
		expect(getResumeSessionId()).toBeNull();
		expect(localStorage.getItem('admin_token_sess-legacy')).toBeNull();
	});

	it('re-pointing to a guest session clears a stale owner marker', () => {
		setResumeSession('sess-acct', 'user-1');
		setResumeSession('sess-guest', null);
		clearAccountResumeSession();
		// The guest pointer must survive — the owner marker was reset.
		expect(getResumeSessionId()).toBe('sess-guest');
	});
});
