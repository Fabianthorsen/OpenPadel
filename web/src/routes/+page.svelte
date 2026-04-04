<script lang="ts">
  import { goto } from '$app/navigation';
  import { api } from '$lib/api/client';

  let step = $state<'home' | 'setup'>('home');
  let courts = $state(2);
  let points = $state(24);
  let name = $state('');
  let creating = $state(false);
  let error = $state('');
  let shaking = $state(false);

  function shake() {
    shaking = false;
    requestAnimationFrame(() => { shaking = true; });
    setTimeout(() => { shaking = false; }, 400);
  }
  let joinCode = $state('');

  async function create() {
    if (!name.trim()) { error = 'Enter your name'; shake(); return; }
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

  function joinByCode() {
    const code = joinCode.trim().toUpperCase();
    if (code) goto(`/s/${code}`);
  }
</script>

{#if step === 'home'}
  <main class="flex min-h-svh flex-col items-center px-6 py-12">
  <div class="flex w-full max-w-sm flex-1 flex-col">
    <div class="flex flex-1 flex-col justify-center space-y-12">
      <!-- Brand -->
      <div class="space-y-1">
        <h1 class="text-[28px] font-[800] text-[var(--primary)]">NotTennis</h1>
        <p class="text-[var(--text-secondary)]">Padel, organised.</p>
      </div>

      <!-- Actions -->
      <div class="space-y-4">
        <button
          onclick={() => (step = 'setup')}
          class="w-full rounded-2xl bg-[var(--primary)] px-4 py-4 text-[15px] font-semibold text-white transition-colors hover:bg-[var(--primary-hover)]"
        >
          Create Tournament →
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
            class="min-w-0 flex-1 rounded-2xl bg-[var(--surface-raised)] px-4 py-3.5 text-sm outline-none focus:ring-2 focus:ring-[var(--primary)]/20"
          />
          <button
            type="submit"
            disabled={!joinCode.trim()}
            class="rounded-2xl bg-[var(--surface-raised)] px-5 text-sm font-semibold text-[var(--text-primary)] transition-colors hover:bg-[var(--border)] disabled:opacity-40"
          >
            Join →
          </button>
        </form>
      </div>
    </div>

  </div>
  </main>

{:else}
  <main class="flex min-h-svh flex-col items-center px-6 py-6">
  <div class="w-full max-w-sm">
    <!-- Nav -->
    <nav class="flex items-center justify-between">
      <button
        onclick={() => (step = 'home')}
        class="flex h-8 w-8 items-center justify-center rounded-full text-lg text-[var(--text-secondary)] hover:bg-[var(--surface-raised)]"
      >
        ×
      </button>
      <span class="text-sm font-semibold text-[var(--primary)]">NotTennis</span>
      <div class="w-8"></div>
    </nav>

    <!-- Header -->
    <div class="mt-8 space-y-2">
      <h1 class="text-[34px] font-[800]">Create New<br />Tournament</h1>
      <p class="text-[var(--text-secondary)]">Set the stage for your next Padel session.</p>
    </div>

    <!-- Form -->
    <div class="mt-8 space-y-7">

      <!-- Game mode -->
      <div class="space-y-2.5">
        <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">Game Mode</p>
        <div class="flex flex-wrap gap-2">
          <span class="inline-flex items-center gap-1.5 rounded-full bg-[var(--primary)] px-4 py-2 text-sm font-semibold text-white">
            Americano
          </span>
          <span class="inline-flex items-center rounded-full bg-[var(--surface-raised)] px-4 py-2 text-sm text-[var(--text-disabled)]">
            Mexicano (Coming soon)
          </span>
        </div>
        <p class="text-xs text-[var(--text-secondary)]">Americano is currently the only supported mode.</p>
      </div>

      <!-- Courts -->
      <div class="space-y-2.5">
        <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">Courts</p>
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
        <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">Points per game</p>
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
          {points === 16 ? 'Quick format' : points === 24 ? 'Standard — recommended' : 'Long format'}
        </p>
      </div>

      <!-- Your name -->
      <div class="space-y-2.5">
        <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">Organiser name</p>
        <input
          bind:value={name}
          oninput={() => (error = '')}
          placeholder="Your name"
          maxlength="32"
          class="w-full rounded-2xl bg-[var(--surface-raised)] px-4 py-3.5 text-sm outline-none ring-2 transition-colors {error ? 'ring-[var(--destructive)]/60' : 'ring-transparent focus:ring-[var(--primary)]/30'} {shaking ? 'shake' : ''}"
        />
      </div>

      <!-- Info note -->
      <div class="flex gap-3 rounded-2xl bg-[var(--surface-raised)] px-4 py-3.5">
        <span class="mt-px shrink-0 text-[var(--text-secondary)]">ℹ</span>
        <p class="text-sm text-[var(--text-secondary)]">
          You'll be able to invite players after creating the session.
        </p>
      </div>

      <button
        onclick={create}
        disabled={creating}
        class="w-full rounded-2xl bg-[var(--primary)] px-4 py-4 text-[15px] font-semibold text-white transition-colors hover:bg-[var(--primary-hover)] disabled:opacity-50"
      >
        {creating ? 'Creating…' : 'Create & Get Invite Link →'}
      </button>
    </div>
  </div>
  </main>
{/if}
