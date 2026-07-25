<script lang="ts">
	import { Dialog as SheetPrimitive } from 'bits-ui';
	import { guardBodyLock } from '$lib/overlayGuard.svelte';

	let { open = $bindable(false), ...restProps }: SheetPrimitive.RootProps = $props();

	// Safety net for the bits-ui body-scroll-lock leak that can freeze the page
	// after the sheet closes. See $lib/overlayGuard.
	guardBodyLock(() => open);
</script>

<SheetPrimitive.Root bind:open {...restProps} />
