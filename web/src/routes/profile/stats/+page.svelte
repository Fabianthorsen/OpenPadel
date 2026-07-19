<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth.svelte';
	import { api } from '$lib/api/client';
	import { _ } from 'svelte-i18n';
	import { ChevronLeft } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';
	import { translateApiError } from '$lib/i18n/errors';
	import Footer from '$lib/components/Footer.svelte';
	import { Spinner } from '$lib/components/ui/spinner';
	import { Section } from '$lib/components/ui/section';
	import { SectionLabel } from '$lib/components/ui/section-label';
	import { StatTile, ModeSection, StatLegend, MODE_METRICS } from '$lib/components/ui/stats';

	let summary = $state<App.CareerSummary | null>(null);
	let modes = $state<App.ModeStats[]>([]);
	let loading = $state(true);
	let showLegend = $state(false);

	async function load() {
		if (!auth.token) return;
		try {
			const [profileRes, statsRes] = await Promise.all([
				api.auth.profile(auth.token),
				api.auth.stats(auth.token)
			]);
			summary = profileRes.stats;
			modes = statsRes.modes;
		} catch (e) {
			// Mirror the profile page: surface the failure rather than silently
			// rendering an empty hero indistinguishable from a real zero-games user.
			toast.error(
				e instanceof Error ? translateApiError(e.message) : translateApiError('server_error')
			);
		} finally {
			loading = false;
		}
	}

	onMount(async () => {
		if (!auth.token) {
			goto('/auth');
			return;
		}
		await load();
	});

	// The cross-mode hero mirrors the profile summary (ADR 0007): numbers are
	// hidden (em dash) at zero games so an empty career reads as intentional
	// rather than a jarring 0%.
	const hasGames = $derived(!!summary && summary.games > 0);
	const pointWinPct = $derived(hasGames ? `${Math.round(summary!.point_win_pct)}%` : '–');
	const winRate = $derived(hasGames ? `${Math.round(summary!.winrate)}%` : '–');
	const games = $derived(summary ? `${summary.games}` : '–');

	// A mode's translated section title from its id, so the page stays fully
	// data-driven off the per-mode aggregates.
	function modeTitle(mode: App.ModeStats['mode']): string {
		return mode === 'mexicano' ? $_('stats_mode_mexicano') : $_('stats_mode_americano');
	}
</script>

<main class="pt-safe-page mx-auto max-w-[480px] space-y-8 px-6 pb-10">
	<!-- Header -->
	<div class="flex items-center gap-4">
		<a
			href="/profile"
			class="text-text-secondary hover:text-text-primary transition-colors"
			aria-label="Back"
		>
			<ChevronLeft size={24} />
		</a>
		<h1 class="text-2xl font-[800]">{$_('stats_page_title')}</h1>
	</div>

	{#if loading}
		<div class="flex justify-center py-12">
			<Spinner />
		</div>
	{:else}
		<!-- Cross-mode summary hero: the three numbers that stay honest blended
		     across game modes (ADR 0007). -->
		<section class="space-y-3">
			<SectionLabel>{$_('stats_summary_title')}</SectionLabel>
			<div class="grid grid-cols-3 gap-3">
				<StatTile value={pointWinPct} label={$_('profile_point_win_pct')} accent />
				<StatTile value={winRate} label={$_('profile_win_rate')} />
				<StatTile value={games} label={$_('profile_games')} />
			</div>
		</section>

		<!-- One section per Game Mode, aggregated separately so every number is
		     compared like-for-like within one scoring model. -->
		{#each modes as mode (mode.mode)}
			<ModeSection title={modeTitle(mode.mode)} stats={mode} metrics={MODE_METRICS} />
		{/each}

		<!-- Collapsible key so the per-mode metrics are explained on demand
		     without crowding the numbers. -->
		<Section title={$_('stats_legend_title')} bind:open={showLegend}>
			{#snippet children()}
				<StatLegend metrics={MODE_METRICS} />
			{/snippet}
		</Section>
	{/if}

	<Footer />
</main>
