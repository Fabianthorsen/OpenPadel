<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/state';
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';
  import { auth } from '$lib/auth.svelte';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import { SectionLabel } from '$lib/components/ui/section-label';
  import { PillToggleGroup, PillToggleItem } from '$lib/components/ui/pill-toggle-group';
  import { Switch } from '$lib/components/ui/switch';
  import Footer from '$lib/components/Footer.svelte';
  import PullToRefresh from '$lib/components/PullToRefresh.svelte';
  import { _ } from 'svelte-i18n';
  import { initials } from '$lib/utils';
  import { toast } from 'svelte-sonner';
  import { ApiError } from '$lib/api/client';
  import { translateApiError } from '$lib/i18n/errors';
  import { Calendar } from '$lib/components/ui/calendar';
  import { type DateValue, today, getLocalTimeZone } from '@internationalized/date';

  let step = $state<'home' | 'setup'>('home');
  let gameMode = $state<'americano' | 'tennis'>('americano');
  let courts = $state(2);
  let points = $state(24);
  let setsToWin = $state(2);
  let gamesPerSet = $state(6);
  let tournamentName = $state('');
  let scheduleEnabled = $state(false);
  let calendarDate = $state<DateValue | undefined>(undefined);
  // Slider: 0 = 08:00, 1 = 08:30, ..., 27 = 21:30
  let timeSlot = $state(20); // default 18:00

  function slotToLabel(slot: number) {
    const totalMins = 8 * 60 + slot * 30;
    const h = String(Math.floor(totalMins / 60)).padStart(2, '0');
    const m = String(totalMins % 60).padStart(2, '0');
    return `${h}:${m}`;
  }

  function calculateNextHourSlot(): number {
    const now = new Date();
    const currentHour = now.getHours();
    const currentMinutes = now.getMinutes();

    // Calculate next whole hour
    const nextHour = currentMinutes > 0 ? currentHour + 1 : currentHour;

    // Clamp to 8-21 range (08:00 to 21:30)
    const clampedHour = Math.min(21, Math.max(8, nextHour));

    // Convert hour to slot (slot 0 = 08:00, slot 27 = 21:30)
    // slot = (hour * 60 - 8 * 60) / 30
    return Math.round((clampedHour * 60 - 8 * 60) / 30);
  }

  const scheduleTime = $derived(slotToLabel(timeSlot));
  let creating = $state(false);
  let joinCode = $state('');
  let rejoinSession = $state<App.Session | null>(null);
  let rejoinHref = $state('');

  async function loadRejoin() {
    const lastId = localStorage.getItem('last_session_id');
    if (!lastId) { rejoinSession = null; return; }
    try {
      const token = localStorage.getItem(`admin_token_${lastId}`) ?? undefined;
      const s = await api.sessions.get(lastId, token);
      if (s.status === 'lobby' || s.status === 'active') {
        rejoinSession = s;
        rejoinHref = token ? `/s/${lastId}?token=${token}` : `/s/${lastId}`;
      } else {
        rejoinSession = null;
      }
    } catch {
      localStorage.removeItem('last_session_id');
      rejoinSession = null;
    }
  }

  onMount(async () => {
    // If ?create=1, go straight to setup (used by profile "New tournament" link)
    if (page.url.searchParams.get('create') === '1') {
      step = 'setup';
    }

    if (page.url.searchParams.get('deleted') === '1') {
      toast($_('home_account_deleted'));
    }
    if (page.url.searchParams.get('notfound') === '1') {
      toast.error($_('home_session_not_found'));
    }

    await loadRejoin();
  });

  // Redirect logged-in users to profile
  $effect(() => {
    if (auth.ready && auth.user && step === 'home') {
      const notfound = page.url.searchParams.get('notfound');
      const create = page.url.searchParams.get('create');
      goto((notfound ? '/profile?notfound=1' : '/profile') + (create ? (notfound ? '&create=1' : '?create=1') : ''));
    }
  });

  async function create() {
    const effectiveName = auth.user!.display_name;
    creating = true;
    try {
      let iso: string | undefined;
      if (scheduleEnabled && calendarDate) {
        const [h, m] = scheduleTime.split(':').map(Number);
        const d = calendarDate.toDate(getLocalTimeZone());
        d.setHours(h, m, 0, 0);
        iso = d.toISOString();
      }
      const session = await api.sessions.create(courts, points, tournamentName.trim(), gameMode, setsToWin, gamesPerSet, iso);
      const adminToken = session.admin_token!;
      localStorage.setItem(`admin_token_${session.id}`, adminToken);
      const player = await api.players.join(session.id, effectiveName, auth.token ?? undefined, adminToken);
      localStorage.setItem(`player_id_${session.id}`, player.id);
      localStorage.setItem('last_session_id', session.id);
      goto(`/s/${session.id}?token=${adminToken}`);
    } catch (e) {
      toast.error(e instanceof ApiError ? translateApiError(e.message) : translateApiError('server_error'));
      creating = false;
    }
  }

  function joinByCode() {
    const code = joinCode.trim().toUpperCase();
    if (code) goto(`/s/${code}`);
  }
</script>

