<script lang="ts">
	import { goto } from '$app/navigation';
	import { api, ApiError } from '$lib/api/client';
	import { auth } from '$lib/auth.svelte';
	import { toast } from 'svelte-sonner';
	import { translateApiError } from '$lib/i18n/errors';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import * as Drawer from '$lib/components/ui/drawer';

	let { open = $bindable(false) }: { open?: boolean } = $props();

	let name = $state('');
	let description = $state('');
	let avatarIcon = $state('Star');
	let avatarColor = $state('forest');
	let creating = $state(false);

	const avatarIcons = ['Zap', 'Star', 'Flame', 'Shield', 'Crown', 'Trophy', 'Target', 'Rocket'];
	const avatarColors = [{ label: 'Forest', value: 'forest' }];

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
				avatar_icon: avatarIcon,
				avatar_color: avatarColor
			});
			toast.success(`Club "${club.name}" created!`);
			open = false;
			name = '';
			description = '';
			avatarIcon = 'Star';
			avatarColor = 'forest';
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
		<div class="mx-auto w-full max-w-sm space-y-6">
			<div class="space-y-2">
				<Drawer.Title>Create a Club</Drawer.Title>
				<Drawer.Description>
					Start a new club to organize games and build your community.
				</Drawer.Description>
			</div>

			<div class="space-y-4">
				<!-- Club Name -->
				<div class="space-y-2">
					<label for="club-name" class="text-sm font-medium">Club Name</label>
					<Input
						id="club-name"
						bind:value={name}
						placeholder="e.g., Bouvet Padel"
						onkeydown={handleKeydown}
						disabled={creating}
					/>
				</div>

				<!-- Description -->
				<div class="space-y-2">
					<label for="club-desc" class="text-sm font-medium">Description</label>
					<textarea
						id="club-desc"
						bind:value={description}
						placeholder="Optional description of your club"
						class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
						rows="3"
						disabled={creating}
					></textarea>
				</div>

				<!-- Avatar Icon -->
				<div class="space-y-2">
					<label for="avatar-icon" class="text-sm font-medium">Icon</label>
					<div class="grid grid-cols-4 gap-2">
						{#each avatarIcons as icon}
							<button
								onclick={() => (avatarIcon = icon)}
								class="p-2 rounded-lg border-2 transition-colors"
								class:border-primary={avatarIcon === icon}
								class:border-input={avatarIcon !== icon}
								disabled={creating}
							>
								<span class="text-sm">{icon}</span>
							</button>
						{/each}
					</div>
				</div>

				<!-- Avatar Color -->
				<div class="space-y-2">
					<label for="avatar-color" class="text-sm font-medium">Color</label>
					<div class="flex gap-2">
						{#each avatarColors as color}
							<button
								onclick={() => (avatarColor = color.value)}
								class="px-3 py-2 rounded-lg border-2 transition-colors text-sm"
								class:border-primary={avatarColor === color.value}
								class:border-input={avatarColor !== color.value}
								disabled={creating}
							>
								{color.label}
							</button>
						{/each}
					</div>
				</div>
			</div>

			<div class="flex gap-3 pt-4">
				<Drawer.Close class="flex-1">
					<Button variant="outline" disabled={creating} class="w-full">
						Cancel
					</Button>
				</Drawer.Close>
				<Button onclick={create} disabled={creating || !name.trim()} class="flex-1">
					{creating ? 'Creating...' : 'Create Club'}
				</Button>
			</div>
		</div>
	</Drawer.Content>
</Drawer.Root>
