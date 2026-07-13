<script lang="ts">
	// Test-only fixture: composes the compound Drawer so render tests can drive it.
	import * as Drawer from './index.js';

	let {
		open = $bindable(false),
		onOpenChange
	}: { open?: boolean; onOpenChange?: (open: boolean) => void } = $props();
</script>

<Drawer.Root bind:open {onOpenChange}>
	<Drawer.Trigger>Open drawer</Drawer.Trigger>
	<Drawer.Content size="md">
		<Drawer.Header>
			<Drawer.Title>Settings</Drawer.Title>
			<Drawer.Description>Adjust your preferences</Drawer.Description>
			<Drawer.Close>Close</Drawer.Close>
		</Drawer.Header>
		<Drawer.Body>Body content</Drawer.Body>
	</Drawer.Content>
</Drawer.Root>
