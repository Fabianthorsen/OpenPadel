<script lang="ts">
	import { api, ApiError } from '$lib/api/client';
	import { auth } from '$lib/auth.svelte';
	import { _ } from 'svelte-i18n';
	import { toast } from 'svelte-sonner';
	import { translateApiError } from '$lib/i18n/errors';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import { SectionLabel } from '$lib/components/ui/section-label';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import * as Drawer from '$lib/components/ui/drawer';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import { ShieldMinus, ShieldPlus, UserMinus, X } from '@lucide/svelte';

	let {
		open = $bindable(false),
		club,
		onchanged,
		ondeleted
	}: {
		open?: boolean;
		club: App.ClubDetail;
		/** Called after a successful edit or roster change so the parent can reload. */
		onchanged: () => void | Promise<void>;
		/** Called after the club is deleted. */
		ondeleted: () => void;
	} = $props();

	// Seeded from `club` by the $effect below (kept in sync when a new club opens).
	let name = $state('');
	let description = $state('');
	let saving = $state(false);
	// Guards concurrent roster mutations; also disables the row being acted on.
	let busyUserId = $state<string | null>(null);
	let showDeleteConfirm = $state(false);

	const members = $derived(club.members);
	const myId = $derived(auth.user?.id ?? '');

	// Re-seed the edit fields whenever a different club is opened.
	$effect(() => {
		name = club.club.name;
		description = club.club.description;
	});

	function errMsg(e: unknown): string {
		return e instanceof ApiError ? translateApiError(e.message) : translateApiError('server_error');
	}

	async function saveEdits() {
		if (!auth.token || !name.trim()) {
			toast.error($_('create_club_name_required'));
			return;
		}
		saving = true;
		try {
			await api.clubs.update(auth.token, club.club.id, {
				name: name.trim(),
				description: description.trim(),
				avatar_icon: club.club.avatar_icon,
				avatar_color: club.club.avatar_color
			});
			toast.success($_('club_admin_updated'));
			await onchanged();
		} catch (e) {
			toast.error(errMsg(e));
		} finally {
			saving = false;
		}
	}

	async function setRole(userId: string, role: 'admin' | 'member') {
		if (!auth.token || busyUserId) return;
		busyUserId = userId;
		try {
			await api.clubs.setMemberRole(auth.token, club.club.id, userId, role);
			await onchanged();
		} catch (e) {
			toast.error(errMsg(e));
		} finally {
			busyUserId = null;
		}
	}

	async function removeMember(userId: string) {
		if (!auth.token || busyUserId) return;
		busyUserId = userId;
		try {
			await api.clubs.removeMember(auth.token, club.club.id, userId);
			toast.success($_('club_admin_member_removed'));
			await onchanged();
		} catch (e) {
			toast.error(errMsg(e));
		} finally {
			busyUserId = null;
		}
	}

	async function deleteClub() {
		if (!auth.token) return;
		showDeleteConfirm = false;
		try {
			await api.clubs.remove(auth.token, club.club.id);
			toast.success($_('club_admin_deleted'));
			open = false;
			ondeleted();
		} catch (e) {
			toast.error(errMsg(e));
		}
	}
</script>

