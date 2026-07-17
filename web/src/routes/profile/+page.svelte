<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { auth } from '$lib/auth.svelte';
	import { api } from '$lib/api/client';
	import { clearPlayerSession } from '$lib/playerSession';
	import { _ } from 'svelte-i18n';
	import { CalendarDays, Radio, UserPlus, X, Search, Check, Settings } from '@lucide/svelte';
	import Footer from '$lib/components/Footer.svelte';
	import CreateDrawer from '$lib/components/CreateDrawer.svelte';
	import CreateClubDrawer from '$lib/components/CreateClubDrawer.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import RatingGate from '$lib/components/RatingGate.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import ClubCard from '$lib/components/ui/ClubCard.svelte';
	import { Section } from '$lib/components/ui/section';
	import { JoinCodeInput } from '$lib/components/ui/join-code-input';
	import { ExpandableList } from '$lib/components/ui/expandable-list';
	import { Spinner } from '$lib/components/ui/spinner';
	import { toast } from 'svelte-sonner';
	import { translateApiError } from '$lib/i18n/errors';
	import { userStream, type UserStream } from '$lib/stores/userStream.svelte';

	const stream: UserStream = userStream(() => auth.token);
	let offInvite: Array<() => void> = [];

	onMount(() => {
		return () => {
			offInvite.forEach((off) => off());
			stream.close();
		};
	});

	async function refreshInvites() {
		if (!auth.token) return;
		try {
			invites = await api.invites.list(auth.token);
		} catch {
			/* heal on next load */
		}
		// Viewing invites is an acknowledgement — clear the app-icon badge.
		navigator.clearAppBadge?.().catch(() => {});
	}

	let showCreateDrawer = $state(false);

	let stats = $state<App.CareerSummary | null>(null);
	let tournaments = $state<App.TournamentEntry[]>([]);
	let upcoming = $state<App.UpcomingEntry[]>([]);
	let loading = $state(true);
	let joinCode = $state('');

	let showStats = $state(true);
	let showContacts = $state(false);
	let showUpcoming = $state(false);
	let showHistory = $state(false);
	let showClubs = $state(false);

	let clubs = $state<App.ClubListItem[]>([]);
	let clubsLoading = $state(false);
	let showCreateClubDrawer = $state(false);

	let invites = $state<App.Invite[]>([]);
	let contacts = $state<App.Contact[]>([]);
	let contactSearch = $state('');
	let searchResults = $state<App.UserSearchResult[]>([]);
	let searchLoading = $state(false);
	let searchDebounce: ReturnType<typeof setTimeout>;
	let contactToDelete = $state<App.Contact | null>(null);
	let showContactDeleteConfirm = $state(false);

	function onContactSearchInput() {
		clearTimeout(searchDebounce);
		if (contactSearch.length < 2) {
			searchResults = [];
			searchLoading = false;
			return;
		}
		searchLoading = true;
		searchDebounce = setTimeout(async () => {
			try {
				searchResults = await api.contacts.search(auth.token!, contactSearch);
			} finally {
				searchLoading = false;
			}
		}, 300);
	}

	async function acceptInvite(inviteID: string, sessionID: string) {
		if (!auth.token) return;
		await api.invites.accept(inviteID, auth.token);
		invites = invites.filter((i) => i.id !== inviteID);
		window.location.href = `/s/${sessionID}`;
	}

	async function declineInvite(inviteID: string) {
		if (!auth.token) return;
		await api.invites.decline(inviteID, auth.token);
		invites = invites.filter((i) => i.id !== inviteID);
	}

	// Leave an upcoming tournament straight from the profile — works for any
	// session the user belongs to, including legacy ones (server matches by
	// user_id). Only offered for lobby entries; a started tournament can't be left.
	let leaveTarget = $state<App.UpcomingEntry | null>(null);
	let leaving = $state(false);

	async function confirmLeave() {
		if (!auth.token || !leaveTarget) return;
		const sessionID = leaveTarget.session_id;
		leaving = true;
		try {
			await api.sessions.leave(sessionID, auth.token);
			upcoming = upcoming.filter((u) => u.session_id !== sessionID);
			clearPlayerSession(sessionID);
			toast.success($_('lobby_left'));
			leaveTarget = null;
		} catch (e) {
			toast.error(
				e instanceof Error ? translateApiError(e.message) : translateApiError('server_error')
			);
		} finally {
			leaving = false;
		}
	}

	async function addContact(userID: string) {
		await api.contacts.add(auth.token!, userID);
		contactSearch = '';
		searchResults = [];
		contacts = await api.contacts.list(auth.token!);
	}

	async function removeContact(userID: string) {
		await api.contacts.remove(auth.token!, userID);
		contacts = contacts.filter((c) => c.user_id !== userID);
		searchResults = searchResults.map((r) => (r.id === userID ? { ...r, is_contact: false } : r));
	}

	function handleDeleteContact(contact: App.Contact) {
		contactToDelete = contact;
		showContactDeleteConfirm = true;
	}

	async function confirmDeleteContact() {
		if (contactToDelete) {
			await removeContact(contactToDelete.user_id);
			showContactDeleteConfirm = false;
		}
	}

	function sessionHref(sessionId: string) {
		if (typeof localStorage === 'undefined') return `/s/${sessionId}`;
		const token = localStorage.getItem(`admin_token_${sessionId}`);
		return token ? `/s/${sessionId}?token=${token}` : `/s/${sessionId}`;
	}

	function joinByCode(code: string) {
		goto(`/s/${code}`);
	}

	async function load() {
		if (!auth.token) return;
		try {
			const [profileRes, historyRes, contactsRes, invitesRes] = await Promise.all([
				api.auth.profile(auth.token),
				api.auth.history(auth.token),
				api.contacts.list(auth.token),
				api.invites.list(auth.token)
			]);
			stats = profileRes.stats;
			tournaments = historyRes.tournaments;
			upcoming = (historyRes.upcoming ?? []).sort((a, b) => {
				if (!a.scheduled_at && !b.scheduled_at) return 0;
				if (!a.scheduled_at) return 1;
				if (!b.scheduled_at) return -1;
				return new Date(a.scheduled_at).getTime() - new Date(b.scheduled_at).getTime();
			});
			contacts = contactsRes;
			invites = invitesRes;
			// Landing on the profile clears any pending-invite app-icon badge.
			navigator.clearAppBadge?.().catch(() => {});
			showUpcoming = upcoming.length > 0;
			showHistory = tournaments.length > 0;
		} catch (err) {
			console.error('Failed to load profile data:', err);
			toast.error('Failed to load profile');
		} finally {
			loading = false;
		}
	}

	async function loadClubs() {
		if (!auth.token) return;
		clubsLoading = true;
		try {
			const result = await api.clubs.list(auth.token);
			clubs = result ?? [];
			showClubs = clubs.length > 0;
		} catch (err) {
			console.error('Failed to load clubs:', err);
			clubs = [];
			toast.error('Failed to load clubs');
		} finally {
			clubsLoading = false;
		}
	}

	$effect(() => {
		if (auth.ready && auth.token) {
			loadClubs();
		}
	});

	onMount(async () => {
		// Wait for auth to be ready before checking token
		while (!auth.ready) {
			await new Promise((resolve) => setTimeout(resolve, 10));
		}

		if (!auth.token) {
			goto('/auth');
			return;
		}
		if (page.url.searchParams.get('notfound') === '1') {
			toast.error(translateApiError('session_not_found'));
		}
		if (page.url.searchParams.get('create') === '1') {
			showCreateDrawer = true;
		}
		await load();

		offInvite = [
			stream.onEvent('invite_received', refreshInvites),
			stream.onEvent('invite_revoked', refreshInvites)
		];
		stream.start();
	});

	// Summary numbers are hidden (em dash) at zero games so an empty profile reads
	// as intentional rather than a jarring 0% (ADR 0007).
	const hasGames = $derived(!!stats && stats.games > 0);
	const pointWinPct = $derived(hasGames ? `${Math.round(stats!.point_win_pct)}%` : '–');
	const winRate = $derived(hasGames ? `${Math.round(stats!.winrate)}%` : '–');

	const memberSince = $derived(
		auth.user
			? new Date(auth.user.created_at).toLocaleDateString(undefined, {
					month: 'long',
					year: 'numeric'
				})
			: ''
	);

	function formatDate(iso: string) {
		return new Date(iso).toLocaleDateString(undefined, {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	function ordinal(n: number) {
		const s = ['th', 'st', 'nd', 'rd'];
		const v = n % 100;
		return n + (s[(v - 20) % 10] ?? s[v] ?? s[0]);
	}
</script>

<main class="pt-safe-page mx-auto max-w-[480px] space-y-8 px-6 pb-10">
	<!-- Header -->
	<div class="flex items-center justify-between gap-4">
		<div class="flex min-w-0 items-center gap-4">
			<Avatar
				icon={auth.user?.avatar_icon ?? ''}
				color={auth.user?.avatar_color ?? 'forest'}
				name={auth.user?.display_name ?? ''}
				size="lg"
			/>
			<div class="min-w-0">
				<h1 class="truncate text-2xl font-[800]">{auth.user?.display_name}</h1>
				{#if memberSince}
					<p class="text-text-secondary text-sm">Member since {memberSince}</p>
				{/if}
			</div>
		</div>
		<a
			href="/profile/settings"
			class="text-text-secondary hover:text-text-primary flex-shrink-0 transition-colors"
			aria-label="Settings"
		>
			<Settings size={24} />
		</a>
	</div>

	<!-- Pending invites -->
	{#if invites.length > 0}
		<div class="space-y-2">
			<p class="text-text-secondary text-[11px] font-bold tracking-[0.1em] uppercase">Invites</p>
			{#each invites as invite}
				<div class="bg-surface-raised flex items-center gap-3 rounded-2xl px-4 py-3.5">
					<div class="min-w-0 flex-1">
						<p class="truncate text-sm font-semibold">{invite.session_name}</p>
						<p class="text-text-secondary text-xs">From {invite.from_display_name}</p>
					</div>
					<button
						onclick={() => acceptInvite(invite.id, invite.session_id)}
						class="bg-primary flex items-center gap-1 rounded-full px-3 py-1.5 text-xs font-semibold text-white"
					>
						<Check size={12} /> Accept
					</button>
					<button
						onclick={() => declineInvite(invite.id)}
						class="bg-surface text-text-disabled hover:text-destructive border-border flex items-center justify-center rounded-full border p-1.5 transition-colors"
						aria-label="Decline"
					>
						<X size={14} />
					</button>
				</div>
			{/each}
		</div>
	{/if}

	<!-- New tournament + join code -->
	<div class="space-y-3">
		<button
			onclick={() => (showCreateDrawer = true)}
			class="bg-primary block w-full rounded-2xl px-4 py-4 text-center text-[15px] font-semibold text-white"
		>
			{$_('profile_new_tournament')}
		</button>
		<div class="flex items-center gap-3">
			<div class="bg-border h-px flex-1"></div>
			<span class="text-text-disabled text-xs">{$_('home_join_code_divider')}</span>
			<div class="bg-border h-px flex-1"></div>
		</div>
		<JoinCodeInput bind:value={joinCode} onComplete={joinByCode} />
	</div>

	{#if loading}
		<div class="flex justify-center py-12">
			<Spinner />
		</div>
	{:else if stats}
		<!-- Cross-mode career summary: the three numbers that stay honest blended
		     across game modes (ADR 0007). -->
		{@const s = stats}
		<Section title={$_('profile_stats_section')} bind:open={showStats}>
			{#snippet children()}
				<div class="grid grid-cols-3 gap-3">
					<div class="bg-surface-raised flex flex-col items-center gap-1.5 rounded-2xl px-3 py-5">
						<p class="text-primary text-3xl leading-none font-[800] tabular-nums">{pointWinPct}</p>
						<p
							class="text-text-disabled text-center text-[11px] font-bold tracking-[0.1em] uppercase"
						>
							{$_('profile_point_win_pct')}
						</p>
					</div>
					<div class="bg-surface-raised flex flex-col items-center gap-1.5 rounded-2xl px-3 py-5">
						<p class="text-3xl leading-none font-[800] tabular-nums">{winRate}</p>
						<p
							class="text-text-disabled text-center text-[11px] font-bold tracking-[0.1em] uppercase"
						>
							{$_('profile_win_rate')}
						</p>
					</div>
					<div class="bg-surface-raised flex flex-col items-center gap-1.5 rounded-2xl px-3 py-5">
						<p class="text-3xl leading-none font-[800] tabular-nums">{s.games}</p>
						<p
							class="text-text-disabled text-center text-[11px] font-bold tracking-[0.1em] uppercase"
						>
							{$_('profile_games')}
						</p>
					</div>
				</div>
				<!-- Expand the headline into the full per-mode Career Stats page (#228). -->
				<a href="/profile/stats" class="text-primary mt-3 block text-center text-sm font-semibold">
					{$_('profile_see_all_stats')}
				</a>
			{/snippet}
		</Section>

		<!-- Contacts -->
		<Section title={$_('profile_contacts_title')} bind:open={showContacts}>
			{#snippet children()}
				<div class="space-y-3">
					<!-- Search -->
					<div class="relative">
						<div class="pointer-events-none absolute inset-y-0 left-3.5 flex items-center">
							<Search size={15} class="text-text-disabled" />
						</div>
						<input
							type="text"
							placeholder={$_('profile_contacts_search_placeholder')}
							bind:value={contactSearch}
							oninput={onContactSearchInput}
							class="bg-surface-raised focus:ring-primary w-full rounded-xl py-2.5 pr-4 pl-9 text-sm transition-shadow outline-none focus:ring-2"
						/>
					</div>

					<!-- Search results section -->
					{#if contactSearch.length >= 2}
						<div class="space-y-2">
							<p class="text-text-secondary px-1 text-xs font-semibold">
								Search results {#if searchResults.length > 0}({searchResults.length}){/if}
							</p>
							{#if searchResults.length > 0}
								<div class="space-y-1.5">
									{#each searchResults as result}
										<div class="bg-surface-raised flex items-center gap-3 rounded-2xl px-4 py-3">
											<Avatar icon="" color="forest" name={result.display_name} size="sm" />
											<p class="flex-1 truncate text-sm font-semibold">{result.display_name}</p>
											{#if result.is_contact}
												<button
													onclick={() =>
														handleDeleteContact({
															user_id: result.id,
															display_name: result.display_name
														} as App.Contact)}
													class="text-text-disabled hover:text-destructive transition-colors"
													aria-label="Remove contact"
												>
													<X size={16} />
												</button>
											{:else}
												<button
													onclick={() => addContact(result.id)}
													class="text-primary"
													aria-label="Add contact"
												>
													<UserPlus size={16} />
												</button>
											{/if}
										</div>
									{/each}
								</div>
							{:else if !searchLoading}
								<p class="text-text-disabled py-1 text-sm">{$_('profile_contacts_search_empty')}</p>
							{/if}
						</div>
					{/if}

					<!-- Saved contacts section (always visible) -->
					{#if contacts.length === 0 && contactSearch.length < 2}
						<p class="text-text-disabled py-1 text-sm">{$_('profile_contacts_empty')}</p>
					{:else if contacts.length > 0}
						<div class="space-y-2">
							{#if contactSearch.length >= 2}
								<p class="text-text-secondary px-1 text-xs font-semibold">Your contacts</p>
							{/if}
							<div class="space-y-1.5">
								{#each contacts as contact}
									<div class="bg-surface-raised flex items-center gap-3 rounded-2xl px-4 py-3">
										<div
											class="bg-primary-muted text-primary flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-xs font-[800]"
										>
											{contact.display_name[0].toUpperCase()}
										</div>
										<p class="flex-1 truncate text-sm font-semibold">{contact.display_name}</p>
										<button
											onclick={() => handleDeleteContact(contact)}
											class="text-text-disabled hover:text-destructive transition-colors"
											aria-label="Remove contact"
										>
											<X size={16} />
										</button>
									</div>
								{/each}
							</div>
						</div>
					{/if}
				</div>
			{/snippet}
		</Section>

		<!-- Delete contact confirmation -->
		<ConfirmDialog
			open={showContactDeleteConfirm}
			title="Delete Contact?"
			description="Remove {contactToDelete?.display_name ||
				'this contact'} from your contacts. This action cannot be undone."
			confirmLabel="Delete"
			cancelLabel="Cancel"
			destructive={true}
			onconfirm={confirmDeleteContact}
			oncancel={() => (showContactDeleteConfirm = false)}
		/>

		<!-- Leave tournament confirmation -->
		<ConfirmDialog
			open={leaveTarget !== null}
			title={$_('leave_dialog_title')}
			description={$_('leave_dialog_desc')}
			confirmLabel={$_('leave_dialog_confirm')}
			cancelLabel={$_('leave_dialog_cancel')}
			destructive={true}
			onconfirm={confirmLeave}
			oncancel={() => (leaveTarget = null)}
		/>

		<!-- Upcoming -->
		<Section title={$_('profile_upcoming_label')} bind:open={showUpcoming}>
			{#snippet children()}
				{#if upcoming.length === 0}
					<p class="text-text-disabled py-1 text-sm">{$_('profile_upcoming_empty')}</p>
				{:else}
					<ExpandableList items={upcoming} showCount={5}>
						{#snippet itemContent(t)}
							<div
								class="bg-surface-raised hover:bg-border flex items-center gap-2 rounded-2xl pr-2 transition-colors"
							>
								<a
									href={sessionHref(t.session_id)}
									class="flex min-w-0 flex-1 items-center gap-4 py-3.5 pl-4"
								>
									<div
										class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full {t.status ===
										'playing'
											? 'bg-primary/15 text-primary'
											: 'bg-primary-muted text-primary'}"
									>
										{#if t.status === 'playing'}<Radio size={18} />{:else}<CalendarDays
												size={18}
											/>{/if}
									</div>
									<div class="min-w-0 flex-1">
										<div class="flex items-center gap-2">
											<p class="truncate text-sm font-semibold">{t.name}</p>
											{#if t.status === 'playing'}
												<span
													class="bg-primary/15 text-primary shrink-0 rounded-full px-2 py-0.5 text-[10px] font-bold tracking-wide uppercase"
													>Live</span
												>
											{/if}
										</div>
										<p class="text-text-secondary text-xs">
											{t.player_count}
											{$_('profile_upcoming_players')} · Americano
										</p>
									</div>
								</a>
								{#if t.status === 'lobby'}
									<button
										onclick={() => (leaveTarget = t)}
										disabled={leaving}
										class="text-text-disabled hover:text-destructive shrink-0 p-2 transition-colors disabled:opacity-40"
										aria-label={$_('lobby_leave')}
									>
										<X size={16} />
									</button>
								{:else}
									<span class="text-text-secondary pr-2 text-sm">→</span>
								{/if}
							</div>
						{/snippet}
					</ExpandableList>
				{/if}
			{/snippet}
		</Section>

		<!-- Tournament history -->
		<Section title={$_('profile_history_label')} bind:open={showHistory}>
			{#snippet children()}
				{#if tournaments.length === 0}
					<p class="text-text-disabled py-2 text-sm">{$_('profile_history_empty')}</p>
				{:else}
					<ExpandableList items={tournaments} showCount={5}>
						{#snippet itemContent(t)}
							<a
								href="/s/{t.session_id}"
								class="bg-surface-raised hover:bg-border flex items-center gap-4 rounded-2xl px-4 py-3.5 transition-colors"
							>
								<div
									class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full text-xs font-[800]
                    {t.rank === 1 ? 'bg-primary text-white' : 'bg-border text-text-secondary'}"
								>
									{ordinal(t.rank)}
								</div>
								<div class="min-w-0 flex-1">
									<p class="truncate text-sm font-semibold">{t.name}</p>
									<p class="text-text-secondary text-xs">
										{formatDate(t.played_at)} · {t.points} pts
										{#if t.ended_early}
											· <span class="text-text-disabled">{$_('profile_ended_early')}</span>
										{/if}
									</p>
								</div>
								<span class="text-text-secondary text-sm">→</span>
							</a>
						{/snippet}
					</ExpandableList>
				{/if}
			{/snippet}
		</Section>
	{/if}

	{#if (clubs && clubs.length > 0) || clubsLoading}
		<Section title="Clubs" bind:open={showClubs}>
			{#snippet children()}
				{#if clubsLoading}
					<div class="flex items-center justify-center py-8">
						<Spinner label="Loading clubs..." />
					</div>
				{:else if clubs && clubs.length > 0}
					<div class="space-y-2">
						{#each clubs as club}
							<ClubCard {club} onclick={() => goto(`/clubs/${club.id}`)} />
						{/each}
					</div>
					<button
						onclick={() => (showCreateClubDrawer = true)}
						class="text-primary hover:text-primary-hover w-full rounded-2xl px-4 py-3.5 text-sm font-semibold transition-colors"
					>
						+ Create Club
					</button>
				{/if}
			{/snippet}
		</Section>
	{/if}

	<Footer />
</main>

<CreateDrawer bind:open={showCreateDrawer} />
<CreateClubDrawer bind:open={showCreateClubDrawer} />

<!-- Home backfill gate (#213): legacy accounts with a null self_rating must pick a
     level before using the dashboard. Only mounts here on home, never on deep links. -->
{#if auth.ready && auth.user && auth.user.self_rating == null}
	<RatingGate />
{/if}
