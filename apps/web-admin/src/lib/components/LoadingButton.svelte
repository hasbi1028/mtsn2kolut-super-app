<script lang="ts">
	import { Button, type ButtonProps, type ButtonSize, type ButtonVariant } from '$lib/components/ui/button';

	type Props = Omit<ButtonProps, 'children'> & {
		loading?: boolean;
		loadingLabel?: string;
		label?: string;
		variant?: ButtonVariant;
		size?: ButtonSize;
		children?: import('svelte').Snippet;
		onclick?: (event: MouseEvent) => void;
		type?: 'button' | 'submit' | 'reset';
		title?: string;
	};

	let {
		loading = false,
		loadingLabel = 'Memproses...',
		label = '',
		disabled,
		children,
		...restProps
	}: Props = $props();
</script>

<Button {...restProps} disabled={loading || disabled}>
	{#if loading}
		<span class="inline-flex items-center gap-2">
			<span class="size-3.5 animate-spin rounded-full border-2 border-current border-t-transparent"></span>
			{loadingLabel}
		</span>
	{:else if children}
		{@render children()}
	{:else}
		{label}
	{/if}
</Button>

