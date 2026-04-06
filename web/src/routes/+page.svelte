<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/state';
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';
  import { auth } from '$lib/auth.svelte';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import Footer from '$lib/components/Footer.svelte';
  import { _ } from 'svelte-i18n';
  import { initials } from '$lib/utils';
  import { fly } from 'svelte/transition';
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

  const scheduleTime = $derived(slotToLabel(timeSlot));
  let creating = $state(false);
  let error = $state('');
  let joinCode = $state('');
  let rejoinSession = $state<App.Session | null>(null);
  let rejoinHref = $state('');
  let showDeletedBanner = $state(false);
  let showNotFoundBanner = $state(false);

  onMount(async () => {
    // If ?create=1, go straight to setup (used by profile "New tournament" link)
    if (page.url.searchParams.get('create') === '1') {
      step = 'setup';
    }

    if (page.url.searchParams.get('deleted') === '1') {
      showDeletedBanner = true;
      setTimeout(() => { showDeletedBanner = false; }, 5000);
    }
    if (page.url.searchParams.get('notfound') === '1') {
      showNotFoundBanner = true;
      setTimeout(() => { showNotFoundBanner = false; }, 5000);
    }

    // Load rejoin session in parallel with auth check
    const lastId = localStorage.getItem('last_session_id');
    if (lastId) {
      try {
        const token = localStorage.getItem(`admin_token_${lastId}`) ?? undefined;
        const s = await api.sessions.get(lastId, token);
        if (s.status === 'lobby' || s.status === 'active') {
          rejoinSession = s;
          rejoinHref = token ? `/s/${lastId}?token=${token}` : `/s/${lastId}`;
        }
      } catch {
        localStorage.removeItem('last_session_id');
      }
    }
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
    error = '';
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
      error = e instanceof Error ? e.message : 'Something went wrong';
      creating = false;
    }
  }

  function joinByCode() {
    const code = joinCode.trim().toUpperCase();
    if (code) goto(`/s/${code}`);
  }
</script>

