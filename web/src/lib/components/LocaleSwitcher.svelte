<script lang="ts">
  import { locale, setLocale } from '$lib/i18n';
</script>

<div class="flex items-center gap-2">
  {#each ['en', 'no'] as lang}
    <button
      onclick={() => setLocale(lang)}
      class="text-[11px] font-semibold uppercase tracking-wide transition-colors
        {$locale === lang ? 'text-text-secondary' : 'text-text-disabled hover:text-text-secondary'}"
    >
      {lang}
    </button>
    {#if lang === 'en'}
      <span class="text-border">·</span>
    {/if}
  {/each}
</div>
