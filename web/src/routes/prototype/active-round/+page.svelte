<!--
  PROTOTYPE — wipe me. Active-round view exploration for wayfinder #170.
  Three radically-different variants on this throwaway route, switchable via ?variant= and the
  floating bar (dev only). Score-ENTRY interaction is a separate ticket (#171) — variants show the
  score affordance, not the entry UI. Run: `make dev/web`, open /prototype/active-round.
-->
<script lang="ts">
	import { page } from '$app/state';
	import VariantA from './VariantA.svelte';
	import VariantB from './VariantB.svelte';
	import VariantC from './VariantC.svelte';
	import PrototypeSwitcher from './PrototypeSwitcher.svelte';
	import { session, round } from './mock';

	const variant = $derived(page.url.searchParams.get('variant') ?? 'A');
	const isAdmin = $derived(page.url.searchParams.get('admin') !== '0');
	const variants = [
		{ key: 'A', name: 'Glanceable court list' },
		{ key: 'B', name: 'Focused one-court' },
		{ key: 'C', name: 'Dense scoreboard' }
	];
</script>

<svelte:head><title>PROTOTYPE — Active round #170</title></svelte:head>

<div class="bg-background min-h-svh">
	{#if variant === 'B'}
		<VariantB {session} {round} {isAdmin} />
	{:else if variant === 'C'}
		<VariantC {session} {round} {isAdmin} />
	{:else}
		<VariantA {session} {round} {isAdmin} />
	{/if}
	<PrototypeSwitcher {variants} current={variant} admin={isAdmin} />
</div>
