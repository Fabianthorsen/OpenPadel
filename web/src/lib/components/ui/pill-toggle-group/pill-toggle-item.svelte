<script lang="ts">
	import { ToggleGroup as ToggleGroupPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils.js';

	let {
		value,
		class: className = '',
		children,
		...restProps
	}: {
		value: string;
		class?: string;
		children?: any;
	} = $props();
</script>

<ToggleGroupPrimitive.Item
	{value}
	class={cn(
		'flex-1 min-w-0 rounded-full py-2.5 text-sm font-semibold transition-colors',
		'bg-surface text-text-primary border border-border',
		'data-[state=on]:bg-primary data-[state=on]:text-white',
		className
	)}
	{...restProps}
>
	{@render children?.()}
</ToggleGroupPrimitive.Item>
