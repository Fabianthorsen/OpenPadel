<script lang="ts">
	import { _ } from 'svelte-i18n';
	import type { Round, Match } from '$lib/types';

	let {
		rounds,
		currentRound,
		courts = 1,
		session
	}: {
		rounds: Round[];
		currentRound: number;
		courts: number;
		session: any;
	} = $props();

	const nextRound = $derived(rounds.find(r => r.round_number === currentRound + 1));

	const courtAssignments = $derived.by(() => {
		if (!nextRound) return [];

		const assignments: { court: number; matches: Match[] }[] = [];

		for (let c = 1; c <= courts; c++) {
			const courtMatches = nextRound.matches.filter(m => m.court === c);
			if (courtMatches.length > 0) {
				assignments.push({ court: c, matches: courtMatches });
			}
		}

		return assignments;
	});

	const benchedPlayers = $derived.by(() => {
		if (!nextRound) return [];

		const allPlayers = new Set(session?.players?.map((p: any) => p.id) || []);
		const playingPlayers = new Set();

		nextRound.matches.forEach((match: Match) => {
			playingPlayers.add(match.player_a_id);
			playingPlayers.add(match.player_b_id);
			playingPlayers.add(match.player_c_id);
			playingPlayers.add(match.player_d_id);
		});

		return Array.from(allPlayers).filter(id => !playingPlayers.has(id));
	});

	function getPlayerName(playerId: string) {
		return session?.players?.find((p: any) => p.id === playerId)?.display_name || playerId;
	}
</script>

{#if nextRound}
	<div class="space-y-4">
		<div class="bg-surface-raised rounded-lg p-4">
			<p class="text-xs font-semibold text-text-secondary mb-3">
				Round {currentRound + 1} of {session?.rounds_total || '?'}
			</p>

			<div class="space-y-3">
				{#each courtAssignments as assignment (assignment.court)}
					<div class="space-y-1">
						<p class="text-sm font-semibold">🎾 Court {assignment.court}</p>
						{#each assignment.matches as match (match.id)}
							<p class="text-xs text-text-secondary pl-4">
								{getPlayerName(match.player_a_id)} + {getPlayerName(match.player_b_id)}
								<span class="text-text-tertiary">vs</span>
								{getPlayerName(match.player_c_id)} + {getPlayerName(match.player_d_id)}
							</p>
						{/each}
					</div>
				{/each}

				{#if benchedPlayers.length > 0}
					<div class="pt-2 border-t border-border space-y-1">
						<p class="text-sm font-semibold">🪑 Bench</p>
						<p class="text-xs text-text-secondary pl-4">
							{benchedPlayers.map(id => getPlayerName(id)).join(', ')}
						</p>
					</div>
				{/if}
			</div>
		</div>
	</div>
{/if}
