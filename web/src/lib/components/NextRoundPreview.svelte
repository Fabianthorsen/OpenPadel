<script lang="ts">
	import { _ } from 'svelte-i18n';
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';
	import type { Round, Match } from '$lib/types';

	let {
		currentRound,
		courts = 1,
		session,
		sessionId
	}: {
		currentRound: number;
		courts: number;
		session: any;
		sessionId: string;
	} = $props();

	let nextRound = $state<Round | null>(null);
	let loading = $state(true);

	onMount(async () => {
		try {
			const response = await api.rounds.list(sessionId);
			const allRounds = response.rounds || [];
			nextRound = allRounds.find((r: any) => r.round_number === currentRound + 1) || null;
		} catch (e) {
			console.error('Failed to load next round:', e);
		} finally {
			loading = false;
		}
	});

	const courtAssignments = $derived.by(() => {
		if (!nextRound || loading) return [];

		const assignments: { court: number; matches: Match[] }[] = [];

		for (let c = 1; c <= courts; c++) {
			const courtMatches = (nextRound as any).matches.filter((m: any) => m.court === c);
			if (courtMatches.length > 0) {
				assignments.push({ court: c, matches: courtMatches });
			}
		}

		return assignments;
	});

	const benchedPlayers = $derived.by(() => {
		if (!nextRound || loading) return [];

		const allPlayers = new Set(session?.players?.map((p: any) => p.id) || []);
		const playingPlayers = new Set();

		(nextRound as any).matches.forEach((match: any) => {
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

{#if loading}
	<div class="bg-surface-raised rounded-lg p-4 animate-pulse">
		<p class="text-xs font-semibold text-text-secondary mb-3">Loading next round...</p>
	</div>
{:else if nextRound}
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
