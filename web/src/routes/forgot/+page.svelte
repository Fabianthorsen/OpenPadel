<script lang="ts">
	import { api } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import AuthShell from '$lib/components/AuthShell.svelte';
	import { _ } from 'svelte-i18n';

	let email = $state('');
	let loading = $state(false);
	let sent = $state(false);

	async function submit() {
		loading = true;
		try {
			await api.auth.forgotPassword(email.trim());
		} catch {
			// Always show success — never reveal whether the email exists
		} finally {
			loading = false;
			sent = true;
		}
	}
</script>

<AuthShell subtitle={$_('forgot_subtitle')} backHref="/auth">
	{#snippet children()}
		{#if sent}
			<div class="bg-surface-raised space-y-1 rounded-2xl px-5 py-5">
				<p class="font-semibold">{$_('forgot_sent_title')}</p>
				<p class="text-text-secondary text-sm">{$_('forgot_sent_desc')}</p>
			</div>
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
						class="bg-surface-raised rounded-2xl px-4 py-3.5 text-sm"
					/>
				</div>

				<Button type="submit" disabled={loading || !email.trim()} size="cta">
					{loading ? $_('forgot_button_loading') : $_('forgot_button')}
				</Button>
			</form>
		{/if}
	{/snippet}
</AuthShell>
