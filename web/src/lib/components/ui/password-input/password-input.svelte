<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import { Eye, EyeOff } from '@lucide/svelte';
	import { cn } from '$lib/utils.js';

	/**
	 * Password input with a show/hide toggle. Composes the Input primitive.
	 *
	 * @example
	 * <Label for="pw">Password</Label>
	 * <PasswordInput id="pw" bind:value={password} autocomplete="current-password" />
	 */
	let {
		value = $bindable(''),
		class: className,
		...restProps
	}: {
		/** value: two-way binding. */
		value?: string;
		class?: string;
		[key: string]: unknown;
	} = $props();

	let show = $state(false);
</script>

<div class="relative" data-slot="password-input">
	<Input
		type={show ? 'text' : 'password'}
		bind:value
		class={cn('pr-10', className)}
		{...restProps}
	/>
	<button
		type="button"
		onclick={() => (show = !show)}
		aria-label={show ? 'Hide password' : 'Show password'}
		aria-pressed={show}
		class="text-text-disabled hover:text-text-secondary absolute inset-y-0 right-0 flex w-10 items-center justify-center"
	>
		{#if show}<EyeOff size={16} />{:else}<Eye size={16} />{/if}
	</button>
</div>
