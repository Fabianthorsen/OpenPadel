<script lang="ts">
  import { api, ApiError } from '$lib/api/client';
  import { Crown, Share, Check, Search, UserPlus, Clock, Info } from 'lucide-svelte';
  import { onMount } from 'svelte';
  import { sessionName } from '$lib/utils';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import Avatar from '$lib/components/ui/Avatar.svelte';
  import { SectionLabel } from '$lib/components/ui/section-label';
  import { PillToggleGroup, PillToggleItem } from '$lib/components/ui/pill-toggle-group';
  const MAX_COURTS = 4;
  import { Calendar } from '$lib/components/ui/calendar';
  import Stepper from '$lib/components/ui/stepper/Stepper.svelte';
  import * as Dialog from '$lib/components/ui/dialog';
  import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
  import { _ } from 'svelte-i18n';
  import { auth } from '$lib/auth.svelte';
  import Footer from '$lib/components/Footer.svelte';
  import { toast } from 'svelte-sonner';
  import { translateApiError } from '$lib/i18n/errors';
  import { goto } from '$app/navigation';
  import { type DateValue, today, getLocalTimeZone, parseAbsoluteToLocal } from '@internationalized/date';
  import type { sessionStream } from '$lib/stores/sessionStream.svelte';
  type SessionStream = ReturnType<typeof sessionStream>;

  let {
    session,
    isAdmin,
    onRefresh,
    onStarted,
    stream,
  }: {
    session: App.Session;
    isAdmin: boolean;
    onRefresh: () => void;
    onStarted: () => void;
    stream?: SessionStream;
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

  // ── Inline config editing state ──
  let editingName = $state(false);
  let nameInputEl = $state<HTMLInputElement | null>(null);
  $effect(() => { if (editingName && nameInputEl) nameInputEl.focus(); });
  let nameInput = $state('');
  let configMode = $state<'americano' | 'mexicano'>('americano');
  let configCourts = $state(2);
  let configPoints = $state(24);
  let configRounds = $state(7);
  let scheduleEnabled = $state(false);
  let calendarDate = $state<DateValue | undefined>(undefined);
  let timeSlot = $state(20);

  // Sync config state whenever session prop changes (after refresh).
  $effect(() => {
    nameInput = session.name ?? '';
    configMode = session.game_mode as 'americano' | 'mexicano';
    configCourts = session.courts;
    configPoints = session.points;
    configRounds = session.rounds_total ?? 7;
    scheduleEnabled = !!session.scheduled_at;
    if (session.scheduled_at) {
      try { calendarDate = parseAbsoluteToLocal(session.scheduled_at); } catch { calendarDate = undefined; }
      const d = new Date(session.scheduled_at);
      const slot = Math.round((d.getHours() * 60 + d.getMinutes() - 8 * 60) / 30);
      timeSlot = Math.max(0, Math.min(27, slot));
    } else {
      calendarDate = undefined;
    }
  });

  function slotToLabel(slot: number) {
    const totalMins = 8 * 60 + slot * 30;
    const h = String(Math.floor(totalMins / 60)).padStart(2, '0');
    const m = String(totalMins % 60).padStart(2, '0');
    return `${h}:${m}`;
  }
  const scheduleTime = $derived(slotToLabel(timeSlot));
  const timeHour = $derived(8 + Math.floor(timeSlot / 2));
  const timeMinute = $derived((timeSlot % 2) * 30);

  function onHourChange(h: number) {
    timeSlot = (h - 8) * 2 + (timeSlot % 2);
    commitScheduleTime();
  }
  function onMinuteChange(m: number) {
    timeSlot = (timeHour - 8) * 2 + Math.round(m / 30);
    commitScheduleTime();
  }

  function calculateNextHourSlot(): number {
    const now = new Date();
    const nextHour = now.getMinutes() > 0 ? now.getHours() + 1 : now.getHours();
    const clamped = Math.min(21, Math.max(8, nextHour));
    return Math.round((clamped * 60 - 8 * 60) / 30);
  }


  async function patchConfig(patch: Parameters<typeof api.sessions.update>[1]) {
    const adminToken = localStorage.getItem(`admin_token_${session.id}`) ?? '';
    try {
      await api.sessions.update(session.id, patch, adminToken);
      onRefresh();
    } catch (e) {
      toast.error(e instanceof ApiError ? translateApiError(e.message) : translateApiError('server_error'));
      // Reset local state to match server
      configMode = session.game_mode as 'americano' | 'mexicano';
      configCourts = session.courts;
      configPoints = session.points;
      configRounds = session.rounds_total ?? 7;
    }
  }

  async function commitName() {
    editingName = false;
    if (nameInput === (session.name ?? '')) return;
    await patchConfig({ name: nameInput });
  }

  function onModeChange(mode: 'americano' | 'mexicano') {
    configMode = mode;
    if (mode === 'mexicano' && configCourts < 2) configCourts = 2;
    const patch: Parameters<typeof api.sessions.update>[1] = { game_mode: mode, courts: configCourts };
    if (mode === 'mexicano') patch.rounds_total = configRounds;
    patchConfig(patch);
  }

  function onCourtsChange(n: number) {
    configCourts = n;
    patchConfig({ courts: n });
  }

  function onPointsChange(n: number) {
    configPoints = n;
    patchConfig({ points: n });
  }

  function onRoundsChange(n: number) {
    configRounds = n;
    patchConfig({ rounds_total: n });
  }

  async function commitSchedule(enabled: boolean) {
    scheduleEnabled = enabled;
    if (!enabled) {
      calendarDate = undefined;
      timeSlot = 20;
      await patchConfig({ scheduled_at: '' });
      return;
    }
    calendarDate = today(getLocalTimeZone());
    timeSlot = calculateNextHourSlot();
  }

  async function commitScheduleTime() {
    if (!scheduleEnabled || !calendarDate) return;
    const [h, m] = scheduleTime.split(':').map(Number);
    const d = calendarDate.toDate(getLocalTimeZone());
    d.setHours(h, m, 0, 0);
    await patchConfig({ scheduled_at: d.toISOString() });
  }

  onMount(() => {
    if (isAdmin) {
      api.invites.listForSession(session.id).catch(() => []).then(v => { sessionInvites = v; });
    }
    if (stream) {
      return stream.onEvent('session_updated', async () => {
        if (isAdmin) {
          sessionInvites = await api.invites.listForSession(session.id).catch(() => []);
        }
      });
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

  const isMexicano = $derived(session.game_mode === 'mexicano');
  const gameModeName = $derived(
    session.game_mode === 'mexicano'
      ? $_('create_mexicano_soon')
      : 'Americano'
  );
  let showRules = $state(false);
  const activePlayers = $derived(session.players.filter((p) => p.active));

  const requiredPlayers = $derived(session.courts * 4);
  const maxPlayers = $derived(isMexicano ? requiredPlayers : undefined);
  const isFull = $derived(maxPlayers ? activePlayers.length >= maxPlayers : false);

  // Use server-computed can_start if available, otherwise fall back to local logic
  const canStart = $derived(
    session.can_start !== undefined
      ? session.can_start
      : (isMexicano
          ? activePlayers.length === requiredPlayers
          : activePlayers.length >= requiredPlayers)
  );

  const creatorName = $derived(activePlayers.find((p) => p.id === session.creator_player_id)?.name ?? '');

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
      goto('/');
    } catch {
      cancelling = false;
      showCancelDialog = false;
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
    <div class="h-8 w-8 animate-spin rounded-full border-2 border-border border-t-primary"></div>
    <p class="text-sm text-text-secondary">{$_('lobby_cancelling')}</p>
  </main>

<!-- ── Join / invite screen (visitor hasn't joined yet) ── -->
{:else if !isAdmin && !alreadyJoined}
  <main class="flex min-h-svh flex-col items-center px-6 pb-12 pt-safe-page">
    <div class="flex w-full max-w-sm justify-end">
      <a href="/" class="flex h-7 w-7 items-center justify-center rounded-full text-text-disabled transition-colors hover:bg-surface-raised hover:text-text-secondary" aria-label="Back">×</a>
    </div>
    <div class="flex w-full max-w-sm flex-1 flex-col justify-center space-y-8">

      <!-- Brand + session info -->
      <div class="space-y-1">
        <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-primary">OpenPadel</p>
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
            class="mt-1.5 shrink-0 text-text-disabled hover:text-text-secondary transition-colors"
          >
            <Info size={18} />
          </button>
        </div>
        {#if session.name}
          <p class="text-text-secondary">{session.name}</p>
        {/if}
        <p class="text-sm text-text-secondary">
          {$_(session.courts === 1 ? 'active_courts_one' : 'active_courts_other', { values: { n: session.courts } })} · {session.points + ' ' + $_('invite_points')} · {gameModeName}{#if session.rounds_total} · {session.rounds_total} rds{/if}{#if session.scheduled_at} · {new Date(session.scheduled_at).toLocaleString(undefined, { weekday: 'short', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}{/if}
        </p>
      </div>

      <div class="space-y-4">
        {#if auth.user}
          <!-- Logged in: show account card + join -->
          <div class="rounded-2xl bg-surface-raised px-4 py-3.5 flex items-center gap-3">
            <Avatar icon={auth.user.avatar_icon} color={auth.user.avatar_color} name={auth.user.display_name} ring="ring-2 ring-primary/30" />
            <div class="flex-1 min-w-0">
              <p class="text-sm font-semibold truncate">{auth.user.display_name}</p>
              <p class="text-xs text-text-secondary truncate">{auth.user.email}</p>
            </div>
          </div>

          {#if isFull}
            <div class="flex items-start gap-2 rounded-2xl bg-red-50 px-4 py-3 text-sm text-red-900">
              <Info size={16} class="shrink-0 mt-0.5" />
              <p>This session has reached its maximum player limit for {gameModeName}.</p>
            </div>
          {/if}

          <Button
            onclick={join}
            disabled={joining || !joinName.trim() || isFull}
            class="h-auto w-full rounded-2xl bg-primary px-4 py-4 text-[15px] font-semibold text-white hover:bg-primary-hover"
          >
            {joining ? $_('invite_joining') : $_('invite_join_button')}
          </Button>
        {:else}
          <!-- Not logged in: account options first, guest below -->
          <a
            href="/auth?redirect=/s/{session.id}"
            class="flex h-auto w-full items-center justify-center rounded-2xl bg-primary px-4 py-4 text-[15px] font-semibold text-white hover:bg-primary-hover"
          >
            {$_('invite_sign_in')}
          </a>
          <p class="text-center text-sm text-text-secondary">
            {$_('invite_no_account')}
            <a href="/auth?register=1&redirect=/s/{session.id}" class="font-semibold text-primary">{$_('invite_create_account')}</a>
          </p>

          <div class="flex items-center gap-3">
            <div class="h-px flex-1 bg-border"></div>
            <span class="text-xs text-text-disabled">{$_('invite_or_guest')}</span>
            <div class="h-px flex-1 bg-border"></div>
          </div>

          <!-- Guest fallback -->
          <form onsubmit={(e) => { e.preventDefault(); join(); }} class="flex gap-2">
            <Input
              bind:value={joinName}
              placeholder={$_('invite_name_placeholder')}
              maxlength={32}
              class="flex-1 rounded-2xl border-0 bg-surface-raised px-4 py-3 text-sm"
            />
            <Button
              type="submit"
              disabled={joining || !joinName.trim() || isFull}
              class="h-auto rounded-2xl bg-surface-raised px-4 text-sm font-semibold text-text-secondary hover:text-text-primary shadow-none"
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
  <main class="mx-auto max-w-[480px] px-6 pb-6 pt-safe-page space-y-6">
    <nav class="flex items-center justify-between">
      <div class="space-y-0.5">
        <p class="text-xs text-text-secondary">
          {#if session.scheduled_at}
            {new Date(session.scheduled_at).toLocaleString(undefined, { weekday: 'short', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}
          {:else}
            {$_('lobby_waiting')}
          {/if}
        </p>
        <!-- Click-to-edit name (admin only) -->
        {#if isAdmin && editingName}
          <input
            bind:this={nameInputEl}
            class="text-sm font-semibold text-primary bg-transparent border-b border-primary focus:outline-none w-full"
            bind:value={nameInput}
            maxlength={48}
            placeholder={$_('lobby_name_placeholder')}
            onblur={commitName}
            onkeydown={(e) => e.key === 'Enter' && commitName()}
          />
        {:else if isAdmin}
          <button
            class="text-sm font-semibold text-primary hover:opacity-70 transition-opacity text-left"
            onclick={() => { nameInput = session.name ?? ''; editingName = true; }}
          >
            {session.name || $_('lobby_name_placeholder')}
          </button>
        {:else}
          <p class="text-sm font-semibold text-primary">{sessionName(session)}</p>
        {/if}
      </div>
      <div class="flex items-center gap-2">
        <div class="text-right text-xs text-text-secondary">
          {$_(session.courts === 1 ? 'active_courts_one' : 'active_courts_other', { values: { n: session.courts } })} · {session.points + ' pts'} · {gameModeName}{#if session.rounds_total} · {session.rounds_total} rds{/if}
        </div>
        <button
          onclick={() => (showRules = true)}
          aria-label={$_('lobby_rules_button')}
          class="text-text-disabled hover:text-text-secondary transition-colors"
        >
          <Info size={16} />
        </button>
        {#if isAdmin || alreadyJoined}
          <a
            href="/"
            class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-text-disabled transition-colors hover:bg-surface-raised hover:text-text-secondary"
            aria-label="Back to home"
          >×</a>
        {/if}
      </div>
    </nav>

    <!-- Admin config section -->
    {#if isAdmin}
      <div class="rounded-2xl bg-surface-raised px-5 py-4 space-y-5">
        <SectionLabel>{$_('lobby_config_label')}</SectionLabel>

        <!-- Game mode -->
        <div class="space-y-2">
          <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-text-disabled">{$_('create_game_mode_label')}</p>
          <div class="grid grid-cols-2 gap-2">
            {#each (['americano', 'mexicano'] as const) as mode}
              <button
                type="button"
                onclick={() => onModeChange(mode)}
                class="rounded-xl border-2 px-3 py-3 text-left transition-colors {configMode === mode ? 'border-primary bg-primary/10' : 'border-border bg-surface'}"
              >
                <p class="text-sm font-semibold capitalize">{mode}</p>
                <p class="text-[11px] text-text-secondary mt-0.5">{$_(mode === 'americano' ? 'create_americano_hint' : 'create_mexicano_hint')}</p>
              </button>
            {/each}
          </div>
        </div>

        <!-- Courts -->
        <div class="space-y-2">
          <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-text-disabled">{$_('create_courts_label')}</p>
          <PillToggleGroup
            value={configCourts.toString()}
            onValueChange={(v) => v && onCourtsChange(parseInt(v))}
          >
            {#each Array.from({ length: MAX_COURTS }, (_, i) => i + 1) as n}
              <PillToggleItem value={n.toString()} disabled={configMode === 'mexicano' && n === 1}>{n}</PillToggleItem>
            {/each}
          </PillToggleGroup>
        </div>
        </div>

        <!-- Points -->
        <div class="space-y-2">
          <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-text-disabled">{$_('create_points_label')}</p>
          <PillToggleGroup
            value={configPoints.toString()}
            onValueChange={(v) => v && onPointsChange(parseInt(v))}
          >
            {#each [16, 24, 32] as p}
              <PillToggleItem value={p.toString()}>{p}</PillToggleItem>
            {/each}
          </PillToggleGroup>
        </div>

        <!-- Mexicano: rounds stepper -->
        {#if configMode === 'mexicano'}
          <div class="flex items-center justify-between gap-4">
            <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-text-disabled">{$_('lobby_rounds_label')}</p>
            <Stepper bind:value={configRounds} min={1} max={20} onchange={onRoundsChange} />
          </div>
        {/if}

        <!-- Schedule -->
        <div class="space-y-2">
          <button
            type="button"
            onclick={() => commitSchedule(!scheduleEnabled)}
            class="w-full flex items-center justify-between rounded-xl border px-4 py-3 transition-colors {scheduleEnabled ? 'border-primary bg-primary/10' : 'border-border bg-surface'}"
          >
            <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-text-disabled">{$_('create_schedule_label')}</p>
            {#if scheduleEnabled && calendarDate}
              <p class="text-sm font-semibold text-primary">{calendarDate.toDate(getLocalTimeZone()).toLocaleDateString(undefined, { weekday: 'short', month: 'short', day: 'numeric' })}</p>
            {:else}
              <span class="text-xs text-text-disabled">{scheduleEnabled ? '–' : $_('lobby_schedule_tap_to_add')}</span>
            {/if}
          </button>
          {#if scheduleEnabled}
            <div class="rounded-xl bg-surface overflow-hidden">
              <Calendar
                bind:value={calendarDate}
                minValue={today(getLocalTimeZone())}
                weekStartsOn={1}
                onValueChange={() => commitScheduleTime()}
              />
              <div class="px-4 pb-4 space-y-3">
                <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-text-disabled">{$_('create_schedule_time_label')}</p>
                <div class="flex items-center justify-center gap-3">
                  <div class="flex flex-col items-center gap-1">
                    <p class="text-[10px] text-text-disabled uppercase tracking-[0.1em]">{$_('schedule_hour_label')}</p>
                    <Stepper value={timeHour} min={8} max={21} onchange={onHourChange} />
                  </div>
                  <p class="text-xl font-[800] text-text-secondary pb-1">:</p>
                  <div class="flex flex-col items-center gap-1">
                    <p class="text-[10px] text-text-disabled uppercase tracking-[0.1em]">{$_('schedule_minute_label')}</p>
                    <Stepper value={timeMinute} min={0} max={30} step={30} onchange={onMinuteChange} />
                  </div>
                </div>
              </div>
            </div>
          {/if}
        </div>
      </div>
    {/if}

    <!-- Join code + share -->
    <div class="rounded-2xl bg-surface-raised px-5 py-4 space-y-3">
      <SectionLabel>{$_('lobby_join_code')}</SectionLabel>
      <div class="flex gap-2">
        {#each session.id.split('') as char}
          <div class="flex flex-1 items-center justify-center rounded-xl bg-surface py-3 text-2xl font-[700] text-text-primary font-mono">
            {char}
          </div>
        {/each}
      </div>
      <div class="flex items-center gap-2 rounded-xl bg-surface px-3 py-2.5">
        <span class="flex-1 truncate text-xs text-text-secondary">{joinUrl}</span>
        <Button
          onclick={copyLink}
          variant="ghost"
          class="h-auto shrink-0 p-1.5 text-primary hover:bg-transparent hover:text-primary-hover"
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
    {#if isAdmin}
      <div class="space-y-2">
        <SectionLabel>{$_('lobby_invite_label')}</SectionLabel>
        <div class="relative">
          <div class="pointer-events-none absolute inset-y-0 left-3.5 flex items-center">
            <Search size={15} class="text-text-disabled" />
          </div>
          <Input
            bind:value={playerSearch}
            oninput={onPlayerSearchInput}
            placeholder={$_('lobby_add_player_placeholder')}
            maxlength={32}
            class="w-full rounded-2xl border-0 bg-surface-raised py-3 pl-9 pr-4 text-sm"
          />
        </div>
        {#if playerResults.length > 0}
          <div class="space-y-1.5">
            {#each playerResults as result}
              <div class="flex items-center gap-3 rounded-2xl bg-surface-raised px-4 py-3">
                <Avatar icon={result.avatar_icon} color={result.avatar_color} name={result.display_name} size="sm" ring="ring-2 ring-primary/30" />
                <p class="flex-1 text-sm font-semibold truncate">{result.display_name}</p>
                <button
                  onclick={() => inviteUser(result.id)}
                  class="flex items-center gap-1 rounded-full bg-primary px-3 py-1.5 text-xs font-semibold text-white"
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
            disabled={joining || isFull}
            class="flex w-full items-center gap-3 rounded-2xl border border-dashed border-border px-4 py-3 text-sm text-text-secondary transition-colors hover:border-primary hover:text-primary disabled:opacity-50"
          >
            <UserPlus size={15} class="shrink-0" />
            Add "{playerSearch.trim()}" as guest
          </button>
        {/if}
      </div>
    {/if}

    <!-- Player list -->
    <div class="space-y-2">
      <SectionLabel>
        {$_('lobby_players_label')} ({activePlayers.length})
      </SectionLabel>
      {#if activePlayers.length === 0 && sessionInvites.length === 0}
        <p class="text-sm text-text-disabled">{$_('lobby_waiting_players')}</p>
      {:else}
        <div class="rounded-2xl bg-surface-raised divide-y divide-border">
          {#each activePlayers as player (player.id)}
            <div class="flex items-center gap-3 px-4 py-3">
              <Avatar icon={player.avatar_icon} color={player.avatar_color} name={player.name} size="sm" ring="ring-2 ring-primary/30" />
              <span class="text-sm font-medium">{player.name}</span>
              <div class="ml-auto flex items-center gap-1.5">
                {#if player.id === session.creator_player_id}
                  <Crown size={13} class="text-primary" />
                {/if}
                {#if player.id === myPlayerId}
                  <span class="text-xs text-text-disabled">{$_('lobby_you')}</span>
                {/if}
                {#if isAdmin && player.id !== session.creator_player_id && player.id !== myPlayerId}
                  <button
                    onclick={() => removePlayer(player.id)}
                    class="ml-1 flex h-5 w-5 items-center justify-center rounded-full text-text-disabled transition-colors hover:bg-destructive/10 hover:text-destructive"
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
              <Avatar name={invite.to_display_name ?? '?'} size="sm" ring="ring-2 ring-primary/30" />
              <span class="flex-1 text-sm font-medium text-text-secondary truncate">{invite.to_display_name}</span>
              <div class="ml-auto flex items-center gap-1 text-text-disabled">
                <Clock size={11} />
                <span class="text-xs">Invited</span>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>

    <!-- Join form (non-admin who hasn't joined) -->
    {#if !isAdmin && !alreadyJoined}
      <div class="space-y-2">
        {#if isFull}
          <div class="flex items-start gap-2 rounded-2xl bg-red-50 px-4 py-3 text-sm text-red-900">
            <Info size={16} class="shrink-0 mt-0.5" />
            <p>This session has reached its maximum player limit.</p>
          </div>
        {/if}
        <form onsubmit={(e) => { e.preventDefault(); join(); }} class="flex gap-2">
          <Input
            bind:value={joinName}
            placeholder={$_('lobby_join_placeholder')}
            maxlength={32}
            class="flex-1 rounded-2xl border-0 bg-surface-raised px-4 py-3 text-sm"
            disabled={isFull}
          />
          <Button
            type="submit"
            disabled={joining || !joinName.trim() || isFull}
            class="h-auto rounded-2xl bg-primary px-4 text-sm font-semibold text-white"
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
          class="h-auto w-full rounded-2xl bg-primary px-4 py-4 text-[15px] font-semibold text-white hover:bg-primary-hover"
        >
          {starting ? $_('lobby_start_loading') : $_('lobby_start_button')}
        </Button>
        {#if !canStart}
          {#if session.validation_errors && session.validation_errors.length > 0}
            {#each session.validation_errors as err}
              <p class="text-center text-xs text-text-disabled">
                {translateApiError(err.code, err.params)}
              </p>
            {/each}
          {:else}
            <p class="text-center text-xs text-text-disabled">
              {#if isMexicano}
                {$_('lobby_mexicano_exact_players', { values: { n: requiredPlayers, current: activePlayers.length } })}
              {:else}
                {$_('lobby_need_players', { values: { n: requiredPlayers } })}
              {/if}
            </p>
          {/if}
        {/if}
        <button
          onclick={() => (showCancelDialog = true)}
          disabled={cancelling}
          class="h-auto w-full rounded-2xl border border-border px-4 py-3.5 text-sm font-semibold text-text-secondary transition-colors hover:border-destructive hover:text-destructive disabled:opacity-40"
        >
          {cancelling ? $_('lobby_cancelling') : $_('lobby_cancel')}
        </button>
        {#if isDev}
          <div class="flex justify-center pt-2">
            <button
              onclick={seedPlayers}
              disabled={seeding}
              class="rounded-full border border-dashed border-border px-4 py-1.5 text-xs text-text-disabled transition-colors hover:border-text-disabled hover:text-text-secondary disabled:opacity-40"
            >
              {seeding ? $_('lobby_dev_seeding') : $_('lobby_dev_seed')}
            </button>
          </div>
        {/if}
      </div>
    {:else}
      <div class="rounded-2xl bg-surface-raised px-4 py-3 text-center">
        <p class="text-sm text-text-secondary">{$_('lobby_waiting_admin')}</p>
      </div>
    {/if}
  </main>
{/if}

<Dialog.Root bind:open={showRules}>
  <Dialog.Content class="w-full max-w-sm">
    <Dialog.Header>
      <Dialog.Title>{gameModeName}</Dialog.Title>
    </Dialog.Header>
    <div class="space-y-2">
      <ul class="space-y-2">
        {#each $_(`rules_${session.game_mode}`).split('\n') as line}
          {#if line.trim()}
            <li class="flex gap-2 text-sm text-text-secondary">
              <span class="mt-0.5 shrink-0 text-primary">·</span>
              <span>{line.trim()}</span>
            </li>
          {/if}
        {/each}
      </ul>
    </div>
    <Dialog.Footer>
      <Dialog.Close class="w-full rounded-2xl border border-border px-4 py-3.5 text-sm font-semibold text-text-secondary transition-colors hover:bg-surface-raised">
        {$_('leaderboard_close')}
      </Dialog.Close>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

<ConfirmDialog
  open={showCancelDialog}
  title={$_('cancel_dialog_title')}
  description={$_('cancel_dialog_desc')}
  confirmLabel={$_('cancel_dialog_confirm')}
  cancelLabel={$_('cancel_dialog_cancel')}
  destructive
  onconfirm={cancel}
  oncancel={() => (showCancelDialog = false)}
/>
