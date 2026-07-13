<script lang="ts" module>
	import type { HTMLAttributes } from 'svelte/elements';
	import type { Snippet } from 'svelte';

	export interface DrawerFooterProps extends HTMLAttributes<HTMLDivElement> {
		children?: Snippet;
	}
</script>

<script lang="ts">
	import { cn } from '$lib/utils.js';

	let { class: className, children, ...restProps }: DrawerFooterProps = $props();
</script>

<div
	data-slot="drawer-footer"
	class={cn('border-border flex gap-3 border-t px-4 py-3', className)}
	{...restProps}
>
	{@render children?.()}
</div>
