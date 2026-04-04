import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import type { Component, ComponentProps } from 'svelte';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export type WithElementRef<T, E extends Element = HTMLElement> = T & {
  ref?: E | null;
};

export type WithoutChildren<T> = Omit<T, 'children'>;
export type WithoutChildrenOrChild<T> = Omit<T, 'children' | 'child'>;
export type AsChild<T extends Component> = { asChild?: boolean; child?: Component<{ props: ComponentProps<T> }> };
