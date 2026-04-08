<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { api } from '$lib/api/client';
  import { _ } from 'svelte-i18n';
  import { shortName, sessionName } from '$lib/utils';
  import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';

  let {
    session,
    isAdmin,
    onRefresh,
  }: {
    session: App.Session;
    isAdmin: boolean;
    onRefresh: () => void;
  } = $props();

  let match = $state<App.TennisMatch | null>(null);
  let addingPoint = $state<'a' | 'b' | null>(null);
  let showCloseDialog = $state(false);
  let showCancelDialog = $state(false);
  let interval: ReturnType<typeof setInterval>;

  async function loadMatch() {
    try {
      match = await api.tennis.getMatch(session.id);
    } catch {
      // swallow — keep polling
    }
  }

  onMount(() => {
    loadMatch();
    interval = setInterval(loadMatch, 2000);
  });
  onDestroy(() => clearInterval(interval));

  async function setServer(team: 'a' | 'b') {
    if (match?.state.winner) return;
    try { match = await api.tennis.setServer(session.id, team); } catch {}
  }

  async function addPoint(team: 'a' | 'b') {
    if (addingPoint || match?.state.winner) return;
    addingPoint = team;
    try {
      match = await api.tennis.addPoint(session.id, team);
      if (match.state.winner) {
        onRefresh();
      }
    } catch {
      // swallow
    } finally {
      addingPoint = null;
    }
  }

  async function closeMatch() {
    const adminToken = localStorage.getItem(`admin_token_${session.id}`) ?? '';
    try {
      await api.sessions.close(session.id, adminToken);
      showCloseDialog = false;
      onRefresh();
    } catch {
      // Error closing, stay on current view
    }
  }

  async function cancelMatch() {
    const adminToken = localStorage.getItem(`admin_token_${session.id}`) ?? '';
    try {
      await api.sessions.cancel(session.id, adminToken);
      location.href = '/';
    } catch {
      // Error canceling, stay on current view
    }
  }

  function pointLabel(p: number): string {
    return ['0', '15', '30', '40'][p] ?? '40';
  }

  const state = $derived(match?.state as App.TennisState | undefined);
  const teamANames = $derived(match?.teams.a.map((t: App.TennisTeam) => shortName(t.name)).join(' & ') ?? '');
  const teamBNames = $derived(match?.teams.b.map((t: App.TennisTeam) => shortName(t.name)).join(' & ') ?? '');

  // Current game score — golden point at 3:3
  const gameScoreA = $derived(
    state?.in_tiebreak
      ? (state.tiebreak_a ?? 0).toString()
      : state?.points_a === 3 && state?.points_b === 3
      ? 'GP'
      : pointLabel(state?.points_a ?? 0)
  );
  const gameScoreB = $derived(
    state?.in_tiebreak
      ? (state.tiebreak_b ?? 0).toString()
      : state?.points_a === 3 && state?.points_b === 3
      ? 'GP'
      : pointLabel(state?.points_b ?? 0)
  );

  const scoreLeadsA = $derived(
    state?.in_tiebreak
      ? (state.tiebreak_a ?? 0) > (state.tiebreak_b ?? 0)
      : (state?.points_a ?? 0) > (state?.points_b ?? 0)
  );
  const scoreLeadsB = $derived(
    state?.in_tiebreak
      ? (state.tiebreak_b ?? 0) > (state.tiebreak_a ?? 0)
      : (state?.points_b ?? 0) > (state?.points_a ?? 0)
  );


</script>

