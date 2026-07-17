<script lang="ts">
	import Avatar from '$lib/components/ui/Avatar.svelte';

	let {
		club,
		onclick
	}: {
		club: App.ClubListItem;
		onclick: () => void;
	} = $props();
</script>

<button
	{onclick}
	class="bg-surface-raised hover:bg-border flex w-full items-center gap-3 rounded-2xl px-4 py-3.5 text-left transition-colors"
>
	<Avatar color={club.avatar_color} name={club.name} size="sm" />
	<div class="min-w-0 flex-1">
		<p class="truncate text-sm font-semibold">{club.name}</p>
		<p class="text-text-secondary text-xs">{club.roster_count} members • {club.my_role}</p>
	</div>
	<span class="text-text-secondary text-sm">→</span>
</button>
