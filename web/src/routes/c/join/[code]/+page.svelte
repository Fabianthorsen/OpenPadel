<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { api, ApiError } from '$lib/api/client';
	import { auth } from '$lib/auth.svelte';
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
				toast.error('Could not load this club invite');
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
			toast.success('Welcome to the club!');
			goto(`/clubs/${id}`);
		} catch (err) {
			if (err instanceof ApiError && err.status === 404) {
				invalid = true;
				toast.error('This club invite is no longer valid');
			} else {
				toast.error('Could not join the club');
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
					<p class="text-text-primary text-lg font-semibold">Invite not valid</p>
					<p class="text-text-secondary text-sm">
						This club invite link has expired or been revoked. Ask a club admin for a fresh link.
					</p>
				</div>
				<Button onclick={() => goto('/')} variant="secondary" size="cta">Back to home</Button>
			</div>
		{:else if preview}
			<!-- Club invite card — deliberately distinct from a Session join link:
			     "join this club", not "join this game". -->
			<div class="bg-surface-raised w-full space-y-6 rounded-3xl px-6 py-8 text-center">
				<p class="text-text-secondary text-[11px] font-semibold tracking-[0.14em] uppercase">
					Club invite
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
							{preview.member_count === 1 ? 'member' : 'members'}
						</p>
					</div>
				</div>

				{#if auth.ready && !auth.user}
					<div class="space-y-3">
						<p class="text-text-secondary text-sm">
							Log in or create an account to join <span class="font-semibold">{preview.name}</span>.
						</p>
						<Button onclick={() => goto(authHref)} size="cta">Log in to join</Button>
						<p class="text-text-secondary text-sm">
							New here?
							<a href={registerHref} class="text-primary font-semibold">Create an account</a>
						</p>
					</div>
				{:else}
					<Button onclick={join} disabled={joining} size="cta">
						{joining ? 'Joining…' : 'Join club'}
					</Button>
				{/if}
			</div>
		{/if}
	</div>

	<Footer />
</main>
