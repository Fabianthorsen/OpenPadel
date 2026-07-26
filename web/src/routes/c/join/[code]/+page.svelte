<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { api, ApiError } from '$lib/api/client';
	import { auth } from '$lib/auth.svelte';
	import { _ } from 'svelte-i18n';
	import { Button } from '$lib/components/ui/button';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import { Spinner } from '$lib/components/ui/spinner';
	import Footer from '$lib/components/Footer.svelte';
	import { toast } from 'svelte-sonner';

	const code = $derived(page.params.code as string);
	const authHref = $derived(`/auth?redirect=/c/join/${code}`);
	const registerHref = $derived(`/auth?register=1&redirect=/c/join/${code}`);

	let preview = $state<App.ClubJoinPreview | null>(null);
	let loading = $state(true);
	let invalid = $state(false);
	let joining = $state(false);

	async function loadPreview() {
		try {
			loading = true;
			invalid = false;
			preview = await api.clubs.joinPreview(code);
		} catch (err) {
			if (err instanceof ApiError && err.status === 404) {
				invalid = true;
			} else {
				toast.error($_('club_join_load_error'));
				invalid = true;
			}
		} finally {
			loading = false;
		}
	}

	async function join() {
		if (!auth.token) {
			goto(authHref);
			return;
		}
		try {
			joining = true;
			const { id } = await api.clubs.join(auth.token, code);
			toast.success($_('club_join_welcome'));
			goto(`/clubs/${id}`);
		} catch (err) {
			if (err instanceof ApiError && err.status === 404) {
				invalid = true;
				toast.error($_('club_join_invalid_toast'));
			} else {
				toast.error($_('club_join_error'));
			}
		} finally {
			joining = false;
		}
	}

	onMount(loadPreview);
</script>

<main class="pt-safe-page mx-auto flex min-h-screen max-w-[480px] flex-col px-6 pb-10">
	<div class="flex flex-1 flex-col items-center justify-center">
		{#if loading}
			<Spinner />
		{:else if invalid}
			<div class="space-y-6 text-center">
				<div class="space-y-2">
					<p class="text-text-primary text-lg font-semibold">{$_('club_join_invalid_title')}</p>
					<p class="text-text-secondary text-sm">{$_('club_join_invalid_desc')}</p>
				</div>
				<Button onclick={() => goto('/')} variant="secondary" size="cta"
					>{$_('club_join_back_home')}</Button
				>
			</div>
		{:else if preview}
			<!-- Club invite card — deliberately distinct from a Session join link:
			     "join this club", not "join this game". -->
			<div class="bg-surface-raised w-full space-y-6 rounded-3xl px-6 py-8 text-center">
				<p class="text-text-secondary text-[11px] font-semibold tracking-[0.14em] uppercase">
					{$_('club_join_eyebrow')}
				</p>

				<div class="flex flex-col items-center gap-3">
					<Avatar
						icon={preview.avatar_icon}
						color={preview.avatar_color}
						name={preview.name}
						size="xl"
					/>
					<div class="space-y-1">
						<h1 class="text-text-primary text-2xl font-bold">{preview.name}</h1>
						<p class="text-text-secondary text-sm">
							{preview.member_count}
							{preview.member_count === 1 ? $_('club_member') : $_('club_members')}
						</p>
					</div>
				</div>

				{#if auth.ready && !auth.user}
					<div class="space-y-3">
						<p class="text-text-secondary text-sm">
							{$_('club_join_login_prompt', { values: { name: preview.name } })}
						</p>
						<Button onclick={() => goto(authHref)} size="cta">{$_('club_join_login_button')}</Button
						>
						<p class="text-text-secondary text-sm">
							{$_('club_join_new_here')}
							<a href={registerHref} class="text-primary font-semibold"
								>{$_('club_join_create_account')}</a
							>
						</p>
					</div>
				{:else}
					<Button onclick={join} disabled={joining} size="cta">
						{joining ? $_('club_join_joining') : $_('club_join_button')}
					</Button>
				{/if}
			</div>
		{/if}
	</div>

	<Footer />
</main>
