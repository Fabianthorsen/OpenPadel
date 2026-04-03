<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { api } from '$lib/api/client';

  let { sessionId }: { sessionId: string } = $props();

  let leaderboard = $state<App.Leaderboard | null>(null);
  let updatedAt = $state('');
  let interval: ReturnType<typeof setInterval>;

  async function load() {
    try {
      leaderboard = await api.leaderboard.get(sessionId);
      updatedAt = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } catch {
      // silently retry on next interval
    }
  }

  onMount(() => {
    load();
    interval = setInterval(load, 15_000);
  });

  onDestroy(() => clearInterval(interval));
</script>

<main class="mx-auto max-w-[480px] px-4 py-6 space-y-4">
  {#if !leaderboard}
    <p class="text-sm text-[var(--text-secondary)]">Loading…</p>
  {:else}
    <div class="rounded-lg border border-[var(--border)] bg-[var(--surface)] overflow-hidden">
      <!-- Header row -->
      <div class="grid grid-cols-[2rem_1fr_3rem] gap-2 border-b border-[var(--border)] px-4 py-2.5">
        <span class="text-xs font-semibold uppercase tracking-wide text-[var(--text-disabled)]">#</span>
        <span class="text-xs font-semibold uppercase tracking-wide text-[var(--text-disabled)]">Player</span>
        <span class="text-right text-xs font-semibold uppercase tracking-wide text-[var(--text-disabled)]">Pts</span>
      </div>

      <!-- Standings -->
      {#each leaderboard.standings as s (s.player_id)}
        <div class="grid grid-cols-[2rem_1fr_3rem] items-center gap-2 border-b border-[var(--border)] px-4 py-3 last:border-0">
          <span
            class="text-sm font-semibold tabular-nums {s.rank === 1
              ? 'text-[var(--primary)]'
              : 'text-[var(--text-disabled)]'}"
          >
            {s.rank}
          </span>
          <span class="text-sm font-medium truncate">{s.name}</span>
          <span class="text-right text-sm font-semibold tabular-nums">{s.points}</span>
        </div>
      {/each}
    </div>

    {#if updatedAt}
      <p class="text-center text-xs text-[var(--text-disabled)]">Updated {updatedAt}</p>
    {/if}
  {/if}
</main>
