<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth.svelte';
	import { api } from '$lib/api/client';
	import { _ } from 'svelte-i18n';
	import { ChevronLeft } from 'lucide-svelte';
	import Footer from '$lib/components/Footer.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import Switch from '$lib/components/ui/switch/switch.svelte';
	import LocaleSwitcher from '$lib/components/LocaleSwitcher.svelte';
	import Section from '$lib/components/ui/section/section.svelte';
	import { toast } from 'svelte-sonner';
	import { subscribeToPush, unsubscribeFromPush } from '$lib/push';

	const AVATAR_ICONS = [
		'Zap',
		'Star',
		'Flame',
		'Shield',
		'Crown',
		'Trophy',
		'Target',
		'Rocket',
		'Ghost',
		'Cat',
		'Dog',
		'Bird',
		'Leaf',
		'Sun',
		'Moon',
		'Snowflake',
		'Mountain',
		'Waves',
		'Music',
		'Heart',
		'Smile',
		'Fish',
		'Swords',
		'Dumbbell',
		'Bike',
		'Footprints'
	];

	const AVATAR_COLORS = ['forest', 'ocean', 'sunset', 'grape', 'peach'] as const;

	let displayName = $state(auth.user?.display_name ?? '');
	let pickerIcon = $state(auth.user?.avatar_icon ?? '');
	let pickerColor = $state((auth.user?.avatar_color ?? 'forest') as (typeof AVATAR_COLORS)[number]);
	let savingProfile = $state(false);
	let profileError = $state('');

	let pushSupported = $state(false);
	let pushEnabled = $state(false);
	let pushToggling = $state(false);

	// Install prompt
	let isStandalone = $state(false);
	let isIOS = $state(false);
	let deferredInstallPrompt = $state<any>(null);
	let installDismissed = $state(false);

	let showDeleteConfirm = $state(false);
	let deleting = $state(false);

	async function checkPushState() {
		const reg = await navigator.serviceWorker.ready;
		const sub = await reg.pushManager.getSubscription();
		pushEnabled = !!sub && Notification.permission === 'granted';
	}

	async function togglePush() {
		if (!auth.token) return;
		pushToggling = true;
		try {
			if (pushEnabled) {
				await unsubscribeFromPush(auth.token);
			} else {
				await subscribeToPush(auth.token);
			}
			await checkPushState();
		} catch (e) {
			const msg = e instanceof Error ? e.message : 'unknown';
			const label =
				msg === 'notifications_blocked'
					? $_('pref_notifications_blocked', { values: { app: 'OpenPadel' } })
					: msg === 'sw_timeout'
						? $_('pref_notifications_sw_timeout')
						: msg;
			toast.error(label);
		} finally {
			pushToggling = false;
		}
	}

	async function saveProfile() {
		if (!auth.token || !auth.user) return;
		if (!displayName.trim()) {
			profileError = $_('settings_name_required');
			return;
		}
		savingProfile = true;
		profileError = '';
		try {
			const updated = await api.auth.updateProfile(
				auth.token,
				displayName.trim(),
				pickerIcon,
				pickerColor
			);
			auth.updateUser(updated);
			toast.success($_('settings_profile_saved'));
		} catch (e) {
			profileError = e instanceof Error ? e.message : 'Failed to save profile';
			toast.error(profileError);
		} finally {
			savingProfile = false;
		}
	}

	async function deleteAccount() {
		if (!auth.token) return;
		deleting = true;
		try {
			await api.auth.deleteAccount(auth.token);
			await auth.logout();
			goto('/?deleted=1');
		} finally {
			deleting = false;
		}
	}

	onMount(async () => {
		if (!auth.token) {
			goto('/auth');
			return;
		}

		// Install detection
		isStandalone =
			window.matchMedia('(display-mode: standalone)').matches ||
			(navigator as any).standalone === true;
		isIOS = /iphone|ipad|ipod/i.test(navigator.userAgent);
		const isMobile = /iphone|ipad|ipod|android/i.test(navigator.userAgent);
		installDismissed = !isMobile || localStorage.getItem('install_dismissed') === '1';

		window.addEventListener('beforeinstallprompt', (e: any) => {
			e.preventDefault();
			deferredInstallPrompt = e;
		});

		// Push subscription check
		if ('serviceWorker' in navigator && 'PushManager' in window) {
			pushSupported = true;
			try {
				const swReady = Promise.race([
					navigator.serviceWorker.ready,
					new Promise<never>((_, reject) => setTimeout(() => reject(new Error('timeout')), 3000))
				]);
				const reg = (await swReady) as ServiceWorkerRegistration;
				const sub = await reg.pushManager.getSubscription();
				pushEnabled = !!sub && Notification.permission === 'granted';
			} catch {
				// SW not ready (dev mode or not yet installed)
			}
		}
	});
