<!-- PROTOTYPE — wipe me. Variant B: focused one-court (court tabs, calmed version of the shipped design). -->
<script lang="ts">
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import RoundIndicator from '$lib/components/RoundIndicator.svelte';
	import { SectionLabel } from '$lib/components/ui/section-label';
	import { Clock } from '@lucide/svelte';
	import TopBar from './TopBar.svelte';
	import BottomNav from './BottomNav.svelte';
	import { playerById, teamLabel, statusOf, timeLeft, type Match } from './mock';

	let { session, round, isAdmin }: { session: any; round: any; isAdmin: boolean } = $props();

	let active = $state(0);
	const m = $derived(round.matches[active] as Match);
	const st = $derived(statusOf(m));
	const sa = $derived(m.score?.a ?? m.live?.a ?? 0);
	const sb = $derived(m.score?.b ?? m.live?.b ?? 0);
	const up = $derived(st === 'upcoming');
</script>

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

		<!-- court tabs -->
		<div class="mt-4 flex gap-2">
			{#each round.matches as mm, i (mm.id)}
				<button
					onclick={() => (active = i)}
					class="flex flex-1 items-center justify-center gap-1.5 rounded-xl py-2.5 text-[13px] font-bold tracking-wide uppercase transition-colors {active ===
					i
						? 'bg-primary text-primary-foreground'
						: 'bg-surface-raised text-text-secondary'}"
				>
					Court {mm.court}
					{#if mm.score !== null}<span
							class="h-1.5 w-1.5 rounded-full {active === i ? 'bg-white/60' : 'bg-primary'}"
						></span>{/if}
				</button>
			{/each}
		</div>

		<!-- focused court -->
		<div class="bg-surface border-border mt-4 rounded-3xl border p-6 shadow-sm">
			<div class="flex justify-center">
				{#if st === 'live'}
					<span
						class="text-primary inline-flex items-center gap-1.5 text-[11px] font-bold tracking-wide uppercase"
						><span class="bg-primary h-1.5 w-1.5 animate-pulse rounded-full"></span>Live</span
					>
				{:else if st === 'final'}
					<span class="text-text-disabled text-[11px] font-bold tracking-wide uppercase"
						>Final result</span
					>
				{:else}
					<span class="text-text-disabled text-[11px] font-semibold tracking-wide uppercase"
						>Not started</span
					>
				{/if}
			</div>

			<div class="mt-4 flex flex-col items-center gap-2">
				<div class="flex">
					<Avatar
						name={playerById[m.team_a[0]]?.name ?? ''}
						color={playerById[m.team_a[0]]?.avatar_color}
						icon={playerById[m.team_a[0]]?.avatar_icon}
						size="md"
						ring="ring-2 ring-surface"
					/>
					<div class="-ml-3">
						<Avatar
							name={playerById[m.team_a[1]]?.name ?? ''}
							color={playerById[m.team_a[1]]?.avatar_color}
							icon={playerById[m.team_a[1]]?.avatar_icon}
							size="md"
							ring="ring-2 ring-surface"
						/>
					</div>
				</div>
				<p class="text-text-primary text-[15px] font-semibold">{teamLabel(m.team_a)}</p>
				<p
					class="text-text-primary text-5xl font-[800] tabular-nums {up
						? 'text-text-disabled'
						: ''}"
				>
					{up ? '–' : sa}
				</p>
			</div>

			<div class="border-border my-4 border-t"></div>

			<div class="flex flex-col items-center gap-2">
				<p
					class="text-text-primary text-5xl font-[800] tabular-nums {up
						? 'text-text-disabled'
						: ''}"
				>
					{up ? '–' : sb}
				</p>
				<p class="text-text-primary text-[15px] font-semibold">{teamLabel(m.team_b)}</p>
				<div class="flex">
					<Avatar
						name={playerById[m.team_b[0]]?.name ?? ''}
						color={playerById[m.team_b[0]]?.avatar_color}
						icon={playerById[m.team_b[0]]?.avatar_icon}
						size="md"
						ring="ring-2 ring-surface"
					/>
					<div class="-ml-3">
						<Avatar
							name={playerById[m.team_b[1]]?.name ?? ''}
							color={playerById[m.team_b[1]]?.avatar_color}
							icon={playerById[m.team_b[1]]?.avatar_icon}
							size="md"
							ring="ring-2 ring-surface"
						/>
					</div>
				</div>
			</div>
		</div>

		{#if isAdmin}
			<button
				class="bg-primary text-primary-foreground mt-4 w-full rounded-2xl py-4 text-[15px] font-bold"
				>{st === 'final' ? 'Edit score' : 'Enter scores'}</button
			>
		{/if}

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
	</main>
	<BottomNav active="scoring" />
</div>
