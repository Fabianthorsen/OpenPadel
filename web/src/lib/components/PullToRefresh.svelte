<script lang="ts">
  import type { Snippet } from 'svelte';

  let {
    onRefresh,
    children
  }: {
    onRefresh: () => Promise<void>;
    children: Snippet;
  } = $props();

  const THRESHOLD = 80;
  const MAX_PULL = 120;

  let dragStartY = 0;
  let dragOffset = $state(0);
  let dragging = $state(false);
  let refreshing = $state(false);

  function getScrollTop(): number {
    return window.scrollY || document.documentElement.scrollTop || 0;
  }

  function onTouchStart(e: TouchEvent) {
    if (getScrollTop() > 0) return;
    dragStartY = e.touches[0].clientY;
    dragging = true;
    dragOffset = 0;
  }

  function onTouchMove(e: TouchEvent) {
    if (!dragging) return;
    const delta = e.touches[0].clientY - dragStartY;
    if (delta <= 0) { dragOffset = 0; return; }
    if (getScrollTop() === 0 && delta > 0) e.preventDefault();
    dragOffset = Math.min(MAX_PULL, delta);
  }

  async function onTouchEnd() {
    if (!dragging) return;
    dragging = false;
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

  // Non-passive touchmove to allow preventDefault (same pattern as Lobby.svelte)
  function nonPassiveTouchMove(node: HTMLElement) {
    node.addEventListener('touchmove', onTouchMove, { passive: false });
    return { destroy() { node.removeEventListener('touchmove', onTouchMove); } };
  }

  let progress = $derived(Math.min(1, dragOffset / THRESHOLD));
  let arrowRotation = $derived(Math.min(180, progress * 180));
</script>

<div
  class="relative min-h-full"
  role="region"
  aria-label="Pull to refresh"
  ontouchstart={onTouchStart}
  ontouchend={onTouchEnd}
  use:nonPassiveTouchMove
>
  <!-- Indicator -->
  <div
    class="flex items-center justify-center overflow-hidden text-[var(--text-disabled)] transition-[height,opacity] duration-200 ease-out"
    style="height: {dragOffset > 0 || refreshing ? (refreshing ? 48 : dragOffset) : 0}px; opacity: {refreshing ? 1 : progress};"
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
        style="transform: rotate({arrowRotation}deg); transition: transform 0.15s ease;"
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

  <!-- Content pushed down while pulling -->
  <div
    class="will-change-transform"
    style="transform: translateY({dragOffset}px); transition: {dragging ? 'none' : 'transform 0.2s ease'};"
  >
    {@render children()}
  </div>
</div>
