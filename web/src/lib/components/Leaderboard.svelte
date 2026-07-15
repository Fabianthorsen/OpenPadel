<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import { _ } from 'svelte-i18n';
	import { Trophy } from 'lucide-svelte';
	import { shortName } from '$lib/utils';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import AvatarWithContactBadge from '$lib/components/ui/AvatarWithContactBadge.svelte';
	import Spinner from '$lib/components/ui/spinner/spinner.svelte';
	import { Button } from '$lib/components/ui/button';
	import { auth } from '$lib/auth.svelte';
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';
	import type { SessionStream } from '$lib/stores/sessionStream.svelte';

	let {
		sessionId,
		sessionName = '',
		complete = false,
		stream = null,
		inSheet = false
	}: {
		sessionId: string;
		sessionName?: string;
		complete?: boolean;
		stream?: SessionStream | null;
		inSheet?: boolean;
	} = $props();

	let leaderboard = $state<App.Leaderboard | null>(null);
	let addedContacts = $state<Record<string, boolean>>({});
	let existingContacts = $state<Record<string, boolean>>({});

	async function load() {
		try {
			leaderboard = await api.leaderboard.get(sessionId);
		} catch {
			// swallow
		}
	}

	async function loadContacts() {
		if (!auth.token) return;
		const contacts = await api.contacts.list(auth.token);
		existingContacts = Object.fromEntries(contacts.map((c) => [c.user_id, true]));
	}

	async function addContact(userID: string) {
		if (!auth.token) return;
		await api.contacts.add(auth.token, userID);
		addedContacts = { ...addedContacts, [userID]: true };
	}

	async function shareResults() {
		const text = `${sessionName || 'Tournament results'} ${window.location.href}`;
		if (navigator.share) {
			try {
				await navigator.share({ title: sessionName, text });
			} catch {
				// User cancelled
			}
		} else {
			navigator.clipboard.writeText(text).then(() => {
				toast.success($_('leaderboard_share_copied'));
			});
		}
	}

	function newSession() {
		goto('/?create=1');
	}

	function closeSession() {
		goto(auth.user ? '/profile' : '/');
	}

	onMount(() => {
		load();
		if (complete) {
			loadContacts();
			return;
		}
		if (!stream) return;
		return stream.onEvent('round_updated', () => {
			load();
		});
	});

	const leader = $derived(leaderboard?.standings[0] ?? null);

	const podiumOrder = $derived(
		leaderboard
			? [
					leaderboard.standings[1], // 2nd — left
					leaderboard.standings[0], // 1st — centre
					leaderboard.standings[2] // 3rd — right
				].filter(Boolean)
			: []
	);
</script>

