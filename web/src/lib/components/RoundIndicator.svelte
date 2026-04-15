<script lang="ts">
  let { current, total }: { current: number; total: number } = $props();

  // Build the list of items to show: numbers or 'dots'
  const items = $derived((): (number | 'dots')[] => {
    if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1);
    const set = new Set<number>([1, current, total]);
    const result: (number | 'dots')[] = [];
    let prev = 0;
    for (const n of [...set].sort((a, b) => a - b)) {
      if (n - prev > 1) result.push('dots');
      result.push(n);
      prev = n;
    }
    return result;
  });
</script>

<div class="flex items-center justify-center gap-1.5">
  {#each items() as item}
    {#if item === 'dots'}
      <span class="text-xs text-text-disabled">·</span>
    {:else}
      <div
        class="flex h-7 w-7 items-center justify-center rounded-full text-xs font-semibold transition-colors
          {item === current
            ? 'bg-primary text-white'
            : item < current
              ? 'bg-primary-muted text-primary'
              : 'bg-surface-raised text-text-disabled'}"
      >
        {item}
      </div>
    {/if}
  {/each}
</div>
