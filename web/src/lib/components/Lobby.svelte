<script lang="ts">
  import { api } from '$lib/api/client';
  import { Crown, Share2, Check } from 'lucide-svelte';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
  import { _ } from 'svelte-i18n';

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
  let showCancelDialog = $state(false);
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
    if (navigator.share) {
      await navigator.share({ title: session.name || 'NotTennis', url: joinUrl }).catch(() => {});
    } else {
      await navigator.clipboard.writeText(joinUrl);
      copied = true;
      setTimeout(() => (copied = false), 2000);
    }
  }

  async function join() {
    joinError = '';
    const name = joinName.trim();
    if (!name) return;
    joining = true;
    try {
      const player = await api.players.join(session.id, name);
      // Only claim this player as "you" if you're not already in the session
      if (!isAdmin) {
        localStorage.setItem(`player_id_${session.id}`, player.id);
        localStorage.setItem('last_session_id', session.id);
      }
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
    cancelling = true;
    try {
      const adminToken = localStorage.getItem(`admin_token_${session.id}`) ?? '';
      await api.sessions.cancel(session.id, adminToken);
      location.href = '/';
    } catch {
      cancelling = false;
    }
  }

  async function removePlayer(playerId: string) {
    const adminToken = localStorage.getItem(`admin_token_${session.id}`) ?? '';
    await api.players.remove(session.id, playerId, adminToken).catch(() => {});
    onRefresh();
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

{#if cancelling}
  <main class="flex min-h-svh flex-col items-center justify-center gap-3 px-6">
    <div class="h-8 w-8 animate-spin rounded-full border-2 border-[var(--border)] border-t-[var(--primary)]"></div>
    <p class="text-sm text-[var(--text-secondary)]">{$_('lobby_cancelling')}</p>
  </main>

<!-- ── Join / invite screen (visitor hasn't joined yet) ── -->
{:else if !isAdmin && !alreadyJoined}
  <main class="flex min-h-svh flex-col px-6 py-12">
    <div class="flex flex-1 flex-col">
      <p class="text-center text-sm font-semibold text-[var(--primary)]">NotTennis</p>

      <div class="mt-10 space-y-3">
        <h1 class="text-[32px] font-[800]">
          {#if creatorName}
            {$_('invite_title_with_creator', { values: { creator: creatorName } })}
          {:else}
            {$_('invite_title_generic')}
          {/if}
        </h1>
        <p class="text-[var(--text-secondary)]">{$_('invite_subtitle')}</p>
      </div>

      <div class="mt-10 space-y-5">
        <div class="space-y-2.5">
          <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('invite_name_label')}</p>
          <form onsubmit={(e) => { e.preventDefault(); join(); }}>
            <Input
              bind:value={joinName}
              placeholder={$_('invite_name_placeholder')}
              maxlength={32}
              class="rounded-2xl border-0 bg-[var(--surface-raised)] px-4 py-3.5 text-sm"
            />
          </form>
        </div>

        {#if joinError}
          <p class="text-sm text-[var(--destructive)]">{joinError}</p>
        {/if}

        <Button
          onclick={join}
          disabled={joining || !joinName.trim()}
          class="h-auto w-full rounded-2xl bg-[var(--primary)] px-4 py-4 text-[15px] font-semibold text-white hover:bg-[var(--primary-hover)]"
        >
          {joining ? $_('invite_joining') : $_('invite_join_button')}
        </Button>
      </div>
    </div>
  </main>

<!-- ── Lobby (admin or already joined) ── -->
{:else}
  <main class="mx-auto max-w-[480px] px-6 py-6 space-y-6">
    <nav class="flex items-center justify-between">
      <div class="space-y-0.5">
        <p class="text-xs text-[var(--text-secondary)]">{$_('lobby_waiting')}</p>
        <p class="text-sm font-semibold text-[var(--primary)]">{session.name || 'NotTennis'}</p>
      </div>
      <div class="text-right text-xs text-[var(--text-secondary)]">
        {$_(session.courts === 1 ? 'active_courts_one' : 'active_courts_other', { values: { n: session.courts } })} · {session.points} pts · Americano
      </div>
    </nav>

    <!-- Join code + share -->
    <div class="rounded-2xl bg-[var(--surface-raised)] px-5 py-4 space-y-3">
      <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('lobby_join_code')}</p>
      <div class="flex gap-2">
        {#each session.id.split('') as char}
          <div class="flex flex-1 items-center justify-center rounded-xl bg-[var(--surface)] py-3 text-2xl font-[700] text-[var(--text-primary)] font-mono">
            {char}
          </div>
        {/each}
      </div>
      <div class="flex items-center gap-2 rounded-xl bg-[var(--surface)] px-3 py-2.5">
        <span class="flex-1 truncate text-xs text-[var(--text-secondary)]">{joinUrl}</span>
        <Button
          onclick={copyLink}
          variant="ghost"
          class="h-auto shrink-0 p-1.5 text-[var(--primary)] hover:bg-transparent hover:text-[var(--primary-hover)]"
        >
          {#if copied}
            <Check size={16} />
          {:else}
            <Share2 size={16} />
          {/if}
        </Button>
      </div>
    </div>

    <!-- Admin: add player manually -->
    {#if isAdmin}
      <div class="space-y-2">
        <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('lobby_add_player_label')}</p>
        <form onsubmit={(e) => { e.preventDefault(); join(); }} class="flex gap-2">
          <Input
            bind:value={joinName}
            placeholder={$_('lobby_add_player_placeholder')}
            maxlength={32}
            class="flex-1 rounded-2xl border-0 bg-[var(--surface-raised)] px-4 py-3 text-sm"
          />
          <Button
            type="submit"
            disabled={joining || !joinName.trim()}
            class="h-auto rounded-2xl bg-[var(--primary)] px-4 text-sm font-semibold text-white"
          >
            {joining ? $_('lobby_add_loading') : $_('lobby_add_button')}
          </Button>
        </form>
        {#if joinError}
          <p class="text-sm text-[var(--destructive)]">{joinError}</p>
        {/if}
      </div>
    {/if}

    <!-- Player list -->
    <div class="space-y-2">
      <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">
        {$_('lobby_players_label')} ({activePlayers.length})
      </p>
      {#if activePlayers.length === 0}
        <p class="text-sm text-[var(--text-disabled)]">{$_('lobby_waiting_players')}</p>
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
                  <span class="text-xs text-[var(--text-disabled)]">{$_('lobby_you')}</span>
                {/if}
                {#if isAdmin && player.id !== session.creator_player_id}
                  <button
                    onclick={() => removePlayer(player.id)}
                    class="ml-1 flex h-5 w-5 items-center justify-center rounded-full text-[var(--text-disabled)] transition-colors hover:bg-[var(--destructive)]/10 hover:text-[var(--destructive)]"
                    aria-label="Remove player"
                  >
                    ×
                  </button>
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
          <Input
            bind:value={joinName}
            placeholder={$_('lobby_join_placeholder')}
            maxlength={32}
            class="flex-1 rounded-2xl border-0 bg-[var(--surface-raised)] px-4 py-3 text-sm"
          />
          <Button
            type="submit"
            disabled={joining || !joinName.trim()}
            class="h-auto rounded-2xl bg-[var(--primary)] px-4 text-sm font-semibold text-white"
          >
            {joining ? $_('lobby_join_loading') : $_('lobby_join_button')}
          </Button>
        </form>
        {#if joinError}
          <p class="text-sm text-[var(--destructive)]">{joinError}</p>
        {/if}
      </div>
    {/if}

    <!-- Admin controls -->
    {#if isAdmin}
      <div class="space-y-2">
        <Button
          onclick={start}
          disabled={starting || !canStart}
          class="h-auto w-full rounded-2xl bg-[var(--primary)] px-4 py-4 text-[15px] font-semibold text-white hover:bg-[var(--primary-hover)]"
        >
          {starting ? $_('lobby_start_loading') : $_('lobby_start_button')}
        </Button>
        {#if !canStart}
          <p class="text-center text-xs text-[var(--text-disabled)]">
            {$_('lobby_need_players', { values: { n: session.courts * 4 } })}
          </p>
        {/if}
        <button
          onclick={() => (showCancelDialog = true)}
          disabled={cancelling}
          class="h-auto w-full rounded-2xl border border-[var(--border)] px-4 py-3.5 text-sm font-semibold text-[var(--text-secondary)] transition-colors hover:border-[var(--destructive)] hover:text-[var(--destructive)] disabled:opacity-40"
        >
          {cancelling ? $_('lobby_cancelling') : $_('lobby_cancel')}
        </button>
        {#if isDev}
          <div class="flex justify-center pt-2">
            <button
              onclick={seedPlayers}
              disabled={seeding}
              class="rounded-full border border-dashed border-[var(--border)] px-4 py-1.5 text-xs text-[var(--text-disabled)] transition-colors hover:border-[var(--text-disabled)] hover:text-[var(--text-secondary)] disabled:opacity-40"
            >
              {seeding ? $_('lobby_dev_seeding') : $_('lobby_dev_seed')}
            </button>
          </div>
        {/if}
      </div>
    {:else}
      <div class="rounded-2xl bg-[var(--surface-raised)] px-4 py-3 text-center">
        <p class="text-sm text-[var(--text-secondary)]">{$_('lobby_waiting_admin')}</p>
      </div>
    {/if}
  </main>
{/if}

<ConfirmDialog
  bind:open={showCancelDialog}
  title={$_('cancel_dialog_title')}
  description={$_('cancel_dialog_desc')}
  confirmLabel={$_('cancel_dialog_confirm')}
  cancelLabel={$_('cancel_dialog_cancel')}
  destructive
  onconfirm={cancel}
/>
