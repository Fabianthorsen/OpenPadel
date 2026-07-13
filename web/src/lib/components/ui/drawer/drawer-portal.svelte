<script lang="ts">
	import { Dialog as DialogPrimitive } from 'bits-ui';

	let { children }: any = $props();
</script>

<DialogPrimitive.Portal>
	{@render children?.()}
</DialogPrimitive.Portal>
