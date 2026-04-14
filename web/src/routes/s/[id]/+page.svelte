<script lang="ts">
  import { page } from '$app/state';
  import { goto } from '$app/navigation';
  import { onMount, onDestroy } from 'svelte';
  import { fly } from 'svelte/transition';
  import { api, ApiError } from '$lib/api/client';
  import Lobby from '$lib/components/Lobby.svelte';
  import ActiveSession from '$lib/components/ActiveSession.svelte';
  import TennisMatch from '$lib/components/TennisMatch.svelte';
  import Leaderboard from '$lib/components/Leaderboard.svelte';
  import TennisResult from '$lib/components/TennisResult.svelte';
  import PullToRefresh from '$lib/components/PullToRefresh.svelte';
  import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
  import { toast } from 'svelte-sonner';
  import { _ } from 'svelte-i18n';
  import { translateApiError } from '$lib/i18n/errors';
  import { sessionDialog } from '$lib/stores/sessionDialog';
  import { numpad as numpadStore } from '$lib/stores/numpad';

  let session = $state<App.Session | null>(null);
  let currentRound = $state<App.Round | null>(null);
  let interval: ReturnType<typeof setInterval>;

  // Poll faster in lobby so players see each other join without waiting.
  // Active sessions use 2s so live score taps are visible to other viewers quickly.
  const POLL_LOBBY = 3_000;
  const POLL_ACTIVE = 2_000;

  const sessionId = $derived(page.params.id as string);

  function getAdminToken(): string | null {
    if (typeof localStorage === 'undefined') return null;
    const stored = localStorage.getItem(`admin_token_${sessionId}`);
    if (stored) return stored;
    const param = new URL(location.href).searchParams.get('token');
    if (param) {
      localStorage.setItem(`admin_token_${sessionId}`, param);
      window.history.replaceState({}, '', location.pathname);
    }
    return param;
  }

  const isAdmin = $derived(!!getAdminToken());

  async function load() {
    const token = getAdminToken() ?? undefined;
    try {
      session = await api.sessions.get(sessionId, token);
      if (session.status !== 'lobby' && session.game_mode !== 'tennis') {
        currentRound = await api.rounds.current(sessionId).catch(() => null);
      }
    } catch (e) {
      if (e instanceof ApiError && e.status === 404) {
        goto('/?notfound=1');
        return;
      }
      toast.error(e instanceof ApiError ? translateApiError(e.message) : translateApiError('server_error'));
    }
  }

  function scheduleNext() {
    clearInterval(interval);
    const delay = session?.status === 'lobby' ? POLL_LOBBY : POLL_ACTIVE;
    interval = setInterval(() => { load().then(scheduleNext); }, delay);
  }

  function onVisibilityChange() {
    if (document.hidden) {
      clearInterval(interval);
    } else {
      load().then(scheduleNext);
    }
  }

  onMount(() => {
    document.addEventListener('visibilitychange', onVisibilityChange);
    load().then(scheduleNext);
  });

  onDestroy(() => {
    document.removeEventListener('visibilitychange', onVisibilityChange);
    clearInterval(interval);
    numpadStore.close();
    sessionDialog.close();
  });

  async function handleDialogConfirm() {
    const callback = $sessionDialog.onConfirm;
    sessionDialog.close();
    if (callback) {
      await callback();
    }
  }

  function handleDialogCancel() {
    sessionDialog.close();
  }

  // Numpad drag-to-close state
  let numpadDragStartY = 0;
  let numpadDragOffset = $state(0);
  let numpadDragging = $state(false);
  let numpadDragLastY = 0;
  let numpadDragLastTime = 0;
  let numpadDragVelocity = 0;

  function numpadTouchStart(e: TouchEvent) {
    if (!$numpadStore) return;
    numpadDragStartY = e.touches[0].clientY;
    numpadDragLastY = numpadDragStartY;
    numpadDragLastTime = Date.now();
    numpadDragOffset = 0;
    numpadDragVelocity = 0;
    numpadDragging = true;
  }

  function numpadTouchMove(e: TouchEvent) {
    if (!numpadDragging) return;
    const now = Date.now();
    const currentY = e.touches[0].clientY;
    const delta = currentY - numpadDragStartY;

    if (delta > 0) {
      e.preventDefault();
      // Cap drag at the numpad element's own height
      const numpadElement = (e.target as HTMLElement)?.closest('[role="presentation"]');
      const maxDrag = numpadElement?.getBoundingClientRect().height ?? 300;
      numpadDragOffset = Math.min(maxDrag, delta);
      numpadDragVelocity = (currentY - numpadDragLastY) / Math.max(16, now - numpadDragLastTime);
      numpadDragLastY = currentY;
      numpadDragLastTime = now;
    }
  }

  function numpadTouchEnd() {
    if (!numpadDragging || !$numpadStore) return;
    numpadDragging = false;
    const shouldClose = numpadDragOffset > 80 || (numpadDragVelocity > 150 && numpadDragOffset > 20);
    if (shouldClose) {
      $numpadStore.onClose();
    }
    numpadDragOffset = 0;
  }

  function numpadHandleKeydown(e: KeyboardEvent) {
    if (!$numpadStore) return;
    if (e.key >= '0' && e.key <= '9') {
      e.preventDefault();
      $numpadStore.onDigit(e.key);
    } else if (e.key === 'Backspace') {
      e.preventDefault();
      $numpadStore.onDelete();
    } else if (e.key === 'Enter') {
      e.preventDefault();
      $numpadStore.onConfirm();
    }
  }

  function nonPassiveNumpadTouchMove(node: HTMLElement) {
    const handleTouchMove = (e: TouchEvent) => numpadTouchMove(e);
    node.addEventListener('touchmove', handleTouchMove, { passive: false });
    return {
      destroy() {
        node.removeEventListener('touchmove', handleTouchMove);
      }
    };
  }
