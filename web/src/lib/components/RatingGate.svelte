<script lang="ts">
	import { _ } from 'svelte-i18n';
	import { auth } from '$lib/auth.svelte';
	import { api } from '$lib/api/client';
	import { toast } from 'svelte-sonner';
	import { translateApiError } from '$lib/i18n/errors';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';
	import RatingPicker from '$lib/components/RatingPicker.svelte';

	// Blocking home backfill gate (#213). A legacy user whose account predates
	// ratings has a null self_rating; this makes them pick a level before they can
	// use the dashboard. `open` is a fixed `true` that is never rebound and the
	// close affordances are disabled, so the gate is genuinely non-dismissible —
	// but it only ever mounts on the home surface, so a deep link into a session
	// never renders it. Building on Dialog gives the focus trap and scroll lock a
	// blocking interstitial needs.
	let selected = $state<number | null>(null);
	let saving = $state(false);

	async function save() {
		if (selected === null || !auth.token) return;
		saving = true;
		try {
			const updated = await api.auth.updateSelfRating(auth.token, selected);
			auth.updateUser(updated);
		} catch (e) {
			const msg =
				e instanceof Error ? translateApiError(e.message) : translateApiError('server_error');
			toast.error(msg);
		} finally {
			saving = false;
		}
	}
</script>

<Dialog.Root open={true}>
	<Dialog.Content
		showCloseButton={false}
		escapeKeydownBehavior="ignore"
		interactOutsideBehavior="ignore"
		class="max-h-[85svh] space-y-6 overflow-y-auto"
	>
		<div class="space-y-2">
			<Dialog.Title class="text-2xl font-[800]">{$_('rating_gate_title')}</Dialog.Title>
			<Dialog.Description class="text-text-secondary text-sm">
				{$_('rating_gate_subtitle')}
			</Dialog.Description>
		</div>

		<RatingPicker bind:value={selected} name="self_rating" disabled={saving} />

		<Button onclick={save} disabled={selected === null || saving} size="cta">
			{saving ? $_('rating_gate_saving') : $_('rating_gate_save')}
		</Button>
	</Dialog.Content>
</Dialog.Root>
