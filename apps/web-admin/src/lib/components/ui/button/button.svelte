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

	const variantClasses: Record<Variant, string> = {
		default: 'preset-filled-primary-500 text-on-primary hover:opacity-90',
		destructive: 'preset-filled-error-500 text-on-error hover:opacity-90',
		outline: 'preset-tonal text-on-surface hover:opacity-80',
		secondary: 'preset-tonal-secondary text-on-surface hover:opacity-80',
		ghost: 'hover:bg-muted text-on-surface',
		link: 'text-primary underline-offset-4 hover:underline'
	};

	const sizeClasses: Record<Size, string> = {
		default: 'h-10 px-4 py-2',
		xs: 'h-7 rounded-md px-2 text-xs',
		sm: 'h-9 rounded-md px-3 text-sm',
		lg: 'h-11 rounded-md px-8',
		icon: 'h-10 w-10'
	};
</script>

{#if href}
	<a
		{href}
		class={cn('btn inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2', variantClasses[variant], sizeClasses[size], className, disabled && 'pointer-events-none opacity-50')}
		aria-disabled={disabled || undefined}
		tabindex={disabled ? -1 : undefined}
		{...restProps}
	>
		{@render children?.()}
	</a>
{:else}
	<button
		class={cn('btn inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2', variantClasses[variant], sizeClasses[size], className)}
		{disabled}
		{...restProps}
	>
		{@render children?.()}
	</button>
{/if}