</script>

<div class="flex flex-col w-screen h-screen overflow-hidden">
  {#if !session}
    <main class="flex flex-1 items-center justify-center px-4">
      <p class="text-sm text-[var(--text-secondary)]">Loading…</p>
    </main>
  {:else if session.status === 'lobby'}
    <PullToRefresh disabled={$sessionDialog.isOpen || !!$numpadStore} onRefresh={load}>
      <Lobby {session} {isAdmin} onRefresh={load} onStarted={load} />
    </PullToRefresh>
  {:else if session.status === 'active' && session.game_mode === 'tennis'}
    <PullToRefresh disabled={$sessionDialog.isOpen || !!$numpadStore} onRefresh={load}>
      <TennisMatch {session} {isAdmin} onRefresh={load} />
    </PullToRefresh>
  {:else if session.status === 'active' && currentRound}
    <PullToRefresh disabled={$sessionDialog.isOpen || !!$numpadStore} onRefresh={load}>
      <ActiveSession {session} {currentRound} {isAdmin} onRefresh={load} />
    </PullToRefresh>
  {:else if session.status === 'complete' && session.game_mode === 'tennis'}
    <TennisResult {session} />
  {:else if session.status === 'complete'}
    <Leaderboard sessionId={session.id} sessionName={session.name} complete />
  {:else}
    <main class="flex flex-1 items-center justify-center px-4">
      <p class="text-sm text-[var(--text-secondary)]">Loading…</p>
    </main>
  {/if}
</div>

<!-- Session dialogs (rendered outside PullToRefresh at document root level) -->
{#if $sessionDialog.isOpen}
  <div class="fixed inset-0 z-50">
    <ConfirmDialog
      open={$sessionDialog.isOpen}
      title={$sessionDialog.type === 'close' ? $_('close_dialog_title') : $_('cancel_dialog_title')}
      description={$sessionDialog.type === 'close' ? $_('close_dialog_desc') : $_('cancel_dialog_desc')}
      confirmLabel={$sessionDialog.type === 'close' ? $_('close_dialog_confirm') : $_('cancel_dialog_confirm')}
      cancelLabel={$sessionDialog.type === 'close' ? $_('close_dialog_cancel') : $_('cancel_dialog_cancel')}
      destructive
      onconfirm={handleDialogConfirm}
      oncancel={handleDialogCancel}
    />
  </div>
{/if}

<!-- Numpad overlay (rendered outside PullToRefresh transform context) -->
{#if $numpadStore}
  <div
    role="presentation"
    class="fixed inset-0 z-40 bg-black/40"
    onclick={$numpadStore.onClose}
    onkeydown={(e) => e.key === 'Escape' && $numpadStore.onClose()}
  ></div>
  <div
    transition:fly={{ y: 500, duration: 300, opacity: 1 }}
    role="presentation"
    class="fixed left-0 right-0 z-50 mx-auto max-w-[480px] rounded-t-3xl bg-[var(--surface)] px-5 pt-6 pb-[max(1.5rem,env(safe-area-inset-bottom))] shadow-2xl flex flex-col gap-3"
    ontouchstart={numpadTouchStart}
    ontouchend={numpadTouchEnd}
    onkeydown={numpadHandleKeydown}
    use:nonPassiveNumpadTouchMove
    style="bottom: 0; transform: translateY({numpadDragOffset}px); transition: {numpadDragging ? 'none' : 'transform 0.2s ease'};"
  >
    <p class="mb-3 text-center text-[10px] font-bold uppercase tracking-widest text-[var(--text-disabled)]">
      Target: {$numpadStore.targetPoints}
    </p>
    <p class="mb-6 text-center text-[64px] font-[800] leading-none tabular-nums transition-transform
      {$numpadStore.shaking ? 'animate-[shake_0.4s_ease-in-out]' : ''}">
      {$numpadStore.value || '0'}
    </p>
    <div class="grid grid-cols-3 gap-3">
      {#each ['1','2','3','4','5','6','7','8','9'] as d}
        <button
          onclick={() => $numpadStore.onDigit(d)}
          class="rounded-2xl bg-[var(--surface-raised)] py-4 text-xl font-[800] transition-all active:scale-95 select-none"
        >{d}</button>
      {/each}
      <button onclick={$numpadStore.onDelete} class="rounded-2xl bg-[var(--surface-raised)] py-4 text-xl font-[800] transition-all active:scale-95 select-none">⌫</button>
      <button onclick={() => $numpadStore.onDigit('0')} class="rounded-2xl bg-[var(--surface-raised)] py-4 text-xl font-[800] transition-all active:scale-95 select-none">0</button>
      <button onclick={$numpadStore.onConfirm} class="rounded-2xl bg-[var(--primary)] py-4 text-xl font-[800] text-white transition-all active:scale-95 select-none">✓</button>
    </div>
  </div>
{/if}
