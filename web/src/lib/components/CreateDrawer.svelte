<script lang="ts">
  import { goto } from '$app/navigation';
  import { api } from '$lib/api/client';
  import { auth } from '$lib/auth.svelte';
  import { _ } from 'svelte-i18n';
  import { initials } from '$lib/utils';
  import { fly } from 'svelte/transition';
  import { Calendar } from '$lib/components/ui/calendar';
  import { Input } from '$lib/components/ui/input';
  import { type DateValue, today, getLocalTimeZone } from '@internationalized/date';

  let { open = $bindable(false) }: { open?: boolean } = $props();

  let gameMode = $state<'americano' | 'tennis'>('americano');
  let courts = $state(2);
  let points = $state(24);
  let setsToWin = $state(2);
  let gamesPerSet = $state(6);
  let tournamentName = $state('');
  let scheduleEnabled = $state(false);
  let calendarDate = $state<DateValue | undefined>(undefined);
  let timeSlot = $state(20); // default 18:00
  let creating = $state(false);
  let error = $state('');

  function slotToLabel(slot: number) {
    const totalMins = 8 * 60 + slot * 30;
    const h = String(Math.floor(totalMins / 60)).padStart(2, '0');
    const m = String(totalMins % 60).padStart(2, '0');
    return `${h}:${m}`;
  }
  const scheduleTime = $derived(slotToLabel(timeSlot));

  function close() {
    open = false;
  }

  $effect(() => {
    if (open) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }
    return () => { document.body.style.overflow = ''; };
  });

  // Drag-down-to-close
  let dragStartY = 0;
  let dragOffset = $state(0);
  let dragging = $state(false);

  function onDragStart(e: TouchEvent) {
    dragStartY = e.touches[0].clientY;
    dragOffset = 0;
    dragging = true;
  }

  function onDragMove(e: TouchEvent) {
    if (!dragging) return;
    const delta = e.touches[0].clientY - dragStartY;
    dragOffset = Math.max(0, delta); // only allow dragging down
  }

  function onDragEnd() {
    dragging = false;
    if (dragOffset > 120) {
      close();
    }
    dragOffset = 0;
  }

  async function create() {
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
      const player = await api.players.join(session.id, auth.user!.display_name, auth.token ?? undefined, adminToken);
      localStorage.setItem(`player_id_${session.id}`, player.id);
      localStorage.setItem('last_session_id', session.id);
      goto(`/s/${session.id}?token=${adminToken}`);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Something went wrong';
      creating = false;
    }
  }
</script>

