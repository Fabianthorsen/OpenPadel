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
		disabled?: boolean;
	} = $props();
</script>

<ToggleGroupPrimitive.Item
	{value}
	class={cn(
		'min-w-0 flex-1 rounded-full py-2.5 text-sm font-semibold transition-colors',
		'bg-surface text-text-primary border-border border',
		'data-[state=on]:bg-primary data-[state=on]:text-primary-foreground data-[state=on]:border-primary',
		'disabled:cursor-not-allowed disabled:opacity-40',
		className
	)}
	{...restProps}
>
	{@render children?.()}
</ToggleGroupPrimitive.Item>
