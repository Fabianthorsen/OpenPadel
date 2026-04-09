<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { api } from '$lib/api/client';
  import { _ } from 'svelte-i18n';
  import { Trophy, UserPlus, Check } from 'lucide-svelte';
  import { shortName } from '$lib/utils';
  import Avatar from '$lib/components/ui/Avatar.svelte';
  import { auth } from '$lib/auth.svelte';

  let {
    sessionId,
    sessionName = '',
    complete = false,
  }: {
    sessionId: string;
    sessionName?: string;
    complete?: boolean;
  } = $props();

  let leaderboard = $state<App.Leaderboard | null>(null);
  let interval: ReturnType<typeof setInterval>;
  // user_id → true once added this session
  let addedContacts = $state<Record<string, boolean>>({});
  // user_id → true if already a contact before arriving
  let existingContacts = $state<Record<string, boolean>>({});

  async function load() {
    try {
      leaderboard = await api.leaderboard.get(sessionId);
    } catch {
      // silently retry on next interval
    }
  }

  async function loadContacts() {
    if (!auth.token) return;
    const contacts = await api.contacts.list(auth.token);
    existingContacts = Object.fromEntries(contacts.map(c => [c.user_id, true]));
  }

  async function addContact(userID: string) {
    if (!auth.token) return;
    await api.contacts.add(auth.token, userID);
    addedContacts = { ...addedContacts, [userID]: true };
  }

  onMount(() => {
    load();
    if (!complete) interval = setInterval(load, 15_000);
    if (complete) loadContacts();
  });

  onDestroy(() => clearInterval(interval));


  const leader = $derived(leaderboard?.standings[0] ?? null);

  const podiumOrder = $derived(
    leaderboard ? [
      leaderboard.standings[1], // 2nd — left
      leaderboard.standings[0], // 1st — centre
      leaderboard.standings[2], // 3rd — right
    ].filter(Boolean) : []
  );

</script>

