<script lang="ts">
	import { ToggleGroup as ToggleGroupPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils.js';

	let {
		value = $bindable(),
		class: className = '',
		children,
		...restProps
	}: {
		value?: string;
		class?: string;
		children?: any;
		onValueChange?: (value: string) => void;
		[key: string]: any;
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
