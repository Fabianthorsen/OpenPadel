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
	class="bg-card flex w-full items-center gap-4 rounded-lg border p-4 text-left transition-colors hover:bg-slate-50 dark:hover:bg-slate-900"
>
	<Avatar color={club.avatar_color} name={club.name} size="md" />
	<div class="min-w-0 flex-1">
		<h2 class="text-sm font-semibold">{club.name}</h2>
		<p class="text-xs text-slate-500">{club.roster_count} members • {club.my_role}</p>
	</div>
	<span class="text-slate-400">→</span>
</button>
