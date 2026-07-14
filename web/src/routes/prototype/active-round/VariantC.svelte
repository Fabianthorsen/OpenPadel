<!-- PROTOTYPE — wipe me. Variant C: dense scoreboard (whole round as a compact venue-style board). -->
<script lang="ts">
	import RoundIndicator from '$lib/components/RoundIndicator.svelte';
	import { Clock } from '@lucide/svelte';
	import TopBar from './TopBar.svelte';
	import BottomNav from './BottomNav.svelte';
	import { playerById, teamLabel, statusOf, timeLeft, type Match } from './mock';

	let { session, round, isAdmin }: { session: any; round: any; isAdmin: boolean } = $props();
</script>

<div class="pb-28">
	<main class="mx-auto w-full max-w-[480px] px-4 pt-4">
		<TopBar name={session.name} />

		<div class="mt-4 flex items-end justify-between">
			<div>
				<p class="text-text-disabled text-[11px] font-bold tracking-[0.1em] uppercase">
					{session.game_mode}
				</p>
				<h2 class="text-text-primary text-[28px] leading-none font-[800] tracking-tight">
					Round {round.number}<span class="text-text-disabled text-lg font-semibold">
						/ {session.rounds_total}</span
					>
				</h2>
			</div>
			<span
				class="text-text-secondary inline-flex items-center gap-1 font-mono text-xs font-semibold"
				><Clock size={13} /> {timeLeft}</span
			>
		</div>
		<div class="mt-3"><RoundIndicator current={round.number} total={session.rounds_total} /></div>

		<div class="border-border mt-4 overflow-hidden rounded-2xl border">
			{#each round.matches as m, i (m.id)}
				{@const st = statusOf(m as Match)}
				{@const sa = m.score?.a ?? m.live?.a ?? 0}
				{@const sb = m.score?.b ?? m.live?.b ?? 0}
				{@const up = st === 'upcoming'}
				<div
					class="{i > 0 ? 'border-border border-t' : ''} {st === 'live'
						? 'bg-primary-muted/40'
						: 'bg-surface'}"
				>
					<div class="flex items-center gap-3 px-4 py-3">
						<div class="flex w-9 shrink-0 flex-col items-center">
							<span class="text-text-disabled text-[9px] font-bold tracking-wide uppercase"
								>Court</span
							>
							<span class="text-text-primary text-lg font-[800] tabular-nums">{m.court}</span>
						</div>
						<div class="min-w-0 flex-1 space-y-1.5">
							<div class="flex items-center justify-between gap-2">
								<span
									class="text-text-primary truncate text-sm {st === 'final' && sa > sb
										? 'font-bold'
										: 'font-medium'}">{teamLabel(m.team_a)}</span
								>
								<span
									class="shrink-0 text-base font-[800] tabular-nums {up
										? 'text-text-disabled'
										: 'text-text-primary'}">{up ? '–' : sa}</span
								>
							</div>
							<div class="flex items-center justify-between gap-2">
								<span
									class="text-text-primary truncate text-sm {st === 'final' && sb > sa
										? 'font-bold'
										: 'font-medium'}">{teamLabel(m.team_b)}</span
								>
								<span
									class="shrink-0 text-base font-[800] tabular-nums {up
										? 'text-text-disabled'
										: 'text-text-primary'}">{up ? '–' : sb}</span
								>
							</div>
						</div>
						<div class="w-12 shrink-0 text-right">
							{#if st === 'live'}
								<span class="text-primary text-[10px] font-bold tracking-wide uppercase"
									>● Live</span
								>
							{:else if st === 'final'}
								<span class="text-text-disabled text-[10px] font-bold tracking-wide uppercase"
									>Final</span
								>
							{:else}
								<span class="text-text-disabled text-[10px] font-semibold uppercase">—</span>
							{/if}
						</div>
					</div>
					{#if isAdmin}
						<button
							class="text-primary hover:bg-surface-raised border-border/60 w-full border-t py-2 text-xs font-semibold transition-colors"
							>{st === 'final' ? 'Edit' : 'Enter score'}</button
						>
					{/if}
				</div>
			{/each}
		</div>

		<div class="text-text-secondary mt-3 flex items-center gap-2 px-1 text-xs">
			<span class="text-text-disabled font-bold tracking-wide uppercase">Bench</span>
			<span class="text-text-primary font-medium"
				>{round.bench.map((id: string) => playerById[id]?.name).join(', ')}</span
			>
		</div>

		{#if isAdmin}
			<button
				class="bg-primary text-primary-foreground mt-4 w-full rounded-2xl py-4 text-[15px] font-bold disabled:opacity-40"
				disabled>Next round →</button
			>
		{/if}
	</main>
	<BottomNav active="scoring" />
</div>
