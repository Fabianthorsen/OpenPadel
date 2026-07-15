<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { auth } from '$lib/auth.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import PasswordInput from '$lib/components/ui/password-input/password-input.svelte';
	import AuthShell from '$lib/components/AuthShell.svelte';
	import { _ } from 'svelte-i18n';
	import { toast } from 'svelte-sonner';
	import { ApiError } from '$lib/api/client';
	import { translateApiError } from '$lib/i18n/errors';

	const redirect = page.url.searchParams.get('redirect') ?? '';

	$effect(() => {
		if (auth.ready && auth.user) goto(redirect || '/profile');
	});

	const resetSuccess = page.url.searchParams.get('reset') === '1';

	let mode = $state<'login' | 'register'>(
		page.url.searchParams.get('register') === '1' ? 'register' : 'login'
	);
	let email = $state('');
	let password = $state('');
	let firstName = $state('');
	let lastName = $state('');
	let loading = $state(false);
	let error = $state('');

	// Show reset success toast once translations load
	let resetToastShown = false;
	$effect(() => {
		if (resetSuccess && !resetToastShown && $_('reset_success_banner') !== 'reset_success_banner') {
			resetToastShown = true;
			toast.success($_('reset_success_banner'));
		}
	});

	async function submit() {
		error = '';
		loading = true;
		try {
			if (mode === 'login') {
				await auth.login(email, password);
			} else {
				await auth.register(email, `${firstName.trim()} ${lastName.trim()}`.trim(), password);
			}
			goto(redirect || '/');
		} catch (e) {
			error =
				e instanceof ApiError ? translateApiError(e.message) : translateApiError('server_error');
		} finally {
			loading = false;
		}
	}
</script>

<AuthShell subtitle={mode === 'login' ? $_('auth_login_subtitle') : $_('auth_register_subtitle')}>
	{#snippet children()}
		<form
			onsubmit={(e) => {
				e.preventDefault();
				submit();
			}}
			class="space-y-4"
		>
			{#if mode === 'register'}
				<div class="flex gap-3">
					<div class="flex-1 space-y-2">
						<Label
							for="firstName"
							class="text-text-secondary text-[11px] font-semibold tracking-[0.1em] uppercase"
						>
							{$_('auth_first_name_label')}
						</Label>
						<Input
							id="firstName"
							bind:value={firstName}
							placeholder={$_('auth_first_name_placeholder')}
							maxlength={32}
							autocomplete="given-name"
							class="bg-surface-raised rounded-2xl px-4 py-3.5 text-sm"
						/>
					</div>
					<div class="flex-1 space-y-2">
						<Label
							for="lastName"
							class="text-text-secondary text-[11px] font-semibold tracking-[0.1em] uppercase"
						>
							{$_('auth_last_name_label')}
						</Label>
						<Input
							id="lastName"
							bind:value={lastName}
							placeholder={$_('auth_last_name_placeholder')}
							maxlength={32}
							autocomplete="family-name"
							class="bg-surface-raised rounded-2xl px-4 py-3.5 text-sm"
						/>
					</div>
				</div>
			{/if}

			<div class="space-y-2">
				<Label
					for="email"
					class="text-text-secondary text-[11px] font-semibold tracking-[0.1em] uppercase"
				>
					{$_('auth_email_label')}
				</Label>
				<Input
					id="email"
					bind:value={email}
					type="email"
					placeholder={$_('auth_email_placeholder')}
					autocomplete="email"
					ariaInvalid={!!error}
					class="bg-surface-raised rounded-2xl px-4 py-3.5 text-sm"
				/>
			</div>

			<div class="space-y-2">
				<Label
					for="password"
					class="text-text-secondary text-[11px] font-semibold tracking-[0.1em] uppercase"
				>
					{$_('auth_password_label')}
				</Label>
				<PasswordInput
					id="password"
					bind:value={password}
					placeholder={$_('auth_password_placeholder')}
					autocomplete={mode === 'login' ? 'current-password' : 'new-password'}
					ariaInvalid={!!error}
					class="bg-surface-raised rounded-2xl px-4 py-3.5 text-sm"
				/>
			</div>

			{#if error}
				<p class="text-destructive text-sm" role="alert">{error}</p>
			{/if}

			<Button type="submit" disabled={loading} size="cta">
				{loading ? '…' : mode === 'login' ? $_('auth_login_button') : $_('auth_register_button')}
			</Button>

			{#if mode === 'login'}
				<div class="flex justify-center">
					<a href="/forgot" class="text-text-disabled hover:text-text-secondary text-xs">
						{$_('auth_forgot_password')}
					</a>
				</div>
			{/if}
		</form>

		<!-- Mode toggle -->
		<div class="flex justify-center gap-1 text-center text-xs">
			<span class="text-text-secondary">
				{mode === 'login' ? $_('auth_no_account') : $_('auth_have_account')}
			</span>
			<button
				type="button"
				onclick={() => {
					mode = mode === 'login' ? 'register' : 'login';
					error = '';
				}}
				class="text-primary hover:text-primary-hover font-semibold"
			>
				{mode === 'login' ? $_('auth_switch_register') : $_('auth_switch_login')}
			</button>
		</div>
	{/snippet}
</AuthShell>
