<script lang="ts">
	import EyeIcon from '@lucide/svelte/icons/eye';
	import EyeOffIcon from '@lucide/svelte/icons/eye-off';
	import type { HTMLInputAttributes } from 'svelte/elements';
	import { cn } from '$lib/utils';
	import { Input } from '$lib/components/ui/input';

	type Props = Omit<HTMLInputAttributes, 'type' | 'files'> & {
		value?: string;
	};

	let {
		value = $bindable(''),
		class: className,
		disabled,
		'aria-label': ariaLabel,
		...restProps
	}: Props = $props();

	let visible = $state(false);
	const toggleLabel = $derived(visible ? 'Sembunyikan password' : 'Lihat password');
</script>

<div class="relative">
	<Input
		{...restProps}
		type={visible ? 'text' : 'password'}
		bind:value
		{disabled}
		aria-label={ariaLabel}
		class={cn('pr-10', className)}
	/>
	<button
		type="button"
		class="absolute inset-y-0 right-0 inline-flex w-10 items-center justify-center rounded-r-lg text-muted-foreground transition hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50"
		onclick={() => (visible = !visible)}
		disabled={disabled}
		aria-label={toggleLabel}
		aria-pressed={visible}
		title={toggleLabel}
	>
		{#if visible}
			<EyeOffIcon class="h-4 w-4" aria-hidden="true" />
		{:else}
			<EyeIcon class="h-4 w-4" aria-hidden="true" />
		{/if}
	</button>
</div>
