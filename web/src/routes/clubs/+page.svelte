<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api/client';
	import { auth } from '$lib/auth.svelte';
	import { Button } from '$lib/components/ui/button';
	import ClubCard from '$lib/components/ClubCard.svelte';
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
			toast.error('Failed to load clubs');
			clubs = [];
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadClubs();
	});
</script>

<div class="flex h-screen flex-col">
	<div class="flex-1 overflow-y-auto pb-20">
		<div class="mx-auto max-w-2xl p-6">
			<div class="mb-6 flex items-center justify-between">
				<h1 class="text-2xl font-bold">My Clubs</h1>
				<Button onclick={() => goto('/')}>←</Button>
			</div>

			{#if loading}
				<p class="text-center text-slate-500">Loading clubs...</p>
			{:else if clubs.length === 0}
				<div class="py-12 text-center">
					<p class="mb-4 text-slate-500">No clubs yet. Create one to get started!</p>
					<Button onclick={() => goto('/')}>Create a Club</Button>
				</div>
			{:else}
				<div class="space-y-3">
					{#each clubs as club}
						<ClubCard {club} onclick={() => goto(`/clubs/${club.id}`)} />
					{/each}
				</div>
			{/if}
		</div>
	</div>

	<Footer />
</div>