<main class="mx-auto max-w-[480px] space-y-6 px-4 {inSheet ? 'pb-6' : 'pt-safe-page pb-24'}">
	{#if !leaderboard}
		<div class="flex flex-col items-center justify-center gap-3 py-12">
			<Spinner />
			<p class="text-text-secondary text-sm">{$_('loading')}</p>
		</div>
	{:else if complete}
		<!-- ── Final Results ── -->

		<!-- Heading -->
		<div class="space-y-0.5 pt-4 text-center">
			<p class="text-text-disabled text-[11px] font-bold tracking-[0.1em] uppercase">
				{$_('leaderboard_final')}
			</p>
			{#if sessionName}
				<p class="text-xl font-[800]">{sessionName}</p>
			{/if}
		</div>

		<!-- Podium -->
		<div class="flex items-end justify-center gap-3 pt-6 pb-2">
			{#each podiumOrder as s}
				{@const isFirst = s.rank === 1}
				<div
					class="flex flex-col items-center {isFirst
						? 'order-2 -mb-0'
						: s.rank === 2
							? 'order-1'
							: 'order-3'} max-w-[120px] flex-1"
				>
					<!-- Trophy for winner -->
					{#if isFirst}
						<div class="text-primary mb-1"><Trophy size={28} /></div>
					{/if}

					<!-- Avatar -->
					<div
						class={isFirst
							? 'ring-primary-muted rounded-full shadow-lg ring-4'
							: 'ring-border rounded-full ring-2'}
					>
						<Avatar
							icon={s.avatar_icon}
							color={s.avatar_color}
							name={s.name}
							size={isFirst ? 'xl' : 'lg'}
						/>
					</div>

					<!-- Rank badge: gold/silver/bronze medal colors -->
					<div
						class="mt-2 flex h-6 w-6 items-center justify-center rounded-full text-xs font-[800] text-white
            {isFirst
							? 'bg-[var(--color-medal-gold)]'
							: s.rank === 2
								? 'bg-[var(--color-medal-silver)]'
								: 'bg-[var(--color-medal-bronze)]'}"
					>
						{s.rank}
					</div>

					<p
						class="mt-1.5 w-full truncate text-center text-sm font-[800] {isFirst
							? 'text-text-primary'
							: 'text-text-secondary'}"
					>
						{shortName(s.name)}
					</p>
					<p
						class="text-[10px] font-bold tracking-widest uppercase {isFirst
							? 'text-primary'
							: 'text-text-disabled'}"
					>
						{s.points}
						{$_('leaderboard_pts')}
					</p>
					<div class="mt-1 flex items-center gap-1.5 text-[11px] font-bold tabular-nums">
						<span class="text-primary">{s.wins ?? 0}W</span>
						<span class="text-text-disabled">·</span>
						<span class="text-text-disabled">{s.draws ?? 0}D</span>
						<span class="text-text-disabled">·</span>
						<span class="text-destructive"
							>{(s.games_played ?? 0) - (s.wins ?? 0) - (s.draws ?? 0)}L</span
						>
					</div>

					<!-- Podium bar: gold/silver/bronze medal colors -->
					<div
						class="mt-3 w-full rounded-t-xl
            {isFirst
							? 'h-12 bg-[var(--color-medal-gold)]'
							: s.rank === 2
								? 'h-8 bg-[var(--color-medal-silver)]'
								: 'h-5 bg-[var(--color-medal-bronze)]'}"
					></div>
				</div>
			{/each}
		</div>

		<!-- Rest of standings (4th+) -->
		{#if leaderboard.standings.length > 3}
			<div class="space-y-1">
				<p class="text-text-disabled px-1 text-[11px] font-bold tracking-[0.1em] uppercase">
					{$_('leaderboard_ranking')}
				</p>
				{#each leaderboard.standings.slice(3) as s (s.player_id)}
					{@const isContact = !!(existingContacts[s.user_id ?? ''] || addedContacts[s.user_id ?? ''])}
					{@const showBadge = !!(auth.token && s.user_id && s.user_id !== auth.user?.id)}
					<div class="bg-surface-raised flex items-center gap-3 rounded-2xl px-4 py-3">
						<span class="text-text-disabled w-6 text-sm font-[800] tabular-nums">{s.rank}</span>
						<AvatarWithContactBadge
							icon={s.avatar_icon}
							color={s.avatar_color}
							name={s.name}
							size="sm"
							ring="ring-2 ring-primary/30"
							{showBadge}
							{isContact}
							targetName={s.name}
							onAdd={() => addContact(s.user_id!)}
						/>
						<span class="flex-1 truncate text-sm font-semibold">{shortName(s.name)}</span>
						<div class="flex items-center gap-1 text-[11px] font-bold tabular-nums">
							<span class="text-primary">{s.wins ?? 0}W</span>
							<span class="text-text-disabled">·</span>
							<span class="text-text-disabled">{s.draws ?? 0}D</span>
							<span class="text-text-disabled">·</span>
							<span class="text-destructive"
								>{(s.games_played ?? 0) - (s.wins ?? 0) - (s.draws ?? 0)}L</span
							>
						</div>
						<span class="text-base font-[800] tabular-nums">{s.points}</span>
						<span class="text-text-disabled text-[10px] font-bold tracking-widest uppercase"
							>{$_('leaderboard_pts')}</span
						>
					</div>
				{/each}
			</div>
		{/if}

		<!-- Actions -->
		<div class="flex flex-col gap-3 pt-2">
			<div class="flex gap-2">
				<Button onclick={shareResults} variant="secondary" class="flex-1">
					{$_('leaderboard_share')}
				</Button>
				<Button onclick={newSession} class="flex-1">
					{$_('leaderboard_new_session')}
				</Button>
			</div>
			<Button onclick={closeSession} variant="outline" class="w-full">
				{$_('leaderboard_close')}
			</Button>
		</div>
	{:else}
		<!-- ── Live Standings (calm & lean) ── -->
		<div class="space-y-4">
			<!-- Header: title + round/live info -->
			<div>
				<div class="flex items-center justify-between px-1 pb-2">
					<h2 class="text-text-secondary text-[13px] font-bold tracking-[0.1em] uppercase">
						{sessionName || $_('leaderboard_current')}
					</h2>
					{#if leaderboard.current_round}
						<div class="text-text-secondary flex items-center gap-2 text-xs">
							<span>
								{leaderboard.total_rounds
									? $_('leaderboard_round_of', {
											values: {
												current: leaderboard.current_round,
												total: leaderboard.total_rounds
											}
										})
									: $_('active_round_open', { values: { current: leaderboard.current_round } })}
							</span>
							<span class="flex items-center gap-1">
								<span class="bg-primary inline-block h-1.5 w-1.5 animate-pulse rounded-full"></span>
								{$_('leaderboard_live')}
							</span>
						</div>
					{/if}
				</div>
			</div>

			<!-- Standings rows: rank · avatar+name · points (calm, no colors) -->
			<div class="space-y-0.5">
				{#each leaderboard.standings as s (s.player_id)}
					{@const isRank1 = s.rank === 1}
					<div
						class="bg-surface flex items-center gap-3 rounded-2xl px-4 py-3.5"
						aria-live="polite"
						aria-label="{s.rank}. {s.name}: {s.points} points"
					>
						<!-- Rank number: rank 1 in primary, others disabled -->
						<span
							class="w-6 text-sm font-[800] tabular-nums {isRank1
								? 'text-primary'
								: 'text-text-disabled'}"
						>
							{s.rank}
						</span>

						<!-- Avatar + name: rank 1 emphasised, others neutral -->
						<div class="flex min-w-0 flex-1 items-center gap-2.5">
							<Avatar
								icon={s.avatar_icon}
								color={s.avatar_color}
								name={s.name}
								size="sm"
								ring="ring-2 ring-primary/30"
							/>
							<span
								class="truncate text-sm {isRank1
									? 'text-primary font-bold'
									: 'text-text-primary font-semibold'}"
							>
								{shortName(s.name)}
							</span>
						</div>

						<!-- Points: right-aligned, the ranking metric -->
						<span class="text-text-primary text-base font-[800] tabular-nums">{s.points}</span>
					</div>
				{/each}
			</div>
		</div>
	{/if}
</main>