<main class="mx-auto max-w-[480px] px-4 pb-24 pt-4 space-y-6">
  {#if !leaderboard}
    <p class="text-sm text-[var(--text-secondary)]">Loading…</p>

  {:else if complete}

    <!-- ── Final Results ── -->

    <!-- Heading -->
    <div class="pt-4 text-center space-y-0.5">
      <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-[var(--text-disabled)]">{$_('leaderboard_final')}</p>
      {#if sessionName}
        <p class="text-xl font-[800]">{sessionName}</p>
      {/if}
    </div>

    <!-- Podium -->
    <div class="flex items-end justify-center gap-3 pt-6 pb-2">
      {#each podiumOrder as s}
        {@const isFirst = s.rank === 1}
        <div class="flex flex-col items-center {isFirst ? 'order-2 -mb-0' : s.rank === 2 ? 'order-1' : 'order-3'} flex-1 max-w-[120px]">

          <!-- Trophy for winner -->
          {#if isFirst}
            <div class="mb-1 text-[var(--primary)]"><Trophy size={28} /></div>
          {/if}

          <!-- Avatar -->
          <div class="{isFirst ? 'ring-4 ring-[var(--primary-muted)] shadow-lg rounded-full' : 'ring-2 ring-[var(--border)] rounded-full'}">
            <Avatar icon={s.avatar_icon} color={s.avatar_color} name={s.name} size={isFirst ? 'xl' : 'lg'} />
          </div>

          <!-- Rank badge -->
          <div class="mt-2 flex h-6 w-6 items-center justify-center rounded-full text-xs font-[800] text-white
            {isFirst ? 'bg-[var(--primary)]' : 'bg-[#4a7856]'}">
            {s.rank}
          </div>

          <p class="mt-1.5 text-sm font-[800] text-center truncate w-full {isFirst ? 'text-[var(--text-primary)]' : 'text-[var(--text-secondary)]'}">{shortName(s.name)}</p>
          <p class="text-[10px] font-bold uppercase tracking-widest {isFirst ? 'text-[var(--primary)]' : 'text-[var(--text-disabled)]'}">{s.points} {$_('leaderboard_pts')}</p>
          <div class="mt-1 flex items-center gap-1.5 text-[11px] font-bold tabular-nums">
            <span class="text-[var(--primary)]">{s.wins ?? 0}W</span>
            <span class="text-[var(--text-disabled)]">·</span>
            <span class="text-[var(--text-disabled)]">{s.draws ?? 0}D</span>
            <span class="text-[var(--text-disabled)]">·</span>
            <span class="text-[#c0392b]">{(s.games_played ?? 0) - (s.wins ?? 0) - (s.draws ?? 0)}L</span>
          </div>
          {#if auth.token && s.user_id && s.user_id !== auth.user?.id}
            {@const isContact = existingContacts[s.user_id] || addedContacts[s.user_id]}
            <button
              onclick={() => !isContact && addContact(s.user_id!)}
              disabled={isContact}
              class="mt-2 flex items-center gap-1 rounded-full px-2.5 py-1 text-[10px] font-bold transition-colors
                {isContact ? 'bg-[var(--border)] text-[var(--text-disabled)] cursor-default' : 'bg-[var(--primary-muted)] text-[var(--primary)] hover:bg-[var(--primary)] hover:text-white'}"
            >
              {#if isContact}<Check size={10} />{:else}<UserPlus size={10} />{/if}
              {isContact ? 'Added' : 'Add'}
            </button>
          {/if}

          <!-- Podium bar -->
          <div class="mt-3 w-full rounded-t-xl
            {isFirst ? 'h-12 bg-[var(--primary)]' : s.rank === 2 ? 'h-8 bg-[#4a7856]/60' : 'h-5 bg-[#a8c5b0]/60'}">
          </div>
        </div>
      {/each}
    </div>

    <!-- Rest of standings (4th+) -->
    {#if leaderboard.standings.length > 3}
      <div class="space-y-1">
        <p class="px-1 text-[11px] font-bold uppercase tracking-[0.1em] text-[var(--text-disabled)]">{$_('leaderboard_ranking')}</p>
        {#each leaderboard.standings.slice(3) as s (s.player_id)}
          {@const isContact = existingContacts[s.user_id ?? ''] || addedContacts[s.user_id ?? '']}
          <div class="flex items-center gap-3 rounded-2xl bg-[var(--surface-raised)] px-4 py-3">
            <span class="w-6 text-sm font-[800] tabular-nums text-[var(--text-disabled)]">{s.rank}</span>
            <Avatar icon={s.avatar_icon} color={s.avatar_color} name={s.name} size="sm" ring="ring-2 ring-[var(--primary)]/30" />
            <span class="flex-1 truncate text-sm font-semibold">{shortName(s.name)}</span>
            <div class="flex items-center gap-1 text-[11px] font-bold tabular-nums">
              <span class="text-[var(--primary)]">{s.wins ?? 0}W</span>
              <span class="text-[var(--text-disabled)]">·</span>
              <span class="text-[var(--text-disabled)]">{s.draws ?? 0}D</span>
              <span class="text-[var(--text-disabled)]">·</span>
              <span class="text-[#c0392b]">{(s.games_played ?? 0) - (s.wins ?? 0) - (s.draws ?? 0)}L</span>
            </div>
            <span class="text-base font-[800] tabular-nums">{s.points}</span>
            <span class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-disabled)]">{$_('leaderboard_pts')}</span>
            {#if auth.token && s.user_id && s.user_id !== auth.user?.id}
              <button
                onclick={() => !isContact && addContact(s.user_id!)}
                disabled={isContact}
                class="flex items-center gap-1 rounded-full px-2.5 py-1 text-[10px] font-bold transition-colors
                  {isContact ? 'bg-[var(--border)] text-[var(--text-disabled)] cursor-default' : 'bg-[var(--primary-muted)] text-[var(--primary)] hover:bg-[var(--primary)] hover:text-white'}"
              >
                {#if isContact}<Check size={10} />{:else}<UserPlus size={10} />{/if}
              </button>
            {/if}
          </div>
        {/each}
      </div>
    {/if}


    <div class="flex justify-center">
      <a
        href="/"
        class="rounded-full border border-[var(--border)] px-5 py-2 text-sm text-[var(--text-secondary)] hover:border-[var(--text-secondary)] hover:text-[var(--text-primary)] transition-colors"
      >
        ✕ {$_('leaderboard_close')}
      </a>
    </div>


  {:else}

    <!-- ── Live Standings ── -->

    <!-- Leader hero card -->
    {#if leader}
      <div class="relative overflow-hidden rounded-2xl bg-[var(--primary)] px-6 py-6">
        <svg class="absolute inset-0 h-full w-full opacity-10" preserveAspectRatio="none" viewBox="0 0 100 100">
          <line x1="50" y1="0" x2="50" y2="100" stroke="white" stroke-width="0.5"/>
          <line x1="0" y1="50" x2="100" y2="50" stroke="white" stroke-width="0.5"/>
          <rect x="20" y="20" width="60" height="60" fill="none" stroke="white" stroke-width="0.5"/>
        </svg>
        <div class="relative z-10 flex items-center gap-5">
          <Avatar icon={leader.avatar_icon} color={leader.avatar_color} name={leader.name} size="lg" ring="ring-2 ring-white/30" />
          <div class="flex-1 min-w-0">
            <div class="mb-0.5">
              <span class="rounded-full bg-white/20 px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-widest text-white">
                {$_('leaderboard_leader')}
              </span>
            </div>
            <p class="text-2xl font-[800] text-white truncate">{leader.name}</p>
            <div class="mt-2 flex items-center gap-4">
              <div>
                <span class="text-xl font-[800] text-white">{leader.points}</span>
                <span class="ml-1 text-[10px] font-bold uppercase tracking-wider text-white/60">{$_('leaderboard_pts')}</span>
              </div>
              {#if (leader.games_played ?? 0) > 0}
                <div class="h-6 w-px bg-white/20"></div>
                <div>
                  <span class="text-xl font-[800] text-white">{leader.wins ?? 0}/{leader.draws ?? 0}/{(leader.games_played ?? 0) - (leader.wins ?? 0) - (leader.draws ?? 0)}</span>
                  <span class="ml-1 text-[10px] font-bold uppercase tracking-wider text-white/60">{$_('leaderboard_wl')}</span>
                </div>
              {/if}
            </div>
          </div>
        </div>
      </div>
    {/if}

    <!-- Standings -->
    <div class="space-y-1">
      <div class="flex items-center justify-between px-1 pb-1">
        <h3 class="text-[13px] font-bold uppercase tracking-[0.1em] text-[var(--text-secondary)]">
          {$_('leaderboard_current')}
        </h3>
        {#if leaderboard.current_round}
          <span class="text-xs text-[var(--text-disabled)]">
            {leaderboard.total_rounds
              ? $_('leaderboard_round_of', { values: { current: leaderboard.current_round, total: leaderboard.total_rounds } })
              : $_('active_round_open', { values: { current: leaderboard.current_round } })}
          </span>
        {/if}
      </div>

      <div class="grid grid-cols-[2rem_1fr_3rem_3.5rem_3rem] gap-2 px-4 pb-1">
        <span class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-disabled)]">#</span>
        <span class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-disabled)]">{$_('leaderboard_player')}</span>
        <span class="text-center text-[10px] font-bold uppercase tracking-widest text-[var(--text-disabled)]">{$_('leaderboard_games')}</span>
        <span class="text-center text-[10px] font-bold uppercase tracking-widest text-[var(--text-disabled)]">{$_('leaderboard_wl')}</span>
        <span class="text-right text-[10px] font-bold uppercase tracking-widest text-[var(--text-disabled)]">{$_('leaderboard_pts')}</span>
      </div>

      {#each leaderboard.standings as s, i (s.player_id)}
        {@const podiumBg = s.rank === 1 ? 'bg-[var(--primary)]' : s.rank === 2 ? 'bg-[#4a7856]' : s.rank === 3 ? 'bg-[#a8c5b0]' : i % 2 === 0 ? 'bg-[var(--surface-raised)]' : 'bg-transparent'}
        {@const isPodium = s.rank <= 3}
        <div class="grid grid-cols-[2rem_1fr_3rem_3.5rem_3rem] items-center gap-2 rounded-2xl px-4 py-3.5 {podiumBg}">
          <span class="text-sm font-[800] tabular-nums {isPodium ? 'text-white' : 'text-[var(--text-disabled)]'}">{s.rank}</span>
          <div class="flex items-center gap-2.5 min-w-0">
            <Avatar icon={s.avatar_icon} color={isPodium ? 'white' : s.avatar_color} name={s.name} size="sm" ring={isPodium ? 'ring-2 ring-white/30' : 'ring-2 ring-[var(--primary)]/30'} />
            <span class="truncate text-sm font-semibold {isPodium ? 'text-white' : ''}">{shortName(s.name)}</span>
          </div>
          <span class="text-center text-sm {isPodium ? 'text-white/70' : 'text-[var(--text-secondary)]'}">{s.games_played ?? 0}</span>
          <span class="text-center text-sm font-semibold {isPodium ? 'text-white/70' : 'text-[var(--text-secondary)]'}">
            {s.wins ?? 0}/{s.draws ?? 0}/{(s.games_played ?? 0) - (s.wins ?? 0) - (s.draws ?? 0)}
          </span>
          <span class="text-right text-base font-[800] tabular-nums {isPodium ? 'text-white' : ''}">{s.points}</span>
        </div>
      {/each}
    </div>

  {/if}
</main>
