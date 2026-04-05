<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/state';
  import { auth } from '$lib/auth.svelte';
  import { api } from '$lib/api/client';
  import { _ } from 'svelte-i18n';
  import { initials } from '$lib/utils';
  import { CalendarDays, Radio, ChevronDown } from 'lucide-svelte';
  import Footer from '$lib/components/Footer.svelte';
  import { fly, slide } from 'svelte/transition';
  import { subscribeToPush, unsubscribeFromPush } from '$lib/push';

  let stats = $state<App.CareerStats | null>(null);
  let tournaments = $state<App.TournamentEntry[]>([]);
  let upcoming = $state<App.UpcomingEntry[]>([]);
  let loading = $state(true);
  let showDeleteConfirm = $state(false);
  let deleting = $state(false);
  let showNotFoundBanner = $state(false);
  let joinChars = $state(['', '', '', '']);
  let joinInputs = $state<HTMLInputElement[]>([]);

  let showStats = $state(true);
  let showUpcoming = $state(false);
  let showHistory = $state(false);
  let showPreferences = $state(true);

  let pushSupported = $state(false);
  let pushEnabled = $state(false);
  let pushToggling = $state(false);
  let pushError = $state('');

  // Install prompt
  let isStandalone = $state(false);
  let isIOS = $state(false);
  let deferredInstallPrompt = $state<any>(null);
  let installDismissed = $state(false);

  function joinByCode() {
    const code = joinChars.join('').trim();
    if (code.length === 4) goto(`/s/${code}`);
  }

  function onJoinInput(i: number, e: Event) {
    const val = (e.currentTarget as HTMLInputElement).value.toUpperCase().replace(/[^A-Z0-9]/g, '');
    joinChars[i] = val.slice(-1);
    if (val && i < 3) joinInputs[i + 1]?.focus();
    if (joinChars.every(c => c)) joinByCode();
  }

  function onJoinKeydown(i: number, e: KeyboardEvent) {
    if (e.key === 'Backspace' && !joinChars[i] && i > 0) {
      joinChars[i - 1] = '';
      joinInputs[i - 1]?.focus();
    }
  }

  function onJoinPaste(e: ClipboardEvent) {
    const text = e.clipboardData?.getData('text')?.toUpperCase().replace(/[^A-Z0-9]/g, '') ?? '';
    if (text.length >= 4) {
      e.preventDefault();
      joinChars = text.slice(0, 4).split('');
      joinInputs[3]?.focus();
      joinByCode();
    }
  }

  onMount(async () => {
    if (!auth.token) { goto('/auth'); return; }
    if (page.url.searchParams.get('notfound') === '1') {
      showNotFoundBanner = true;
      setTimeout(() => { showNotFoundBanner = false; }, 4000);
    }
    try {
      const [profileRes, historyRes] = await Promise.all([
        api.auth.profile(auth.token),
        api.auth.history(auth.token),
      ]);
      stats = profileRes.stats;
      tournaments = historyRes.tournaments;
      upcoming = historyRes.upcoming;
      showUpcoming = upcoming.length > 0;
      showHistory = tournaments.length > 0;
    } finally {
      loading = false;
    }

    // Install detection — runs immediately, no SW needed
    isStandalone = window.matchMedia('(display-mode: standalone)').matches
      || (navigator as any).standalone === true;
    isIOS = /iphone|ipad|ipod/i.test(navigator.userAgent);
    const isMobile = /iphone|ipad|ipod|android/i.test(navigator.userAgent);
    installDismissed = !isMobile || localStorage.getItem('install_dismissed') === '1';

    window.addEventListener('beforeinstallprompt', (e: any) => {
      e.preventDefault();
      deferredInstallPrompt = e;
    });

    // Push subscription check — requires active SW, may hang in dev
    if ('serviceWorker' in navigator && 'PushManager' in window) {
      pushSupported = true;
      try {
        const swReady = Promise.race([
          navigator.serviceWorker.ready,
          new Promise<never>((_, reject) => setTimeout(() => reject(new Error('timeout')), 3000)),
        ]);
        const reg = await swReady as ServiceWorkerRegistration;
        const sub = await reg.pushManager.getSubscription();
        pushEnabled = !!sub && Notification.permission === 'granted';
      } catch {
        // SW not ready (dev mode or not yet installed) — push state unknown
      }
    }
  });

  const winRate = $derived(
    stats && stats.games_played > 0
      ? Math.round((stats.wins / stats.games_played) * 100)
      : 0
  );

  const memberSince = $derived(
    auth.user
      ? new Date(auth.user.created_at).toLocaleDateString(undefined, { month: 'long', year: 'numeric' })
      : ''
  );

  async function checkPushState() {
    const reg = await navigator.serviceWorker.ready;
    const sub = await reg.pushManager.getSubscription();
    pushEnabled = !!sub && Notification.permission === 'granted';
  }

  async function togglePush() {
    if (!auth.token) return;
    pushToggling = true;
    pushError = '';
    try {
      if (pushEnabled) {
        await unsubscribeFromPush(auth.token);
      } else {
        await subscribeToPush(auth.token);
      }
      await checkPushState();
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'unknown';
      pushError = msg === 'notifications_blocked' || msg === 'sw_timeout' ? msg : msg;
    } finally {
      pushToggling = false;
    }
  }

  async function deleteAccount() {
    if (!auth.token) return;
    deleting = true;
    try {
      await api.auth.deleteAccount(auth.token);
      await auth.logout();
      goto('/?deleted=1');
    } finally {
      deleting = false;
    }
  }

  function formatDate(iso: string) {
    return new Date(iso).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
  }

  function ordinal(n: number) {
    const s = ['th', 'st', 'nd', 'rd'];
    const v = n % 100;
    return n + (s[(v - 20) % 10] ?? s[v] ?? s[0]);
  }
