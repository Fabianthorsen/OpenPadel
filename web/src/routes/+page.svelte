<script lang="ts">
  import { goto } from '$app/navigation';
  import { api } from '$lib/api/client';
  import { Button } from '$lib/components/ui/button';
  import { Input } from '$lib/components/ui/input';
  import { setLocale, locale } from '$lib/i18n';
  import { _ } from 'svelte-i18n';

  let step = $state<'home' | 'setup'>('home');
  let courts = $state(2);
  let points = $state(24);
  let name = $state('');
  let creating = $state(false);
  let error = $state('');
  let shaking = $state(false);
  let joinCode = $state('');

  function shake() {
    shaking = false;
    requestAnimationFrame(() => { shaking = true; });
    setTimeout(() => { shaking = false; }, 400);
  }

  async function create() {
    if (!name.trim()) { error = $_('create_error_name'); shake(); return; }
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
        <p class="text-[var(--text-secondary)]">{$_('home_tagline')}</p>
      </div>

      <!-- Actions -->
      <div class="space-y-4">
        <Button
          onclick={() => (step = 'setup')}
          class="h-auto w-full rounded-2xl bg-[var(--primary)] px-4 py-4 text-[15px] font-semibold text-white hover:bg-[var(--primary-hover)]"
        >
          {$_('home_create_tournament')}
        </Button>

        <div class="flex items-center gap-3">
          <div class="h-px flex-1 bg-[var(--border)]"></div>
          <span class="text-xs text-[var(--text-disabled)]">{$_('home_join_code_divider')}</span>
          <div class="h-px flex-1 bg-[var(--border)]"></div>
        </div>

        <form onsubmit={(e) => { e.preventDefault(); joinByCode(); }} class="flex gap-2">
          <Input
            bind:value={joinCode}
            oninput={(e: Event) => { joinCode = (e.currentTarget as HTMLInputElement).value.toUpperCase(); }}
            placeholder={$_('home_join_code_placeholder')}
            maxlength={4}
            autocomplete="off"
            autocorrect="off"
            autocapitalize="characters"
            spellcheck={false}
            class="min-w-0 flex-1 rounded-2xl border-0 bg-[var(--surface-raised)] px-4 py-3.5 text-sm"
          />
          <Button
            type="submit"
            disabled={!joinCode.trim()}
            variant="secondary"
            class="h-auto rounded-2xl bg-[var(--surface-raised)] px-5 text-sm font-semibold text-[var(--text-primary)] hover:bg-[var(--border)]"
          >
            {$_('home_join_button')}
          </Button>
        </form>
      </div>
    </div>

    <!-- Language toggle -->
    <div class="flex justify-center gap-3 pt-8">
      {#each [['en', 'EN'], ['no', 'NO']] as [lang, label]}
        <button
          onclick={() => setLocale(lang)}
          class="text-xs font-semibold transition-colors {$locale === lang ? 'text-[var(--primary)]' : 'text-[var(--text-disabled)] hover:text-[var(--text-secondary)]'}"
        >
          {label}
        </button>
      {/each}
    </div>
  </div>
  </main>

{:else}
  <main class="flex min-h-svh flex-col items-center px-6 py-6">
  <div class="w-full max-w-sm">
    <!-- Nav -->
    <nav class="flex items-center justify-between">
      <Button
        onclick={() => (step = 'home')}
        variant="ghost"
        class="flex h-8 w-8 items-center justify-center rounded-full p-0 text-lg text-[var(--text-secondary)]"
      >
        ×
      </Button>
      <span class="text-sm font-semibold text-[var(--primary)]">NotTennis</span>
      <div class="w-8"></div>
    </nav>

    <!-- Header -->
    <div class="mt-8 space-y-2">
      <h1 class="text-[34px] font-[800]">{$_('create_title_line1')}<br />{$_('create_title_line2')}</h1>
      <p class="text-[var(--text-secondary)]">{$_('create_subtitle')}</p>
    </div>

    <!-- Form -->
    <div class="mt-8 space-y-7">

      <!-- Game mode -->
      <div class="space-y-2.5">
        <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('create_game_mode_label')}</p>
        <div class="flex flex-wrap gap-2">
          <span class="inline-flex items-center gap-1.5 rounded-full bg-[var(--primary)] px-4 py-2 text-sm font-semibold text-white">
            Americano
          </span>
          <span class="inline-flex items-center rounded-full bg-[var(--surface-raised)] px-4 py-2 text-sm text-[var(--text-disabled)]">
            {$_('create_mexicano_soon')}
          </span>
        </div>
        <p class="text-xs text-[var(--text-secondary)]">{$_('create_game_mode_hint')}</p>
      </div>

      <!-- Courts -->
      <div class="space-y-2.5">
        <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('create_courts_label')}</p>
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
        <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('create_points_label')}</p>
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
          {points === 16 ? $_('create_points_quick') : points === 24 ? $_('create_points_standard') : $_('create_points_long')}
        </p>
      </div>

      <!-- Your name -->
      <div class="space-y-2.5">
        <p class="text-[11px] font-semibold uppercase tracking-[0.1em] text-[var(--text-secondary)]">{$_('create_organiser_label')}</p>
        <Input
          bind:value={name}
          oninput={() => (error = '')}
          placeholder={$_('create_organiser_placeholder')}
          maxlength={32}
          aria-invalid={!!error}
          class="rounded-2xl border-0 bg-[var(--surface-raised)] px-4 py-3.5 text-sm {shaking ? 'shake' : ''}"
        />
      </div>

      <!-- Info note -->
      <div class="flex gap-3 rounded-2xl bg-[var(--surface-raised)] px-4 py-3.5">
        <span class="mt-px shrink-0 text-[var(--text-secondary)]">ℹ</span>
        <p class="text-sm text-[var(--text-secondary)]">{$_('create_info_note')}</p>
      </div>

      <Button
        onclick={create}
        disabled={creating}
        class="h-auto w-full rounded-2xl bg-[var(--primary)] px-4 py-4 text-[15px] font-semibold text-white hover:bg-[var(--primary-hover)]"
      >
        {creating ? $_('create_button_loading') : $_('create_button')}
      </Button>
    </div>
  </div>
  </main>
{/if}
