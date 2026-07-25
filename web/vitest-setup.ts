import '@testing-library/jest-dom/vitest';
import { afterEach, vi } from 'vitest';
import { cleanup } from '@testing-library/svelte';

// Bits UI's presence layer and body-scroll-lock tear down on a promise/microtask
// tail that outlives Testing Library's synchronous unmount. If that tail runs
// after jsdom disposes the environment it throws "document is not defined" as an
// unhandled error and fails the run. Unmount, then yield so the deferred teardown
// settles while `document` still exists.
//
// A scroll-locking layer (Drawer/Dialog/Select) defers its resetBodyStyle on a
// ~24ms timer and leaves `overflow: hidden` on <body> until it fires, so when a
// lock is still pending we wait past that delay; every other test keeps the
// zero-cost single-macrotask yield.
afterEach(async () => {
	cleanup();
	const lockPending = document.body.style.overflow === 'hidden';
	await new Promise((resolve) => setTimeout(resolve, lockPending ? 30 : 0));
});

// jsdom lacks several browser APIs that Bits UI (Dialog focus scope, presence
// layer, floating positioning) touches. Stub them so component tests can render
// without throwing. These are test-only shims, not app behaviour.

window.matchMedia ??= vi.fn().mockImplementation((query: string) => ({
	matches: false,
	media: query,
	onchange: null,
	addEventListener: vi.fn(),
	removeEventListener: vi.fn(),
	addListener: vi.fn(),
	removeListener: vi.fn(),
	dispatchEvent: vi.fn()
})) as unknown as typeof window.matchMedia;

class ObserverStub {
	observe() {}
	unobserve() {}
	disconnect() {}
}
window.ResizeObserver ??= ObserverStub as unknown as typeof ResizeObserver;
window.IntersectionObserver ??= ObserverStub as unknown as typeof IntersectionObserver;

Element.prototype.scrollIntoView ??= vi.fn();

// Pointer capture — jsdom has no PointerEvent capture model; Bits UI's Select
// trigger calls these on pointerdown. No-op shims let the dropdown open in tests.
Element.prototype.hasPointerCapture ??= vi.fn().mockReturnValue(false);
Element.prototype.setPointerCapture ??= vi.fn();
Element.prototype.releasePointerCapture ??= vi.fn();

// Web Animations API — jsdom has no implementation; Bits UI's presence layer
// probes for running animations before unmounting.
Element.prototype.animate ??= vi.fn().mockReturnValue({
	cancel: vi.fn(),
	finished: Promise.resolve(),
	onfinish: null
}) as unknown as Element['animate'];
Element.prototype.getAnimations ??= vi.fn().mockReturnValue([]);
