<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api/client';
	import { auth } from '$lib/auth.svelte';
	import { Button } from '$lib/components/ui/button';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import MemberRow from '$lib/components/ui/MemberRow.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import { Section } from '$lib/components/ui/section';
	import { Spinner } from '$lib/components/ui/spinner';
	import Footer from '$lib/components/Footer.svelte';
	import ClubAdminDrawer from '$lib/components/ClubAdminDrawer.svelte';
	import CreateDrawer from '$lib/components/CreateDrawer.svelte';
	import { toast } from 'svelte-sonner';
	import { translateApiError } from '$lib/i18n/errors';
	import { onMount, onDestroy } from 'svelte';
	import { userStream, type UserStream } from '$lib/stores/userStream.svelte';
	import {
		CalendarDays,
		Check,
		Copy,
		Plus,
		Radio,
		RefreshCw,
		Search,
		Settings,
		UserPlus,
		Users
	} from '@lucide/svelte';

	let club = $state<App.ClubDetail | null>(null);
	let loading = $state(true);
	let error = $state('');
	let linkCopied = $state(false);
	let rotating = $state(false);
	let adminDrawerOpen = $state(false);

	// Club events (upcoming games owned by the Club). The first is the "Next up"
	// hero; the rest fall into a compact list below.
	let events = $state<App.UpcomingEntry[]>([]);
	let createOpen = $state(false);
	const nextEvent = $derived(events[0] ?? null);
	const laterEvents = $derived(events.slice(1));

	// A club event is a normal Session — created via the shared create flow with the
	// Club preset, so members hear about it automatically (push + this live feed).
	const stream: UserStream = userStream(() => auth.token);

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

	// Member-facing invite: any Member can invite a registered User to the Club.
	let inviteSearch = $state('');
	let inviteResults = $state<App.UserSearchResult[]>([]);
	let inviteSearching = $state(false);
	let invitedIds = $state<Set<string>>(new Set());
	let inviteDebounce: ReturnType<typeof setTimeout>;

	// Members already on the roster shouldn't show up as invitable.
	const memberIds = $derived(new Set(members.map((m) => m.user_id)));

	function onInviteSearchInput() {
		clearTimeout(inviteDebounce);
		if (inviteSearch.length < 2) {
			inviteResults = [];
			inviteSearching = false;
			return;
		}
		inviteSearching = true;
		inviteDebounce = setTimeout(async () => {
			if (!auth.token) return;
			try {
				inviteResults = await api.contacts.search(auth.token, inviteSearch);
			} finally {
				inviteSearching = false;
			}
		}, 300);
	}

	async function inviteUser(userId: string) {
		if (!auth.token || !club) return;
		try {
			await api.clubs.invites.send(auth.token, club.club.id, userId);
			invitedIds = new Set(invitedIds).add(userId);
			toast.success('Invite sent');
		} catch (e) {
			toast.error(
				e instanceof Error ? translateApiError(e.message) : translateApiError('server_error')
			);
		}
	}

	let clubId = $state('');

	async function loadClub(id: string) {
		if (!auth.token) {
			goto('/auth');
			return;
		}

		try {
			loading = true;
			club = await api.clubs.detail(auth.token, id);
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

	// reloadClub refetches without toggling the page spinner, so the admin drawer
	// stays mounted while roster/detail changes land.
	async function reloadClub() {
		if (!auth.token || !clubId) return;
		try {
			club = await api.clubs.detail(auth.token, clubId);
		} catch {
			// A stale drawer isn't worth a hard error; the next open reloads.
		}
	}

	// Compact "Mon 14:30"-style label for a scheduled event; falls back silently
	// on an unparseable value rather than showing "Invalid Date".
	function formatEventTime(iso: string): string {
		const d = new Date(iso);
		if (isNaN(d.getTime())) return '';
		return d.toLocaleString(undefined, {
			weekday: 'short',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	async function loadEvents() {
		if (!auth.token || !clubId) return;
		try {
			events = await api.clubs.events(auth.token, clubId);
		} catch {
			// The events feed is secondary to the roster; a fetch blip just leaves
			// the last-known list in place rather than erroring the whole page.
		}
	}

	onMount(() => {
		clubId = window.location.pathname.split('/')[2];
		if (clubId) {
			loadClub(clubId);
			loadEvents();
		}
		// A new club event anywhere in a Club we're a member of pushes a live nudge;
		// refresh the feed so the "Next up" hero reflects it without a reload.
		stream.start();
		stream.onEvent('club_event_created', loadEvents);
	});

	onDestroy(() => stream.close());
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
			{#snippet action()}
				{#if isAdmin}
					<button
						onclick={() => (adminDrawerOpen = true)}
						class="text-text-secondary hover:text-text-primary flex-shrink-0 transition-colors"
						aria-label="Manage club"
					>
						<Settings size={22} />
					</button>
				{/if}
			{/snippet}
			{#if club.club.description}
				<p class="text-text-secondary text-sm leading-relaxed">{club.club.description}</p>
			{/if}
		</PageHeader>

		<!-- Club games — the primary thing: getting the crowd into a game. Any member
		     can schedule one (Club preset), and the "Next up" hero foregrounds the
		     soonest game so it's the first thing a member sees. -->
		<section class="space-y-3">
			<div class="flex items-center justify-between">
				<h2 class="text-text-primary text-sm font-[800]">Club games</h2>
				<Button onclick={() => (createOpen = true)} size="sm" variant="default">
					<Plus size={16} />
					Schedule
				</Button>
			</div>

			{#if nextEvent}
				<a
					href={`/s/${nextEvent.session_id}`}
					class="bg-primary-muted hover:bg-primary/15 block rounded-2xl p-4 transition-colors"
				>
					<div class="flex items-center gap-2">
						<span class="text-primary text-[11px] font-bold tracking-[0.1em] uppercase">
							{nextEvent.status === 'playing' ? 'Live now' : 'Next up'}
						</span>
						{#if nextEvent.status === 'playing'}
							<Radio size={14} class="text-primary" />
						{/if}
					</div>
					<p class="text-text-primary mt-1 truncate text-base font-[800]">{nextEvent.name}</p>
					<p class="text-text-secondary mt-0.5 text-xs">
						{nextEvent.player_count}
						players
						{#if nextEvent.scheduled_at}· {formatEventTime(nextEvent.scheduled_at)}{/if}
					</p>
				</a>

				{#if laterEvents.length > 0}
					<!-- Below the hero, events collapse to compact single-line rows: a small
					     status dot, the name, then meta (players · time) right-aligned. -->
					<div class="divide-border bg-surface-raised divide-y overflow-hidden rounded-2xl">
						{#each laterEvents as ev}
							<a
								href={`/s/${ev.session_id}`}
								class="hover:bg-border flex items-center gap-3 px-4 py-2.5 transition-colors"
							>
								{#if ev.status === 'playing'}
									<Radio size={14} class="text-primary shrink-0" />
								{:else}
									<CalendarDays size={14} class="text-text-disabled shrink-0" />
								{/if}
								<p class="text-text-primary min-w-0 flex-1 truncate text-sm font-semibold">
									{ev.name}
								</p>
								<span class="text-text-secondary flex shrink-0 items-center gap-1 text-xs">
									<Users size={12} class="text-text-disabled" />
									{ev.player_count}
									{#if ev.scheduled_at}· {formatEventTime(ev.scheduled_at)}{/if}
								</span>
							</a>
						{/each}
					</div>
				{/if}
			{:else}
				<div class="bg-surface-raised rounded-2xl px-4 py-6 text-center">
					<p class="text-text-secondary text-sm">
						No upcoming games. Schedule one — the whole club will be notified.
					</p>
				</div>
			{/if}
		</section>

		<!-- Invite members — invite a registered User by name (a Club invite that adds
		     them to the roster). Falling back below the "or" divider is the shareable
		     join link, styled distinctly from a Session join link: any member can share
		     it, only admins can reset (rotate) it. -->
		<Section title="Invite members" collapsible={false}>
			{#snippet children()}
				<div class="space-y-4">
					<div class="space-y-3">
						<div class="relative">
							<div class="pointer-events-none absolute inset-y-0 left-3.5 flex items-center">
								<Search size={15} class="text-text-disabled" />
							</div>
							<input
								type="text"
								placeholder="Search people by name"
								bind:value={inviteSearch}
								oninput={onInviteSearchInput}
								class="bg-surface-raised focus:ring-primary w-full rounded-xl py-2.5 pr-4 pl-9 text-sm transition-shadow outline-none focus:ring-2"
							/>
						</div>

						{#if inviteSearch.length >= 2}
							{#if inviteResults.length > 0}
								<div class="space-y-1.5">
									{#each inviteResults as result}
										<div class="bg-surface-raised flex items-center gap-3 rounded-2xl px-4 py-3">
											<Avatar
												icon={result.avatar_icon}
												color={result.avatar_color}
												name={result.display_name}
												size="sm"
											/>
											<p class="flex-1 truncate text-sm font-semibold">{result.display_name}</p>
											{#if memberIds.has(result.id)}
												<span class="text-text-disabled text-xs">Member</span>
											{:else if invitedIds.has(result.id)}
												<span class="text-primary flex items-center gap-1 text-xs font-semibold">
													<Check size={14} /> Invited
												</span>
											{:else}
												<button
													onclick={() => inviteUser(result.id)}
													class="text-primary"
													aria-label="Invite {result.display_name}"
												>
													<UserPlus size={16} />
												</button>
											{/if}
										</div>
									{/each}
								</div>
							{:else if !inviteSearching}
								<p class="text-text-disabled py-1 text-sm">No people found</p>
							{/if}
						{/if}
					</div>

					<!-- or -->
					<div class="flex items-center gap-3">
						<div class="bg-border h-px flex-1"></div>
						<span class="text-text-disabled text-xs">or share a link</span>
						<div class="bg-border h-px flex-1"></div>
					</div>

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

		<CreateDrawer bind:open={createOpen} club={{ id: club.club.id, name: club.club.name }} />

		{#if isAdmin}
			<ClubAdminDrawer
				bind:open={adminDrawerOpen}
				{club}
				onchanged={reloadClub}
				ondeleted={() => goto('/profile')}
			/>
		{/if}
	{/if}

	<Footer />
</main>
