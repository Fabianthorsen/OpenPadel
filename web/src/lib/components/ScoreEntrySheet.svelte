<script lang="ts">
	import { _ } from 'svelte-i18n';
	import { toast } from 'svelte-sonner';
	import * as Sheet from '$lib/components/ui/sheet';
	import { Button } from '$lib/components/ui/button';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import { Check } from 'lucide-svelte';

	interface Props {
		matchId: string;
		courtNumber: number;
		roundNumber: number;
		pointsTarget: number;
		teamAName: string;
		teamBName: string;
		teamAPlayers?: Array<{ avatar_icon: string; avatar_color: string; name: string }>;
		teamBPlayers?: Array<{ avatar_icon: string; avatar_color: string; name: string }>;
		initialScoreA?: number;
		initialScoreB?: number;
		onSubmit: (scoreA: number, scoreB: number) => Promise<void>;
		onClose: () => void;
		onLiveSave: (scoreA: number, scoreB: number) => void;
	}

	let {
		matchId,
		courtNumber,
		roundNumber,
		pointsTarget,
		teamAName,
		teamBName,
		teamAPlayers = [],
		teamBPlayers = [],
		initialScoreA = 0,
		initialScoreB = 0,
		onSubmit,
		onClose,
		onLiveSave
	}: Props = $props();

	let scoreA = $state(initialScoreA);
	let scoreB = $state(initialScoreB);
	let focusedTeam = $state<'a' | 'b' | null>(null);
	let keypadValue = $state('');
	let submitting = $state(false);
	let shaking = $state(false);

	const isValid = $derived(scoreA + scoreB === pointsTarget);
	const sum = $derived(scoreA + scoreB);
	const isOver = $derived(sum > pointsTarget);

	function updateScore(team: 'a' | 'b', value: number) {
		const clamped = Math.max(0, Math.min(pointsTarget, value));
		if (team === 'a') {
			scoreA = clamped;
		} else {
			scoreB = clamped;
		}
		onLiveSave(scoreA, scoreB);
	}

	function adjustScore(team: 'a' | 'b', delta: number) {
		const current = team === 'a' ? scoreA : scoreB;
		updateScore(team, current + delta);
	}

	function focusScore(team: 'a' | 'b') {
		focusedTeam = team;
		keypadValue = '';
	}

	function keypadDigit(d: string) {
		if (!focusedTeam) return;

		let next: string;
		if (keypadValue === '' || (keypadValue === '0' && d !== '0')) {
			next = d;
		} else {
			next = (keypadValue + d).replace(/^0+(\d)/, '$1');
		}

		const parsed = parseInt(next || '0');
		if (parsed > pointsTarget) {
			shaking = true;
			setTimeout(() => {
				shaking = false;
			}, 400);
			return;
		}

		keypadValue = next;
	}

	function keypadDelete() {
		keypadValue = keypadValue.slice(0, -1);
	}

	function keypadConfirm() {
		if (!focusedTeam) return;

		const entered = parseInt(keypadValue || '0');
		if (entered > pointsTarget) {
			shaking = true;
			setTimeout(() => {
				shaking = false;
			}, 400);
			return;
		}

		// Auto-complement
		const otherScore = pointsTarget - entered;
		if (focusedTeam === 'a') {
			scoreA = entered;
			scoreB = otherScore;
		} else {
			scoreB = entered;
			scoreA = otherScore;
		}

		onLiveSave(scoreA, scoreB);
		focusedTeam = null;
		keypadValue = '';
	}

	async function handleSubmit() {
		if (!isValid || submitting) return;

		submitting = true;
		try {
			await onSubmit(scoreA, scoreB);
			toast.success($_('toast_score_confirmed'));
			onClose();
		} catch (e) {
			toast.error($_('api_error_server_error'));
		} finally {
			submitting = false;
		}
	}
</script>

