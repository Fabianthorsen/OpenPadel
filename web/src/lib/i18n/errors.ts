import { get } from 'svelte/store';
import { _ } from 'svelte-i18n';

export function translateApiError(code: string, params?: Record<string, string | number | boolean | null | undefined>): string {
  const t = get(_);
  const key = `api_error_${code}`;
  const msg = params ? t(key, { values: params }) : t(key);
  // svelte-i18n returns the key itself when no translation is found
  return msg === key ? t('api_error_server_error') : msg;
}
