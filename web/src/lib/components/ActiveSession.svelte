<script lang="ts">
  import { untrack } from 'svelte';
  import { toast } from 'svelte-sonner';
  import { ApiError } from '$lib/api/client';
  import { translateApiError } from '$lib/i18n/errors';
  import { api } from '$lib/api/client';
  import { _ } from 'svelte-i18n';
  import { sessionDialog } from '$lib/stores/sessionDialog';
  import { Activity, ChartBar, Users, Pencil, Shield, LayoutGrid, Check, X } from 'lucide-svelte';
  import { sessionName } from '$lib/utils';
  import Avatar from '$lib/components/ui/Avatar.svelte';
  import { SectionLabel } from '$lib/components/ui/section-label';
  import { ToggleGroup, ToggleGroupItem } from '$lib/components/ui/toggle-group';
  import RoundIndicator from './RoundIndicator.svelte';
  import Leaderboard from './Leaderboard.svelte';
  import { numpad as numpadStore } from '$lib/stores/numpad';

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
  let editing = $state<Record<string, boolean>>({});
  let initialServer = $state<Record<string, 'a' | 'b'>>({});
  const saveTimeout: Record<string, ReturnType<typeof setTimeout>> = {};
  let advancing = $state(false);
  let cancelling = $state(false);
  let closing = $state(false);

  // Court tabs
  let activeCourt = $state(0);
  $effect(() => {
    const max = currentRound.matches.length - 1;
    if (activeCourt > max) activeCourt = 0;
  });

  let showCourtsOverview = $state(false);

  // Numpad (mobile-optimized: drag-to-close, keyboard input, overwrite)
  type NumpadState = { matchId: string; team: 'a' | 'b'; value: string; fresh: boolean };
  let numpad = $state<NumpadState | null>(null);

  function openNumpad(matchId: string, team: 'a' | 'b') {
    const current = scores[matchId]?.[team] ?? 0;
    const value = current > 0 ? String(current) : '';
    numpad = { matchId, team, value, fresh: true };
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
    // Overwrite on first digit after opening/confirming, then append normally
    let next: string;
    if (numpad.fresh && numpad.value && numpad.value !== '0') {
      next = d; // Replace value on first digit
    } else {
      next = (numpad.value + d).replace(/^0+(\d)/, '$1');
    }
    if (parseInt(next || '0') > session.points) return;
    numpad = { ...numpad, value: next, fresh: false }; // After first digit, normal append mode
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
    if (entered > session.points) {
      numpadStore.update({ shaking: true });
      setTimeout(() => {
        numpadStore.update({ shaking: false });
      }, 400);
      return;
    }
    const other = session.points - entered;
    const { matchId, team } = numpad;
    localScores[matchId] = team === 'a' ? { a: entered, b: other } : { a: other, b: entered };
    scheduleLiveSave(matchId);
    numpad = null;
    numpadStore.close();
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

  const someScored = $derived(
    currentRound.matches.some((m) => m.score !== null && !editing[m.id])
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
    submitting[matchId] = true;
    const s = scores[matchId];
    try {
      await api.scores.submit(session.id, matchId, s.a, s.b, '');
      editing[matchId] = false;
      toast.success($_('toast_score_confirmed'));
      onRefresh();
    } catch (e) {
      toast.error(e instanceof ApiError ? translateApiError(e.message) : translateApiError('server_error'));
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
      toast.error(e instanceof ApiError ? translateApiError(e.message) : translateApiError('server_error'));
      closing = false;
    }
  }

  async function cancelSession() {
    cancelling = true;
    try {
      const adminToken = localStorage.getItem(`admin_token_${session.id}`) ?? '';
      await api.sessions.cancel(session.id, adminToken);
      location.href = '/';
    } catch (e) {
      toast.error(e instanceof ApiError ? translateApiError(e.message) : translateApiError('server_error'));
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


{#if cancelling}
  <main class="flex min-h-svh flex-col items-center justify-center gap-3 px-6">
    <div class="h-8 w-8 animate-spin rounded-full border-2 border-border border-t-primary"></div>
    <p class="text-sm text-text-secondary">{$_('lobby_cancelling')}</p>
  </main>
{:else}
<div class="flex flex-col h-full">
  <div class="flex-1 min-h-0 overflow-y-auto">

<!-- ── SCORING TAB ── -->
{#if tab === 'scoring'}
  <main class="mx-auto w-full max-w-[480px] px-4 pb-6 pt-safe-page space-y-4">

    <!-- Nav -->
    <div class="flex items-center justify-between">
      <p class="text-sm font-semibold text-primary">{sessionName(session)}</p>
      <a
        href="/"
        class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-text-disabled transition-colors hover:bg-surface-raised hover:text-text-secondary"
        aria-label="Back to home"
      >×</a>
    </div>

    <!-- Admin role badge -->
    {#if isAdmin}
      <div class="flex w-fit items-center gap-1.5 rounded-full bg-surface-raised px-3 py-1.5">
        <Pencil size={11} class="text-primary" />
        <span class="text-[10px] font-bold uppercase tracking-widest text-primary">Active Scorekeeper</span>
      </div>
    {/if}

    <!-- Match info row -->
    <div class="flex items-center justify-between gap-3">
      <div class="min-w-0">
        <p class="text-[10px] font-bold uppercase tracking-widest text-text-disabled">
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
        class="flex shrink-0 items-center gap-1.5 rounded-xl bg-surface-raised px-3 py-2 text-xs font-semibold text-text-secondary transition-colors hover:bg-border"
      >
        <LayoutGrid size={13} />
        Overview
      </button>
    </div>

    {#if session.game_mode !== 'americano'}
      {#if timeLeft !== null}
        <p class="text-xs font-mono font-semibold text-text-secondary">⏱ {timeLeft}</p>
      {/if}

      {#if timeExpired}
        <div class="rounded-2xl border border-amber-500/30 bg-amber-500/10 px-5 py-4 text-center">
          <p class="text-sm font-bold text-amber-600 dark:text-amber-400">{$_('active_time_expired')}</p>
        </div>
      {/if}
    {/if}

    {#if session.rounds_total != null}
      <RoundIndicator current={currentRound.number} total={session.rounds_total} />
    {/if}

    <!-- Court tabs -->
    {#if currentRound.matches.length > 1}
      <ToggleGroup
        type="single"
        value={activeCourt.toString()}
        onValueChange={(val) => activeCourt = parseInt(val)}
        class="flex gap-2 overflow-x-auto pb-1 -mx-4 px-4 scrollbar-none"
      >
        {#each currentRound.matches as match, i}
          {@const isFinalized = match.score !== null && !editing[match.id]}
          <ToggleGroupItem
            value={i.toString()}
            class="flex shrink-0 items-center gap-1.5 rounded-xl px-3 py-2 text-[11px] font-bold uppercase tracking-widest transition-colors bg-surface-raised text-text-secondary data-[state=on]:bg-primary data-[state=on]:text-white"
          >
            🎾 Court {match.court}
            {#if isFinalized}
              <span class="h-1.5 w-1.5 rounded-full {activeCourt === i ? 'bg-white/60' : 'bg-primary'}"></span>
            {/if}
          </ToggleGroupItem>
        {/each}
      </ToggleGroup>
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
            class="w-full rounded-3xl overflow-hidden border border-primary/40 text-left"
          >
            <div class="flex items-center gap-3 px-5 py-4
              {isDraw ? 'bg-surface-raised' : sa > sb ? 'bg-primary' : 'bg-surface-raised'}">
              <div class="flex">
                <Avatar icon={p1?.avatar_icon} color={p1?.avatar_color} name={p1?.name ?? ''} size="sm" ring={!isDraw && sa > sb ? 'ring-2 ring-white/30' : 'ring-2 ring-primary/30'} />
                <div class="-ml-2">
                  <Avatar icon={p2?.avatar_icon} color={p2?.avatar_color} name={p2?.name ?? ''} size="sm" ring={!isDraw && sa > sb ? 'ring-2 ring-white/30' : 'ring-2 ring-primary/30'} />
                </div>
              </div>
              <p class="flex-1 font-semibold truncate
                {isDraw ? 'text-text-primary' : sa > sb ? 'text-white' : 'text-text-disabled'}">
                {teamLabel(match.team_a)}
              </p>
              <span class="text-2xl font-[800] tabular-nums
                {isDraw ? 'text-text-primary' : sa > sb ? 'text-white' : 'text-text-disabled'}">{sa}</span>
            </div>
            <div class="h-px bg-border"></div>
            <div class="flex items-center gap-3 px-5 py-4
              {isDraw ? 'bg-surface-raised' : sb > sa ? 'bg-primary' : 'bg-surface-raised'}">
              <div class="flex">
                <Avatar icon={p3?.avatar_icon} color={p3?.avatar_color} name={p3?.name ?? ''} size="sm" ring={!isDraw && sb > sa ? 'ring-2 ring-white/30' : 'ring-2 ring-primary/30'} />
                <div class="-ml-2">
                  <Avatar icon={p4?.avatar_icon} color={p4?.avatar_color} name={p4?.name ?? ''} size="sm" ring={!isDraw && sb > sa ? 'ring-2 ring-white/30' : 'ring-2 ring-primary/30'} />
                </div>
              </div>
              <p class="flex-1 font-semibold truncate
                {isDraw ? 'text-text-primary' : sb > sa ? 'text-white' : 'text-text-disabled'}">
                {teamLabel(match.team_b)}
              </p>
              <span class="text-2xl font-[800] tabular-nums
                {isDraw ? 'text-text-primary' : sb > sa ? 'text-white' : 'text-text-disabled'}">{sb}</span>
            </div>
          </button>

        {:else}
          <!-- Team A card -->
          <div class="relative overflow-hidden rounded-3xl border border-white/25 bg-[#3d7a24] px-5 pt-6 pb-5">
            <svg class="pointer-events-none absolute inset-0 h-full w-full opacity-10" preserveAspectRatio="none" viewBox="0 0 100 100">
              <line x1="50" y1="0" x2="50" y2="100" stroke="white" stroke-width="0.5"/>
              <rect x="10" y="10" width="80" height="80" fill="none" stroke="white" stroke-width="0.5"/>
            </svg>
            <div class="relative z-10 space-y-3">
              <div class="flex flex-col items-center gap-2">
                <div class="flex justify-center">
                  <Avatar icon={p1?.avatar_icon} color={p1?.avatar_color} name={p1?.name ?? ''} size="md" ring="ring-2 ring-white/30" />
                  <div class="-ml-3">
                    <Avatar icon={p2?.avatar_icon} color={p2?.avatar_color} name={p2?.name ?? ''} size="md" ring="ring-2 ring-white/30" />
                  </div>
                </div>
                <p class="text-base font-bold text-white">{teamLabel(match.team_a)}</p>
                <p class="text-[10px] font-bold uppercase tracking-widest text-white/50">Team A</p>
              </div>
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
            </div>
          </div>

          <!-- Team B card -->
          <div class="relative overflow-hidden rounded-3xl bg-[#3d7a24] px-5 pt-5 pb-6">
            <svg class="pointer-events-none absolute inset-0 h-full w-full opacity-10" preserveAspectRatio="none" viewBox="0 0 100 100">
              <line x1="50" y1="0" x2="50" y2="100" stroke="white" stroke-width="0.5"/>
              <rect x="10" y="10" width="80" height="80" fill="none" stroke="white" stroke-width="0.5"/>
            </svg>
            <div class="relative z-10 space-y-3">
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
              <div class="flex flex-col items-center gap-2">
                <p class="text-[10px] font-bold uppercase tracking-widest text-white/50">Team B</p>
                <p class="text-base font-bold text-white">{teamLabel(match.team_b)}</p>
                <div class="flex justify-center">
                  <Avatar icon={p3?.avatar_icon} color={p3?.avatar_color} name={p3?.name ?? ''} size="md" ring="ring-2 ring-white/30" />
                  <div class="-ml-3">
                    <Avatar icon={p4?.avatar_icon} color={p4?.avatar_color} name={p4?.name ?? ''} size="md" ring="ring-2 ring-white/30" />
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Finalize button + helper text -->
          {#if isAdmin}
            <button
              onclick={() => submitScore(match.id)}
              disabled={s.a + s.b !== session.points || submitting[match.id]}
              class="flex w-full items-center justify-center gap-2 rounded-2xl bg-primary px-4 py-4 text-[15px] font-[700] text-white transition-all active:scale-[0.98] disabled:opacity-40"
            >
              <Check size={18} />
              {submitting[match.id] ? '…' : $_('active_finalize_result')}
            </button>
            <p class="text-center text-xs text-text-disabled">{$_('active_scores_synced')}</p>
          {/if}
        {/if}
      {/if}
    {/each}

    <!-- Official card -->
    {#if adminPlayer}
      <div class="flex items-center gap-3 rounded-2xl bg-surface-raised px-4 py-3">
        <Avatar icon={adminPlayer.avatar_icon} color={adminPlayer.avatar_color} name={adminPlayer.name} size="sm" ring="ring-2 ring-primary/30" />
        <p class="flex-1 text-sm text-text-secondary">
          Official: <span class="font-semibold text-text-primary">{adminPlayer.name}</span>
        </p>
        <Shield size={14} class="text-primary" />
      </div>
    {/if}

    <!-- Bench -->
    {#if benchNames.length > 0}
      {@const benchPlayer = playerById[currentRound.bench[0]]}
      <div class="flex items-center gap-3 rounded-2xl bg-surface-raised px-4 py-3">
        <Avatar icon={benchPlayer?.avatar_icon} color={benchPlayer?.avatar_color} name={benchNames[0]} size="sm" ring="ring-2 ring-primary/30" />
        <p class="text-sm text-text-secondary">
          {$_('active_bench')}: <span class="font-semibold text-text-primary">{benchNames.join(', ')}</span>
        </p>
      </div>
    {/if}

    <!-- Next round / waiting -->
    {#if allScored && isAdmin}
      {@const isFinalRound = session.rounds_total != null && currentRound.number === session.rounds_total}
      <button
        onclick={isFinalRound ? onRefresh : advanceRound}
        disabled={advancing}
        class="w-full rounded-2xl bg-primary px-4 py-4 text-[15px] font-[700] text-white transition-all active:scale-[0.98] disabled:opacity-60"
      >
        {advancing ? '…' : isFinalRound ? $_('active_final_results') : $_('active_next_round')}
      </button>
    {:else if someScored && !allScored && isAdmin}
      <button
        disabled
        class="w-full rounded-2xl bg-primary px-4 py-4 text-[15px] font-[700] text-white disabled:opacity-40"
      >
        {$_('active_next_round')}
      </button>
      <p class="text-center text-xs text-text-disabled">{$_('active_courts_pending')}</p>
    {:else if allScored}
      <div class="rounded-2xl bg-surface-raised px-4 py-3 text-center text-sm text-text-secondary">
        {$_('active_round_complete')}
      </div>
    {/if}

    <!-- Admin: end tournament -->
    {#if isAdmin}
      <div class="flex flex-col items-center gap-1 pb-2">
        <button
          onclick={() => sessionDialog.open('close', closeSession)}
          disabled={closing || cancelling}
          class="rounded-full bg-destructive px-5 py-2 text-xs font-semibold text-white transition-all active:scale-95 disabled:opacity-40"
        >{$_('active_close')}</button>
        <button
          onclick={() => sessionDialog.open('cancel', cancelSession)}
          disabled={closing || cancelling}
          class="px-4 py-1.5 text-xs text-text-disabled transition-colors hover:text-destructive disabled:opacity-40"
        >{$_('active_cancel')}</button>
      </div>
    {/if}

  </main>

<!-- ── STANDINGS TAB ── -->
{:else if tab === 'standings'}
  <main class="mx-auto w-full max-w-[480px] px-4 pb-6 pt-safe-page">
    <div class="mb-4 flex items-center justify-between">
      <p class="text-sm font-semibold text-primary">{sessionName(session)}</p>
      <a href="/" class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-text-disabled transition-colors hover:bg-surface-raised" aria-label="Back to home">×</a>
    </div>
    <Leaderboard sessionId={session.id} sessionName={sessionName(session)} />
  </main>

<!-- ── PLAYERS TAB ── -->
{:else if tab === 'players'}
  <main class="mx-auto w-full max-w-[480px] px-4 pb-6 pt-safe-page space-y-4">
    <div class="flex items-center justify-between">
      <p class="text-sm font-semibold text-primary">{sessionName(session)}</p>
      <a href="/" class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-text-disabled transition-colors hover:bg-surface-raised" aria-label="Back to home">×</a>
    </div>

    <!-- On court -->
    {#each [session.players.filter(p => p.active && !benchIds.has(p.id))] as onCourt}
    <div>
      <SectionLabel class="mb-2">
        On court ({onCourt.length})
      </SectionLabel>
      <div class="divide-y divide-border rounded-2xl bg-surface-raised">
        {#each onCourt as player (player.id)}
          <div class="flex items-center gap-3 px-4 py-3">
            <Avatar icon={player.avatar_icon} color={player.avatar_color} name={player.name} size="sm" ring="ring-2 ring-primary/30" />
            <span class="flex-1 text-sm font-medium">{player.name}</span>
            {#if player.id === session.creator_player_id}
              <span class="text-[10px] font-bold text-primary">Admin</span>
            {/if}
          </div>
        {/each}
      </div>
    </div>
    {/each}

    <!-- Bench -->
    {#if currentRound.bench.length > 0}
      <div>
        <SectionLabel class="mb-2">
          Bench ({currentRound.bench.length})
        </SectionLabel>
        <div class="divide-y divide-border rounded-2xl bg-surface-raised">
          {#each currentRound.bench as id}
            {@const p = playerById[id]}
            <div class="flex items-center gap-3 px-4 py-3">
              <Avatar icon={p?.avatar_icon} color={p?.avatar_color} name={p?.name ?? ''} size="sm" ring="ring-2 ring-primary/30" />
              <span class="flex-1 text-sm font-medium">{p?.name ?? id}</span>
            </div>
          {/each}
        </div>
      </div>
    {/if}
  </main>
{/if}

  </div><!-- end scroll wrapper -->

  <!-- Bottom nav: plain flex child, always at bottom -->
  <div class="shrink-0 flex border-t border-border bg-background backdrop-blur-sm shadow-lg" style="padding-bottom: max(1.5rem, env(safe-area-inset-bottom));">
    <button
      onclick={() => tab = 'scoring'}
      class="flex flex-1 flex-col items-center gap-1 py-3 transition-colors {tab === 'scoring' ? 'text-primary' : 'text-text-secondary'}"
    >
      <Activity size={20} />
      <span class="text-[10px] font-semibold uppercase tracking-wide">Scoring</span>
    </button>
    <button
      onclick={() => tab = 'standings'}
      class="flex flex-1 flex-col items-center gap-1 py-3 transition-colors {tab === 'standings' ? 'text-primary' : 'text-text-secondary'}"
    >
      <ChartBar size={20} />
      <span class="text-[10px] font-semibold uppercase tracking-wide">{$_('active_tab_standings')}</span>
    </button>
    <button
      onclick={() => tab = 'players'}
      class="flex flex-1 flex-col items-center gap-1 py-3 transition-colors {tab === 'players' ? 'text-primary' : 'text-text-secondary'}"
    >
      <Users size={20} />
      <span class="text-[10px] font-semibold uppercase tracking-wide">Players</span>
    </button>
  </div>
</div><!-- end flex col h-full -->

{/if}

<!-- ── COURTS OVERVIEW BOTTOM SHEET ── -->
{#if showCourtsOverview}
  <div
    role="presentation"
    class="fixed inset-0 z-40 bg-black/40"
    onclick={() => showCourtsOverview = false}
    onkeydown={(e) => e.key === 'Escape' && (showCourtsOverview = false)}
  ></div>
  <div class="fixed inset-x-0 bottom-0 z-50 rounded-t-3xl bg-surface px-4 pt-5 pb-[max(1.5rem,env(safe-area-inset-bottom))] space-y-3">
    <div class="mb-1 flex items-center justify-between">
      <h3 class="text-lg font-[800]">Courts Overview</h3>
      <button onclick={() => showCourtsOverview = false} class="text-text-disabled hover:text-text-secondary">
        <X size={20} />
      </button>
    </div>
    <div class="space-y-2">
      {#each currentRound.matches as match}
        {@const s = scores[match.id] ?? { a: 0, b: 0 }}
        {@const isFinalized = match.score !== null}
        {@const inProgress = !isFinalized && (s.a + s.b > 0)}
        <div class="flex items-center gap-3 rounded-2xl bg-surface-raised px-4 py-3">
          <div class="flex w-8 shrink-0 flex-col items-center">
            <p class="text-[9px] font-bold uppercase tracking-widest text-text-disabled">C</p>
            <p class="text-xl font-[800] leading-tight">{match.court}</p>
          </div>
          <div class="min-w-0 flex-1">
            <p class="truncate text-sm font-semibold">{teamLabel(match.team_a)}</p>
            <p class="truncate text-xs text-text-secondary">vs {teamLabel(match.team_b)}</p>
          </div>
          <div class="flex items-center gap-2">
            <span class="text-lg font-[800] tabular-nums">{s.a}–{s.b}</span>
            {#if isFinalized}
              <Check size={14} class="text-primary" />
            {:else if inProgress}
              <div class="h-2 w-2 rounded-full bg-amber-400"></div>
            {:else}
              <div class="h-2 w-2 rounded-full bg-border"></div>
            {/if}
          </div>
        </div>
      {/each}
    </div>
  </div>
{/if}

