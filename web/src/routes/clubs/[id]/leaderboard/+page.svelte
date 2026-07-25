<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api/client';
	import { auth } from '$lib/auth.svelte';
	import { Button } from '$lib/components/ui/button';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import { Spinner } from '$lib/components/ui/spinner';
	import Footer from '$lib/components/Footer.svelte';
	import { shortName, formScore } from '$lib/utils';
	import { toast } from 'svelte-sonner';
	import { onMount } from 'svelte';

	let clubId = $state('');
	let club = $state<App.ClubDetail | null>(null);
	let board = $state<App.ClubLeaderboard | null>(null);
	let loading = $state(true);
	let error = $state('');

	async function load(id: string) {
		if (!auth.token) {
			goto('/auth');
			return;
		}
		try {
			loading = true;
			// The club detail gives us the name for the header and doubles as the
			// membership gate — a non-member 403s here just like the board endpoint.
			[club, board] = await Promise.all([
				api.clubs.detail(auth.token, id),
				api.clubs.leaderboard(auth.token, id)
			]);
		} catch (err: any) {
			if (err.status === 403) {
				error = 'You are not a member of this club';
			} else if (err.status === 404) {
				error = 'Club not found';
			} else {
				error = 'Failed to load leaderboard';
			}
			toast.error(error);
		} finally {
			loading = false;
		}
	}

	const ranked = $derived(board?.ranked ?? []);
	const provisional = $derived(board?.provisional ?? []);

	onMount(() => {
		clubId = window.location.pathname.split('/')[2];
		if (clubId) load(clubId);
	});
</script>

<main class="pt-safe-page mx-auto max-w-[480px] space-y-6 px-6 pb-10">
	{#if loading}
		<div class="flex justify-center py-12">
			<Spinner />
		</div>
	{:else if error}
		<div class="space-y-6 py-12">
			<p class="text-destructive text-center text-sm font-semibold">{error}</p>
			<Button onclick={() => goto('/profile')} variant="default" size="cta">Back to Profile</Button>
		</div>
	{:else if club && board}
		<PageHeader
			title="Leaderboard"
			backHref={`/clubs/${clubId}`}
			avatar={{ icon: club.club.avatar_icon, color: club.club.avatar_color, name: club.club.name }}
			subtitle={club.club.name}
		>
			<p class="text-text-secondary text-sm leading-relaxed">
				Ranked by current form — your average points margin per game over the last 90 days.
			</p>
		</PageHeader>

		{#if ranked.length === 0 && provisional.length === 0}
			<div class="bg-surface-raised rounded-2xl px-4 py-8 text-center">
				<p class="text-text-secondary text-sm">
					No games yet. Play a club game and the leaderboard fills in — members rank once they've
					played {board.min_games}.
				</p>
			</div>
		{:else}
			<!-- Ranked board: rank · avatar+name · form (calm, rank-1 in primary). -->
			{#if ranked.length > 0}
				<section class="space-y-0.5">
					{#each ranked as e (e.user_id)}
						{@const isRank1 = e.rank === 1}
						<div
							class="bg-surface flex items-center gap-3 rounded-2xl px-4 py-3.5"
							aria-label="{e.rank}. {e.name}: form {formScore(e.form)}"
						>
							<span
								class="w-6 text-sm font-[800] tabular-nums {isRank1
									? 'text-primary'
									: 'text-text-disabled'}"
							>
								{e.rank}
							</span>
							<div class="flex min-w-0 flex-1 items-center gap-2.5">
								<Avatar
									icon={e.avatar_icon}
									color={e.avatar_color}
									name={e.name}
									size="sm"
									ring="ring-2 ring-primary/30"
								/>
								<div class="min-w-0">
									<span
										class="block truncate text-sm {isRank1
											? 'text-primary font-bold'
											: 'text-text-primary font-semibold'}"
									>
										{shortName(e.name)}
									</span>
									<span class="flex items-center gap-1 text-[11px] font-bold tabular-nums">
										<span class="text-primary">{e.wins}W</span>
										<span class="text-text-disabled">·</span>
										<span class="text-text-disabled">{e.draws}D</span>
										<span class="text-text-disabled">·</span>
										<span class="text-destructive">{e.losses}L</span>
									</span>
								</div>
							</div>
							<span class="text-text-primary text-base font-[800] tabular-nums">
								{formScore(e.form)}
							</span>
						</div>
					{/each}
				</section>
			{/if}

			<!-- Provisional: members who haven't played enough to rank yet. -->
			{#if provisional.length > 0}
				<section class="space-y-2">
					<p class="text-text-disabled px-1 text-[11px] font-bold tracking-[0.1em] uppercase">
						Not yet ranked
					</p>
					<div class="space-y-0.5">
						{#each provisional as p (p.user_id)}
							<div class="bg-surface flex items-center gap-3 rounded-2xl px-4 py-3">
								<span class="w-6"></span>
								<div class="flex min-w-0 flex-1 items-center gap-2.5">
									<Avatar icon={p.avatar_icon} color={p.avatar_color} name={p.name} size="sm" />
									<span class="text-text-secondary truncate text-sm font-semibold">
										{shortName(p.name)}
									</span>
								</div>
								<span class="text-text-disabled text-xs font-semibold">
									{p.games_to_go} more to rank
								</span>
							</div>
						{/each}
					</div>
				</section>
			{/if}
		{/if}
	{/if}

	<Footer />
</main>
