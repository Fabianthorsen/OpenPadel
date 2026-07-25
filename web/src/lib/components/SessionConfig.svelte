<script lang="ts">
	import { api, ApiError } from '$lib/api/client';
	import { Calendar } from '$lib/components/ui/calendar';
	import Stepper from '$lib/components/ui/stepper/Stepper.svelte';
	import { SegmentedControl, type SegmentedOption } from '$lib/components/ui/segmented-control';
	import { calculateAmericanoRounds } from '$lib/americano';
	import { Button } from '$lib/components/ui/button';
	import { Switch } from '$lib/components/ui/switch';
	import * as Drawer from '$lib/components/ui/drawer';
	import { Trophy, LayoutGrid, Target, Repeat, CalendarClock } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';
	import { translateApiError } from '$lib/i18n/errors';
	import { _ } from 'svelte-i18n';
	import {
		type DateValue,
		today,
		getLocalTimeZone,
		parseAbsoluteToLocal
	} from '@internationalized/date';

	let {
		session,
		open = $bindable(false),
		sessionId,
		onRefresh
	}: {
		session: App.Session;
		open: boolean;
		sessionId: string;
		/** Called after a config patch succeeds so the parent reloads the session. */
		onRefresh?: () => void;
	} = $props();

	// ── Config editing state ──
	let configMode = $state<'americano' | 'mexicano'>('americano');
	let configCourts = $state(2);
	let configPoints = $state(24);
	let configRounds = $state(7);
	let roundsMode = $state<'fixed' | 'unlimited'>('fixed');
	let scheduleEnabled = $state(false);
	let calendarDate = $state<DateValue | undefined>(undefined);
	let timeSlot = $state(20);

	let isUpdating = $state(false);

	// Sync config state whenever session prop changes.
	$effect(() => {
		configMode = session.game_mode as 'americano' | 'mexicano';
		configCourts = session.courts;
		configPoints = session.points;
		configRounds = session.rounds_total ?? 7;
		roundsMode =
			session.rounds_total === null || session.rounds_total === undefined ? 'unlimited' : 'fixed';
		scheduleEnabled = !!session.scheduled_at;
		if (session.scheduled_at) {
			try {
				calendarDate = parseAbsoluteToLocal(session.scheduled_at);
			} catch {
				calendarDate = undefined;
			}
			const d = new Date(session.scheduled_at);
			const slot = Math.round((d.getHours() * 60 + d.getMinutes() - 8 * 60) / 30);
			timeSlot = Math.max(0, Math.min(27, slot));
		} else {
			calendarDate = undefined;
		}
	});

	function slotToLabel(slot: number) {
		const totalMins = 8 * 60 + slot * 30;
		const h = String(Math.floor(totalMins / 60)).padStart(2, '0');
		const m = String(totalMins % 60).padStart(2, '0');
		return `${h}:${m}`;
	}

	const scheduleTime = $derived(slotToLabel(timeSlot));
	const timeHour = $derived(8 + Math.floor(timeSlot / 2));
	const timeMinute = $derived((timeSlot % 2) * 30);

	function onHourChange(h: number) {
		timeSlot = (h - 8) * 2 + (timeSlot % 2);
		commitScheduleTime();
	}

	function onMinuteChange(m: number) {
		timeSlot = (timeHour - 8) * 2 + Math.round(m / 30);
		commitScheduleTime();
	}

	function calculateNextHourSlot(): number {
		const now = new Date();
		const nextHour = now.getMinutes() > 0 ? now.getHours() + 1 : now.getHours();
		const clamped = Math.min(21, Math.max(8, nextHour));
		return Math.round((clamped * 60 - 8 * 60) / 30);
	}

	async function patchConfig(patch: Parameters<typeof api.sessions.update>[1]) {
		isUpdating = true;
		const adminToken = localStorage.getItem(`admin_token_${sessionId}`) ?? '';
		try {
			await api.sessions.update(sessionId, patch, adminToken);
			onRefresh?.();
		} catch (e) {
			toast.error(
				e instanceof ApiError ? translateApiError(e.message) : translateApiError('server_error')
			);
			// Reset local state to match server
			configMode = session.game_mode as 'americano' | 'mexicano';
			configCourts = session.courts;
			configPoints = session.points;
			configRounds = session.rounds_total ?? 7;
		} finally {
			isUpdating = false;
		}
	}

	function onModeChange(mode: 'americano' | 'mexicano') {
		configMode = mode;
		if (mode === 'mexicano' && configCourts < 2) configCourts = 2;
		const patch: Parameters<typeof api.sessions.update>[1] = {
			game_mode: mode,
			courts: configCourts
		};
		if (mode === 'mexicano') patch.rounds_total = configRounds;
		patchConfig(patch);
	}

	function onCourtsChange(n: number) {
		configCourts = n;
		patchConfig({ courts: n });
	}

	function onPointsChange(n: number) {
		configPoints = n;
		patchConfig({ points: n });
	}

	function onRoundsChange(n: number) {
		configRounds = n;
		patchConfig({ rounds_total: n });
	}

	function onRoundsModeChange(mode: 'fixed' | 'unlimited') {
		roundsMode = mode;
		if (mode === 'unlimited') {
			patchConfig({ rounds_total: null });
			return;
		}
		// A non-null rounds_total just signals "limited". Mexicano is user-picked;
		// Americano's count is derived by the backend, so we send the display value
		// (the backend recomputes the fair count from the roster at start anyway).
		patchConfig({ rounds_total: configMode === 'americano' ? americanoRounds : configRounds });
	}

	async function commitSchedule(enabled: boolean) {
		scheduleEnabled = enabled;
		if (!enabled) {
			calendarDate = undefined;
			timeSlot = 20;
			await patchConfig({ scheduled_at: '' });
			return;
		}
		calendarDate = today(getLocalTimeZone());
		timeSlot = calculateNextHourSlot();
	}

	async function commitScheduleTime() {
		if (!scheduleEnabled || !calendarDate) return;
		const [h, m] = scheduleTime.split(':').map(Number);
		const d = calendarDate.toDate(getLocalTimeZone());
		d.setHours(h, m, 0, 0);
		await patchConfig({ scheduled_at: d.toISOString() });
	}

	const MAX_COURTS = 4;
	const POINTS_OPTIONS = [16, 24, 32];

	// Derived so they react to mode/court changes (Mexicano disables 1 court).
	const modeOptions: SegmentedOption[] = [
		{ value: 'americano', label: 'Americano' },
		{ value: 'mexicano', label: 'Mexicano' }
	];
	const courtOptions = $derived(
		Array.from({ length: MAX_COURTS }, (_, i) => ({
			value: String(i + 1),
			label: String(i + 1),
			disabled: configMode === 'mexicano' && i + 1 === 1
		}))
	);
	const pointOptions: SegmentedOption[] = POINTS_OPTIONS.map((p) => ({
		value: String(p),
		label: String(p)
	}));

	// Mirror the backend, which schedules from active players only (removed players
	// are soft-deactivated but stay in session.players).
	const activePlayerCount = $derived(session.players.filter((p) => p.active).length);

	const minRounds = 1;
	const maxRounds = $derived(
		Math.max(
			7,
			configMode === 'americano' ? calculateAmericanoRounds(activePlayerCount, configCourts) : 5
		)
	);

	// Display-only preview of the fair Americano round count. Recomputes on every
	// settings change (players joining/leaving, court count). The backend is the
	// source of truth and recomputes the same value from the roster at start; this
	// just shows the admin roughly how long the tournament will be. Not shown when
	// the session is unlimited.
	const americanoRounds = $derived(calculateAmericanoRounds(activePlayerCount, configCourts));
