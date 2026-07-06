<script lang="ts">
	import { Label } from '$lib/components/ui/label';

	interface Props {
		class?: string;
		children?: any;
		[key: string]: any;
	}

	let { class: className = '', children, ...restProps }: Props = $props();
</script>

<Label
	class="text-text-secondary text-[11px] font-semibold tracking-[0.1em] uppercase {className}"
	{...restProps}
>
	{@render children?.()}
</Label>
