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
		default: 'badge',
		secondary: 'badge badge-soft',
		destructive: 'badge badge-error',
		outline: 'badge badge-outline'
	};
</script>

<div class={cn('badge', variantClasses[variant], className)} {...restProps}>
	{@render children?.()}
</div>
