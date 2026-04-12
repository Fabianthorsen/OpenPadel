<script lang="ts">
  import { page } from '$app/state';
  import { goto } from '$app/navigation';
  import { onMount, onDestroy } from 'svelte';
  import { api, ApiError } from '$lib/api/client';
  import Lobby from '$lib/components/Lobby.svelte';
  import ActiveSession from '$lib/components/ActiveSession.svelte';
  import TennisMatch from '$lib/components/TennisMatch.svelte';
  import Leaderboard from '$lib/components/Leaderboard.svelte';
  import TennisResult from '$lib/components/TennisResult.svelte';
  import PullToRefresh from '$lib/components/PullToRefresh.svelte';
  import { toast } from 'svelte-sonner';
  import { translateApiError } from '$lib/i18n/errors';

  let session = $state<App.Session | null>(null);
  let currentRound = $state<App.Round | null>(null);
  let interval: ReturnType<typeof setInterval>;

  // Poll faster in lobby so players see each other join without waiting.
  // Active sessions use 2s so live score taps are visible to other viewers quickly.
  const POLL_LOBBY = 3_000;
  const POLL_ACTIVE = 2_000;

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

  async function load() {
    const token = getAdminToken() ?? undefined;
    try {
      session = await api.sessions.get(sessionId, token);
      if (session.status !== 'lobby' && session.game_mode !== 'tennis') {
        currentRound = await api.rounds.current(sessionId).catch(() => null);
      }
    } catch (e) {
      if (e instanceof ApiError && e.status === 404) {
        goto('/?notfound=1');
        return;
      }
      toast.error(e instanceof ApiError ? translateApiError(e.message) : translateApiError('server_error'));
    }
  }

  function scheduleNext() {
    clearInterval(interval);
    const delay = session?.status === 'lobby' ? POLL_LOBBY : POLL_ACTIVE;
    interval = setInterval(() => { load().then(scheduleNext); }, delay);
  }

  onMount(() => { load().then(scheduleNext); });
  onDestroy(() => clearInterval(interval));
</script>

{#if !session}
  <main class="flex min-h-svh items-center justify-center px-4">
    <p class="text-sm text-[var(--text-secondary)]">Loading…</p>
  </main>
{:else if session.status === 'lobby'}
  <PullToRefresh onRefresh={load}>
    <Lobby {session} {isAdmin} onRefresh={load} onStarted={load} />
  </PullToRefresh>
{:else if session.status === 'active' && session.game_mode === 'tennis'}
  <PullToRefresh onRefresh={load}>
    <TennisMatch {session} {isAdmin} onRefresh={load} />
  </PullToRefresh>
{:else if session.status === 'active' && currentRound}
  <PullToRefresh onRefresh={load}>
    <ActiveSession {session} {currentRound} {isAdmin} onRefresh={load} />
  </PullToRefresh>
{:else if session.status === 'complete' && session.game_mode === 'tennis'}
  <TennisResult {session} />
{:else if session.status === 'complete'}
  <Leaderboard sessionId={session.id} sessionName={session.name} complete />
{:else}
  <main class="flex min-h-svh items-center justify-center px-4">
    <p class="text-sm text-[var(--text-secondary)]">Loading…</p>
  </main>
{/if}
