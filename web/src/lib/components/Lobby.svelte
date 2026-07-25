<script lang="ts">
	import { api, ApiError } from '$lib/api/client';
	import {
		Crown,
		Share,
		Check,
		Search,
		UserPlus,
		Clock,
		Info,
		Settings,
		Pencil,
		Users
	} from '@lucide/svelte';
	import { onMount } from 'svelte';
	import { sessionName } from '$lib/utils';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import { SectionLabel } from '$lib/components/ui/section-label';
	import { OrDivider } from '$lib/components/ui/or-divider';
	import { PillToggleGroup, PillToggleItem } from '$lib/components/ui/pill-toggle-group';
	const MAX_COURTS = 4;
	import { Calendar } from '$lib/components/ui/calendar';
	import Stepper from '$lib/components/ui/stepper/Stepper.svelte';
	import * as Dialog from '$lib/components/ui/dialog';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import { _ } from 'svelte-i18n';
	import { auth } from '$lib/auth.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import SessionConfig from '$lib/components/SessionConfig.svelte';
	import RatingPicker from '$lib/components/RatingPicker.svelte';
	import { calculateAmericanoRounds } from '$lib/americano';
	import {
		savePlayerSession,
		getPlayerId,
		getPlayerToken,
		clearPlayerSession
	} from '$lib/playerSession';
	import { toast } from 'svelte-sonner';
	import { translateApiError } from '$lib/i18n/errors';
	import { goto } from '$app/navigation';
	import {
		type DateValue,
		today,
		getLocalTimeZone,
		parseAbsoluteToLocal
	} from '@internationalized/date';
	import type { sessionStream } from '$lib/stores/sessionStream.svelte';
	type SessionStream = ReturnType<typeof sessionStream>;

	let {
		session,
		isAdmin,
		onRefresh,
		onStarted,
		onLeave,
		stream
	}: {
		session: App.Session;
		isAdmin: boolean;
		onRefresh: () => void;
		onStarted: () => void;
		onLeave?: (leaving: boolean) => void;
		stream?: SessionStream;
	} = $props();

	const isDev = import.meta.env.DEV;
	// Dev "fill lobby" pools. seedAccounts mirror the registered users created by
	// `make seed` (backend cmd/seed) — filling logs in as each to add a real
	// account-linked Player. guestNames are pure guests (no account). The fill
	// interleaves the two so the roster is a genuine mix, not guests-only.
	const seedPassword = 'password123';
	const seedAccounts = [
		{ email: 'alice@openpadel.local', name: 'Alice' },
		{ email: 'bob@openpadel.local', name: 'Bob' },
		{ email: 'carol@openpadel.local', name: 'Carol' },
		{ email: 'dave@openpadel.local', name: 'Dave' },
		{ email: 'erik@openpadel.local', name: 'Erik' },
		{ email: 'fiona@openpadel.local', name: 'Fiona' },
		{ email: 'grace@openpadel.local', name: 'Grace' },
		{ email: 'henry@openpadel.local', name: 'Henry' }
	];
	const guestNames = ['Ola', 'Kari', 'Nils', 'Ingrid', 'Sven', 'Astrid', 'Lars', 'Sofie'];

	let copied = $state(false);
	let starting = $state(false);
	let cancelling = $state(false);
	let showCancelDialog = $state(false);
	let leavingSession = $state(false);
	let showLeaveDialog = $state(false);
	let seeding = $state(false);
	let joinName = $state('');
	// Guests self-joining by link must pick a skill level (#210); registered
	// users seed their rating from their account, so this stays null for them.
	let guestRating = $state<number | null>(null);
	let joining = $state(false);

	let playerSearch = $state('');
	let playerResults = $state<App.UserSearchResult[]>([]);
	let playerSearchLoading = $state(false);
	let playerSearchDebounce: ReturnType<typeof setTimeout>;
	let sessionInvites = $state<App.Invite[]>([]);
	// For a club event, the invite surface is scoped to the owning Club: pick from
	// its roster, or invite everyone at once (#128). Loaded for a signed-in admin.
	let clubMembers = $state<App.ClubMember[]>([]);
	let invitingAll = $state(false);
	const isClubEvent = $derived(!!session.club_id);

	// SessionConfig drawer state
	let configDrawerOpen = $state(false);

	// ── Inline config editing state ──
	let editingName = $state(false);
	let nameInputEl = $state<HTMLInputElement | null>(null);
	$effect(() => {
		if (editingName && nameInputEl) nameInputEl.focus();
	});
	let nameInput = $state('');
	let configMode = $state<'americano' | 'mexicano'>('americano');
	let configCourts = $state(2);
	let configPoints = $state(24);
	let configRounds = $state(7);
	let roundsMode = $state<'fixed' | 'unlimited'>('fixed');
	let scheduleEnabled = $state(false);
	let calendarDate = $state<DateValue | undefined>(undefined);
	let timeSlot = $state(20);

	// Sync config state whenever session prop changes (after refresh).
	$effect(() => {
		nameInput = session.name ?? '';
		configMode = session.game_mode as 'americano' | 'mexicano';
		configCourts = session.courts;
		configPoints = session.points;
		configRounds = session.rounds_total ?? 7;
		roundsMode =
			session.rounds_total === null || session.rounds_total === undefined ? 'unlimited' : 'fixed';
		scheduleEnabled = !!session.scheduled_at;
		if (session.scheduled_at) {
			try {
				calendarDate = parseAbsoluteToLocal(session.scheduled_at);
			} catch {
				calendarDate = undefined;
			}
			const d = new Date(session.scheduled_at);
			const slot = Math.round((d.getHours() * 60 + d.getMinutes() - 8 * 60) / 30);
			timeSlot = Math.max(0, Math.min(27, slot));
		} else {
			calendarDate = undefined;
		}
	});

	function slotToLabel(slot: number) {
		const totalMins = 8 * 60 + slot * 30;
		const h = String(Math.floor(totalMins / 60)).padStart(2, '0');
		const m = String(totalMins % 60).padStart(2, '0');
		return `${h}:${m}`;
	}
	const scheduleTime = $derived(slotToLabel(timeSlot));
	const timeHour = $derived(8 + Math.floor(timeSlot / 2));
	const timeMinute = $derived((timeSlot % 2) * 30);

	function onHourChange(h: number) {
		timeSlot = (h - 8) * 2 + (timeSlot % 2);
		commitScheduleTime();
	}
	function onMinuteChange(m: number) {
		timeSlot = (timeHour - 8) * 2 + Math.round(m / 30);
		commitScheduleTime();
	}

	function calculateNextHourSlot(): number {
		const now = new Date();
		const nextHour = now.getMinutes() > 0 ? now.getHours() + 1 : now.getHours();
		const clamped = Math.min(21, Math.max(8, nextHour));
		return Math.round((clamped * 60 - 8 * 60) / 30);
	}

	async function patchConfig(patch: Parameters<typeof api.sessions.update>[1]) {
		const adminToken = localStorage.getItem(`admin_token_${session.id}`) ?? '';
		try {
			await api.sessions.update(session.id, patch, adminToken);
			onRefresh();
		} catch (e) {
			toast.error(
				e instanceof ApiError ? translateApiError(e.message) : translateApiError('server_error')
			);
			// Reset local state to match server
			configMode = session.game_mode as 'americano' | 'mexicano';
			configCourts = session.courts;
			configPoints = session.points;
			configRounds = session.rounds_total ?? 7;
		}
	}

	async function commitName() {
		editingName = false;
		if (nameInput === (session.name ?? '')) return;
		await patchConfig({ name: nameInput });
	}

	function onModeChange(mode: 'americano' | 'mexicano') {
		configMode = mode;
		if (mode === 'mexicano' && configCourts < 2) configCourts = 2;
		const patch: Parameters<typeof api.sessions.update>[1] = {
			game_mode: mode,
			courts: configCourts
		};
		if (mode === 'mexicano') patch.rounds_total = configRounds;
		patchConfig(patch);
	}

	function onCourtsChange(n: number) {
		configCourts = n;
		patchConfig({ courts: n });
	}

	function onPointsChange(n: number) {
		configPoints = n;
		patchConfig({ points: n });
	}

	function onRoundsChange(n: number) {
		configRounds = n;
		patchConfig({ rounds_total: n });
	}

	function onRoundsModeChange(mode: 'fixed' | 'unlimited') {
		roundsMode = mode;
		if (mode === 'unlimited') {
			patchConfig({ rounds_total: null });
		} else {
			patchConfig({ rounds_total: configRounds });
		}
	}

	async function commitSchedule(enabled: boolean) {
		scheduleEnabled = enabled;
		if (!enabled) {
			calendarDate = undefined;
			timeSlot = 20;
			await patchConfig({ scheduled_at: '' });
			return;
		}
		calendarDate = today(getLocalTimeZone());
		timeSlot = calculateNextHourSlot();
	}

	async function commitScheduleTime() {
		if (!scheduleEnabled || !calendarDate) return;
		const [h, m] = scheduleTime.split(':').map(Number);
		const d = calendarDate.toDate(getLocalTimeZone());
		d.setHours(h, m, 0, 0);
		await patchConfig({ scheduled_at: d.toISOString() });
	}

	onMount(() => {
		if (isAdmin) {
			api.invites
				.listForSession(session.id)
				.catch(() => [])
				.then((v) => {
					sessionInvites = v;
				});
			if (auth.token && session.club_id) {
				api.clubs
					.detail(auth.token, session.club_id)
					.then((d) => {
						clubMembers = d.members;
					})
					.catch(() => {});
			}
		}
		if (stream) {
			return stream.onEvent('session_updated', async () => {
				if (isAdmin) {
					sessionInvites = await api.invites.listForSession(session.id).catch(() => []);
				}
			});
		}
	});

	const joinedUserIds = $derived(new Set(session.players.map((p) => p.user_id).filter(Boolean)));
	const pendingInvites = $derived(
		sessionInvites.filter((inv) => !joinedUserIds.has(inv.to_user_id))
	);

	// Club-event invite list: the owning Club's members who aren't already in the
	// Session or already invited, filtered by the search box. The admin themselves
	// is dropped (they're the creator, already on the roster).
	const invitedUserIds = $derived(new Set(sessionInvites.map((inv) => inv.to_user_id)));
	const invitableClubMembers = $derived(
		clubMembers.filter((m) => {
			if (m.user_id === auth.user?.id) return false;
			if (joinedUserIds.has(m.user_id) || invitedUserIds.has(m.user_id)) return false;
			const q = playerSearch.trim().toLowerCase();
			return q === '' || m.display_name.toLowerCase().includes(q);
		})
	);

	function onPlayerSearchInput() {
		clearTimeout(playerSearchDebounce);
		if (playerSearch.length < 2) {
			playerResults = [];
			playerSearchLoading = false;
			return;
		}
		playerSearchLoading = true;
		playerSearchDebounce = setTimeout(async () => {
			try {
				playerResults = await api.contacts.search(auth.token!, playerSearch);
			} finally {
				playerSearchLoading = false;
			}
		}, 300);
	}

	async function inviteUser(userID: string) {
		if (!auth.token) return;
		await api.invites.send(session.id, userID, auth.token).catch(() => {});
		playerSearch = '';
		playerResults = [];
		sessionInvites = await api.invites.listForSession(session.id).catch(() => []);
	}

	// Fan the owning Club's whole roster out into Session invites in one tap (#128).
	// Members already in or already invited are skipped server-side, so a repeat
	// tap is safe; the toast reports how many invites actually went out.
	async function inviteClubAll() {
		if (!auth.token || !session.club_id || invitingAll) return;
		invitingAll = true;
		try {
			const created = await api.invites.sendClub(session.id, session.club_id, auth.token);
			sessionInvites = await api.invites.listForSession(session.id).catch(() => []);
			const club = session.club_name ?? '';
			if (created.length > 0) {
				toast.success($_('lobby_invite_club_sent', { values: { count: created.length, club } }));
			} else {
				toast.info($_('lobby_invite_club_none', { values: { club } }));
			}
		} catch (e) {
			toast.error(
				e instanceof ApiError ? translateApiError(e.message) : translateApiError('server_error')
			);
		} finally {
			invitingAll = false;
		}
	}

	// Optional level the admin picks when adding a guest by hand (#211). Blank
	// (null) lets the server fall back to the median, the one path that yields an
	// unrated-by-default Player.
	let guestAddRating = $state<number | null>(null);

	async function addGuest(name: string) {
		if (!name) return;
		joining = true;
		try {
			// Admin adds the guest by name; the admin token authorises the join so
			// the server lets the guest fall back to the median rating (#210, #211).
			const adminToken = localStorage.getItem(`admin_token_${session.id}`) ?? undefined;
			await api.players.join(session.id, name, undefined, adminToken, guestAddRating ?? undefined);
			playerSearch = '';
			guestAddRating = null;
			toast.success($_('lobby_player_joined'));
			onRefresh();
		} catch (e) {
			toast.error(
				e instanceof ApiError ? translateApiError(e.message) : translateApiError('server_error')
			);
		} finally {
			joining = false;
		}
	}

	// ── Inline rating edit (admin only, #211) ──
	// AdminToken is the only gate; a matching CreatorUserID never grants edit rights.
	let editingPlayer = $state<App.Player | null>(null);
	let editRating = $state<number | null>(null);
	let savingRating = $state(false);

	function openRatingEdit(player: App.Player) {
		editingPlayer = player;
		editRating = player.rating;
	}

	async function saveRating() {
		if (!editingPlayer || editRating == null) return;
		savingRating = true;
		try {
			const adminToken = localStorage.getItem(`admin_token_${session.id}`) ?? '';
			await api.players.updateRating(session.id, editingPlayer.id, editRating, adminToken);
			editingPlayer = null;
			onRefresh();
		} catch (e) {
			toast.error(
				e instanceof ApiError ? translateApiError(e.message) : translateApiError('server_error')
			);
		} finally {
			savingRating = false;
		}
	}

	// Pre-fill name from account when on the invite screen
	$effect(() => {
		if (auth.user && !isAdmin && !alreadyJoined && !joinName) {
			joinName = auth.user.display_name;
		}
	});

	const joinUrl = $derived(
		typeof location !== 'undefined' ? `${location.origin}/s/${session.id}` : ''
	);

	const isMexicano = $derived(session.game_mode === 'mexicano');
	const gameModeName = $derived(session.game_mode === 'mexicano' ? 'Mexicano' : 'Americano');
	let showRules = $state(false);
	const activePlayers = $derived(session.players.filter((p) => p.active));

	// Rounds shown in the header summary. Unlimited (rounds_total null) → no count.
	// Americano is derived live from the active roster (recomputes as players join
	// and leave); Mexicano shows its user-picked count.
	const summaryRounds = $derived(
		session.rounds_total == null
			? null
			: session.game_mode === 'americano'
				? calculateAmericanoRounds(activePlayers.length, session.courts)
				: session.rounds_total
	);

	const requiredPlayers = $derived(session.courts * 4);
	const maxPlayers = $derived(isMexicano ? requiredPlayers : undefined);
	const isFull = $derived(maxPlayers ? activePlayers.length >= maxPlayers : false);

	// Use server-computed can_start if available, otherwise fall back to local logic
	const canStart = $derived(
		session.can_start !== undefined
			? session.can_start
			: isMexicano
				? activePlayers.length === requiredPlayers
				: activePlayers.length >= requiredPlayers
	);

	const creatorName = $derived(
		activePlayers.find((p) => p.id === session.creator_player_id)?.name ?? ''
	);

	const myPlayerId = $derived(getPlayerId(session.id));
	const alreadyJoined = $derived(
		(!!myPlayerId && activePlayers.some((p) => p.id === myPlayerId)) ||
			(!!auth.user && activePlayers.some((p) => p.user_id === auth.user!.id))
	);

	async function copyLink() {
		if (navigator.share) {
			await navigator.share({ title: sessionName(session), url: joinUrl }).catch(() => {});
		} else {
			await navigator.clipboard.writeText(joinUrl);
			copied = true;
			setTimeout(() => (copied = false), 2000);
		}
	}

	async function join() {
		const name = joinName.trim();
		if (!name) return;
		// A guest (no account) must bring a rating; registered users seed theirs
		// from their account server-side, so we send none for them.
		const rating = auth.user ? undefined : (guestRating ?? undefined);
		if (!auth.user && rating === undefined) return;
		joining = true;
		try {
			const player = await api.players.join(
				session.id,
				name,
				isAdmin ? undefined : (auth.token ?? undefined),
				undefined,
				rating
			);
			if (!isAdmin) {
				savePlayerSession(session.id, player);
				localStorage.setItem('last_session_id', session.id);
			}
			toast.success($_('lobby_player_joined'));
			joinName = '';
			onRefresh();
		} catch (e) {
			toast.error(
				e instanceof ApiError ? translateApiError(e.message) : translateApiError('server_error')
			);
		} finally {
			joining = false;
		}
	}

	// Add a seeded registered user as an account-linked Player. Logs in as them
	// (they all share seedPassword) purely to mint a token for the join call —
	// this does not touch the current admin's auth session. Their rating comes
	// from their account. Falls back to a guest join if the account isn't seeded
	// yet (run `make seed`).
	async function joinAsAccount(acct: { email: string; name: string }) {
		try {
			const { token } = await api.auth.login(acct.email, seedPassword);
			await api.players.join(session.id, acct.name, token);
		} catch {
			await joinAsGuest(acct.name);
		}
	}

	// Add a pure guest (no account) with a random 1–5 rating so the balancer has
	// real spread to work with.
	async function joinAsGuest(name: string) {
		const rating = 1 + Math.floor(Math.random() * 5);
		await api.players.join(session.id, name, undefined, undefined, rating).catch(() => {});
	}

	async function seedPlayers() {
		seeding = true;
		try {
			const existing = new Set(activePlayers.map((p) => p.name));
			// Fill to a random roster size that fits the court count: courts×4 up to
			// courts×4+3 (1 court → 4–7, 2 courts → 8–11).
			const min = session.courts * 4;
			const target = min + Math.floor(Math.random() * 4);
			const wanted = Math.max(0, target - activePlayers.length);

			const accounts = seedAccounts.filter((a) => !existing.has(a.name));
			const guests = guestNames.filter((n) => !existing.has(n));

			// Interleave registered accounts and guests so the roster is a genuine
			// mix; fall back to whichever pool still has entries, then to generated
			// guest names if both run dry.
			let ai = 0;
			let gi = 0;
			for (let n = 0; n < wanted; n++) {
				if (n % 2 === 0 && ai < accounts.length) {
					await joinAsAccount(accounts[ai++]);
				} else if (gi < guests.length) {
					await joinAsGuest(guests[gi++]);
				} else if (ai < accounts.length) {
					await joinAsAccount(accounts[ai++]);
				} else {
					await joinAsGuest(`Player ${activePlayers.length + n + 1}`);
				}
			}
		} finally {
			seeding = false;
			onRefresh();
		}
	}

	async function cancel() {
		cancelling = true;
		// Flag the deletion before the request so the page ignores the resulting
		// session_updated/poll 404 instead of bouncing to "Tournament does not exist".
		onLeave?.(true);
		try {
			const adminToken = localStorage.getItem(`admin_token_${session.id}`) ?? '';
			await api.sessions.cancel(session.id, adminToken);
			goto('/');
		} catch {
			onLeave?.(false); // delete failed — re-enable the page's normal loading
			cancelling = false;
			showCancelDialog = false;
		}
	}

	async function leave() {
		leavingSession = true;
		try {
			if (auth.token) {
				// Retroactive path: server resolves membership by user_id, no stored
				// player id required.
				await api.sessions.leave(session.id, auth.token);
			} else if (myPlayerId) {
				// Guest self-removal proven by the per-player secret stored at join (#241).
				await api.players.leave(session.id, myPlayerId, getPlayerToken(session.id));
			} else {
				return;
			}
			clearPlayerSession(session.id);
			toast.success($_('lobby_left'));
			goto(auth.user ? '/profile' : '/');
		} catch (e) {
			toast.error(
				e instanceof ApiError ? translateApiError(e.message) : translateApiError('server_error')
			);
			leavingSession = false;
			showLeaveDialog = false;
		}
	}

	async function removePlayer(playerId: string) {
		const adminToken = localStorage.getItem(`admin_token_${session.id}`) ?? '';
		await api.players.remove(session.id, playerId, adminToken).catch(() => {});
		onRefresh();
	}

	async function start() {
		starting = true;
		try {
			const adminToken = localStorage.getItem(`admin_token_${session.id}`) ?? '';
			await api.sessions.start(session.id, adminToken);
			onStarted();
		} catch {
			starting = false;
		}
	}
