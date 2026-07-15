<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import { auth } from '$lib/auth.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { SectionLabel } from '$lib/components/ui/section-label';
	import { PillToggleGroup, PillToggleItem } from '$lib/components/ui/pill-toggle-group';
	import { Switch } from '$lib/components/ui/switch';
	import Footer from '$lib/components/Footer.svelte';
	import { _ } from 'svelte-i18n';
	import { initials } from '$lib/utils';
	import { toast } from 'svelte-sonner';
	import { ApiError } from '$lib/api/client';
	import { translateApiError } from '$lib/i18n/errors';
	import { Calendar } from '$lib/components/ui/calendar';
	import { type DateValue, today, getLocalTimeZone } from '@internationalized/date';

	let step = $state<'home' | 'setup'>('home');
	let courts = $state(2);
	let points = $state(24);
	let tournamentName = $state('');
	let scheduleEnabled = $state(false);
	let calendarDate = $state<DateValue | undefined>(undefined);
	// Slider: 0 = 08:00, 1 = 08:30, ..., 27 = 21:30
	let timeSlot = $state(20); // default 18:00

	function slotToLabel(slot: number) {
		const totalMins = 8 * 60 + slot * 30;
		const h = String(Math.floor(totalMins / 60)).padStart(2, '0');
		const m = String(totalMins % 60).padStart(2, '0');
		return `${h}:${m}`;
	}

	function calculateNextHourSlot(): number {
		const now = new Date();
		const currentHour = now.getHours();
		const currentMinutes = now.getMinutes();

		// Calculate next whole hour
		const nextHour = currentMinutes > 0 ? currentHour + 1 : currentHour;

		// Clamp to 8-21 range (08:00 to 21:30)
		const clampedHour = Math.min(21, Math.max(8, nextHour));

		// Convert hour to slot (slot 0 = 08:00, slot 27 = 21:30)
		// slot = (hour * 60 - 8 * 60) / 30
		return Math.round((clampedHour * 60 - 8 * 60) / 30);
	}

	const scheduleTime = $derived(slotToLabel(timeSlot));
	let creating = $state(false);
	let joinCode = $state('');
	let rejoinSession = $state<App.Session | null>(null);
	let rejoinHref = $state('');

	async function loadRejoin() {
		const lastId = localStorage.getItem('last_session_id');
		if (!lastId) {
			rejoinSession = null;
			return;
		}
		try {
			const token = localStorage.getItem(`admin_token_${lastId}`) ?? undefined;
			const s = await api.sessions.get(lastId, token);
			if (s.status === 'lobby' || s.status === 'playing') {
				rejoinSession = s;
				rejoinHref = token ? `/s/${lastId}?token=${token}` : `/s/${lastId}`;
			} else {
				rejoinSession = null;
			}
		} catch {
			localStorage.removeItem('last_session_id');
			rejoinSession = null;
		}
	}

	onMount(async () => {
		// If ?create=1, go straight to setup (used by profile "New tournament" link)
		if (page.url.searchParams.get('create') === '1') {
			step = 'setup';
		}

		if (page.url.searchParams.get('deleted') === '1') {
			toast($_('home_account_deleted'));
		}
		if (page.url.searchParams.get('notfound') === '1') {
			toast.error($_('home_session_not_found'));
		}

		await loadRejoin();
	});

	// Redirect logged-in users to profile
	$effect(() => {
		if (auth.ready && auth.user && step === 'home') {
			const notfound = page.url.searchParams.get('notfound');
			const create = page.url.searchParams.get('create');
			goto(
				(notfound ? '/profile?notfound=1' : '/profile') +
					(create ? (notfound ? '&create=1' : '?create=1') : '')
			);
		}
	});

	async function create() {
		const effectiveName = auth.user!.display_name;
		creating = true;
		try {
			let iso: string | undefined;
			if (scheduleEnabled && calendarDate) {
				const [h, m] = scheduleTime.split(':').map(Number);
				const d = calendarDate.toDate(getLocalTimeZone());
				d.setHours(h, m, 0, 0);
				iso = d.toISOString();
			}
			const session = await api.sessions.create(
				{
					courts,
					points,
					name: tournamentName.trim(),
					game_mode: 'americano',
					scheduled_at: iso
				},
				auth.token ?? undefined
			);
			const adminToken = session.admin_token!;
			localStorage.setItem(`admin_token_${session.id}`, adminToken);
			const player = await api.players.join(
				session.id,
				effectiveName,
				auth.token ?? undefined,
				adminToken
			);
			localStorage.setItem(`player_id_${session.id}`, player.id);
			localStorage.setItem('last_session_id', session.id);
			goto(`/s/${session.id}?token=${adminToken}`);
		} catch (e) {
			toast.error(
				e instanceof ApiError ? translateApiError(e.message) : translateApiError('server_error')
			);
			creating = false;
		}
	}

	function joinByCode() {
		const code = joinCode.trim().toUpperCase();
		if (code) goto(`/s/${code}`);
	}
