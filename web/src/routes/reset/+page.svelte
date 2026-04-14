<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/state';
  import { api } from '$lib/api/client';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import { _ } from 'svelte-i18n';

  const token = page.url.searchParams.get('token') ?? '';

  let password = $state('');
  let loading = $state(false);
  let error = $state('');

  async function submit() {
    error = '';
    loading = true;
    try {
      await api.auth.resetPassword(token, password);
      goto('/auth?reset=1');
    } catch (e) {
      error = e instanceof Error ? e.message : 'Something went wrong';
    } finally {
      loading = false;
    }
  }
</script>

<main class="flex min-h-svh flex-col items-center px-6 pb-12 pt-safe-page">
  <div class="flex w-full max-w-sm flex-1 flex-col justify-center space-y-8">

    <div class="space-y-1">
      <h1 class="text-[28px] font-[800] text-[var(--primary)]">OpenPadel</h1>
      <p class="text-[var(--text-secondary)]">{$_('reset_subtitle')}</p>
    </div>

    {#if !token}
      <p class="text-sm text-[var(--destructive)]">{$_('reset_invalid_link')}</p>
    {:else}
      <form onsubmit={(e) => { e.preventDefault(); submit(); }} class="space-y-4">
        <div class="space-y-2">
          <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('reset_new_password_label')}</p>
          <Input
            bind:value={password}
            type="password"
            placeholder={$_('auth_password_placeholder')}
            autocomplete="new-password"
            class="rounded-2xl border-0 bg-[var(--surface-raised)] px-4 py-3.5 text-sm"
          />
        </div>

        {#if error}
          <p class="text-sm text-[var(--destructive)]">{error}</p>
        {/if}

        <Button
          type="submit"
          disabled={loading || password.length < 8}
          class="h-auto w-full rounded-2xl bg-[var(--primary)] px-4 py-4 text-[15px] font-semibold text-white hover:bg-[var(--primary-hover)]"
        >
          {loading ? $_('reset_button_loading') : $_('reset_button')}
        </Button>
      </form>
    {/if}

  </div>
</main>
