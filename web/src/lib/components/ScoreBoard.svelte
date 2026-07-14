<script lang="ts">
	import { _ } from 'svelte-i18n';
	import { Card } from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import Avatar from '$lib/components/ui/Avatar.svelte';

	interface Team {
		players: Array<{ avatar_icon: string; avatar_color: string; name: string }>;
		name: string;
		score: number;
	}

	let {
		teamA,
		teamB,
		scored,
		live,
		pointsTarget,
		isAdmin,
		submitting = false,
		onAdjust,
		onScoreTap,
		onFinalize
	}: {
		teamA: Team;
		teamB: Team;
		scored: boolean;
		live: boolean;
		pointsTarget: number;
		isAdmin: boolean;
		submitting?: boolean;
		onAdjust: (team: 'a' | 'b', delta: number) => void;
		onScoreTap: (team: 'a' | 'b') => void;
		onFinalize: () => void;
	} = $props();

	const editable = $derived(isAdmin && !scored);
	const canFinalize = $derived(teamA.score + teamB.score === pointsTarget);

	function scoreColor(mine: number, other: number) {
		if (!scored) return 'text-text-primary';
		if (mine > other) return 'text-primary font-bold';
		if (mine < other) return 'text-text-disabled';
		return 'text-text-primary';
	}
</script>

<Card class="border-border bg-surface rounded-3xl p-6">
	<!-- Status chip -->
	<div class="mb-4 flex justify-center">
		{#if scored}
			<span
				class="bg-surface-raised text-text-primary inline-block rounded-full px-3 py-1 text-xs font-semibold"
			>
				{$_('court_status_final')}
			</span>
		{:else if live}
			<span
				class="bg-primary/10 text-primary inline-flex items-center gap-1 rounded-full px-3 py-1 text-xs font-semibold"
			>
				<span class="bg-primary inline-block h-2 w-2 animate-pulse rounded-full"></span>
				{$_('court_status_live')}
			</span>
		{:else}
			<span
				class="bg-surface-raised text-text-secondary inline-block rounded-full px-3 py-1 text-xs font-semibold"
			>
				{$_('court_status_upcoming')}
			</span>
		{/if}
	</div>

	<!-- Team A -->
	<div class="flex flex-col items-center gap-3 text-center">
		<div class="flex justify-center">
			{#each teamA.players as player, i}
				<div class={i > 0 ? '-ml-3' : ''}>
					<Avatar
						icon={player.avatar_icon}
						color={player.avatar_color}
						name={player.name}
						size="md"
					/>
				</div>
			{/each}
		</div>
		<p class="text-[15px] font-semibold">{teamA.name}</p>
		{#if editable}
			<div class="flex items-center gap-5">
				<button
					onclick={() => onAdjust('a', -1)}
					disabled={teamA.score === 0}
					class="bg-surface-raised flex h-11 w-11 items-center justify-center rounded-full text-2xl font-bold transition-all active:scale-95 disabled:opacity-40"
					aria-label="Decrease Team A score">−</button
				>
				<button
					onclick={() => onScoreTap('a')}
					class="text-text-primary text-5xl font-[800] tabular-nums transition-transform active:scale-95"
					aria-label="Set Team A score">{teamA.score}</button
				>
				<button
					onclick={() => onAdjust('a', 1)}
					disabled={teamA.score + teamB.score >= pointsTarget}
					class="bg-surface-raised flex h-11 w-11 items-center justify-center rounded-full text-2xl font-bold transition-all active:scale-95 disabled:opacity-40"
					aria-label="Increase Team A score">+</button
				>
			</div>
		{:else}
			<p class="text-5xl font-[800] tabular-nums {scoreColor(teamA.score, teamB.score)}">
				{teamA.score}
			</p>
		{/if}
	</div>

	<div class="bg-border my-4 h-px"></div>

	<!-- Team B -->
	<div class="flex flex-col items-center gap-3 text-center">
		{#if editable}
			<div class="flex items-center gap-5">
				<button
					onclick={() => onAdjust('b', -1)}
					disabled={teamB.score === 0}
					class="bg-surface-raised flex h-11 w-11 items-center justify-center rounded-full text-2xl font-bold transition-all active:scale-95 disabled:opacity-40"
					aria-label="Decrease Team B score">−</button
				>
				<button
					onclick={() => onScoreTap('b')}
					class="text-text-primary text-5xl font-[800] tabular-nums transition-transform active:scale-95"
					aria-label="Set Team B score">{teamB.score}</button
				>
				<button
					onclick={() => onAdjust('b', 1)}
					disabled={teamA.score + teamB.score >= pointsTarget}
					class="bg-surface-raised flex h-11 w-11 items-center justify-center rounded-full text-2xl font-bold transition-all active:scale-95 disabled:opacity-40"
					aria-label="Increase Team B score">+</button
				>
			</div>
		{:else}
			<p class="text-5xl font-[800] tabular-nums {scoreColor(teamB.score, teamA.score)}">
				{teamB.score}
			</p>
		{/if}
		<p class="text-[15px] font-semibold">{teamB.name}</p>
		<div class="flex justify-center">
			{#each teamB.players as player, i}
				<div class={i > 0 ? '-ml-3' : ''}>
					<Avatar
						icon={player.avatar_icon}
						color={player.avatar_color}
						name={player.name}
						size="md"
					/>
				</div>
			{/each}
		</div>
	</div>

	<!-- Finalize (admin only) -->
	{#if editable}
		<div class="mt-6">
			<Button
				variant="default"
				size="cta"
				disabled={!canFinalize || submitting}
				onclick={onFinalize}
			>
				{submitting ? '…' : $_('active_finalize_result')}
			</Button>
		</div>
	{/if}
</Card>
