<script lang="ts">
  import { untrack } from 'svelte';
  import { fly } from 'svelte/transition';
  import { api } from '$lib/api/client';
  import { _ } from 'svelte-i18n';
  import { Activity, ChartBar, Users, Pencil, Shield, LayoutGrid, Check, X } from 'lucide-svelte';
  import { sessionName } from '$lib/utils';
  import Avatar from '$lib/components/ui/Avatar.svelte';
  import RoundIndicator from './RoundIndicator.svelte';
  import Leaderboard from './Leaderboard.svelte';
  import ConfirmDialog from './ConfirmDialog.svelte';

  let {
    session,
    currentRound,
    isAdmin,
    onRefresh,
  }: {
    session: App.Session;
    currentRound: App.Round;
    isAdmin: boolean;
    onRefresh: () => void;
  } = $props();

  type Tab = 'scoring' | 'standings' | 'players';
  let tab = $state<Tab>(untrack(() => currentRound.matches.every((m) => m.score !== null)) ? 'standings' : 'scoring');

  const playerName = $derived(Object.fromEntries(session.players.map((p) => [p.id, p.name])));
  const playerById = $derived(Object.fromEntries(session.players.map((p) => [p.id, p])));

  let localScores = $state<Record<string, { a: number; b: number }>>({});
  let submitting = $state<Record<string, boolean>>({});
  let submitError = $state<Record<string, string>>({});
  let editing = $state<Record<string, boolean>>({});
  let initialServer = $state<Record<string, 'a' | 'b'>>({});
  const saveTimeout: Record<string, ReturnType<typeof setTimeout>> = {};
  let advancing = $state(false);
  let showCancelDialog = $state(false);
  let cancelling = $state(false);
  let showCloseDialog = $state(false);
  let closing = $state(false);
  let actionError = $state('');

  // Court tabs
  let activeCourt = $state(0);
  $effect(() => {
    const max = currentRound.matches.length - 1;
    if (activeCourt > max) activeCourt = 0;
  });

  let showCourtsOverview = $state(false);

  // Numpad
  type NumpadState = { matchId: string; team: 'a' | 'b'; value: string };
  let numpad = $state<NumpadState | null>(null);
  let numpadShaking = $state(false);

  function openNumpad(matchId: string, team: 'a' | 'b') {
    const current = scores[matchId]?.[team] ?? 0;
    numpad = { matchId, team, value: current > 0 ? String(current) : '' };
  }

  function numpadDigit(d: string) {
    if (!numpad) return;
    const next = (numpad.value + d).replace(/^0+(\d)/, '$1');
    if (parseInt(next || '0') > session.points) return;
    numpad = { ...numpad, value: next };
  }

  function numpadDelete() {
    if (!numpad) return;
    numpad = { ...numpad, value: numpad.value.slice(0, -1) };
  }

  function numpadConfirm() {
    if (!numpad) return;
    const entered = parseInt(numpad.value || '0');
    if (entered > session.points) {
      numpadShaking = true;
      setTimeout(() => { numpadShaking = false; }, 400);
      return;
    }
    const other = session.points - entered;
    const { matchId, team } = numpad;
    localScores[matchId] = team === 'a' ? { a: entered, b: other } : { a: other, b: entered };
    scheduleLiveSave(matchId);
    numpad = null;
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
    localScores[matchId] = { ...s, [team]: Math.max(0, Math.min(session.points, s[team] + delta)) };
    scheduleLiveSave(matchId);
  }

  async function submitScore(matchId: string) {
    clearTimeout(saveTimeout[matchId]);
    submitError[matchId] = '';
    submitting[matchId] = true;
    const s = scores[matchId];
    try {
      await api.scores.submit(session.id, matchId, s.a, s.b, '');
      editing[matchId] = false;
      onRefresh();
    } catch (e) {
      submitError[matchId] = e instanceof Error ? e.message : 'Failed to submit';
    } finally {
      submitting[matchId] = false;
    }
  }

  async function closeSession() {
    closing = true;
    actionError = '';
    try {
      const adminToken = localStorage.getItem(`admin_token_${session.id}`) ?? '';
      await api.sessions.close(session.id, adminToken);
      showCloseDialog = false;
      closing = false;
      onRefresh();
    } catch (e) {
      actionError = e instanceof Error ? e.message : 'Could not end tournament';
      closing = false;
      setTimeout(() => { actionError = ''; }, 4000);
    }
  }

  async function cancelSession() {
    cancelling = true;
    actionError = '';
    try {
      const adminToken = localStorage.getItem(`admin_token_${session.id}`) ?? '';
      await api.sessions.cancel(session.id, adminToken);
      location.href = '/';
    } catch (e) {
      actionError = e instanceof Error ? e.message : 'Could not cancel tournament';
      cancelling = false;
      setTimeout(() => { actionError = ''; }, 4000);
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
  const adminPlayer = $derived(session.creator_player_id ? playerById[session.creator_player_id] : null);

  // Timer countdown
  let now = $state(Date.now());
  $effect(() => {
    const interval = setInterval(() => { now = Date.now(); }, 1000);
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
</script>

{#if actionError}
  <div transition:fly={{ y: -48, duration: 400 }} class="fixed inset-x-0 top-0 z-50 flex items-center justify-center bg-[var(--destructive)] px-4 py-3 text-sm font-semibold text-white">
    {actionError}
  </div>
{/if}

{#if cancelling}
  <main class="flex min-h-svh flex-col items-center justify-center gap-3 px-6">
    <div class="h-8 w-8 animate-spin rounded-full border-2 border-[var(--border)] border-t-[var(--primary)]"></div>
    <p class="text-sm text-[var(--text-secondary)]">{$_('lobby_cancelling')}</p>
  </main>
{:else}

<!-- Bottom nav -->
<div class="fixed bottom-0 left-0 right-0 z-10 flex border-t border-[var(--border)] bg-[var(--background)]/90 backdrop-blur-sm pb-[env(safe-area-inset-bottom)]">
  <button
    onclick={() => tab = 'scoring'}
    class="flex flex-1 flex-col items-center gap-1 py-3 transition-colors {tab === 'scoring' ? 'text-[var(--primary)]' : 'text-[var(--text-secondary)]'}"
  >
    <Activity size={20} />
    <span class="text-[10px] font-semibold uppercase tracking-wide">Scoring</span>
  </button>
  <button
    onclick={() => tab = 'standings'}
    class="flex flex-1 flex-col items-center gap-1 py-3 transition-colors {tab === 'standings' ? 'text-[var(--primary)]' : 'text-[var(--text-secondary)]'}"
  >
    <ChartBar size={20} />
    <span class="text-[10px] font-semibold uppercase tracking-wide">{$_('active_tab_standings')}</span>
  </button>
  <button
    onclick={() => tab = 'players'}
    class="flex flex-1 flex-col items-center gap-1 py-3 transition-colors {tab === 'players' ? 'text-[var(--primary)]' : 'text-[var(--text-secondary)]'}"
  >
    <Users size={20} />
    <span class="text-[10px] font-semibold uppercase tracking-wide">Players</span>
  </button>
</div>

<!-- ── SCORING TAB ── -->
{#if tab === 'scoring'}
  <main class="mx-auto max-w-[480px] px-4 pb-32 pt-6 space-y-4">

    <!-- Nav -->
    <div class="flex items-center justify-between">
      <p class="text-sm font-semibold text-[var(--primary)]">{sessionName(session)}</p>
      <a
        href="/"
        class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-[var(--text-disabled)] transition-colors hover:bg-[var(--surface-raised)] hover:text-[var(--text-secondary)]"
        aria-label="Back to home"
      >×</a>
    </div>

    <!-- Admin role badge -->
    {#if isAdmin}
      <div class="flex w-fit items-center gap-1.5 rounded-full bg-[var(--surface-raised)] px-3 py-1.5">
        <Pencil size={11} class="text-[var(--primary)]" />
        <span class="text-[10px] font-bold uppercase tracking-widest text-[var(--primary)]">Active Scorekeeper</span>
      </div>
    {/if}

    <!-- Match info row -->
    <div class="flex items-center justify-between gap-3">
      <div class="min-w-0">
        <p class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-disabled)]">
          {session.game_mode} tournament
        </p>
        <button onclick={() => showCourtsOverview = true} class="text-left">
          <h2 class="text-[28px] font-[800] leading-tight tracking-tight">
            {session.rounds_total != null
              ? $_('active_round_of', { values: { current: currentRound.number, total: session.rounds_total } })
              : $_('active_round_open', { values: { current: currentRound.number } })}
          </h2>
        </button>
      </div>
      <button
        onclick={() => showCourtsOverview = true}
        class="flex shrink-0 items-center gap-1.5 rounded-xl bg-[var(--surface-raised)] px-3 py-2 text-xs font-semibold text-[var(--text-secondary)] transition-colors hover:bg-[var(--border)]"
      >
        <LayoutGrid size={13} />
        Overview
      </button>
    </div>

    {#if timeLeft !== null}
      <p class="text-xs font-mono font-semibold text-[var(--text-secondary)]">⏱ {timeLeft}</p>
    {/if}

    {#if timeExpired}
      <div class="rounded-2xl border border-amber-500/30 bg-amber-500/10 px-5 py-4 text-center">
        <p class="text-sm font-bold text-amber-600 dark:text-amber-400">{$_('active_time_expired')}</p>
      </div>
    {/if}

    {#if session.rounds_total != null}
      <RoundIndicator current={currentRound.number} total={session.rounds_total} />
    {/if}

    <!-- Court tabs -->
    {#if currentRound.matches.length > 1}
      <div class="flex gap-2 overflow-x-auto pb-1 -mx-4 px-4 scrollbar-none">
        {#each currentRound.matches as match, i}
          {@const isFinalized = match.score !== null && !editing[match.id]}
          <button
            onclick={() => activeCourt = i}
            class="flex shrink-0 items-center gap-1.5 rounded-xl px-3 py-2 text-[11px] font-bold uppercase tracking-widest transition-colors
              {activeCourt === i ? 'bg-[var(--primary)] text-white' : 'bg-[var(--surface-raised)] text-[var(--text-secondary)]'}"
          >
            🎾 Court {match.court}
            {#if isFinalized}
              <span class="h-1.5 w-1.5 rounded-full {activeCourt === i ? 'bg-white/60' : 'bg-[var(--primary)]'}"></span>
            {/if}
          </button>
        {/each}
      </div>
    {/if}

    <!-- Score cards (one per court, only active shown) -->
    {#each currentRound.matches as match, i (match.id)}
      {#if i === activeCourt}
        {@const s = scores[match.id] ?? { a: 0, b: 0 }}
        {@const scored = match.score !== null && !editing[match.id]}
        {@const p1 = playerById[match.team_a[0]]}
        {@const p2 = playerById[match.team_a[1]]}
        {@const p3 = playerById[match.team_b[0]]}
        {@const p4 = playerById[match.team_b[1]]}

        {#if scored}
          <!-- Finalized compact result card (tap to re-enter) -->
          {@const sa = match.score!.a}
          {@const sb = match.score!.b}
          {@const isDraw = sa === sb}
          <button
            onclick={() => { localScores[match.id] = { a: sa, b: sb }; editing[match.id] = true; }}
            class="w-full rounded-3xl overflow-hidden text-left"
          >
            <div class="flex items-center gap-3 px-5 py-4
              {isDraw ? 'bg-[var(--surface-raised)]' : sa > sb ? 'bg-[var(--primary)]' : 'bg-[var(--surface-raised)]'}">
              <div class="flex">
                <Avatar icon={p1?.avatar_icon} color={p1?.avatar_color} name={p1?.name ?? ''} size="sm" />
                <div class="-ml-2 rounded-full ring-2 ring-[var(--background)]">
                  <Avatar icon={p2?.avatar_icon} color={p2?.avatar_color} name={p2?.name ?? ''} size="sm" />
                </div>
              </div>
              <p class="flex-1 font-semibold truncate
                {isDraw ? 'text-[var(--text-primary)]' : sa > sb ? 'text-white' : 'text-[var(--text-disabled)]'}">
                {teamLabel(match.team_a)}
              </p>
              <span class="text-2xl font-[800] tabular-nums
                {isDraw ? 'text-[var(--text-primary)]' : sa > sb ? 'text-white' : 'text-[var(--text-disabled)]'}">{sa}</span>
            </div>
            <div class="h-px bg-[var(--border)]"></div>
            <div class="flex items-center gap-3 px-5 py-4
              {isDraw ? 'bg-[var(--surface-raised)]' : sb > sa ? 'bg-[var(--primary)]' : 'bg-[var(--surface-raised)]'}">
              <div class="flex">
                <Avatar icon={p3?.avatar_icon} color={p3?.avatar_color} name={p3?.name ?? ''} size="sm" />
                <div class="-ml-2 rounded-full ring-2 ring-[var(--background)]">
                  <Avatar icon={p4?.avatar_icon} color={p4?.avatar_color} name={p4?.name ?? ''} size="sm" />
                </div>
              </div>
              <p class="flex-1 font-semibold truncate
                {isDraw ? 'text-[var(--text-primary)]' : sb > sa ? 'text-white' : 'text-[var(--text-disabled)]'}">
                {teamLabel(match.team_b)}
              </p>
              <span class="text-2xl font-[800] tabular-nums
                {isDraw ? 'text-[var(--text-primary)]' : sb > sa ? 'text-white' : 'text-[var(--text-disabled)]'}">{sb}</span>
            </div>
          </button>

        {:else}
          <!-- Active dark score card -->
          <div class="relative overflow-hidden rounded-3xl bg-[#3d7a24] px-5 pt-7 pb-6 space-y-2">
            <!-- Court line pattern -->
            <svg class="pointer-events-none absolute inset-0 h-full w-full opacity-10" preserveAspectRatio="none" viewBox="0 0 100 100">
              <line x1="50" y1="0" x2="50" y2="100" stroke="white" stroke-width="0.5"/>
              <line x1="0" y1="50" x2="100" y2="50" stroke="white" stroke-width="0.5"/>
              <rect x="10" y="10" width="80" height="80" fill="none" stroke="white" stroke-width="0.5"/>
            </svg>
            <div class="relative z-10 space-y-2">

            <!-- Team A: avatars + names -->
            <div class="flex flex-col items-center gap-2">
              <div class="flex justify-center">
                <Avatar icon={p1?.avatar_icon} color={p1?.avatar_color} name={p1?.name ?? ''} size="md" />
                <div class="-ml-3 rounded-full ring-2 ring-[#3d7a24]">
                  <Avatar icon={p2?.avatar_icon} color={p2?.avatar_color} name={p2?.name ?? ''} size="md" />
                </div>
              </div>
              <p class="text-base font-bold text-white">{teamLabel(match.team_a)}</p>
              <p class="text-[10px] font-bold uppercase tracking-widest text-white/50">Team A</p>
            </div>

            <!-- Team A score row -->
            <div class="flex items-center justify-between gap-2">
              <button
                onclick={() => adjust(match.id, 'a', -1)}
                disabled={s.a === 0}
                class="flex h-14 w-14 shrink-0 items-center justify-center rounded-full bg-white text-2xl font-bold text-[#3d7a24] shadow-sm transition-all active:scale-95 disabled:opacity-40"
              >−</button>
              <button
                onclick={() => openNumpad(match.id, 'a')}
                class="flex-1 text-center text-[80px] font-[800] leading-none tabular-nums text-white"
              >{s.a}</button>
              <button
                onclick={() => adjust(match.id, 'a', 1)}
                disabled={s.a + s.b >= session.points}
                class="flex h-14 w-14 shrink-0 items-center justify-center rounded-full bg-white text-2xl font-bold text-[#3d7a24] shadow-sm transition-all active:scale-95 disabled:opacity-40"
              >+</button>
            </div>

            <!-- Team B score row -->
            <div class="flex items-center justify-between gap-2">
              <button
                onclick={() => adjust(match.id, 'b', -1)}
                disabled={s.b === 0}
                class="flex h-14 w-14 shrink-0 items-center justify-center rounded-full bg-white text-2xl font-bold text-[#3d7a24] shadow-sm transition-all active:scale-95 disabled:opacity-40"
              >−</button>
              <button
                onclick={() => openNumpad(match.id, 'b')}
                class="flex-1 text-center text-[80px] font-[800] leading-none tabular-nums text-white"
              >{s.b}</button>
              <button
                onclick={() => adjust(match.id, 'b', 1)}
                disabled={s.a + s.b >= session.points}
                class="flex h-14 w-14 shrink-0 items-center justify-center rounded-full bg-white text-2xl font-bold text-[#3d7a24] shadow-sm transition-all active:scale-95 disabled:opacity-40"
              >+</button>
            </div>

            <!-- Team B: names + avatars -->
            <div class="flex flex-col items-center gap-2">
              <p class="text-[10px] font-bold uppercase tracking-widest text-white/50">Team B</p>
              <p class="text-base font-bold text-white">{teamLabel(match.team_b)}</p>
              <div class="flex justify-center">
                <Avatar icon={p3?.avatar_icon} color={p3?.avatar_color} name={p3?.name ?? ''} size="md" />
                <div class="-ml-3 rounded-full ring-2 ring-[#3d7a24]">
                  <Avatar icon={p4?.avatar_icon} color={p4?.avatar_color} name={p4?.name ?? ''} size="md" />
                </div>
              </div>
            </div>
          </div><!-- close z-10 wrapper -->
          </div><!-- close card -->

          <!-- Finalize button + helper text -->
          {#if isAdmin}
            {#if submitError[match.id]}
              <p class="text-center text-xs text-[var(--destructive)]">{submitError[match.id]}</p>
            {/if}
            <button
              onclick={() => submitScore(match.id)}
              disabled={s.a + s.b !== session.points || submitting[match.id]}
              class="flex w-full items-center justify-center gap-2 rounded-2xl bg-[var(--primary)] px-4 py-4 text-[15px] font-[700] text-white transition-all active:scale-[0.98] disabled:opacity-40"
            >
              <Check size={18} />
              {submitting[match.id] ? '…' : 'Finalize Result'}
            </button>
            <p class="text-center text-xs text-[var(--text-disabled)]">Scores are synced live to all player devices</p>
          {/if}
        {/if}
      {/if}
    {/each}

    <!-- Official card -->
    {#if adminPlayer}
      <div class="flex items-center gap-3 rounded-2xl bg-[var(--surface-raised)] px-4 py-3">
        <Avatar icon={adminPlayer.avatar_icon} color={adminPlayer.avatar_color} name={adminPlayer.name} size="sm" />
        <p class="flex-1 text-sm text-[var(--text-secondary)]">
          Official: <span class="font-semibold text-[var(--text-primary)]">{adminPlayer.name}</span>
        </p>
        <Shield size={14} class="text-[var(--primary)]" />
      </div>
    {/if}

    <!-- Bench -->
    {#if benchNames.length > 0}
      {@const benchPlayer = playerById[currentRound.bench[0]]}
      <div class="flex items-center gap-3 rounded-2xl bg-[var(--surface-raised)] px-4 py-3">
        <Avatar icon={benchPlayer?.avatar_icon} color={benchPlayer?.avatar_color} name={benchNames[0]} size="sm" />
        <p class="text-sm text-[var(--text-secondary)]">
          {$_('active_bench')}: <span class="font-semibold text-[var(--text-primary)]">{benchNames.join(', ')}</span>
        </p>
      </div>
    {/if}

    <!-- Next round / waiting -->
    {#if allScored && isAdmin}
      {@const isFinalRound = session.rounds_total != null && currentRound.number === session.rounds_total}
      <button
        onclick={isFinalRound ? onRefresh : advanceRound}
        disabled={advancing}
        class="w-full rounded-2xl bg-[var(--primary)] px-4 py-4 text-[15px] font-[700] text-white transition-all active:scale-[0.98] disabled:opacity-60"
      >
        {advancing ? '…' : isFinalRound ? $_('active_final_results') : $_('active_next_round')}
      </button>
    {:else if allScored}
      <div class="rounded-2xl bg-[var(--surface-raised)] px-4 py-3 text-center text-sm text-[var(--text-secondary)]">
        {$_('active_round_complete')}
      </div>
    {/if}

    <!-- Admin: end tournament -->
    {#if isAdmin}
      <div class="flex flex-col items-center gap-1 pb-2">
        <button
          onclick={() => showCloseDialog = true}
          disabled={closing || cancelling}
          class="rounded-full bg-[var(--destructive)] px-5 py-2 text-xs font-semibold text-white transition-all active:scale-95 disabled:opacity-40"
        >{$_('active_close')}</button>
        <button
          onclick={() => showCancelDialog = true}
          disabled={closing || cancelling}
          class="px-4 py-1.5 text-xs text-[var(--text-disabled)] transition-colors hover:text-[var(--destructive)] disabled:opacity-40"
        >{$_('active_cancel')}</button>
      </div>
    {/if}

  </main>

<!-- ── STANDINGS TAB ── -->
{:else if tab === 'standings'}
  <main class="mx-auto max-w-[480px] px-4 pb-28 pt-6">
    <div class="mb-4 flex items-center justify-between">
      <p class="text-sm font-semibold text-[var(--primary)]">{sessionName(session)}</p>
      <a href="/" class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-[var(--text-disabled)] transition-colors hover:bg-[var(--surface-raised)]" aria-label="Back to home">×</a>
    </div>
    <Leaderboard sessionId={session.id} sessionName={sessionName(session)} />
  </main>

<!-- ── PLAYERS TAB ── -->
{:else if tab === 'players'}
  <main class="mx-auto max-w-[480px] px-4 pb-28 pt-6 space-y-4">
    <div class="flex items-center justify-between">
      <p class="text-sm font-semibold text-[var(--primary)]">{sessionName(session)}</p>
      <a href="/" class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-[var(--text-disabled)] transition-colors hover:bg-[var(--surface-raised)]" aria-label="Back to home">×</a>
    </div>

    <!-- Bench (prominently at top) -->
    {#if currentRound.bench.length > 0}
      <div>
        <p class="mb-2 text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">
          Sitting out ({currentRound.bench.length})
        </p>
        <div class="divide-y divide-[var(--border)] rounded-2xl bg-[var(--surface-raised)]">
          {#each currentRound.bench as id}
            {@const p = playerById[id]}
            <div class="flex items-center gap-3 px-4 py-3">
              <Avatar icon={p?.avatar_icon} color={p?.avatar_color} name={p?.name ?? ''} size="sm" />
              <span class="flex-1 text-sm font-medium">{p?.name ?? id}</span>
              <span class="rounded-full bg-[var(--border)] px-2 py-0.5 text-[10px] font-bold text-[var(--text-disabled)]">Bench</span>
            </div>
          {/each}
        </div>
      </div>
    {/if}

    <!-- All active players -->
    {#each [session.players.filter(p => p.active)] as activePlayers}
    <div>
      <p class="mb-2 text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">
        Players ({activePlayers.length})
      </p>
      <div class="divide-y divide-[var(--border)] rounded-2xl bg-[var(--surface-raised)]">
        {#each activePlayers as player (player.id)}
          <div class="flex items-center gap-3 px-4 py-3">
            <Avatar icon={player.avatar_icon} color={player.avatar_color} name={player.name} size="sm" />
            <span class="flex-1 text-sm font-medium">{player.name}</span>
            {#if player.id === session.creator_player_id}
              <span class="text-[10px] font-bold text-[var(--primary)]">Admin</span>
            {:else if benchIds.has(player.id)}
              <span class="text-[10px] text-[var(--text-disabled)]">Bench</span>
            {/if}
          </div>
        {/each}
      </div>
    </div>
    {/each}
  </main>
{/if}

{/if}

<!-- ── COURTS OVERVIEW BOTTOM SHEET ── -->
{#if showCourtsOverview}
  <div
    role="presentation"
    class="fixed inset-0 z-40 bg-black/40"
    onclick={() => showCourtsOverview = false}
    onkeydown={(e) => e.key === 'Escape' && (showCourtsOverview = false)}
  ></div>
  <div class="fixed inset-x-0 bottom-0 z-50 rounded-t-3xl bg-[var(--surface)] px-4 pt-5 pb-[max(1.5rem,env(safe-area-inset-bottom))] space-y-3">
    <div class="mb-1 flex items-center justify-between">
      <h3 class="text-lg font-[800]">Courts Overview</h3>
      <button onclick={() => showCourtsOverview = false} class="text-[var(--text-disabled)] hover:text-[var(--text-secondary)]">
        <X size={20} />
      </button>
    </div>
    <div class="space-y-2">
      {#each currentRound.matches as match}
        {@const s = scores[match.id] ?? { a: 0, b: 0 }}
        {@const isFinalized = match.score !== null}
        {@const inProgress = !isFinalized && (s.a + s.b > 0)}
        <div class="flex items-center gap-3 rounded-2xl bg-[var(--surface-raised)] px-4 py-3">
          <div class="flex w-8 shrink-0 flex-col items-center">
            <p class="text-[9px] font-bold uppercase tracking-widest text-[var(--text-disabled)]">C</p>
            <p class="text-xl font-[800] leading-tight">{match.court}</p>
          </div>
          <div class="min-w-0 flex-1">
            <p class="truncate text-sm font-semibold">{teamLabel(match.team_a)}</p>
            <p class="truncate text-xs text-[var(--text-secondary)]">vs {teamLabel(match.team_b)}</p>
          </div>
          <div class="flex items-center gap-2">
            <span class="text-lg font-[800] tabular-nums">{s.a}–{s.b}</span>
            {#if isFinalized}
              <Check size={14} class="text-[var(--primary)]" />
            {:else if inProgress}
              <div class="h-2 w-2 rounded-full bg-amber-400"></div>
            {:else}
              <div class="h-2 w-2 rounded-full bg-[var(--border)]"></div>
            {/if}
          </div>
        </div>
      {/each}
    </div>
  </div>
{/if}

<!-- ── NUMPAD BOTTOM SHEET ── -->
{#if numpad}
  <div
    role="presentation"
    class="fixed inset-0 z-40 bg-black/40"
    onclick={() => numpad = null}
    onkeydown={(e) => e.key === 'Escape' && (numpad = null)}
  ></div>
  <div class="fixed inset-x-0 bottom-0 z-50 mx-auto max-w-[480px] rounded-t-3xl bg-[var(--surface)] px-5 pt-6 pb-[max(1.5rem,env(safe-area-inset-bottom))]">
    <p class="mb-3 text-center text-[10px] font-bold uppercase tracking-widest text-[var(--text-disabled)]">
      Target: {session.points}
    </p>
    <p class="mb-6 text-center text-[64px] font-[800] leading-none tabular-nums transition-transform
      {numpadShaking ? 'animate-[shake_0.4s_ease-in-out]' : ''}">
      {numpad.value || '0'}
    </p>
    <div class="grid grid-cols-3 gap-3">
      {#each ['1','2','3','4','5','6','7','8','9'] as d}
        <button
          onclick={() => numpadDigit(d)}
          class="rounded-2xl bg-[var(--surface-raised)] py-4 text-xl font-[800] transition-all active:scale-95"
        >{d}</button>
      {/each}
      <button onclick={numpadDelete} class="rounded-2xl bg-[var(--surface-raised)] py-4 text-xl font-[800] transition-all active:scale-95">⌫</button>
      <button onclick={() => numpadDigit('0')} class="rounded-2xl bg-[var(--surface-raised)] py-4 text-xl font-[800] transition-all active:scale-95">0</button>
      <button onclick={numpadConfirm} class="rounded-2xl bg-[var(--primary)] py-4 text-xl font-[800] text-white transition-all active:scale-95">✓</button>
    </div>
  </div>
{/if}

<ConfirmDialog
  bind:open={showCloseDialog}
  title={$_('close_dialog_title')}
  description={$_('close_dialog_desc')}
  confirmLabel={$_('close_dialog_confirm')}
  cancelLabel={$_('close_dialog_cancel')}
  destructive
  onconfirm={closeSession}
/>

<ConfirmDialog
  bind:open={showCancelDialog}
  title={$_('cancel_dialog_title')}
  description={$_('cancel_dialog_desc')}
  confirmLabel={$_('cancel_dialog_confirm')}
  cancelLabel={$_('cancel_dialog_cancel')}
  destructive
  onconfirm={cancelSession}
/>
