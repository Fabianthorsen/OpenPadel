<script lang="ts">
	import { goto } from '$app/navigation';
	import { api, ApiError } from '$lib/api/client';
	import { savePlayerSession } from '$lib/playerSession';
	import { auth } from '$lib/auth.svelte';
	import { _ } from 'svelte-i18n';
	import { translateApiError } from '$lib/i18n/errors';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { SegmentedControl, type SegmentedOption } from '$lib/components/ui/segmented-control';
	import * as Drawer from '$lib/components/ui/drawer';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import { Users, Check } from '@lucide/svelte';

	let {
		open = $bindable(false),
		club = null
	}: { open?: boolean; club?: { id: string; name: string } | null } = $props();

	let gameMode = $state<'americano' | 'mexicano'>('americano');
	let name = $state('');
	let creating = $state(false);
	let error = $state('');

	// Second creation flow: from the generic "New Tournament" button (no preset
	// club), a member can optionally attach the game to one of their clubs, making
	// it a club event. A preset club (from the club page) always wins and hides
	// the picker.
	let myClubs = $state<App.ClubListItem[]>([]);
	let clubsLoaded = $state(false);
	let selectedClubId = $state('');

	// Load the caller's clubs the first time the drawer opens for a personal
	// creation. api.clubs.list only returns clubs they're a member of, which is
	// exactly what the create endpoint will accept as club_id.
	$effect(() => {
		if (open && !club && auth.token && !clubsLoaded) {
			clubsLoaded = true;
			api.clubs
				.list(auth.token)
				.then((v) => {
					myClubs = v;
				})
				.catch(() => {});
		}
	});

	// The club the new session will belong to: a preset always wins, else the
	// picked club, else none (personal).
	const effectiveClubId = $derived(club?.id ?? selectedClubId);
	const effectiveClubName = $derived(
		club?.name ?? myClubs.find((c) => c.id === selectedClubId)?.name ?? ''
	);

	const modeOptions: SegmentedOption[] = [
		{ value: 'americano', label: 'Americano' },
		{ value: 'mexicano', label: 'Mexicano' }
	];

	async function create() {
		if (!auth.user) {
			error = 'Not authenticated';
			return;
		}

		creating = true;
		error = '';
		try {
			const defaults =
				gameMode === 'mexicano'
					? { courts: 2, points: 24, rounds_total: 7 }
					: { courts: 2, points: 24 };
			const session = await api.sessions.create(
				{
					game_mode: gameMode,
					name: name.trim(),
					...defaults,
					...(effectiveClubId ? { club_id: effectiveClubId } : {})
				},
				auth.token ?? undefined
			);
			const adminToken = session.admin_token!;
			localStorage.setItem(`admin_token_${session.id}`, adminToken);
			const player = await api.players.join(
				session.id,
				auth.user.display_name,
				auth.token ?? undefined,
				adminToken
			);
			savePlayerSession(session.id, player);
			localStorage.setItem('last_session_id', session.id);
			open = false;
			goto(`/s/${session.id}?token=${adminToken}`);
		} catch (e) {
			error =
				e instanceof ApiError ? translateApiError(e.message) : translateApiError('server_error');
			creating = false;
		}
	}
</script>

<Drawer.Root bind:open>
	<Drawer.Content class="mx-auto w-full max-w-[480px] overflow-hidden">
		<Drawer.Header>
			<div class="flex w-full items-center justify-between">
				<h2 class="text-lg font-[800]">
					{#if club}{$_('create_club_event_title')}{:else}{$_('create_title_line1')}
						{$_('create_title_line2')}{/if}
				</h2>
				<Drawer.Close
					class="bg-surface-raised text-text-secondary hover:bg-border hidden h-8 w-8 items-center justify-center rounded-full text-xl leading-none transition-colors md:flex"
					>×</Drawer.Close
				>
			</div>
		</Drawer.Header>

		<div class="flex-1 space-y-6 overflow-y-auto px-6 pb-8">
			{#if effectiveClubId}
				<!-- Owned by a Club (preset or picked): the whole roster is told about it
				     automatically — no personal invites needed. -->
				<div class="bg-primary/10 text-primary flex items-center gap-2 rounded-2xl px-4 py-3">
					<Users size={16} class="shrink-0" />
					<p class="text-sm font-semibold">
						{$_('create_club_event_banner', { values: { club: effectiveClubName } })}
					</p>
				</div>
			{/if}

			{#if !club && myClubs.length > 0}
				<!-- Optional: attach this game to one of the caller's clubs, turning it
				     into a club event. Default is a personal game. Hidden when opened
				     from a club page (club preset) or when the caller has no clubs. -->
				<div class="space-y-2.5">
					<p class="text-text-disabled text-[11px] font-semibold tracking-[0.1em] uppercase">
						{$_('create_attach_club_label')}
					</p>
					<div class="space-y-1.5">
						<button
							type="button"
							onclick={() => (selectedClubId = '')}
							aria-pressed={selectedClubId === ''}
							class="flex w-full items-center gap-3 rounded-2xl border px-4 py-3 text-left transition-colors {selectedClubId ===
							''
								? 'border-primary bg-primary/5'
								: 'border-border bg-surface-raised'}"
						>
							<span class="flex-1 text-sm font-semibold">{$_('create_attach_club_none')}</span>
							{#if selectedClubId === ''}<Check size={16} class="text-primary shrink-0" />{/if}
						</button>
						{#each myClubs as c (c.id)}
							<button
								type="button"
								onclick={() => (selectedClubId = c.id)}
								aria-pressed={selectedClubId === c.id}
								class="flex w-full items-center gap-3 rounded-2xl border px-4 py-3 text-left transition-colors {selectedClubId ===
								c.id
									? 'border-primary bg-primary/5'
									: 'border-border bg-surface-raised'}"
							>
								<Avatar icon={c.avatar_icon} color={c.avatar_color} name={c.name} size="sm" />
								<span class="flex-1 truncate text-sm font-semibold">{c.name}</span>
								{#if selectedClubId === c.id}<Check size={16} class="text-primary shrink-0" />{/if}
							</button>
						{/each}
					</div>
				</div>
			{/if}

			<!-- Game mode -->
			<div class="space-y-3">
				<SegmentedControl
					options={modeOptions}
					bind:value={gameMode}
					ariaLabel={$_('create_game_mode_label')}
				/>
				<p class="text-text-secondary text-sm">
					{gameMode === 'mexicano' ? $_('create_mexicano_hint') : $_('create_americano_hint')}
				</p>
			</div>

			<!-- Session name (optional) -->
			<div class="space-y-2.5">
				<p class="text-text-disabled text-[11px] font-semibold tracking-[0.1em] uppercase">
					{$_('create_tournament_name_label')}
				</p>
				<Input
					bind:value={name}
					placeholder={$_('create_tournament_name_placeholder')}
					maxlength={48}
					disabled={creating}
					class="bg-surface-raised rounded-2xl border-0 px-4 py-3.5 text-sm"
				/>
			</div>

			{#if error}
				<p class="text-destructive text-sm">{error}</p>
			{/if}

			<Button onclick={create} disabled={creating} size="cta" variant="default">
				{#if creating}{$_('create_button_loading')}{:else if effectiveClubId}{$_(
						'create_club_event_button'
					)}{:else}{$_('create_button')}{/if}
			</Button>
		</div>
	</Drawer.Content>
</Drawer.Root>
