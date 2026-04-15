<script lang="ts">
	import { Label } from '$lib/components/ui/label';
	import type { HTMLAttributes } from 'svelte/elements';

	interface Props extends HTMLAttributes<HTMLLabelElement> {
		children?: any;
	}

	let { class: className = '', children, ...restProps }: Props = $props();
</script>

<Label
	class="text-[11px] font-semibold uppercase tracking-[0.1em] text-text-secondary {className}"
	{...restProps}
>
	{@render children?.()}
</Label>
