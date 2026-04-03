<script lang="ts">
  import { goto } from '$app/navigation';
  import { api } from '$lib/api/client';

  let step = $state<'home' | 'setup'>('home');
  let courts = $state(2);
  let points = $state(24);
  let name = $state('');
  let creating = $state(false);
  let error = $state('');
  let joinCode = $state('');

  function joinByCode() {
    const code = joinCode.trim().toUpperCase();
    if (code) goto(`/s/${code}`);
  }

  async function create() {
    if (!name.trim()) { error = 'Enter your name'; return; }
    creating = true;
    error = '';
    try {
      const session = await api.sessions.create(courts, points);
      const token = session.admin_token!;
      localStorage.setItem(`admin_token_${session.id}`, token);
      const player = await api.players.join(session.id, name.trim(), token);
      localStorage.setItem(`player_id_${session.id}`, player.id);
      goto(`/s/${session.id}?token=${token}`);
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

      <div class="space-y-4">
        <button
          onclick={() => (step = 'setup')}
          class="w-full rounded-lg bg-[var(--primary)] px-4 py-3 text-[15px] font-semibold text-white transition-colors hover:bg-[var(--primary-hover)]"
        >
          Create session
        </button>

        <div class="flex items-center gap-3">
          <div class="h-px flex-1 bg-[var(--border)]"></div>
          <span class="text-xs text-[var(--text-disabled)]">or join with a code</span>
          <div class="h-px flex-1 bg-[var(--border)]"></div>
        </div>

        <form onsubmit={(e) => { e.preventDefault(); joinByCode(); }} class="flex gap-2">
          <input
            bind:value={joinCode}
            oninput={(e) => { joinCode = (e.currentTarget as HTMLInputElement).value.toUpperCase(); }}
            placeholder="Session code"
            maxlength="4"
            autocomplete="off"
            autocorrect="off"
            autocapitalize="characters"
            spellcheck={false}
            class="min-w-0 flex-1 rounded-lg border border-[var(--border)] bg-[var(--surface)] px-3 py-2.5 text-sm outline-none focus:border-[var(--border-strong)] focus:ring-2 focus:ring-[var(--primary)]/20"
          />
          <button
            type="submit"
            disabled={!joinCode.trim()}
            class="rounded-lg border border-[var(--border)] bg-[var(--surface)] px-4 py-2.5 text-sm font-semibold text-[var(--text-primary)] transition-colors hover:bg-[var(--surface-raised)] disabled:opacity-40"
          >
            Join →
          </button>
        </form>
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
          4 players per court · {courts * 4} playing each round · {points} pts per game
        </div>

        <!-- Your name -->
        <div class="space-y-2">
          <p class="text-sm font-medium text-[var(--text-primary)]">Your name</p>
          <input
            bind:value={name}
            placeholder="e.g. Fabian"
            maxlength="32"
            class="w-full rounded-lg border border-[var(--border)] bg-[var(--surface)] px-3 py-2.5 text-sm outline-none focus:border-[var(--border-strong)] focus:ring-2 focus:ring-[var(--primary)]/20"
          />
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
