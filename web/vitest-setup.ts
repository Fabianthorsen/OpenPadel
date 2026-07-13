import '@testing-library/jest-dom/vitest';
import { vi } from 'vitest';

// jsdom lacks several browser APIs that Bits UI (Dialog focus scope, presence
// layer, floating positioning) touches. Stub them so component tests can render
// without throwing. These are test-only shims, not app behaviour.

if (!window.matchMedia) {
	window.matchMedia = vi.fn().mockImplementation((query: string) => ({
		matches: false,
		media: query,
		onchange: null,
		addEventListener: vi.fn(),
		removeEventListener: vi.fn(),
		addListener: vi.fn(),
		removeListener: vi.fn(),
		dispatchEvent: vi.fn()
	}));
}

class ObserverStub {
	observe() {}
	unobserve() {}
	disconnect() {}
}
window.ResizeObserver ??= ObserverStub as unknown as typeof ResizeObserver;
window.IntersectionObserver ??= ObserverStub as unknown as typeof IntersectionObserver;

Element.prototype.scrollIntoView ??= vi.fn();

// Web Animations API — jsdom has no implementation; Bits UI's presence layer
// probes for running animations before unmounting.
Element.prototype.animate ??= vi.fn().mockReturnValue({
	cancel: vi.fn(),
	finished: Promise.resolve(),
	onfinish: null
}) as unknown as Element['animate'];
Element.prototype.getAnimations ??= vi.fn().mockReturnValue([]);
