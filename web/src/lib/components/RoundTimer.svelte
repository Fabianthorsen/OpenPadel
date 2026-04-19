<script lang="ts">
  import { onMount } from 'svelte';

  let {
    roundStartedAt,
    roundDurationSeconds,
    bufferSeconds,
  }: {
    roundStartedAt: string;
    roundDurationSeconds: number;
    bufferSeconds: number;
  } = $props();

  let now = $state(Date.now());

  const remaining = $derived((): number => {
    const roundStart = new Date(roundStartedAt).getTime();
    const endTime = roundStart + roundDurationSeconds * 1000;
    return Math.max(0, endTime - now);
  });

  const minutes = $derived(Math.floor(remaining() / 60000));
  const seconds = $derived(Math.floor((remaining() % 60000) / 1000));
  const displaySeconds = $derived(String(seconds).padStart(2, '0'));

  const colorClass = $derived((): string => {
    const r = remaining();
    if (r > 60000) return 'bg-emerald-500';
    if (r > 30000) return 'bg-amber-500';
    if (r > 0) return 'bg-red-500 animate-pulse';
    return 'bg-red-600';
  });

  const isBuzzer = $derived(remaining() <= 0);

  onMount(() => {
    const interval = setInterval(() => {
      now = Date.now();
    }, 1000);

    // Attempt vibration on buzzer
    if (isBuzzer) {
      navigator.vibrate?.([200, 100, 200]);
    }

    return () => {
      clearInterval(interval);
    };
  });
</script>

<div
  data-testid="timer-container"
  class="flex flex-col items-center justify-center rounded-lg {colorClass()} p-6 transition-colors duration-300"
>
  <div data-testid="timer-display" class="text-5xl font-bold text-white font-mono">
    {minutes}:{displaySeconds}
  </div>

  {#if isBuzzer}
    <div data-testid="buzzer-text" class="mt-4 text-sm font-semibold text-white text-center">
      FINISH CURRENT POINT
    </div>
  {/if}
</div>
