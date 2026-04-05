<script lang="ts">
  import { locale, setLocale } from '$lib/i18n';
</script>

<div class="flex items-center gap-2">
  {#each ['en', 'no'] as lang}
    <button
      onclick={() => setLocale(lang)}
      class="text-[11px] font-semibold uppercase tracking-wide transition-colors
        {$locale === lang ? 'text-[var(--text-secondary)]' : 'text-[var(--text-disabled)] hover:text-[var(--text-secondary)]'}"
    >
      {lang}
    </button>
    {#if lang === 'en'}
      <span class="text-[var(--border)]">·</span>
    {/if}
  {/each}
</div>
