<!-- PROTOTYPE — wipe me. Variant A: glanceable court list (all courts at once, calm cards, drill-in to score). -->
<script lang="ts">
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import RoundIndicator from '$lib/components/RoundIndicator.svelte';
	import { SectionLabel } from '$lib/components/ui/section-label';
	import { Clock } from '@lucide/svelte';
	import TopBar from './TopBar.svelte';
	import BottomNav from './BottomNav.svelte';
	import { playerById, teamLabel, statusOf, timeLeft, type Match } from './mock';

	let { session, round, isAdmin }: { session: any; round: any; isAdmin: boolean } = $props();
</script>

{#snippet teamRow(ids: [string, string], sc: number, win: boolean, upcoming: boolean)}
	<div class="flex items-center gap-3 px-4 py-3 {win ? 'bg-primary-muted' : ''}">
		<div class="flex">
			{#each ids as id, i (id)}
				<div class={i > 0 ? '-ml-2' : ''}>
					<Avatar
						name={playerById[id]?.name ?? ''}
						color={playerById[id]?.avatar_color}
						icon={playerById[id]?.avatar_icon}
						size="sm"
						ring="ring-2 ring-surface"
					/>
				</div>
			{/each}
		</div>
		<span class="text-text-primary flex-1 truncate text-[15px] {win ? 'font-bold' : 'font-medium'}"
			>{teamLabel(ids)}</span
		>
		<span
			class="text-2xl font-[800] tabular-nums {upcoming
				? 'text-text-disabled'
				: win
					? 'text-primary'
					: 'text-text-primary'}">{upcoming ? '–' : sc}</span
		>
	</div>
{/snippet}

<div class="pb-28">
	<main class="mx-auto w-full max-w-[480px] px-4 pt-4">
		<TopBar name={session.name} />

		<div class="mt-4">
			<SectionLabel
				>{session.game_mode} · Round {round.number} of {session.rounds_total}</SectionLabel
			>
			<div class="mt-2 flex items-center justify-between">
				<RoundIndicator current={round.number} total={session.rounds_total} />
				<span
					class="text-text-secondary inline-flex items-center gap-1 font-mono text-xs font-semibold"
					><Clock size={13} /> {timeLeft}</span
				>
			</div>
		</div>

		<div class="mt-4 space-y-3">
			{#each round.matches as m (m.id)}
				{@const st = statusOf(m as Match)}
				{@const sa = m.score?.a ?? m.live?.a ?? 0}
				{@const sb = m.score?.b ?? m.live?.b ?? 0}
				{@const up = st === 'upcoming'}
				<div class="bg-surface border-border overflow-hidden rounded-2xl border shadow-sm">
					<div class="flex items-center justify-between px-4 pt-3 pb-1">
						<span class="text-text-disabled text-[11px] font-bold tracking-[0.1em] uppercase"
							>Court {m.court}</span
						>
						{#if st === 'live'}
							<span
								class="text-primary inline-flex items-center gap-1.5 text-[11px] font-bold tracking-wide uppercase"
								><span class="bg-primary h-1.5 w-1.5 animate-pulse rounded-full"></span>Live</span
							>
						{:else if st === 'final'}
							<span class="text-text-disabled text-[11px] font-bold tracking-wide uppercase"
								>Final</span
							>
						{:else}
							<span class="text-text-disabled text-[11px] font-semibold tracking-wide uppercase"
								>Not started</span
							>
						{/if}
					</div>
					{@render teamRow(m.team_a, sa, st === 'final' && sa > sb, up)}
					<div class="border-border mx-4 border-t"></div>
					{@render teamRow(m.team_b, sb, st === 'final' && sb > sa, up)}
					{#if isAdmin}
						<button
							class="text-primary hover:bg-surface-raised border-border w-full border-t py-2.5 text-sm font-semibold transition-colors"
						>
							{st === 'final' ? 'Edit score' : st === 'live' ? 'Enter score →' : 'Start scoring →'}
						</button>
					{/if}
				</div>
			{/each}
		</div>

		<div class="mt-4">
			<SectionLabel class="mb-1.5">Bench</SectionLabel>
			<div class="bg-surface-raised flex items-center gap-2.5 rounded-xl px-4 py-2.5 text-sm">
				{#each round.bench as id (id)}
					<Avatar
						name={playerById[id]?.name ?? ''}
						color={playerById[id]?.avatar_color}
						icon={playerById[id]?.avatar_icon}
						size="sm"
					/>
					<span class="text-text-primary font-medium">{playerById[id]?.name}</span>
				{/each}
			</div>
		</div>

		{#if isAdmin}
			<button
				class="bg-primary text-primary-foreground mt-5 w-full rounded-2xl py-4 text-[15px] font-bold disabled:opacity-40"
				disabled>Next round →</button
			>
			<p class="text-text-disabled mt-1.5 text-center text-xs">1 court still needs a score</p>
		{/if}
	</main>
	<BottomNav active="scoring" />
</div>
