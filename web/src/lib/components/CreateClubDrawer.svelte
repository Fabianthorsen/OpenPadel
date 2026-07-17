<script lang="ts">
	import { goto } from '$app/navigation';
	import { api, ApiError } from '$lib/api/client';
	import { auth } from '$lib/auth.svelte';
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
			toast.error('Club name is required');
			return;
		}

		creating = true;
		try {
			const club = await api.clubs.create(auth.token || '', {
				name: name.trim(),
				description: description.trim(),
				avatar_icon: 'Star',
				avatar_color: 'forest'
			});
			toast.success(`Club "${club.name}" created!`);
			open = false;
			name = '';
			description = '';
			goto(`/clubs/${club.id}`);
		} catch (e) {
			const msg = e instanceof ApiError ? translateApiError(e.message) : 'Failed to create club';
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
</script>

<Drawer.Root bind:open>
	<Drawer.Content>
		<div class="pb-safe-page mx-auto w-full max-w-sm space-y-6 px-6 py-6">
			<div class="space-y-2">
				<Drawer.Title>Create a Club</Drawer.Title>
				<Drawer.Description>
					Start a new club to organize games and build your community.
				</Drawer.Description>
			</div>

			<div class="space-y-4">
				<!-- Club Name -->
				<div class="space-y-2.5">
					<SectionLabel>Club Name</SectionLabel>
					<Input
						bind:value={name}
						placeholder="e.g., Bouvet Padel"
						onkeydown={handleKeydown}
						disabled={creating}
						class="bg-surface-raised rounded-2xl border-0 px-4 py-3.5 text-sm"
					/>
				</div>

				<!-- Description -->
				<div class="space-y-2.5">
					<SectionLabel>Description</SectionLabel>
					<Textarea
						bind:value={description}
						placeholder="Optional description"
						rows="3"
						disabled={creating}
						class="bg-surface-raised"
					/>
				</div>
			</div>

			<div class="space-y-3 pt-4">
				<Button
					onclick={create}
					disabled={creating || !name.trim()}
					class="bg-primary hover:bg-primary-hover h-auto w-full rounded-2xl px-4 py-4 text-[15px] font-semibold text-white"
				>
					{creating ? 'Creating...' : 'Create Club'}
				</Button>
				<button
					onclick={() => (showCancelConfirm = true)}
					disabled={creating}
					class="text-text-secondary hover:text-text-primary w-full text-center text-sm transition-colors disabled:opacity-50"
				>
					Cancel
				</button>
			</div>

			<ConfirmDialog
				open={showCancelConfirm}
				title="Discard changes?"
				description="You'll lose your club details."
				confirmLabel="Discard"
				cancelLabel="Keep editing"
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
