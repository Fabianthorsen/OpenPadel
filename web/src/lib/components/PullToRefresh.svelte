<script lang="ts">
  import type { Snippet } from 'svelte';

  let {
    onRefresh,
    disabled = false,
    children
  }: {
    onRefresh: () => Promise<void>;
    disabled?: boolean;
    children: Snippet;
  } = $props();

  const THRESHOLD = 80;
  const MAX_PULL = 120;

  let dragStartY = 0;
  let pendingDelta = 0;
  let rafId: number | null = null;
  let dragOffset = $state(0);
  let dragging = $state(false);
  let refreshing = $state(false);
  let scrollContainer: HTMLElement | null = null;

  function getScrollTop(): number {
    return scrollContainer?.scrollTop ?? 0;
  }

  function onTouchStart(e: TouchEvent) {
    if (disabled || getScrollTop() > 0) return;
    dragStartY = e.touches[0].clientY;
    dragging = true;
    dragOffset = 0;
  }

  function onTouchMove(e: TouchEvent) {
    if (!dragging || disabled) return;
    const delta = e.touches[0].clientY - dragStartY;
    if (delta <= 0) { dragOffset = 0; return; }
    if (getScrollTop() === 0 && delta > 0) e.preventDefault();
    // Batch updates to sync with display refresh rate
    pendingDelta = delta;
    if (!rafId) {
      rafId = requestAnimationFrame(() => {
        dragOffset = Math.min(MAX_PULL, pendingDelta);
        rafId = null;
      });
    }
  }

  async function onTouchEnd() {
    if (!dragging || disabled) return;
    dragging = false;
    if (rafId) { cancelAnimationFrame(rafId); rafId = null; }
    if (dragOffset >= THRESHOLD) {
      refreshing = true;
      dragOffset = 0;
      try {
        await Promise.all([onRefresh(), new Promise(r => setTimeout(r, 1000))]);
      } finally {
        refreshing = false;
      }
    } else {
      dragOffset = 0;
    }
  }

  // Non-passive touchmove to allow preventDefault
  function nonPassiveTouchMove(node: HTMLElement) {
    node.addEventListener('touchmove', onTouchMove, { passive: false });
    return { destroy() { node.removeEventListener('touchmove', onTouchMove); } };
  }

  let progress = $derived(Math.min(1, dragOffset / THRESHOLD));
  let arrowRotation = $derived(Math.min(180, progress * 180));
</script>

<div
  class="relative flex flex-1 flex-col h-full overflow-hidden"
  role="region"
  aria-label="Pull to refresh"
  ontouchstart={onTouchStart}
  ontouchend={onTouchEnd}
  use:nonPassiveTouchMove
>
  <!-- Indicator: fixed 48px, positioned above content, revealed by content translateY -->
  <div
    class="absolute inset-x-0 top-0 flex h-12 items-center justify-center text-[var(--text-disabled)]"
    style="opacity: {refreshing ? 1 : progress};"
  >
    {#if refreshing}
      <!-- Dotted spinner: 8 dots arranged in a circle, staggered opacity -->
      <div class="relative size-7" aria-label="Loading">
        {#each Array(8) as _, i}
          {@const size = 4 - i * 0.35}
          <div
            class="absolute rounded-full bg-current"
            style="
              width: {size}px;
              height: {size}px;
              top: 0;
              left: 50%;
              margin-left: -{size / 2}px;
              transform-origin: {size / 2}px 14px;
              transform: rotate({i * 45}deg);
              animation: ptr-fade 0.8s linear infinite;
              animation-delay: {(8 - i) * -0.1}s;
              opacity: {1 - i * 0.1};
            "
          ></div>
        {/each}
      </div>
    {:else}
      <!-- Pull arrow, rotates as you drag -->
      <svg
        style="transform: rotate({arrowRotation}deg); transition: {dragging ? 'none' : 'transform 0.2s ease'};"
        xmlns="http://www.w3.org/2000/svg"
        width="24"
        height="24"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <line x1="12" y1="5" x2="12" y2="19" />
        <polyline points="19 12 12 19 5 12" />
      </svg>
    {/if}
  </div>

  <!-- Content pushed down via transform (GPU-accelerated, no reflows) -->
  <div
    bind:this={scrollContainer}
    class="flex-1 overflow-y-auto will-change-transform"
    style="transform: translateY({refreshing ? 48 : dragOffset}px); transition: {dragging ? 'none' : 'transform 0.2s ease'};"
  >
    {@render children()}
  </div>
</div>
