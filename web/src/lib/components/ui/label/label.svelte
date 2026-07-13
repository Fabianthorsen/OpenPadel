<script lang="ts">
	import { Label as LabelPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils.js';

	/**
	 * Label component for form inputs.
	 * Always pair with Input using htmlFor to connect by id for accessibility.
	 *
	 * @example
	 * <Label htmlFor="email">Email</Label>
	 * <Input id="email" type="email" />
	 *
	 * @example
	 * <Label htmlFor="email">
	 *   Email <span class="text-destructive">*</span>
	 * </Label>
	 * <Input id="email" type="email" ariaInvalid={hasError} />
	 */
	let {
		ref = $bindable(null),
		class: className,
		...restProps
	}: LabelPrimitive.RootProps = $props();
</script>

<LabelPrimitive.Root
	bind:ref
	data-slot="label"
	class={cn(
		'flex items-center gap-2 text-sm leading-none font-medium select-none group-data-[disabled=true]:pointer-events-none group-data-[disabled=true]:opacity-50 peer-disabled:cursor-not-allowed peer-disabled:opacity-50',
		className
	)}
	{...restProps}
/>
