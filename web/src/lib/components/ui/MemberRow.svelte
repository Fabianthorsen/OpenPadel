<script lang="ts">
	import Avatar from '$lib/components/ui/Avatar.svelte';

	let {
		member
	}: {
		member: App.ClubMember;
	} = $props();
</script>

<div class="bg-surface-raised flex items-center gap-3 rounded-2xl px-4 py-3.5">
	<Avatar color={member.avatar_color} name={member.display_name} size="sm" />
	<div class="min-w-0 flex-1">
		<p class="truncate text-sm font-semibold">{member.display_name}</p>
		<p class="text-text-secondary text-xs capitalize">{member.role}</p>
	</div>
</div>
