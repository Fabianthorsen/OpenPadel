<script lang="ts" module>
	import { type VariantProps, tv } from 'tailwind-variants';
	import type { HTMLAttributes } from 'svelte/elements';
	import type { Snippet } from 'svelte';
	import { type WithElementRef } from '$lib/utils.js';

	/**
	 * Drawer content container. Slides in from bottom of screen.
	 * Positioned absolutely over page content with backdrop overlay.
	 *
	 * @example
	 * <Drawer>
	 *   <DrawerTrigger>Open</DrawerTrigger>
	 *   <DrawerContent size="md">
	 *     <DrawerHeader>
	 *       <DrawerTitle>Settings</DrawerTitle>
	 *     </DrawerHeader>
	 *     <DrawerBody>Content</DrawerBody>
	 *   </DrawerContent>
	 * </Drawer>
	 *
	 * @example
	 * <DrawerContent size="lg">
	 *   Large bottom drawer for complex forms
	 * </DrawerContent>
	 */
	export const drawerContentVariants = tv({
		base: 'fixed inset-x-0 bottom-0 z-50 flex flex-col gap-4 border-input bg-background p-4 shadow-lg',
		variants: {
			size: {
				sm: 'max-h-[40vh] md:max-h-[300px]',
				md: 'max-h-[60vh] md:max-h-[480px]',
				lg: 'max-h-[80vh] md:max-h-[640px]'
			}
		},
		defaultVariants: {
			size: 'md'
		}
	});

	export type DrawerSize = VariantProps<typeof drawerContentVariants>['size'];

	/**
	 * Drawer content props. Main container for drawer content sliding up from bottom.
	 *
	 * Size variants control maximum height:
	 * - sm: 40vh on mobile, 300px on desktop (compact)
	 * - md: 60vh on mobile, 480px on desktop (standard)
	 * - lg: 80vh on mobile, 640px on desktop (tall)
	 */
	export interface DrawerContentProps extends WithElementRef<HTMLAttributes<HTMLDivElement>> {
		/** size: sm | md | lg. Controls drawer max-height. */
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
		size = 'md',
		children,
		id,
		...restProps
	}: DrawerContentProps & { id?: string } = $props();
</script>

<DialogPrimitive.Portal>
	<DialogPrimitive.Overlay
		data-slot="drawer-overlay"
		class="fixed inset-0 z-40 bg-black/40 transition-opacity duration-200 ease-out data-[state=closed]:opacity-0 data-[state=open]:opacity-100"
		aria-hidden="true"
	/>
	<DialogPrimitive.Content
		bind:ref
		data-slot="drawer-content"
		{id}
		aria-modal="true"
		class={cn(drawerContentVariants({ size }), className)}
		{...restProps}
	>
		{@render children?.()}
	</DialogPrimitive.Content>
</DialogPrimitive.Portal>

<style>
	/* Open state: slide-up animation */
	:global([data-slot='drawer-content'][data-state='open']) {
		animation: slide-up 0.3s cubic-bezier(0.4, 0, 0.2, 1) forwards;
	}

	/* Closed state: slide-down transition */
	:global([data-slot='drawer-content'][data-state='closed']) {
		animation: slide-down 0.3s cubic-bezier(0.4, 0, 0.2, 1) forwards;
	}

	@keyframes slide-up {
		from {
			transform: translateY(100%);
		}
		to {
			transform: translateY(0);
		}
	}

	@keyframes slide-down {
		from {
			transform: translateY(0);
		}
		to {
			transform: translateY(100%);
		}
	}
</style>
