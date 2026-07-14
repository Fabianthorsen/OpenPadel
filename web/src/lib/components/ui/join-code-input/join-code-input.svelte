<script lang="ts" module>
	/** Number of characters in a session join code. */
	const LENGTH = 4;

	const sanitize = (s: string) => s.toUpperCase().replace(/[^A-Z0-9]/g, '');
</script>

<script lang="ts">
	import { cn } from '$lib/utils.js';

	/**
	 * Session join-code entry — OTP-style boxes with auto-advance, backspace-to-previous,
	 * and paste. The single shared join-code control (Home + Profile).
	 *
	 * @example
	 * <JoinCodeInput onComplete={(code) => goto(`/s/${code}`)} />
	 */
	let {
		value = $bindable(''),
		onComplete,
		class: className
	}: {
		/** value: reflects the entered code; updates as boxes are filled. */
		value?: string;
		/** Called with the full code once every box is filled. */
		onComplete?: (code: string) => void;
		class?: string;
	} = $props();

	let chars = $state<string[]>(Array(LENGTH).fill(''));
	let inputs = $state<HTMLInputElement[]>([]);

	function sync() {
		value = chars.join('');
		if (value.length === LENGTH && chars.every((c) => c)) onComplete?.(value);
	}

	function onInput(i: number, e: Event) {
		const raw = sanitize((e.currentTarget as HTMLInputElement).value);
		chars[i] = raw.slice(-1);
		if (raw && i < LENGTH - 1) inputs[i + 1]?.focus();
		sync();
	}

	function onKeydown(i: number, e: KeyboardEvent) {
		if (e.key === 'Backspace' && !chars[i] && i > 0) {
			chars[i - 1] = '';
			inputs[i - 1]?.focus();
			sync();
		}
	}

	function onPaste(e: ClipboardEvent) {
		const text = sanitize(e.clipboardData?.getData('text') ?? '');
		if (text.length >= LENGTH) {
			e.preventDefault();
			chars = text.slice(0, LENGTH).split('');
			inputs[LENGTH - 1]?.focus();
			sync();
		}
	}
</script>

<div
	class={cn('flex justify-center gap-2', className)}
	onpaste={onPaste}
	data-slot="join-code-input"
>
	{#each chars as ch, i (i)}
		<input
			bind:this={inputs[i]}
			value={ch}
			oninput={(e) => onInput(i, e)}
			onkeydown={(e) => onKeydown(i, e)}
			maxlength={1}
			autocomplete="off"
			autocorrect="off"
			autocapitalize="characters"
			spellcheck={false}
			aria-label={`Join code character ${i + 1} of ${LENGTH}`}
			class="bg-surface-raised text-text-primary focus:ring-primary size-12 rounded-xl text-center font-mono text-lg font-bold uppercase transition-shadow outline-none focus:ring-2"
		/>
	{/each}
</div>
