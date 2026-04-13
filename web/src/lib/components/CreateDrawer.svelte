<script lang="ts">
  import { goto } from '$app/navigation';
  import { api } from '$lib/api/client';
  import { auth } from '$lib/auth.svelte';
  import { _ } from 'svelte-i18n';
  import { initials } from '$lib/utils';
  import { ChevronDown, Check } from 'lucide-svelte';
  import { fly, slide } from 'svelte/transition';
  import { Calendar } from '$lib/components/ui/calendar';
  import { Input } from '$lib/components/ui/input';
  import { type DateValue, today, getLocalTimeZone } from '@internationalized/date';
  import { onMount } from 'svelte';

  let { open = $bindable(false) }: { open?: boolean } = $props();

  let gameMode = $state<'americano' | 'mexicano' | 'tennis'>('americano');
  let mexicanoRounds = $state<number | null>(null); // null = no round limit
  let courtDuration = $state<number | null>(null);  // null = no timer
  let customTimeMode = $state(false);
  let customTimeRaw = $state('');
  let customInputEl = $state<HTMLInputElement | null>(null);
  $effect(() => {
    if (customTimeMode && customInputEl) customInputEl.focus();
  });
  $effect(() => {
    if (customTimeMode) {
      const v = parseInt(customTimeRaw);
      courtDuration = (v >= 15 && v <= 300) ? v : null;
    }
  });
  function pickRounds(n: number) {
    mexicanoRounds = mexicanoRounds === n ? null : n;
    courtDuration = null;
    customTimeMode = false;
    customTimeRaw = '';
  }
  function pickTime(min: number) {
    courtDuration = courtDuration === min && !customTimeMode ? null : min;
    mexicanoRounds = null;
    customTimeMode = false;
    customTimeRaw = '';
  }
  function pickCustomTime() {
    customTimeMode = true;
    mexicanoRounds = null;
    courtDuration = null;
    customTimeRaw = '';
  }
  let courts = $state(2);
  let points = $state(24);
  let setsToWin = $state(2);
  let gamesPerSet = $state(6);
  let tournamentName = $state('');
  let scheduleEnabled = $state(false);
  let calendarDate = $state<DateValue | undefined>(undefined);
  let timeSlot = $state(20); // default 18:00

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
  let creating = $state(false);
  let error = $state('');

  let contacts = $state<App.Contact[]>([]);
  let selectedContacts = $state<Set<string>>(new Set());
  let showContacts = $state(false);

  onMount(async () => {
    if (auth.token) {
      contacts = await api.contacts.list(auth.token);
    }
  });

  function toggleContact(userID: string) {
    const next = new Set(selectedContacts);
    if (next.has(userID)) next.delete(userID);
    else next.add(userID);
    selectedContacts = next;
  }

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
      const session = await api.sessions.create(courts, points, tournamentName.trim(), gameMode, setsToWin, gamesPerSet, iso, gameMode === 'mexicano' ? mexicanoRounds ?? undefined : undefined, courtDuration ?? undefined);
      const adminToken = session.admin_token!;
      localStorage.setItem(`admin_token_${session.id}`, adminToken);
      const player = await api.players.join(session.id, auth.user!.display_name, auth.token ?? undefined, adminToken);
      localStorage.setItem(`player_id_${session.id}`, player.id);
      localStorage.setItem('last_session_id', session.id);
      // Send invites to selected contacts
      await Promise.all(
        [...selectedContacts].map(uid => api.invites.send(session.id, uid, auth.token!))
      );
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
            onclick={() => { gameMode = 'mexicano'; if (courts < 2) courts = 2; }}
            class="flex-1 rounded-full py-2.5 text-sm font-semibold transition-colors {gameMode === 'mexicano'
              ? 'bg-[var(--primary)] text-white'
              : 'bg-[var(--surface-raised)] text-[var(--text-primary)]'}"
          >Mexicano</button>
          <button
            onclick={() => (gameMode = 'tennis')}
            class="flex-1 rounded-full py-2.5 text-sm font-semibold transition-colors {gameMode === 'tennis'
              ? 'bg-[var(--primary)] text-white'
              : 'bg-[var(--surface-raised)] text-[var(--text-primary)]'}"
          >{$_('create_mode_tennis')}</button>
        </div>
        {#if gameMode === 'mexicano'}
          <p class="text-xs text-[var(--text-secondary)]">{$_('create_mexicano_hint')}</p>
        {/if}
      </div>

      {#if gameMode === 'mexicano'}
        <!-- Mexicano: rounds or time (mutually exclusive) -->
        <div class="space-y-2.5">
          <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('create_duration_label')}</p>
          <!-- Rounds row -->
          <div class="flex gap-2">
            {#each [4, 6, 8, 10] as n}
              <button
                onclick={() => pickRounds(n)}
                class="flex-1 rounded-full py-2.5 text-sm font-semibold transition-colors {mexicanoRounds === n
                  ? 'bg-[var(--primary)] text-white'
                  : 'bg-[var(--surface-raised)] text-[var(--text-primary)]'}"
              >{n} {$_('create_duration_rounds')}</button>
            {/each}
          </div>
          <!-- Divider -->
          <div class="flex items-center gap-3">
            <div class="h-px flex-1 bg-[var(--border)]"></div>
            <span class="text-[11px] font-semibold text-[var(--text-disabled)]">{$_('create_duration_or')}</span>
            <div class="h-px flex-1 bg-[var(--border)]"></div>
          </div>
          <!-- Time row -->
          <div class="flex gap-2">
            {#each [60, 90, 120] as min}
              <button
                onclick={() => pickTime(min)}
                class="flex-1 rounded-full py-2.5 text-sm font-semibold transition-colors {courtDuration === min && !customTimeMode
                  ? 'bg-[var(--primary)] text-white'
                  : 'bg-[var(--surface-raised)] text-[var(--text-primary)]'}"
              >{min}m</button>
            {/each}
            {#if customTimeMode}
              <div class="flex flex-1 items-center justify-center gap-0.5 rounded-full bg-[var(--primary)] px-2 py-2.5 text-sm font-semibold text-white">
                <input
                  bind:this={customInputEl}
                  type="number"
                  min="15"
                  max="300"
                  bind:value={customTimeRaw}
                  placeholder="90"
                  class="w-10 bg-transparent text-center font-semibold text-white placeholder-white/60 focus:outline-none [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
                />
                <span>m</span>
              </div>
            {:else}
              <button
                onclick={pickCustomTime}
                class="flex-1 rounded-full py-2.5 text-sm font-semibold transition-colors bg-[var(--surface-raised)] text-[var(--text-primary)]"
              >{$_('create_duration_custom')}</button>
            {/if}
          </div>
          <p class="text-xs text-[var(--text-secondary)]">
            {#if mexicanoRounds}
              {$_('create_mexicano_rounds_hint_fixed', { values: { n: mexicanoRounds } })}
            {:else if courtDuration}
              {$_('create_duration_hint', { values: { n: courtDuration } })}
            {:else}
              {$_('create_mexicano_rounds_hint_open')}
            {/if}
          </p>
        </div>
      {/if}

      {#if gameMode === 'americano' || gameMode === 'mexicano'}
        <!-- Courts -->
        <div class="space-y-2.5">
          <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('create_courts_label')}</p>
          <div class="flex gap-2">
            {#each (gameMode === 'mexicano' ? [2, 3, 4] : [1, 2, 3, 4]) as n}
              <button
                onclick={() => (courts = n)}
                class="flex-1 rounded-full py-2.5 text-sm font-semibold transition-colors {courts === n
                  ? 'bg-[var(--primary)] text-white'
                  : 'bg-[var(--surface-raised)] text-[var(--text-primary)]'}"
              >{n}</button>
            {/each}
          </div>
          {#if gameMode === 'mexicano'}
            <p class="text-xs text-[var(--text-secondary)]">{$_('create_mexicano_courts_hint', { values: { n: courts * 4 } })}</p>
          {/if}
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
            onclick={() => {
              scheduleEnabled = !scheduleEnabled;
              if (!scheduleEnabled) {
                calendarDate = undefined;
                timeSlot = 20;
              } else {
                calendarDate = today(getLocalTimeZone());
                timeSlot = calculateNextHourSlot();
              }
            }}
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

      <!-- Contacts picker -->
      {#if contacts.length > 0}
        <div class="space-y-2.5">
          <button
            onclick={() => showContacts = !showContacts}
            class="flex w-full items-center justify-between"
          >
            <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">
              {$_('create_contacts_invite_label')}
              {#if selectedContacts.size > 0}
                <span class="ml-1.5 rounded-full bg-[var(--primary)] px-1.5 py-0.5 text-[10px] text-white">{selectedContacts.size}</span>
              {/if}
            </p>
            <ChevronDown size={14} class="text-[var(--text-disabled)] transition-transform duration-200 {showContacts ? 'rotate-180' : ''}" />
          </button>

          {#if showContacts}
            <div transition:slide={{ duration: 200 }} class="space-y-1.5">
              {#each contacts as contact}
                {@const selected = selectedContacts.has(contact.user_id)}
                <button
                  onclick={() => toggleContact(contact.user_id)}
                  class="flex w-full items-center gap-3 rounded-2xl px-4 py-3 transition-colors
                    {selected ? 'bg-[var(--primary)]' : 'bg-[var(--surface-raised)]'}"
                >
                  <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-xs font-[800]
                    {selected ? 'bg-white/20 text-white' : 'bg-[var(--primary-muted)] text-[var(--primary)]'}">
                    {initials(contact.display_name)}
                  </div>
                  <span class="flex-1 text-left text-sm font-semibold {selected ? 'text-white' : ''}">{contact.display_name}</span>
                  {#if selected}
                    <Check size={16} class="shrink-0 text-white" />
                  {/if}
                </button>
              {/each}
            </div>
          {/if}
        </div>
      {/if}

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
