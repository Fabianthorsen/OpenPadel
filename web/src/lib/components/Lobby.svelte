<script lang="ts">
  import { api } from '$lib/api/client';

  let {
    session,
    isAdmin,
    onStarted,
  }: {
    session: App.Session;
    isAdmin: boolean;
    onStarted: () => void;
  } = $props();

  let copied = $state(false);
  let starting = $state(false);
  let joinName = $state('');
  let joining = $state(false);
  let joinError = $state('');

  const joinUrl = $derived(
    typeof location !== 'undefined'
      ? `${location.origin}/s/${session.id}`
      : ''
  );

  const activePlayers = $derived(session.players.filter((p) => p.active));
  const canStart = $derived(activePlayers.length >= session.courts * 4 + 1);

  async function copyLink() {
    await navigator.clipboard.writeText(joinUrl);
    copied = true;
    setTimeout(() => (copied = false), 2000);
  }

  async function share() {
    if (navigator.share) {
      await navigator.share({ title: 'Join my padel session', url: joinUrl });
    } else {
      copyLink();
    }
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
    } catch (e) {
      joinError = e instanceof Error ? e.message : 'Could not join';
    } finally {
      joining = false;
    }
  }

  async function start() {
    starting = true;
    try {
      const adminToken = localStorage.getItem(`admin_token_${session.id}`) ?? '';
      await api.sessions.start(session.id, adminToken);
      onStarted();
    } catch (e) {
      starting = false;
    }
  }

  const myPlayerId = $derived(
    typeof localStorage !== 'undefined'
      ? localStorage.getItem(`player_id_${session.id}`)
      : null
  );
  const alreadyJoined = $derived(
    !!myPlayerId && activePlayers.some((p) => p.id === myPlayerId)
  );
</script>

<main class="mx-auto max-w-[480px] px-4 py-8 space-y-6">
  <!-- Header -->
  <div>
    <p class="text-sm text-[var(--text-secondary)]">Waiting to start</p>
    <h1 class="text-[26px] font-[650]">NotTennis</h1>
    <p class="text-sm text-[var(--text-secondary)]">
      {session.courts} {session.courts === 1 ? 'court' : 'courts'} · {session.points} pts
    </p>
  </div>

  <!-- Invite link (admin) -->
  {#if isAdmin}
    <div class="space-y-2">
      <p class="text-sm font-medium text-[var(--text-secondary)]">Invite players</p>
      <div class="flex items-center gap-2 rounded-lg border border-[var(--border)] bg-[var(--surface)] px-3 py-2.5">
        <span class="flex-1 truncate text-sm text-[var(--text-secondary)]">{joinUrl}</span>
        <button
          onclick={copyLink}
          class="shrink-0 text-sm font-medium text-[var(--primary)] transition-colors hover:text-[var(--primary-hover)]"
        >
          {copied ? 'Copied!' : 'Copy'}
        </button>
      </div>
      <button
        onclick={share}
        class="w-full rounded-lg border border-[var(--border)] bg-[var(--surface)] px-4 py-2.5 text-sm font-medium text-[var(--text-primary)] transition-colors hover:bg-[var(--surface-raised)]"
      >
        Share link
      </button>
    </div>
  {/if}

  <!-- Join form (non-admin who hasn't joined yet) -->
  {#if !isAdmin && !alreadyJoined}
    <div class="space-y-2">
      <p class="text-sm font-medium text-[var(--text-secondary)]">Join this session</p>
      <form onsubmit={(e) => { e.preventDefault(); join(); }} class="flex gap-2">
        <input
          bind:value={joinName}
          placeholder="Your name"
          maxlength="32"
          class="flex-1 rounded-lg border border-[var(--border)] bg-[var(--surface)] px-3 py-2.5 text-sm outline-none focus:border-[var(--border-strong)] focus:ring-2 focus:ring-[var(--primary)]/20"
        />
        <button
          type="submit"
          disabled={joining || !joinName.trim()}
          class="rounded-lg bg-[var(--primary)] px-4 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-[var(--primary-hover)] disabled:opacity-50"
        >
          {joining ? '…' : 'Join'}
        </button>
      </form>
      {#if joinError}
        <p class="text-sm text-[var(--destructive)]">{joinError}</p>
      {/if}
    </div>
  {/if}

  <!-- Player list -->
  <div class="space-y-2">
    <p class="text-sm font-medium text-[var(--text-secondary)]">
      Players ({activePlayers.length})
    </p>
    {#if activePlayers.length === 0}
      <p class="text-sm text-[var(--text-disabled)]">Waiting for players to join…</p>
    {:else}
      <div class="rounded-lg border border-[var(--border)] bg-[var(--surface)] divide-y divide-[var(--border)]">
        {#each activePlayers as player (player.id)}
          <div class="flex items-center gap-3 px-4 py-3">
            <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-[var(--primary-muted)] text-sm font-semibold text-[var(--primary)]">
              {player.name[0].toUpperCase()}
            </div>
            <span class="text-sm font-medium">{player.name}</span>
            {#if player.id === myPlayerId}
              <span class="ml-auto text-xs text-[var(--text-disabled)]">you</span>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>

  <!-- Start button (admin only) -->
  {#if isAdmin}
    <div class="space-y-1.5">
      <button
        onclick={start}
        disabled={starting || !canStart}
        class="w-full rounded-lg bg-[var(--primary)] px-4 py-3 text-[15px] font-semibold text-white transition-colors hover:bg-[var(--primary-hover)] disabled:opacity-40"
      >
        {starting ? 'Starting…' : 'Start session →'}
      </button>
      {#if !canStart}
        <p class="text-center text-xs text-[var(--text-disabled)]">
          Need at least {session.courts * 4 + 1} players to start
        </p>
      {/if}
    </div>
  {:else if alreadyJoined}
    <div class="rounded-lg bg-[var(--surface-raised)] px-4 py-3 text-center">
      <p class="text-sm text-[var(--text-secondary)]">Waiting for the admin to start…</p>
    </div>
  {/if}
</main>
