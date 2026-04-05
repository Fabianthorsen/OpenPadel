<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { auth } from '$lib/auth.svelte';
  import { api } from '$lib/api/client';
  import { _ } from 'svelte-i18n';
  import { initials } from '$lib/utils';

  let stats = $state<App.CareerStats | null>(null);
  let tournaments = $state<App.TournamentEntry[]>([]);
  let loading = $state(true);
  let showDeleteConfirm = $state(false);
  let deleting = $state(false);

  onMount(async () => {
    if (!auth.token) { goto('/auth'); return; }
    try {
      const [profileRes, historyRes] = await Promise.all([
        api.auth.profile(auth.token),
        api.auth.history(auth.token),
      ]);
      stats = profileRes.stats;
      tournaments = historyRes.tournaments;
    } finally {
      loading = false;
    }
  });

  const winRate = $derived(
    stats && stats.games_played > 0
      ? Math.round((stats.wins / stats.games_played) * 100)
      : 0
  );

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

<main class="mx-auto max-w-[480px] px-6 py-10 space-y-8">

  <!-- Header -->
  <div class="flex items-center gap-4">
    <div class="flex h-16 w-16 shrink-0 items-center justify-center rounded-full bg-[var(--primary)] text-2xl font-[800] text-white">
      {auth.user ? initials(auth.user.display_name) : '?'}
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

    <!-- Tournament history -->
    <div class="space-y-3">
      <p class="text-[11px] font-bold uppercase tracking-[0.1em] text-[var(--text-disabled)]">{$_('profile_history_label')}</p>

      {#if tournaments.length === 0}
        <p class="text-sm text-[var(--text-disabled)] py-4 text-center">{$_('profile_history_empty')}</p>
      {:else}
        <div class="space-y-2">
          {#each tournaments as t}
            <a
              href="/s/{t.session_id}"
              class="flex items-center gap-4 rounded-2xl bg-[var(--surface-raised)] px-4 py-3.5 transition-colors hover:bg-[var(--border)]"
            >
              <!-- Rank badge -->
              <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full {t.rank === 1 ? 'bg-[var(--primary)] text-white' : 'bg-[var(--border)] text-[var(--text-secondary)]'} text-xs font-[800]">
                {ordinal(t.rank)}
              </div>
              <!-- Info -->
              <div class="flex-1 min-w-0">
                <p class="truncate font-semibold text-sm">{t.name}</p>
                <p class="text-xs text-[var(--text-secondary)]">{formatDate(t.played_at)} · {t.games_played} {$_('profile_history_games')} · {t.points} pts</p>
              </div>
              <span class="text-sm text-[var(--text-secondary)]">→</span>
            </a>
          {/each}
        </div>
      {/if}
    </div>

  {/if}

  <!-- Actions -->
  <div class="space-y-3 pt-2">
    <a
      href="/?create=1"
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
    <button
      onclick={() => (showDeleteConfirm = true)}
      class="w-full px-4 py-2 text-sm text-[var(--text-disabled)] hover:text-[var(--destructive)] transition-colors"
    >
      {$_('profile_delete_account')}
    </button>
  </div>

</main>

<!-- Delete account confirmation dialog -->
{#if showDeleteConfirm}
  <div class="fixed inset-0 z-50 flex items-end justify-center bg-black/40 p-4 sm:items-center">
    <div class="w-full max-w-sm rounded-3xl bg-[var(--surface)] p-6 space-y-4 shadow-xl">
      <div class="space-y-1">
        <h2 class="text-lg font-[800]">{$_('profile_delete_title')}</h2>
        <p class="text-sm text-[var(--text-secondary)]">{$_('profile_delete_desc')}</p>
      </div>
      <div class="space-y-2 pt-1">
        <button
          onclick={deleteAccount}
          disabled={deleting}
          class="w-full rounded-2xl bg-[var(--destructive)] px-4 py-3.5 text-sm font-semibold text-white disabled:opacity-60"
        >
          {deleting ? $_('profile_delete_loading') : $_('profile_delete_confirm')}
        </button>
        <button
          onclick={() => (showDeleteConfirm = false)}
          class="w-full rounded-2xl border border-[var(--border)] px-4 py-3.5 text-sm font-semibold text-[var(--text-secondary)]"
        >
          {$_('profile_delete_cancel')}
        </button>
      </div>
    </div>
  </div>
{/if}
