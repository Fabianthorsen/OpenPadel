<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api/client';
	import { auth } from '$lib/auth.svelte';
	import { _ } from 'svelte-i18n';
	import { Button } from '$lib/components/ui/button';
	import ClubCard from '$lib/components/ui/ClubCard.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import { Spinner } from '$lib/components/ui/spinner';
	import Footer from '$lib/components/Footer.svelte';
	import { toast } from 'svelte-sonner';
	import { onMount } from 'svelte';

	let clubs = $state<App.ClubListItem[]>([]);
	let loading = $state(true);

	async function loadClubs() {
		if (!auth.token) {
			goto('/auth');
			return;
		}

		try {
			loading = true;
			clubs = await api.clubs.list(auth.token);
		} catch (err) {
			console.error('Failed to load clubs:', err);
			toast.error($_('profile_clubs_load_error'));
			clubs = [];
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadClubs();
	});
</script>

<main class="pt-safe-page mx-auto max-w-[480px] space-y-8 px-6 pb-10">
	<PageHeader title={$_('clubs_title')} backHref="/" />

	{#if loading}
		<div class="flex justify-center py-12">
			<Spinner label={$_('profile_clubs_loading')} />
		</div>
	{:else if clubs.length === 0}
		<div class="space-y-4 py-8 text-center">
			<p class="text-text-secondary text-sm">{$_('clubs_empty')}</p>
			<Button onclick={() => goto('/')}>{$_('profile_create_club')}</Button>
		</div>
	{:else}
		<div class="space-y-2">
			{#each clubs as club (club.id)}
				<ClubCard {club} onclick={() => goto(`/clubs/${club.id}`)} />
			{/each}
		</div>
	{/if}

	<Footer />
</main>
