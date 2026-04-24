<script lang="ts">
  import { goto } from '$app/navigation';
  import { api, ApiError } from '$lib/api/client';
  import { auth } from '$lib/auth.svelte';
  import { _ } from 'svelte-i18n';
  import { translateApiError } from '$lib/i18n/errors';
  import { PillToggleGroup, PillToggleItem } from '$lib/components/ui/pill-toggle-group';
  import * as Drawer from '$lib/components/ui/drawer';

  let { open = $bindable(false) }: { open?: boolean } = $props();

  let gameMode = $state<'americano' | 'mexicano'>('americano');
  let creating = $state(false);
  let error = $state('');

  async function create() {
    creating = true;
    error = '';
    try {
      const defaults = gameMode === 'mexicano'
        ? { courts: 2, points: 24, rounds_total: 7 }
        : { courts: 2, points: 24 };
      const session = await api.sessions.create({
        game_mode: gameMode,
        name: '',
        ...defaults,
      });
      const adminToken = session.admin_token!;
      localStorage.setItem(`admin_token_${session.id}`, adminToken);
      const player = await api.players.join(session.id, auth.user!.display_name, auth.token ?? undefined, adminToken);
      localStorage.setItem(`player_id_${session.id}`, player.id);
      localStorage.setItem('last_session_id', session.id);
      open = false;
      goto(`/s/${session.id}?token=${adminToken}`);
    } catch (e) {
      error = e instanceof ApiError ? translateApiError(e.message) : translateApiError('server_error');
      creating = false;
    }
  }
</script>

<Drawer.Root bind:open>
  <Drawer.Content class="flex flex-col overflow-hidden sm:data-[vaul-drawer-direction=bottom]:left-1/2 sm:data-[vaul-drawer-direction=bottom]:-translate-x-1/2 sm:data-[vaul-drawer-direction=bottom]:w-[480px] sm:data-[vaul-drawer-direction=bottom]:max-w-[480px] sm:data-[vaul-drawer-direction=bottom]:bottom-6">
    <Drawer.Header>
      <div class="flex items-center justify-between w-full">
        <h2 class="text-lg font-[800]">{$_('create_title_line1')} {$_('create_title_line2')}</h2>
        <Drawer.Close class="hidden md:flex h-8 w-8 items-center justify-center rounded-full bg-surface-raised text-text-secondary hover:bg-border transition-colors text-xl leading-none">×</Drawer.Close>
      </div>
    </Drawer.Header>

    <div class="flex-1 overflow-y-auto px-6 pb-8 space-y-6">

      <!-- Game mode -->
      <div class="space-y-3">
        <PillToggleGroup bind:value={gameMode}>
          <PillToggleItem value="americano">Americano</PillToggleItem>
          <PillToggleItem value="mexicano">Mexicano</PillToggleItem>
        </PillToggleGroup>
        <p class="text-sm text-text-secondary">
          {gameMode === 'mexicano' ? $_('create_mexicano_hint') : $_('create_americano_hint')}
        </p>
      </div>

      {#if error}
        <p class="text-sm text-destructive">{error}</p>
      {/if}

      <button
        onclick={create}
        disabled={creating}
        class="w-full rounded-2xl bg-primary px-4 py-4 text-[15px] font-semibold text-white disabled:opacity-60 hover:bg-primary-hover transition-colors"
      >
        {creating ? $_('create_button_loading') : $_('create_button')}
      </button>

    </div>
  </Drawer.Content>
</Drawer.Root>
