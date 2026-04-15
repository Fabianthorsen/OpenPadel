<script lang="ts">
	import { ToggleGroup as ToggleGroupPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils.js';

	let {
		value = $bindable(),
		class: className = '',
		children,
		...restProps
	}: {
		value?: string | $bindable();
		class?: string;
		children?: any;
	} = $props();
</script>

<ToggleGroupPrimitive.Root
	type="single"
	bind:value
	class={cn('flex gap-2', className)}
	{...restProps}
>
	{@render children?.()}
</ToggleGroupPrimitive.Root>
