<script lang="ts" module>
	import type { HTMLAttributes } from 'svelte/elements';
	import type { Snippet } from 'svelte';

	/**
	 * Drawer body section. Scrollable main content area.
	 *
	 * @example
	 * <DrawerBody>Form fields and content here</DrawerBody>
	 */
	export interface DrawerBodyProps extends HTMLAttributes<HTMLDivElement> {
		/** Child content (form fields, text, etc.). */
		children?: Snippet;
	}
</script>

<script lang="ts">
	import { cn } from '$lib/utils.js';

	let { class: className, children, ...restProps }: DrawerBodyProps = $props();
</script>

<div
	data-slot="drawer-body"
	class={cn('flex-1 overflow-y-auto px-4 py-3', className)}
	{...restProps}
>
	{@render children?.()}
</div>
