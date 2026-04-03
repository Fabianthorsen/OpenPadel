<script lang="ts">
  import { untrack } from 'svelte';
  import { api } from '$lib/api/client';
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

  // Map player id → name
  const playerName = $derived(
    Object.fromEntries(session.players.map((p) => [p.id, p.name]))
  );

  // Per-match score state: matchId → { a, b }
  let scores = $state<Record<string, { a: number; b: number }>>({});
  let submitting = $state<Record<string, boolean>>({});
  let submitError = $state<Record<string, string>>({});

  // Initialise scores from already-submitted data
  $effect(() => {
    for (const m of currentRound.matches) {
      if (m.score && !(m.id in scores)) {
        scores[m.id] = { a: m.score.a, b: m.score.b };
      } else if (!(m.id in scores)) {
        scores[m.id] = { a: 0, b: 0 };
      }
    }
  });

  const allScored = $derived(
    currentRound.matches.every((m) => m.score !== null)
  );

  function adjust(matchId: string, team: 'a' | 'b', delta: number) {
    const s = scores[matchId] ?? { a: 0, b: 0 };
    scores[matchId] = { ...s, [team]: Math.max(0, Math.min(session.points, s[team] + delta)) };
  }

  function setScore(matchId: string, team: 'a' | 'b', raw: string) {
    const s = scores[matchId] ?? { a: 0, b: 0 };
    scores[matchId] = { ...s, [team]: Math.max(0, parseInt(raw) || 0) };
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

  const benchNames = $derived(
    currentRound.bench.map((id) => playerName[id] ?? id)
  );
</script>

<!-- Tab bar -->
<div class="sticky top-0 z-10 flex border-b border-[var(--border)] bg-[var(--background)]">
  {#each ([['round', 'Round'], ['leaderboard', 'Leaderboard']] as const) as [id, label]}
    <button
      onclick={() => (tab = id)}
      class="flex-1 py-3 text-sm font-medium transition-colors
        {tab === id
          ? 'border-b-2 border-[var(--primary)] text-[var(--primary)]'
          : 'text-[var(--text-secondary)]'}"
    >
      {label}
    </button>
  {/each}
</div>

{#if tab === 'round'}
  <main class="mx-auto max-w-[480px] space-y-5 px-4 py-6">
    <!-- Round header -->
    <div class="space-y-3">
      <div class="flex items-center justify-between">
        <p class="text-sm font-medium text-[var(--text-secondary)]">
          Round {currentRound.number} of {session.rounds_total}
        </p>
        <span class="rounded-full bg-[var(--primary-muted)] px-2.5 py-0.5 text-xs font-semibold text-[var(--primary)]">
          LIVE
        </span>
      </div>
      <RoundIndicator current={currentRound.number} total={session.rounds_total ?? 0} />
    </div>

    <!-- Court cards -->
    {#each currentRound.matches as match (match.id)}
      {@const s = scores[match.id] ?? { a: 0, b: 0 }}
      {@const scored = match.score !== null}
      <div class="rounded-lg border border-[var(--border)] bg-[var(--surface)]">
        <div class="border-b border-[var(--border)] px-4 py-2.5">
          <p class="text-xs font-semibold uppercase tracking-wide text-[var(--text-secondary)]">
            Court {match.court}
          </p>
        </div>

        <div class="px-4 py-3 space-y-3">
          <!-- Teams -->
          <div class="space-y-1 text-center">
            <p class="text-[15px] font-semibold">
              {playerName[match.team_a[0]]} · {playerName[match.team_a[1]]}
            </p>
            <p class="text-xs text-[var(--text-disabled)]">vs</p>
            <p class="text-[15px] font-semibold">
              {playerName[match.team_b[0]]} · {playerName[match.team_b[1]]}
            </p>
          </div>

          <!-- Score display (already scored) -->
          {#if scored}
            <div class="flex items-center justify-center gap-4 rounded-lg bg-[var(--surface-raised)] py-3">
              <span class="text-2xl font-bold tabular-nums">{match.score!.a}</span>
              <span class="text-sm text-[var(--text-disabled)]">–</span>
              <span class="text-2xl font-bold tabular-nums">{match.score!.b}</span>
            </div>
            <button
              onclick={() => { scores[match.id] = { a: match.score!.a, b: match.score!.b }; }}
              class="w-full text-xs text-[var(--text-disabled)] underline-offset-2 hover:underline"
            >
              Edit score
            </button>

          <!-- Score entry (not yet scored) -->
          {:else}
            <div class="space-y-2">
              {#each (['a', 'b'] as const) as team}
                <div class="space-y-1">
                  <p class="text-xs text-[var(--text-secondary)]">
                    {team === 'a'
                      ? `${playerName[match.team_a[0]]} · ${playerName[match.team_a[1]]}`
                      : `${playerName[match.team_b[0]]} · ${playerName[match.team_b[1]]}`}
                  </p>
                  <div class="flex items-center gap-3">
                    <button
                      onclick={() => adjust(match.id, team, -1)}
                      disabled={s[team] === 0}
                      class="flex h-10 w-10 items-center justify-center rounded-lg border border-[var(--border)] text-lg font-semibold hover:bg-[var(--surface-raised)] disabled:opacity-30"
                    >−</button>
                    <input
                      type="number"
                      inputmode="numeric"
                      min="0"
                      max={session.points}
                      value={s[team]}
                      oninput={(e) => setScore(match.id, team, (e.target as HTMLInputElement).value)}
                      class="h-10 w-full rounded-lg border border-[var(--border)] bg-[var(--surface)] text-center text-xl font-bold tabular-nums outline-none focus:border-[var(--border-strong)]"
                    />
                    <button
                      onclick={() => adjust(match.id, team, 1)}
                      disabled={s.a + s.b >= session.points}
                      class="flex h-10 w-10 items-center justify-center rounded-lg border border-[var(--border)] text-lg font-semibold hover:bg-[var(--surface-raised)] disabled:opacity-30"
                    >+</button>
                  </div>
                </div>
              {/each}

              <div class="flex items-center justify-between pt-1">
                <p class="text-xs text-[var(--text-disabled)]">
                  {#if s.a + s.b === session.points}
                    ✓
                  {:else if s.a + s.b < session.points}
                    {session.points - s.a - s.b} left
                  {:else}
                    {s.a + s.b - session.points} over
                  {/if}
                </p>
                {#if submitError[match.id]}
                  <p class="text-xs text-[var(--destructive)]">{submitError[match.id]}</p>
                {/if}
                <button
                  onclick={() => submitScore(match.id)}
                  disabled={s.a + s.b !== session.points || submitting[match.id]}
                  class="rounded-lg bg-[var(--primary)] px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-[var(--primary-hover)] disabled:opacity-40"
                >
                  {submitting[match.id] ? '…' : 'Confirm'}
                </button>
              </div>
            </div>
          {/if}
        </div>
      </div>
    {/each}

    <!-- Bench -->
    {#if benchNames.length > 0}
      <p class="text-sm text-[var(--text-disabled)]">
        Bench — {benchNames.join(', ')}
      </p>
    {/if}

    <!-- All scored: next round -->
    {#if allScored && isAdmin}
      <button
        onclick={onRefresh}
        class="w-full rounded-lg bg-[var(--primary)] px-4 py-3 text-[15px] font-semibold text-white transition-colors hover:bg-[var(--primary-hover)]"
      >
        {currentRound.number === session.rounds_total ? 'See final results →' : 'Next round →'}
      </button>
    {:else if allScored}
      <div class="rounded-lg bg-[var(--surface-raised)] px-4 py-3 text-center text-sm text-[var(--text-secondary)]">
        Round complete — waiting for admin to advance
      </div>
    {/if}
  </main>

{:else}
  <Leaderboard sessionId={session.id} />
{/if}
