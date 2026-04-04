<script lang="ts">
  import { untrack } from 'svelte';
  import { api } from '$lib/api/client';
  import { _ } from 'svelte-i18n';
  import RoundIndicator from './RoundIndicator.svelte';
  import Leaderboard from './Leaderboard.svelte';

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

  type Tab = 'round' | 'leaderboard';
  let tab = $state<Tab>(untrack(() => currentRound.matches.every((m) => m.score !== null)) ? 'leaderboard' : 'round');

  const playerName = $derived(
    Object.fromEntries(session.players.map((p) => [p.id, p.name]))
  );

  let scores = $state<Record<string, { a: number; b: number }>>({});
  let submitting = $state<Record<string, boolean>>({});
  let submitError = $state<Record<string, string>>({});

  $effect(() => {
    for (const m of currentRound.matches) {
      if (m.score && !(m.id in scores)) {
        scores[m.id] = { a: m.score.a, b: m.score.b };
      } else if (!(m.id in scores)) {
        scores[m.id] = { a: 0, b: 0 };
      }
    }
  });

  const allScored = $derived(currentRound.matches.every((m) => m.score !== null));

  function adjust(matchId: string, team: 'a' | 'b', delta: number) {
    const s = scores[matchId] ?? { a: 0, b: 0 };
    scores[matchId] = { ...s, [team]: Math.max(0, Math.min(session.points, s[team] + delta)) };
  }

  async function submitScore(matchId: string) {
    submitError[matchId] = '';
    submitting[matchId] = true;
    const s = scores[matchId];
    try {
      await api.scores.submit(session.id, matchId, s.a, s.b, '');
      onRefresh();
    } catch (e) {
      submitError[matchId] = e instanceof Error ? e.message : 'Failed to submit';
    } finally {
      submitting[matchId] = false;
    }
  }

  const benchNames = $derived(currentRound.bench.map((id) => playerName[id] ?? id));
</script>

