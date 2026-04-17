<script lang="ts">
  import { goto } from '$app/navigation';
  import { api } from '$lib/api/client';
  import { auth } from '$lib/auth.svelte';
  import { _ } from 'svelte-i18n';
  import { initials } from '$lib/utils';
  import { ChevronDown, Check } from 'lucide-svelte';
  import { Calendar } from '$lib/components/ui/calendar';
  import { Input } from '$lib/components/ui/input';
  import { SectionLabel } from '$lib/components/ui/section-label';
  import { Switch } from '$lib/components/ui/switch';
  import { Separator } from '$lib/components/ui/separator';
  import { PillToggleGroup, PillToggleItem } from '$lib/components/ui/pill-toggle-group';
  import * as Collapsible from '$lib/components/ui/collapsible';
  import * as Drawer from '$lib/components/ui/drawer';
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

<Drawer.Root bind:open>
    <Drawer.Content class="flex flex-col overflow-hidden sm:data-[vaul-drawer-direction=bottom]:left-1/2 sm:data-[vaul-drawer-direction=bottom]:-translate-x-1/2 sm:data-[vaul-drawer-direction=bottom]:w-[480px] sm:data-[vaul-drawer-direction=bottom]:max-w-[480px] sm:data-[vaul-drawer-direction=bottom]:bottom-6">
      <Drawer.Header>
        <div class="flex items-center justify-between w-full">
          <h2 class="text-lg font-[800]">{$_('create_title_line1')} {$_('create_title_line2')}</h2>
          <Drawer.Close class="hidden md:flex h-8 w-8 items-center justify-center rounded-full bg-surface-raised text-text-secondary hover:bg-border transition-colors text-xl leading-none">×</Drawer.Close>
        </div>
      </Drawer.Header>

      <div class="flex-1 overflow-y-auto px-6 pb-8 space-y-6">

      <!-- Game mode -->
      <div class="space-y-2.5">
        <SectionLabel>{$_('create_game_mode_label')}</SectionLabel>
        <PillToggleGroup bind:value={gameMode}>
          <PillToggleItem value="americano">Americano</PillToggleItem>
          <PillToggleItem value="mexicano">Mexicano</PillToggleItem>
          <PillToggleItem value="tennis">{$_('create_mode_tennis')}</PillToggleItem>
        </PillToggleGroup>
        {#if gameMode === 'mexicano'}
          <p class="text-xs text-text-secondary">{$_('create_mexicano_hint')}</p>
        {/if}
      </div>

      {#if gameMode === 'mexicano'}
        <!-- Mexicano: rounds or time (mutually exclusive) -->
        <div class="space-y-2.5">
          <SectionLabel>{$_('create_duration_label')}</SectionLabel>
          <!-- Rounds row -->
          <PillToggleGroup
            value={mexicanoRounds?.toString() ?? ''}
            onValueChange={(val) => val && pickRounds(parseInt(val))}
          >
            {#each [4, 6, 8, 10] as n}
              <PillToggleItem value={n.toString()}>
                {n} {$_('create_duration_rounds')}
              </PillToggleItem>
            {/each}
          </PillToggleGroup>
          <!-- Divider -->
          <div class="flex items-center gap-3">
            <Separator class="flex-1" />
            <span class="text-[11px] font-semibold text-text-disabled">{$_('create_duration_or')}</span>
            <Separator class="flex-1" />
          </div>
          <!-- Time row -->
          <PillToggleGroup
            value={!customTimeMode && courtDuration ? courtDuration.toString() : customTimeMode ? 'custom' : ''}
            onValueChange={(val) => {
              if (val === 'custom') {
                pickCustomTime();
              } else if (val) {
                pickTime(parseInt(val));
              }
            }}
          >
            {#each [60, 90, 120] as min}
              <PillToggleItem value={min.toString()}>
                {min}m
              </PillToggleItem>
            {/each}
            {#if customTimeMode}
              <div class="flex flex-1 items-center justify-center gap-0.5 rounded-full bg-primary px-2 py-2.5 text-sm font-semibold text-white">
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
              <PillToggleItem value="custom">
                {$_('create_duration_custom')}
              </PillToggleItem>
            {/if}
          </PillToggleGroup>
          <p class="text-xs text-text-secondary">
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
          <SectionLabel>{$_('create_courts_label')}</SectionLabel>
          <PillToggleGroup
            value={courts.toString()}
            onValueChange={(val) => courts = parseInt(val)}
          >
            {#each (gameMode === 'mexicano' ? [2, 3, 4] : [1, 2, 3, 4]) as n}
              <PillToggleItem value={n.toString()}>
                {n}
              </PillToggleItem>
            {/each}
          </PillToggleGroup>
          {#if gameMode === 'mexicano'}
            <p class="text-xs text-text-secondary">{$_('create_mexicano_courts_hint', { values: { n: courts * 4 } })}</p>
          {/if}
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
        <!-- Sets to win -->
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

        <!-- Games per set -->
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

      <!-- Tournament name -->
      <div class="space-y-2.5">
        <SectionLabel>{$_('create_tournament_name_label')}</SectionLabel>
        <Input
          bind:value={tournamentName}
          placeholder={$_('create_tournament_name_placeholder')}
          maxlength={48}
          class="rounded-2xl border-0 bg-surface-raised px-4 py-3.5 text-sm"
        />
      </div>

      <!-- Schedule -->
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
              <input type="range" min="0" max="27" step="1" bind:value={timeSlot} class="w-full accent-primary" />
              <div class="flex justify-between text-[10px] text-text-disabled">
                <span>08:00</span><span>21:30</span>
              </div>
            </div>
          </div>
        {/if}
      </div>

      <!-- Contacts picker -->
      {#if contacts.length > 0}
        <Collapsible.Root bind:open={showContacts} class="space-y-2.5">
          <Collapsible.Trigger class="flex w-full items-center justify-between">
            <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-text-secondary">
              {$_('create_contacts_invite_label')}
              {#if selectedContacts.size > 0}
                <span class="ml-1.5 rounded-full bg-primary px-1.5 py-0.5 text-[10px] text-white">{selectedContacts.size}</span>
              {/if}
            </p>
            <ChevronDown size={14} class="text-text-disabled transition-transform duration-200 data-[state=open]:rotate-180" />
          </Collapsible.Trigger>

          <Collapsible.Content class="space-y-1.5">
              {#each contacts as contact}
                {@const selected = selectedContacts.has(contact.user_id)}
                <button
                  onclick={() => toggleContact(contact.user_id)}
                  class="flex w-full items-center gap-3 rounded-2xl px-4 py-3 transition-colors
                    {selected ? 'bg-primary' : 'bg-surface-raised'}"
                >
                  <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-xs font-[800]
                    {selected ? 'bg-white/20 text-white' : 'bg-primary-muted text-primary'}">
                    {initials(contact.display_name)}
                  </div>
                  <span class="flex-1 text-left text-sm font-semibold {selected ? 'text-white' : ''}">{contact.display_name}</span>
                  {#if selected}
                    <Check size={16} class="shrink-0 text-white" />
                  {/if}
                </button>
              {/each}
          </Collapsible.Content>
        </Collapsible.Root>
      {/if}

{#if error}
        <p class="text-sm text-destructive">{error}</p>
      {/if}

      <button
        onclick={create}
        disabled={creating}
        class="w-full rounded-2xl bg-primary px-4 py-4 text-[15px] font-semibold text-white disabled:opacity-60 hover:bg-primary-hover transition-colors"
      >
        {creating ? $_('create_button_loading') : $_('create_button')}
      </button>

      </div>
    </Drawer.Content>
  </Drawer.Root>
