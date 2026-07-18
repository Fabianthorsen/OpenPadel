<script lang="ts">
	import { untrack } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { ApiError } from '$lib/api/client';
	import { translateApiError } from '$lib/i18n/errors';
	import { api } from '$lib/api/client';
	import { _ } from 'svelte-i18n';
	import { sessionDialog } from '$lib/stores/sessionDialog';
	import { Pencil, Shield, Clock, Trophy } from '@lucide/svelte';
	import { sessionName } from '$lib/utils';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import { Button } from '$lib/components/ui/button';
	import { SectionLabel } from '$lib/components/ui/section-label';
	import * as Sheet from '$lib/components/ui/sheet';
	import RoundIndicator from './RoundIndicator.svelte';
	import Leaderboard from './Leaderboard.svelte';
	import ScoreBoard from './ScoreBoard.svelte';
	import TeamScore from './TeamScore.svelte';
	import { numpad as numpadStore } from '$lib/stores/numpad';
	import { auth } from '$lib/auth.svelte';
	import type { SessionStream } from '$lib/stores/sessionStream.svelte';
	import { Card } from '$lib/components/ui/card';
	import Spinner from '$lib/components/ui/spinner/spinner.svelte';

	let {
		session,
		currentRound,
		isAdmin,
		onRefresh,
		stream
	}: {
		session: App.Session;
		currentRound: App.Round;
		isAdmin: boolean;
		onRefresh: () => void;
		stream: SessionStream;
	} = $props();

	const playerName = $derived(Object.fromEntries(session.players.map((p) => [p.id, p.name])));
	const playerById = $derived(Object.fromEntries(session.players.map((p) => [p.id, p])));
	// player_id → 1–5 Rating for the live standings peek (leaderboard API has no rating).
	const playerRatings = $derived(Object.fromEntries(session.players.map((p) => [p.id, p.rating])));
	const maxScore = $derived(session.points);

	let localScores = $state<Record<string, { a: number; b: number }>>({});
	let submitting = $state<Record<string, boolean>>({});
	let editing = $state<Record<string, boolean>>({});
	let initialServer = $state<Record<string, 'a' | 'b'>>({});
	let advancing = $state(false);
	let cancelling = $state(false);
	let closing = $state(false);
	let showEndMenu = $state(false);
	let showStandingsSheet = $state(false);

	// Numpad (mobile-optimized: drag-to-close, keyboard input, overwrite).
	// mode 'adjust' (single court) just tracks the score; 'final' (multi court)
	// submits the result on confirm, since there is no separate finalize button.
	type NumpadState = {
		matchId: string;
		team: 'a' | 'b';
		value: string;
		fresh: boolean;
		mode: 'adjust' | 'final';
	};
	let numpad = $state<NumpadState | null>(null);

	function openNumpad(matchId: string, team: 'a' | 'b', mode: 'adjust' | 'final' = 'adjust') {
		const current = scores[matchId]?.[team] ?? 0;
		const value = current > 0 ? String(current) : '';
		numpad = { matchId, team, value, fresh: true, mode };
		numpadStore.open({
			value,
			fresh: true,
			targetPoints: session.points,
			shaking: false,
			onDigit: numpadDigit,
			onDelete: numpadDelete,
			onConfirm: numpadConfirm,
			onClose: () => {
				numpad = null;
				numpadStore.close();
			}
		});
	}

	function numpadDigit(d: string) {
		if (!numpad) return;
		let next: string;
		if (numpad.fresh && numpad.value && numpad.value !== '0') {
			next = d;
		} else {
			next = (numpad.value + d).replace(/^0+(\d)/, '$1');
		}
		if (parseInt(next || '0') > maxScore) return;
		numpad = { ...numpad, value: next, fresh: false };
		numpadStore.update({ value: next, fresh: false });
	}

	function numpadDelete() {
		if (!numpad) return;
		const next = numpad.value.slice(0, -1);
		numpad = { ...numpad, value: next };
		numpadStore.update({ value: next, fresh: false });
	}

	function numpadConfirm() {
		if (!numpad) return;
		const entered = parseInt(numpad.value || '0');
		if (entered > maxScore) {
			numpadStore.update({ shaking: true });
			setTimeout(() => {
				numpadStore.update({ shaking: false });
			}, 400);
			return;
		}
		const { matchId, team, mode } = numpad;

		const other = session.points - entered;
		const next = team === 'a' ? { a: entered, b: other } : { a: other, b: entered };
		localScores[matchId] = next;
		numpad = null;
		numpadStore.close();
		if (mode === 'final') {
			submitScore(matchId, next.a, next.b);
		} else {
			scheduleLiveSave(matchId);
		}
	}

	const scores = $derived.by(() => {
		const result: Record<string, { a: number; b: number }> = {};
		for (const m of currentRound.matches) {
			if (editing[m.id]) {
				result[m.id] = localScores[m.id] ?? { a: 0, b: 0 };
			} else if (m.score) {
				result[m.id] = { a: m.score.a, b: m.score.b };
			} else if (localScores[m.id] !== undefined) {
				result[m.id] = localScores[m.id];
			} else if (m.live) {
				result[m.id] = { a: m.live.a, b: m.live.b };
			} else {
				result[m.id] = { a: 0, b: 0 };
			}
		}
		return result;
	});

	$effect(() => {
		const matches = currentRound.matches;
		untrack(() => {
			for (const m of matches) {
				if (!(m.id in initialServer)) {
					initialServer[m.id] = m.live?.server ?? 'a';
				}
			}
		});
	});

	const allScored = $derived(
		currentRound.matches.every((m) => m.score !== null) &&
			currentRound.matches.every((m) => !editing[m.id])
	);

	const someScored = $derived(currentRound.matches.some((m) => m.score !== null && !editing[m.id]));

	function scheduleLiveSave(matchId: string) {
		clearTimeout(saveTimeout[matchId]);
		saveTimeout[matchId] = setTimeout(async () => {
			const current = localScores[matchId];
			if (!current) return;
			const srv = initialServer[matchId] ?? 'a';
			await api.scores.updateLive(session.id, matchId, current.a, current.b, srv).catch(() => {});
		}, 400);
	}

	function adjust(matchId: string, team: 'a' | 'b', delta: number) {
		const s = scores[matchId] ?? { a: 0, b: 0 };
		localScores[matchId] = { ...s, [team]: Math.max(0, Math.min(maxScore, s[team] + delta)) };
		scheduleLiveSave(matchId);
	}

	async function submitScore(matchId: string, scoreA?: number, scoreB?: number) {
		clearTimeout(saveTimeout[matchId]);
		submitting[matchId] = true;
		const s =
			scoreA !== undefined && scoreB !== undefined ? { a: scoreA, b: scoreB } : scores[matchId];
		try {
			await api.scores.submit(session.id, matchId, s.a, s.b, '');
			editing[matchId] = false;
			toast.success($_('toast_score_confirmed'));
			onRefresh();
		} catch (e) {
			toast.error(
				e instanceof ApiError ? translateApiError(e.message) : translateApiError('server_error')
			);
		} finally {
			submitting[matchId] = false;
		}
	}

	async function closeSession() {
		closing = true;
		try {
			const adminToken = localStorage.getItem(`admin_token_${session.id}`) ?? '';
			await api.sessions.close(session.id, adminToken);
			sessionDialog.close();
			closing = false;
			onRefresh();
		} catch (e) {
			toast.error(
				e instanceof ApiError ? translateApiError(e.message) : translateApiError('server_error')
			);
			closing = false;
		}
	}

	async function cancelSession() {
		cancelling = true;
		try {
			const adminToken = localStorage.getItem(`admin_token_${session.id}`) ?? '';
			await api.sessions.cancel(session.id, adminToken);
			location.href = auth.user ? '/profile' : '/';
		} catch (e) {
			toast.error(
				e instanceof ApiError ? translateApiError(e.message) : translateApiError('server_error')
			);
			cancelling = false;
		}
	}

	async function advanceRound() {
		advancing = true;
		try {
			const adminToken = localStorage.getItem(`admin_token_${session.id}`) ?? '';
			await api.rounds.advance(session.id, adminToken);
			onRefresh();
		} catch {
			// ignore — button stays visible so admin can retry
		} finally {
			advancing = false;
		}
	}

	const benchNames = $derived(currentRound.bench.map((id) => playerName[id] ?? id));
	const benchIds = $derived(new Set(currentRound.bench));
	const adminPlayer = $derived(
		session.creator_player_id ? playerById[session.creator_player_id] : null
	);

	// Timer countdown
	let now = $state(Date.now());
	$effect(() => {
		const interval = setInterval(() => {
			now = Date.now();
		}, 1000);
		return () => clearInterval(interval);
	});
	const endsAtMs = $derived(session.ends_at ? new Date(session.ends_at).getTime() : null);
	const timeExpired = $derived(endsAtMs !== null && now >= endsAtMs);
	const timeLeft = $derived.by(() => {
		if (endsAtMs === null) return null;
		const ms = endsAtMs - now;
		if (ms <= 0) return null;
		const totalSecs = Math.ceil(ms / 1000);
		const h = Math.floor(totalSecs / 3600);
		const m = Math.floor((totalSecs % 3600) / 60);
		const s = totalSecs % 60;
		if (h > 0) return `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
		return `${m}:${String(s).padStart(2, '0')}`;
	});

	function shortPlayerName(name: string) {
		const parts = name.trim().split(' ');
		if (parts.length === 1) return parts[0];
		return `${parts[0]} ${parts[1][0]}.`;
	}

	function teamLabel(ids: readonly [string, string]) {
		return `${shortPlayerName(playerName[ids[0]] ?? '?')} & ${shortPlayerName(playerName[ids[1]] ?? '?')}`;
	}

	const saveTimeout: Record<string, ReturnType<typeof setTimeout>> = {};

	// Determine if layout is single-court or multi-court
	const isSingleCourt = $derived(currentRound.matches.length === 1);
</script>

{#if cancelling}
	<main class="flex min-h-svh flex-col items-center justify-center gap-3 px-6">
		<Spinner label={$_('lobby_cancelling')} />
		<p class="text-text-secondary text-sm">{$_('lobby_cancelling')}</p>
	</main>
{:else}
	<div class="flex h-full flex-col">
		<div class="min-h-0 flex-1 overflow-y-auto">
			<main class="pt-safe-page mx-auto w-full max-w-[480px] space-y-4 px-4 pb-6">
				<!-- Top bar: Session name + Standings pill + Back button -->
				<div
					class="bg-background/95 sticky top-0 z-40 flex items-center justify-between gap-2 pb-4 backdrop-blur-sm"
				>
					<p class="text-primary text-sm font-semibold">{sessionName(session)}</p>
					<button
						onclick={() => (showStandingsSheet = true)}
						class="bg-surface-raised hover:bg-surface-raised/80 flex shrink-0 items-center gap-1.5 rounded-full px-3 py-2 text-xs font-semibold transition-colors"
						aria-haspopup="dialog"
						aria-label="View standings"
					>
						<Trophy size={14} />
						<span>{$_('standings_label')}</span>
					</button>
					<a
						href="/"
						class="text-text-disabled hover:bg-surface-raised flex h-8 w-8 shrink-0 items-center justify-center rounded-full transition-colors"
						aria-label="Back to home">×</a
					>
				</div>

				<!-- Admin role badge -->
				{#if isAdmin}
					<div class="bg-surface-raised flex w-fit items-center gap-1.5 rounded-full px-3 py-1.5">
						<Pencil size={11} class="text-primary" />
						<span class="text-primary text-[10px] font-bold tracking-widest uppercase"
							>{$_('active_official')}</span
						>
					</div>
				{/if}

				<!-- Round header -->
				<div>
					<SectionLabel class="mb-3">
						{session.game_mode} · {session.rounds_total != null
							? $_('active_round_label')
							: $_('active_round_open_label')}
					</SectionLabel>
					<h2 class="text-[28px] leading-tight font-[800] tracking-tight">
						{session.rounds_total != null
							? $_('active_round_of', {
									values: { current: currentRound.number, total: session.rounds_total }
								})
							: $_('active_round_open', { values: { current: currentRound.number } })}
					</h2>

					<!-- Round indicator + timer row -->
					<div class="mt-4 flex items-center justify-between gap-4">
						{#if session.rounds_total != null}
							<RoundIndicator current={currentRound.number} total={session.rounds_total} />
						{:else}
							<div></div>
						{/if}

						{#if session.game_mode !== 'americano' && timeLeft !== null}
							<div class="text-text-secondary flex items-center gap-2 font-mono text-xs">
								<Clock size={14} />
								<span>{timeLeft}</span>
							</div>
						{/if}
					</div>
				</div>

				<!-- Time expired notice -->
				{#if session.game_mode !== 'americano' && timeExpired}
					<div class="border-warning/30 bg-warning/10 rounded-2xl border px-5 py-4 text-center">
						<p class="text-warning text-sm font-bold">
							{$_('active_time_expired')}
						</p>
					</div>
				{/if}

				<!-- Courts: Adaptive layout -->
				{#if isSingleCourt}
					<!-- Single court: ScoreBoard with inline steppers + tap-to-numpad -->
					{@const match = currentRound.matches[0]}
					{@const s = scores[match.id] ?? { a: 0, b: 0 }}
					{@const scored = match.score !== null && !editing[match.id]}

					<ScoreBoard
						teamA={{
							players: [playerById[match.team_a[0]], playerById[match.team_a[1]]].filter(Boolean),
							name: teamLabel(match.team_a),
							score: s.a
						}}
						teamB={{
							players: [playerById[match.team_b[0]], playerById[match.team_b[1]]].filter(Boolean),
							name: teamLabel(match.team_b),
							score: s.b
						}}
						{scored}
						live={!!match.live}
						pointsTarget={maxScore}
						{isAdmin}
						submitting={!!submitting[match.id]}
						onAdjust={(team, delta) => adjust(match.id, team, delta)}
						onScoreTap={(team) => openNumpad(match.id, team)}
						onFinalize={() => submitScore(match.id)}
					/>
				{:else}
					<!-- Multiple courts: Glanceable list layout -->
					<div class="space-y-3">
						{#each currentRound.matches as match}
							{@const s = scores[match.id] ?? { a: 0, b: 0 }}
							{@const scored = match.score !== null && !editing[match.id]}
							{@const p1 = playerById[match.team_a[0]]}
							{@const p2 = playerById[match.team_a[1]]}
							{@const p3 = playerById[match.team_b[0]]}
							{@const p4 = playerById[match.team_b[1]]}

							<Card
								class={`overflow-hidden rounded-2xl border shadow-sm transition-colors ${match.live ? 'border-primary/50 bg-primary-muted/30' : 'border-border bg-surface'}`}
							>
								<!-- Header row: Court label + Status -->
								<div class="border-border flex items-center justify-between border-b px-4 py-3">
									<span class="text-text-secondary text-[11px] font-bold tracking-widest uppercase">
										{$_('active_court_label', { values: { number: match.court } })}
									</span>
									{#if match.score !== null}
										<span
											class="bg-surface-raised text-text-primary inline-block rounded-full px-3 py-1 text-xs font-semibold"
										>
											{$_('court_status_final')}
										</span>
									{:else if match.live}
										<span
											class="bg-primary/10 text-primary inline-flex items-center gap-1 rounded-full px-3 py-1 text-xs font-semibold"
										>
											<span class="bg-primary inline-block h-2 w-2 animate-pulse rounded-full"
											></span>
											{$_('court_status_live')}
										</span>
									{:else}
										<span
											class="bg-surface-raised text-text-secondary inline-block rounded-full px-3 py-1 text-xs font-semibold"
										>
											{$_('court_status_upcoming')}
										</span>
									{/if}
								</div>

								<!-- Team A row -->
								<div
									class="flex items-center gap-3 px-4 py-3 {scored && s.a > s.b
										? 'bg-primary-muted'
										: ''}"
								>
									<div class="flex">
										<Avatar
											icon={p1?.avatar_icon}
											color={p1?.avatar_color}
											name={p1?.name ?? ''}
											size="sm"
										/>
										<div class="-ml-2">
											<Avatar
												icon={p2?.avatar_icon}
												color={p2?.avatar_color}
												name={p2?.name ?? ''}
												size="sm"
											/>
										</div>
									</div>
									<p class="flex-1 truncate text-sm font-semibold">{teamLabel(match.team_a)}</p>
									<TeamScore
										score={s.a}
										opponentScore={s.b}
										{scored}
										interactive={isAdmin}
										underline
										label="{scored ? 'Edit' : 'Set'} Team A score"
										onTap={() => openNumpad(match.id, 'a', 'final')}
									/>
								</div>

								<div class="bg-border mx-4 h-px"></div>

								<!-- Team B row -->
								<div
									class="flex items-center gap-3 px-4 py-3 {scored && s.b > s.a
										? 'bg-primary-muted'
										: ''}"
								>
									<div class="flex">
										<Avatar
											icon={p3?.avatar_icon}
											color={p3?.avatar_color}
											name={p3?.name ?? ''}
											size="sm"
										/>
										<div class="-ml-2">
											<Avatar
												icon={p4?.avatar_icon}
												color={p4?.avatar_color}
												name={p4?.name ?? ''}
												size="sm"
											/>
										</div>
									</div>
									<p class="flex-1 truncate text-sm font-semibold">{teamLabel(match.team_b)}</p>
									<TeamScore
										score={s.b}
										opponentScore={s.a}
										{scored}
										interactive={isAdmin}
										underline
										label="{scored ? 'Edit' : 'Set'} Team B score"
										onTap={() => openNumpad(match.id, 'b', 'final')}
									/>
								</div>
							</Card>
						{/each}
					</div>
				{/if}

				<!-- Bench -->
				{#if benchNames.length > 0}
					<div>
						<SectionLabel class="mb-3">{$_('active_bench')}</SectionLabel>
						<div class="bg-surface-raised rounded-2xl px-4 py-3">
							<div class="flex flex-wrap gap-2">
								{#each currentRound.bench as id}
									{@const p = playerById[id]}
									<div class="flex items-center gap-2">
										<Avatar
											icon={p?.avatar_icon}
											color={p?.avatar_color}
											name={p?.name ?? ''}
											size="sm"
										/>
										<span class="text-sm font-medium">{p?.name ?? id}</span>
									</div>
								{/each}
							</div>
						</div>
					</div>
				{/if}

				<!-- Admin: Official card -->
				{#if adminPlayer}
					<div class="bg-surface-raised flex items-center gap-3 rounded-2xl px-4 py-3">
						<Avatar
							icon={adminPlayer.avatar_icon}
							color={adminPlayer.avatar_color}
							name={adminPlayer.name}
							size="sm"
						/>
						<p class="text-text-secondary flex-1 text-sm">
							{$_('active_official_label')}:
							<span class="text-text-primary font-semibold">{adminPlayer.name}</span>
						</p>
						<Shield size={14} class="text-primary" />
					</div>
				{/if}

				<!-- Round actions (admin only) -->
				{#if isAdmin}
					{#if allScored}
						{@const isFinalRound =
							session.rounds_total != null && currentRound.number === session.rounds_total}
						<Button
							variant="default"
							size="cta"
							disabled={advancing}
							onclick={isFinalRound ? onRefresh : advanceRound}
						>
							{advancing
								? isFinalRound
									? $_('active_final_results_loading')
									: $_('active_next_round_loading')
								: isFinalRound
									? $_('active_final_results')
									: $_('active_next_round')}
						</Button>
					{:else if someScored}
						<Button variant="default" size="cta" disabled>
							{$_('active_next_round')}
						</Button>
						<p class="text-text-disabled text-center text-xs">{$_('active_courts_pending')}</p>
					{/if}

					<!-- End session button -->
					<Button
						variant="destructive-solid"
						size="cta"
						disabled={closing || cancelling}
						onclick={() => (showEndMenu = true)}
					>
						{$_('active_close')}
					</Button>
				{:else if allScored}
					<div
						class="bg-surface-raised text-text-secondary rounded-2xl px-4 py-3 text-center text-sm"
					>
						{$_('active_round_complete')}
					</div>
				{/if}
			</main>
		</div>
	</div>

	<!-- Standings bottom sheet -->
	<Sheet.Root bind:open={showStandingsSheet}>
		<Sheet.Content side="bottom" class="mx-auto w-full max-w-[480px]">
			<Sheet.Header>
				<Sheet.Title>{$_('standings_label')}</Sheet.Title>
			</Sheet.Header>
			<div class="max-h-[70vh] overflow-y-auto">
				<Leaderboard
					sessionId={session.id}
					sessionName={sessionName(session)}
					{stream}
					ratings={playerRatings}
				/>
			</div>
		</Sheet.Content>
	</Sheet.Root>

	<!-- End tournament bottom sheet -->
	<Sheet.Root bind:open={showEndMenu}>
		<Sheet.Content side="bottom" class="mx-auto w-full max-w-[480px]">
			<Sheet.Header>
				<Sheet.Title>{$_('active_close')}</Sheet.Title>
			</Sheet.Header>
			<Sheet.Footer>
				<Button
					variant="default"
					size="cta"
					disabled={closing || cancelling}
					onclick={() => {
						showEndMenu = false;
						closeSession();
					}}
				>
					{closing ? '…' : $_('end_menu_save')}
				</Button>
				<Button
					variant="destructive-solid"
					size="cta"
					disabled={closing || cancelling}
					onclick={() => {
						showEndMenu = false;
						cancelSession();
					}}
				>
					{cancelling ? '…' : $_('end_menu_discard')}
				</Button>
				<Sheet.Close
					class="border-border text-text-secondary hover:bg-surface-raised h-auto w-full rounded-2xl border px-4 py-3.5 text-sm font-semibold transition-colors"
				>
					{$_('end_menu_keep')}
				</Sheet.Close>
			</Sheet.Footer>
		</Sheet.Content>
	</Sheet.Root>
{/if}