<!-- Bottom nav -->
<div class="fixed bottom-0 left-0 right-0 z-10 flex border-t border-[var(--border)] bg-[var(--background)]/90 backdrop-blur-sm">
  {#each (['round', 'leaderboard'] as const) as id}
    <button
      onclick={() => (tab = id)}
      class="flex-1 py-3 text-xs font-semibold uppercase tracking-wide transition-colors
        {tab === id ? 'text-[var(--primary)]' : 'text-[var(--text-secondary)]'}"
    >
      {id === 'round' ? $_('active_tab_live') : $_('active_tab_standings')}
    </button>
  {/each}
</div>

{#if tab === 'round'}
  <main class="mx-auto max-w-[480px] px-4 pb-24 pt-6 space-y-4">

    <!-- Header -->
    <div class="flex items-start justify-between">
      <div>
        <h2 class="text-[32px] font-[800] leading-none tracking-tight">
          Round {currentRound.number} of {session.rounds_total}
        </h2>
        <p class="mt-1 text-sm text-[var(--text-secondary)]">
          Americano · {session.courts} {session.courts === 1 ? 'court' : 'courts'}
        </p>
      </div>
      <div class="rounded-xl bg-[var(--surface-raised)] px-4 py-2 text-center">
        <p class="text-[10px] font-bold uppercase tracking-wider text-[var(--text-secondary)]">{$_('active_target_label')}</p>
        <p class="text-xl font-[800] text-[var(--text-primary)]">{session.points}</p>
      </div>
    </div>

    <RoundIndicator current={currentRound.number} total={session.rounds_total ?? 0} />

    <!-- Court cards -->
    {#each currentRound.matches as match (match.id)}
      {@const s = scores[match.id] ?? { a: 0, b: 0 }}
      {@const scored = match.score !== null}

      <div class="space-y-2">
        <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">
          {$_('active_court', { values: { n: match.court } })}
        </p>

        {#if scored}
          <!-- Scored: compact summary -->
          <div class="rounded-2xl bg-[var(--surface-raised)] px-5 py-4">
            <div class="flex items-center justify-between">
              <div class="space-y-0.5">
                <p class="text-sm font-semibold">{playerName[match.team_a[0]]} · {playerName[match.team_a[1]]}</p>
                <p class="text-xs text-[var(--text-secondary)]">{playerName[match.team_b[0]]} · {playerName[match.team_b[1]]}</p>
              </div>
              <div class="flex items-center gap-3">
                <span class="text-3xl font-[800] tabular-nums">{match.score!.a}</span>
                <span class="text-sm text-[var(--text-disabled)]">–</span>
                <span class="text-3xl font-[800] tabular-nums">{match.score!.b}</span>
              </div>
            </div>
            <button
              onclick={() => { scores[match.id] = { a: match.score!.a, b: match.score!.b }; }}
              class="mt-3 w-full text-xs text-[var(--text-disabled)] underline-offset-2 hover:underline"
            >
              {$_('active_edit_score')}
            </button>
          </div>

        {:else}
          <!-- Score entry: two team cards -->
          {#each (['a', 'b'] as const) as team}
            {@const teamPlayers = team === 'a'
              ? [playerName[match.team_a[0]], playerName[match.team_a[1]]]
              : [playerName[match.team_b[0]], playerName[match.team_b[1]]]}
            <div class="rounded-2xl bg-[var(--surface-raised)] px-5 py-4">
              <!-- Team label + players -->
              <div class="mb-4 space-y-1">
                <p class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-secondary)]">
                  {team === 'a' ? $_('active_team_a') : $_('active_team_b')}
                </p>
                {#each teamPlayers as pname, i}
                  <p class="font-[700] text-[var(--text-primary)] {i > 0 ? 'opacity-75' : ''}">{pname}</p>
                {/each}
              </div>

              <!-- Score tapper -->
              <div class="flex items-center justify-between gap-4">
                <button
                  onclick={() => adjust(match.id, team, -1)}
                  disabled={s[team] === 0}
                  class="flex h-16 w-16 items-center justify-center rounded-full bg-[var(--surface)] text-2xl font-bold text-[var(--text-secondary)] shadow-sm transition-all active:scale-95 disabled:opacity-30"
                >
                  −
                </button>
                <span class="text-[72px] font-[800] leading-none tabular-nums text-[var(--text-primary)]">
                  {s[team]}
                </span>
                <button
                  onclick={() => adjust(match.id, team, 1)}
                  disabled={s.a + s.b >= session.points}
                  class="flex h-16 w-16 items-center justify-center rounded-full bg-[var(--primary-muted)] text-2xl font-bold text-[var(--primary)] shadow-sm transition-all active:scale-95 disabled:opacity-30"
                >
                  +
                </button>
              </div>
            </div>
          {/each}

          <!-- Remaining + confirm -->
          <div class="flex items-center justify-between px-1">
            <p class="text-sm text-[var(--text-disabled)]">
              {#if s.a + s.b === session.points}
                {$_('active_points_done')}
              {:else if s.a + s.b < session.points}
                {$_('active_points_left', { values: { n: session.points - s.a - s.b } })}
              {:else}
                {$_('active_points_over', { values: { n: s.a + s.b - session.points } })}
              {/if}
            </p>
            {#if submitError[match.id]}
              <p class="text-xs text-[var(--destructive)]">{submitError[match.id]}</p>
            {/if}
          </div>

          <button
            onclick={() => submitScore(match.id)}
            disabled={s.a + s.b !== session.points || submitting[match.id]}
            class="w-full rounded-2xl bg-[var(--primary)] px-4 py-4 text-[15px] font-[700] text-white transition-all active:scale-[0.98] disabled:opacity-40"
          >
            {submitting[match.id] ? '…' : $_('active_finalise')}
          </button>
        {/if}
      </div>
    {/each}

    <!-- Bench -->
    {#if benchNames.length > 0}
      <div class="flex items-center gap-3 rounded-2xl bg-[var(--surface-raised)] px-4 py-3">
        <div class="flex h-8 w-8 items-center justify-center rounded-full bg-[var(--border)] text-[11px] font-bold text-[var(--text-secondary)]">
          {benchNames[0][0].toUpperCase()}
        </div>
        <p class="text-sm text-[var(--text-secondary)]">
          {$_('active_bench')}: <span class="font-semibold text-[var(--text-primary)]">{benchNames.join(', ')}</span>
        </p>
      </div>
    {/if}

    <!-- Next round / waiting -->
    {#if allScored && isAdmin}
      <button
        onclick={onRefresh}
        class="w-full rounded-2xl bg-[var(--primary)] px-4 py-4 text-[15px] font-[700] text-white transition-all active:scale-[0.98]"
      >
        {currentRound.number === session.rounds_total ? $_('active_final_results') : $_('active_next_round')}
      </button>
    {:else if allScored}
      <div class="rounded-2xl bg-[var(--surface-raised)] px-4 py-3 text-center text-sm text-[var(--text-secondary)]">
        {$_('active_round_complete')}
      </div>
    {/if}

  </main>

{:else}
  <div class="pb-16">
    <Leaderboard sessionId={session.id} />
  </div>
{/if}
