<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api/client';
	import { auth } from '$lib/auth.svelte';
	import { Button } from '$lib/components/ui/button';
	import Footer from '$lib/components/Footer.svelte';
	import { initials } from '$lib/utils';
	import { toast } from 'svelte-sonner';
	import { onMount } from 'svelte';

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
		<Button onclick={() => goto('/profile')}>Back to profile</Button>
	</div>
{:else if club}
	<div class="flex flex-col h-screen">
		<div class="flex-1 overflow-y-auto pb-20">
			<!-- Header with back button and member count -->
			<div class="pt-safe-page bg-surface-raised sticky top-0 z-10 flex items-center justify-between px-6 py-4 border-b">
				<button
					onclick={() => goto('/profile')}
					class="text-text-secondary hover:text-text-primary transition-colors text-lg"
				>
					‹
				</button>
				<h1 class="flex-1 text-center text-lg font-bold">{club.club.name}</h1>
				<div class="w-8 text-center">
					<p class="text-xs font-semibold text-text-secondary">{club.roster_count}</p>
				</div>
			</div>

			<!-- Club Details -->
			<div class="px-6 py-6 border-b">
				<div class="flex items-start gap-4">
					<div
						class="w-14 h-14 rounded-lg flex items-center justify-center text-lg font-bold text-white flex-shrink-0"
						style="background-color: var(--avatar-color-{club.club.avatar_color}, #666)"
					>
						{initials(club.club.name)}
					</div>
					<div class="flex-1 min-w-0">
						{#if club.club.description}
							<p class="text-sm text-text-secondary break-words">{club.club.description}</p>
						{/if}
						<p class="text-xs text-text-disabled mt-2">{club.roster_count} members</p>
					</div>
				</div>
			</div>

			<!-- Roster -->
			<div class="px-6 py-6">
				<p class="text-text-secondary text-[11px] font-bold tracking-[0.1em] uppercase mb-4">Members</p>
				<div class="space-y-2">
					{#each club.members as member}
						<div
							class="flex items-center gap-3 p-3 rounded-lg border bg-card hover:bg-surface-raised transition-colors"
						>
							<div
								class="w-10 h-10 rounded flex items-center justify-center text-xs font-bold text-white flex-shrink-0"
								style="background-color: var(--avatar-color-{member.avatar_color}, #666)"
							>
								{initials(member.display_name)}
							</div>
							<div class="flex-1 min-w-0">
								<p class="font-medium text-sm">{member.display_name}</p>
								<p class="text-xs text-text-secondary capitalize">{member.role}</p>
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
