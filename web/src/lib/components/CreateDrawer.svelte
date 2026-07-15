<script lang="ts">
	import { goto } from '$app/navigation';
	import { api, ApiError } from '$lib/api/client';
	import { auth } from '$lib/auth.svelte';
	import { _ } from 'svelte-i18n';
	import { translateApiError } from '$lib/i18n/errors';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { ToggleGroup as RadioGroup, ToggleGroupItem } from '$lib/components/ui/toggle-group';
	import * as Drawer from '$lib/components/ui/drawer';

	let { open = $bindable(false) }: { open?: boolean } = $props();

	let gameMode = $state<'americano' | 'mexicano'>('americano');
	let name = $state('');
	let creating = $state(false);
	let error = $state('');

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
					...defaults
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
			localStorage.setItem(`player_id_${session.id}`, player.id);
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
				<h2 class="text-lg font-[800]">{$_('create_title_line1')} {$_('create_title_line2')}</h2>
				<Drawer.Close
					class="bg-surface-raised text-text-secondary hover:bg-border hidden h-8 w-8 items-center justify-center rounded-full text-xl leading-none transition-colors md:flex"
					>×</Drawer.Close
				>
			</div>
		</Drawer.Header>

		<div class="flex-1 space-y-6 overflow-y-auto px-6 pb-8">
			<!-- Game mode -->
			<div class="space-y-3">
				<RadioGroup bind:value={gameMode}>
					<ToggleGroupItem value="americano" class="flex-1">Americano</ToggleGroupItem>
					<ToggleGroupItem value="mexicano" class="flex-1">Mexicano</ToggleGroupItem>
				</RadioGroup>
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
				{creating ? $_('create_button_loading') : $_('create_button')}
			</Button>
		</div>
	</Drawer.Content>
</Drawer.Root>
