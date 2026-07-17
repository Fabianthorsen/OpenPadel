<script lang="ts">
	import Avatar from '$lib/components/ui/Avatar.svelte';

	let {
		member
	}: {
		member: App.ClubMember;
	} = $props();
</script>

<div
	class="bg-card hover:bg-surface-raised flex items-center gap-3 rounded-lg border p-3 transition-colors"
>
	<Avatar color={member.avatar_color} name={member.display_name} size="md" />
	<div class="min-w-0 flex-1">
		<p class="text-sm font-medium">{member.display_name}</p>
		<p class="text-text-secondary text-xs capitalize">{member.role}</p>
	</div>
</div>
