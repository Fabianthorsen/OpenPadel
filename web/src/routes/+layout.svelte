<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	import { setupI18n } from '$lib/i18n';
	import { isLoading } from 'svelte-i18n';
	import { auth } from '$lib/auth.svelte';
	import { onMount } from 'svelte';
	import { Toaster } from 'svelte-sonner';

	setupI18n();
	onMount(() => auth.init());

	let { children } = $props();
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>

{#if !$isLoading}
  {@render children()}
  <Toaster position="bottom-center" />
{/if}
