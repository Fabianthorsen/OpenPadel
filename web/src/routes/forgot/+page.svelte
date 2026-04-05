<script lang="ts">
  import { api } from '$lib/api/client';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import { _ } from 'svelte-i18n';

  let email = $state('');
  let loading = $state(false);
  let sent = $state(false);

  async function submit() {
    loading = true;
    try {
      await api.auth.forgotPassword(email.trim());
    } catch {
      // Always show success — never reveal whether the email exists
    } finally {
      loading = false;
      sent = true;
    }
  }
</script>

<main class="flex min-h-svh flex-col items-center px-6 py-12">
  <div class="flex w-full max-w-sm flex-1 flex-col justify-center space-y-8">

    <div class="space-y-1">
      <h1 class="text-[28px] font-[800] text-[var(--primary)]">NotTennis</h1>
      <p class="text-[var(--text-secondary)]">{$_('forgot_subtitle')}</p>
    </div>

    {#if sent}
      <div class="rounded-2xl bg-[var(--surface-raised)] px-5 py-5 space-y-1">
        <p class="font-semibold">{$_('forgot_sent_title')}</p>
        <p class="text-sm text-[var(--text-secondary)]">{$_('forgot_sent_desc')}</p>
      </div>
    {:else}
      <form onsubmit={(e) => { e.preventDefault(); submit(); }} class="space-y-4">
        <div class="space-y-2">
          <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('auth_email_label')}</p>
          <Input
            bind:value={email}
            type="email"
            placeholder={$_('auth_email_placeholder')}
            autocomplete="email"
            class="rounded-2xl border-0 bg-[var(--surface-raised)] px-4 py-3.5 text-sm"
          />
        </div>

        <Button
          type="submit"
          disabled={loading || !email.trim()}
          class="h-auto w-full rounded-2xl bg-[var(--primary)] px-4 py-4 text-[15px] font-semibold text-white hover:bg-[var(--primary-hover)]"
        >
          {loading ? $_('forgot_button_loading') : $_('forgot_button')}
        </Button>
      </form>
    {/if}

    <div class="flex justify-center">
      <a href="/auth" class="text-xs text-[var(--text-disabled)] hover:text-[var(--text-secondary)]">
        ← {$_('auth_back_home')}
      </a>
    </div>

  </div>
</main>