</script>

{#if showNotFoundBanner}
  <div transition:fly={{ y: -48, duration: 400 }} class="fixed inset-x-0 top-0 z-50 flex items-center justify-center bg-[var(--destructive)] px-4 py-3 text-sm font-semibold text-white">
    {$_('home_session_not_found')}
  </div>
{/if}

<main class="mx-auto max-w-[480px] px-6 pb-24 pt-8 space-y-8">

  <!-- Header -->
  <div class="flex items-center gap-4">
    <div class="flex h-16 w-16 shrink-0 items-center justify-center rounded-full bg-[var(--primary)] text-2xl font-[800] text-white">
      {auth.user ? initials(auth.user.display_name) : '?'}
    </div>
    <div class="min-w-0">
      <h1 class="text-2xl font-[800] truncate">{auth.user?.display_name}</h1>
      {#if memberSince}
        <p class="text-sm text-[var(--text-secondary)]">Member since {memberSince}</p>
      {/if}
    </div>
  </div>

  {#if loading}
    <div class="flex justify-center py-12">
      <div class="h-7 w-7 animate-spin rounded-full border-2 border-[var(--border)] border-t-[var(--primary)]"></div>
    </div>
  {:else}

    <!-- Preferences -->
    <div class="space-y-3">
      <button
        onclick={() => showPreferences = !showPreferences}
        class="flex w-full items-center justify-between"
      >
        <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('pref_section')}</p>
        <ChevronDown size={14} class="text-[var(--text-disabled)] transition-transform duration-200 {showPreferences ? 'rotate-180' : ''}" />
      </button>

      {#if showPreferences}
        <div transition:slide={{ duration: 200 }} class="space-y-2">
          {#if pushSupported}
            <div class="flex items-center gap-4 rounded-2xl bg-[var(--surface-raised)] px-4 py-3.5">
              <div class="flex-1">
                <p class="text-sm font-semibold">{$_('pref_notifications_title')}</p>
                <p class="text-xs text-[var(--text-secondary)]">{$_('pref_notifications_desc')}</p>
              </div>
              <button
                onclick={togglePush}
                disabled={pushToggling}
                class="relative h-7 w-12 shrink-0 rounded-full transition-colors duration-200 disabled:opacity-50
                  {pushEnabled ? 'bg-[var(--primary)]' : 'bg-[var(--border-strong)]'}"
                aria-label="Toggle notifications"
              >
                <span class="absolute top-0.5 left-0.5 h-6 w-6 rounded-full bg-white shadow transition-transform duration-200
                  {pushEnabled ? 'translate-x-5' : 'translate-x-0'}"></span>
              </button>
            </div>
            {#if pushError}
              <p class="px-1 text-xs text-[var(--destructive)]">
                {pushError === 'notifications_blocked'
                  ? $_('pref_notifications_blocked', { values: { app: 'NotTennis' } })
                  : pushError === 'sw_timeout'
                  ? $_('pref_notifications_sw_timeout')
                  : pushError}
              </p>
            {/if}
          {/if}

          <!-- Install prompt — shown regardless of pushSupported -->
          {#if !isStandalone && !installDismissed && (isIOS || deferredInstallPrompt)}
            <div class="flex items-start gap-3 rounded-2xl bg-[var(--primary-muted)] px-4 py-3.5">
              <div class="flex-1 space-y-0.5">
                {#if isIOS}
                  <p class="text-sm font-semibold text-[var(--primary)]">{$_('pwa_ios_title')}</p>
                  <p class="text-xs text-[var(--text-secondary)]">
                    {$_('pwa_ios_tap')} <span class="font-semibold">{$_('pwa_ios_share')}</span> {$_('pwa_ios_then')} <span class="font-semibold">{$_('pwa_ios_add')}</span> {$_('pwa_ios_suffix')}
                  </p>
                {:else if deferredInstallPrompt}
                  <p class="text-sm font-semibold text-[var(--primary)]">{$_('pwa_android_title')}</p>
                  <p class="text-xs text-[var(--text-secondary)]">{$_('pwa_android_desc')}</p>
                {/if}
              </div>
              <div class="flex shrink-0 items-center gap-2">
                {#if deferredInstallPrompt && !isIOS}
                  <button
                    onclick={async () => { deferredInstallPrompt.prompt(); const { outcome } = await deferredInstallPrompt.userChoice; if (outcome === 'accepted') { deferredInstallPrompt = null; isStandalone = true; } }}
                    class="rounded-full bg-[var(--primary)] px-3 py-1 text-xs font-semibold text-white"
                  >{$_('pwa_install_btn')}</button>
                {/if}
                <button
                  onclick={() => { installDismissed = true; localStorage.setItem('install_dismissed', '1'); }}
                  class="text-[var(--text-disabled)] hover:text-[var(--text-secondary)] text-lg leading-none"
                  aria-label="Dismiss"
                >✕</button>
              </div>
            </div>
          {/if}
        </div>
      {/if}
    </div>

    {#if stats}

      <!-- Stats -->
      <div class="space-y-3">
        <button
          onclick={() => showStats = !showStats}
          class="flex w-full items-center justify-between"
        >
          <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('profile_stats_section')}</p>
          <ChevronDown size={14} class="text-[var(--text-disabled)] transition-transform duration-200 {showStats ? 'rotate-180' : ''}" />
        </button>

        {#if showStats}
          <div transition:slide={{ duration: 200 }} class="grid grid-cols-2 gap-3">
            <div class="rounded-2xl bg-[var(--surface-raised)] px-5 py-5 flex flex-col items-center justify-center gap-1">
              <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-[var(--text-disabled)]">{$_('profile_tournaments')}</p>
              <p class="text-3xl font-[800]">{stats.tournaments}</p>
            </div>
            <div class="rounded-2xl bg-[var(--surface-raised)] px-5 py-5 flex flex-col items-center justify-center gap-1">
              <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-[var(--text-disabled)]">{$_('profile_win_rate')}</p>
              <p class="text-3xl font-[800]">{winRate}<span class="text-lg text-[var(--text-secondary)]">%</span></p>
            </div>
            <div class="rounded-2xl bg-[var(--surface-raised)] px-5 py-5 flex flex-col items-center justify-center gap-1">
              <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-[var(--text-disabled)]">{$_('profile_games')}</p>
              <p class="text-3xl font-[800]">{stats.games_played}</p>
            </div>
            <div class="rounded-2xl bg-[var(--surface-raised)] px-5 py-5 flex flex-col items-center justify-center gap-1.5">
              <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-[var(--text-disabled)]">{$_('leaderboard_wl')}</p>
              <div class="flex items-center gap-1.5 text-2xl font-[800] tabular-nums">
                <span class="text-[var(--primary)]">{stats.wins}V</span>
                <span class="text-[var(--text-disabled)] text-base">·</span>
                <span class="text-[var(--text-disabled)]">{stats.draws}U</span>
                <span class="text-[var(--text-disabled)] text-base">·</span>
                <span class="text-[#c0392b]">{stats.losses}T</span>
              </div>
            </div>
          </div>
        {/if}
      </div>

      <!-- Upcoming -->
      <div class="space-y-3">
        <button
          onclick={() => showUpcoming = !showUpcoming}
          class="flex w-full items-center justify-between"
        >
          <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('profile_upcoming_label')}</p>
          <ChevronDown size={14} class="text-[var(--text-disabled)] transition-transform duration-200 {showUpcoming ? 'rotate-180' : ''}" />
        </button>

        {#if showUpcoming}
          <div transition:slide={{ duration: 200 }} class="space-y-2">
            {#if upcoming.length === 0}
              <p class="text-sm text-[var(--text-disabled)] py-1">{$_('profile_upcoming_empty')}</p>
            {:else}
              {#each upcoming as t}
                <a href="/s/{t.session_id}" class="flex items-center gap-4 rounded-2xl bg-[var(--surface-raised)] px-4 py-3.5 transition-colors hover:bg-[var(--border)]">
                  <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full {t.status === 'active' ? 'bg-emerald-500/15 text-emerald-500' : 'bg-[var(--primary-muted)] text-[var(--primary)]'}">
                    {#if t.status === 'active'}<Radio size={18} />{:else}<CalendarDays size={18} />{/if}
                  </div>
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center gap-2">
                      <p class="truncate font-semibold text-sm">{t.name}</p>
                      {#if t.status === 'active'}
                        <span class="shrink-0 rounded-full bg-emerald-500/15 px-2 py-0.5 text-[10px] font-bold uppercase tracking-wide text-emerald-500">Live</span>
                      {/if}
                    </div>
                    <p class="text-xs text-[var(--text-secondary)]">{t.player_count} {$_('profile_upcoming_players')}</p>
                  </div>
                  <span class="text-sm text-[var(--text-secondary)]">→</span>
                </a>
              {/each}
            {/if}
          </div>
        {/if}
      </div>

      <!-- Tournament history -->
      <div class="space-y-3">
        <button
          onclick={() => showHistory = !showHistory}
          class="flex w-full items-center justify-between"
        >
          <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('profile_history_label')}</p>
          <ChevronDown size={14} class="text-[var(--text-disabled)] transition-transform duration-200 {showHistory ? 'rotate-180' : ''}" />
        </button>

        {#if showHistory}
          <div transition:slide={{ duration: 200 }} class="space-y-2">
            {#if tournaments.length === 0}
              <p class="text-sm text-[var(--text-disabled)] py-2">{$_('profile_history_empty')}</p>
            {:else}
              {#each tournaments as t}
                <a href="/s/{t.session_id}" class="flex items-center gap-4 rounded-2xl bg-[var(--surface-raised)] px-4 py-3.5 transition-colors hover:bg-[var(--border)]">
                  <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full text-xs font-[800]
                    {t.rank === 1 ? 'bg-[var(--primary)] text-white' : 'bg-[var(--border)] text-[var(--text-secondary)]'}">
                    {ordinal(t.rank)}
                  </div>
                  <div class="flex-1 min-w-0">
                    <p class="truncate font-semibold text-sm">{t.name}</p>
                    <p class="text-xs text-[var(--text-secondary)]">{formatDate(t.played_at)} · {t.points} pts</p>
                  </div>
                  <span class="text-sm text-[var(--text-secondary)]">→</span>
                </a>
              {/each}
            {/if}
          </div>
        {/if}
      </div>

    {/if}

    <!-- Actions -->
    <div class="space-y-3">
      <a href="/?create=1" class="block w-full rounded-2xl bg-[var(--primary)] px-4 py-4 text-center text-[15px] font-semibold text-white">
        {$_('profile_new_tournament')}
      </a>

      <div class="flex items-center gap-3">
        <div class="h-px flex-1 bg-[var(--border)]"></div>
        <span class="text-xs text-[var(--text-disabled)]">{$_('home_join_code_divider')}</span>
        <div class="h-px flex-1 bg-[var(--border)]"></div>
      </div>

      <div class="flex justify-center gap-2" onpaste={onJoinPaste}>
        {#each joinChars as _, i}
          <input
            bind:this={joinInputs[i]}
            value={joinChars[i]}
            oninput={(e) => onJoinInput(i, e)}
            onkeydown={(e) => onJoinKeydown(i, e)}
            maxlength={1}
            autocomplete="off"
            autocorrect="off"
            autocapitalize="characters"
            spellcheck={false}
            class="w-12 rounded-xl bg-[var(--surface-raised)] py-2.5 text-center text-lg font-[700] font-mono text-[var(--text-primary)] outline-none focus:ring-2 focus:ring-[var(--primary)] transition-shadow"
          />
        {/each}
      </div>

      <div class="pt-2 space-y-2">
        <button
          onclick={() => auth.logout().then(() => goto('/'))}
          class="w-full rounded-2xl border border-[var(--border)] px-4 py-3.5 text-sm font-semibold text-[var(--text-secondary)] transition-colors hover:border-[var(--destructive)] hover:text-[var(--destructive)]"
        >
          {$_('auth_sign_out')}
        </button>
        <button
          onclick={() => (showDeleteConfirm = true)}
          class="w-full px-4 py-2 text-sm text-[var(--text-disabled)] hover:text-[var(--destructive)] transition-colors"
        >
          {$_('profile_delete_account')}
        </button>
      </div>
    </div>

  {/if}

  <Footer />

</main>

{#if showDeleteConfirm}
  <div class="fixed inset-0 z-50 flex items-end justify-center bg-black/40 p-4 sm:items-center">
    <div class="w-full max-w-sm rounded-3xl bg-[var(--surface)] p-6 space-y-4 shadow-xl">
      <div class="space-y-1">
        <h2 class="text-lg font-[800]">{$_('profile_delete_title')}</h2>
        <p class="text-sm text-[var(--text-secondary)]">{$_('profile_delete_desc')}</p>
      </div>
      <div class="space-y-2 pt-1">
        <button onclick={deleteAccount} disabled={deleting} class="w-full rounded-2xl bg-[var(--destructive)] px-4 py-3.5 text-sm font-semibold text-white disabled:opacity-60">
          {deleting ? $_('profile_delete_loading') : $_('profile_delete_confirm')}
        </button>
        <button onclick={() => (showDeleteConfirm = false)} class="w-full rounded-2xl border border-[var(--border)] px-4 py-3.5 text-sm font-semibold text-[var(--text-secondary)]">
          {$_('profile_delete_cancel')}
        </button>
      </div>
    </div>
  </div>
{/if}
