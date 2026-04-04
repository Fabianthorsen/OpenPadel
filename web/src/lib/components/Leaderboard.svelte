<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { api } from '$lib/api/client';
  import { _ } from 'svelte-i18n';

  let {
    sessionId,
    complete = false,
  }: {
    sessionId: string;
    complete?: boolean;
  } = $props();

  let leaderboard = $state<App.Leaderboard | null>(null);
  let interval: ReturnType<typeof setInterval>;

  async function load() {
    try {
      leaderboard = await api.leaderboard.get(sessionId);
    } catch {
      // silently retry on next interval
    }
  }

  onMount(() => {
    load();
    if (!complete) interval = setInterval(load, 15_000);
  });

  onDestroy(() => clearInterval(interval));

  const leader = $derived(leaderboard?.standings[0] ?? null);
</script>

<main class="mx-auto max-w-[480px] px-4 pb-6 pt-4 space-y-6">
  {#if !leaderboard}
    <p class="text-sm text-[var(--text-secondary)]">Loading…</p>

  {:else}

    <!-- Leader hero card -->
    {#if leader}
      <div class="relative overflow-hidden rounded-2xl bg-[var(--primary)] px-6 py-6">
        <!-- subtle court lines -->
        <svg class="absolute inset-0 h-full w-full opacity-10" preserveAspectRatio="none" viewBox="0 0 100 100">
          <line x1="50" y1="0" x2="50" y2="100" stroke="white" stroke-width="0.5"/>
          <line x1="0" y1="50" x2="100" y2="50" stroke="white" stroke-width="0.5"/>
          <rect x="20" y="20" width="60" height="60" fill="none" stroke="white" stroke-width="0.5"/>
        </svg>

        <div class="relative z-10 flex items-center gap-5">
          <!-- Avatar -->
          <div class="flex h-16 w-16 shrink-0 items-center justify-center rounded-full bg-white/20 text-2xl font-[800] text-white">
            {leader.name[0].toUpperCase()}
          </div>

          <!-- Info -->
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2 mb-0.5">
              <span class="rounded-full bg-white/20 px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-widest text-white">
                {$_('leaderboard_leader')}
              </span>
              <span class="text-[11px] font-bold uppercase tracking-widest text-white/60">
                {$_('leaderboard_rank1')}
              </span>
            </div>
            <p class="text-2xl font-[800] text-white truncate">{leader.name}</p>
            <div class="mt-2 flex items-center gap-4">
              <div>
                <span class="text-xl font-[800] text-white">{leader.points}</span>
                <span class="ml-1 text-[10px] font-bold uppercase tracking-wider text-white/60">{$_('leaderboard_pts')}</span>
              </div>
              {#if (leader.games_played ?? 0) > 0}
                <div class="h-6 w-px bg-white/20"></div>
                <div>
                  <span class="text-xl font-[800] text-white">{leader.wins ?? 0}/{leader.draws ?? 0}/{(leader.games_played ?? 0) - (leader.wins ?? 0) - (leader.draws ?? 0)}</span>
                  <span class="ml-1 text-[10px] font-bold uppercase tracking-wider text-white/60">{$_('leaderboard_wl')}</span>
                </div>
              {/if}
            </div>
          </div>
        </div>
      </div>
    {/if}

    <!-- Standings -->
    <div class="space-y-1">
      <div class="flex items-center justify-between px-1 pb-1">
        <h3 class="text-[13px] font-bold uppercase tracking-[0.1em] text-[var(--text-secondary)]">
          {complete ? $_('leaderboard_final') : $_('leaderboard_current')}
        </h3>
        {#if leaderboard.current_round && leaderboard.total_rounds}
          <span class="text-xs text-[var(--text-disabled)]">
            {$_('leaderboard_round_of', { values: { current: leaderboard.current_round, total: leaderboard.total_rounds } })}
          </span>
        {/if}
      </div>

      <!-- Header -->
      <div class="grid grid-cols-[2rem_1fr_3rem_3.5rem_3rem] gap-2 px-4 pb-1">
        <span class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-disabled)]">#</span>
        <span class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-disabled)]">{$_('leaderboard_player')}</span>
        <span class="text-center text-[10px] font-bold uppercase tracking-widest text-[var(--text-disabled)]">{$_('leaderboard_games')}</span>
        <span class="text-center text-[10px] font-bold uppercase tracking-widest text-[var(--text-disabled)]">{$_('leaderboard_wl')}</span>
        <span class="text-right text-[10px] font-bold uppercase tracking-widest text-[var(--text-disabled)]">{$_('leaderboard_pts')}</span>
      </div>

      <!-- Rows -->
      {#each leaderboard.standings as s, i (s.player_id)}
        <div class="grid grid-cols-[2rem_1fr_3rem_3.5rem_3rem] items-center gap-2 rounded-2xl px-4 py-3.5
          {i % 2 === 0 ? 'bg-[var(--surface-raised)]' : 'bg-transparent'}">
          <span class="text-sm font-[800] tabular-nums {s.rank === 1 ? 'text-[var(--primary)]' : 'text-[var(--text-disabled)]'}">
            {s.rank}
          </span>
          <div class="flex items-center gap-2.5 min-w-0">
            <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-[var(--primary-muted)] text-xs font-[800] text-[var(--primary)]">
              {s.name[0].toUpperCase()}
            </div>
            <span class="truncate text-sm font-semibold">{s.name}</span>
          </div>
          <span class="text-center text-sm text-[var(--text-secondary)]">{s.games_played ?? 0}</span>
          <span class="text-center text-sm font-semibold {s.rank === 1 ? 'text-[var(--primary)]' : 'text-[var(--text-secondary)]'}">
            {s.wins ?? 0}/{s.draws ?? 0}/{(s.games_played ?? 0) - (s.wins ?? 0) - (s.draws ?? 0)}
          </span>
          <span class="text-right text-base font-[800] tabular-nums">{s.points}</span>
        </div>
      {/each}
    </div>

    {#if complete}
      <a
        href="/"
        class="block w-full rounded-2xl bg-[var(--primary)] px-4 py-4 text-center text-[15px] font-semibold text-white hover:bg-[var(--primary-hover)]"
      >
        {$_('leaderboard_new_session')}
      </a>
    {/if}

  {/if}
</main>
