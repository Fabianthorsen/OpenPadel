<script lang="ts">
  import { goto } from '$app/navigation';
  import { api } from '$lib/api/client';

  let step = $state<'home' | 'setup'>('home');
  let courts = $state(2);
  let points = $state(24);
  let creating = $state(false);
  let error = $state('');

  async function create() {
    creating = true;
    error = '';
    try {
      const session = await api.sessions.create(courts, points);
      localStorage.setItem(`admin_token_${session.id}`, session.admin_token!);
      goto(`/s/${session.id}?token=${session.admin_token}`);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Something went wrong';
      creating = false;
    }
  }
</script>

<main class="flex min-h-svh flex-col items-center justify-center px-4">
  <div class="w-full max-w-sm space-y-8">

    {#if step === 'home'}
      <!-- Landing -->
      <div class="space-y-1">
        <h1 class="text-[26px] font-[650] text-[var(--text-primary)]">NotTennis</h1>
        <p class="text-sm text-[var(--text-secondary)]">Padel, organised.</p>
      </div>

      <div class="space-y-3">
        <button
          onclick={() => (step = 'setup')}
          class="w-full rounded-lg bg-[var(--primary)] px-4 py-3 text-[15px] font-semibold text-white transition-colors hover:bg-[var(--primary-hover)]"
        >
          Create session
        </button>
        <p class="text-center text-sm text-[var(--text-secondary)]">
          Have a link?
          <a href="#join" class="text-[var(--primary)] underline-offset-2 hover:underline">
            Join a session →
          </a>
        </p>
      </div>

    {:else}
      <!-- Setup -->
      <div class="space-y-1">
        <button
          onclick={() => (step = 'home')}
          class="mb-1 text-sm text-[var(--text-secondary)] hover:text-[var(--text-primary)]"
        >
          ← Back
        </button>
        <h1 class="text-[22px] font-[650]">New session</h1>
        <p class="text-sm text-[var(--text-secondary)]">
          Players join after you share the link.
        </p>
      </div>

      <div class="space-y-5">
        <!-- Courts -->
        <div class="space-y-2">
          <p class="text-sm font-medium text-[var(--text-primary)]">Courts</p>
          <div class="grid grid-cols-4 gap-2">
            {#each [1, 2, 3, 4] as n}
              <button
                onclick={() => (courts = n)}
                class="rounded-lg border py-2.5 text-sm font-semibold transition-colors {courts === n
                  ? 'border-[var(--primary)] bg-[var(--primary-muted)] text-[var(--primary)]'
                  : 'border-[var(--border)] bg-[var(--surface)] text-[var(--text-primary)] hover:bg-[var(--surface-raised)]'}"
              >
                {n}
              </button>
            {/each}
          </div>
        </div>

        <!-- Points -->
        <div class="space-y-2">
          <p class="text-sm font-medium text-[var(--text-primary)]">Points per game</p>
          <div class="grid grid-cols-3 gap-2">
            {#each [16, 24, 32] as p}
              <button
                onclick={() => (points = p)}
                class="rounded-lg border py-2.5 text-sm font-semibold transition-colors {points === p
                  ? 'border-[var(--primary)] bg-[var(--primary-muted)] text-[var(--primary)]'
                  : 'border-[var(--border)] bg-[var(--surface)] text-[var(--text-primary)] hover:bg-[var(--surface-raised)]'}"
              >
                {p}
              </button>
            {/each}
          </div>
          <p class="text-xs text-[var(--text-disabled)]">
            {points === 16 ? 'Quick format' : points === 24 ? 'Standard — recommended' : 'Long format'}
          </p>
        </div>

        <!-- Summary -->
        <div class="rounded-lg bg-[var(--surface-raised)] px-4 py-3 text-sm text-[var(--text-secondary)]">
          {courts} {courts === 1 ? 'court' : 'courts'} · {courts * 4} players on court per round
          · {points} pts per game
        </div>

        {#if error}
          <p class="text-sm text-[var(--destructive)]">{error}</p>
        {/if}

        <button
          onclick={create}
          disabled={creating}
          class="w-full rounded-lg bg-[var(--primary)] px-4 py-3 text-[15px] font-semibold text-white transition-colors hover:bg-[var(--primary-hover)] disabled:opacity-50"
        >
          {creating ? 'Creating…' : 'Create & get link →'}
        </button>
      </div>
    {/if}

  </div>
</main>
