<script lang="ts">
	import { cn } from '$lib/utils';

	let {
		class: className = '',
		rounded = 'rounded-xl'
	}: {
		class?: string;
		rounded?: string;
	} = $props();
</script>

<div
	class={cn(
		'animate-pulse bg-gradient-to-r from-muted via-accent/50 to-muted bg-[length:200%_100%]',
		rounded,
		className
	)}
	aria-hidden="true"
></div>
