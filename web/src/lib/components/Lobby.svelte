<script lang="ts">
  import { api } from '$lib/api/client';
  import { Crown } from 'lucide-svelte';

  let {
    session,
    isAdmin,
    onRefresh,
    onStarted,
  }: {
    session: App.Session;
    isAdmin: boolean;
    onRefresh: () => void;
    onStarted: () => void;
  } = $props();

  const isDev = import.meta.env.DEV;
  const devNames = ['Alice', 'Bob', 'Carlos', 'Diana', 'Erik', 'Fiona', 'Gio', 'Hanna', 'Ivan', 'Julia', 'Karl', 'Lena'];

  let copied = $state(false);
  let starting = $state(false);
  let cancelling = $state(false);
  let seeding = $state(false);
  let joinName = $state('');
  let joining = $state(false);
  let joinError = $state('');

  const joinUrl = $derived(
    typeof location !== 'undefined' ? `${location.origin}/s/${session.id}` : ''
  );

  const activePlayers = $derived(session.players.filter((p) => p.active));
  const canStart = $derived(activePlayers.length >= session.courts * 4);
  const creatorName = $derived(activePlayers.find((p) => p.id === session.creator_player_id)?.name ?? '');

  const myPlayerId = $derived(
    typeof localStorage !== 'undefined' ? localStorage.getItem(`player_id_${session.id}`) : null
  );
  const alreadyJoined = $derived(!!myPlayerId && activePlayers.some((p) => p.id === myPlayerId));

  async function copyLink() {
    await navigator.clipboard.writeText(joinUrl);
    copied = true;
    setTimeout(() => (copied = false), 2000);
  }

  async function join() {
    joinError = '';
    const name = joinName.trim();
    if (!name) return;
    joining = true;
    try {
      const player = await api.players.join(session.id, name);
      localStorage.setItem(`player_id_${session.id}`, player.id);
      joinName = '';
      onRefresh();
    } catch (e) {
      joinError = e instanceof Error ? e.message : 'Could not join';
    } finally {
      joining = false;
    }
  }

  async function seedPlayers() {
    seeding = true;
    const existing = new Set(activePlayers.map((p) => p.name));
    const needed = session.courts * 4 + 2;
    const toAdd = devNames.filter((n) => !existing.has(n)).slice(0, Math.max(0, needed - activePlayers.length));
    for (const name of toAdd) {
      await api.players.join(session.id, name).catch(() => {});
    }
    seeding = false;
    onRefresh();
  }

  async function cancel() {
    if (!confirm('Cancel this session? This cannot be undone.')) return;
    cancelling = true;
    try {
      const adminToken = localStorage.getItem(`admin_token_${session.id}`) ?? '';
      await api.sessions.cancel(session.id, adminToken);
      location.href = '/';
    } catch {
      cancelling = false;
    }
  }

  async function start() {
    starting = true;
    try {
      const adminToken = localStorage.getItem(`admin_token_${session.id}`) ?? '';
      await api.sessions.start(session.id, adminToken);
      onStarted();
    } catch {
      starting = false;
    }
  }
</script>

