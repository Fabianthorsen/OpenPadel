<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { api } from '$lib/api/client';
  import { _ } from 'svelte-i18n';
  import { Trophy } from 'lucide-svelte';
  import { initials } from '$lib/utils';

  let {
    sessionId,
    sessionName = '',
    complete = false,
  }: {
    sessionId: string;
    sessionName?: string;
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

  const podiumOrder = $derived(
    leaderboard ? [
      leaderboard.standings[1], // 2nd — left
      leaderboard.standings[0], // 1st — centre
      leaderboard.standings[2], // 3rd — right
    ].filter(Boolean) : []
  );

  const funStats = $derived.by(() => {
    if (!leaderboard || leaderboard.standings.length === 0) return [];
    const s = leaderboard.standings;

    const mostWins = s.reduce((a, b) => (b.wins ?? 0) > (a.wins ?? 0) ? b : a);
    const fewestLosses = s.reduce((a, b) => {
      const la = (a.games_played ?? 0) - (a.wins ?? 0) - (a.draws ?? 0);
      const lb = (b.games_played ?? 0) - (b.wins ?? 0) - (b.draws ?? 0);
      return lb < la ? b : a;
    });
    const mostDraws = s.reduce((a, b) => (b.draws ?? 0) > (a.draws ?? 0) ? b : a);

    const stats = [
      { badge: 'W', label: 'stat_most_wins', name: mostWins.name, value: `${mostWins.wins ?? 0}W` },
      { badge: 'L', label: 'stat_iron_wall', name: fewestLosses.name, value: `${(fewestLosses.games_played ?? 0) - (fewestLosses.wins ?? 0) - (fewestLosses.draws ?? 0)}L` },
    ];
    if ((mostDraws.draws ?? 0) > 0) {
      stats.push({ badge: 'D', label: 'stat_diplomat', name: mostDraws.name, value: `${mostDraws.draws ?? 0}D` });
    }
    return stats;
  });
</script>

<main class="mx-auto max-w-[480px] px-4 pb-24 pt-4 space-y-6">
  {#if !leaderboard}
    <p class="text-sm text-[var(--text-secondary)]">Loading…</p>

  {:else if complete}

    <!-- ── Final Results ── -->

    <!-- Heading -->
    <div class="pt-4 text-center space-y-0.5">
      <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-[var(--text-disabled)]">{$_('leaderboard_final')}</p>
      {#if sessionName}
        <p class="text-xl font-[800]">{sessionName}</p>
      {/if}
    </div>

    <!-- Podium -->
    <div class="flex items-end justify-center gap-3 pt-6 pb-2">
      {#each podiumOrder as s}
        {@const isFirst = s.rank === 1}
        <div class="flex flex-col items-center {isFirst ? 'order-2 -mb-0' : s.rank === 2 ? 'order-1' : 'order-3'} flex-1 max-w-[120px]">

          <!-- Trophy for winner -->
          {#if isFirst}
            <div class="mb-1 text-[var(--primary)]"><Trophy size={28} /></div>
          {/if}

          <!-- Avatar -->
          <div class="flex shrink-0 items-center justify-center rounded-full font-[800] text-white
            {isFirst ? 'h-20 w-20 text-2xl bg-[var(--primary)] ring-4 ring-[var(--primary-muted)] shadow-lg' : 'h-14 w-14 text-base bg-[#4a7856] ring-2 ring-[var(--border)]'}">
            {initials(s.name)}
          </div>

          <!-- Rank badge -->
          <div class="mt-2 flex h-6 w-6 items-center justify-center rounded-full text-xs font-[800] text-white
            {isFirst ? 'bg-[var(--primary)]' : 'bg-[#4a7856]'}">
            {s.rank}
          </div>

          <p class="mt-1.5 text-sm font-[800] text-center truncate w-full {isFirst ? 'text-[var(--text-primary)]' : 'text-[var(--text-secondary)]'}">{s.name}</p>
          <p class="text-[10px] font-bold uppercase tracking-widest {isFirst ? 'text-[var(--primary)]' : 'text-[var(--text-disabled)]'}">{s.points} {$_('leaderboard_pts')}</p>

          <!-- Podium bar -->
          <div class="mt-3 w-full rounded-t-xl
            {isFirst ? 'h-12 bg-[var(--primary)]' : s.rank === 2 ? 'h-8 bg-[#4a7856]/60' : 'h-5 bg-[#a8c5b0]/60'}">
          </div>
        </div>
      {/each}
    </div>

    <!-- Rest of standings (4th+) -->
    {#if leaderboard.standings.length > 3}
      <div class="space-y-1">
        <p class="px-1 text-[11px] font-bold uppercase tracking-[0.1em] text-[var(--text-disabled)]">{$_('leaderboard_ranking')}</p>
        {#each leaderboard.standings.slice(3) as s (s.player_id)}
          <div class="flex items-center gap-3 rounded-2xl bg-[var(--surface-raised)] px-4 py-3">
            <span class="w-6 text-sm font-[800] tabular-nums text-[var(--text-disabled)]">{s.rank}</span>
            <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-[var(--primary-muted)] text-xs font-[800] text-[var(--primary)]">
              {initials(s.name)}
            </div>
            <span class="flex-1 truncate text-sm font-semibold">{s.name}</span>
            <span class="text-base font-[800] tabular-nums">{s.points}</span>
            <span class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-disabled)]">{$_('leaderboard_pts')}</span>
          </div>
        {/each}
      </div>
    {/if}

    <!-- Fun stats -->
    {#if funStats.length > 0}
      <div class="space-y-1">
        <p class="px-1 text-[11px] font-bold uppercase tracking-[0.1em] text-[var(--text-disabled)]">{$_('stat_title')}</p>
        {#each funStats as stat}
          <div class="flex items-center gap-3 rounded-2xl bg-[var(--surface-raised)] px-4 py-3">
            <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-[var(--primary-muted)] text-xs font-[800] text-[var(--primary)]">{stat.badge}</div>
            <div class="flex-1 min-w-0">
              <p class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-disabled)]">{$_(stat.label)}</p>
              <p class="text-sm font-semibold truncate">{stat.name}</p>
            </div>
            <span class="text-sm font-[800] tabular-nums text-[var(--primary)]">{stat.value}</span>
          </div>
        {/each}
      </div>
    {/if}

    <a
      href="/"
      class="block w-full rounded-2xl bg-[var(--primary)] px-4 py-4 text-center text-[15px] font-semibold text-white hover:bg-[var(--primary-hover)]"
    >
      {$_('leaderboard_new_session')}
    </a>

  {:else}

    <!-- ── Live Standings ── -->

    <!-- Leader hero card -->
    {#if leader}
      <div class="relative overflow-hidden rounded-2xl bg-[var(--primary)] px-6 py-6">
        <svg class="absolute inset-0 h-full w-full opacity-10" preserveAspectRatio="none" viewBox="0 0 100 100">
          <line x1="50" y1="0" x2="50" y2="100" stroke="white" stroke-width="0.5"/>
          <line x1="0" y1="50" x2="100" y2="50" stroke="white" stroke-width="0.5"/>
          <rect x="20" y="20" width="60" height="60" fill="none" stroke="white" stroke-width="0.5"/>
        </svg>
        <div class="relative z-10 flex items-center gap-5">
          <div class="flex h-16 w-16 shrink-0 items-center justify-center rounded-full bg-white/20 text-2xl font-[800] text-white">
            {initials(leader.name)}
          </div>
          <div class="flex-1 min-w-0">
            <div class="mb-0.5">
              <span class="rounded-full bg-white/20 px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-widest text-white">
                {$_('leaderboard_leader')}
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
          {$_('leaderboard_current')}
        </h3>
        {#if leaderboard.current_round && leaderboard.total_rounds}
          <span class="text-xs text-[var(--text-disabled)]">
            {$_('leaderboard_round_of', { values: { current: leaderboard.current_round, total: leaderboard.total_rounds } })}
          </span>
        {/if}
      </div>

      <div class="grid grid-cols-[2rem_1fr_3rem_3.5rem_3rem] gap-2 px-4 pb-1">
        <span class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-disabled)]">#</span>
        <span class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-disabled)]">{$_('leaderboard_player')}</span>
        <span class="text-center text-[10px] font-bold uppercase tracking-widest text-[var(--text-disabled)]">{$_('leaderboard_games')}</span>
        <span class="text-center text-[10px] font-bold uppercase tracking-widest text-[var(--text-disabled)]">{$_('leaderboard_wl')}</span>
        <span class="text-right text-[10px] font-bold uppercase tracking-widest text-[var(--text-disabled)]">{$_('leaderboard_pts')}</span>
      </div>

      {#each leaderboard.standings as s, i (s.player_id)}
        {@const podiumBg = s.rank === 1 ? 'bg-[var(--primary)]' : s.rank === 2 ? 'bg-[#4a7856]' : s.rank === 3 ? 'bg-[#a8c5b0]' : i % 2 === 0 ? 'bg-[var(--surface-raised)]' : 'bg-transparent'}
        {@const isPodium = s.rank <= 3}
        <div class="grid grid-cols-[2rem_1fr_3rem_3.5rem_3rem] items-center gap-2 rounded-2xl px-4 py-3.5 {podiumBg}">
          <span class="text-sm font-[800] tabular-nums {isPodium ? 'text-white' : 'text-[var(--text-disabled)]'}">{s.rank}</span>
          <div class="flex items-center gap-2.5 min-w-0">
            <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full {isPodium ? 'bg-white/20 text-white' : 'bg-[var(--primary-muted)] text-[var(--primary)]'} text-xs font-[800]">
              {initials(s.name)}
            </div>
            <span class="truncate text-sm font-semibold {isPodium ? 'text-white' : ''}">{s.name}</span>
          </div>
          <span class="text-center text-sm {isPodium ? 'text-white/70' : 'text-[var(--text-secondary)]'}">{s.games_played ?? 0}</span>
          <span class="text-center text-sm font-semibold {isPodium ? 'text-white/70' : 'text-[var(--text-secondary)]'}">
            {s.wins ?? 0}/{s.draws ?? 0}/{(s.games_played ?? 0) - (s.wins ?? 0) - (s.draws ?? 0)}
          </span>
          <span class="text-right text-base font-[800] tabular-nums {isPodium ? 'text-white' : ''}">{s.points}</span>
        </div>
      {/each}
    </div>

  {/if}
</main>
