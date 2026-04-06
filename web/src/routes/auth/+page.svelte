<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/state';
  import { auth } from '$lib/auth.svelte';
  import { fly } from 'svelte/transition';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import { _ } from 'svelte-i18n';

  const redirect = page.url.searchParams.get('redirect') ?? '';

  $effect(() => {
    if (auth.ready && auth.user) goto(redirect || '/profile');
  });

  const resetSuccess = page.url.searchParams.get('reset') === '1';

  let mode = $state<'login' | 'register'>(page.url.searchParams.get('register') === '1' ? 'register' : 'login');
  let banner = $state<{ text: string; kind: 'success' | 'error' } | null>(
    resetSuccess ? { text: '', kind: 'success' } : null
  );
  let email = $state('');
  let password = $state('');
  let firstName = $state('');
  let lastName = $state('');
  let loading = $state(false);

  // Set success banner text once translations load
  $effect(() => {
    if (resetSuccess && banner?.kind === 'success') {
      banner = { text: $_('reset_success_banner'), kind: 'success' };
      setTimeout(() => { banner = null; }, 5000);
    }
  });

  function showError(msg: string) {
    banner = { text: msg, kind: 'error' };
  }

  async function submit() {
    banner = null;
    loading = true;
    try {
      if (mode === 'login') {
        await auth.login(email, password);
      } else {
        await auth.register(email, `${firstName.trim()} ${lastName.trim()}`.trim(), password);
      }
      goto(redirect || '/');
    } catch (e) {
      showError(e instanceof Error ? e.message : 'Something went wrong');
    } finally {
      loading = false;
    }
  }
</script>

{#if banner}
  <div
    transition:fly={{ y: -48, duration: 400 }}
    class="fixed inset-x-0 top-0 z-50 flex items-center justify-center px-4 py-3 text-sm font-semibold text-white
      {banner.kind === 'error' ? 'bg-[var(--destructive)]' : 'bg-[var(--primary)]'}"
  >
    {banner.text}
  </div>
{/if}

<main class="flex min-h-svh flex-col items-center px-6 py-12" class:pt-16={!!banner}>
  <div class="flex w-full max-w-sm flex-1 flex-col justify-center space-y-8">

    <!-- Brand -->
    <div class="space-y-1">
      <h1 class="text-[28px] font-[800] text-[var(--primary)]">OpenPadel</h1>
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

      <Button
        type="submit"
        disabled={loading}
        class="h-auto w-full rounded-2xl bg-[var(--primary)] px-4 py-4 text-[15px] font-semibold text-white hover:bg-[var(--primary-hover)]"
      >
        {loading ? '…' : mode === 'login' ? $_('auth_login_button') : $_('auth_register_button')}
      </Button>

      {#if mode === 'login'}
        <div class="flex justify-center">
          <a href="/forgot" class="text-xs text-[var(--text-disabled)] hover:text-[var(--text-secondary)]">
            {$_('auth_forgot_password')}
          </a>
        </div>
      {/if}
    </form>

    <!-- Back -->
    <div class="flex justify-center">
      <a href="/" class="text-xs text-[var(--text-disabled)] hover:text-[var(--text-secondary)]">
        ← {$_('auth_back_home')}
      </a>
    </div>

  </div>
</main>
