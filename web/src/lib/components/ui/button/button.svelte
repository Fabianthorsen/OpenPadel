<script lang="ts" module>
	import { cn, type WithElementRef } from '$lib/utils.js';
	import type { HTMLAnchorAttributes, HTMLButtonAttributes } from 'svelte/elements';
	import { type VariantProps, tv } from 'tailwind-variants';

	/**
	 * Button component for user interactions.
	 * Renders as `<button>` by default or `<a>` when `href` is provided.
	 *
	 * @example
	 * <Button>Click me</Button>
	 *
	 * @example
	 * <Button variant="destructive" disabled>Disabled</Button>
	 *
	 * @example
	 * <Button href="/path">Link button</Button>
	 *
	 * @example
	 * <Button variant="ghost" size="icon">
	 *   <Icon name="menu" />
	 * </Button>
	 */
	export const buttonVariants = tv({
		base: "focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive dark:aria-invalid:border-destructive/50 rounded-lg border border-transparent bg-clip-padding text-sm font-medium focus-visible:ring-3 active:not-aria-[haspopup]:translate-y-px aria-invalid:ring-3 [&_svg:not([class*='size-'])]:size-4 group/button inline-flex shrink-0 items-center justify-center whitespace-nowrap transition-all outline-none select-none disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:shrink-0",
		variants: {
			variant: {
				default: 'bg-primary text-primary-foreground [a]:hover:bg-primary/80',
				outline:
					'border-border bg-background hover:bg-muted hover:text-foreground dark:bg-input/30 dark:border-input dark:hover:bg-input/50 aria-expanded:bg-muted aria-expanded:text-foreground',
				secondary:
					'bg-secondary text-secondary-foreground hover:bg-secondary/80 aria-expanded:bg-secondary aria-expanded:text-secondary-foreground',
				ghost:
					'hover:bg-muted hover:text-foreground dark:hover:bg-muted/50 aria-expanded:bg-muted aria-expanded:text-foreground',
				destructive:
					'bg-destructive/10 hover:bg-destructive/20 focus-visible:ring-destructive/20 dark:focus-visible:ring-destructive/40 dark:bg-destructive/20 text-destructive focus-visible:border-destructive/40 dark:hover:bg-destructive/30',
				link: 'text-primary underline-offset-4 hover:underline'
			},
			size: {
				default:
					'h-8 gap-1.5 px-2.5 has-data-[icon=inline-end]:pr-2 has-data-[icon=inline-start]:pl-2',
				xs: "h-6 gap-1 rounded-[min(var(--radius-md),10px)] px-2 text-xs in-data-[slot=button-group]:rounded-lg has-data-[icon=inline-end]:pr-1.5 has-data-[icon=inline-start]:pl-1.5 [&_svg:not([class*='size-'])]:size-3",
				sm: "h-7 gap-1 rounded-[min(var(--radius-md),12px)] px-2.5 text-[0.8rem] in-data-[slot=button-group]:rounded-lg has-data-[icon=inline-end]:pr-1.5 has-data-[icon=inline-start]:pl-1.5 [&_svg:not([class*='size-'])]:size-3.5",
				lg: 'h-9 gap-1.5 px-2.5 has-data-[icon=inline-end]:pr-2 has-data-[icon=inline-start]:pl-2',
				icon: 'size-8',
				'icon-xs':
					"size-6 rounded-[min(var(--radius-md),10px)] in-data-[slot=button-group]:rounded-lg [&_svg:not([class*='size-'])]:size-3",
				'icon-sm':
					'size-7 rounded-[min(var(--radius-md),12px)] in-data-[slot=button-group]:rounded-lg',
				'icon-lg': 'size-9'
			}
		},
		defaultVariants: {
			variant: 'default',
			size: 'default'
		}
	});

	export type ButtonVariant = VariantProps<typeof buttonVariants>['variant'];
	export type ButtonSize = VariantProps<typeof buttonVariants>['size'];

	/**
	 * Button component props.
	 *
	 * Pass `href` to render as `<a>` (link button), omit for `<button>` element.
	 * When `href` is set and `disabled=true`, the link is removed (aria-disabled applied instead).
	 */
	export type ButtonProps = WithElementRef<HTMLButtonAttributes> &
		WithElementRef<HTMLAnchorAttributes> & {
			/**
			 * Visual style: primary | outline | secondary | ghost | destructive | link.
			 * See tokens.colors for palette.
			 * - primary: main CTA, preferred action (default)
			 * - outline: secondary action with border (desktop-friendly, less prominent)
			 * - secondary: supporting action, alternative flow
			 * - ghost: low-emphasis action, deemphasized (works on any background)
			 * - destructive: destructive action, demands attention
			 * - link: text-only link style, underline on hover
			 */
			variant?: ButtonVariant;

			/**
			 * Size: xs | sm | default | lg | icon | icon-xs | icon-sm | icon-lg.
			 * - xs: 6px height, extra small (compact UI)
			 * - sm: 7px height, small (tight spacing)
			 * - default: 8px height, standard (most use cases)
			 * - lg: 9px height, large (prominent)
			 * - icon: 8px icon button (rounded, no padding)
			 * - icon-xs: 6px icon button
			 * - icon-sm: 7px icon button
			 * - icon-lg: 9px icon button
			 */
			size?: ButtonSize;

			/** Disabled state: button/link cannot interact, appears inactive. */
			disabled?: boolean;

			/** Expanded state: true when associated content is visible (dropdown, menu). */
			ariaExpanded?: boolean;

			/** Invalid state: indicates aria-invalid for error styling. */
			ariaInvalid?: boolean;

			/** Pressed state: true when toggle button is active. */
			ariaPressed?: boolean;

			/** href: render as link (`<a>`). Omit for button element (`<button>`). */
			href?: string;
		};
</script>

<script lang="ts">
	let {
		class: className,
		variant = 'default',
		size = 'default',
		ref = $bindable(null),
		href = undefined,
		type = 'button',
		disabled,
		children,
		...restProps
	}: ButtonProps = $props();
</script>

{#if href}
	<a
		bind:this={ref}
		data-slot="button"
		class={cn(buttonVariants({ variant, size }), className)}
		href={disabled ? undefined : href}
		aria-disabled={disabled}
		role={disabled ? 'link' : undefined}
		tabindex={disabled ? -1 : undefined}
		{...restProps}
	>
		{@render children?.()}
	</a>
{:else}
	<button
		bind:this={ref}
		data-slot="button"
		class={cn(buttonVariants({ variant, size }), className)}
		{type}
		{disabled}
		{...restProps}
	>
		{@render children?.()}
	</button>
{/if}
