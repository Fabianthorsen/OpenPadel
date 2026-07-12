<script lang="ts">
	import { goto } from '$app/navigation';
	import { api, ApiError } from '$lib/api/client';
	import { auth } from '$lib/auth.svelte';
	import { toast } from 'svelte-sonner';
	import { translateApiError } from '$lib/i18n/errors';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { SectionLabel } from '$lib/components/ui/section-label';
	import * as Drawer from '$lib/components/ui/drawer';

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
		<div class="mx-auto w-full max-w-sm space-y-6 pb-safe-page px-6 py-6">
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
					<textarea
						bind:value={description}
						placeholder="Optional description"
						class="bg-surface-raised w-full rounded-2xl border-0 px-4 py-3.5 text-sm placeholder:text-text-disabled focus:outline-none focus:ring-1 focus:ring-primary"
						rows="3"
						disabled={creating}
					></textarea>
				</div>
			</div>

			<div class="space-y-3 pt-4">
				<Button onclick={create} disabled={creating || !name.trim()} class="bg-primary hover:bg-primary-hover h-auto w-full rounded-2xl px-4 py-4 text-[15px] font-semibold text-white">
					{creating ? 'Creating...' : 'Create Club'}
				</Button>
				<button
					onclick={() => (showCancelConfirm = true)}
					disabled={creating}
					class="w-full text-center text-sm text-text-secondary hover:text-text-primary transition-colors disabled:opacity-50"
				>
					Cancel
				</button>
			</div>

			{#if showCancelConfirm}
				<div class="fixed inset-0 bg-black/50 flex items-end z-50">
					<div class="w-full bg-surface-raised rounded-t-2xl p-6 space-y-4">
						<div class="space-y-2">
							<h2 class="font-semibold text-base">Discard changes?</h2>
							<p class="text-sm text-text-secondary">You'll lose your club details.</p>
						</div>
						<div class="flex gap-3">
							<button
								onclick={() => (showCancelConfirm = false)}
								class="flex-1 h-auto rounded-2xl px-4 py-3 text-sm font-semibold text-text-primary bg-surface-raised hover:bg-border transition-colors"
							>
								Keep editing
							</button>
							<Drawer.Close class="flex-1">
								<button class="w-full h-auto rounded-2xl px-4 py-3 text-sm font-semibold text-white bg-primary hover:bg-primary-hover transition-colors">
									Discard
								</button>
							</Drawer.Close>
						</div>
					</div>
				</div>
			{/if}
		</div>
	</Drawer.Content>
</Drawer.Root>
