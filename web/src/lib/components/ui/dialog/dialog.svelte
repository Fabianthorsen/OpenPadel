<script lang="ts">
	import { Dialog as DialogPrimitive } from 'bits-ui';
	import { guardBodyLock } from '$lib/overlayGuard.svelte';

	let { open = $bindable(false), ...restProps }: DialogPrimitive.RootProps = $props();

	// Safety net for the bits-ui body-scroll-lock leak that can freeze the page
	// after the dialog closes. See $lib/overlayGuard.
	guardBodyLock(() => open);
</script>

<DialogPrimitive.Root bind:open {...restProps} />
