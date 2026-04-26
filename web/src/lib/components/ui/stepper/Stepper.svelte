<script lang="ts">
  let {
    value = $bindable(7),
    min = 1,
    max = 20,
    step = 1,
    onchange,
  }: {
    value?: number;
    min?: number;
    max?: number;
    step?: number;
    onchange?: (v: number) => void;
  } = $props();

  function decrement() {
    if (value > min) {
      value = Math.max(min, value - step);
      onchange?.(value);
    }
  }

  function increment() {
    if (value < max) {
      value = Math.min(max, value + step);
      onchange?.(value);
    }
  }
</script>

<div class="flex items-center gap-3">
  <button
    onclick={decrement}
    disabled={value <= min}
    class="flex h-9 w-9 items-center justify-center rounded-full bg-surface-raised text-lg font-semibold transition-colors hover:bg-border disabled:opacity-40 select-none"
    aria-label="Decrease"
  >−</button>
  <span class="min-w-[2ch] text-center text-xl font-[800] tabular-nums">{value}</span>
  <button
    onclick={increment}
    disabled={value >= max}
    class="flex h-9 w-9 items-center justify-center rounded-full bg-surface-raised text-lg font-semibold transition-colors hover:bg-border disabled:opacity-40 select-none"
    aria-label="Increase"
  >+</button>
</div>
