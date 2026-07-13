<script lang="ts" module>
	import type { DialogTriggerProps } from 'bits-ui';

	/**
	 * Drawer trigger button. Opens the drawer when clicked.
	 * Renders as a standard button; typically paired with Button component for styling.
	 *
	 * @example
	 * <DrawerTrigger>Open Drawer</DrawerTrigger>
	 *
	 * @example
	 * <DrawerTrigger disabled>Disabled</DrawerTrigger>
	 */
	export interface DrawerTriggerProps extends DialogTriggerProps {
		/** If true, trigger button is disabled. */
		disabled?: boolean;
	}
</script>

<script lang="ts">
	import { Dialog as DialogPrimitive } from 'bits-ui';

	let {
		ref = $bindable(null),
		disabled,
		class: className,
		children,
		...restProps
	}: DrawerTriggerProps = $props();
</script>

<DialogPrimitive.Trigger
	bind:ref
	data-slot="drawer-trigger"
	class={className}
	{disabled}
	{...restProps}
>
	{@render children?.()}
</DialogPrimitive.Trigger>
