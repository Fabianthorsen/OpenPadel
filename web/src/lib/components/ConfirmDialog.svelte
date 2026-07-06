<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog';

	let {
		open,
		title,
		description,
		confirmLabel,
		cancelLabel = 'Cancel',
		destructive = false,
		onconfirm,
		oncancel
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

<Dialog.Root bind:open>
	<Dialog.Content class="w-full max-w-sm">
		<Dialog.Header>
			<Dialog.Title>{title}</Dialog.Title>
		</Dialog.Header>
		{#if description}
			<p class="text-text-secondary text-sm">{description}</p>
		{/if}

		<Dialog.Footer class="flex gap-2 pt-4">
			<button
				onclick={oncancel}
				class="border-border text-text-secondary hover:bg-surface-raised h-auto flex-1 rounded-2xl border px-4 py-3.5 text-sm font-semibold transition-colors"
			>
				{cancelLabel}
			</button>
			<button
				onclick={onconfirm}
				class="h-auto flex-1 rounded-2xl px-4 py-4 text-[15px] font-semibold transition-colors
          {destructive
					? 'bg-destructive text-white hover:opacity-90'
					: 'bg-primary hover:bg-primary-hover text-white'}"
			>
				{confirmLabel}
			</button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
