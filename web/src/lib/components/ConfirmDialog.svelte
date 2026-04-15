<script lang="ts">
  import * as Sheet from '$lib/components/ui/sheet';

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

<Sheet.Root bind:open>
  <Sheet.Content class="w-full max-w-sm sm:max-w-md">
    <div class="space-y-5">
      <div class="space-y-1.5">
        <Sheet.Header>
          <Sheet.Title>{title}</Sheet.Title>
        </Sheet.Header>
        {#if description}
          <p class="text-sm text-text-secondary">{description}</p>
        {/if}
      </div>

      <div class="space-y-2">
        <button
          onclick={onconfirm}
          class="h-auto w-full rounded-2xl px-4 py-4 text-[15px] font-semibold transition-colors
            {destructive
              ? 'bg-destructive text-white hover:opacity-90'
              : 'bg-primary text-white hover:bg-primary-hover'}"
        >
          {confirmLabel}
        </button>
        <button
          onclick={oncancel}
          class="h-auto w-full rounded-2xl border border-border px-4 py-3.5 text-sm font-semibold text-text-secondary transition-colors hover:bg-surface-raised"
        >
          {cancelLabel}
        </button>
      </div>
    </div>
  </Sheet.Content>
</Sheet.Root>
