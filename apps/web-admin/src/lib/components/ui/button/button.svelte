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
		default: 'btn btn-primary',
		destructive: 'btn btn-error',
		outline: 'btn btn-outline',
		secondary: 'btn btn-soft',
		ghost: 'btn btn-ghost',
		link: 'btn btn-link'
	};

	const sizeClasses: Record<Size, string> = {
		default: '',
		xs: 'btn-xs',
		sm: 'btn-sm',
		lg: 'btn-lg',
		icon: 'btn-square'
	};
</script>

{#if href}
	<a
		{href}
		class={cn('btn', variantClasses[variant], sizeClasses[size], className)}
		aria-disabled={disabled || undefined}
		tabindex={disabled ? -1 : undefined}
		{...restProps}
	>
		{@render children?.()}
	</a>
{:else}
	<button
		class={cn('btn', variantClasses[variant], sizeClasses[size], className)}
		{disabled}
		{...restProps}
	>
		{@render children?.()}
	</button>
{/if}