<main class="mx-auto max-w-[480px] px-4 py-6 space-y-4">

  <!-- Nav -->
  <nav class="flex items-center justify-between px-2">
    <div class="space-y-0.5">
      <p class="text-xs text-[var(--text-secondary)]">{$_('create_mode_tennis')}</p>
      <p class="text-sm font-semibold text-[var(--primary)]">{sessionName(session)}</p>
    </div>
    <div class="flex items-center gap-3">
      <p class="text-xs text-[var(--text-secondary)]">{$_(session.sets_to_win === 3 ? 'create_sets_bo5' : 'create_sets_bo3')}</p>
      {#if isAdmin}
        <div class="flex items-center gap-2">
          <button
            onclick={() => (showCloseDialog = true)}
            class="rounded-full bg-[var(--destructive)] px-3 py-1 text-xs font-semibold text-white transition-all active:scale-95"
          >{$_('active_close')}</button>
          <button
            onclick={() => (showCancelDialog = true)}
            class="text-xs text-[var(--text-disabled)] hover:text-[var(--destructive)] transition-colors"
          >{$_('active_cancel')}</button>
        </div>
      {/if}
    </div>
  </nav>

  {#if !match}
    <div class="flex justify-center py-16">
      <div class="h-7 w-7 animate-spin rounded-full border-2 border-[var(--border)] border-t-[var(--primary)]"></div>
    </div>
  {:else}

    <!-- Score card -->
    <div class="rounded-2xl bg-[var(--surface-raised)] px-5 py-5 space-y-1">

      <!-- Card header -->
      <div class="flex items-center gap-2 mb-2">
        <div class="flex-1">
          {#if state?.winner}
            <span class="rounded-full bg-[var(--primary)] px-3 py-0.5 text-[10px] font-bold uppercase text-white">
              {$_('tennis_winner', { values: { team: state.winner === 'a' ? teamANames : teamBNames } })}
            </span>
          {:else}
            <span class="flex items-center gap-1.5 w-fit rounded-full bg-red-100 px-2.5 py-0.5 text-[10px] font-bold uppercase text-red-500">
              <span class="h-1.5 w-1.5 rounded-full bg-red-500 animate-pulse"></span>
              Live
            </span>
          {/if}
        </div>
        <!-- Column headers aligned with score boxes -->
        {#each state?.sets ?? [] as _set, i}
          <div class="w-11 text-center text-[9px] font-bold uppercase tracking-wider text-[var(--text-disabled)] shrink-0">
            {$_('tennis_set')} {i + 1}
          </div>
        {/each}
        {#if !state?.winner}
          <div class="w-11 text-center text-[9px] font-bold uppercase tracking-wider text-[var(--primary)] shrink-0">
            {$_('tennis_set')} {(state?.sets?.length ?? 0) + 1}
          </div>
          <div class="w-14 text-center text-[9px] font-bold uppercase tracking-wider text-[var(--text-disabled)] shrink-0">
            {state?.in_tiebreak ? $_('tennis_tiebreak') : $_('tennis_game')}
          </div>
        {/if}
      </div>

      <!-- Team A row -->
      <div class="flex items-center gap-2 py-2">
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-1.5">
            {#if state?.server === 'a'}
              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" class="text-[var(--primary)] shrink-0">
                <circle cx="12" cy="12" r="10"/>
                <path d="M6 12c0-3.3 1.5-6.2 4-8"/>
                <path d="M18 12c0 3.3-1.5 6.2-4 8"/>
              </svg>
            {/if}
            <p class="text-sm font-[700] leading-snug truncate">{teamANames}</p>
          </div>
          <p class="text-[10px] text-[var(--text-disabled)] uppercase tracking-wide mt-0.5">{$_('tennis_team_a')}</p>
        </div>
        <!-- Completed set score boxes -->
        {#each state?.sets ?? [] as set}
          <div class="h-11 w-11 rounded-lg flex items-center justify-center text-sm font-[700] shrink-0
            {set[0] > set[1] ? 'bg-[var(--surface)] text-[var(--text-primary)]' : 'bg-transparent text-[var(--text-disabled)]'}">
            {set[0]}
          </div>
        {/each}
        <!-- Current set games (live) -->
        {#if !state?.winner}
          <div class="h-11 w-11 rounded-lg flex items-center justify-center text-sm font-[700] shrink-0 bg-[var(--surface)] text-[var(--text-primary)]">
            {state?.games_a ?? 0}
          </div>
          <!-- Current game point score -->
          <div class="h-14 w-14 rounded-xl flex items-center justify-center shrink-0
            {scoreLeadsA ? 'bg-[var(--primary)] text-white' : 'bg-[var(--surface)] text-[var(--text-primary)]'}">
            <p class="text-2xl font-[800] tabular-nums">{gameScoreA}</p>
          </div>
        {/if}
      </div>

      <!-- Divider -->
      <div class="h-px bg-[var(--border)] mx-1"></div>

      <!-- Team B row -->
      <div class="flex items-center gap-2 py-2">
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-1.5">
            {#if state?.server === 'b'}
              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" class="text-[var(--primary)] shrink-0">
                <circle cx="12" cy="12" r="10"/>
                <path d="M6 12c0-3.3 1.5-6.2 4-8"/>
                <path d="M18 12c0 3.3-1.5 6.2-4 8"/>
              </svg>
            {/if}
            <p class="text-sm font-[700] leading-snug truncate">{teamBNames}</p>
          </div>
          <p class="text-[10px] text-[var(--text-disabled)] uppercase tracking-wide mt-0.5">{$_('tennis_team_b')}</p>
        </div>
        <!-- Completed set score boxes -->
        {#each state?.sets ?? [] as set}
          <div class="h-11 w-11 rounded-lg flex items-center justify-center text-sm font-[700] shrink-0
            {set[1] > set[0] ? 'bg-[var(--surface)] text-[var(--text-primary)]' : 'bg-transparent text-[var(--text-disabled)]'}">
            {set[1]}
          </div>
        {/each}
        <!-- Current set games (live) -->
        {#if !state?.winner}
          <div class="h-11 w-11 rounded-lg flex items-center justify-center text-sm font-[700] shrink-0 bg-[var(--surface)] text-[var(--text-primary)]">
            {state?.games_b ?? 0}
          </div>
          <!-- Current game point score -->
          <div class="h-14 w-14 rounded-xl flex items-center justify-center shrink-0
            {scoreLeadsB ? 'bg-[var(--primary)] text-white' : 'bg-[var(--surface)] text-[var(--text-primary)]'}">
            <p class="text-2xl font-[800] tabular-nums">{gameScoreB}</p>
          </div>
        {/if}
      </div>

      <!-- Deuce / tiebreak notice -->
      {#if state?.points_a === 3 && state?.points_b === 3 && !state?.in_tiebreak}
        <p class="text-center text-[11px] font-semibold text-[var(--text-secondary)] pt-1">{$_('tennis_deuce')} — {$_('tennis_golden_point')}</p>
      {:else if state?.in_tiebreak}
        <p class="text-center text-[11px] font-semibold text-[var(--text-secondary)] pt-1">{$_('tennis_tiebreak')}</p>
      {/if}
    </div>

    <!-- Point buttons -->
    {#if !state?.winner}
      <div class="grid grid-cols-2 gap-3">

        <!-- Team A column -->
        <div class="space-y-2">
          <p class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-disabled)] px-1">{$_('tennis_update_a')}</p>
          <button
            onclick={() => addPoint('a')}
            disabled={!!addingPoint}
            class="w-full rounded-2xl bg-[var(--surface-raised)] py-7 flex flex-col items-center gap-0.5 disabled:opacity-60 active:scale-95 transition-transform"
          >
            <span class="text-3xl font-[200] leading-none text-[var(--text-primary)]">+</span>
            <span class="text-[11px] font-[800] uppercase tracking-wide text-[var(--text-primary)]">{$_('active_team_a')}</span>
          </button>
          <button
            onclick={() => setServer('a')}
            class="w-full rounded-xl py-2.5 text-[11px] font-[700] uppercase tracking-wide transition-all active:scale-95
              {state?.server === 'a'
                ? 'bg-[var(--primary)] text-white'
                : 'bg-[var(--surface-raised)] text-[var(--text-secondary)]'}"
          >{$_('active_serve')}</button>
        </div>

        <!-- Team B column -->
        <div class="space-y-2">
          <p class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-disabled)] px-1">{$_('tennis_update_b')}</p>
          <button
            onclick={() => addPoint('b')}
            disabled={!!addingPoint}
            class="w-full rounded-2xl bg-[var(--surface-raised)] py-7 flex flex-col items-center gap-0.5 disabled:opacity-60 active:scale-95 transition-transform"
          >
            <span class="text-3xl font-[200] leading-none text-[var(--text-primary)]">+</span>
            <span class="text-[11px] font-[800] uppercase tracking-wide text-[var(--text-primary)]">{$_('active_team_b')}</span>
          </button>
          <button
            onclick={() => setServer('b')}
            class="w-full rounded-xl py-2.5 text-[11px] font-[700] uppercase tracking-wide transition-all active:scale-95
              {state?.server === 'b'
                ? 'bg-[var(--primary)] text-white'
                : 'bg-[var(--surface-raised)] text-[var(--text-secondary)]'}"
          >{$_('active_serve')}</button>
        </div>

      </div>
    {/if}

  {/if}

</main>

<ConfirmDialog
  bind:open={showCloseDialog}
  title={$_('close_dialog_title')}
  description={$_('close_dialog_desc')}
  confirmLabel={$_('close_dialog_confirm')}
  cancelLabel={$_('close_dialog_cancel')}
  destructive
  onconfirm={closeMatch}
/>

<ConfirmDialog
  bind:open={showCancelDialog}
  title={$_('cancel_dialog_title')}
  description={$_('cancel_dialog_desc')}
  confirmLabel={$_('cancel_dialog_confirm')}
  cancelLabel={$_('cancel_dialog_cancel')}
  destructive
  onconfirm={cancelMatch}
/>
