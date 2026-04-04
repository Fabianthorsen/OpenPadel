import { browser } from '$app/environment';
import { init, register, locale } from 'svelte-i18n';

register('en', () => import('./en.json'));
register('no', () => import('./no.json'));

export function setupI18n() {
  const savedLocale = browser ? (localStorage.getItem('locale') ?? navigator.language.split('-')[0]) : 'en';
  const chosen = ['en', 'no'].includes(savedLocale) ? savedLocale : 'en';

  init({
    fallbackLocale: 'en',
    initialLocale: chosen,
  });
}

export function setLocale(lang: string) {
  locale.set(lang);
  if (browser) localStorage.setItem('locale', lang);
}

export { locale };
