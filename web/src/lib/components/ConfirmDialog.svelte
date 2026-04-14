<script lang="ts">
  import { fly, fade } from 'svelte/transition';

  function nonPassiveBackdropTouch(node: HTMLElement) {
    node.addEventListener('touchmove', (e: TouchEvent) => e.preventDefault(), { passive: false });
    return { destroy() {} };
  }

  let {
    open,
    title,
    description,
    confirmLabel,
    cancelLabel = 'Cancel',
    destructive = false,
    onconfirm,
    oncancel,
  }: {
    open: boolean;
    title: string;
    description?: string;
    confirmLabel: string;
    cancelLabel?: string;
    destructive?: boolean;
    onconfirm: () => void;
    oncancel: () => void;
  } = $props();

</script>

{#if open}
  <!-- Backdrop -->
  <div
    transition:fade={{ duration: 150 }}
    class="fixed inset-0 z-50 flex items-end justify-center bg-black/40 backdrop-blur-sm sm:items-center touch-none"
    onclick={oncancel}
    onkeydown={(e) => e.key === 'Escape' && oncancel()}
    role="presentation"
    tabindex="-1"
    use:nonPassiveBackdropTouch
  >
    <!-- Sheet -->
    <div
      transition:fly={{ y: 400, duration: 280, opacity: 1 }}
      class="w-full max-w-sm rounded-t-3xl bg-[var(--surface)] px-6 pb-8 pt-6 space-y-5 sm:rounded-3xl sm:mx-4"
      onclick={(e) => e.stopPropagation()}
      role="presentation"
    >
      <!-- Handle -->
      <div class="mx-auto h-1 w-10 rounded-full bg-[var(--border)] sm:hidden"></div>

      <div class="space-y-1.5">
        <h2 class="text-[18px] font-[800] text-[var(--text-primary)]">{title}</h2>
        {#if description}
          <p class="text-sm text-[var(--text-secondary)]">{description}</p>
        {/if}
      </div>

      <div class="space-y-2">
        <button
          onclick={onconfirm}
          class="h-auto w-full rounded-2xl px-4 py-4 text-[15px] font-semibold transition-colors
            {destructive
              ? 'bg-[var(--destructive)] text-white hover:opacity-90'
              : 'bg-[var(--primary)] text-white hover:bg-[var(--primary-hover)]'}"
        >
          {confirmLabel}
        </button>
        <button
          onclick={oncancel}
          class="h-auto w-full rounded-2xl border border-[var(--border)] px-4 py-3.5 text-sm font-semibold text-[var(--text-secondary)] transition-colors hover:bg-[var(--surface-raised)]"
        >
          {cancelLabel}
        </button>
      </div>
    </div>
  </div>
{/if}
