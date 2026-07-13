import type { HTMLAttributes } from 'svelte/elements';
import type { WithElementRef } from '$lib/utils.js';

/**
 * Base props for drawer sub-components that wrap Bits UI Dialog primitives.
 * Bits UI expects `id?: string`, but Svelte's HTMLAttributes types `id` as
 * `string | null`; this narrows it so `{...restProps}` stays assignable.
 */
export type DrawerPrimitiveProps<T extends HTMLElement = HTMLElement> = WithElementRef<
	Omit<HTMLAttributes<T>, 'id'>
> & { id?: string };
