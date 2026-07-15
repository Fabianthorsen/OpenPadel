<script lang="ts">
	import { UserPlus, Check } from 'lucide-svelte';
	import { _ } from 'svelte-i18n';

	let {
		isContact = false,
		onAdd,
		targetName = ''
	}: {
		isContact?: boolean;
		onAdd?: () => Promise<void>;
		targetName?: string;
	} = $props();

	let loading = $state(false);

	async function handleClick() {
		if (isContact || loading || !onAdd) return;
		loading = true;
		try {
			await onAdd();
		} finally {
			loading = false;
		}
	}
</script>

<!-- Add-contact badge overlaid on avatar's bottom-right corner -->
<button
	onclick={handleClick}
	disabled={isContact || loading}
	aria-label={isContact ? $_('avatar_contact_added', { values: { name: targetName } }) : $_('avatar_contact_add', { values: { name: targetName } })}
	class="absolute bottom-0 right-0 translate-x-1/3 translate-y-1/3 flex items-center justify-center
		w-6 h-6 rounded-full transition-all
		{isContact
			? 'bg-primary text-primary-foreground'
			: 'bg-primary hover:scale-110 hover:shadow-md text-primary-foreground cursor-pointer'}
		disabled:opacity-60 disabled:cursor-default"
	title={isContact ? 'Added' : 'Add contact'}
>
	{#if loading}
		<div class="w-3 h-3 border-1 border-transparent border-t-current rounded-full animate-spin" />
	{:else if isContact}
		<Check size={14} />
	{:else}
		<UserPlus size={14} />
	{/if}
</button>
