<script lang="ts">
	import { _ } from 'svelte-i18n';
	import { formStrip, recentRecord } from './form';

	/**
	 * Cross-mode recent form: a Point Win % sparkline over the player's most recent
	 * matches (both Game Modes, oldest→newest), plus the win/draw/loss record over
	 * that window — all derived client-side from the per-Match results series
	 * (ADR 0007). Form is mode-agnostic, so the series is passed whole rather than
	 * segmented.
	 *
	 * The 50% midline is the crux of the read: a match's Point Win % exceeds 50%
	 * exactly when it was won (points > conceded), so a dot's position above/below
	 * the line encodes the win/draw/loss outcome — the same thing its colour shows.
	 * Outcome therefore never rides on colour alone. The curve traces trajectory:
	 * a climb is improving form, a dip a slump.
	 *
	 * Renders nothing when there are no matches, so an empty career never shows a
	 * hollow strip. A single match reads its own real Point Win %, never a jarring
	 * 0%/100%.
	 *
	 * @example
	 * <FormStrip series={allMatches} />
	 */
	let { series }: { series: App.MatchResult[] } = $props();

	const bars = $derived(formStrip(series));
	const record = $derived(recentRecord(series));
	// The win/draw/loss record label, shown as text and reused as the SVG's a11y
	// label so screen readers get the same summary sighted users read.
	const recordLabel = $derived(
		$_('stats_form_record', {
			values: { wins: record.wins, draws: record.draws, losses: record.losses }
		})
	);

	// Sparkline geometry in a fixed viewBox; the SVG scales to the card width while
	// the container's matching aspect-ratio keeps the markers perfectly round.
	const W = 320;
	const H = 64;
	// A wider left gutter leaves room for the 50% axis label; the plot area spans
	// PAD_LEFT..(W − PAD_RIGHT).
	const PAD_LEFT = 30;
	const PAD_RIGHT = 12;
	const PAD_Y = 12;

	function xAt(i: number, n: number): number {
		if (n <= 1) return (PAD_LEFT + (W - PAD_RIGHT)) / 2;
		return PAD_LEFT + (W - PAD_RIGHT - PAD_LEFT) * (i / (n - 1));
	}
	function yAt(pct: number): number {
		const clamped = Math.max(0, Math.min(100, pct));
		return H - PAD_Y - (clamped / 100) * (H - 2 * PAD_Y);
	}

	const points = $derived(
		bars.map((bar, i) => ({
			cx: xAt(i, bars.length),
			cy: yAt(bar.pointWinPct),
			pct: bar.pointWinPct,
			result: bar.result
		}))
	);

	// A smooth cubic-Bézier curve through the points, using monotone (Fritsch–
	// Carlson) tangents so the curve never overshoots past the data — critical here
	// because an overshooting spline could bulge across the 50% midline and imply a
	// win/loss the matches don't support.
	function smoothPath(pts: { cx: number; cy: number }[]): string {
		const n = pts.length;
		if (n < 2) return '';
		if (n === 2) return `M${pts[0].cx} ${pts[0].cy} L${pts[1].cx} ${pts[1].cy}`;

		// Secant slopes between consecutive points.
		const delta: number[] = [];
		for (let i = 0; i < n - 1; i++) {
			delta.push((pts[i + 1].cy - pts[i].cy) / (pts[i + 1].cx - pts[i].cx));
		}
		// Initial tangents: endpoints take the adjacent secant; interiors average the
		// two secants, but flatten to 0 at a local extremum (secants of opposite sign
		// or a plateau) so the curve turns without overshooting past the point.
		const m: number[] = new Array(n);
		m[0] = delta[0];
		m[n - 1] = delta[n - 2];
		for (let i = 1; i < n - 1; i++) {
			m[i] = delta[i - 1] * delta[i] <= 0 ? 0 : (delta[i - 1] + delta[i]) / 2;
		}
		// Enforce monotonicity: flatten at plateaus and clamp steep tangents.
		for (let i = 0; i < n - 1; i++) {
			if (delta[i] === 0) {
				m[i] = 0;
				m[i + 1] = 0;
				continue;
			}
			const a = m[i] / delta[i];
			const b = m[i + 1] / delta[i];
			const s = a * a + b * b;
			if (s > 9) {
				const tau = 3 / Math.sqrt(s);
				m[i] = tau * a * delta[i];
				m[i + 1] = tau * b * delta[i];
			}
		}
		// Emit a cubic Bézier per segment with control points a third of the way in.
		let d = `M${pts[0].cx.toFixed(1)} ${pts[0].cy.toFixed(1)}`;
		for (let i = 0; i < n - 1; i++) {
			const dx = (pts[i + 1].cx - pts[i].cx) / 3;
			const c1x = pts[i].cx + dx;
			const c1y = pts[i].cy + m[i] * dx;
			const c2x = pts[i + 1].cx - dx;
			const c2y = pts[i + 1].cy - m[i + 1] * dx;
			d += ` C${c1x.toFixed(1)} ${c1y.toFixed(1)} ${c2x.toFixed(1)} ${c2y.toFixed(1)} ${pts[
				i + 1
			].cx.toFixed(1)} ${pts[i + 1].cy.toFixed(1)}`;
		}
		return d;
	}

	const linePath = $derived(smoothPath(points));
	const midY = yAt(50);

	function dotColor(result: App.MatchResult['result']): string {
		if (result === 'win') return 'var(--color-positive)';
		if (result === 'loss') return 'var(--color-destructive)';
		return 'var(--color-warning)';
	}
</script>

{#if bars.length > 0}
	<div data-slot="form-strip" class="bg-surface-raised space-y-3 rounded-2xl px-4 py-4">
		<svg
			class="block w-full"
			style="aspect-ratio: {W} / {H};"
			viewBox="0 0 {W} {H}"
			role="img"
			aria-label={recordLabel}
		>
			<!-- 50% "even" reference: dots above it are winning matches, below are losses.
			     The label makes the one meaningful gridline explicit without a full axis. -->
			<line
				x1={PAD_LEFT}
				y1={midY}
				x2={W - PAD_RIGHT}
				y2={midY}
				stroke="var(--color-border-strong)"
				stroke-width="1.5"
				stroke-dasharray="4 3"
				vector-effect="non-scaling-stroke"
			/>
			<text
				x={PAD_LEFT - 6}
				y={midY}
				text-anchor="end"
				dominant-baseline="central"
				font-size="9"
				font-weight="700"
				fill="var(--color-text-disabled)"
			>
				50%
			</text>
			{#if linePath}
				<path
					d={linePath}
					fill="none"
					stroke="var(--color-text-secondary)"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					vector-effect="non-scaling-stroke"
				/>
			{/if}
			{#each points as p, i (i)}
				<circle
					cx={p.cx}
					cy={p.cy}
					r="4"
					style="fill: {dotColor(p.result)}"
					stroke="var(--color-surface-raised)"
					stroke-width="1.5"
				>
					<title>{Math.round(p.pct)}%</title>
				</circle>
			{/each}
		</svg>
		<div class="text-text-primary text-[11px] font-bold tracking-[0.08em] uppercase tabular-nums">
			{recordLabel}
		</div>
	</div>
{/if}
