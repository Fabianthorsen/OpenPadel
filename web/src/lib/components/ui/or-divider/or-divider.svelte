<script lang="ts">
	/**
	 * A horizontal rule with a centred label — "──── or ────". Separates two
	 * alternative actions (e.g. invite individuals vs. invite a whole club).
	 */
	let { label, class: className = '' }: { label: string; class?: string } = $props();
</script>

<div class="flex items-center gap-3 {className}">
	<div class="bg-border h-px flex-1"></div>
	<span class="text-text-disabled text-xs">{label}</span>
	<div class="bg-border h-px flex-1"></div>
</div>
