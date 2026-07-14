<script lang="ts">
	import { Section } from './index';

	let {
		title = 'History',
		collapsible = true,
		open = true
	}: { title?: string; collapsible?: boolean; open?: boolean } = $props();
</script>

<Section {title} {collapsible} {open}>
	{#snippet children()}
		<p>section body content</p>
	{/snippet}
</Section>
