<script lang="ts">
  import { goto } from '$app/navigation';
  import { api } from '$lib/api/client';

  let creating = $state(false);
  let error = $state('');

  async function startSession() {
    creating = true;
    error = '';
    try {
      const session = await api.sessions.create(2, 24);
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
    <div class="space-y-1">
      <h1 class="text-[26px] font-[650] text-[var(--text-primary)]">NotTennis</h1>
      <p class="text-sm text-[var(--text-secondary)]">Padel, organised.</p>
    </div>

    <div class="space-y-3">
      <button
        onclick={startSession}
        disabled={creating}
        class="w-full rounded-lg bg-[var(--primary)] px-4 py-3 text-[15px] font-semibold text-white transition-colors hover:bg-[var(--primary-hover)] disabled:opacity-60"
      >
        {creating ? 'Creating…' : 'Start session'}
      </button>

      <p class="text-center text-sm text-[var(--text-secondary)]">
        Have a link? <a href="/join" class="text-[var(--primary)] underline-offset-2 hover:underline"
          >Join a session →</a
        >
      </p>
    </div>

    {#if error}
      <p class="text-sm text-[var(--destructive)]">{error}</p>
    {/if}
  </div>
</main>
