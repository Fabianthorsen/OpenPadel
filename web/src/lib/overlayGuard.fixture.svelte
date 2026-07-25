<script lang="ts">
	// Test-only fixture: drives guardBodyLock via a plain `open` prop so the
	// open→close transition logic can be exercised without a real bits-ui modal
	// (jsdom runs no animations, so the lock never actually leaks on its own).
	import { guardBodyLock } from './overlayGuard.svelte';

	let { open = false }: { open?: boolean } = $props();

	guardBodyLock(() => open);
</script>
