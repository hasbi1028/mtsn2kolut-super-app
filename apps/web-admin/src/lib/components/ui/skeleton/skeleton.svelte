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
		'animate-pulse bg-gradient-to-r from-slate-200 via-slate-100 to-slate-200 bg-[length:200%_100%]',
		rounded,
		className
	)}
	aria-hidden="true"
></div>
