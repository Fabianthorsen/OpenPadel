<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	import { setupI18n } from '$lib/i18n';
	import { isLoading } from 'svelte-i18n';

	setupI18n();

	let { children } = $props();
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>

{#if !$isLoading}
  {@render children()}
{/if}
