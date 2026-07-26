<script lang="ts">
	import { goto } from '$app/navigation';
	import { api, ApiError } from '$lib/api/client';
	import { auth } from '$lib/auth.svelte';
	import { _ } from 'svelte-i18n';
	import { toast } from 'svelte-sonner';
	import { translateApiError } from '$lib/i18n/errors';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import { SectionLabel } from '$lib/components/ui/section-label';
	import * as Drawer from '$lib/components/ui/drawer';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';

	let { open = $bindable(false) }: { open?: boolean } = $props();

	let name = $state('');
	let description = $state('');
	let creating = $state(false);
	let showCancelConfirm = $state(false);

	async function create() {
		if (!name.trim()) {
			toast.error($_('create_club_name_required'));
			return;
		}

		creating = true;
		try {
			const club = await api.clubs.create(auth.token || '', {
				name: name.trim(),
				description: description.trim(),
				avatar_icon: 'Star',
				avatar_color: 'sky'
			});
			toast.success($_('create_club_created', { values: { name: club.name } }));
			open = false;
			name = '';
			description = '';
			goto(`/clubs/${club.id}`);
		} catch (e) {
			const msg = e instanceof ApiError ? translateApiError(e.message) : $_('create_club_error');
			toast.error(msg);
		} finally {
			creating = false;
		}
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter' && !creating && name.trim()) {
			event.preventDefault();
			create();
		}
	}

	function disableOtherInputs(focused: HTMLElement) {
		const container = focused.closest('[data-form-group]');
		if (!container) return;
		const inputs = container.querySelectorAll('input, textarea');
		inputs.forEach((input) => {
			if (input !== focused) {
				(input as HTMLInputElement | HTMLTextAreaElement).disabled = true;
			}
		});
	}

	function enableOtherInputs() {
		if (creating) return;
		const formGroup = document.querySelector('[data-form-group]');
		if (!formGroup) return;
		const inputs = formGroup.querySelectorAll('input, textarea');
		inputs.forEach((input) => {
			(input as HTMLInputElement | HTMLTextAreaElement).disabled = false;
		});
	}
</script>

<Drawer.Root bind:open>
	<Drawer.Content>
		<div class="pb-safe-page mx-auto w-full max-w-sm space-y-6 px-6 py-6">
			<div class="space-y-2">
				<Drawer.Title>{$_('create_club_title')}</Drawer.Title>
				<Drawer.Description>{$_('create_club_desc')}</Drawer.Description>
			</div>

			<div class="space-y-4" data-form-group>
				<!-- Club Name -->
				<div class="space-y-2.5">
					<SectionLabel>{$_('create_club_name_label')}</SectionLabel>
					<Input
						bind:value={name}
						placeholder={$_('create_club_name_placeholder')}
						onkeydown={handleKeydown}
						onfocus={(e) => disableOtherInputs(e.target as HTMLElement)}
						onblur={enableOtherInputs}
						disabled={creating}
						class="bg-surface-raised rounded-2xl border-0 px-4 py-3.5 text-sm"
					/>
				</div>

				<!-- Description -->
				<div class="space-y-2.5">
					<SectionLabel>{$_('create_club_description_label')}</SectionLabel>
					<Textarea
						bind:value={description}
						placeholder={$_('create_club_description_placeholder')}
						rows="3"
						onfocus={(e) => disableOtherInputs(e.target as HTMLElement)}
						onblur={enableOtherInputs}
						disabled={creating}
						class="bg-surface-raised w-full resize-none rounded-2xl border-0 px-4 py-3.5 text-sm"
					/>
				</div>
			</div>

			<div class="space-y-3 pt-4">
				<Button
					onclick={create}
					disabled={creating || !name.trim()}
					class="bg-primary hover:bg-primary-hover h-auto w-full rounded-2xl px-4 py-4 text-[15px] font-semibold text-white"
				>
					{creating ? $_('create_club_creating') : $_('create_club_submit')}
				</Button>
				<button
					onclick={() => (showCancelConfirm = true)}
					disabled={creating}
					class="text-text-secondary hover:text-text-primary w-full text-center text-sm transition-colors disabled:opacity-50"
				>
					{$_('create_club_cancel')}
				</button>
			</div>

			<ConfirmDialog
				open={showCancelConfirm}
				title={$_('create_club_discard_title')}
				description={$_('create_club_discard_desc')}
				confirmLabel={$_('create_club_discard_confirm')}
				cancelLabel={$_('create_club_discard_keep')}
				onconfirm={() => {
					open = false;
					name = '';
					description = '';
					showCancelConfirm = false;
				}}
				oncancel={() => (showCancelConfirm = false)}
			/>
		</div>
	</Drawer.Content>
</Drawer.Root>
