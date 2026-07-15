<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import { auth } from '$lib/auth.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import Footer from '$lib/components/Footer.svelte';
	import { _ } from 'svelte-i18n';
	import { initials } from '$lib/utils';
	import { toast } from 'svelte-sonner';

	let joinCode = $state('');
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
			const notfound = page.url.searchParams.get('notfound');
			goto(notfound ? '/profile?notfound=1' : '/profile');
		}
	});

	function joinByCode() {
		const code = joinCode.trim().toUpperCase();
		if (code) goto(`/s/${code}`);
	}
</script>

<main class="pt-safe-page flex min-h-svh flex-col items-center px-6 pb-12">
	<div class="flex w-full max-w-sm flex-1 flex-col">
		<div class="flex flex-1 flex-col justify-center space-y-12">
			<!-- Brand -->
			<div class="space-y-1">
				<h1 class="text-primary text-[28px] font-[800]">OpenPadel</h1>
				<p class="text-text-secondary">{$_('home_tagline')}</p>
			</div>

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

				{#if auth.ready && !auth.user}
					<a
						href="/auth"
						class="bg-primary hover:bg-primary-hover flex h-auto w-full items-center justify-center rounded-2xl px-4 py-4 text-[15px] font-semibold text-white"
					>
						{$_('auth_sign_in')} →
					</a>
					<p class="text-text-secondary text-center text-sm">
						{$_('auth_no_account')}
						<a href="/auth?register=1" class="text-primary hover:text-primary-hover font-semibold">
							{$_('auth_switch_register')}
						</a>
					</p>
				{/if}

				<div class="flex items-center gap-3">
					<div class="bg-border h-px flex-1"></div>
					<span class="text-text-disabled text-xs">{$_('home_join_code_divider')}</span>
					<div class="bg-border h-px flex-1"></div>
				</div>

				<form
					onsubmit={(e) => {
						e.preventDefault();
						joinByCode();
					}}
					class="flex gap-2"
				>
					<Input
						bind:value={joinCode}
						oninput={(e: Event) => {
							joinCode = (e.currentTarget as HTMLInputElement).value.toUpperCase();
						}}
						placeholder={$_('home_join_code_placeholder')}
						maxlength={4}
						autocomplete="off"
						autocorrect="off"
						autocapitalize="characters"
						spellcheck={false}
						class="bg-surface-raised min-w-0 flex-1 rounded-2xl border-0 px-4 py-3.5 text-sm"
					/>
					<Button
						type="submit"
						disabled={!joinCode.trim()}
						variant="secondary"
						class="bg-surface-raised text-text-primary hover:bg-border h-auto rounded-2xl px-5 text-sm font-semibold"
					>
						{$_('home_join_button')}
					</Button>
				</form>
			</div>
		</div>

		<!-- Auth (logged-in pill — guests see the sign-in button above instead) -->
		{#if auth.user}
			<div class="flex justify-center pt-6">
				<a
					href="/profile"
					class="hover:bg-surface-raised flex items-center gap-3 rounded-2xl px-3 py-2 transition-colors"
				>
					<div
						class="bg-primary-muted text-primary flex h-7 w-7 items-center justify-center rounded-full text-xs font-[800]"
					>
						{initials(auth.user.display_name)}
					</div>
					<span class="text-sm font-semibold">{auth.user.display_name}</span>
					<span class="text-text-disabled text-xs">→</span>
				</a>
			</div>
		{/if}
	</div>
	<Footer />
</main>
