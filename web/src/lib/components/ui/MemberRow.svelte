<script lang="ts">
	import { _ } from 'svelte-i18n';
	import Avatar from '$lib/components/ui/Avatar.svelte';

	let {
		member
	}: {
		member: App.ClubMember;
	} = $props();
</script>

<div class="flex items-center gap-3 rounded-xl px-2 py-2">
	<Avatar color={member.avatar_color} name={member.display_name} size="md" />
	<div class="flex min-w-0 flex-1 flex-col gap-0.5">
		<p class="text-text-primary truncate text-sm font-semibold">{member.display_name}</p>
		<p class="text-text-secondary text-xs">{$_(`club_role_${member.role}`)}</p>
	</div>
</div>