</script>

<main class="pt-safe-page mx-auto max-w-[480px] space-y-8 px-6 pb-10">
	<!-- Header -->
	<div class="flex items-center gap-4">
		<a
			href="/profile"
			class="text-text-secondary hover:text-text-primary transition-colors"
			aria-label="Back"
		>
			<ChevronLeft size={24} />
		</a>
		<h1 class="text-2xl font-[800]">{$_('settings_title')}</h1>
	</div>

	<!-- Profile section -->
	<Section title={$_('settings_profile_section')} collapsible={false}>
		{#snippet children()}
			<div class="space-y-4">
				<!-- Avatar picker -->
				<div class="space-y-2">
					<p class="text-sm font-semibold">{$_('settings_avatar')}</p>
					<div class="flex gap-4">
						<Avatar icon={pickerIcon} color={pickerColor} name={displayName} size="lg" />
						<div class="flex-1 space-y-3">
							<!-- Icon grid -->
							<div class="space-y-2">
								<p class="text-text-secondary text-xs font-semibold">
									{$_('settings_avatar_icon')}
								</p>
								<div class="grid grid-cols-4 gap-1.5">
									<!-- Initials option -->
									<button
										onclick={() => (pickerIcon = '')}
										class="flex items-center justify-center rounded-lg p-1.5 transition-colors {pickerIcon ===
										''
											? 'bg-primary/15 ring-primary ring-2'
											: 'bg-surface-raised hover:bg-border'}"
										aria-label="Use initials"
										title="Initials"
									>
										<Avatar icon="" color={pickerColor} name={displayName} size="sm" />
									</button>
									{#each AVATAR_ICONS.slice(0, 7) as icon}
										<button
											onclick={() => (pickerIcon = icon)}
											class="flex items-center justify-center rounded-lg p-1.5 transition-colors {pickerIcon ===
											icon
												? 'bg-primary/15 ring-primary ring-2'
												: 'bg-surface-raised hover:bg-border'}"
											aria-label={icon}
											title={icon}
										>
											<Avatar {icon} color={pickerColor} name="" size="sm" />
										</button>
									{/each}
								</div>
								<button
									class="text-primary hover:text-primary/80 w-full text-left text-xs font-semibold transition-colors"
									onclick={() => {
										/* TODO: expand to show all icons */
									}}
								>
									{$_('settings_avatar_more')}
								</button>
							</div>

							<!-- Color swatches -->
							<div class="space-y-2">
								<p class="text-text-secondary text-xs font-semibold">
									{$_('settings_avatar_color')}
								</p>
								<div class="flex gap-2">
									{#each AVATAR_COLORS as color}
										<button
											onclick={() => (pickerColor = color)}
											class="h-8 w-8 rounded-full ring-offset-2 transition-all {pickerColor ===
											color
												? 'ring-text-secondary ring-2'
												: 'hover:opacity-80'}"
											style="background-color: var(--color-{color})"
											aria-label={color}
											title={color}
										></button>
									{/each}
								</div>
							</div>
						</div>
					</div>
				</div>

				<!-- Name edit -->
				<div class="space-y-2">
					<label for="display-name" class="text-sm font-semibold">{$_('settings_name')}</label>
					<input
						id="display-name"
						type="text"
						bind:value={displayName}
						maxlength="32"
						class="bg-surface-raised focus:ring-primary w-full rounded-xl px-4 py-2.5 text-sm transition-shadow outline-none focus:ring-2"
					/>
					{#if profileError}
						<p class="text-destructive text-xs">{profileError}</p>
					{/if}
				</div>

				<!-- Save button -->
				<button
					onclick={saveProfile}
					disabled={savingProfile}
					class="bg-primary w-full rounded-2xl py-3.5 text-sm font-semibold text-white disabled:opacity-50"
				>
					{savingProfile ? $_('settings_saving') : $_('settings_save')}
				</button>
			</div>
		{/snippet}
	</Section>

	<!-- Preferences section -->
	<Section title={$_('pref_section')} collapsible={false}>
		{#snippet children()}
			<div class="space-y-3">
				<!-- Notifications -->
				{#if pushSupported}
					<div class="flex items-center justify-between gap-4">
						<div class="flex-1">
							<p class="text-sm font-semibold">{$_('pref_notifications_title')}</p>
							<p class="text-text-secondary text-xs">{$_('pref_notifications_desc')}</p>
						</div>
						<Switch bind:checked={pushEnabled} onchange={togglePush} disabled={pushToggling} />
					</div>
				{/if}

				<!-- Install prompt -->
				{#if !isStandalone && !installDismissed && (isIOS || deferredInstallPrompt)}
					<div class="bg-primary-muted flex items-start gap-3 rounded-2xl px-4 py-3.5">
						<div class="flex-1 space-y-0.5">
							{#if isIOS}
								<p class="text-primary text-sm font-semibold">{$_('pwa_ios_title')}</p>
								<p class="text-text-secondary text-xs">
									{$_('pwa_ios_tap')} <span class="font-semibold">{$_('pwa_ios_share')}</span>
									{$_('pwa_ios_then')} <span class="font-semibold">{$_('pwa_ios_add')}</span>
									{$_('pwa_ios_suffix')}
								</p>
							{:else if deferredInstallPrompt}
								<p class="text-primary text-sm font-semibold">{$_('pwa_android_title')}</p>
								<p class="text-text-secondary text-xs">{$_('pwa_android_desc')}</p>
							{/if}
						</div>
						<div class="flex shrink-0 items-center gap-2">
							{#if deferredInstallPrompt && !isIOS}
								<button
									onclick={async () => {
										deferredInstallPrompt.prompt();
										const { outcome } = await deferredInstallPrompt.userChoice;
										if (outcome === 'accepted') {
											deferredInstallPrompt = null;
											isStandalone = true;
										}
									}}
									class="bg-primary rounded-full px-3 py-1 text-xs font-semibold text-white"
									>{$_('pwa_install_btn')}</button
								>
							{/if}
							<button
								onclick={() => {
									installDismissed = true;
									localStorage.setItem('install_dismissed', '1');
								}}
								class="text-text-disabled hover:text-text-secondary text-lg leading-none"
								aria-label="Dismiss">✕</button
							>
						</div>
					</div>
				{/if}

				<!-- Locale -->
				<div class="flex items-center justify-between gap-4">
					<div class="flex-1">
						<p class="text-sm font-semibold">{$_('settings_language')}</p>
						<p class="text-text-secondary text-xs">{$_('settings_language_desc')}</p>
					</div>
					<LocaleSwitcher />
				</div>
			</div>
		{/snippet}
	</Section>

	<!-- Account section -->
	<Section title={$_('settings_account_section')} collapsible={false}>
		{#snippet children()}
			<div class="space-y-2">
				<button
					onclick={() => auth.logout().then(() => goto('/'))}
					class="border-border text-text-secondary hover:border-text-secondary hover:text-text-secondary w-full rounded-2xl border px-4 py-3.5 text-sm font-semibold transition-colors"
				>
					{$_('auth_sign_out')}
				</button>
				<button
					onclick={() => (showDeleteConfirm = true)}
					class="text-destructive hover:bg-destructive/10 w-full rounded-2xl px-4 py-3.5 text-sm font-semibold transition-colors"
				>
					{$_('profile_delete_account')}
				</button>
			</div>
		{/snippet}
	</Section>

	<!-- Delete account confirmation -->
	<ConfirmDialog
		open={showDeleteConfirm}
		title={$_('profile_delete_title')}
		description={$_('profile_delete_desc')}
		confirmLabel={$_('profile_delete_confirm')}
		cancelLabel={$_('profile_delete_cancel')}
		destructive={true}
		onconfirm={deleteAccount}
		oncancel={() => (showDeleteConfirm = false)}
	/>

	<Footer />
</main>
