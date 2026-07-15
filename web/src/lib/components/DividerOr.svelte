<script lang="ts">
	let { label = 'or' }: { label?: string } = $props();
</script>

<div class="flex items-center gap-3">
	<div class="bg-border h-px flex-1"></div>
	<span class="text-text-disabled text-xs">{label}</span>
	<div class="bg-border h-px flex-1"></div>
</div>
