<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api/client';
	import { auth } from '$lib/auth.svelte';
	import { Button } from '$lib/components/ui/button';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import MemberRow from '$lib/components/ui/MemberRow.svelte';
	import { Section } from '$lib/components/ui/section';
	import { Spinner } from '$lib/components/ui/spinner';
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

<main class="pt-safe-page mx-auto max-w-[480px] space-y-8 px-6 pb-10">
	{#if loading}
		<div class="flex justify-center py-12">
			<Spinner />
		</div>
	{:else if error}
		<div class="space-y-6 py-12">
			<div class="space-y-2">
				<p class="text-destructive text-center text-sm font-semibold">{error}</p>
			</div>
			<Button onclick={() => goto('/profile')} variant="default" size="cta">Back to Profile</Button>
		</div>
	{:else if club}
		<!-- Header -->
		<div class="flex items-center justify-between gap-4">
			<button
				onclick={() => goto('/profile')}
				class="text-text-secondary hover:text-text-primary flex-shrink-0 transition-colors"
				aria-label="Back"
			>
				‹
			</button>
			<div class="min-w-0 flex-1">
				<h1 class="truncate text-2xl font-[800]">{club.club.name}</h1>
			</div>
			<div class="flex-shrink-0 text-right">
				<p class="text-text-secondary text-xs font-semibold">{club.roster_count}</p>
				<p class="text-text-disabled text-[11px]">members</p>
			</div>
		</div>

		<!-- Club Info -->
		{#if club.club.description}
			<div class="bg-surface-raised space-y-2 rounded-2xl px-4 py-3.5">
				<p class="text-sm">{club.club.description}</p>
			</div>
		{/if}

		<!-- Members Section -->
		<Section title={`Members (${club.roster_count})`} collapsible={false}>
			{#snippet children()}
				{#if club.members && club.members.length > 0}
					<div class="space-y-2">
						{#each club.members as member}
							<MemberRow {member} />
						{/each}
					</div>
				{:else}
					<p class="text-text-disabled py-1 text-sm">No members yet</p>
				{/if}
			{/snippet}
		</Section>
	{/if}

	<Footer />
</main>