</script>

<Drawer.Root bind:open>
	<Drawer.Content size="lg" class="mx-auto w-full max-w-[480px]">
		<Drawer.Header class="flex-row items-center justify-between">
			<Drawer.Title>{$_('lobby_edit_config')}</Drawer.Title>
			<Drawer.Close
				class="bg-surface-raised text-text-secondary hover:bg-border flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-xl leading-none transition-colors"
				>×</Drawer.Close
			>
		</Drawer.Header>

		<Drawer.Body class="space-y-5">
			<!-- Game Mode -->
			<section class="space-y-2.5">
				<div class="flex items-center gap-2">
					<Trophy size={15} class="text-primary" />
					<p class="text-sm font-semibold">{$_('create_game_mode_label')}</p>
				</div>
				<SegmentedControl
					options={modeOptions}
					value={configMode}
					onChange={(v) => onModeChange(v as 'americano' | 'mexicano')}
					ariaLabel={$_('create_game_mode_label')}
				/>
				<p class="text-text-secondary text-xs leading-relaxed">
					{configMode === 'mexicano' ? $_('create_mexicano_hint') : $_('create_americano_hint')}
				</p>
			</section>

			<div class="bg-border h-px"></div>

			<!-- Courts -->
			<section class="space-y-2.5">
				<div class="flex items-center gap-2">
					<LayoutGrid size={15} class="text-primary" />
					<p class="text-sm font-semibold">{$_('create_courts_label')}</p>
				</div>
				<SegmentedControl
					options={courtOptions}
					value={configCourts.toString()}
					onChange={(v) => onCourtsChange(parseInt(v))}
					ariaLabel={$_('create_courts_label')}
				/>
			</section>

			<div class="bg-border h-px"></div>

			<!-- Points -->
			<section class="space-y-2.5">
				<div class="flex items-center gap-2">
					<Target size={15} class="text-primary" />
					<p class="text-sm font-semibold">{$_('create_points_label')}</p>
				</div>
				<SegmentedControl
					options={pointOptions}
					value={configPoints.toString()}
					onChange={(v) => onPointsChange(parseInt(v))}
					ariaLabel={$_('create_points_label')}
				/>
			</section>

			<!-- Rounds (both modes support fixed / unlimited) -->
			<div class="bg-border h-px"></div>
			<section class="space-y-2.5">
				<div class="flex items-center gap-2">
					<Repeat size={15} class="text-primary" />
					<p class="text-sm font-semibold">{$_('lobby_rounds_label')}</p>
				</div>
				<SegmentedControl
					options={[
						{ value: 'fixed', label: $_('lobby_rounds_mode_fixed') },
						{ value: 'unlimited', label: $_('lobby_rounds_mode_unlimited') }
					]}
					value={roundsMode}
					onChange={(v) => onRoundsModeChange(v as 'fixed' | 'unlimited')}
					ariaLabel={$_('lobby_rounds_label')}
				/>
				<!-- Mexicano lets the admin pick the count. Americano's is derived by the
				     backend from the roster and shown in the lobby header, not here. -->
				{#if roundsMode === 'fixed' && configMode === 'mexicano'}
					<div
						class="bg-surface border-border flex items-center justify-between rounded-xl border px-4 py-2.5"
					>
						<span class="text-text-secondary text-sm">{$_('lobby_rounds_label')}</span>
						<Stepper
							bind:value={configRounds}
							onchange={onRoundsChange}
							min={minRounds}
							max={maxRounds}
							step={1}
						/>
					</div>
				{:else if roundsMode === 'fixed' && configMode === 'americano'}
					<p class="text-text-secondary text-xs">{$_('rounds_auto_hint')}</p>
				{/if}
			</section>

			<div class="bg-border h-px"></div>

			<!-- Schedule -->
			<section class="space-y-3">
				<div class="flex items-center justify-between gap-3">
					<div class="flex items-center gap-2">
						<CalendarClock size={15} class="text-primary" />
						<p class="text-sm font-semibold">{$_('schedule_session')}</p>
					</div>
					<Switch
						checked={scheduleEnabled}
						onCheckedChange={commitSchedule}
						aria-label={$_('schedule_session')}
					/>
				</div>
				{#if scheduleEnabled && calendarDate}
					<div class="space-y-4">
						<Calendar bind:value={calendarDate} onchange={commitScheduleTime} />
						<div class="grid grid-cols-2 gap-3">
							<div
								class="bg-surface border-border flex items-center justify-between rounded-xl border px-4 py-2.5"
							>
								<span class="text-text-secondary text-xs font-semibold">
									{$_('schedule_hour_label')}
								</span>
								<Stepper value={timeHour} onchange={onHourChange} min={8} max={21} step={1} />
							</div>
							<div
								class="bg-surface border-border flex items-center justify-between rounded-xl border px-4 py-2.5"
							>
								<span class="text-text-secondary text-xs font-semibold">
									{$_('schedule_minute_label')}
								</span>
								<Stepper value={timeMinute} onchange={onMinuteChange} min={0} max={30} step={30} />
							</div>
						</div>
					</div>
				{/if}
			</section>
		</Drawer.Body>

		<Drawer.Footer>
			<Button class="w-full" onclick={() => (open = false)}>
				{$_('done')}
			</Button>
		</Drawer.Footer>
	</Drawer.Content>
</Drawer.Root>
