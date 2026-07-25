/**
 * Safety net for the bits-ui body-scroll-lock leak.
 *
 * While a modal (Dialog / Drawer / Sheet) is open, bits-ui locks the page by
 * setting `pointer-events: none` and `overflow: hidden` on `<body>`, and only
 * restores it when the modal's Content component unmounts. That unmount is
 * deferred until the exit animation reports done via the Web Animations API
 * (`node.getAnimations()` → `Promise.all(a.finished)`). On a backgrounded or
 * throttled mobile tab those promises can stall, the Content never unmounts, and
 * the body lock leaks — every click is swallowed and the page appears frozen.
 *
 * These helpers re-check shortly after a modal closes and release the lock if it
 * is still applied while no modal is actually open. Wire {@link guardBodyLock}
 * into a modal root so every usage is covered without per-call-site code.
 */

/** An open bits-ui modal keeps `data-state="open"` on its `role="dialog"` node. */
const OPEN_MODAL_SELECTOR =
	'[role="dialog"][data-state="open"],[role="alertdialog"][data-state="open"]';

/**
 * How long to wait after a close before checking for a stuck lock. Comfortably
 * past the longest exit animation (~300ms) plus bits-ui's own ~24ms deferred
 * cleanup, so the normal path has already cleared and this becomes a no-op.
 */
const RELEASE_DELAY_MS = 500;

/** Body properties bits-ui's scroll lock applies; cleared together on release. */
const LOCK_PROPS = [
	'pointer-events',
	'overflow',
	'padding-right',
	'margin-right',
	'--scrollbar-width'
];

/**
 * Clear a leaked body scroll-lock, but only when it is genuinely stuck: the body
 * is still locked *and* no modal is actually open. Safe to call at any time —
 * it's a no-op unless both conditions hold, so it never fights a legitimately
 * open dialog or a body that some other code styled.
 */
export function releaseStuckBodyLock(): void {
	if (typeof document === 'undefined') return;
	const body = document.body;
	// Only the pointer-events lock actually freezes the page; if it isn't set,
	// there is nothing stuck to clean up.
	if (body.style.pointerEvents !== 'none') return;
	// A genuinely open modal still needs the lock — leave it alone.
	if (document.querySelector(OPEN_MODAL_SELECTOR)) return;
	for (const prop of LOCK_PROPS) body.style.removeProperty(prop);
}

/**
 * Schedule a stuck-lock check to run after {@link RELEASE_DELAY_MS}. Returns a
 * canceller. Kept runes-free so it can be unit-tested with fake timers.
 */
export function scheduleBodyLockRelease(): () => void {
	const id = setTimeout(releaseStuckBodyLock, RELEASE_DELAY_MS);
	return () => clearTimeout(id);
}

/**
 * Watch an overlay's `open` flag and run the safety net after each close. Call
 * once during a modal root's initialisation, e.g. `guardBodyLock(() => open)`.
 * Only genuine open→close transitions are guarded, so a modal that mounts closed
 * schedules nothing.
 */
export function guardBodyLock(isOpen: () => boolean): void {
	let armed = false;
	$effect(() => {
		if (isOpen()) {
			armed = true;
			return;
		}
		if (!armed) return;
		// Just closed: bits-ui should restore the body itself, but re-check in case
		// its animation-gated cleanup never runs. The returned canceller fires if
		// the modal reopens (or the root is destroyed) before the check.
		return scheduleBodyLockRelease();
	});
}
