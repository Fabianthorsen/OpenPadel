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
  // Rubber-band resistance factor — higher = more resistance
  const RESISTANCE = 2.5;

  let dragStartY = 0;
  let pendingDelta = 0;
  let rafId: number | null = null;
  let dragOffset = $state(0);
  let dragging = $state(false);
  let refreshing = $state(false);
  let triggered = $state(false); // crossed THRESHOLD during this drag
  let scrollContainer: HTMLElement | null = null;
  let container: HTMLElement | null = null;

  function getScrollTop(): number {
    return scrollContainer?.scrollTop ?? 0;
  }

  // Rubber-band formula: feels immediate at first, then increasingly resistant
  function rubberBand(delta: number): number {
    return (delta * THRESHOLD) / (delta + THRESHOLD * RESISTANCE);
  }

  function onTouchStart(e: TouchEvent) {
    if (disabled) return;
    let el: HTMLElement | null = e.target as HTMLElement;
    while (el && el !== container) {
      if (el.scrollTop > 0) return;
      el = el.parentElement;
    }
    if (getScrollTop() > 0) return;
    dragStartY = e.touches[0].clientY;
    dragging = true;
    triggered = false;
    dragOffset = 0;
  }

  function onTouchMove(e: TouchEvent) {
    if (!dragging || disabled) return;
    const delta = e.touches[0].clientY - dragStartY;
    if (delta <= 0) { dragOffset = 0; triggered = false; return; }
    if (getScrollTop() === 0 && delta > 0) e.preventDefault();
    pendingDelta = delta;
    if (!rafId) {
      rafId = requestAnimationFrame(() => {
        dragOffset = rubberBand(pendingDelta);
        triggered = pendingDelta >= THRESHOLD;
        rafId = null;
      });
    }
  }

  async function onTouchEnd() {
    if (!dragging || disabled) return;
    dragging = false;
    if (rafId) { cancelAnimationFrame(rafId); rafId = null; }
    if (triggered) {
      refreshing = true;
      dragOffset = 0;
      triggered = false;
      try {
        await Promise.all([onRefresh(), new Promise(r => setTimeout(r, 1000))]);
      } finally {
        refreshing = false;
      }
    } else {
      dragOffset = 0;
      triggered = false;
    }
  }

  // Non-passive touchmove to allow preventDefault
  function nonPassiveTouchMove(node: HTMLElement) {
    node.addEventListener('touchmove', onTouchMove, { passive: false });
    return { destroy() { node.removeEventListener('touchmove', onTouchMove); } };
  }

  let progress = $derived(Math.min(1, dragOffset / (THRESHOLD / RESISTANCE)));
  let arrowRotation = $derived(Math.min(180, progress * 180));
  // Spring easing for the snap-back — slight overshoot feels physical
  const springEase = 'cubic-bezier(0.34, 1.4, 0.64, 1)';
</script>

<div
  bind:this={container}
  class="relative flex flex-1 flex-col h-screen-safe overflow-clip"
  role="region"
  aria-label="Pull to refresh"
  ontouchstart={onTouchStart}
  ontouchend={onTouchEnd}
  use:nonPassiveTouchMove
>
  <!-- Indicator: positioned above content, revealed by content translateY. Padded below notch/island. -->
  <div
    class="absolute inset-x-0 top-0 flex items-end justify-center pb-2"
    style="
      height: calc(4rem + env(safe-area-inset-top, 0px));
      padding-top: env(safe-area-inset-top, 0px);
      opacity: {refreshing ? 1 : progress};
      color: {triggered ? 'var(--primary)' : 'var(--text-disabled)'};
      transition: color 0.15s ease;
    "
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
      <!-- Pull arrow: rotates and scales as you drag -->
      <svg
        style="
          transform: rotate({arrowRotation}deg) scale({0.6 + progress * 0.4});
          transition: {dragging ? 'none' : `transform 0.15s ${springEase}`};
        "
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
    class="flex-1 will-change-transform min-h-0"
    style="transform: translateY({refreshing ? 64 : dragOffset}px); transition: {dragging ? 'none' : `transform 0.45s ${springEase}`};"
  >
    <div
      bind:this={scrollContainer}
      class="h-full overflow-y-auto"
    >
      {@render children()}
    </div>
  </div>
</div>
