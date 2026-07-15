<script lang="ts">
	import { api, ApiError } from '$lib/api/client';
	import { Calendar } from '$lib/components/ui/calendar';
	import Stepper from '$lib/components/ui/stepper/Stepper.svelte';
	import { PillToggleGroup, PillToggleItem } from '$lib/components/ui/pill-toggle-group';
	import { Button } from '$lib/components/ui/button';
	import * as Drawer from '$lib/components/ui/drawer';
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
		sessionId
	}: {
		session: App.Session;
		open: boolean;
		sessionId: string;
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

	function gcd(a: number, b: number): number {
		return b === 0 ? a : gcd(b, a % b);
	}

	function calculateAmericanoRounds(players: number, courts: number): number {
		const benchSize = players - courts * 4;
		if (benchSize <= 0) {
			return players - 1;
		}
		const cycle = players / gcd(players, benchSize);
		const target = players - 1;
		return Math.ceil(target / cycle) * cycle;
	}

	async function patchConfig(patch: Parameters<typeof api.sessions.update>[1]) {
		isUpdating = true;
		const adminToken = localStorage.getItem(`admin_token_${sessionId}`) ?? '';
		try {
			await api.sessions.update(sessionId, patch, adminToken);
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
		} else {
			patchConfig({ rounds_total: configRounds });
		}
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

	const maxCourts = 4;
	const courtsDisabled = configMode === 'mexicano' ? [1] : [];

	const minRounds = 1;
	const maxRounds = Math.max(
		7,
		configMode === 'americano' ? calculateAmericanoRounds(session.players.length, configCourts) : 5
	);
</script>

<Drawer.Root bind:open>
	<Drawer.Content class="mx-auto w-full max-w-[480px]">
		<Drawer.Header>
			<h2 class="text-lg font-[800]">{$_('lobby_edit_config')}</h2>
			<Drawer.Close
				class="bg-surface-raised text-text-secondary hover:bg-border h-8 w-8 items-center justify-center rounded-full text-xl leading-none transition-colors"
				>×</Drawer.Close
			>
		</Drawer.Header>

		<div class="space-y-6 overflow-y-auto px-6 pb-8">
			<!-- Game Mode -->
			<div class="space-y-2">
				<p class="text-sm font-semibold">{$_('create_game_mode_label')}</p>
				<PillToggleGroup value={configMode} onchange={onModeChange}>
					<PillToggleItem value="americano">Americano</PillToggleItem>
					<PillToggleItem value="mexicano">Mexicano</PillToggleItem>
				</PillToggleGroup>
			</div>

			<!-- Courts -->
			<div class="space-y-2">
				<p class="text-sm font-semibold">{$_('create_courts_label')}</p>
				<PillToggleGroup
					value={configCourts.toString()}
					onchange={(v: string) => onCourtsChange(parseInt(v))}
				>
					{#each Array.from({ length: maxCourts }, (_, i) => i + 1) as court}
						<PillToggleItem value={court.toString()} disabled={courtsDisabled.includes(court)}>
							{court}
						</PillToggleItem>
					{/each}
				</PillToggleGroup>
			</div>

			<!-- Points -->
			<div class="space-y-2">
				<p class="text-sm font-semibold">{$_('create_points_label')}</p>
				<Stepper bind:value={configPoints} onchange={onPointsChange} min={11} max={31} step={1} />
			</div>

			<!-- Rounds -->
			{#if configMode === 'mexicano'}
				<div class="space-y-2">
					<p class="text-sm font-semibold">{$_('lobby_rounds_label')}</p>
					<PillToggleGroup
						value={roundsMode}
						onchange={(v: string) => onRoundsModeChange(v as 'fixed' | 'unlimited')}
					>
						<PillToggleItem value="fixed">{$_('lobby_rounds_mode_fixed')}</PillToggleItem>
						<PillToggleItem value="unlimited">{$_('lobby_rounds_mode_unlimited')}</PillToggleItem>
					</PillToggleGroup>
					{#if roundsMode === 'fixed'}
						<Stepper
							bind:value={configRounds}
							onchange={onRoundsChange}
							min={minRounds}
							max={maxRounds}
							step={1}
						/>
					{/if}
				</div>
			{/if}

			<!-- Schedule -->
			<div class="space-y-2">
				<label class="flex items-center gap-2">
					<input
						type="checkbox"
						checked={scheduleEnabled}
						onchange={(e) => commitSchedule(e.currentTarget.checked)}
						class="rounded"
					/>
					<span class="text-sm font-semibold">{$_('schedule_session')}</span>
				</label>
				{#if scheduleEnabled && calendarDate}
					<div class="space-y-4">
						<Calendar bind:value={calendarDate} onchange={commitScheduleTime} />
						<div class="grid grid-cols-2 gap-3">
							<div class="space-y-1">
								<p class="text-text-secondary text-xs font-semibold">{$_('schedule_hour_label')}</p>
								<Stepper value={timeHour} onchange={onHourChange} min={8} max={21} step={1} />
							</div>
							<div class="space-y-1">
								<p class="text-text-secondary text-xs font-semibold">
									{$_('schedule_minute_label')}
								</p>
								<Stepper value={timeMinute} onchange={onMinuteChange} min={0} max={30} step={30} />
							</div>
						</div>
						<p class="text-text-secondary text-sm">
							{$_('scheduled_for')}
							{calendarDate.toString()}
							{scheduleTime}
						</p>
					</div>
				{/if}
			</div>

			<Button class="w-full" onclick={() => (open = false)}>
				{$_('done')}
			</Button>
		</div>
	</Drawer.Content>
</Drawer.Root>