<Drawer.Root bind:open>
	<Drawer.Content size="lg">
		<!-- min-h-0 lets this flex child shrink so its own scroll engages instead of
		     the touch scroll leaking to the page behind the drawer. -->
		<div
			class="pb-safe-page mx-auto min-h-0 w-full max-w-sm flex-1 space-y-8 overflow-y-auto px-6 py-6"
		>
			<div class="flex items-start justify-between gap-3">
				<div class="space-y-2">
					<Drawer.Title>{$_('club_manage')}</Drawer.Title>
					<Drawer.Description>{$_('club_admin_desc')}</Drawer.Description>
				</div>
				<Drawer.Close
					class="text-text-secondary hover:text-text-primary -mt-1 shrink-0"
					aria-label={$_('club_admin_close')}
				>
					<X size={22} />
				</Drawer.Close>
			</div>

			<!-- Edit details -->
			<div class="space-y-4">
				<div class="space-y-2.5">
					<SectionLabel>{$_('create_club_name_label')}</SectionLabel>
					<Input
						bind:value={name}
						placeholder={$_('create_club_name_placeholder')}
						disabled={saving}
						class="bg-surface-raised rounded-2xl border-0 px-4 py-3.5 text-sm"
					/>
				</div>
				<div class="space-y-2.5">
					<SectionLabel>{$_('create_club_description_label')}</SectionLabel>
					<Textarea
						bind:value={description}
						placeholder={$_('create_club_description_placeholder')}
						rows="3"
						disabled={saving}
						class="bg-surface-raised w-full resize-none rounded-2xl border-0 px-4 py-3.5 text-sm"
					/>
				</div>
				<Button
					onclick={saveEdits}
					disabled={saving || !name.trim()}
					class="bg-primary hover:bg-primary-hover h-auto w-full rounded-2xl px-4 py-4 text-[15px] font-semibold text-white"
				>
					{saving ? $_('club_admin_saving') : $_('club_admin_save')}
				</Button>
			</div>

			<!-- Manage roster -->
			<div class="space-y-3">
				<SectionLabel>{$_('club_admin_roster', { values: { count: members.length } })}</SectionLabel
				>
				<div class="space-y-2">
					{#each members as member (member.user_id)}
						<div class="bg-surface-raised flex items-center gap-3 rounded-2xl px-4 py-3">
							<Avatar color={member.avatar_color} name={member.display_name} size="sm" />
							<div class="min-w-0 flex-1">
								<p class="truncate text-sm font-semibold">
									{member.display_name}{#if member.user_id === myId}<span
											class="text-text-disabled font-normal"
										>
											{$_('club_admin_you')}</span
										>{/if}
								</p>
								<p class="text-text-secondary text-xs">{$_(`club_role_${member.role}`)}</p>
							</div>

							{#if member.role === 'admin'}
								<button
									onclick={() => setRole(member.user_id, 'member')}
									disabled={busyUserId !== null}
									class="text-text-secondary hover:text-text-primary shrink-0 p-1.5 disabled:opacity-40"
									aria-label={$_('club_admin_demote', { values: { name: member.display_name } })}
									title={$_('club_admin_demote_title')}
								>
									<ShieldMinus size={17} />
								</button>
							{:else}
								<button
									onclick={() => setRole(member.user_id, 'admin')}
									disabled={busyUserId !== null}
									class="text-text-secondary hover:text-primary shrink-0 p-1.5 disabled:opacity-40"
									aria-label={$_('club_admin_promote', { values: { name: member.display_name } })}
									title={$_('club_admin_promote_title')}
								>
									<ShieldPlus size={17} />
								</button>
							{/if}
							<button
								onclick={() => removeMember(member.user_id)}
								disabled={busyUserId !== null}
								class="text-text-secondary hover:text-destructive shrink-0 p-1.5 disabled:opacity-40"
								aria-label={$_('club_admin_remove', { values: { name: member.display_name } })}
								title={$_('club_admin_remove_title')}
							>
								<UserMinus size={17} />
							</button>
						</div>
					{/each}
				</div>
			</div>

			<!-- Danger zone -->
			<div class="border-border space-y-3 border-t pt-6">
				<button
					onclick={() => (showDeleteConfirm = true)}
					class="text-destructive hover:bg-destructive/10 w-full rounded-2xl px-4 py-3.5 text-sm font-semibold transition-colors"
				>
					{$_('club_admin_delete')}
				</button>
				<p class="text-text-disabled text-center text-xs">{$_('club_admin_delete_hint')}</p>
			</div>

			<Drawer.Close
				class="text-text-secondary hover:text-text-primary w-full text-center text-sm transition-colors"
			>
				{$_('club_admin_done')}
			</Drawer.Close>

			<ConfirmDialog
				open={showDeleteConfirm}
				title={$_('club_admin_delete_confirm_title')}
				description={$_('club_admin_delete_confirm_desc')}
				confirmLabel={$_('club_admin_delete')}
				cancelLabel={$_('club_admin_delete_keep')}
				destructive
				onconfirm={deleteClub}
				oncancel={() => (showDeleteConfirm = false)}
			/>
		</div>
	</Drawer.Content>
</Drawer.Root>
