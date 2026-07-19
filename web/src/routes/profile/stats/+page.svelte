<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth.svelte';
	import { api } from '$lib/api/client';
	import { _ } from 'svelte-i18n';
	import { ChevronLeft, Info } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';
	import { translateApiError } from '$lib/i18n/errors';
	import Footer from '$lib/components/Footer.svelte';
	import { Spinner } from '$lib/components/ui/spinner';
	import * as Dialog from '$lib/components/ui/dialog';
	import { SectionLabel } from '$lib/components/ui/section-label';
	import {
		StatTile,
		ModeSection,
		FormStrip,
		StatLegend,
		MODE_METRICS,
		FORM_WINDOW
	} from '$lib/components/ui/stats';

	let summary = $state<App.CareerSummary | null>(null);
	let modes = $state<App.ModeStats[]>([]);
	let series = $state<App.MatchResult[]>([]);
	let loading = $state(true);
	let showInfo = $state(false);

	async function load() {
		if (!auth.token) return;
		try {
			const [profileRes, statsRes] = await Promise.all([
				api.auth.profile(auth.token),
				api.auth.stats(auth.token)
			]);
			summary = profileRes.stats;
			modes = statsRes.modes;
			series = statsRes.series;
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

	// Placement stats blend both modes (a finishing rank compares like-for-like
	// where point-share does not — ADR 0007). Hidden (em dash) at zero games, same
	// as the hero, so an empty career never shows a hollow "1st / 0 titles".
	const titles = $derived(hasGames ? `${summary!.titles}` : '–');
	const podiums = $derived(hasGames ? `${summary!.podiums}` : '–');
	const bestFinish = $derived(hasGames ? `${summary!.best_finish}` : '–');
	const averageFinish = $derived(hasGames ? summary!.average_finish.toFixed(1) : '–');

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
		<h1 class="flex-1 text-2xl font-[800]">{$_('stats_page_title')}</h1>
		<button
			type="button"
			class="text-text-secondary hover:text-text-primary transition-colors"
			aria-label={$_('stats_legend_title')}
			onclick={() => (showInfo = true)}
		>
			<Info size={22} />
		</button>
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

		<!-- Cross-mode placement: titles, podiums and finishing position blend both
		     modes, since a finishing rank compares like-for-like (ADR 0007). -->
		<section class="space-y-3">
			<SectionLabel>{$_('stats_placement_title')}</SectionLabel>
			<div class="grid grid-cols-2 gap-3">
				<StatTile value={titles} label={$_('stats_titles')} />
				<StatTile value={podiums} label={$_('stats_podiums')} />
				<StatTile value={bestFinish} label={$_('stats_best_finish')} />
				<StatTile value={averageFinish} label={$_('stats_average_finish')} />
			</div>
		</section>

		<!-- One section per Game Mode, aggregated separately so every number is
		     compared like-for-like within one scoring model. -->
		<!-- Cross-mode recent form: one sparkline over the player's most recent
		     matches, both modes mixed (form is mode-agnostic). The header names the
		     actual number of games shown. Hidden until there's history so an empty
		     career never shows a hollow strip. -->
		{#if series.length > 0}
			<section class="space-y-3">
				<SectionLabel>
					{$_('stats_form_title', {
						values: { count: Math.min(series.length, FORM_WINDOW) }
					})}
				</SectionLabel>
				<FormStrip {series} />
			</section>
		{/if}

		{#each modes as mode (mode.mode)}
			<ModeSection title={modeTitle(mode.mode)} stats={mode} metrics={MODE_METRICS} />
		{/each}
	{/if}

	<Footer />
</main>

<!-- The metric key lives behind the header info button rather than a section, so
     the numbers aren't crowded by explanations the reader rarely needs. -->
<Dialog.Root bind:open={showInfo}>
	<Dialog.Content class="max-h-[85vh] gap-5 overflow-y-auto rounded-2xl p-6 sm:max-w-[420px]">
		<Dialog.Header class="space-y-0">
			<Dialog.Title class="text-text-primary text-xl font-[800]">
				{$_('stats_legend_title')}
			</Dialog.Title>
		</Dialog.Header>

		<!-- Form curve: the least self-evident element, so it leads, with a colour
		     key that reinforces the win/draw/loss read. -->
		<section class="space-y-2.5">
			<h3 class="text-text-primary text-sm font-semibold">{$_('stats_form_legend_label')}</h3>
			<p class="text-text-secondary text-sm leading-relaxed">{$_('stats_form_desc')}</p>
			<div class="text-text-secondary flex flex-wrap gap-x-4 gap-y-1.5 pt-0.5 text-xs font-medium">
				<span class="inline-flex items-center gap-1.5">
					<span class="bg-positive size-2.5 rounded-full"></span>{$_('stats_result_win')}
				</span>
				<span class="inline-flex items-center gap-1.5">
					<span class="bg-warning size-2.5 rounded-full"></span>{$_('stats_result_draw')}
				</span>
				<span class="inline-flex items-center gap-1.5">
					<span class="bg-destructive size-2.5 rounded-full"></span>{$_('stats_result_loss')}
				</span>
			</div>
		</section>

		<hr class="border-border" />

		<!-- Per-mode metric key, driven by the same catalog as the stat grids. -->
		<StatLegend metrics={MODE_METRICS} />
	</Dialog.Content>
</Dialog.Root>
