<!-- PROTOTYPE — wipe me. Floating variant switcher (dev only). -->
<script lang="ts">
	import { goto } from '$app/navigation';

	let {
		variants,
		current,
		admin
	}: { variants: { key: string; name: string }[]; current: string; admin: boolean } = $props();

	const idx = $derived(
		Math.max(
			0,
			variants.findIndex((v) => v.key === current)
		)
	);

	function go(key: string, adm: boolean) {
		goto(`?variant=${key}&admin=${adm ? '1' : '0'}`, {
			replaceState: true,
			keepFocus: true,
			noScroll: true
		});
	}
	function prev() {
		go(variants[(idx - 1 + variants.length) % variants.length].key, admin);
	}
	function next() {
		go(variants[(idx + 1) % variants.length].key, admin);
	}
	function onKey(e: KeyboardEvent) {
		const t = e.target as HTMLElement | null;
		if (t && (t.tagName === 'INPUT' || t.tagName === 'TEXTAREA' || t.isContentEditable)) return;
		if (e.key === 'ArrowLeft') prev();
		if (e.key === 'ArrowRight') next();
	}
</script>

<svelte:window onkeydown={onKey} />

{#if import.meta.env.DEV}
	<div
		class="fixed bottom-20 left-1/2 z-[100] flex -translate-x-1/2 items-center gap-3 rounded-full border border-white/10 bg-black/85 px-3 py-2 text-white shadow-xl backdrop-blur"
	>
		<button
			onclick={prev}
			aria-label="Previous variant"
			class="flex h-7 w-7 items-center justify-center rounded-full hover:bg-white/15">←</button
		>
		<span class="text-xs font-semibold whitespace-nowrap">{current} — {variants[idx].name}</span>
		<button
			onclick={next}
			aria-label="Next variant"
			class="flex h-7 w-7 items-center justify-center rounded-full hover:bg-white/15">→</button
		>
		<span class="mx-1 h-4 w-px bg-white/20"></span>
		<label class="flex cursor-pointer items-center gap-1.5 text-xs">
			<input type="checkbox" checked={admin} onchange={() => go(current, !admin)} />
			Admin
		</label>
	</div>
{/if}
