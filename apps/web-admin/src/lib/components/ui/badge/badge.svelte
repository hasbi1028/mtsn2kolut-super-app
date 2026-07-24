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
		default: 'bg-primary/10 text-primary border border-primary/20',
		secondary: 'bg-secondary/30 text-secondary-foreground border border-secondary/30',
		destructive: 'bg-destructive/10 text-destructive border border-destructive/20',
		outline: 'bg-transparent text-foreground border border-border'
	};
</script>

<span class={cn(
	'inline-flex items-center rounded-full px-2.5 py-0.5 text-[10px] font-semibold leading-none uppercase tracking-wider whitespace-nowrap',
	variantClasses[variant],
	className
)} {...restProps}>
	{@render children?.()}
</span>