{#if open}
  <!-- Backdrop -->
  <button
    class="fixed inset-0 z-40 bg-black/40"
    onclick={close}
    aria-label="Close"
  ></button>

  <!-- Drawer -->
  <div
    transition:fly={{ y: 600, duration: 320, opacity: 1 }}
    class="fixed bottom-0 left-1/2 z-50 flex w-full max-w-[480px] flex-col rounded-t-3xl bg-[var(--surface)] shadow-2xl"
    style="height: 78vh; max-height: 78vh; transform: translateX(-50%) translateY({dragOffset}px); transition: {dragging ? 'none' : 'transform 0.3s ease'};"
  >
    <!-- Handle + header (drag target) -->
    <div
      role="presentation"
      class="flex shrink-0 flex-col items-center px-6 pt-3 pb-4 touch-none"
      ontouchstart={onDragStart}
      ontouchmove={onDragMove}
      ontouchend={onDragEnd}
    >
      <div class="mb-4 h-1 w-10 rounded-full bg-[var(--border-strong)]"></div>
      <div class="flex w-full items-center justify-between">
        <h2 class="text-lg font-[800]">{$_('create_title_line1')} {$_('create_title_line2')}</h2>
        <button
          onclick={close}
          class="hidden md:flex h-8 w-8 items-center justify-center rounded-full bg-[var(--surface-raised)] text-[var(--text-secondary)] hover:bg-[var(--border)] transition-colors text-xl leading-none"
        >×</button>
        <div class="md:hidden w-8"></div>
      </div>
    </div>

    <!-- Scrollable content -->
    <div class="flex-1 overflow-y-auto px-6 pb-8 space-y-6">

      <!-- Game mode -->
      <div class="space-y-2.5">
        <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('create_game_mode_label')}</p>
        <div class="flex gap-2">
          <button
            onclick={() => (gameMode = 'americano')}
            class="flex-1 rounded-full py-2.5 text-sm font-semibold transition-colors {gameMode === 'americano'
              ? 'bg-[var(--primary)] text-white'
              : 'bg-[var(--surface-raised)] text-[var(--text-primary)]'}"
          >Americano</button>
          <button
            onclick={() => (gameMode = 'tennis')}
            class="flex-1 rounded-full py-2.5 text-sm font-semibold transition-colors {gameMode === 'tennis'
              ? 'bg-[var(--primary)] text-white'
              : 'bg-[var(--surface-raised)] text-[var(--text-primary)]'}"
          >{$_('create_mode_tennis')}</button>
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
                  : 'bg-[var(--surface-raised)] text-[var(--text-primary)]'}"
              >{n}</button>
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
                  : 'bg-[var(--surface-raised)] text-[var(--text-primary)]'}"
              >{p}</button>
            {/each}
          </div>
          <p class="text-xs text-[var(--text-secondary)]">
            {points === 16 ? $_('create_points_quick') : points === 24 ? $_('create_points_standard') : $_('create_points_long')}
          </p>
        </div>
      {:else}
        <!-- Sets to win -->
        <div class="space-y-2.5">
          <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('create_sets_label')}</p>
          <div class="flex gap-2">
            <button
              onclick={() => (setsToWin = 2)}
              class="flex-1 rounded-full py-2.5 text-sm font-semibold transition-colors {setsToWin === 2
                ? 'bg-[var(--primary)] text-white'
                : 'bg-[var(--surface-raised)] text-[var(--text-primary)]'}"
            >{$_('create_sets_bo3')}</button>
            <button
              onclick={() => (setsToWin = 3)}
              class="flex-1 rounded-full py-2.5 text-sm font-semibold transition-colors {setsToWin === 3
                ? 'bg-[var(--primary)] text-white'
                : 'bg-[var(--surface-raised)] text-[var(--text-primary)]'}"
            >{$_('create_sets_bo5')}</button>
          </div>
        </div>

        <!-- Games per set -->
        <div class="space-y-2.5">
          <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('create_games_per_set_label')}</p>
          <div class="flex gap-2">
            <button
              onclick={() => (gamesPerSet = 4)}
              class="flex-1 rounded-full py-2.5 text-sm font-semibold transition-colors {gamesPerSet === 4
                ? 'bg-[var(--primary)] text-white'
                : 'bg-[var(--surface-raised)] text-[var(--text-primary)]'}"
            >{$_('create_games_per_set_4')}</button>
            <button
              onclick={() => (gamesPerSet = 6)}
              class="flex-1 rounded-full py-2.5 text-sm font-semibold transition-colors {gamesPerSet === 6
                ? 'bg-[var(--primary)] text-white'
                : 'bg-[var(--surface-raised)] text-[var(--text-primary)]'}"
            >{$_('create_games_per_set_6')}</button>
          </div>
        </div>
      {/if}

      <!-- Tournament name -->
      <div class="space-y-2.5">
        <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('create_tournament_name_label')}</p>
        <Input
          bind:value={tournamentName}
          placeholder={$_('create_tournament_name_placeholder')}
          maxlength={48}
          class="rounded-2xl border-0 bg-[var(--surface-raised)] px-4 py-3.5 text-sm"
        />
      </div>

      <!-- Schedule -->
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
              <input type="range" min="0" max="27" step="1" bind:value={timeSlot} class="w-full accent-[var(--primary)]" />
              <div class="flex justify-between text-[10px] text-[var(--text-disabled)]">
                <span>08:00</span><span>21:30</span>
              </div>
            </div>
          </div>
        {/if}
      </div>

      <!-- Organiser -->
      <div class="space-y-2.5">
        <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('create_organiser_label')}</p>
        <div class="flex items-center gap-3 rounded-2xl bg-[var(--surface-raised)] px-4 py-3.5">
          <div class="flex h-7 w-7 items-center justify-center rounded-full bg-[var(--primary-muted)] text-xs font-[800] text-[var(--primary)]">
            {auth.user ? initials(auth.user.display_name) : '?'}
          </div>
          <span class="text-sm font-semibold">{auth.user?.display_name}</span>
        </div>
      </div>

      {#if error}
        <p class="text-sm text-[var(--destructive)]">{error}</p>
      {/if}

      <button
        onclick={create}
        disabled={creating}
        class="w-full rounded-2xl bg-[var(--primary)] px-4 py-4 text-[15px] font-semibold text-white disabled:opacity-60 hover:bg-[var(--primary-hover)] transition-colors"
      >
        {creating ? $_('create_button_loading') : $_('create_button')}
      </button>

    </div>
  </div>
{/if}
