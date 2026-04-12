<script lang="ts">
  import { api } from '$lib/api/client';
  import { Crown, Share, Check, Search, UserPlus, Clock, Info } from 'lucide-svelte';
  import { onMount } from 'svelte';
  import { initials, sessionName } from '$lib/utils';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import Avatar from '$lib/components/ui/Avatar.svelte';
  import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
  import { _ } from 'svelte-i18n';
  import { auth } from '$lib/auth.svelte';
  import Footer from '$lib/components/Footer.svelte';
  import { toast } from 'svelte-sonner';
  import { ApiError } from '$lib/api/client';
  import { translateApiError } from '$lib/i18n/errors';

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

  let playerSearch = $state('');
  let playerResults = $state<App.UserSearchResult[]>([]);
  let playerSearchLoading = $state(false);
  let playerSearchDebounce: ReturnType<typeof setTimeout>;
  let sessionInvites = $state<App.Invite[]>([]);

  onMount(async () => {
    if (isAdmin) {
      sessionInvites = await api.invites.listForSession(session.id).catch(() => []);
    }
  });

  const joinedUserIds = $derived(new Set(session.players.map(p => p.user_id).filter(Boolean)));
  const pendingInvites = $derived(sessionInvites.filter(inv => !joinedUserIds.has(inv.to_user_id)));

  function onPlayerSearchInput() {
    clearTimeout(playerSearchDebounce);
    if (playerSearch.length < 2) { playerResults = []; playerSearchLoading = false; return; }
    playerSearchLoading = true;
    playerSearchDebounce = setTimeout(async () => {
      try {
        playerResults = await api.contacts.search(auth.token!, playerSearch);
      } finally {
        playerSearchLoading = false;
      }
    }, 300);
  }

  async function inviteUser(userID: string) {
    if (!auth.token) return;
    await api.invites.send(session.id, userID, auth.token).catch(() => {});
    playerSearch = '';
    playerResults = [];
    sessionInvites = await api.invites.listForSession(session.id).catch(() => []);
  }

  async function addGuest(name: string) {
    if (!name) return;
    joining = true;
    try {
      await api.players.join(session.id, name);
      playerSearch = '';
      toast.success($_('lobby_player_joined'));
      onRefresh();
    } catch (e) {
      toast.error(e instanceof ApiError ? translateApiError(e.message) : translateApiError('server_error'));
    } finally {
      joining = false;
    }
  }

  // Pre-fill name from account when on the invite screen
  $effect(() => {
    if (auth.user && !isAdmin && !alreadyJoined && !joinName) {
      joinName = auth.user.display_name;
    }
  });

  const joinUrl = $derived(
    typeof location !== 'undefined' ? `${location.origin}/s/${session.id}` : ''
  );

  const isTennis = $derived(session.game_mode === 'tennis');
  const isMexicano = $derived(session.game_mode === 'mexicano');
  const gameModeName = $derived(
    session.game_mode === 'mexicano' ? 'Mexicano' : session.game_mode === 'tennis' ? 'Tennis' : 'Americano'
  );
  let showRules = $state(false);
  const activePlayers = $derived(session.players.filter((p) => p.active));

  // Tennis team state — maps player_id → 'a' | 'b' | null
  let teamAssignments = $state<Record<string, 'a' | 'b'>>({});
  let savingTeams = $state(false);
  let draggingId = $state<string | null>(null);
  let dragOverZone = $state<'a' | 'b' | 'pool' | null>(null);

  // Touch drag state
  let touchGhost = $state<{ name: string; x: number; y: number } | null>(null);
  let zoneAEl = $state<HTMLElement | null>(null);
  let zoneBEl = $state<HTMLElement | null>(null);
  let zonePoolEl = $state<HTMLElement | null>(null);

  function getZoneAtPoint(x: number, y: number): 'a' | 'b' | 'pool' | null {
    for (const [zone, el] of [['a', zoneAEl], ['b', zoneBEl], ['pool', zonePoolEl]] as const) {
      if (!el) continue;
      const r = el.getBoundingClientRect();
      if (x >= r.left && x <= r.right && y >= r.top && y <= r.bottom) return zone;
    }
    return null;
  }

  function onTouchStart(e: TouchEvent, playerId: string, name: string) {
    const t = e.touches[0];
    draggingId = playerId;
    touchGhost = { name, x: t.clientX, y: t.clientY };
  }

  function onTouchMove(e: TouchEvent) {
    if (!draggingId) return;
    try { e.preventDefault(); } catch {}
    const t = e.touches[0];
    const player = activePlayers.find(p => p.id === draggingId);
    if (player) touchGhost = { name: player.name, x: t.clientX, y: t.clientY };
    dragOverZone = getZoneAtPoint(t.clientX, t.clientY);
  }

  // Svelte makes touchmove passive by default — use an action to register non-passive.
  function nonPassiveTouchMove(node: HTMLElement) {
    node.addEventListener('touchmove', onTouchMove, { passive: false });
    return { destroy() { node.removeEventListener('touchmove', onTouchMove); } };
  }

  function onTouchEnd(e: TouchEvent) {
    if (!draggingId) return;
    const t = e.changedTouches[0];
    const zone = getZoneAtPoint(t.clientX, t.clientY);
    if (zone === 'pool') {
      unassignPlayer(draggingId);
    } else if (zone === 'a' || zone === 'b') {
      assignPlayer(draggingId, zone);
    }
    draggingId = null;
    touchGhost = null;
    dragOverZone = null;
  }

  const teamA = $derived(activePlayers.filter((p) => teamAssignments[p.id] === 'a'));
  const teamB = $derived(activePlayers.filter((p) => teamAssignments[p.id] === 'b'));
  const unassigned = $derived(activePlayers.filter((p) => !teamAssignments[p.id]));

  const requiredPlayers = $derived(session.courts * 4);
  const canStart = $derived(
    isTennis
      ? teamA.length === 2 && teamB.length === 2
      : isMexicano
        ? activePlayers.length === requiredPlayers
        : activePlayers.length >= requiredPlayers
  );

  const creatorName = $derived(activePlayers.find((p) => p.id === session.creator_player_id)?.name ?? '');

  function unassignPlayer(playerId: string) {
    const next = { ...teamAssignments };
    delete next[playerId];
    teamAssignments = next;
    saveTeams();
  }

  function assignPlayer(playerId: string, team: 'a' | 'b') {
    const current = teamAssignments[playerId];
    // Clicking a player already on this team moves them back to pool
    if (current === team) {
      unassignPlayer(playerId);
      return;
    }
    // Check team capacity
    const count = activePlayers.filter((p) => teamAssignments[p.id] === team).length;
    if (count >= 2) return;
    teamAssignments = { ...teamAssignments, [playerId]: team };
    saveTeams();
  }

  async function saveTeams() {
    if (!isAdmin) return;
    savingTeams = true;
    const adminToken = localStorage.getItem(`admin_token_${session.id}`) ?? '';
    const teams = activePlayers
      .filter((p) => teamAssignments[p.id])
      .map((p) => ({ player_id: p.id, team: teamAssignments[p.id] as 'a' | 'b' }));
    try {
      await api.tennis.setTeams(session.id, teams, adminToken);
    } catch (e) {
      toast.error(e instanceof ApiError ? translateApiError(e.message) : translateApiError('server_error'));
    } finally {
      savingTeams = false;
    }
  }

  const myPlayerId = $derived(
    typeof localStorage !== 'undefined' ? localStorage.getItem(`player_id_${session.id}`) : null
  );
  const alreadyJoined = $derived(
    (!!myPlayerId && activePlayers.some((p) => p.id === myPlayerId)) ||
    (!!auth.user && activePlayers.some((p) => p.user_id === auth.user!.id))
  );

  async function copyLink() {
    if (navigator.share) {
      await navigator.share({ title: sessionName(session), url: joinUrl }).catch(() => {});
    } else {
      await navigator.clipboard.writeText(joinUrl);
      copied = true;
      setTimeout(() => (copied = false), 2000);
    }
  }

  async function join() {
    const name = joinName.trim();
    if (!name) return;
    joining = true;
    try {
      const player = await api.players.join(session.id, name, isAdmin ? undefined : (auth.token ?? undefined));
      if (!isAdmin) {
        localStorage.setItem(`player_id_${session.id}`, player.id);
        localStorage.setItem('last_session_id', session.id);
      }
      toast.success($_('lobby_player_joined'));
      joinName = '';
      onRefresh();
    } catch (e) {
      toast.error(e instanceof ApiError ? translateApiError(e.message) : translateApiError('server_error'));
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
{:else if !isAdmin && !alreadyJoined && isTennis && activePlayers.length >= 4}
  <main class="flex min-h-svh flex-col items-center justify-center px-6 gap-4">
    <div class="w-full max-w-sm rounded-2xl bg-[var(--surface-raised)] px-5 py-6 text-center space-y-2">
      <p class="text-lg font-[800]">{sessionName(session)}</p>
      <p class="text-sm text-[var(--text-secondary)]">{$_('tennis_session_full')}</p>
    </div>
    <a href="/" class="text-sm text-[var(--text-disabled)] hover:text-[var(--text-secondary)]">← {$_('auth_back_home')}</a>
  </main>
{:else if !isAdmin && !alreadyJoined}
  <main class="flex min-h-svh flex-col items-center px-6 py-12">
    <div class="flex w-full max-w-sm justify-end">
      <a href="/" class="flex h-7 w-7 items-center justify-center rounded-full text-[var(--text-disabled)] transition-colors hover:bg-[var(--surface-raised)] hover:text-[var(--text-secondary)]" aria-label="Back">×</a>
    </div>
    <div class="flex w-full max-w-sm flex-1 flex-col justify-center space-y-8">

      <!-- Brand + session info -->
      <div class="space-y-1">
        <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-[var(--primary)]">OpenPadel</p>
        <div class="flex items-start gap-2">
          <h1 class="text-[28px] font-[800] leading-tight">
            {#if creatorName}
              {$_('invite_title_with_creator', { values: { creator: creatorName, mode: gameModeName } })}
            {:else}
              {$_('invite_title_generic', { values: { mode: gameModeName } })}
            {/if}
          </h1>
          <button
            onclick={() => (showRules = true)}
            aria-label={$_('lobby_rules_button')}
            class="mt-1.5 shrink-0 text-[var(--text-disabled)] hover:text-[var(--text-secondary)] transition-colors"
          >
            <Info size={18} />
          </button>
        </div>
        {#if session.name}
          <p class="text-[var(--text-secondary)]">{session.name}</p>
        {/if}
        <p class="text-sm text-[var(--text-secondary)]">
          {#if isTennis}
            {$_('create_mode_tennis')} · {$_(session.sets_to_win === 3 ? 'create_sets_bo5' : 'create_sets_bo3')}{#if session.scheduled_at} · {new Date(session.scheduled_at).toLocaleString(undefined, { weekday: 'short', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}{/if}
          {:else}
            {$_(session.courts === 1 ? 'active_courts_one' : 'active_courts_other', { values: { n: session.courts } })} · {session.points} {$_('invite_points')} · {gameModeName}{#if session.rounds_total} · {session.rounds_total} rds{:else if session.court_duration_minutes} · {session.court_duration_minutes} min{/if}{#if session.scheduled_at} · {new Date(session.scheduled_at).toLocaleString(undefined, { weekday: 'short', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}{/if}
          {/if}
        </p>
      </div>

      <div class="space-y-4">
        {#if auth.user}
          <!-- Logged in: show account card + join -->
          <div class="rounded-2xl bg-[var(--surface-raised)] px-4 py-3.5 flex items-center gap-3">
            <Avatar icon={auth.user.avatar_icon} color={auth.user.avatar_color} name={auth.user.display_name} ring="ring-2 ring-[var(--primary)]/30" />
            <div class="flex-1 min-w-0">
              <p class="text-sm font-semibold truncate">{auth.user.display_name}</p>
              <p class="text-xs text-[var(--text-secondary)] truncate">{auth.user.email}</p>
            </div>
          </div>

          <Button
            onclick={join}
            disabled={joining || !joinName.trim()}
            class="h-auto w-full rounded-2xl bg-[var(--primary)] px-4 py-4 text-[15px] font-semibold text-white hover:bg-[var(--primary-hover)]"
          >
            {joining ? $_('invite_joining') : $_('invite_join_button')}
          </Button>
        {:else}
          <!-- Not logged in: account options first, guest below -->
          <a
            href="/auth?redirect=/s/{session.id}"
            class="flex h-auto w-full items-center justify-center rounded-2xl bg-[var(--primary)] px-4 py-4 text-[15px] font-semibold text-white hover:bg-[var(--primary-hover)]"
          >
            {$_('invite_sign_in')}
          </a>
          <p class="text-center text-sm text-[var(--text-secondary)]">
            {$_('invite_no_account')}
            <a href="/auth?register=1&redirect=/s/{session.id}" class="font-semibold text-[var(--primary)]">{$_('invite_create_account')}</a>
          </p>

          <div class="flex items-center gap-3">
            <div class="h-px flex-1 bg-[var(--border)]"></div>
            <span class="text-xs text-[var(--text-disabled)]">{$_('invite_or_guest')}</span>
            <div class="h-px flex-1 bg-[var(--border)]"></div>
          </div>

          <!-- Guest fallback -->
          <form onsubmit={(e) => { e.preventDefault(); join(); }} class="flex gap-2">
            <Input
              bind:value={joinName}
              placeholder={$_('invite_name_placeholder')}
              maxlength={32}
              class="flex-1 rounded-2xl border-0 bg-[var(--surface-raised)] px-4 py-3 text-sm"
            />
            <Button
              type="submit"
              disabled={joining || !joinName.trim()}
              class="h-auto rounded-2xl bg-[var(--surface-raised)] px-4 text-sm font-semibold text-[var(--text-secondary)] hover:text-[var(--text-primary)] shadow-none"
            >
              {joining ? '…' : $_('invite_guest_join')}
            </Button>
          </form>

        {/if}
      </div>

    </div>
    <Footer />
  </main>

<!-- ── Lobby (admin or already joined) ── -->
{:else}
  <main class="mx-auto max-w-[480px] px-6 py-6 space-y-6">
    <nav class="flex items-center justify-between">
      <div class="space-y-0.5">
        <p class="text-xs text-[var(--text-secondary)]">
          {#if session.scheduled_at}
            {new Date(session.scheduled_at).toLocaleString(undefined, { weekday: 'short', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}
          {:else}
            {$_('lobby_waiting')}
          {/if}
        </p>
        <p class="text-sm font-semibold text-[var(--primary)]">{sessionName(session)}</p>
      </div>
      <div class="flex items-center gap-2">
        <div class="text-right text-xs text-[var(--text-secondary)]">
          {#if isTennis}
            {$_('create_mode_tennis')} · {$_(session.sets_to_win === 3 ? 'create_sets_bo5' : 'create_sets_bo3')}
          {:else}
            {$_(session.courts === 1 ? 'active_courts_one' : 'active_courts_other', { values: { n: session.courts } })} · {session.points} pts · {gameModeName}{#if session.rounds_total} · {session.rounds_total} rds{:else if session.court_duration_minutes} · {session.court_duration_minutes} min{/if}
          {/if}
        </div>
        <button
          onclick={() => (showRules = true)}
          aria-label={$_('lobby_rules_button')}
          class="text-[var(--text-disabled)] hover:text-[var(--text-secondary)] transition-colors"
        >
          <Info size={16} />
        </button>
        {#if isAdmin || alreadyJoined}
          <a
            href="/"
            class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-[var(--text-disabled)] transition-colors hover:bg-[var(--surface-raised)] hover:text-[var(--text-secondary)]"
            aria-label="Back to home"
          >×</a>
        {/if}
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
            <Share size={16} />
          {/if}
        </Button>
      </div>
    </div>

    <!-- Admin: invite or add guest -->
    {#if isAdmin && !(isTennis && activePlayers.length >= 4)}
      <div class="space-y-2">
        <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('lobby_add_player_label')}</p>
        <div class="relative">
          <div class="pointer-events-none absolute inset-y-0 left-3.5 flex items-center">
            <Search size={15} class="text-[var(--text-disabled)]" />
          </div>
          <Input
            bind:value={playerSearch}
            oninput={onPlayerSearchInput}
            placeholder={$_('lobby_add_player_placeholder')}
            maxlength={32}
            class="w-full rounded-2xl border-0 bg-[var(--surface-raised)] py-3 pl-9 pr-4 text-sm"
          />
        </div>
        {#if playerResults.length > 0}
          <div class="space-y-1.5">
            {#each playerResults as result}
              <div class="flex items-center gap-3 rounded-2xl bg-[var(--surface-raised)] px-4 py-3">
                <Avatar icon={result.avatar_icon} color={result.avatar_color} name={result.display_name} size="sm" ring="ring-2 ring-[var(--primary)]/30" />
                <p class="flex-1 text-sm font-semibold truncate">{result.display_name}</p>
                <button
                  onclick={() => inviteUser(result.id)}
                  class="flex items-center gap-1 rounded-full bg-[var(--primary)] px-3 py-1.5 text-xs font-semibold text-white"
                >
                  <UserPlus size={12} /> Invite
                </button>
              </div>
            {/each}
          </div>
        {/if}
        {#if playerSearch.trim().length > 0 && !playerSearchLoading}
          <button
            onclick={() => addGuest(playerSearch.trim())}
            disabled={joining}
            class="flex w-full items-center gap-3 rounded-2xl border border-dashed border-[var(--border)] px-4 py-3 text-sm text-[var(--text-secondary)] transition-colors hover:border-[var(--primary)] hover:text-[var(--primary)] disabled:opacity-50"
          >
            <UserPlus size={15} class="shrink-0" />
            Add "{playerSearch.trim()}" as guest
          </button>
        {/if}
      </div>
    {/if}

    <!-- Player list -->
    <div class="space-y-2">
      <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">
        {$_('lobby_players_label')} ({activePlayers.length})
      </p>
      {#if activePlayers.length === 0 && sessionInvites.length === 0}
        <p class="text-sm text-[var(--text-disabled)]">{$_('lobby_waiting_players')}</p>
      {:else}
        <div class="rounded-2xl bg-[var(--surface-raised)] divide-y divide-[var(--border)]">
          {#each activePlayers as player (player.id)}
            <div class="flex items-center gap-3 px-4 py-3">
              <Avatar icon={player.avatar_icon} color={player.avatar_color} name={player.name} size="sm" ring="ring-2 ring-[var(--primary)]/30" />
              <span class="text-sm font-medium">{player.name}</span>
              <div class="ml-auto flex items-center gap-1.5">
                {#if player.id === session.creator_player_id}
                  <Crown size={13} class="text-[var(--primary)]" />
                {/if}
                {#if player.id === myPlayerId}
                  <span class="text-xs text-[var(--text-disabled)]">{$_('lobby_you')}</span>
                {/if}
                {#if isAdmin && player.id !== session.creator_player_id && player.id !== myPlayerId}
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
          {#each pendingInvites as invite (invite.id)}
            <div class="flex items-center gap-3 px-4 py-3 opacity-60">
              <Avatar name={invite.to_display_name ?? '?'} size="sm" ring="ring-2 ring-[var(--primary)]/30" />
              <span class="flex-1 text-sm font-medium text-[var(--text-secondary)] truncate">{invite.to_display_name}</span>
              <div class="ml-auto flex items-center gap-1 text-[var(--text-disabled)]">
                <Clock size={11} />
                <span class="text-xs">Invited</span>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>

    <!-- Tennis team assignment (admin only, shown when all 4 players present) -->
    {#if isTennis && isAdmin && activePlayers.length >= 4}
      <!-- Touch ghost element -->
      {#if touchGhost}
        <div class="pointer-events-none fixed z-50 -translate-x-1/2 -translate-y-1/2 flex items-center gap-2 rounded-full bg-[var(--primary)] px-3 py-1.5 text-sm font-semibold text-white shadow-lg opacity-90"
          style="left:{touchGhost.x}px;top:{touchGhost.y}px">
          <div class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-white/20 text-[10px] font-bold">{initials(touchGhost.name)}</div>
          <span>{touchGhost.name}</span>
        </div>
      {/if}

      <div class="space-y-3" use:nonPassiveTouchMove ontouchend={onTouchEnd} role="group" aria-label="Team assignment">
        <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('tennis_assign_teams')}</p>

        <!-- Unassigned pool -->
        <div
          bind:this={zonePoolEl}
          class="min-h-[52px] rounded-2xl border-2 border-dashed border-[var(--border)] px-3 py-2 flex flex-wrap gap-2 transition-colors {dragOverZone === 'pool' ? 'border-[var(--primary)] bg-[var(--primary-muted)]' : ''}"
          ondragover={(e) => { e.preventDefault(); dragOverZone = 'pool'; }}
          ondragleave={() => { if (dragOverZone === 'pool') dragOverZone = null; }}
          ondrop={(e) => { e.preventDefault(); dragOverZone = null; if (draggingId) { unassignPlayer(draggingId); draggingId = null; } }}
          role="region"
          aria-label="Unassigned players"
        >
          {#if unassigned.length === 0}
            <p class="text-xs text-[var(--text-disabled)] py-1">All assigned</p>
          {/if}
          {#each unassigned as player (player.id)}
            <div
              draggable="true"
              ondragstart={(e) => { draggingId = player.id; e.dataTransfer!.effectAllowed = 'move'; }}
              ondragend={() => { draggingId = null; dragOverZone = null; }}
              ontouchstart={(e) => onTouchStart(e, player.id, player.name)}
              class="flex touch-none cursor-grab items-center gap-2 rounded-full bg-[var(--surface-raised)] px-3 py-1.5 text-sm font-semibold select-none {draggingId === player.id ? 'opacity-40' : ''}"
            >
              <Avatar icon={player.avatar_icon} color={player.avatar_color} name={player.name} size="sm" ring="ring-2 ring-[var(--primary)]/30" />
              <span>{player.name}</span>
            </div>
          {/each}
        </div>

        <!-- Team columns -->
        <div class="grid grid-cols-2 gap-3">
          <!-- Team A drop zone -->
          <div
            bind:this={zoneAEl}
            class="min-h-[96px] rounded-2xl border-2 transition-colors p-3 space-y-2 {dragOverZone === 'a' ? 'border-[var(--primary)] bg-[var(--primary-muted)]' : 'border-transparent bg-[var(--surface-raised)]'}"
            ondragover={(e) => { e.preventDefault(); dragOverZone = 'a'; }}
            ondragleave={() => { if (dragOverZone === 'a') dragOverZone = null; }}
            ondrop={(e) => { e.preventDefault(); dragOverZone = null; if (draggingId) { assignPlayer(draggingId, 'a'); draggingId = null; } }}
            role="region"
            aria-label="Team A"
          >
            <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-[var(--primary)]">{$_('tennis_team_a')}</p>
            {#each teamA as player (player.id)}
              <div
                draggable="true"
                ondragstart={(e) => { draggingId = player.id; e.dataTransfer!.effectAllowed = 'move'; }}
                ondragend={() => { draggingId = null; dragOverZone = null; }}
                ontouchstart={(e) => onTouchStart(e, player.id, player.name)}
                onclick={() => assignPlayer(player.id, 'a')}
                class="flex touch-none cursor-grab items-center gap-2 rounded-xl bg-[var(--primary)] px-3 py-2 text-sm font-semibold text-white select-none {draggingId === player.id ? 'opacity-40' : ''}"
                role="button"
                tabindex="0"
                onkeydown={(e) => e.key === 'Enter' && assignPlayer(player.id, 'a')}
              >
                <Avatar icon={player.avatar_icon} color={player.avatar_color} name={player.name} size="sm" ring="ring-2 ring-[var(--primary)]/30" />
                <span class="truncate">{player.name}</span>
              </div>
            {/each}
            {#if teamA.length < 2}
              <p class="text-xs text-[var(--text-disabled)]">Drop here</p>
            {/if}
          </div>

          <!-- Team B drop zone -->
          <div
            bind:this={zoneBEl}
            class="min-h-[96px] rounded-2xl border-2 transition-colors p-3 space-y-2 {dragOverZone === 'b' ? 'border-[var(--primary)] bg-[var(--primary-muted)]' : 'border-transparent bg-[var(--surface-raised)]'}"
            ondragover={(e) => { e.preventDefault(); dragOverZone = 'b'; }}
            ondragleave={() => { if (dragOverZone === 'b') dragOverZone = null; }}
            ondrop={(e) => { e.preventDefault(); dragOverZone = null; if (draggingId) { assignPlayer(draggingId, 'b'); draggingId = null; } }}
            role="region"
            aria-label="Team B"
          >
            <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('tennis_team_b')}</p>
            {#each teamB as player (player.id)}
              <div
                draggable="true"
                ondragstart={(e) => { draggingId = player.id; e.dataTransfer!.effectAllowed = 'move'; }}
                ondragend={() => { draggingId = null; dragOverZone = null; }}
                ontouchstart={(e) => onTouchStart(e, player.id, player.name)}
                onclick={() => assignPlayer(player.id, 'b')}
                class="flex touch-none cursor-grab items-center gap-2 rounded-xl bg-[var(--border)] px-3 py-2 text-sm font-semibold text-[var(--text-primary)] select-none {draggingId === player.id ? 'opacity-40' : ''}"
                role="button"
                tabindex="0"
                onkeydown={(e) => e.key === 'Enter' && assignPlayer(player.id, 'b')}
              >
                <Avatar icon={player.avatar_icon} color={player.avatar_color} name={player.name} size="sm" ring="ring-2 ring-[var(--primary)]/30" />
                <span class="truncate">{player.name}</span>
              </div>
            {/each}
            {#if teamB.length < 2}
              <p class="text-xs text-[var(--text-disabled)]">Drop here</p>
            {/if}
          </div>
        </div>
      </div>
    {/if}

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
            {#if isTennis}
              {$_(activePlayers.length < 4 ? 'lobby_need_players' : 'tennis_start_locked', { values: { n: 4 } })}
            {:else if isMexicano}
              {$_('lobby_mexicano_exact_players', { values: { n: requiredPlayers, current: activePlayers.length } })}
            {:else}
              {$_('lobby_need_players', { values: { n: requiredPlayers } })}
            {/if}
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

{#if showRules}
  <div
    class="fixed inset-0 z-50 flex items-end justify-center bg-black/40 backdrop-blur-sm pb-safe sm:items-center"
    onclick={() => (showRules = false)}
    onkeydown={(e) => e.key === 'Escape' && (showRules = false)}
    role="presentation"
    tabindex="-1"
  >
    <div
      class="w-full max-w-sm rounded-t-3xl bg-[var(--surface)] px-6 pb-8 pt-6 space-y-4 sm:rounded-3xl sm:mx-4"
      onclick={(e) => e.stopPropagation()}
      role="presentation"
    >
      <div class="mx-auto h-1 w-10 rounded-full bg-[var(--border)] sm:hidden"></div>
      <h2 class="text-[18px] font-[800]">{gameModeName}</h2>

      <ul class="space-y-2">
        {#each $_(`rules_${session.game_mode}`).split('\n') as line}
          {#if line.trim()}
            <li class="flex gap-2 text-sm text-[var(--text-secondary)]">
              <span class="mt-0.5 shrink-0 text-[var(--primary)]">·</span>
              <span>{line.trim()}</span>
            </li>
          {/if}
        {/each}
      </ul>

      <button
        onclick={() => (showRules = false)}
        class="w-full rounded-2xl border border-[var(--border)] px-4 py-3.5 text-sm font-semibold text-[var(--text-secondary)] transition-colors hover:bg-[var(--surface-raised)]"
      >
        {$_('leaderboard_close')}
      </button>
    </div>
  </div>
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
