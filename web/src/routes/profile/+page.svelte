<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/state';
  import { auth } from '$lib/auth.svelte';
  import { api } from '$lib/api/client';
  import { _ } from 'svelte-i18n';
  import { CalendarDays, Radio, ChevronDown, UserPlus, X, Search, Check } from 'lucide-svelte';
  import Footer from '$lib/components/Footer.svelte';
  import CreateDrawer from '$lib/components/CreateDrawer.svelte';
  import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
  import Avatar from '$lib/components/ui/Avatar.svelte';
  import * as Collapsible from '$lib/components/ui/collapsible';
  import * as Dialog from '$lib/components/ui/dialog';
  import { toast } from 'svelte-sonner';
  import { translateApiError } from '$lib/i18n/errors';
  import { subscribeToPush, unsubscribeFromPush } from '$lib/push';

  const AVATAR_ICONS = [
    'Zap', 'Star', 'Flame', 'Shield', 'Crown', 'Trophy', 'Target', 'Rocket',
    'Ghost', 'Cat', 'Dog', 'Bird', 'Leaf', 'Sun', 'Moon', 'Snowflake',
    'Mountain', 'Waves', 'Music', 'Heart', 'Smile', 'Fish', 'Swords', 'Dumbbell',
    'Bike', 'Footprints',
  ];

  let pickerIcon = $state(auth.user?.avatar_icon ?? '');
  let savingAvatar = $state(false);
  let showAvatarPicker = $state(false);

  $effect(() => {
    document.body.style.overflow = showAvatarPicker ? 'hidden' : '';
    return () => { document.body.style.overflow = ''; };
  });

  let showCreateDrawer = $state(false);

  let stats = $state<App.AmericanoCareerStats | null>(null);
  let tennisStats = $state<App.TennisCareerStats | null>(null);
  let tournaments = $state<App.TournamentEntry[]>([]);
  let upcoming = $state<App.UpcomingEntry[]>([]);
  let loading = $state(true);
  let showDeleteConfirm = $state(false);
  let deleting = $state(false);
  let joinChars = $state(['', '', '', '']);
  let joinInputs = $state<HTMLInputElement[]>([]);

  let showStats = $state(true);
  let showContacts = $state(false);
  let showUpcoming = $state(false);
  let showHistory = $state(false);
  let showPreferences = $state(true);

  let invites = $state<App.Invite[]>([]);
  let contacts = $state<App.Contact[]>([]);
  let contactSearch = $state('');
  let searchResults = $state<App.UserSearchResult[]>([]);
  let searchLoading = $state(false);
  let searchDebounce: ReturnType<typeof setTimeout>;
  let contactToDelete = $state<App.Contact | null>(null);
  let showContactDeleteConfirm = $state(false);

  function onContactSearchInput() {
    clearTimeout(searchDebounce);
    if (contactSearch.length < 2) { searchResults = []; searchLoading = false; return; }
    searchLoading = true;
    searchDebounce = setTimeout(async () => {
      try {
        searchResults = await api.contacts.search(auth.token!, contactSearch);
      } finally {
        searchLoading = false;
      }
    }, 300);
  }

  async function acceptInvite(inviteID: string, sessionID: string) {
    if (!auth.token) return;
    await api.invites.accept(inviteID, auth.token);
    invites = invites.filter(i => i.id !== inviteID);
    window.location.href = `/s/${sessionID}`;
  }

  async function declineInvite(inviteID: string) {
    if (!auth.token) return;
    await api.invites.decline(inviteID, auth.token);
    invites = invites.filter(i => i.id !== inviteID);
  }

  async function addContact(userID: string) {
    await api.contacts.add(auth.token!, userID);
    contactSearch = '';
    searchResults = [];
    contacts = await api.contacts.list(auth.token!);
  }

  async function removeContact(userID: string) {
    await api.contacts.remove(auth.token!, userID);
    contacts = contacts.filter(c => c.user_id !== userID);
    searchResults = searchResults.map(r => r.id === userID ? { ...r, is_contact: false } : r);
  }

  function handleDeleteContact(contact: App.Contact) {
    contactToDelete = contact;
    showContactDeleteConfirm = true;
  }

  async function confirmDeleteContact() {
    if (contactToDelete) {
      await removeContact(contactToDelete.user_id);
      showContactDeleteConfirm = false;
    }
  }

  let pushSupported = $state(false);
  let pushEnabled = $state(false);
  let pushToggling = $state(false);

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

  async function load() {
    if (!auth.token) return;
    try {
      const [profileRes, historyRes, contactsRes, invitesRes] = await Promise.all([
        api.auth.profile(auth.token),
        api.auth.history(auth.token),
        api.contacts.list(auth.token),
        api.invites.list(auth.token),
      ]);
      stats = profileRes.stats;
      tennisStats = profileRes.tennis_stats;
      tournaments = historyRes.tournaments;
      upcoming = (historyRes.upcoming ?? []).sort((a, b) => {
        if (!a.scheduled_at && !b.scheduled_at) return 0;
        if (!a.scheduled_at) return 1;
        if (!b.scheduled_at) return -1;
        return new Date(a.scheduled_at).getTime() - new Date(b.scheduled_at).getTime();
      });
      contacts = contactsRes;
      invites = invitesRes;
      showUpcoming = upcoming.length > 0;
      showHistory = tournaments.length > 0;
    } finally {
      loading = false;
    }
  }

  onMount(async () => {
    if (!auth.token) { goto('/auth'); return; }
    if (page.url.searchParams.get('notfound') === '1') {
      toast.error(translateApiError('session_not_found'));
    }
    if (page.url.searchParams.get('create') === '1') {
      showCreateDrawer = true;
    }
    await load();

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

  const tennisWinRate = $derived(
    tennisStats && (tennisStats.wins + tennisStats.losses) > 0
      ? Math.round((tennisStats.wins / (tennisStats.wins + tennisStats.losses)) * 100)
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
    try {
      if (pushEnabled) {
        await unsubscribeFromPush(auth.token);
      } else {
        await subscribeToPush(auth.token);
      }
      await checkPushState();
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'unknown';
      const label = msg === 'notifications_blocked'
        ? $_('pref_notifications_blocked', { values: { app: 'OpenPadel' } })
        : msg === 'sw_timeout'
        ? $_('pref_notifications_sw_timeout')
        : msg;
      toast.error(label);
    } finally {
      pushToggling = false;
    }
  }

  async function saveAvatar() {
    if (!auth.token || !auth.user) return;
    savingAvatar = true;
    try {
      const updated = await api.auth.updateProfile(auth.token, auth.user.display_name, pickerIcon, 'forest');
      auth.updateUser(updated);
    } finally {
      savingAvatar = false;
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



<main class="mx-auto max-w-[480px] px-6 pb-10 pt-safe-page space-y-8">

  <!-- Header -->
  <div class="flex items-center gap-4">
    <button onclick={() => showAvatarPicker = true} class="relative shrink-0 group" aria-label="Edit avatar">
      <Avatar icon={auth.user?.avatar_icon ?? ''} color={auth.user?.avatar_color ?? 'forest'} name={auth.user?.display_name ?? ''} size="lg" />
      <span class="absolute inset-0 flex items-center justify-center rounded-full bg-black/30 opacity-0 group-hover:opacity-100 transition-opacity text-white text-xs font-semibold">Edit</span>
    </button>
    <div class="min-w-0">
      <h1 class="text-2xl font-[800] truncate">{auth.user?.display_name}</h1>
      {#if memberSince}
        <p class="text-sm text-text-secondary">Member since {memberSince}</p>
      {/if}
    </div>
  </div>

  <!-- Pending invites -->
  {#if invites.length > 0}
    <div class="space-y-2">
      <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-text-secondary">Invites</p>
      {#each invites as invite}
        <div class="flex items-center gap-3 rounded-2xl bg-surface-raised px-4 py-3.5">
          <div class="flex-1 min-w-0">
            <p class="truncate text-sm font-semibold">{invite.session_name}</p>
            <p class="text-xs text-text-secondary">From {invite.from_display_name}</p>
          </div>
          <button
            onclick={() => acceptInvite(invite.id, invite.session_id)}
            class="flex items-center gap-1 rounded-full bg-primary px-3 py-1.5 text-xs font-semibold text-white"
          >
            <Check size={12} /> Accept
          </button>
          <button
            onclick={() => declineInvite(invite.id)}
            class="flex items-center justify-center rounded-full bg-surface p-1.5 text-text-disabled hover:text-destructive transition-colors border border-border"
            aria-label="Decline"
          >
            <X size={14} />
          </button>
        </div>
      {/each}
    </div>
  {/if}

  <!-- New tournament + join code -->
  <div class="space-y-3">
    <button onclick={() => showCreateDrawer = true} class="block w-full rounded-2xl bg-primary px-4 py-4 text-center text-[15px] font-semibold text-white">
      {$_('profile_new_tournament')}
    </button>
    <div class="flex items-center gap-3">
      <div class="h-px flex-1 bg-border"></div>
      <span class="text-xs text-text-disabled">{$_('home_join_code_divider')}</span>
      <div class="h-px flex-1 bg-border"></div>
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
          class="w-12 rounded-xl bg-surface-raised py-2.5 text-center text-lg font-[700] font-mono text-text-primary outline-none focus:ring-2 focus:ring-primary transition-shadow"
        />
      {/each}
    </div>
  </div>

  {#if loading}
    <div class="flex justify-center py-12">
      <div class="h-7 w-7 animate-spin rounded-full border-2 border-border border-t-primary"></div>
    </div>
  {:else}

    <!-- Preferences -->
    <Collapsible.Root bind:open={showPreferences} class="space-y-3">
      <Collapsible.Trigger class="flex w-full items-center justify-between">
        <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-text-secondary">{$_('pref_section')}</p>
        <ChevronDown size={14} class="text-text-disabled transition-transform duration-200 data-[state=open]:rotate-180" />
      </Collapsible.Trigger>

      <Collapsible.Content class="space-y-2">
          {#if pushSupported}
            <div class="flex items-center gap-4 rounded-2xl bg-surface-raised px-4 py-3.5">
              <div class="flex-1">
                <p class="text-sm font-semibold">{$_('pref_notifications_title')}</p>
                <p class="text-xs text-text-secondary">{$_('pref_notifications_desc')}</p>
              </div>
              <button
                onclick={togglePush}
                disabled={pushToggling}
                class="relative h-7 w-12 shrink-0 rounded-full transition-colors duration-200 disabled:opacity-50
                  {pushEnabled ? 'bg-primary' : 'bg-border-strong'}"
                aria-label="Toggle notifications"
              >
                <span class="absolute top-0.5 left-0.5 h-6 w-6 rounded-full bg-white shadow transition-transform duration-200
                  {pushEnabled ? 'translate-x-5' : 'translate-x-0'}"></span>
              </button>
            </div>

          {/if}

          <!-- Install prompt — shown regardless of pushSupported -->
          {#if !isStandalone && !installDismissed && (isIOS || deferredInstallPrompt)}
            <div class="flex items-start gap-3 rounded-2xl bg-primary-muted px-4 py-3.5">
              <div class="flex-1 space-y-0.5">
                {#if isIOS}
                  <p class="text-sm font-semibold text-primary">{$_('pwa_ios_title')}</p>
                  <p class="text-xs text-text-secondary">
                    {$_('pwa_ios_tap')} <span class="font-semibold">{$_('pwa_ios_share')}</span> {$_('pwa_ios_then')} <span class="font-semibold">{$_('pwa_ios_add')}</span> {$_('pwa_ios_suffix')}
                  </p>
                {:else if deferredInstallPrompt}
                  <p class="text-sm font-semibold text-primary">{$_('pwa_android_title')}</p>
                  <p class="text-xs text-text-secondary">{$_('pwa_android_desc')}</p>
                {/if}
              </div>
              <div class="flex shrink-0 items-center gap-2">
                {#if deferredInstallPrompt && !isIOS}
                  <button
                    onclick={async () => { deferredInstallPrompt.prompt(); const { outcome } = await deferredInstallPrompt.userChoice; if (outcome === 'accepted') { deferredInstallPrompt = null; isStandalone = true; } }}
                    class="rounded-full bg-primary px-3 py-1 text-xs font-semibold text-white"
                  >{$_('pwa_install_btn')}</button>
                {/if}
                <button
                  onclick={() => { installDismissed = true; localStorage.setItem('install_dismissed', '1'); }}
                  class="text-text-disabled hover:text-text-secondary text-lg leading-none"
                  aria-label="Dismiss"
                >✕</button>
              </div>
            </div>
          {/if}
      </Collapsible.Content>
    </Collapsible.Root>

    {#if stats}

      <!-- Americano stats -->
      <Collapsible.Root bind:open={showStats} class="space-y-3">
        <Collapsible.Trigger class="flex w-full items-center justify-between">
          <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-text-secondary">Americano</p>
          <ChevronDown size={14} class="text-text-disabled transition-transform duration-200 data-[state=open]:rotate-180" />
        </Collapsible.Trigger>

        <Collapsible.Content class="grid grid-cols-2 gap-3">
            <div class="rounded-2xl bg-surface-raised px-5 py-5 flex flex-col items-center gap-1.5">
              <p class="text-3xl font-[800] leading-none">{stats.tournaments}</p>
              <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-text-disabled">{$_('profile_tournaments')}</p>
            </div>
            <div class="rounded-2xl bg-surface-raised px-5 py-5 flex flex-col items-center gap-1.5">
              <p class="text-3xl font-[800] leading-none">{winRate}</p>
              <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-text-disabled">{$_('profile_win_rate')} %</p>
            </div>
            <div class="rounded-2xl bg-surface-raised px-5 py-5 flex flex-col items-center gap-1.5">
              <p class="text-3xl font-[800] leading-none">{stats.games_played}</p>
              <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-text-disabled">{$_('profile_games')}</p>
            </div>
            <div class="rounded-2xl bg-surface-raised px-5 py-5 flex flex-col items-center gap-1.5">
              <div class="flex items-baseline gap-1 leading-none font-[800] tabular-nums">
                <span class="text-2xl text-primary">{stats.wins}V</span>
                <span class="text-base text-text-disabled">·</span>
                <span class="text-2xl text-text-disabled">{stats.draws}U</span>
                <span class="text-base text-text-disabled">·</span>
                <span class="text-2xl text-[#c0392b]">{stats.losses}T</span>
              </div>
              <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-text-disabled">{$_('leaderboard_wl')}</p>
            </div>
        </Collapsible.Content>
      </Collapsible.Root>

      <!-- Tennis (2v2) stats -->
      {#if tennisStats}
        <div class="space-y-3">
          <div class="flex w-full items-center justify-between">
            <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-text-secondary">{$_('create_mode_tennis')}</p>
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div class="rounded-2xl bg-surface-raised px-5 py-5 flex flex-col items-center gap-1.5">
              <p class="text-3xl font-[800] leading-none">{tennisStats.tournaments}</p>
              <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-text-disabled">{$_('profile_tournaments')}</p>
            </div>
            <div class="rounded-2xl bg-surface-raised px-5 py-5 flex flex-col items-center gap-1.5">
              <p class="text-3xl font-[800] leading-none">{tennisWinRate}</p>
              <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-text-disabled">{$_('profile_win_rate')} %</p>
            </div>
          </div>
        </div>
      {/if}

      <!-- Contacts -->
      <Collapsible.Root bind:open={showContacts} class="space-y-3">
        <Collapsible.Trigger class="flex w-full items-center justify-between">
          <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-text-secondary">{$_('profile_contacts_title')}</p>
          <ChevronDown size={14} class="text-text-disabled transition-transform duration-200 data-[state=open]:rotate-180" />
        </Collapsible.Trigger>

        <Collapsible.Content class="space-y-3">
            <!-- Search -->
            <div class="relative">
              <div class="pointer-events-none absolute inset-y-0 left-3.5 flex items-center">
                <Search size={15} class="text-text-disabled" />
              </div>
              <input
                type="text"
                placeholder={$_('profile_contacts_search_placeholder')}
                bind:value={contactSearch}
                oninput={onContactSearchInput}
                class="w-full rounded-xl bg-surface-raised py-2.5 pl-9 pr-4 text-sm outline-none focus:ring-2 focus:ring-primary transition-shadow"
              />
            </div>

            <!-- Search results section -->
            {#if contactSearch.length >= 2}
              <div class="space-y-2">
                <p class="text-xs font-semibold text-text-secondary px-1">
                  Search results {#if searchResults.length > 0}({searchResults.length}){/if}
                </p>
                {#if searchResults.length > 0}
                  <div class="space-y-1.5">
                    {#each searchResults as result}
                      <div class="flex items-center gap-3 rounded-2xl bg-surface-raised px-4 py-3">
                        <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary-muted text-xs font-[800] text-primary">
                          {result.display_name[0].toUpperCase()}
                        </div>
                        <p class="flex-1 text-sm font-semibold truncate">{result.display_name}</p>
                        {#if result.is_contact}
                          <button onclick={() => handleDeleteContact({ user_id: result.id, display_name: result.display_name } as App.Contact)} class="text-text-disabled hover:text-destructive transition-colors" aria-label="Remove contact">
                            <X size={16} />
                          </button>
                        {:else}
                          <button onclick={() => addContact(result.id)} class="text-primary" aria-label="Add contact">
                            <UserPlus size={16} />
                          </button>
                        {/if}
                      </div>
                    {/each}
                  </div>
                {:else if !searchLoading}
                  <p class="text-sm text-text-disabled py-1">{$_('profile_contacts_search_empty')}</p>
                {/if}
              </div>
            {/if}

            <!-- Saved contacts section (always visible) -->
            {#if contacts.length === 0 && contactSearch.length < 2}
              <p class="text-sm text-text-disabled py-1">{$_('profile_contacts_empty')}</p>
            {:else if contacts.length > 0}
              <div class="space-y-2">
                {#if contactSearch.length >= 2}
                  <p class="text-xs font-semibold text-text-secondary px-1">Your contacts</p>
                {/if}
                <div class="space-y-1.5">
                  {#each contacts as contact}
                    <div class="flex items-center gap-3 rounded-2xl bg-surface-raised px-4 py-3">
                      <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary-muted text-xs font-[800] text-primary">
                        {contact.display_name[0].toUpperCase()}
                      </div>
                      <p class="flex-1 text-sm font-semibold truncate">{contact.display_name}</p>
                      <button onclick={() => handleDeleteContact(contact)} class="text-text-disabled hover:text-destructive transition-colors" aria-label="Remove contact">
                        <X size={16} />
                      </button>
                    </div>
                  {/each}
                </div>
              </div>
            {/if}
        </Collapsible.Content>
      </Collapsible.Root>

      <!-- Delete contact confirmation -->
      <ConfirmDialog
        open={showContactDeleteConfirm}
        title="Delete Contact?"
        description="Remove {contactToDelete?.display_name || 'this contact'} from your contacts. This action cannot be undone."
        confirmLabel="Delete"
        cancelLabel="Cancel"
        destructive={true}
        onconfirm={confirmDeleteContact}
        oncancel={() => (showContactDeleteConfirm = false)}
      />

      <!-- Upcoming -->
      <Collapsible.Root bind:open={showUpcoming} class="space-y-3">
        <Collapsible.Trigger class="flex w-full items-center justify-between">
          <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-text-secondary">{$_('profile_upcoming_label')}</p>
          <ChevronDown size={14} class="text-text-disabled transition-transform duration-200 data-[state=open]:rotate-180" />
        </Collapsible.Trigger>

        <Collapsible.Content class="space-y-2">
            {#if upcoming.length === 0}
              <p class="text-sm text-text-disabled py-1">{$_('profile_upcoming_empty')}</p>
            {:else}
              {#each upcoming as t}
                <a href="/s/{t.session_id}" class="flex items-center gap-4 rounded-2xl bg-surface-raised px-4 py-3.5 transition-colors hover:bg-border">
                  <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full {t.status === 'active' ? 'bg-emerald-500/15 text-emerald-500' : 'bg-primary-muted text-primary'}">
                    {#if t.status === 'active'}<Radio size={18} />{:else}<CalendarDays size={18} />{/if}
                  </div>
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center gap-2">
                      <p class="truncate font-semibold text-sm">{t.name}</p>
                      {#if t.status === 'active'}
                        <span class="shrink-0 rounded-full bg-emerald-500/15 px-2 py-0.5 text-[10px] font-bold uppercase tracking-wide text-emerald-500">Live</span>
                      {/if}
                    </div>
                    <p class="text-xs text-text-secondary">{t.player_count} {$_('profile_upcoming_players')} · {t.game_mode === 'tennis' ? $_('create_mode_tennis') : 'Americano'}</p>
                  </div>
                  <span class="text-sm text-text-secondary">→</span>
                </a>
              {/each}
            {/if}
        </Collapsible.Content>
      </Collapsible.Root>

      <!-- Tournament history -->
      <Collapsible.Root bind:open={showHistory} class="space-y-3">
        <Collapsible.Trigger class="flex w-full items-center justify-between">
          <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-text-secondary">{$_('profile_history_label')}</p>
          <ChevronDown size={14} class="text-text-disabled transition-transform duration-200 data-[state=open]:rotate-180" />
        </Collapsible.Trigger>

        <Collapsible.Content class="space-y-2">
            {#if tournaments.length === 0}
              <p class="text-sm text-text-disabled py-2">{$_('profile_history_empty')}</p>
            {:else}
              {#each tournaments as t}
                <a href="/s/{t.session_id}" class="flex items-center gap-4 rounded-2xl bg-surface-raised px-4 py-3.5 transition-colors hover:bg-border">
                  <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full text-xs font-[800]
                    {t.rank === 1 ? 'bg-primary text-white' : 'bg-border text-text-secondary'}">
                    {ordinal(t.rank)}
                  </div>
                  <div class="flex-1 min-w-0">
                    <p class="truncate font-semibold text-sm">{t.name}</p>
                    <p class="text-xs text-text-secondary">
                      {formatDate(t.played_at)} · {t.points} pts
                      {#if t.ended_early}
                        · <span class="text-text-disabled">{$_('profile_ended_early')}</span>
                      {/if}
                    </p>
                  </div>
                  <span class="text-sm text-text-secondary">→</span>
                </a>
              {/each}
            {/if}
        </Collapsible.Content>
      </Collapsible.Root>

    {/if}

    <!-- Actions -->
    <div class="space-y-3">
      <div class="pt-2 space-y-2">
        <button
          onclick={() => auth.logout().then(() => goto('/'))}
          class="w-full rounded-2xl border border-border px-4 py-3.5 text-sm font-semibold text-text-secondary transition-colors hover:border-destructive hover:text-destructive"
        >
          {$_('auth_sign_out')}
        </button>
        <button
          onclick={() => (showDeleteConfirm = true)}
          class="w-full px-4 py-2 text-sm text-text-disabled hover:text-destructive transition-colors"
        >
          {$_('profile_delete_account')}
        </button>
      </div>
    </div>

  {/if}

  <Footer />

</main>

<CreateDrawer bind:open={showCreateDrawer} />

<Dialog.Root bind:open={showAvatarPicker}>
  <Dialog.Content class="w-full max-w-sm">
    <Dialog.Header>
      <div class="flex items-center gap-3">
        <Avatar icon={pickerIcon} color="forest" name={auth.user?.display_name ?? ''} size="md" />
        <Dialog.Title>Choose avatar</Dialog.Title>
      </div>
    </Dialog.Header>
    <div class="grid grid-cols-8 gap-1.5">
      <!-- "Use initials" option -->
      <button
        onclick={() => pickerIcon = ''}
        class="flex items-center justify-center rounded-xl p-1 transition-colors
          {pickerIcon === '' ? 'bg-primary-muted ring-2 ring-primary' : 'bg-surface-raised hover:bg-border'}"
        aria-label="Use initials"
      >
        <Avatar icon="" color="forest" name={auth.user?.display_name ?? ''} size="sm" />
      </button>
      {#each AVATAR_ICONS as icon}
        <button
          onclick={() => pickerIcon = icon}
          class="flex items-center justify-center rounded-xl p-1 transition-colors
            {pickerIcon === icon ? 'bg-primary-muted ring-2 ring-primary' : 'bg-surface-raised hover:bg-border'}"
          aria-label={icon}
        >
          <Avatar {icon} color="forest" name="" size="sm" />
        </button>
      {/each}
    </div>
    <Dialog.Footer>
      <button
        onclick={async () => { await saveAvatar(); showAvatarPicker = false; }}
        disabled={savingAvatar}
        class="w-full rounded-2xl bg-primary py-3.5 text-sm font-semibold text-white disabled:opacity-50"
      >
        {savingAvatar ? 'Saving…' : 'Save'}
      </button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

{#if showDeleteConfirm}
  <div class="fixed inset-0 z-50 flex items-end justify-center bg-black/40 p-4 sm:items-center">
    <div class="w-full max-w-sm rounded-3xl bg-surface p-6 space-y-4 shadow-xl">
      <div class="space-y-1">
        <h2 class="text-lg font-[800]">{$_('profile_delete_title')}</h2>
        <p class="text-sm text-text-secondary">{$_('profile_delete_desc')}</p>
      </div>
      <div class="space-y-2 pt-1">
        <button onclick={deleteAccount} disabled={deleting} class="w-full rounded-2xl bg-destructive px-4 py-3.5 text-sm font-semibold text-white disabled:opacity-60">
          {deleting ? $_('profile_delete_loading') : $_('profile_delete_confirm')}
        </button>
        <button onclick={() => (showDeleteConfirm = false)} class="w-full rounded-2xl border border-border px-4 py-3.5 text-sm font-semibold text-text-secondary">
          {$_('profile_delete_cancel')}
        </button>
      </div>
    </div>
  </div>
{/if}
