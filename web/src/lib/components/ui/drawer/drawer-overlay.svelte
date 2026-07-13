<script lang="ts">
	import { Dialog as DialogPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils.js';

	let { ref = $bindable(null), class: className, ...restProps }: any = $props();
</script>

<DialogPrimitive.Overlay
	bind:ref
	data-slot="drawer-overlay"
	class={cn('animate-fade-in fixed inset-0 z-40 bg-black/40', className)}
	{...restProps}
/>
