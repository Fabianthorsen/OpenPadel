<script lang="ts" module>
	import { tv } from 'tailwind-variants';
	import type { LabelRootProps } from 'bits-ui';

	/**
	 * Label for form inputs. Always pair with an Input by matching the Label's
	 * `for` to the Input's `id` for accessibility.
	 *
	 * @example
	 * <Label for="email">Email</Label>
	 * <Input id="email" type="email" />
	 *
	 * @example
	 * <Label for="email">Email <span class="text-destructive">*</span></Label>
	 * <Input id="email" type="email" ariaInvalid={hasError} />
	 */
	export const labelVariants = tv({
		base: 'flex items-center gap-2 text-sm leading-none font-medium select-none group-data-[disabled=true]:pointer-events-none group-data-[disabled=true]:opacity-50 peer-disabled:cursor-not-allowed peer-disabled:opacity-50'
	});

	export type LabelProps = LabelRootProps;
</script>

<script lang="ts">
	import { Label as LabelPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils.js';

	let { ref = $bindable(null), class: className, ...restProps }: LabelProps = $props();
</script>

<LabelPrimitive.Root
	bind:ref
	data-slot="label"
	class={cn(labelVariants(), className)}
	{...restProps}
/>
