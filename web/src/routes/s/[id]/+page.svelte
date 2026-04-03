<script lang="ts">
  import { page } from '$app/state';
  import { onMount, onDestroy } from 'svelte';
  import { api } from '$lib/api/client';

  let session = $state<App.Session | null>(null);
  let leaderboard = $state<App.Leaderboard | null>(null);
  let currentRound = $state<App.Round | null>(null);
  let error = $state('');
  let interval: ReturnType<typeof setInterval>;

  const sessionId = $derived(page.params.id as string);

  function getAdminToken(): string | null {
    if (typeof localStorage === 'undefined') return null;
    const stored = localStorage.getItem(`admin_token_${sessionId}`);
    if (stored) return stored;
    const param = new URL(location.href).searchParams.get('token');
    if (param) localStorage.setItem(`admin_token_${sessionId}`, param);
    return param;
  }

  const isAdmin = $derived(!!getAdminToken());
  const myPlayerId = $derived(
    typeof localStorage !== 'undefined'
      ? localStorage.getItem(`player_id_${sessionId}`)
      : null
  );

  async function load() {
    const token = getAdminToken() ?? undefined;
    try {
      session = await api.sessions.get(sessionId, token);
      if (session.status !== 'lobby') {
        [leaderboard, currentRound] = await Promise.all([
          api.leaderboard.get(sessionId),
          api.rounds.current(sessionId).catch(() => null),
        ]);
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load session';
    }
  }

  onMount(() => {
    load();
    interval = setInterval(load, 15_000);
  });

  onDestroy(() => clearInterval(interval));
</script>

{#if error}
  <main class="flex min-h-svh items-center justify-center px-4">
    <p class="text-[var(--destructive)]">{error}</p>
  </main>
{:else if !session}
  <main class="flex min-h-svh items-center justify-center px-4">
    <p class="text-sm text-[var(--text-secondary)]">Loading…</p>
  </main>
{:else if session.status === 'lobby'}
  <!-- Lobby — full component coming next -->
  <main class="mx-auto max-w-[480px] px-4 py-8">
    <p class="text-sm text-[var(--text-secondary)]">Lobby — {session.players.length} players joined</p>
  </main>
{:else if session.status === 'active'}
  <!-- Active round — full component coming next -->
  <main class="mx-auto max-w-[480px] px-4 py-8">
    <p class="text-sm text-[var(--text-secondary)]">Round {currentRound?.number} of {session.rounds_total}</p>
  </main>
{:else}
  <!-- Complete — full component coming next -->
  <main class="mx-auto max-w-[480px] px-4 py-8">
    <p class="text-sm text-[var(--text-secondary)]">Session complete</p>
  </main>
{/if}
