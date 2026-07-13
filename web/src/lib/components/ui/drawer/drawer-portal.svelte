<script lang="ts" module>
	import type { Snippet } from 'svelte';

	/**
	 * Drawer portal wrapper. Renders drawer content at document root (not in DOM context).
	 * Used internally by DrawerContent; usually not needed directly.
	 *
	 * @example
	 * <DrawerPortal>
	 *   <DrawerOverlay />
	 *   <DrawerContent>Content</DrawerContent>
	 * </DrawerPortal>
	 */
	export interface DrawerPortalProps {
		children?: Snippet;
	}
</script>

<script lang="ts">
	import { Dialog as DialogPrimitive } from 'bits-ui';

	let { children }: DrawerPortalProps = $props();
</script>

<DialogPrimitive.Portal>
	{@render children?.()}
</DialogPrimitive.Portal>
