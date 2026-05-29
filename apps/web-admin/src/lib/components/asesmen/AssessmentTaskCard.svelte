<script lang="ts">
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';

	type Tone = 'default' | 'primary' | 'warning' | 'success';

	let {
		code,
		title,
		description,
		meta = '',
		href = '',
		cta = '',
		tone = 'default'
	}: {
		code: string;
		title: string;
		description: string;
		meta?: string;
		href?: string;
		cta?: string;
		tone?: Tone;
	} = $props();

	function toneClass(value: Tone) {
		if (value === 'primary') return 'border-primary/20 bg-primary/5';
		if (value === 'warning') return 'border-warning/30 bg-warning/10';
		if (value === 'success') return 'border-success/20 bg-success/10';
		return 'border-border bg-card';
	}
</script>

<article class={`rounded-lg border p-4 shadow-sm ${toneClass(tone)}`}>
	<div class="flex items-start gap-3">
		<Badge class="shrink-0 border-primary/20 bg-primary/10 text-primary" variant="outline">{code}</Badge>
		<div class="min-w-0">
			<h2 class="text-base font-semibold text-foreground">{title}</h2>
			<p class="mt-1 text-sm leading-6 text-muted-foreground">{description}</p>
			{#if meta}
				<p class="mt-2 text-xs font-medium text-muted-foreground">{meta}</p>
			{/if}
		</div>
	</div>
	{#if href && cta}
		<Button href={href} variant="outline" size="sm" class="mt-4 border-primary/20 text-primary hover:bg-primary/10">
			{cta}
		</Button>
	{/if}
</article>
