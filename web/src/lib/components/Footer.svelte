<script lang="ts">
  import { _ } from 'svelte-i18n';
  import LocaleSwitcher from './LocaleSwitcher.svelte';

  let { showLocale = true }: { showLocale?: boolean } = $props();
</script>

<footer class="flex items-center justify-center gap-3 pb-6 pt-8 text-[11px] text-[var(--text-disabled)]">
  {#if showLocale}
    <LocaleSwitcher />
    <span>·</span>
  {/if}
  <a href="https://github.com/fabianthorsen/nottennis" target="_blank" rel="noopener noreferrer" class="flex items-center opacity-40 hover:opacity-70 transition-opacity">
    <img src="/github-logo-clearspace.svg" alt="GitHub" class="h-[14px] w-auto" />
  </a>
  <span>·</span>
  <a href="https://github.com/fabianthorsen/nottennis/issues/new" target="_blank" rel="noopener noreferrer" class="hover:text-[var(--text-secondary)] transition-colors">
    {$_('footer_feedback')}
  </a>
</footer>
