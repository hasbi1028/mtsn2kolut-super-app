<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils';

	type Variant = 'default' | 'destructive' | 'outline' | 'secondary' | 'ghost' | 'link';
	type Size = 'default' | 'xs' | 'sm' | 'lg' | 'icon';

	let {
		variant = 'default',
		size = 'default',
		class: className,
		disabled = false,
		children,
		href,
		...restProps
	}: {
		variant?: Variant;
		size?: Size;
		class?: string;
		disabled?: boolean;
		children?: Snippet;
		href?: string;
		[key: string]: unknown;
	} = $props();

	// Tailwind utility classes — no DaisyUI dependency
	const variantClasses: Record<Variant, string> = {
		default:
			'bg-primary text-primary-foreground border border-primary hover:bg-primary/90 shadow-sm',
		destructive:
			'bg-destructive text-destructive-foreground border border-destructive hover:bg-destructive/90 shadow-sm',
		outline:
			'border border-input bg-background text-foreground hover:bg-accent hover:text-accent-foreground',
		secondary:
			'bg-secondary text-secondary-foreground border border-secondary hover:bg-secondary/80',
		ghost:
			'text-foreground hover:bg-accent hover:text-accent-foreground border border-transparent',
		link:
			'text-primary underline-offset-4 hover:underline border border-transparent',
	};

	const sizeClasses: Record<Size, string> = {
		default: 'h-9 px-4 text-sm',
		xs: 'h-7 px-2 text-xs',
		sm: 'h-8 px-3 text-xs',
		lg: 'h-11 px-6 text-base',
		icon: 'h-9 w-9',
	};
</script>

{#if href}
	<a
		{href}
		class={cn(
			'inline-flex items-center justify-center gap-2 rounded-lg font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring cursor-pointer disabled:pointer-events-none disabled:opacity-40',
			variantClasses[variant],
			sizeClasses[size],
			className
		)}
		aria-disabled={disabled || undefined}
		tabindex={disabled ? -1 : undefined}
		{...restProps}
	>
		{@render children?.()}
	</a>
{:else}
	<button
		class={cn(
			'inline-flex items-center justify-center gap-2 rounded-lg font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring cursor-pointer disabled:pointer-events-none disabled:opacity-40',
			variantClasses[variant],
			sizeClasses[size],
			className
		)}
		{disabled}
		{...restProps}
	>
		{@render children?.()}
	</button>
{/if}
