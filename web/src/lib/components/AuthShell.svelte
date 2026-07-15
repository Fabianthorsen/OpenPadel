<script lang="ts">
	import { _ } from 'svelte-i18n';

	interface Props {
		subtitle: string;
		children?: any;
		backHref?: string;
		class?: string;
	}

	let { subtitle, children, backHref = '/', class: className = '' }: Props = $props();
</script>

<main class={`pt-safe-page flex min-h-svh flex-col items-center px-6 pb-12 ${className}`}>
	<div class="flex w-full max-w-sm flex-1 flex-col justify-center space-y-8">
		<!-- Brand -->
		<div class="space-y-1">
			<h1 class="text-primary text-[28px] font-[800]">OpenPadel</h1>
			<p class="text-text-secondary">{subtitle}</p>
		</div>

		{@render children?.()}

		<!-- Back home link -->
		<div class="flex justify-center">
			<a href={backHref} class="text-text-disabled hover:text-text-secondary text-xs">
				← {$_('auth_back_home')}
			</a>
		</div>
	</div>
</main>