</script>

{#if step === 'home'}
	<main class="pt-safe-page flex min-h-svh flex-col items-center px-6 pb-12">
		<div class="flex w-full max-w-sm flex-1 flex-col">
			<div class="flex flex-1 flex-col justify-center space-y-12">
				<!-- Brand -->
				<div class="space-y-1">
					<h1 class="text-primary text-[28px] font-[800]">OpenPadel</h1>
					<p class="text-text-secondary">{$_('home_tagline')}</p>
				</div>

				<!-- Actions -->
				<div class="space-y-4">
					{#if rejoinSession}
						<a
							href={rejoinHref}
							class="bg-surface-raised hover:bg-border flex items-center gap-3 rounded-2xl px-4 py-3.5 transition-colors"
						>
							<div
								class="bg-primary-muted flex h-9 w-9 shrink-0 items-center justify-center rounded-full"
							>
								<div class="bg-primary h-2.5 w-2.5 animate-pulse rounded-full"></div>
							</div>
							<div class="min-w-0 flex-1">
								<p class="text-text-disabled text-[11px] font-bold tracking-[0.1em] uppercase">
									{$_('home_rejoin_label')}
								</p>
								<p class="truncate text-sm font-semibold">{rejoinSession.name || 'OpenPadel'}</p>
							</div>
							<span class="text-text-secondary text-sm">→</span>
						</a>
					{/if}

					{#if auth.ready && !auth.user}
						<a
							href="/auth"
							class="bg-primary hover:bg-primary-hover flex h-auto w-full items-center justify-center rounded-2xl px-4 py-4 text-[15px] font-semibold text-white"
						>
							{$_('auth_sign_in')} →
						</a>
						<p class="text-text-secondary text-center text-sm">
							{$_('auth_no_account')}
							<a
								href="/auth?register=1"
								class="text-primary hover:text-primary-hover font-semibold"
							>
								{$_('auth_switch_register')}
							</a>
						</p>
					{/if}

					<div class="flex items-center gap-3">
						<div class="bg-border h-px flex-1"></div>
						<span class="text-text-disabled text-xs">{$_('home_join_code_divider')}</span>
						<div class="bg-border h-px flex-1"></div>
					</div>

					<form
						onsubmit={(e) => {
							e.preventDefault();
							joinByCode();
						}}
						class="flex gap-2"
					>
						<Input
							bind:value={joinCode}
							oninput={(e: Event) => {
								joinCode = (e.currentTarget as HTMLInputElement).value.toUpperCase();
							}}
							placeholder={$_('home_join_code_placeholder')}
							maxlength={4}
							autocomplete="off"
							autocorrect="off"
							autocapitalize="characters"
							spellcheck={false}
							class="bg-surface-raised min-w-0 flex-1 rounded-2xl border-0 px-4 py-3.5 text-sm"
						/>
						<Button
							type="submit"
							disabled={!joinCode.trim()}
							variant="secondary"
							class="bg-surface-raised text-text-primary hover:bg-border h-auto rounded-2xl px-5 text-sm font-semibold"
						>
							{$_('home_join_button')}
						</Button>
					</form>
				</div>
			</div>

			<!-- Auth (logged-in pill — guests see the sign-in button above instead) -->
			{#if auth.user}
				<div class="flex justify-center pt-6">
					<a
						href="/profile"
						class="hover:bg-surface-raised flex items-center gap-3 rounded-2xl px-3 py-2 transition-colors"
					>
						<div
							class="bg-primary-muted text-primary flex h-7 w-7 items-center justify-center rounded-full text-xs font-[800]"
						>
							{initials(auth.user.display_name)}
						</div>
						<span class="text-sm font-semibold">{auth.user.display_name}</span>
						<span class="text-text-disabled text-xs">→</span>
					</a>
				</div>
			{/if}
		</div>
		<Footer />
	</main>
{:else}
	<main class="pt-safe-page flex min-h-svh flex-col items-center px-6 pb-6">
		<div class="w-full max-w-sm">
			<!-- Nav -->
			<nav class="flex items-center justify-between">
				<Button
					onclick={() => goto('/profile')}
					variant="ghost"
					class="text-text-secondary flex h-8 w-8 items-center justify-center rounded-full p-0 text-lg"
				>
					×
				</Button>
				<span class="text-primary text-sm font-semibold">OpenPadel</span>
				<div class="w-8"></div>
			</nav>

			<!-- Header -->
			<div class="mt-8 space-y-2">
				<h1 class="text-[34px] font-[800]">
					{$_('create_title_line1')}<br />{$_('create_title_line2')}
				</h1>
				<p class="text-text-secondary">{$_('create_subtitle')}</p>
			</div>

			<!-- Form -->
			<div class="mt-8 space-y-7">
				<!-- Courts -->
				<div class="space-y-2.5">
					<SectionLabel>{$_('create_courts_label')}</SectionLabel>
					<PillToggleGroup
						value={courts.toString()}
						onValueChange={(val) => (courts = parseInt(val))}
					>
						{#each [1, 2, 3, 4] as n}
							<PillToggleItem value={n.toString()}>
								{n}
							</PillToggleItem>
						{/each}
					</PillToggleGroup>
				</div>

				<!-- Points -->
				<div class="space-y-2.5">
					<SectionLabel>{$_('create_points_label')}</SectionLabel>
					<PillToggleGroup
						value={points.toString()}
						onValueChange={(val) => (points = parseInt(val))}
					>
						{#each [16, 24, 32] as p}
							<PillToggleItem value={p.toString()}>
								{p}
							</PillToggleItem>
						{/each}
					</PillToggleGroup>
					<p class="text-text-secondary text-xs">
						{points === 16
							? $_('create_points_quick')
							: points === 24
								? $_('create_points_standard')
								: $_('create_points_long')}
					</p>
				</div>

				<!-- Tournament name (optional) -->
				<div class="space-y-2.5">
					<SectionLabel>{$_('create_tournament_name_label')}</SectionLabel>
					<Input
						bind:value={tournamentName}
						placeholder={$_('create_tournament_name_placeholder')}
						maxlength={48}
						class="bg-surface-raised rounded-2xl border-0 px-4 py-3.5 text-sm"
					/>
				</div>

				<!-- Schedule (optional) -->
				<div class="space-y-2.5">
					<div class="flex items-center justify-between">
						<SectionLabel>{$_('create_schedule_label')}</SectionLabel>
						<Switch
							checked={scheduleEnabled}
							onCheckedChange={(checked) => {
								scheduleEnabled = checked;
								if (!checked) {
									calendarDate = undefined;
									timeSlot = 20;
								} else {
									calendarDate = today(getLocalTimeZone());
									timeSlot = calculateNextHourSlot();
								}
							}}
						/>
					</div>
					{#if scheduleEnabled}
						<div class="bg-surface-raised overflow-hidden rounded-2xl">
							<Calendar
								bind:value={calendarDate}
								minValue={today(getLocalTimeZone())}
								weekStartsOn={1}
							/>
							<div class="space-y-2 px-4 pb-4">
								<div class="flex items-center justify-between">
									<p
										class="text-text-disabled text-[11px] font-semibold tracking-[0.1em] uppercase"
									>
										{$_('create_schedule_time_label')}
									</p>
									<p class="text-primary text-sm font-[800]">{scheduleTime}</p>
								</div>
								<input
									type="range"
									min="0"
									max="27"
									step="1"
									bind:value={timeSlot}
									class="accent-primary w-full"
								/>
								<div class="text-text-disabled flex justify-between text-[10px]">
									<span>08:00</span>
									<span>21:30</span>
								</div>
							</div>
						</div>
					{/if}
				</div>

				<!-- Organiser (always logged in at this point) -->
				<div class="space-y-2.5">
					<SectionLabel>{$_('create_organiser_label')}</SectionLabel>
					<div class="bg-surface-raised flex items-center gap-3 rounded-2xl px-4 py-3.5">
						<div
							class="bg-primary-muted text-primary flex h-7 w-7 items-center justify-center rounded-full text-xs font-[800]"
						>
							{auth.user ? initials(auth.user.display_name) : '?'}
						</div>
						<span class="text-sm font-semibold">{auth.user?.display_name}</span>
					</div>
				</div>

				<!-- Info note -->
				<div class="bg-surface-raised flex gap-3 rounded-2xl px-4 py-3.5">
					<span class="text-text-secondary mt-px shrink-0">ℹ</span>
					<p class="text-text-secondary text-sm">{$_('create_info_note')}</p>
				</div>

				<Button
					onclick={create}
					disabled={creating}
					class="bg-primary hover:bg-primary-hover h-auto w-full rounded-2xl px-4 py-4 text-[15px] font-semibold text-white"
				>
					{creating ? $_('create_button_loading') : $_('create_button')}
				</Button>
			</div>
		</div>
	</main>
{/if}
