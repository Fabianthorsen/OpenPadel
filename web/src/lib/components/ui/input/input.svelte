<script lang="ts" module>
	import { cn, type WithElementRef } from '$lib/utils.js';
	import type { HTMLInputAttributes, HTMLInputTypeAttribute } from 'svelte/elements';
	import { tv } from 'tailwind-variants';

	/**
	 * Input component for form data collection.
	 *
	 * @example
	 * <Input type="text" placeholder="Name" />
	 *
	 * @example
	 * <Input type="email" ariaInvalid={hasError} />
	 *
	 * @example
	 * <Input type="password" disabled />
	 *
	 * Pair with Label for accessibility:
	 * ```svelte
	 * <Label for="email">Email</Label>
	 * <Input id="email" type="email" />
	 * ```
	 */
	export const inputVariants = tv({
		base: 'dark:bg-input/30 border-input focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive dark:aria-invalid:border-destructive/50 disabled:bg-input/50 dark:disabled:bg-input/80 file:text-foreground placeholder:text-muted-foreground h-8 w-full min-w-0 rounded-lg border bg-transparent px-2.5 py-1 text-base transition-colors outline-none file:inline-flex file:h-6 file:border-0 file:bg-transparent file:text-sm file:font-medium focus-visible:ring-3 disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 aria-invalid:ring-3 md:text-sm',
	});

	type InputType = Exclude<HTMLInputTypeAttribute, 'file'>;

	/**
	 * Input component props. Supports all HTML input types: text, password, email, number, search, tel, url, date, time, file, etc.
	 *
	 * File inputs bind to `files` prop; other types bind to `value`.
	 */
	export type InputProps = WithElementRef<
		Omit<HTMLInputAttributes, 'type'> &
			(
				| { type: 'file'; files?: FileList }
				| {
						/** type: text | password | email | number | search | tel | url | date | time | etc. */
						type?: InputType;
						files?: undefined;
				  }
			)
	> & {
		/** value: two-way binding for text-based inputs. */
		value?: string | number | null;

		/** files: two-way binding for file inputs (type='file' only). */
		files?: FileList;

		/** Disabled: input cannot be interacted with, appears inactive. */
		disabled?: boolean;

		/** Invalid: true if input data is invalid; indicates error state for screen readers. */
		ariaInvalid?: boolean;
	};
</script>

<script lang="ts">
	type Props = InputProps;

	let {
		ref = $bindable(null),
		value = $bindable(),
		type,
		files = $bindable(),
		class: className,
		'data-slot': dataSlot = 'input',
		...restProps
	}: Props = $props();
</script>

{#if type === 'file'}
	<input
		bind:this={ref}
		data-slot={dataSlot}
		class={cn(inputVariants(), className)}
		type="file"
		bind:files
		bind:value
		{...restProps}
	/>
{:else}
	<input
		bind:this={ref}
		data-slot={dataSlot}
		class={cn(inputVariants(), className)}
		{type}
		bind:value
		{...restProps}
	/>
{/if}
