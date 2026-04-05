<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { auth } from '$lib/auth.svelte';
  import { api } from '$lib/api/client';
  import { _ } from 'svelte-i18n';

  let stats = $state<App.CareerStats | null>(null);
  let loading = $state(true);

  onMount(async () => {
    if (!auth.token) { goto('/auth'); return; }
    try {
      const res = await api.auth.profile(auth.token);
      stats = res.stats;
    } finally {
      loading = false;
    }
  });

  const winRate = $derived(
    stats && stats.games_played > 0
      ? Math.round((stats.wins / stats.games_played) * 100)
      : 0
  );
</script>

<main class="mx-auto max-w-[480px] px-6 py-10 space-y-8">

  <!-- Header -->
  <div class="flex items-center gap-4">
    <div class="flex h-16 w-16 shrink-0 items-center justify-center rounded-full bg-[var(--primary)] text-2xl font-[800] text-white">
      {auth.user?.display_name[0].toUpperCase() ?? '?'}
    </div>
    <div class="min-w-0">
      <h1 class="text-2xl font-[800] truncate">{auth.user?.display_name}</h1>
      <p class="text-sm text-[var(--text-secondary)] truncate">{auth.user?.email}</p>
    </div>
  </div>

  {#if loading}
    <div class="flex justify-center py-12">
      <div class="h-7 w-7 animate-spin rounded-full border-2 border-[var(--border)] border-t-[var(--primary)]"></div>
    </div>

  {:else if stats}

    <!-- Tournaments played -->
    <div class="rounded-2xl bg-[var(--primary)] px-6 py-5">
      <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-white/60">{$_('profile_tournaments')}</p>
      <p class="mt-1 text-5xl font-[800] text-white">{stats.tournaments}</p>
    </div>

    <!-- Stat grid -->
    <div class="grid grid-cols-2 gap-3">
      <div class="rounded-2xl bg-[var(--surface-raised)] px-5 py-4 space-y-0.5">
        <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-[var(--text-disabled)]">{$_('profile_games')}</p>
        <p class="text-3xl font-[800]">{stats.games_played}</p>
      </div>

      <div class="rounded-2xl bg-[var(--surface-raised)] px-5 py-4 space-y-0.5">
        <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-[var(--text-disabled)]">{$_('profile_win_rate')}</p>
        <p class="text-3xl font-[800]">{winRate}<span class="text-lg text-[var(--text-secondary)]">%</span></p>
      </div>

      <div class="rounded-2xl bg-[var(--surface-raised)] px-5 py-4 space-y-0.5">
        <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-[var(--text-disabled)]">{$_('profile_total_points')}</p>
        <p class="text-3xl font-[800]">{stats.total_points}</p>
      </div>

      <div class="rounded-2xl bg-[var(--surface-raised)] px-5 py-4 space-y-0.5">
        <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-[var(--text-disabled)]">{$_('leaderboard_wl')}</p>
        <p class="text-2xl font-[800]">{stats.wins}<span class="text-[var(--text-disabled)]">/</span>{stats.draws}<span class="text-[var(--text-disabled)]">/</span>{stats.losses}</p>
      </div>
    </div>

    {#if stats.games_played === 0}
      <p class="text-center text-sm text-[var(--text-disabled)]">{$_('profile_no_games')}</p>
    {/if}

  {/if}

  <!-- Actions -->
  <div class="space-y-3 pt-2">
    <a
      href="/"
      class="block w-full rounded-2xl bg-[var(--primary)] px-4 py-4 text-center text-[15px] font-semibold text-white"
    >
      {$_('profile_new_tournament')}
    </a>
    <button
      onclick={() => auth.logout().then(() => goto('/'))}
      class="w-full rounded-2xl border border-[var(--border)] px-4 py-3.5 text-sm font-semibold text-[var(--text-secondary)] transition-colors hover:border-[var(--destructive)] hover:text-[var(--destructive)]"
    >
      {$_('auth_sign_out')}
    </button>
  </div>

</main>
