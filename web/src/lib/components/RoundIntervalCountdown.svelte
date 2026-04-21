<script lang="ts">
  import { onMount } from 'svelte';
  import { _ } from 'svelte-i18n';

  interface Props {
    intervalMinutes: number;
    onComplete: () => void;
    onSkipEarly: () => void;
  }

  let { intervalMinutes, onComplete, onSkipEarly }: Props = $props();

  let remaining = $state(0);
  let progress = $state(100);
  let display = $state('');

  onMount(() => {
    const startTime = Date.now();
    const totalDuration = intervalMinutes * 60 * 1000;
    const endTime = startTime + totalDuration;
    remaining = totalDuration;

    const updateTimer = () => {
      const now = Date.now();
      remaining = Math.max(0, endTime - now);
      progress = (remaining / totalDuration) * 100;

      // Format display MM:SS
      const minutes = Math.floor(remaining / 60000);
      const seconds = Math.floor((remaining % 60000) / 1000);
      const paddedSeconds = String(seconds).padStart(2, '0');
      display = `${minutes}:${paddedSeconds}`;

      if (remaining <= 0) {
        onComplete();
      } else {
        requestAnimationFrame(updateTimer);
      }
    };

    const interval = setInterval(updateTimer, 100);
    updateTimer();

    return () => clearInterval(interval);
  });

  function handleSkip() {
    remaining = 0;
    onSkipEarly();
  }
</script>

<div class="flex flex-col items-center justify-center gap-6 py-8">
  <!-- Timer Display -->
  <div class="text-center space-y-3">
    <p class="text-sm text-text-secondary font-medium">
      {$_('active_countdown_starting')}
    </p>
    <div class="text-5xl font-bold font-mono tracking-tight text-primary">
      {display}
    </div>
  </div>

  <!-- Progress Bar -->
  <div class="w-full max-w-xs h-1.5 bg-surface-raised rounded-full overflow-hidden">
    <div
      class="h-full bg-primary transition-all duration-100"
      style="width: {progress}%"
    ></div>
  </div>

  <!-- Start Now Button -->
  <button
    onclick={handleSkip}
    class="px-4 py-2 rounded-lg bg-surface-raised hover:bg-border text-text-primary font-medium transition-colors text-sm"
  >
    {$_('active_countdown_start_now')}
  </button>
</div>
