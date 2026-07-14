<script lang="ts">
	import * as Drawer from '$lib/components/ui/drawer';
	import { numpad } from '$lib/stores/numpad';
</script>

<!--
	Mobile-optimized score numpad, rendered at page level so it anchors to the
	viewport (see fix "move numpad to page level"). Driven entirely by the
	numpad store: whoever opens it wires the digit/delete/confirm callbacks.
-->
<Drawer.Root open={!!$numpad} onOpenChange={(open) => !open && $numpad?.onClose()}>
	<Drawer.Content class="mx-auto flex max-h-[80vh] w-full max-w-[480px] flex-col gap-3">
		<div class="px-6 pt-6">
			<p
				class="text-text-disabled mb-3 text-center text-[10px] font-bold tracking-widest uppercase"
			>
				Target: {$numpad?.targetPoints}
			</p>
			<p
				class="mb-6 text-center text-[64px] leading-none font-[800] tabular-nums transition-transform
        {$numpad?.shaking ? 'animate-[shake_0.4s_ease-in-out]' : ''}"
			>
				{$numpad?.value || '0'}
			</p>
		</div>
		<div class="flex-1 px-6 pb-[env(safe-area-inset-bottom)]">
			<div class="mx-auto grid max-w-sm grid-cols-3 gap-3">
				{#each ['1', '2', '3', '4', '5', '6', '7', '8', '9'] as d}
					<button
						onclick={() => $numpad?.onDigit(d)}
						class="bg-surface-raised rounded-2xl py-4 text-xl font-[800] transition-all select-none active:scale-95"
						aria-label="Enter {d}">{d}</button
					>
				{/each}
				<button
					onclick={() => $numpad?.onDelete()}
					class="bg-surface-raised rounded-2xl py-4 text-xl font-[800] transition-all select-none active:scale-95"
					aria-label="Delete">⌫</button
				>
				<button
					onclick={() => $numpad?.onDigit('0')}
					class="bg-surface-raised rounded-2xl py-4 text-xl font-[800] transition-all select-none active:scale-95"
					aria-label="Enter 0">0</button
				>
				<button
					onclick={() => $numpad?.onConfirm()}
					class="bg-primary text-primary-foreground rounded-2xl py-4 text-xl font-[800] transition-all select-none active:scale-95"
					aria-label="Confirm">✓</button
				>
			</div>
		</div>
	</Drawer.Content>
</Drawer.Root>
