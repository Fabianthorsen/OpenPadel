<script lang="ts">
	import { goto } from '$app/navigation';
	import { api, ApiError } from '$lib/api/client';
	import { auth } from '$lib/auth.svelte';
	import { _ } from 'svelte-i18n';
	import { translateApiError } from '$lib/i18n/errors';
	import { PillToggleGroup, PillToggleItem } from '$lib/components/ui/pill-toggle-group';
	import * as Drawer from '$lib/components/ui/drawer';

	let { open = $bindable(false) }: { open?: boolean } = $props();

	let gameMode = $state<'americano' | 'mexicano'>('americano');
	let creating = $state(false);
	let error = $state('');

	async function create() {
		creating = true;
		error = '';
		try {
			const defaults =
				gameMode === 'mexicano'
					? { courts: 2, points: 24, rounds_total: 7 }
					: { courts: 2, points: 24 };
			const session = await api.sessions.create({
				game_mode: gameMode,
				name: '',
				...defaults
			});
			const adminToken = session.admin_token!;
			localStorage.setItem(`admin_token_${session.id}`, adminToken);
			const player = await api.players.join(
				session.id,
				auth.user!.display_name,
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
				<PillToggleGroup bind:value={gameMode}>
					<PillToggleItem value="americano">Americano</PillToggleItem>
					<PillToggleItem value="mexicano">Mexicano</PillToggleItem>
				</PillToggleGroup>
				<p class="text-text-secondary text-sm">
					{gameMode === 'mexicano' ? $_('create_mexicano_hint') : $_('create_americano_hint')}
				</p>
			</div>

			{#if error}
				<p class="text-destructive text-sm">{error}</p>
			{/if}

			<button
				onclick={create}
				disabled={creating}
				class="bg-primary hover:bg-primary-hover w-full rounded-2xl px-4 py-4 text-[15px] font-semibold text-white transition-colors disabled:opacity-60"
			>
				{creating ? $_('create_button_loading') : $_('create_button')}
			</button>
		</div>
	</Drawer.Content>
</Drawer.Root>
