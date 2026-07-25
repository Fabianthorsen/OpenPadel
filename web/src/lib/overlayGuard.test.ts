import { describe, it, expect, vi, afterEach } from 'vitest';
import { render } from '@testing-library/svelte';
import { releaseStuckBodyLock, scheduleBodyLockRelease } from './overlayGuard.svelte';
import OverlayGuardFixture from './overlayGuard.fixture.svelte';

/**
 * Regression coverage for the bits-ui body-scroll-lock leak: while a modal is
 * open bits-ui sets `pointer-events: none` on <body> and only restores it when
 * the modal's Content unmounts after its exit animation resolves. If that never
 * happens (rAF/animations throttled on a backgrounded mobile tab) the lock leaks
 * and the whole page stops responding to clicks. The guard re-checks after each
 * close and releases a stuck lock. See $lib/overlayGuard.
 */

/** Reproduce the leaked lock bits-ui leaves on <body>. */
function lockBody(): void {
	document.body.style.pointerEvents = 'none';
	document.body.style.overflow = 'hidden';
}

afterEach(() => {
	// Restore real timers before the global afterEach (which may await a timer).
	vi.useRealTimers();
	document.body.removeAttribute('style');
	document.body.innerHTML = '';
});

describe('releaseStuckBodyLock', () => {
	it('clears a leaked lock when no modal is open', () => {
		lockBody();

		releaseStuckBodyLock();

		expect(document.body.style.pointerEvents).toBe('');
		expect(document.body.style.overflow).toBe('');
	});

	it('leaves the lock in place while a modal is still open', () => {
		lockBody();
		document.body.innerHTML = '<div role="dialog" data-state="open"></div>';

		releaseStuckBodyLock();

		expect(document.body.style.pointerEvents).toBe('none');
	});

	it('ignores a body that is not locked (never touches unrelated styles)', () => {
		document.body.style.paddingRight = '10px';

		releaseStuckBodyLock();

		expect(document.body.style.paddingRight).toBe('10px');
	});
});

describe('scheduleBodyLockRelease', () => {
	it('releases a stuck lock after the delay', () => {
		vi.useFakeTimers();
		lockBody();

		scheduleBodyLockRelease();
		expect(document.body.style.pointerEvents).toBe('none'); // not yet
		vi.advanceTimersByTime(600);

		expect(document.body.style.pointerEvents).toBe('');
	});

	it('can be cancelled before it fires', () => {
		vi.useFakeTimers();
		lockBody();

		const cancel = scheduleBodyLockRelease();
		cancel();
		vi.advanceTimersByTime(600);

		expect(document.body.style.pointerEvents).toBe('none');
	});
});

describe('guardBodyLock (via fixture)', () => {
	it('releases a leaked lock shortly after the overlay closes', async () => {
		vi.useFakeTimers();
		const { rerender } = render(OverlayGuardFixture, { open: false });

		// Open, then simulate bits-ui leaving <body> locked after the close.
		await rerender({ open: true });
		lockBody();
		await rerender({ open: false });

		// Still locked immediately; the guard waits before intervening.
		expect(document.body.style.pointerEvents).toBe('none');
		vi.advanceTimersByTime(600);

		expect(document.body.style.pointerEvents).toBe('');
	});

	it('does not release while the overlay is still open', async () => {
		vi.useFakeTimers();
		const { rerender } = render(OverlayGuardFixture, { open: false });

		await rerender({ open: true });
		lockBody();
		vi.advanceTimersByTime(600);

		expect(document.body.style.pointerEvents).toBe('none');
	});

	it('does not schedule a release for a modal that only ever mounts closed', async () => {
		vi.useFakeTimers();
		render(OverlayGuardFixture, { open: false });

		// An unrelated leaked lock must not be cleared by a modal that never opened.
		lockBody();
		vi.advanceTimersByTime(600);

		expect(document.body.style.pointerEvents).toBe('none');
	});
});