<Sheet.Root open={true}>
	<Sheet.Content side="bottom" class="mx-auto w-full max-w-[480px]">
		<!-- Header -->
		<div class="flex items-center justify-between pb-4">
			<div>
				<p class="text-text-secondary text-xs font-semibold tracking-widest uppercase">
					{$_('active_court_label', { values: { number: courtNumber } })} · Round {roundNumber}
				</p>
			</div>
			<button
				onclick={onClose}
				class="text-text-disabled hover:bg-surface-raised flex h-8 w-8 shrink-0 items-center justify-center rounded-full transition-colors"
				aria-label="Close score entry"
			>
				×
			</button>
		</div>

		<!-- Validity readout -->
		<div class="mb-6 flex items-center justify-center gap-2">
			<span class="text-text-secondary font-mono text-sm">
				{scoreA} + {scoreB}
			</span>
			<span class={`font-mono text-sm ${isValid ? 'text-primary' : 'text-text-secondary'}`}>
				= {pointsTarget}
			</span>
			{#if isValid}
				<Check size={16} class="text-primary" />
			{/if}
		</div>

		<!-- Team A -->
		<div class="mb-6 flex flex-col items-center gap-3">
			{#if teamAPlayers.length > 0}
				<div class="flex justify-center">
					{#each teamAPlayers as player, i}
						<div class={i > 0 ? '-ml-3' : ''}>
							<Avatar
								icon={player.avatar_icon}
								color={player.avatar_color}
								name={player.name}
								size="sm"
							/>
						</div>
					{/each}
				</div>
			{/if}
			<p class="text-sm font-semibold">{teamAName}</p>
			<div class="flex items-center gap-4">
				<button
					onclick={() => adjustScore('a', -1)}
					disabled={scoreA === 0}
					class="bg-surface-raised flex h-12 w-12 items-center justify-center rounded-full text-xl font-bold transition-all active:scale-95 disabled:opacity-40"
					aria-label="Decrease Team A score"
				>
					−
				</button>
				<button
					onclick={() => focusScore('a')}
					class={`text-[64px] font-[800] tabular-nums transition-all ${
						focusedTeam === 'a' ? 'text-primary' : 'text-text-primary'
					}`}
					aria-label="Enter Team A score"
				>
					{scoreA}
				</button>
				<button
					onclick={() => adjustScore('a', 1)}
					disabled={scoreA + scoreB >= pointsTarget}
					class="bg-surface-raised flex h-12 w-12 items-center justify-center rounded-full text-xl font-bold transition-all active:scale-95 disabled:opacity-40"
					aria-label="Increase Team A score"
				>
					+
				</button>
			</div>
		</div>

		<!-- Team B -->
		<div class="mb-6 flex flex-col items-center gap-3">
			<div class="flex items-center gap-4">
				<button
					onclick={() => adjustScore('b', -1)}
					disabled={scoreB === 0}
					class="bg-surface-raised flex h-12 w-12 items-center justify-center rounded-full text-xl font-bold transition-all active:scale-95 disabled:opacity-40"
					aria-label="Decrease Team B score"
				>
					−
				</button>
				<button
					onclick={() => focusScore('b')}
					class={`text-[64px] font-[800] tabular-nums transition-all ${
						focusedTeam === 'b' ? 'text-primary' : 'text-text-primary'
					}`}
					aria-label="Enter Team B score"
				>
					{scoreB}
				</button>
				<button
					onclick={() => adjustScore('b', 1)}
					disabled={scoreA + scoreB >= pointsTarget}
					class="bg-surface-raised flex h-12 w-12 items-center justify-center rounded-full text-xl font-bold transition-all active:scale-95 disabled:opacity-40"
					aria-label="Increase Team B score"
				>
					+
				</button>
			</div>
			<p class="text-sm font-semibold">{teamBName}</p>
			{#if teamBPlayers.length > 0}
				<div class="flex justify-center">
					{#each teamBPlayers as player, i}
						<div class={i > 0 ? '-ml-3' : ''}>
							<Avatar
								icon={player.avatar_icon}
								color={player.avatar_color}
								name={player.name}
								size="sm"
							/>
						</div>
					{/each}
				</div>
			{/if}
		</div>

		<!-- Keypad (integrated) -->
		{#if focusedTeam}
			<div class="mb-6 space-y-3">
				<div class="grid grid-cols-3 gap-2">
					{#each ['1', '2', '3', '4', '5', '6', '7', '8', '9'] as digit}
						<button
							onclick={() => keypadDigit(digit)}
							class={`bg-surface-raised rounded-lg py-3 text-lg font-bold transition-all active:scale-95 ${
								shaking ? 'animate-shake' : ''
							}`}
							aria-label="Enter {digit}"
						>
							{digit}
						</button>
					{/each}
				</div>
				<div class="grid grid-cols-3 gap-2">
					<button
						onclick={() => keypadDelete()}
						class="bg-surface-raised rounded-lg py-3 text-lg font-bold transition-all active:scale-95"
						aria-label="Delete"
					>
						⌫
					</button>
					<button
						onclick={() => keypadDigit('0')}
						class="bg-surface-raised rounded-lg py-3 text-lg font-bold transition-all active:scale-95"
						aria-label="Enter 0"
					>
						0
					</button>
					<button
						onclick={() => keypadConfirm()}
						class="bg-primary text-primary-foreground rounded-lg py-3 text-lg font-bold transition-all active:scale-95"
						aria-label="Confirm"
					>
						✓
					</button>
				</div>
				<p class="text-text-secondary text-center text-xs">
					{keypadValue || '0'}
				</p>
			</div>
		{/if}

		<!-- Finalize button -->
		<Button variant="default" size="cta" disabled={!isValid || submitting} onclick={handleSubmit}>
			{submitting ? '…' : $_('active_finalize_result')}
		</Button>
		<p class="text-text-disabled pt-2 text-center text-xs">
			{$_('active_scores_synced')}
		</p>
	</Sheet.Content>
</Sheet.Root>
