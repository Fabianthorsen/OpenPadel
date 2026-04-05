import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import type { Component, ComponentProps } from 'svelte';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

/** Returns "Firstname L." for multi-word names, or just the name if single word. "Fabian Thorsen" → "Fabian T." */
export function shortName(name: string): string {
  const words = name.trim().split(/\s+/).filter(Boolean);
  if (words.length <= 1) return name.trim();
  return `${words[0]} ${words[words.length - 1][0].toUpperCase()}.`;
}

/** Returns up to 2 uppercase initials from a display name. "Fabian Thorsen" → "FT" */
export function initials(name: string): string {
  const words = name.trim().split(/\s+/).filter(Boolean);
  if (words.length === 0) return '?';
  if (words.length === 1) return words[0][0].toUpperCase();
  return (words[0][0] + words[words.length - 1][0]).toUpperCase();
}

export type WithElementRef<T, E extends Element = HTMLElement> = T & {
  ref?: E | null;
};

export type WithoutChildren<T> = Omit<T, 'children'>;
export type WithoutChildrenOrChild<T> = Omit<T, 'children' | 'child'>;
export type AsChild<T extends Component> = { asChild?: boolean; child?: Component<{ props: ComponentProps<T> }> };
