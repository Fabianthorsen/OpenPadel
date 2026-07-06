<script lang="ts">
	import { _ } from 'svelte-i18n';
	import LocaleSwitcher from './LocaleSwitcher.svelte';

	let { showLocale = true }: { showLocale?: boolean } = $props();
</script>

<footer class="text-text-disabled flex items-center justify-center gap-3 pt-8 pb-6 text-[11px]">
	{#if showLocale}
		<LocaleSwitcher />
		<span>·</span>
	{/if}
	<a
		href="https://github.com/fabianthorsen/openpadel"
		target="_blank"
		rel="noopener noreferrer"
		class="flex items-center opacity-40 transition-opacity hover:opacity-70"
	>
		<img src="/github-logo-clearspace.svg" alt="GitHub" class="h-[14px] w-auto" />
	</a>
	<span>·</span>
	<a
		href="https://github.com/fabianthorsen/openpadel/issues/new"
		target="_blank"
		rel="noopener noreferrer"
		class="hover:text-text-secondary transition-colors"
	>
		{$_('footer_feedback')}
	</a>
</footer>
