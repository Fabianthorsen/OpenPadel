import { describe, it, expect, vi, afterEach } from 'vitest';
import { render, screen, waitFor, within } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import DrawerFixture from './drawer.fixture.svelte';

/**
 * Render/interaction tests for the custom Drawer (the net-new component added in
 * the Phase 1 design-system PR). These exercise our composition of the Bits UI
 * Dialog parts — open/close wiring, ARIA, and focus — not Bits UI internals.
 *
 * Notes on the jsdom environment:
 * - Close behaviour is asserted via the `onOpenChange` callback rather than DOM
 *   removal: the drawer's exit is CSS-animation driven and jsdom runs no
 *   animations, so "element removed" would be flaky.
 * - `pointerEventsCheck` is disabled because Bits UI makes outside elements
 *   `pointer-events: none` while open; that state leaks across tests in the
 *   shared document, which would otherwise block user-event clicks.
 */
const setup = () => userEvent.setup({ pointerEventsCheck: 0 });

afterEach(() => {
	// Bits UI's scroll-lock / inert-outside leaves a style on <body> that would
	// otherwise carry into the next test's fresh render.
	document.body.style.pointerEvents = '';
});

describe('Drawer (render)', () => {
	it('is closed initially — no dialog in the document', () => {
		render(DrawerFixture);
		expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
	});

	it('opens when the trigger is clicked', async () => {
		const user = setup();
		render(DrawerFixture);

		await user.click(screen.getByRole('button', { name: 'Open drawer' }));

		const dialog = await screen.findByRole('dialog');
		expect(within(dialog).getByText('Settings')).toBeInTheDocument();
		expect(within(dialog).getByText('Body content')).toBeInTheDocument();
	});

	it('wires ARIA: aria-modal and aria-labelledby -> title', async () => {
		const user = setup();
		render(DrawerFixture);
		await user.click(screen.getByRole('button', { name: 'Open drawer' }));

		const dialog = await screen.findByRole('dialog');
		expect(dialog).toHaveAttribute('aria-modal', 'true');

		expect(dialog).toHaveAttribute('aria-labelledby');
		const labelledBy = dialog.getAttribute('aria-labelledby');
		expect(document.getElementById(labelledBy!)).toHaveTextContent('Settings');
	});

	it('moves focus into the dialog on open', async () => {
		const user = setup();
		render(DrawerFixture);
		await user.click(screen.getByRole('button', { name: 'Open drawer' }));

		const dialog = await screen.findByRole('dialog');
		// jsdom has no real tab order, so we assert focus lands inside the dialog
		// (Bits UI focus scope), not a full tab cycle.
		await waitFor(() => expect(dialog.contains(document.activeElement)).toBe(true));
	});

	// Overlay-click dismissal is intentionally asserted as presence only. Closing on
	// an overlay click is Bits UI "interact-outside" behaviour, which keys off trusted
	// pointer events that jsdom does not synthesize (a simulated click never fires the
	// dismiss). The close *contract* is covered by the DrawerClose and Escape tests
	// below; here we just assert our internally-rendered backdrop is present.
	it('renders the backdrop overlay while open', async () => {
		const { baseElement } = render(DrawerFixture, { open: true });
		await screen.findByRole('dialog');

		const overlay = baseElement.querySelector('[data-slot="drawer-overlay"]');
		expect(overlay).not.toBeNull();
		expect(overlay).toHaveAttribute('aria-hidden', 'true');
	});

	it('closes via the DrawerClose button', async () => {
		const user = setup();
		const onOpenChange = vi.fn();
		render(DrawerFixture, { open: true, onOpenChange });

		await user.click(await screen.findByRole('button', { name: 'Close' }));

		expect(onOpenChange).toHaveBeenCalledWith(false);
	});

	it('closes on Escape', async () => {
		const user = setup();
		const onOpenChange = vi.fn();
		render(DrawerFixture, { open: true, onOpenChange });
		await screen.findByRole('dialog');

		await user.keyboard('{Escape}');

		expect(onOpenChange).toHaveBeenCalledWith(false);
	});
});
