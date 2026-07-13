<script lang="ts" module>
	import type { HTMLAttributes } from 'svelte/elements';
	import type { Snippet } from 'svelte';

	/**
	 * Drawer header section with title and close button.
	 * Typically contains DrawerTitle and DrawerClose.
	 *
	 * @example
	 * <DrawerHeader>
	 *   <DrawerTitle>Settings</DrawerTitle>
	 *   <DrawerClose>×</DrawerClose>
	 * </DrawerHeader>
	 */
	export interface DrawerHeaderProps extends HTMLAttributes<HTMLDivElement> {
		/** Child content (typically DrawerTitle + DrawerClose). */
		children?: Snippet;
	}
</script>

<script lang="ts">
	import { cn } from '$lib/utils.js';

	let { class: className, children, ...restProps }: DrawerHeaderProps = $props();
</script>

<div
	data-slot="drawer-header"
	class={cn('border-border flex flex-col gap-2 border-b px-4 py-3', className)}
	{...restProps}
>
	{@render children?.()}
</div>
