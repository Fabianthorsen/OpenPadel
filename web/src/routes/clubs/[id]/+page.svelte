<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api/client';
	import { auth } from '$lib/auth.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import Footer from '$lib/components/Footer.svelte';
	import { initials } from '$lib/utils';
	import { toast } from 'svelte-sonner';
	import { onMount } from 'svelte';

	interface PageData {
		params: { id: string };
	}

	let club = $state<App.ClubDetail | null>(null);
	let loading = $state(true);
	let error = $state('');

	async function loadClub(clubId: string) {
		if (!auth.token) {
			goto('/auth');
			return;
		}

		try {
			loading = true;
			club = await api.clubs.detail(auth.token, clubId);
		} catch (err: any) {
			if (err.status === 403) {
				error = 'You are not a member of this club';
			} else if (err.status === 404) {
				error = 'Club not found';
			} else {
				error = 'Failed to load club';
			}
			toast.error(error);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		const clubId = window.location.pathname.split('/')[2];
		if (clubId) {
			loadClub(clubId);
		}
	});
</script>

{#if loading}
	<div class="flex items-center justify-center min-h-screen">
		<p>Loading club...</p>
	</div>
{:else if error}
	<div class="flex flex-col items-center justify-center min-h-screen gap-4">
		<p class="text-red-500">{error}</p>
		<Button onclick={() => goto('/')}>Back to home</Button>
	</div>
{:else if club}
	<div class="flex flex-col h-screen">
		<div class="flex-1 overflow-y-auto pb-20">
			<!-- Club Header -->
			<div class="bg-gradient-to-b from-slate-100 to-white dark:from-slate-900 dark:to-slate-950 p-6 border-b">
				<div class="flex items-start gap-4 max-w-2xl mx-auto">
					<div
						class="w-16 h-16 rounded-lg flex items-center justify-center text-xl font-bold text-white"
						style="background-color: var(--avatar-color-{club.club.avatar_color}, #666)"
					>
						{initials(club.club.name)}
					</div>
					<div class="flex-1">
						<h1 class="text-2xl font-bold">{club.club.name}</h1>
						{#if club.club.description}
							<p class="text-sm text-slate-600 dark:text-slate-400 mt-1">{club.club.description}</p>
						{/if}
						<p class="text-xs text-slate-500 dark:text-slate-500 mt-2">{club.roster_count} members</p>
					</div>
				</div>
			</div>

			<!-- Roster -->
			<div class="max-w-2xl mx-auto p-6">
				<h2 class="text-lg font-semibold mb-4">Members ({club.roster_count})</h2>
				<div class="space-y-2">
					{#each club.members as member}
						<div
							class="flex items-center gap-3 p-3 rounded-lg border bg-card hover:bg-slate-50 dark:hover:bg-slate-900"
						>
							<div
								class="w-10 h-10 rounded flex items-center justify-center text-xs font-bold text-white flex-shrink-0"
								style="background-color: var(--avatar-color-{member.avatar_color}, #666)"
							>
								{initials(member.display_name)}
							</div>
							<div class="flex-1 min-w-0">
								<p class="font-medium text-sm">{member.display_name}</p>
								<p class="text-xs text-slate-500 capitalize">{member.role}</p>
							</div>
						</div>
					{/each}
				</div>
			</div>
		</div>

		<Footer />
	</div>
{/if}

<style>
	:root {
		--avatar-color-forest: #2d5016;
		--avatar-color-blue: #0066cc;
		--avatar-color-green: #00aa00;
	}
</style>
