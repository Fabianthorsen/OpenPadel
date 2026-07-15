<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { api } from '$lib/api/client';
	import { ApiError } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import { Label } from '$lib/components/ui/label';
	import PasswordInput from '$lib/components/ui/password-input/password-input.svelte';
	import AuthShell from '$lib/components/AuthShell.svelte';
	import { _ } from 'svelte-i18n';
	import { toast } from 'svelte-sonner';
	import { translateApiError } from '$lib/i18n/errors';

	const token = page.url.searchParams.get('token') ?? '';

	let password = $state('');
	let loading = $state(false);
	let error = $state('');

	async function submit() {
		error = '';
		loading = true;
		try {
			await api.auth.resetPassword(token, password);
			goto('/auth?reset=1');
		} catch (e) {
			error =
				e instanceof ApiError ? translateApiError(e.message) : translateApiError('server_error');
		} finally {
			loading = false;
		}
	}
</script>

<AuthShell subtitle={$_('reset_subtitle')}>
	{#snippet children()}
		{#if !token}
			<p class="text-destructive text-sm">{$_('reset_invalid_link')}</p>
		{:else}
			<form
				onsubmit={(e) => {
					e.preventDefault();
					submit();
				}}
				class="space-y-4"
			>
				<div class="space-y-2">
					<Label
						for="password"
						class="text-text-secondary text-[11px] font-semibold tracking-[0.1em] uppercase"
					>
						{$_('reset_new_password_label')}
					</Label>
					<PasswordInput
						id="password"
						bind:value={password}
						placeholder={$_('auth_password_placeholder')}
						autocomplete="new-password"
						ariaInvalid={!!error}
						class="bg-surface-raised rounded-2xl px-4 py-3.5 text-sm"
					/>
				</div>

				{#if error}
					<p class="text-destructive text-sm" role="alert">{error}</p>
				{/if}

				<Button type="submit" disabled={loading || password.length < 8} size="cta">
					{loading ? $_('reset_button_loading') : $_('reset_button')}
				</Button>
			</form>
		{/if}
	{/snippet}
</AuthShell>