{#if step === 'home'}
  {#if showDeletedBanner}
    <div transition:fly={{ y: -48, duration: 400 }} class="fixed inset-x-0 top-0 z-50 flex items-center justify-center bg-[var(--primary)] px-4 py-3 text-sm font-semibold text-white">
      {$_('home_account_deleted')}
    </div>
  {/if}
  {#if showNotFoundBanner}
    <div transition:fly={{ y: -48, duration: 400 }} class="fixed inset-x-0 top-0 z-50 flex items-center justify-center bg-[var(--destructive)] px-4 py-3 text-sm font-semibold text-white">
      {$_('home_session_not_found')}
    </div>
  {/if}
  <main class="flex min-h-svh flex-col items-center px-6 py-12" class:pt-16={showDeletedBanner || showNotFoundBanner}>
  <div class="flex w-full max-w-sm flex-1 flex-col">
    <div class="flex flex-1 flex-col justify-center space-y-12">
      <!-- Brand -->
      <div class="space-y-1">
        <h1 class="text-[28px] font-[800] text-[var(--primary)]">NotTennis</h1>
        <p class="text-[var(--text-secondary)]">{$_('home_tagline')}</p>
      </div>

      <!-- Actions -->
      <div class="space-y-4">
        {#if rejoinSession}
          <a
            href={rejoinHref}
            class="flex items-center gap-3 rounded-2xl bg-[var(--surface-raised)] px-4 py-3.5 transition-colors hover:bg-[var(--border)]"
          >
            <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-[var(--primary-muted)]">
              <div class="h-2.5 w-2.5 rounded-full bg-[var(--primary)] animate-pulse"></div>
            </div>
            <div class="flex-1 min-w-0">
              <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-[var(--text-disabled)]">{$_('home_rejoin_label')}</p>
              <p class="truncate text-sm font-semibold">{rejoinSession.name || 'NotTennis'}</p>
            </div>
            <span class="text-sm text-[var(--text-secondary)]">→</span>
          </a>
        {/if}

        {#if auth.ready && !auth.user}
          <a
            href="/auth"
            class="flex h-auto w-full items-center justify-center rounded-2xl bg-[var(--primary)] px-4 py-4 text-[15px] font-semibold text-white hover:bg-[var(--primary-hover)]"
          >
            {$_('auth_sign_in')} →
          </a>
          <p class="text-center text-sm text-[var(--text-secondary)]">
            {$_('auth_no_account')}
            <a href="/auth?register=1" class="font-semibold text-[var(--primary)] hover:text-[var(--primary-hover)]">
              {$_('auth_switch_register')}
            </a>
          </p>
        {/if}

        <div class="flex items-center gap-3">
          <div class="h-px flex-1 bg-[var(--border)]"></div>
          <span class="text-xs text-[var(--text-disabled)]">{$_('home_join_code_divider')}</span>
          <div class="h-px flex-1 bg-[var(--border)]"></div>
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
            class="min-w-0 flex-1 rounded-2xl border-0 bg-[var(--surface-raised)] px-4 py-3.5 text-sm"
          />
          <Button
            type="submit"
            disabled={!joinCode.trim()}
            variant="secondary"
            class="h-auto rounded-2xl bg-[var(--surface-raised)] px-5 text-sm font-semibold text-[var(--text-primary)] hover:bg-[var(--border)]"
          >
            {$_('home_join_button')}
          </Button>
        </form>
      </div>
    </div>

    <!-- Auth (logged-in pill — guests see the sign-in button above instead) -->
    {#if auth.user}
    <div class="flex justify-center pt-6">
      <a href="/profile" class="flex items-center gap-3 rounded-2xl px-3 py-2 transition-colors hover:bg-[var(--surface-raised)]">
        <div class="flex h-7 w-7 items-center justify-center rounded-full bg-[var(--primary-muted)] text-xs font-[800] text-[var(--primary)]">
          {initials(auth.user.display_name)}
        </div>
        <span class="text-sm font-semibold">{auth.user.display_name}</span>
        <span class="text-xs text-[var(--text-disabled)]">→</span>
      </a>
    </div>
    {/if}

  </div>
  <Footer />
  </main>

{:else}
  <main class="flex min-h-svh flex-col items-center px-6 py-6">
  <div class="w-full max-w-sm">
    <!-- Nav -->
    <nav class="flex items-center justify-between">
      <Button
        onclick={() => goto('/profile')}
        variant="ghost"
        class="flex h-8 w-8 items-center justify-center rounded-full p-0 text-lg text-[var(--text-secondary)]"
      >
        ×
      </Button>
      <span class="text-sm font-semibold text-[var(--primary)]">NotTennis</span>
      <div class="w-8"></div>
    </nav>

    <!-- Header -->
    <div class="mt-8 space-y-2">
      <h1 class="text-[34px] font-[800]">{$_('create_title_line1')}<br />{$_('create_title_line2')}</h1>
      <p class="text-[var(--text-secondary)]">{$_('create_subtitle')}</p>
    </div>

    <!-- Form -->
    <div class="mt-8 space-y-7">

      <!-- Game mode -->
      <div class="space-y-2.5">
        <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('create_game_mode_label')}</p>
        <div class="flex gap-2">
          <button
            onclick={() => (gameMode = 'americano')}
            class="flex-1 rounded-full py-2.5 text-sm font-semibold transition-colors {gameMode === 'americano'
              ? 'bg-[var(--primary)] text-white'
              : 'bg-[var(--surface-raised)] text-[var(--text-primary)] hover:bg-[var(--border)]'}"
          >
            Americano
          </button>
          <button
            onclick={() => (gameMode = 'tennis')}
            class="flex-1 rounded-full py-2.5 text-sm font-semibold transition-colors {gameMode === 'tennis'
              ? 'bg-[var(--primary)] text-white'
              : 'bg-[var(--surface-raised)] text-[var(--text-primary)] hover:bg-[var(--border)]'}"
          >
            {$_('create_mode_tennis')}
          </button>
        </div>
      </div>

      {#if gameMode === 'americano'}
      <!-- Courts -->
      <div class="space-y-2.5">
        <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('create_courts_label')}</p>
        <div class="flex gap-2">
          {#each [1, 2, 3, 4] as n}
            <button
              onclick={() => (courts = n)}
              class="flex-1 rounded-full py-2.5 text-sm font-semibold transition-colors {courts === n
                ? 'bg-[var(--primary)] text-white'
                : 'bg-[var(--surface-raised)] text-[var(--text-primary)] hover:bg-[var(--border)]'}"
            >
              {n}
            </button>
          {/each}
        </div>
      </div>

      <!-- Points -->
      <div class="space-y-2.5">
        <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('create_points_label')}</p>
        <div class="flex gap-2">
          {#each [16, 24, 32] as p}
            <button
              onclick={() => (points = p)}
              class="flex-1 rounded-full py-2.5 text-sm font-semibold transition-colors {points === p
                ? 'bg-[var(--primary)] text-white'
                : 'bg-[var(--surface-raised)] text-[var(--text-primary)] hover:bg-[var(--border)]'}"
            >
              {p}
            </button>
          {/each}
        </div>
        <p class="text-xs text-[var(--text-secondary)]">
          {points === 16 ? $_('create_points_quick') : points === 24 ? $_('create_points_standard') : $_('create_points_long')}
        </p>
      </div>
      {:else}
      <!-- Sets to win (tennis) -->
      <div class="space-y-2.5">
        <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('create_sets_label')}</p>
        <div class="flex gap-2">
          <button
            onclick={() => (setsToWin = 2)}
            class="flex-1 rounded-full py-2.5 text-sm font-semibold transition-colors {setsToWin === 2
              ? 'bg-[var(--primary)] text-white'
              : 'bg-[var(--surface-raised)] text-[var(--text-primary)] hover:bg-[var(--border)]'}"
          >
            {$_('create_sets_bo3')}
          </button>
          <button
            onclick={() => (setsToWin = 3)}
            class="flex-1 rounded-full py-2.5 text-sm font-semibold transition-colors {setsToWin === 3
              ? 'bg-[var(--primary)] text-white'
              : 'bg-[var(--surface-raised)] text-[var(--text-primary)] hover:bg-[var(--border)]'}"
          >
            {$_('create_sets_bo5')}
          </button>
        </div>
      </div>

      <!-- Games per set (tennis) -->
      <div class="space-y-2.5">
        <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('create_games_per_set_label')}</p>
        <div class="flex gap-2">
          <button
            onclick={() => (gamesPerSet = 4)}
            class="flex-1 rounded-full py-2.5 text-sm font-semibold transition-colors {gamesPerSet === 4
              ? 'bg-[var(--primary)] text-white'
              : 'bg-[var(--surface-raised)] text-[var(--text-primary)] hover:bg-[var(--border)]'}"
          >
            {$_('create_games_per_set_4')}
          </button>
          <button
            onclick={() => (gamesPerSet = 6)}
            class="flex-1 rounded-full py-2.5 text-sm font-semibold transition-colors {gamesPerSet === 6
              ? 'bg-[var(--primary)] text-white'
              : 'bg-[var(--surface-raised)] text-[var(--text-primary)] hover:bg-[var(--border)]'}"
          >
            {$_('create_games_per_set_6')}
          </button>
        </div>
      </div>
      {/if}

      <!-- Tournament name (optional) -->
      <div class="space-y-2.5">
        <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('create_tournament_name_label')}</p>
        <Input
          bind:value={tournamentName}
          placeholder={$_('create_tournament_name_placeholder')}
          maxlength={48}
          class="rounded-2xl border-0 bg-[var(--surface-raised)] px-4 py-3.5 text-sm"
        />
      </div>

      <!-- Schedule (optional) -->
      <div class="space-y-2.5">
        <div class="flex items-center justify-between">
          <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('create_schedule_label')}</p>
          <button
            type="button"
            onclick={() => { scheduleEnabled = !scheduleEnabled; if (!scheduleEnabled) { calendarDate = undefined; timeSlot = 20; } }}
            aria-label={$_('create_schedule_label')}
            class="relative h-6 w-11 rounded-full transition-colors {scheduleEnabled ? 'bg-[var(--primary)]' : 'bg-[var(--border)]'}"
          >
            <span class="absolute top-0.5 left-0.5 h-5 w-5 rounded-full bg-white shadow transition-transform {scheduleEnabled ? 'translate-x-5' : 'translate-x-0'}"></span>
          </button>
        </div>
        {#if scheduleEnabled}
          <div class="rounded-2xl bg-[var(--surface-raised)] overflow-hidden">
            <Calendar bind:value={calendarDate} minValue={today(getLocalTimeZone())} weekStartsOn={1} />
            <div class="px-4 pb-4 space-y-2">
              <div class="flex items-center justify-between">
                <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-disabled)]">{$_('create_schedule_time_label')}</p>
                <p class="text-sm font-[800] text-[var(--primary)]">{scheduleTime}</p>
              </div>
              <input
                type="range"
                min="0"
                max="27"
                step="1"
                bind:value={timeSlot}
                class="w-full accent-[var(--primary)]"
              />
              <div class="flex justify-between text-[10px] text-[var(--text-disabled)]">
                <span>08:00</span>
                <span>21:30</span>
              </div>
            </div>
          </div>
        {/if}
      </div>

      <!-- Organiser (always logged in at this point) -->
      <div class="space-y-2.5">
        <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('create_organiser_label')}</p>
        <div class="flex items-center gap-3 rounded-2xl bg-[var(--surface-raised)] px-4 py-3.5">
          <div class="flex h-7 w-7 items-center justify-center rounded-full bg-[var(--primary-muted)] text-xs font-[800] text-[var(--primary)]">
            {auth.user ? initials(auth.user.display_name) : '?'}
          </div>
          <span class="text-sm font-semibold">{auth.user?.display_name}</span>
        </div>
      </div>

      <!-- Info note -->
      <div class="flex gap-3 rounded-2xl bg-[var(--surface-raised)] px-4 py-3.5">
        <span class="mt-px shrink-0 text-[var(--text-secondary)]">ℹ</span>
        <p class="text-sm text-[var(--text-secondary)]">{$_('create_info_note')}</p>
      </div>

      {#if error}
        <p class="text-sm text-[var(--destructive)]">{error}</p>
      {/if}

      <Button
        onclick={create}
        disabled={creating}
        class="h-auto w-full rounded-2xl bg-[var(--primary)] px-4 py-4 text-[15px] font-semibold text-white hover:bg-[var(--primary-hover)]"
      >
        {creating ? $_('create_button_loading') : $_('create_button')}
      </Button>
    </div>
  </div>
  </main>
{/if}
