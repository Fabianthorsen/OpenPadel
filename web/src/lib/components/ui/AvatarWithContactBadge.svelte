<script lang="ts">
	import { UserPlus, Check } from 'lucide-svelte';
	import Avatar from './Avatar.svelte';
	import { _ } from 'svelte-i18n';

	let {
		icon = '',
		color = '',
		name = '',
		size = 'md' as 'sm' | 'md' | 'lg' | 'xl',
		ring = '',
		showBadge = false,
		isContact = false,
		onAdd,
		targetName = ''
	}: {
		icon?: string;
		color?: string;
		name?: string;
		size?: 'sm' | 'md' | 'lg' | 'xl';
		ring?: string;
		showBadge?: boolean;
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

{#if showBadge}
	<div class="relative inline-flex shrink-0">
		<Avatar {icon} {color} {name} {size} {ring} />
		<button
			onclick={handleClick}
			disabled={isContact || loading}
			aria-label={isContact
				? $_('avatar_contact_added', { values: { name: targetName } })
				: $_('avatar_contact_add', { values: { name: targetName } })}
			class="absolute right-0 bottom-0 flex h-6 w-6 translate-x-1/3 translate-y-1/3
				items-center justify-center rounded-full border-2 border-[var(--color-border-strong)] bg-[var(--color-surface-raised)] text-[var(--color-text-primary)] shadow-md transition-all
				{isContact ? '' : 'cursor-pointer hover:scale-110'}
				disabled:cursor-default disabled:opacity-60"
		>
			{#if loading}
				<div
					class="h-3 w-3 animate-spin rounded-full border-1 border-transparent border-t-current"
				></div>
			{:else if isContact}
				<Check size={14} strokeWidth={3}></Check>
			{:else}
				<UserPlus size={14} strokeWidth={3}></UserPlus>
			{/if}
		</button>
	</div>
{:else}
	<Avatar {icon} {color} {name} {size} {ring} />
{/if}
