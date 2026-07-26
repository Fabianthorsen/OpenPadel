<script lang="ts">
	import { _ } from 'svelte-i18n';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import { ChevronRight } from '@lucide/svelte';

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
	class="group hover:bg-surface-raised flex w-full items-center gap-3 rounded-xl py-2 pr-1 pl-2 text-left transition-colors"
>
	<Avatar color={club.avatar_color} name={club.name} size="md" />
	<div class="flex min-w-0 flex-1 flex-col gap-0.5">
		<p class="text-text-primary truncate text-sm font-semibold">{club.name}</p>
		<p class="text-text-secondary truncate text-xs">
			{club.roster_count}
			{club.roster_count === 1 ? $_('club_member') : $_('club_members')} · {$_(
				`club_role_${club.my_role}`
			)}
		</p>
	</div>
	<span
		class="text-text-disabled group-hover:text-primary shrink-0 transition-[color,transform] group-hover:translate-x-0.5"
		aria-hidden="true"
	>
		<ChevronRight size={18} strokeWidth={2.25} />
	</span>
</button>
