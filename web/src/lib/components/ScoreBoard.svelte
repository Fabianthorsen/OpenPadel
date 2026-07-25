<script lang="ts">
	import { _ } from 'svelte-i18n';
	import { Card } from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import StepButton from './StepButton.svelte';
	import TeamScore from './TeamScore.svelte';

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
	const atTarget = $derived(teamA.score + teamB.score >= pointsTarget);
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
				class="bg-signal-muted text-signal-strong inline-flex items-center gap-1 rounded-full px-3 py-1 text-xs font-semibold"
			>
				<span class="bg-signal inline-block h-2 w-2 animate-pulse rounded-full"></span>
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
				<StepButton
					direction="decrease"
					label="Decrease Team A score"
					disabled={teamA.score === 0}
					onclick={() => onAdjust('a', -1)}
				/>
				<TeamScore
					score={teamA.score}
					opponentScore={teamB.score}
					scored={false}
					size="lg"
					tile
					interactive
					label="Set Team A score"
					onTap={() => onScoreTap('a')}
				/>
				<StepButton
					direction="increase"
					label="Increase Team A score"
					disabled={atTarget}
					onclick={() => onAdjust('a', 1)}
				/>
			</div>
		{:else}
			<TeamScore score={teamA.score} opponentScore={teamB.score} {scored} size="lg" tile />
		{/if}
	</div>

	<div class="bg-border my-4 h-px"></div>

	<!-- Team B -->
	<div class="flex flex-col items-center gap-3 text-center">
		{#if editable}
			<div class="flex items-center gap-5">
				<StepButton
					direction="decrease"
					label="Decrease Team B score"
					disabled={teamB.score === 0}
					onclick={() => onAdjust('b', -1)}
				/>
				<TeamScore
					score={teamB.score}
					opponentScore={teamA.score}
					scored={false}
					size="lg"
					tile
					interactive
					label="Set Team B score"
					onTap={() => onScoreTap('b')}
				/>
				<StepButton
					direction="increase"
					label="Increase Team B score"
					disabled={atTarget}
					onclick={() => onAdjust('b', 1)}
				/>
			</div>
		{:else}
			<TeamScore score={teamB.score} opponentScore={teamA.score} {scored} size="lg" tile />
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