<!-- ── Join / invite screen (visitor hasn't joined yet) ── -->
{#if !isAdmin && !alreadyJoined}
  <main class="flex min-h-svh flex-col px-6 py-12">
    <div class="flex flex-1 flex-col">
      <!-- Brand -->
      <p class="text-center text-sm font-semibold text-[var(--primary)]">NotTennis</p>

      <!-- Headline -->
      <div class="mt-10 space-y-3">
        <h1 class="text-[32px] font-[800]">
          {#if creatorName}
            {creatorName} invited you to this Americano Tournament
          {:else}
            Join this Americano Tournament
          {/if}
        </h1>
        <p class="text-[var(--text-secondary)]">
          Join the session to track your scores and see your ranking.
        </p>
      </div>

      <!-- Join form -->
      <div class="mt-10 space-y-5">
        <div class="space-y-2.5">
          <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">Your name</p>
          <form onsubmit={(e) => { e.preventDefault(); join(); }}>
            <input
              bind:value={joinName}
              placeholder="Enter your name"
              maxlength="32"
              class="w-full rounded-2xl bg-[var(--surface-raised)] px-4 py-3.5 text-sm outline-none focus:ring-2 focus:ring-[var(--primary)]/20"
            />
          </form>
        </div>

        {#if joinError}
          <p class="text-sm text-[var(--destructive)]">{joinError}</p>
        {/if}

        <button
          onclick={join}
          disabled={joining || !joinName.trim()}
          class="w-full rounded-2xl bg-[var(--primary)] px-4 py-4 text-[15px] font-semibold text-white transition-colors hover:bg-[var(--primary-hover)] disabled:opacity-50"
        >
          {joining ? 'Joining…' : 'Join Tournament →'}
        </button>
      </div>
    </div>

  </main>

<!-- ── Lobby (admin or already joined) ── -->
{:else}
  <main class="mx-auto max-w-[480px] px-6 py-6 space-y-6">
    <!-- Nav -->
    <nav class="flex items-center justify-between">
      <div class="space-y-0.5">
        <p class="text-xs text-[var(--text-secondary)]">Waiting to start</p>
        <p class="text-sm font-semibold text-[var(--primary)]">NotTennis</p>
      </div>
      <div class="text-right text-xs text-[var(--text-secondary)]">
        {session.courts} {session.courts === 1 ? 'court' : 'courts'} · {session.points} pts · Americano
      </div>
    </nav>

    <!-- Join code + share -->
    <div class="rounded-2xl bg-[var(--surface-raised)] px-5 py-4 space-y-3">
      <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">Join code</p>
      <div class="flex gap-2">
        {#each session.id.split('') as char}
          <div class="flex flex-1 items-center justify-center rounded-xl bg-[var(--surface)] py-3 text-2xl font-[700] text-[var(--text-primary)] font-mono">
            {char}
          </div>
        {/each}
      </div>
      <div class="flex items-center gap-2 rounded-xl bg-[var(--surface)] px-3 py-2.5">
        <span class="flex-1 truncate text-xs text-[var(--text-secondary)]">{joinUrl}</span>
        <button
          onclick={copyLink}
          class="shrink-0 text-xs font-semibold text-[var(--primary)] transition-colors hover:text-[var(--primary-hover)]"
        >
          {copied ? 'Copied!' : 'Copy'}
        </button>
      </div>
    </div>

    <!-- Player list -->
    <div class="space-y-2">
      <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">
        Players ({activePlayers.length})
      </p>
      {#if activePlayers.length === 0}
        <p class="text-sm text-[var(--text-disabled)]">Waiting for players to join…</p>
      {:else}
        <div class="rounded-2xl bg-[var(--surface-raised)] divide-y divide-[var(--border)]">
          {#each activePlayers as player (player.id)}
            <div class="flex items-center gap-3 px-4 py-3">
              <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-[var(--primary-muted)] text-sm font-semibold text-[var(--primary)]">
                {player.name[0].toUpperCase()}
              </div>
              <span class="text-sm font-medium">{player.name}</span>
              <div class="ml-auto flex items-center gap-1.5">
                {#if player.id === session.creator_player_id}
                  <Crown size={13} class="text-[var(--primary)]" />
                {/if}
                {#if player.id === myPlayerId}
                  <span class="text-xs text-[var(--text-disabled)]">you</span>
                {/if}
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>

    <!-- Join form (non-admin who hasn't joined) -->
    {#if !isAdmin && !alreadyJoined}
      <div class="space-y-2">
        <form onsubmit={(e) => { e.preventDefault(); join(); }} class="flex gap-2">
          <input
            bind:value={joinName}
            placeholder="Your name"
            maxlength="32"
            class="flex-1 rounded-2xl bg-[var(--surface-raised)] px-4 py-3 text-sm outline-none focus:ring-2 focus:ring-[var(--primary)]/20"
          />
          <button
            type="submit"
            disabled={joining || !joinName.trim()}
            class="rounded-2xl bg-[var(--primary)] px-4 text-sm font-semibold text-white disabled:opacity-50"
          >
            {joining ? '…' : 'Join'}
          </button>
        </form>
        {#if joinError}
          <p class="text-sm text-[var(--destructive)]">{joinError}</p>
        {/if}
      </div>
    {/if}

    <!-- Admin controls -->
    {#if isAdmin}
      <div class="space-y-2">
        <button
          onclick={start}
          disabled={starting || !canStart}
          class="w-full rounded-2xl bg-[var(--primary)] px-4 py-4 text-[15px] font-semibold text-white transition-colors hover:bg-[var(--primary-hover)] disabled:opacity-40"
        >
          {starting ? 'Starting…' : 'Start session →'}
        </button>
        {#if !canStart}
          <p class="text-center text-xs text-[var(--text-disabled)]">
            Need at least {session.courts * 4} players to start
          </p>
        {/if}
        <button
          onclick={cancel}
          disabled={cancelling}
          class="w-full text-sm text-[var(--text-secondary)] underline-offset-2 hover:text-[var(--destructive)] hover:underline disabled:opacity-50"
        >
          {cancelling ? 'Cancelling…' : 'Cancel tournament'}
        </button>
        {#if isDev}
          <button
            onclick={seedPlayers}
            disabled={seeding}
            class="w-full text-sm text-[var(--text-disabled)] underline-offset-2 hover:underline disabled:opacity-50"
          >
            {seeding ? 'Adding…' : '[dev] Fill with test players'}
          </button>
        {/if}
      </div>
    {:else}
      <div class="rounded-2xl bg-[var(--surface-raised)] px-4 py-3 text-center">
        <p class="text-sm text-[var(--text-secondary)]">Waiting for the admin to start…</p>
      </div>
    {/if}
  </main>
{/if}
