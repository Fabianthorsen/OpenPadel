<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import { auth } from '$lib/auth.svelte';
	import { Button } from '$lib/components/ui/button';
	import Spinner from '$lib/components/ui/spinner/spinner.svelte';
	import JoinCodeInput from '$lib/components/ui/join-code-input/join-code-input.svelte';
	import DividerOr from '$lib/components/DividerOr.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import { _ } from 'svelte-i18n';
	import { toast } from 'svelte-sonner';

	let rejoinSession = $state<App.Session | null>(null);
	let rejoinHref = $state('');

	async function loadRejoin() {
		const lastId = localStorage.getItem('last_session_id');
		if (!lastId) {
			rejoinSession = null;
			return;
		}
		try {
			const token = localStorage.getItem(`admin_token_${lastId}`) ?? undefined;
			const s = await api.sessions.get(lastId, token);
			if (s.status === 'lobby' || s.status === 'playing') {
				rejoinSession = s;
				rejoinHref = token ? `/s/${lastId}?token=${token}` : `/s/${lastId}`;
			} else {
				rejoinSession = null;
			}
		} catch {
			localStorage.removeItem('last_session_id');
			rejoinSession = null;
		}
	}

	onMount(async () => {
		if (page.url.searchParams.get('deleted') === '1') {
			toast($_('home_account_deleted'));
		}
		if (page.url.searchParams.get('notfound') === '1') {
			toast.error($_('home_session_not_found'));
		}

		await loadRejoin();
	});

	$effect(() => {
		if (auth.ready && auth.user) {
			const params = new URLSearchParams();
			if (page.url.searchParams.get('create') === '1') params.set('create', '1');
			if (page.url.searchParams.get('notfound') === '1') params.set('notfound', '1');
			const query = params.toString();
			goto(`/profile${query ? `?${query}` : ''}`);
		}
	});
</script>

<main class="pt-safe-page flex min-h-svh flex-col items-center px-6 pb-12">
	<div class="flex w-full max-w-sm flex-1 flex-col">
		<div class="flex flex-1 flex-col justify-center space-y-12">
			<!-- Brand -->
			<div class="space-y-1">
				<h1 class="text-primary text-[28px] font-[800]">OpenPadel</h1>
				<p class="text-text-secondary">{$_('home_tagline')}</p>
			</div>

			<!-- Loading state (auth checking) -->
			{#if !auth.ready}
				<div class="flex items-center justify-center">
					<Spinner label={$_('loading')} />
				</div>
			{:else}
				<!-- Actions -->
				<div class="space-y-4">
					{#if rejoinSession}
						<a
							href={rejoinHref}
							class="bg-surface-raised hover:bg-border flex items-center gap-3 rounded-2xl px-4 py-3.5 transition-colors"
						>
							<div
								class="bg-primary-muted flex h-9 w-9 shrink-0 items-center justify-center rounded-full"
							>
								<div class="bg-primary h-2.5 w-2.5 animate-pulse rounded-full"></div>
							</div>
							<div class="min-w-0 flex-1">
								<p class="text-text-disabled text-[11px] font-bold tracking-[0.1em] uppercase">
									{$_('home_rejoin_label')}
								</p>
								<p class="truncate text-sm font-semibold">{rejoinSession.name || 'OpenPadel'}</p>
							</div>
							<span class="text-text-secondary text-sm">→</span>
						</a>
					{/if}

					{#if !auth.user}
						<Button href="/auth" variant="default" size="cta">
							{$_('auth_sign_in')} →
						</Button>
						<p class="text-text-secondary text-center text-sm">
							{$_('auth_no_account')}
							<a
								href="/auth?register=1"
								class="text-primary hover:text-primary-hover font-semibold"
							>
								{$_('auth_switch_register')}
							</a>
						</p>

						<DividerOr label={$_('home_join_code_divider')} />

						<div class="space-y-3">
							<label for="join-code-input" class="sr-only">
								{$_('home_join_code_placeholder')}
							</label>
							<JoinCodeInput id="join-code-input" onComplete={(code) => goto(`/s/${code}`)} />
						</div>
					{/if}
				</div>
			{/if}
		</div>
	</div>
	<Footer />
</main>
