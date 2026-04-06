<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api/client';
  import { _ } from 'svelte-i18n';

  let {
    session,
  }: {
    session: App.Session;
  } = $props();

  let match = $state<App.TennisMatch | null>(null);

  onMount(async () => {
    try { match = await api.tennis.getMatch(session.id); } catch {}
  });

  const state = $derived(match?.state);
  const winnerTeam = $derived(state?.winner === 'a' ? match?.teams.a : match?.teams.b);
  const finalistTeam = $derived(state?.winner === 'a' ? match?.teams.b : match?.teams.a);
  const winnerNames = $derived(winnerTeam?.map((t) => t.name).join(' & ') ?? '');
  const finalistNames = $derived(finalistTeam?.map((t) => t.name).join(' & ') ?? '');

  // Set scores from winner's perspective
  const sets = $derived(state?.sets ?? []);

  function winnerSetScore(i: number): number {
    const s = sets[i];
    if (!s) return 0;
    return state?.winner === 'a' ? s[0] : s[1];
  }
  function finalistSetScore(i: number): number {
    const s = sets[i];
    if (!s) return 0;
    return state?.winner === 'a' ? s[1] : s[0];
  }

  async function share() {
    const text = `${winnerNames} beat ${finalistNames} — ${sets.map((s, i) => `${winnerSetScore(i)}-${finalistSetScore(i)}`).join(', ')}`;
    if (navigator.share) {
      await navigator.share({ title: 'NotTennis', text }).catch(() => {});
    } else {
      await navigator.clipboard.writeText(text).catch(() => {});
    }
  }
</script>

<main class="mx-auto max-w-[480px] px-5 py-10 space-y-8">

  {#if !match}
    <div class="flex justify-center py-20">
      <div class="h-7 w-7 animate-spin rounded-full border-2 border-[var(--border)] border-t-[var(--primary)]"></div>
    </div>
  {:else}

    <!-- Header -->
    <div class="space-y-1">
      <p class="text-[11px] font-bold uppercase tracking-[0.15em] text-[var(--primary)]">{$_('tennis_result_label')}</p>
      <h1 class="text-4xl font-[900] leading-tight">
        {session.name ? $_('tennis_result_title', { values: { name: session.name } }) : $_('tennis_result_title_default')}
      </h1>
    </div>

    <!-- Score card -->
    <div class="rounded-2xl bg-[var(--surface-raised)] overflow-hidden">

      <!-- Column headers -->
      <div class="flex items-center gap-3 px-5 pt-4 pb-0">
        <div class="flex-1"></div>
        {#each sets as _set, i}
          <div class="w-14 text-center text-[10px] font-bold uppercase tracking-widest text-[var(--text-disabled)]">
            {$_('tennis_set')} {i + 1}
          </div>
        {/each}
      </div>

      <!-- Winners row -->
      <div class="flex items-center gap-3 px-5 py-4">
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-1.5 mb-0.5">
            <span class="text-[var(--primary)] text-sm">★</span>
            <p class="text-[10px] font-bold uppercase tracking-widest text-[var(--primary)]">{$_('tennis_result_winners')}</p>
          </div>
          <p class="text-lg font-[800] leading-snug">{winnerNames}</p>
        </div>
        {#each sets as _set, i}
          <div class="w-14 h-14 rounded-xl bg-[var(--primary)] flex items-center justify-center shrink-0">
            <p class="text-2xl font-[800] tabular-nums text-white">{winnerSetScore(i)}</p>
          </div>
        {/each}
      </div>

      <!-- Divider -->
      <div class="h-px bg-[var(--border)] mx-5"></div>

      <!-- Finalist row -->
      <div class="flex items-center gap-3 px-5 py-4">
        <div class="flex-1 min-w-0">
          <p class="text-[10px] font-bold uppercase tracking-widest text-[var(--text-disabled)] mb-0.5">{$_('tennis_result_finalist')}</p>
          <p class="text-lg font-[700] leading-snug text-[var(--text-secondary)]">{finalistNames}</p>
        </div>
        {#each sets as _set, i}
          <div class="w-14 h-14 rounded-xl bg-[var(--surface)] flex items-center justify-center shrink-0">
            <p class="text-2xl font-[800] tabular-nums text-[var(--text-disabled)]">{finalistSetScore(i)}</p>
          </div>
        {/each}
      </div>

    </div>

    <!-- Actions -->
    <div class="space-y-3">
      <button
        onclick={share}
        class="w-full rounded-2xl bg-[var(--primary)] py-4 flex items-center justify-center gap-2 text-white text-sm font-[800]"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/><line x1="8.59" y1="13.51" x2="15.42" y2="17.49"/><line x1="15.41" y1="6.51" x2="8.59" y2="10.49"/></svg>
        {$_('tennis_result_share')}
      </button>
      <a
        href="/"
        class="w-full rounded-2xl bg-[var(--surface-raised)] py-4 flex items-center justify-center gap-2 text-[var(--text-secondary)] text-sm font-[700]"
      >
        {$_('tennis_result_back')}
      </a>
    </div>

  {/if}

</main>
