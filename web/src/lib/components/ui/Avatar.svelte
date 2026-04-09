<script lang="ts">
	import * as icons from 'lucide-svelte';

	type Size = 'sm' | 'md' | 'lg' | 'xl';

	let {
		icon = '',
		color = '',
		name = '',
		size = 'md' as Size
	}: {
		icon?: string;
		color?: string;
		name?: string;
		size?: 'sm' | 'md' | 'lg' | 'xl';
	} = $props();

	const colorMap: Record<string, string> = {
		forest: 'bg-[#2d5a1a] text-white',
		sky: 'bg-[#0ea5e9] text-white',
		orange: 'bg-[#f97316] text-white',
		coral: 'bg-[#f43f5e] text-white',
		purple: 'bg-[#8b5cf6] text-white',
		teal: 'bg-[#14b8a6] text-white',
		gold: 'bg-[#eab308] text-white',
		slate: 'bg-[#64748b] text-white',
		rose: 'bg-[#ec4899] text-white',
		charcoal: 'bg-[#374151] text-white',
		white: 'bg-white/20 text-white'
	};

	const sizeMap: Record<Size, { circle: string; icon: string; text: string }> = {
		sm: { circle: 'size-7', icon: 'size-3.5', text: 'text-[10px]' },
		md: { circle: 'size-9', icon: 'size-4', text: 'text-xs' },
		lg: { circle: 'size-14', icon: 'size-7', text: 'text-lg' },
		xl: { circle: 'size-20', icon: 'size-10', text: 'text-2xl' }
	};

	const colorClass = $derived(colorMap[color] ?? 'bg-[#2d5a1a] text-white');
	const s = $derived(sizeMap[size]);

	const IconComponent = $derived(icon ? (icons as Record<string, any>)[icon] : null);

	const initials = $derived(
		name
			.split(' ')
			.map((w) => w[0])
			.join('')
			.toUpperCase()
			.slice(0, 2)
	);
</script>

<div class="shrink-0 rounded-full {s.circle} {colorClass} flex items-center justify-center font-semibold">
	{#if IconComponent}
		<IconComponent class={s.icon} strokeWidth={2} />
	{:else}
		<span class={s.text}>{initials || '?'}</span>
	{/if}
</div>
