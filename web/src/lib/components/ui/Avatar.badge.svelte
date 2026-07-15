<script lang="ts">
	import { UserPlus, Check } from '@lucide/svelte';
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
	aria-label={isContact
		? $_('avatar_contact_added', { values: { name: targetName } })
		: $_('avatar_contact_add', { values: { name: targetName } })}
	class="absolute right-0 bottom-0 flex h-6 w-6 translate-x-1/3 translate-y-1/3
		items-center justify-center rounded-full transition-all
		{isContact
		? 'bg-primary text-primary-foreground'
		: 'bg-primary text-primary-foreground cursor-pointer hover:scale-110 hover:shadow-md'}
		disabled:cursor-default disabled:opacity-60"
	title={isContact ? 'Added' : 'Add contact'}
>
	{#if loading}
		<div class="h-3 w-3 animate-spin rounded-full border-1 border-transparent border-t-current" />
	{:else if isContact}
		<Check size={14} />
	{:else}
		<UserPlus size={14} />
	{/if}
</button>
