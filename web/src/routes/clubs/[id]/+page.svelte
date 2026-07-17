<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api/client';
	import { auth } from '$lib/auth.svelte';
	import { Button } from '$lib/components/ui/button';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import MemberRow from '$lib/components/MemberRow.svelte';
	import Footer from '$lib/components/Footer.svelte';
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
	<div class="flex min-h-screen items-center justify-center">
		<p>Loading club...</p>
	</div>
{:else if error}
	<div class="flex min-h-screen flex-col items-center justify-center gap-4">
		<p class="text-red-500">{error}</p>
		<Button onclick={() => goto('/profile')}>Back to profile</Button>
	</div>
{:else if club}
	<div class="flex h-screen flex-col">
		<div class="flex-1 overflow-y-auto pb-20">
			<!-- Header with back button and member count -->
			<div
				class="pt-safe-page bg-surface-raised sticky top-0 z-10 flex items-center justify-between border-b px-6 py-4"
			>
				<button
					onclick={() => goto('/profile')}
					class="text-text-secondary hover:text-text-primary text-lg transition-colors"
				>
					‹
				</button>
				<h1 class="flex-1 text-center text-lg font-bold">{club.club.name}</h1>
				<div class="w-8 text-center">
					<p class="text-text-secondary text-xs font-semibold">{club.roster_count}</p>
				</div>
			</div>

			<!-- Club Details -->
			<div class="border-b px-6 py-6">
				<div class="flex items-start gap-4">
					<Avatar color={club.club.avatar_color} name={club.club.name} size="lg" />
					<div class="min-w-0 flex-1">
						{#if club.club.description}
							<p class="text-text-secondary text-sm break-words">{club.club.description}</p>
						{/if}
						<p class="text-text-disabled mt-2 text-xs">{club.roster_count} members</p>
					</div>
				</div>
			</div>

			<!-- Roster -->
			<div class="px-6 py-6">
				<p class="text-text-secondary mb-4 text-[11px] font-bold tracking-[0.1em] uppercase">
					Members
				</p>
				<div class="space-y-2">
					{#each club.members as member}
						<MemberRow {member} />
					{/each}
				</div>
			</div>
		</div>

		<Footer />
	</div>
{/if}
