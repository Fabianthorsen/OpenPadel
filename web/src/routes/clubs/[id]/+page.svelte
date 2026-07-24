<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api/client';
	import { auth } from '$lib/auth.svelte';
	import { Button } from '$lib/components/ui/button';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import MemberRow from '$lib/components/ui/MemberRow.svelte';
	import { Section } from '$lib/components/ui/section';
	import { Spinner } from '$lib/components/ui/spinner';
	import Footer from '$lib/components/Footer.svelte';
	import { toast } from 'svelte-sonner';
	import { onMount } from 'svelte';
	import { Check, Copy, RefreshCw } from '@lucide/svelte';

	let club = $state<App.ClubDetail | null>(null);
	let loading = $state(true);
	let error = $state('');
	let linkCopied = $state(false);
	let rotating = $state(false);

	// The Club invite link uses the distinct /c/join/:code path — never the club id —
	// so it can't be confused with a Session join link.
	const inviteLink = $derived(
		club && typeof location !== 'undefined'
			? `${location.origin}/c/join/${club.club.join_code}`
			: ''
	);

	async function copyInviteLink() {
		if (!inviteLink) return;
		try {
			if (navigator.share) {
				await navigator.share({ title: club?.club.name, url: inviteLink });
			} else {
				await navigator.clipboard.writeText(inviteLink);
				linkCopied = true;
				setTimeout(() => (linkCopied = false), 2000);
			}
		} catch {
			// User dismissed the share sheet — nothing to do.
		}
	}

	async function rotateLink() {
		if (!auth.token || !club) return;
		try {
			rotating = true;
			const { join_code } = await api.clubs.rotateJoinCode(auth.token, club.club.id);
			club.club.join_code = join_code;
			toast.success('Invite link reset — the old link no longer works');
		} catch {
			toast.error('Could not reset the invite link');
		} finally {
			rotating = false;
		}
	}

	// Derived so the members list doesn't depend on `{#if club}` narrowing, which
	// TS can't carry into the Section's `children` snippet closure.
	const members = $derived(club?.members ?? []);
	const isAdmin = $derived(club?.is_admin ?? false);

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
		<PageHeader
			title={club.club.name}
			backHref="/profile"
			avatar={{ icon: club.club.avatar_icon, color: club.club.avatar_color, name: club.club.name }}
			subtitle={`${club.roster_count} ${club.roster_count === 1 ? 'member' : 'members'}`}
		>
			{#if club.club.description}
				<p class="text-text-secondary text-sm leading-relaxed">{club.club.description}</p>
			{/if}
		</PageHeader>

		<!-- Club invite link — styled distinctly from a Session join link. Any member
		     can share it; only admins can reset (rotate) it. -->
		<Section title="Invite link" collapsible={false}>
			{#snippet children()}
				<div class="space-y-2">
					<div class="bg-surface-raised flex items-center gap-2 rounded-xl px-3 py-2.5">
						<span class="text-text-secondary flex-1 truncate text-xs">{inviteLink}</span>
						<button
							onclick={copyInviteLink}
							class="text-primary hover:text-primary-hover shrink-0 p-1"
							aria-label="Copy invite link"
						>
							{#if linkCopied}
								<Check size={16} />
							{:else}
								<Copy size={16} />
							{/if}
						</button>
						{#if isAdmin}
							<button
								onclick={rotateLink}
								disabled={rotating}
								class="text-text-secondary hover:text-text-primary shrink-0 p-1 disabled:opacity-50"
								aria-label="Reset invite link"
								title="Reset link — revokes the old one"
							>
								<RefreshCw size={15} class={rotating ? 'animate-spin' : ''} />
							</button>
						{/if}
					</div>
					<p class="text-text-disabled text-xs">
						Anyone with this link can join.{#if isAdmin}
							Reset it to revoke the old link.{/if}
					</p>
				</div>
			{/snippet}
		</Section>

		<!-- Members Section -->
		<Section title={`Members (${club.roster_count})`} collapsible={false}>
			{#snippet children()}
				{#if members.length > 0}
					<div class="space-y-2">
						{#each members as member}
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