{#if step === 'home'}
  <PullToRefresh onRefresh={loadRejoin}>
  <main class="flex min-h-svh flex-col items-center px-6 pb-12 pt-safe-page">
  <div class="flex w-full max-w-sm flex-1 flex-col">
    <div class="flex flex-1 flex-col justify-center space-y-12">
      <!-- Brand -->
      <div class="space-y-1">
        <h1 class="text-[28px] font-[800] text-primary">OpenPadel</h1>
        <p class="text-text-secondary">{$_('home_tagline')}</p>
      </div>

      <!-- Actions -->
      <div class="space-y-4">
        {#if rejoinSession}
          <a
            href={rejoinHref}
            class="flex items-center gap-3 rounded-2xl bg-surface-raised px-4 py-3.5 transition-colors hover:bg-border"
          >
            <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-primary-muted">
              <div class="h-2.5 w-2.5 rounded-full bg-primary animate-pulse"></div>
            </div>
            <div class="flex-1 min-w-0">
              <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-text-disabled">{$_('home_rejoin_label')}</p>
              <p class="truncate text-sm font-semibold">{rejoinSession.name || 'OpenPadel'}</p>
            </div>
            <span class="text-sm text-text-secondary">→</span>
          </a>
        {/if}

        {#if auth.ready && !auth.user}
          <a
            href="/auth"
            class="flex h-auto w-full items-center justify-center rounded-2xl bg-primary px-4 py-4 text-[15px] font-semibold text-white hover:bg-primary-hover"
          >
            {$_('auth_sign_in')} →
          </a>
          <p class="text-center text-sm text-text-secondary">
            {$_('auth_no_account')}
            <a href="/auth?register=1" class="font-semibold text-primary hover:text-primary-hover">
              {$_('auth_switch_register')}
            </a>
          </p>
        {/if}

        <div class="flex items-center gap-3">
          <div class="h-px flex-1 bg-border"></div>
          <span class="text-xs text-text-disabled">{$_('home_join_code_divider')}</span>
          <div class="h-px flex-1 bg-border"></div>
        </div>

        <form onsubmit={(e) => { e.preventDefault(); joinByCode(); }} class="flex gap-2">
          <Input
            bind:value={joinCode}
            oninput={(e: Event) => { joinCode = (e.currentTarget as HTMLInputElement).value.toUpperCase(); }}
            placeholder={$_('home_join_code_placeholder')}
            maxlength={4}
            autocomplete="off"
            autocorrect="off"
            autocapitalize="characters"
            spellcheck={false}
            class="min-w-0 flex-1 rounded-2xl border-0 bg-surface-raised px-4 py-3.5 text-sm"
          />
          <Button
            type="submit"
            disabled={!joinCode.trim()}
            variant="secondary"
            class="h-auto rounded-2xl bg-surface-raised px-5 text-sm font-semibold text-text-primary hover:bg-border"
          >
            {$_('home_join_button')}
          </Button>
        </form>
      </div>
    </div>

    <!-- Auth (logged-in pill — guests see the sign-in button above instead) -->
    {#if auth.user}
    <div class="flex justify-center pt-6">
      <a href="/profile" class="flex items-center gap-3 rounded-2xl px-3 py-2 transition-colors hover:bg-surface-raised">
        <div class="flex h-7 w-7 items-center justify-center rounded-full bg-primary-muted text-xs font-[800] text-primary">
          {initials(auth.user.display_name)}
        </div>
        <span class="text-sm font-semibold">{auth.user.display_name}</span>
        <span class="text-xs text-text-disabled">→</span>
      </a>
    </div>
    {/if}

  </div>
  <Footer />
  </main>
  </PullToRefresh>

{:else}
  <main class="flex min-h-svh flex-col items-center px-6 pb-6 pt-safe-page">
  <div class="w-full max-w-sm">
    <!-- Nav -->
    <nav class="flex items-center justify-between">
      <Button
        onclick={() => goto('/profile')}
        variant="ghost"
        class="flex h-8 w-8 items-center justify-center rounded-full p-0 text-lg text-text-secondary"
      >
        ×
      </Button>
      <span class="text-sm font-semibold text-primary">OpenPadel</span>
      <div class="w-8"></div>
    </nav>

    <!-- Header -->
    <div class="mt-8 space-y-2">
      <h1 class="text-[34px] font-[800]">{$_('create_title_line1')}<br />{$_('create_title_line2')}</h1>
      <p class="text-text-secondary">{$_('create_subtitle')}</p>
    </div>

    <!-- Form -->
    <div class="mt-8 space-y-7">

      <!-- Game mode -->
      <div class="space-y-2.5">
        <SectionLabel>{$_('create_game_mode_label')}</SectionLabel>
        <PillToggleGroup
          value={gameMode}
          onValueChange={(val) => gameMode = val as 'americano' | 'tennis'}
        >
          <PillToggleItem value="americano">
            Americano
          </PillToggleItem>
          <PillToggleItem value="tennis">
            {$_('create_mode_tennis')}
          </PillToggleItem>
        </PillToggleGroup>
      </div>

      {#if gameMode === 'americano'}
      <!-- Courts -->
      <div class="space-y-2.5">
        <SectionLabel>{$_('create_courts_label')}</SectionLabel>
        <PillToggleGroup
          value={courts.toString()}
          onValueChange={(val) => courts = parseInt(val)}
        >
          {#each [1, 2, 3, 4] as n}
            <PillToggleItem value={n.toString()}>
              {n}
            </PillToggleItem>
          {/each}
        </PillToggleGroup>
      </div>

      <!-- Points -->
      <div class="space-y-2.5">
        <SectionLabel>{$_('create_points_label')}</SectionLabel>
        <PillToggleGroup
          value={points.toString()}
          onValueChange={(val) => points = parseInt(val)}
        >
          {#each [16, 24, 32] as p}
            <PillToggleItem value={p.toString()}>
              {p}
            </PillToggleItem>
          {/each}
        </PillToggleGroup>
        <p class="text-xs text-text-secondary">
          {points === 16 ? $_('create_points_quick') : points === 24 ? $_('create_points_standard') : $_('create_points_long')}
        </p>
      </div>
      {:else}
      <!-- Sets to win (tennis) -->
      <div class="space-y-2.5">
        <SectionLabel>{$_('create_sets_label')}</SectionLabel>
        <PillToggleGroup
          value={setsToWin.toString()}
          onValueChange={(val) => setsToWin = parseInt(val)}
        >
          <PillToggleItem value="2">
            {$_('create_sets_bo3')}
          </PillToggleItem>
          <PillToggleItem value="3">
            {$_('create_sets_bo5')}
          </PillToggleItem>
        </PillToggleGroup>
      </div>

      <!-- Games per set (tennis) -->
      <div class="space-y-2.5">
        <SectionLabel>{$_('create_games_per_set_label')}</SectionLabel>
        <PillToggleGroup
          value={gamesPerSet.toString()}
          onValueChange={(val) => gamesPerSet = parseInt(val)}
        >
          <PillToggleItem value="4">
            {$_('create_games_per_set_4')}
          </PillToggleItem>
          <PillToggleItem value="6">
            {$_('create_games_per_set_6')}
          </PillToggleItem>
        </PillToggleGroup>
      </div>
      {/if}

      <!-- Tournament name (optional) -->
      <div class="space-y-2.5">
        <SectionLabel>{$_('create_tournament_name_label')}</SectionLabel>
        <Input
          bind:value={tournamentName}
          placeholder={$_('create_tournament_name_placeholder')}
          maxlength={48}
          class="rounded-2xl border-0 bg-surface-raised px-4 py-3.5 text-sm"
        />
      </div>

      <!-- Schedule (optional) -->
      <div class="space-y-2.5">
        <div class="flex items-center justify-between">
          <SectionLabel>{$_('create_schedule_label')}</SectionLabel>
          <Switch
            checked={scheduleEnabled}
            onCheckedChange={(checked) => {
              scheduleEnabled = checked;
              if (!checked) {
                calendarDate = undefined;
                timeSlot = 20;
              } else {
                calendarDate = today(getLocalTimeZone());
                timeSlot = calculateNextHourSlot();
              }
            }}
          />
        </div>
        {#if scheduleEnabled}
          <div class="rounded-2xl bg-surface-raised overflow-hidden">
            <Calendar bind:value={calendarDate} minValue={today(getLocalTimeZone())} weekStartsOn={1} />
            <div class="px-4 pb-4 space-y-2">
              <div class="flex items-center justify-between">
                <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-text-disabled">{$_('create_schedule_time_label')}</p>
                <p class="text-sm font-[800] text-primary">{scheduleTime}</p>
              </div>
              <input
                type="range"
                min="0"
                max="27"
                step="1"
                bind:value={timeSlot}
                class="w-full accent-primary"
              />
              <div class="flex justify-between text-[10px] text-text-disabled">
                <span>08:00</span>
                <span>21:30</span>
              </div>
            </div>
          </div>
        {/if}
      </div>

      <!-- Organiser (always logged in at this point) -->
      <div class="space-y-2.5">
        <SectionLabel>{$_('create_organiser_label')}</SectionLabel>
        <div class="flex items-center gap-3 rounded-2xl bg-surface-raised px-4 py-3.5">
          <div class="flex h-7 w-7 items-center justify-center rounded-full bg-primary-muted text-xs font-[800] text-primary">
            {auth.user ? initials(auth.user.display_name) : '?'}
          </div>
          <span class="text-sm font-semibold">{auth.user?.display_name}</span>
        </div>
      </div>

      <!-- Info note -->
      <div class="flex gap-3 rounded-2xl bg-surface-raised px-4 py-3.5">
        <span class="mt-px shrink-0 text-text-secondary">ℹ</span>
        <p class="text-sm text-text-secondary">{$_('create_info_note')}</p>
      </div>

      <Button
        onclick={create}
        disabled={creating}
        class="h-auto w-full rounded-2xl bg-primary px-4 py-4 text-[15px] font-semibold text-white hover:bg-primary-hover"
      >
        {creating ? $_('create_button_loading') : $_('create_button')}
      </Button>
    </div>
  </div>
  </main>
{/if}
