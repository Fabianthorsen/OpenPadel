<script lang="ts" module>
	import { type VariantProps, tv } from 'tailwind-variants';
	import type { HTMLAttributes } from 'svelte/elements';
	import type { Snippet } from 'svelte';
	import { type WithElementRef } from '$lib/utils.js';

	/**
	 * Drawer content container. Slides in from left or right edge.
	 * Positioned absolutely over page content with backdrop overlay.
	 *
	 * @example
	 * <Drawer>
	 *   <DrawerTrigger>Open</DrawerTrigger>
	 *   <DrawerContent position="right" size="md">
	 *     <DrawerHeader>
	 *       <DrawerTitle>Settings</DrawerTitle>
	 *     </DrawerHeader>
	 *     <DrawerBody>Content</DrawerBody>
	 *   </DrawerContent>
	 * </Drawer>
	 *
	 * @example
	 * <DrawerContent position="left" size="lg">
	 *   Wide left-side drawer for navigation
	 * </DrawerContent>
	 */
	export const drawerContentVariants = tv({
		base: 'fixed inset-y-0 z-50 gap-4 border-input bg-background p-4 shadow-lg transition-transform md:w-full',
		variants: {
			position: {
				left: 'left-0 border-r animate-slide-in-from-left',
				right: 'right-0 border-l animate-slide-in-from-right'
			},
			size: {
				sm: 'w-[320px]',
				md: 'w-[480px]',
				lg: 'w-[640px]'
			}
		},
		defaultVariants: {
			position: 'left',
			size: 'md'
		}
	});

	export type DrawerPosition = VariantProps<typeof drawerContentVariants>['position'];
	export type DrawerSize = VariantProps<typeof drawerContentVariants>['size'];

	/**
	 * Drawer content props. Main container for drawer content.
	 *
	 * Position variants:
	 * - left: slides from left edge (common for navigation)
	 * - right: slides from right edge (common for settings/filters)
	 *
	 * Size variants:
	 * - sm: 320px (mobile-friendly, tight content)
	 * - md: 480px (standard, balanced)
	 * - lg: 640px (wide, complex forms/lists)
	 */
	export interface DrawerContentProps extends WithElementRef<HTMLAttributes<HTMLDivElement>> {
		/** position: left | right. Determines which edge drawer slides from. */
		position?: DrawerPosition;

		/** size: sm | md | lg. Controls drawer width. */
		size?: DrawerSize;

		/** Child content rendered inside drawer. */
		children?: Snippet;
	}
</script>

<script lang="ts">
	import { Dialog as DialogPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils.js';

	let {
		ref = $bindable(null),
		class: className,
		position = 'left',
		size = 'md',
		children,
		id,
		...restProps
	}: DrawerContentProps & { id?: string } = $props();
</script>

<DialogPrimitive.Portal>
	<DialogPrimitive.Overlay
		data-slot="drawer-overlay"
		class="animate-fade-in fixed inset-0 z-40 bg-black/40"
	/>
	<DialogPrimitive.Content
		bind:ref
		data-slot="drawer-content"
		data-position={position}
		{id}
		class={cn(drawerContentVariants({ position, size }), className)}
		{...restProps}
	>
		{@render children?.()}
	</DialogPrimitive.Content>
</DialogPrimitive.Portal>

<style>
	/* Responsive: full width on mobile, constrained on desktop */
	@media (max-width: 768px) {
		:global([data-slot='drawer-content']) {
			width: 90vw !important;
			max-width: 90vw;
		}
	}

	/* RTL support */
	@media (dir: rtl) {
		:global([data-position='left']) {
			left: auto;
			right: 0;
		}
		:global([data-position='right']) {
			right: auto;
			left: 0;
		}
	}
</style>
