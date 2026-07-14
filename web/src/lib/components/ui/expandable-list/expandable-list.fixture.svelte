<script>
	import { ExpandableList } from '.';

	let { items = [], showCount = 5 } = $props();
</script>

<ExpandableList {items} {showCount}>
	{#snippet itemContent(item)}
		<button>{item}</button>
	{/snippet}
</ExpandableList>
