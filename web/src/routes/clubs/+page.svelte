<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api/client';
	import { auth } from '$lib/auth.svelte';
	import { Button } from '$lib/components/ui/button';
	import Footer from '$lib/components/Footer.svelte';
	import { initials } from '$lib/utils';
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

<div class="flex flex-col h-screen">
	<div class="flex-1 overflow-y-auto pb-20">
		<div class="max-w-2xl mx-auto p-6">
			<div class="flex items-center justify-between mb-6">
				<h1 class="text-2xl font-bold">My Clubs</h1>
				<Button onclick={() => goto('/')}>←</Button>
			</div>

			{#if loading}
				<p class="text-center text-slate-500">Loading clubs...</p>
			{:else if clubs.length === 0}
				<div class="text-center py-12">
					<p class="text-slate-500 mb-4">No clubs yet. Create one to get started!</p>
					<Button onclick={() => goto('/')}>Create a Club</Button>
				</div>
			{:else}
				<div class="space-y-3">
					{#each clubs as club}
						<button
							onclick={() => goto(`/clubs/${club.id}`)}
							class="w-full text-left flex items-center gap-4 p-4 rounded-lg border bg-card hover:bg-slate-50 dark:hover:bg-slate-900 transition-colors"
						>
							<div
								class="w-12 h-12 rounded flex items-center justify-center text-sm font-bold text-white flex-shrink-0"
								style="background-color: var(--avatar-color-{club.avatar_color}, #666)"
							>
								{initials(club.name)}
							</div>
							<div class="flex-1 min-w-0">
								<h2 class="font-semibold text-sm">{club.name}</h2>
								<p class="text-xs text-slate-500">{club.roster_count} members • {club.my_role}</p>
							</div>
							<span class="text-slate-400">→</span>
						</button>
					{/each}
				</div>
			{/if}
		</div>
	</div>

	<Footer />
</div>

<style>
	:root {
		--avatar-color-forest: #2d5016;
		--avatar-color-blue: #0066cc;
		--avatar-color-green: #00aa00;
	}
</style>
