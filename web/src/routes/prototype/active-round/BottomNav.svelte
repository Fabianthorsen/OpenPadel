<!-- PROTOTYPE — wipe me. Shared (faux) session bottom nav, for density realism only. -->
<script lang="ts">
	import { Activity, Trophy, Users } from '@lucide/svelte';
	let { active = 'scoring' }: { active?: string } = $props();
	const items = [
		{ key: 'scoring', label: 'Round', icon: Activity },
		{ key: 'standings', label: 'Standings', icon: Trophy },
		{ key: 'players', label: 'Players', icon: Users }
	];
</script>

<div
	class="border-border bg-background/95 fixed inset-x-0 bottom-0 z-40 mx-auto flex max-w-[480px] border-t backdrop-blur"
	style="padding-bottom: max(0.5rem, env(safe-area-inset-bottom));"
>
	{#each items as it (it.key)}
		{@const Icon = it.icon}
		<div
			class="flex flex-1 flex-col items-center gap-1 py-3 {active === it.key
				? 'text-primary'
				: 'text-text-secondary'}"
		>
			<Icon size={20} />
			<span class="text-[10px] font-semibold tracking-wide uppercase">{it.label}</span>
		</div>
	{/each}
</div>
