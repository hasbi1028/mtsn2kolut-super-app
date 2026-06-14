<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	type Variant = 'default' | 'secondary' | 'destructive' | 'outline';

	let {
		variant = 'default',
		class: className,
		children,
		...restProps
	}: {
		variant?: Variant;
		class?: string;
		children?: Snippet;
		[key: string]: unknown;
	} = $props();

	const variantClasses: Record<Variant, string> = {
		default: 'preset-tonal-primary',
		secondary: 'preset-tonal',
		destructive: 'preset-tonal-error',
		outline: 'preset-outlined'
	};
</script>

<div class={cn('badge inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors', variantClasses[variant], className)} {...restProps}>
	{@render children?.()}
</div>
