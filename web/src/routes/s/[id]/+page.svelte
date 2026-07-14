<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { onMount, onDestroy } from 'svelte';
	import { api, ApiError } from '$lib/api/client';
	import Lobby from '$lib/components/Lobby.svelte';
	import ActiveSession from '$lib/components/ActiveSession.svelte';
	import Leaderboard from '$lib/components/Leaderboard.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import Spinner from '$lib/components/ui/spinner/spinner.svelte';
	import * as Drawer from '$lib/components/ui/drawer';
	import { toast } from 'svelte-sonner';
	import { _ } from 'svelte-i18n';
	import { translateApiError } from '$lib/i18n/errors';
	import { sessionDialog } from '$lib/stores/sessionDialog';
	import { numpad as numpadStore } from '$lib/stores/numpad';
	import { sessionStream } from '$lib/stores/sessionStream.svelte';

	let session = $state<App.Session | null>(null);
	let currentRound = $state<App.Round | null>(null);
	let fallbackInterval: ReturnType<typeof setInterval>;

	const FALLBACK_POLL = 30_000;

	const sessionId = $derived(page.params.id as string);

	function getAdminToken(): string | null {
		if (typeof localStorage === 'undefined') return null;
		const stored = localStorage.getItem(`admin_token_${sessionId}`);
		if (stored) return stored;
		const param = new URL(location.href).searchParams.get('token');
		if (param) {
			localStorage.setItem(`admin_token_${sessionId}`, param);
			window.history.replaceState({}, '', location.pathname);
		}
		return param;
	}

	const isAdmin = $derived(!!getAdminToken() || !!session?.is_creator);

	async function load() {
		const token = getAdminToken() ?? undefined;
		try {
			session = await api.sessions.get(sessionId, token);
			// Token recovery: if the server recognises us as creator, persist the token.
			if (session.admin_token && !token) {
				localStorage.setItem(`admin_token_${sessionId}`, session.admin_token);
			}
			if (session.status !== 'lobby') {
				currentRound = await api.rounds.current(sessionId).catch(() => null);
			}
		} catch (e) {
			if (e instanceof ApiError && e.status === 404) {
				goto('/?notfound=1');
				return;
			}
			toast.error(
				e instanceof ApiError ? translateApiError(e.message) : translateApiError('server_error')
			);
		}
	}

	const stream = sessionStream(page.params.id as string);

	onMount(() => {
		load();
		stream.start();
		const cleanupSession = stream.onEvent('session_updated', () => {
			load();
		});
		const cleanupRound = stream.onEvent('round_updated', () => {
			load();
		});
		const cleanupLive = stream.onEvent<{ match_id: string; a: number; b: number; server: string }>(
			'live_score',
			(p) => {
				if (!currentRound) return;
				currentRound = {
					...currentRound,
					matches: currentRound.matches.map((m) =>
						m.id === p.match_id
							? { ...m, live: { a: p.a, b: p.b, server: p.server as 'a' | 'b' } }
							: m
					)
				};
			}
		);
		fallbackInterval = setInterval(load, FALLBACK_POLL);
		return () => {
			cleanupSession();
			cleanupRound();
			cleanupLive();
		};
	});

	onDestroy(() => {
		clearInterval(fallbackInterval);
		stream.close();
		numpadStore.close();
		sessionDialog.close();
	});

	async function handleDialogConfirm() {
		const callback = $sessionDialog.onConfirm;
		sessionDialog.close();
		if (callback) {
			await callback();
		}
	}

	function handleDialogCancel() {
		sessionDialog.close();
	}
</script>

<div class="flex h-screen w-full flex-col overflow-y-auto">
	{#if !session}
		<main class="flex flex-1 flex-col items-center justify-center gap-3 px-4">
			<Spinner />
			<p class="text-text-secondary text-sm">{$_('loading')}</p>
		</main>
	{:else if session.status === 'lobby'}
		<Lobby {session} {isAdmin} onRefresh={load} onStarted={load} {stream} />
	{:else if session.status === 'playing' && currentRound}
		<ActiveSession {session} {currentRound} {isAdmin} onRefresh={load} {stream} />
	{:else if session.status === 'done'}
		<Leaderboard sessionId={session.id} sessionName={session.name} complete />
	{:else}
		<main class="flex flex-1 flex-col items-center justify-center gap-3 px-4">
			<Spinner />
			<p class="text-text-secondary text-sm">Loading…</p>
		</main>
	{/if}
</div>

{#if $sessionDialog.isOpen}
	<div class="fixed inset-0 z-50">
		<ConfirmDialog
			open={$sessionDialog.isOpen}
			title={$sessionDialog.type === 'close' ? $_('close_dialog_title') : $_('cancel_dialog_title')}
			description={$sessionDialog.type === 'close'
				? $_('close_dialog_desc')
				: $_('cancel_dialog_desc')}
			confirmLabel={$sessionDialog.type === 'close'
				? $_('close_dialog_confirm')
				: $_('cancel_dialog_confirm')}
			cancelLabel={$sessionDialog.type === 'close'
				? $_('close_dialog_cancel')
				: $_('cancel_dialog_cancel')}
			destructive
			onconfirm={handleDialogConfirm}
			oncancel={handleDialogCancel}
		/>
	</div>
{/if}

<!-- Numpad drawer -->
<Drawer.Root open={!!$numpadStore} onOpenChange={(open) => !open && $numpadStore?.onClose()}>
	<Drawer.Content class="mx-auto flex max-h-[80vh] w-full max-w-[480px] flex-col gap-3">
		<div class="px-6 pt-6">
			<p
				class="text-text-disabled mb-3 text-center text-[10px] font-bold tracking-widest uppercase"
			>
				Target: {$numpadStore?.targetPoints}
			</p>
			<p
				class="mb-6 text-center text-[64px] leading-none font-[800] tabular-nums transition-transform
        {$numpadStore?.shaking ? 'animate-[shake_0.4s_ease-in-out]' : ''}"
			>
				{$numpadStore?.value || '0'}
			</p>
		</div>
		<div class="flex-1 px-6 pb-[env(safe-area-inset-bottom)]">
			<div class="mx-auto grid max-w-sm grid-cols-3 gap-3">
				{#each ['1', '2', '3', '4', '5', '6', '7', '8', '9'] as d}
					<button
						onclick={() => $numpadStore?.onDigit(d)}
						class="bg-surface-raised rounded-2xl py-4 text-xl font-[800] transition-all select-none active:scale-95"
						>{d}</button
					>
				{/each}
				<button
					onclick={() => $numpadStore?.onDelete()}
					class="bg-surface-raised rounded-2xl py-4 text-xl font-[800] transition-all select-none active:scale-95"
					>⌫</button
				>
				<button
					onclick={() => $numpadStore?.onDigit('0')}
					class="bg-surface-raised rounded-2xl py-4 text-xl font-[800] transition-all select-none active:scale-95"
					>0</button
				>
				<button
					onclick={() => $numpadStore?.onConfirm()}
					class="bg-primary rounded-2xl py-4 text-xl font-[800] text-white transition-all select-none active:scale-95"
					>✓</button
				>
			</div>
		</div>
	</Drawer.Content>
</Drawer.Root>
