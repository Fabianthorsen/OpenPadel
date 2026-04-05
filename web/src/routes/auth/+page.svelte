<script lang="ts">
  import { goto } from '$app/navigation';
  import { auth } from '$lib/auth.svelte';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import { _ } from 'svelte-i18n';

  let mode = $state<'login' | 'register'>('login');
  let email = $state('');
  let password = $state('');
  let firstName = $state('');
  let lastName = $state('');
  let error = $state('');
  let loading = $state(false);

  async function submit() {
    error = '';
    loading = true;
    try {
      if (mode === 'login') {
        await auth.login(email, password);
      } else {
        await auth.register(email, `${firstName.trim()} ${lastName.trim()}`.trim(), password);
      }
      goto('/');
    } catch (e) {
      error = e instanceof Error ? e.message : 'Something went wrong';
    } finally {
      loading = false;
    }
  }
</script>

<main class="flex min-h-svh flex-col items-center px-6 py-12">
  <div class="flex w-full max-w-sm flex-1 flex-col justify-center space-y-8">

    <!-- Brand -->
    <div class="space-y-1">
      <h1 class="text-[28px] font-[800] text-[var(--primary)]">NotTennis</h1>
      <p class="text-[var(--text-secondary)]">{mode === 'login' ? $_('auth_login_subtitle') : $_('auth_register_subtitle')}</p>
    </div>

    <form onsubmit={(e) => { e.preventDefault(); submit(); }} class="space-y-4">

      {#if mode === 'register'}
        <div class="flex gap-3">
          <div class="flex-1 space-y-2">
            <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('auth_first_name_label')}</p>
            <Input
              bind:value={firstName}
              placeholder={$_('auth_first_name_placeholder')}
              maxlength={32}
              autocomplete="given-name"
              class="rounded-2xl border-0 bg-[var(--surface-raised)] px-4 py-3.5 text-sm"
            />
          </div>
          <div class="flex-1 space-y-2">
            <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('auth_last_name_label')}</p>
            <Input
              bind:value={lastName}
              placeholder={$_('auth_last_name_placeholder')}
              maxlength={32}
              autocomplete="family-name"
              class="rounded-2xl border-0 bg-[var(--surface-raised)] px-4 py-3.5 text-sm"
            />
          </div>
        </div>
      {/if}

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

      <div class="space-y-2">
        <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('auth_password_label')}</p>
        <Input
          bind:value={password}
          type="password"
          placeholder={$_('auth_password_placeholder')}
          autocomplete={mode === 'login' ? 'current-password' : 'new-password'}
          class="rounded-2xl border-0 bg-[var(--surface-raised)] px-4 py-3.5 text-sm"
        />
      </div>

      {#if error}
        <p class="text-sm text-[var(--destructive)]">{error}</p>
      {/if}

      <Button
        type="submit"
        disabled={loading}
        class="h-auto w-full rounded-2xl bg-[var(--primary)] px-4 py-4 text-[15px] font-semibold text-white hover:bg-[var(--primary-hover)]"
      >
        {loading ? '…' : mode === 'login' ? $_('auth_login_button') : $_('auth_register_button')}
      </Button>
    </form>

    <!-- Toggle mode -->
    <p class="text-center text-sm text-[var(--text-secondary)]">
      {mode === 'login' ? $_('auth_no_account') : $_('auth_have_account')}
      <button
        onclick={() => { mode = mode === 'login' ? 'register' : 'login'; error = ''; }}
        class="font-semibold text-[var(--primary)] hover:text-[var(--primary-hover)]"
      >
        {mode === 'login' ? $_('auth_switch_register') : $_('auth_switch_login')}
      </button>
    </p>

    <!-- Back to home -->
    <div class="flex justify-center">
      <a href="/" class="text-xs text-[var(--text-disabled)] hover:text-[var(--text-secondary)]">
        ← {$_('auth_back_home')}
      </a>
    </div>

  </div>
</main>