</script>

{#if cancelling}
	<main class="flex min-h-svh flex-col items-center justify-center gap-3 px-6">
		<div class="border-border border-t-primary h-8 w-8 animate-spin rounded-full border-2"></div>
		<p class="text-text-secondary text-sm">{$_('lobby_cancelling')}</p>
	</main>

	<!-- ── Join / invite screen (visitor hasn't joined yet) ── -->
{:else if !isAdmin && !alreadyJoined}
	<main class="pt-safe-page flex min-h-svh flex-col items-center px-6 pb-12">
		<div class="flex w-full max-w-sm justify-end">
			<a
				href="/"
				class="text-text-disabled hover:bg-surface-raised hover:text-text-secondary flex h-7 w-7 items-center justify-center rounded-full transition-colors"
				aria-label="Back">×</a
			>
		</div>
		<div class="flex w-full max-w-sm flex-1 flex-col justify-center space-y-8">
			<!-- Brand + session info -->
			<div class="space-y-1">
				<p class="text-primary text-[11px] font-bold tracking-[0.1em] uppercase">OpenPadel</p>
				<div class="flex items-start gap-2">
					<h1 class="text-[28px] leading-tight font-[800]">
						{#if session.club_name}
							<!-- A club event is a club game, not a personal invite — frame it by
							     the Club, optionally naming the member who scheduled it. -->
							{#if creatorName}
								{$_('invite_title_club_named', {
									values: { creator: creatorName, club: session.club_name }
								})}
							{:else}
								{$_('invite_title_club', { values: { club: session.club_name } })}
							{/if}
						{:else if creatorName && session.name}
							{$_('invite_title_with_creator_named', {
								values: { creator: creatorName, name: session.name }
							})}
						{:else if creatorName}
							{$_('invite_title_with_creator', {
								values: { creator: creatorName, mode: gameModeName }
							})}
						{:else if session.name}
							{$_('invite_title_generic_named', { values: { name: session.name } })}
						{:else}
							{$_('invite_title_generic', { values: { mode: gameModeName } })}
						{/if}
					</h1>
					<button
						onclick={() => (showRules = true)}
						aria-label={$_('lobby_rules_button')}
						class="text-text-disabled hover:text-text-secondary mt-1.5 shrink-0 transition-colors"
					>
						<Info size={18} />
					</button>
				</div>
				<p class="text-text-secondary text-sm">
					{$_(session.courts === 1 ? 'active_courts_one' : 'active_courts_other', {
						values: { n: session.courts }
					})} · {session.points + ' ' + $_('invite_points')} · {gameModeName}{#if summaryRounds != null}
						· {summaryRounds} rds{/if}{#if session.scheduled_at}
						· {new Date(session.scheduled_at).toLocaleString(undefined, {
							weekday: 'short',
							month: 'short',
							day: 'numeric',
							hour: '2-digit',
							minute: '2-digit'
						})}{/if}
				</p>
			</div>

			<div class="space-y-4">
				{#if auth.user}
					<!-- Logged in: show account card + join -->
					<div class="bg-surface-raised flex items-center gap-3 rounded-2xl px-4 py-3.5">
						<Avatar
							icon={auth.user.avatar_icon}
							color={auth.user.avatar_color}
							name={auth.user.display_name}
							ring="ring-2 ring-primary/30"
						/>
						<div class="min-w-0 flex-1">
							<p class="truncate text-sm font-semibold">{auth.user.display_name}</p>
							<p class="text-text-secondary truncate text-xs">{auth.user.email}</p>
						</div>
					</div>

					{#if isFull}
						<div
							class="flex items-start gap-2 rounded-2xl bg-red-50 px-4 py-3 text-sm text-red-900"
						>
							<Info size={16} class="mt-0.5 shrink-0" />
							<p>This session has reached its maximum player limit for {gameModeName}.</p>
						</div>
					{/if}

					<Button
						onclick={join}
						disabled={joining || !joinName.trim() || isFull}
						class="bg-primary hover:bg-primary-hover h-auto w-full rounded-2xl px-4 py-4 text-[15px] font-semibold text-white"
					>
						{joining ? $_('invite_joining') : $_('invite_join_button')}
					</Button>
				{:else}
					<!-- Not logged in: account options first, guest below -->
					<a
						href="/auth?redirect=/s/{session.id}"
						class="bg-primary hover:bg-primary-hover flex h-auto w-full items-center justify-center rounded-2xl px-4 py-4 text-[15px] font-semibold text-white"
					>
						{$_('invite_sign_in')}
					</a>
					<p class="text-text-secondary text-center text-sm">
						{$_('invite_no_account')}
						<a href="/auth?register=1&redirect=/s/{session.id}" class="text-primary font-semibold"
							>{$_('invite_create_account')}</a
						>
					</p>

					<OrDivider label={$_('invite_or_guest')} />

					<!-- Guest fallback: name + required skill level -->
					<form
						onsubmit={(e) => {
							e.preventDefault();
							join();
						}}
						class="space-y-3"
					>
						<Input
							bind:value={joinName}
							placeholder={$_('invite_name_placeholder')}
							maxlength={32}
							class="bg-surface-raised w-full rounded-2xl border-0 px-4 py-3 text-sm"
						/>
						<div class="space-y-2">
							<SectionLabel>{$_('invite_rating_label')}</SectionLabel>
							<RatingPicker compact bind:value={guestRating} disabled={isFull} />
						</div>
						<Button
							type="submit"
							disabled={joining || !joinName.trim() || guestRating == null || isFull}
							class="bg-primary hover:bg-primary-hover h-auto w-full rounded-2xl px-4 py-4 text-[15px] font-semibold text-white shadow-none"
						>
							{joining ? '…' : $_('invite_guest_join')}
						</Button>
					</form>
				{/if}
			</div>
		</div>
		<Footer />
	</main>

	<!-- ── Lobby (admin or already joined) ── -->
{:else}
	<main class="pt-safe-page mx-auto w-full max-w-[480px] space-y-6 px-4 pb-6">
		<nav class="space-y-1">
			<!-- Club badge — marks this as a club game (not a one-off) for everyone who
			     opens it, so a member arriving from the Club home keeps that context. -->
			{#if session.club_name}
				<div class="text-primary flex items-center gap-1.5 text-xs font-bold">
					<Users size={13} class="shrink-0" />
					<span class="truncate tracking-[0.02em] uppercase">{session.club_name}</span>
				</div>
			{/if}
			<!-- Status -->
			<p class="text-text-disabled text-[11px] font-bold tracking-[0.12em] uppercase">
				{#if session.scheduled_at}
					{new Date(session.scheduled_at).toLocaleString(undefined, {
						weekday: 'short',
						month: 'short',
						day: 'numeric',
						hour: '2-digit',
						minute: '2-digit'
					})}
				{:else}
					{$_('lobby_waiting')}
				{/if}
			</p>
			<!-- Name + actions on one row -->
			<div class="flex items-center justify-between gap-2">
				<!-- Click-to-edit name (admin only) -->
				{#if isAdmin && editingName}
					<input
						bind:this={nameInputEl}
						class="text-text-primary border-primary min-w-0 flex-1 border-b bg-transparent text-2xl font-[800] focus:outline-none"
						bind:value={nameInput}
						maxlength={48}
						placeholder={$_('lobby_name_placeholder')}
						onblur={commitName}
						onkeydown={(e) => e.key === 'Enter' && commitName()}
					/>
				{:else if isAdmin}
					<button
						class="group flex min-w-0 flex-1 items-center gap-1.5 text-left"
						onclick={() => {
							nameInput = session.name ?? '';
							editingName = true;
						}}
					>
						<span class="text-text-primary min-w-0 truncate text-2xl font-[800]">
							{session.name || $_('lobby_name_placeholder')}
						</span>
						<Pencil
							size={15}
							class="text-text-disabled group-hover:text-text-secondary shrink-0 transition-colors"
						/>
					</button>
				{:else}
					<h1 class="text-text-primary min-w-0 flex-1 truncate text-2xl font-[800]">
						{sessionName(session)}
					</h1>
				{/if}
				<div class="flex shrink-0 items-center gap-0.5">
					{#if isAdmin}
						<button
							onclick={() => (configDrawerOpen = true)}
							aria-label={$_('lobby_edit_config')}
							class="text-text-disabled hover:bg-surface-raised hover:text-text-secondary flex h-8 w-8 items-center justify-center rounded-full transition-colors"
						>
							<Settings size={17} />
						</button>
					{/if}
					{#if isAdmin || alreadyJoined}
						<a
							href="/"
							class="text-text-disabled hover:bg-surface-raised hover:text-text-secondary flex h-8 w-8 shrink-0 items-center justify-center rounded-full transition-colors"
							aria-label="Back to home">×</a
						>
					{/if}
				</div>
			</div>
			<!-- Config summary — tap the mode for rules -->
			<p class="text-text-secondary text-xs">
				<button
					onclick={() => (showRules = true)}
					class="hover:text-text-primary capitalize underline decoration-dotted underline-offset-2 transition-colors"
				>
					{session.game_mode}
				</button>
				· {$_(session.courts === 1 ? 'active_courts_one' : 'active_courts_other', {
					values: { n: session.courts }
				})} · {session.points}
				{$_('invite_points')}{#if summaryRounds != null}
					· {summaryRounds} rds{/if}
			</p>
		</nav>

		<!-- Join code + share -->
		<div class="bg-surface-raised space-y-3 rounded-2xl px-5 py-4">
			<SectionLabel>{$_('lobby_join_code')}</SectionLabel>
			<div class="flex gap-2">
				{#each session.id.split('') as char}
					<div
						class="bg-surface text-text-primary flex flex-1 items-center justify-center rounded-xl py-3 font-mono text-2xl font-[700]"
					>
						{char}
					</div>
				{/each}
			</div>
			<div class="bg-surface flex items-center gap-2 rounded-xl px-3 py-2.5">
				<span class="text-text-secondary flex-1 truncate text-xs">{joinUrl}</span>
				<Button
					onclick={copyLink}
					variant="ghost"
					class="text-primary hover:text-primary-hover h-auto shrink-0 p-1.5 hover:bg-transparent"
				>
					{#if copied}
						<Check size={16} />
					{:else}
						<Share size={16} />
					{/if}
				</Button>
			</div>
		</div>

		<!-- Admin: invite or add guest -->
		{#if isAdmin}
			<div class="space-y-2">
				<SectionLabel>{$_('lobby_invite_label')}</SectionLabel>
				{#if isClubEvent}
					<!-- Club event: invite from the owning Club's roster, or invite the
					     whole club at once. A club game is only open to its Members, so
					     there's no free-text guest add here. -->
					<div class="relative">
						<div class="pointer-events-none absolute inset-y-0 left-3.5 flex items-center">
							<Search size={15} class="text-text-disabled" />
						</div>
						<Input
							bind:value={playerSearch}
							placeholder={$_('lobby_invite_search_members')}
							maxlength={32}
							class="bg-surface-raised w-full rounded-2xl border-0 py-3 pr-4 pl-9 text-sm"
						/>
					</div>
					{#if invitableClubMembers.length > 0}
						<div class="space-y-1.5">
							{#each invitableClubMembers as m (m.user_id)}
								<div class="bg-surface-raised flex items-center gap-3 rounded-2xl px-4 py-3">
									<Avatar
										icon={m.avatar_icon}
										color={m.avatar_color}
										name={m.display_name}
										size="sm"
										ring="ring-2 ring-primary/30"
									/>
									<p class="flex-1 truncate text-sm font-semibold">{m.display_name}</p>
									<button
										onclick={() => inviteUser(m.user_id)}
										class="bg-primary flex items-center gap-1 rounded-full px-3 py-1.5 text-xs font-semibold text-white"
									>
										<UserPlus size={12} />
										{$_('lobby_invite_button')}
									</button>
								</div>
							{/each}
						</div>
					{:else}
						<p class="text-text-disabled text-sm">{$_('lobby_invite_all_members_present')}</p>
					{/if}

					<OrDivider label={$_('lobby_invite_or')} />

					<button
						onclick={inviteClubAll}
						disabled={invitingAll}
						class="bg-primary flex w-full items-center justify-center gap-2 rounded-2xl px-4 py-3 text-sm font-semibold text-white disabled:opacity-50"
					>
						<Users size={15} class="shrink-0" />
						{invitingAll
							? '…'
							: $_('lobby_invite_all_club', { values: { club: session.club_name } })}
					</button>
				{:else}
					<!-- Personal session: invite by contact search, or add a guest by name. -->
					<div class="relative">
						<div class="pointer-events-none absolute inset-y-0 left-3.5 flex items-center">
							<Search size={15} class="text-text-disabled" />
						</div>
						<Input
							bind:value={playerSearch}
							oninput={onPlayerSearchInput}
							placeholder={$_('lobby_add_player_placeholder')}
							maxlength={32}
							class="bg-surface-raised w-full rounded-2xl border-0 py-3 pr-4 pl-9 text-sm"
						/>
					</div>
					{#if playerResults.length > 0}
						<div class="space-y-1.5">
							{#each playerResults as result}
								<div class="bg-surface-raised flex items-center gap-3 rounded-2xl px-4 py-3">
									<Avatar
										icon={result.avatar_icon}
										color={result.avatar_color}
										name={result.display_name}
										size="sm"
										ring="ring-2 ring-primary/30"
									/>
									<p class="flex-1 truncate text-sm font-semibold">{result.display_name}</p>
									<button
										onclick={() => inviteUser(result.id)}
										class="bg-primary flex items-center gap-1 rounded-full px-3 py-1.5 text-xs font-semibold text-white"
									>
										<UserPlus size={12} /> Invite
									</button>
								</div>
							{/each}
						</div>
					{/if}
					{#if playerSearch.trim().length > 0 && !playerSearchLoading}
						<div class="space-y-2">
							<div class="space-y-1.5">
								<SectionLabel>{$_('lobby_guest_rating_optional')}</SectionLabel>
								<RatingPicker compact bind:value={guestAddRating} disabled={joining || isFull} />
							</div>
							<button
								onclick={() => addGuest(playerSearch.trim())}
								disabled={joining || isFull}
								class="border-border text-text-secondary hover:border-primary hover:text-primary flex w-full items-center gap-3 rounded-2xl border border-dashed px-4 py-3 text-sm transition-colors disabled:opacity-50"
							>
								<UserPlus size={15} class="shrink-0" />
								Add "{playerSearch.trim()}" as guest
							</button>
						</div>
					{/if}
				{/if}
			</div>
		{/if}

		<!-- Player list -->
		<div class="space-y-2">
			<SectionLabel>
				{$_('lobby_players_label')} ({activePlayers.length})
			</SectionLabel>
			{#if activePlayers.length === 0 && sessionInvites.length === 0}
				<p class="text-text-disabled text-sm">{$_('lobby_waiting_players')}</p>
			{:else}
				<div class="bg-surface-raised divide-border divide-y rounded-2xl">
					{#each activePlayers as player (player.id)}
						<div class="flex items-center gap-3 px-4 py-3">
							<Avatar
								icon={player.avatar_icon}
								color={player.avatar_color}
								name={player.name}
								size="sm"
								ring="ring-2 ring-primary/30"
							/>
							<span class="text-sm font-medium">{player.name}</span>
							{#if isAdmin && player.added_by_admin}
								<button
									onclick={() => openRatingEdit(player)}
									class="bg-surface text-text-secondary hover:ring-primary/40 flex h-5 items-center gap-0.5 rounded-full px-1.5 text-xs font-bold tabular-nums transition-shadow hover:ring-2"
									aria-label={$_('lobby_edit_rating_aria', { values: { name: player.name } })}
								>
									{player.rating}
									<Pencil size={9} class="opacity-60" />
								</button>
							{:else}
								<span
									class="bg-surface text-text-secondary flex h-5 min-w-5 items-center justify-center rounded-full px-1.5 text-xs font-bold tabular-nums"
									title={$_('auth_rating_label')}
									aria-label={$_('auth_rating_label')}
								>
									{player.rating}
								</span>
							{/if}
							<div class="ml-auto flex items-center gap-1.5">
								{#if player.id === session.creator_player_id}
									<Crown size={13} class="text-primary" />
								{/if}
								{#if player.id === myPlayerId}
									<span class="text-text-disabled text-xs">{$_('lobby_you')}</span>
								{/if}
								{#if isAdmin && player.id !== session.creator_player_id && player.id !== myPlayerId}
									<button
										onclick={() => removePlayer(player.id)}
										class="text-text-disabled hover:bg-destructive/10 hover:text-destructive ml-1 flex h-5 w-5 items-center justify-center rounded-full transition-colors"
										aria-label="Remove player"
									>
										×
									</button>
								{/if}
							</div>
						</div>
					{/each}
					{#each pendingInvites as invite (invite.id)}
						<div class="flex items-center gap-3 px-4 py-3 opacity-60">
							<Avatar
								name={invite.to_display_name ?? '?'}
								size="sm"
								ring="ring-2 ring-primary/30"
							/>
							<span class="text-text-secondary flex-1 truncate text-sm font-medium"
								>{invite.to_display_name}</span
							>
							<div class="text-text-disabled ml-auto flex items-center gap-1">
								<Clock size={11} />
								<span class="text-xs">Invited</span>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>

		<!-- Join form (non-admin who hasn't joined) -->
		{#if !isAdmin && !alreadyJoined}
			<div class="space-y-2">
				{#if isFull}
					<div class="flex items-start gap-2 rounded-2xl bg-red-50 px-4 py-3 text-sm text-red-900">
						<Info size={16} class="mt-0.5 shrink-0" />
						<p>This session has reached its maximum player limit.</p>
					</div>
				{/if}
				<form
					onsubmit={(e) => {
						e.preventDefault();
						join();
					}}
					class="flex gap-2"
				>
					<Input
						bind:value={joinName}
						placeholder={$_('lobby_join_placeholder')}
						maxlength={32}
						class="bg-surface-raised flex-1 rounded-2xl border-0 px-4 py-3 text-sm"
						disabled={isFull}
					/>
					<Button
						type="submit"
						disabled={joining || !joinName.trim() || isFull}
						class="bg-primary h-auto rounded-2xl px-4 text-sm font-semibold text-white"
					>
						{joining ? $_('lobby_join_loading') : $_('lobby_join_button')}
					</Button>
				</form>
			</div>
		{/if}

		<!-- Admin controls -->
		{#if isAdmin}
			<div class="space-y-2">
				<Button
					onclick={start}
					disabled={starting || !canStart}
					class="bg-primary hover:bg-primary-hover h-auto w-full rounded-2xl px-4 py-4 text-[15px] font-semibold text-white"
				>
					{starting ? $_('lobby_start_loading') : $_('lobby_start_button')}
				</Button>
				{#if !canStart}
					{#if session.validation_errors && session.validation_errors.length > 0}
						{#each session.validation_errors as err}
							<p class="text-text-disabled text-center text-xs">
								{translateApiError(err.code, err.params)}
							</p>
						{/each}
					{:else}
						<p class="text-text-disabled text-center text-xs">
							{#if isMexicano}
								{$_('lobby_mexicano_exact_players', {
									values: { n: requiredPlayers, current: activePlayers.length }
								})}
							{:else}
								{$_('lobby_need_players', { values: { n: requiredPlayers } })}
							{/if}
						</p>
					{/if}
				{/if}
				<button
					onclick={() => (showCancelDialog = true)}
					disabled={cancelling}
					class="border-border text-text-secondary hover:border-destructive hover:text-destructive h-auto w-full rounded-2xl border px-4 py-3.5 text-sm font-semibold transition-colors disabled:opacity-40"
				>
					{cancelling ? $_('lobby_cancelling') : $_('lobby_cancel')}
				</button>
				{#if isDev}
					<div class="flex justify-center pt-2">
						<button
							onclick={seedPlayers}
							disabled={seeding}
							class="border-border text-text-disabled hover:border-text-disabled hover:text-text-secondary rounded-full border border-dashed px-4 py-1.5 text-xs transition-colors disabled:opacity-40"
						>
							{seeding ? $_('lobby_dev_seeding') : $_('lobby_dev_seed')}
						</button>
					</div>
				{/if}
			</div>
		{:else}
			<div class="space-y-2">
				<div class="bg-surface-raised rounded-2xl px-4 py-3 text-center">
					<p class="text-text-secondary text-sm">{$_('lobby_waiting_admin')}</p>
				</div>
				{#if alreadyJoined && !session.is_creator}
					<!-- The creator can't leave their own session's roster — they administer
					     it, so leaving would orphan it. They cancel the session instead. -->
					<button
						onclick={() => (showLeaveDialog = true)}
						disabled={leavingSession}
						class="border-border text-text-secondary hover:border-destructive hover:text-destructive h-auto w-full rounded-2xl border px-4 py-3.5 text-sm font-semibold transition-colors disabled:opacity-40"
					>
						{leavingSession ? $_('lobby_leaving') : $_('lobby_leave')}
					</button>
				{/if}
			</div>
		{/if}
	</main>
{/if}

<Dialog.Root bind:open={showRules}>
	<Dialog.Content class="w-full max-w-sm">
		<Dialog.Header>
			<Dialog.Title>{gameModeName}</Dialog.Title>
		</Dialog.Header>
		<div class="space-y-2">
			<ul class="space-y-2">
				{#each $_(`rules_${session.game_mode}`).split('\n') as line}
					{#if line.trim()}
						<li class="text-text-secondary flex gap-2 text-sm">
							<span class="text-primary mt-0.5 shrink-0">·</span>
							<span>{line.trim()}</span>
						</li>
					{/if}
				{/each}
			</ul>
		</div>
		<Dialog.Footer>
			<Dialog.Close
				class="border-border text-text-secondary hover:bg-surface-raised w-full rounded-2xl border px-4 py-3.5 text-sm font-semibold transition-colors"
			>
				{$_('leaderboard_close')}
			</Dialog.Close>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>

<ConfirmDialog
	open={showCancelDialog}
	title={$_('cancel_dialog_title')}
	description={$_('cancel_dialog_desc')}
	confirmLabel={$_('cancel_dialog_confirm')}
	cancelLabel={$_('cancel_dialog_cancel')}
	destructive
	onconfirm={cancel}
	oncancel={() => (showCancelDialog = false)}
/>

<ConfirmDialog
	open={showLeaveDialog}
	title={$_('leave_dialog_title')}
	description={$_('leave_dialog_desc')}
	confirmLabel={$_('leave_dialog_confirm')}
	cancelLabel={$_('leave_dialog_cancel')}
	destructive
	onconfirm={leave}
	oncancel={() => (showLeaveDialog = false)}
/>

<!-- Admin rating edit (#211) -->
<Dialog.Root
	open={editingPlayer != null}
	onOpenChange={(o) => {
		if (!o) editingPlayer = null;
	}}
>
	<Dialog.Content class="w-full max-w-sm">
		<Dialog.Header>
			<Dialog.Title>{$_('lobby_edit_rating_title')}</Dialog.Title>
		</Dialog.Header>
		{#if editingPlayer}
			<p class="text-text-secondary text-sm">
				{$_('lobby_edit_rating_desc', { values: { name: editingPlayer.name } })}
			</p>
			<RatingPicker
				bind:value={editRating}
				current={editingPlayer.rating}
				disabled={savingRating}
			/>
		{/if}
		<Dialog.Footer class="gap-2">
			<button
				onclick={() => (editingPlayer = null)}
				disabled={savingRating}
				class="border-border text-text-secondary hover:bg-surface-raised flex-1 rounded-2xl border px-4 py-3 text-sm font-semibold transition-colors disabled:opacity-50"
			>
				{$_('lobby_rating_cancel')}
			</button>
			<Button
				onclick={saveRating}
				disabled={savingRating || editRating == null}
				class="bg-primary hover:bg-primary-hover h-auto flex-1 rounded-2xl px-4 py-3 text-sm font-semibold text-white"
			>
				{savingRating ? '…' : $_('lobby_rating_save')}
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>

<!-- Session Config Drawer -->
<SessionConfig bind:open={configDrawerOpen} {session} sessionId={session.id} {onRefresh} />
