<script lang="ts">
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';

	type PhaseAction = {
		label: string;
		href: string;
		variant?: 'default' | 'outline' | 'ghost' | 'secondary' | 'link';
	};

	let {
		code,
		title,
		description,
		badge = 'Asesmen',
		context = '',
		primaryAction,
		secondaryActions = []
	}: {
		code: string;
		title: string;
		description: string;
		badge?: string;
		context?: string;
		primaryAction?: PhaseAction;
		secondaryActions?: PhaseAction[];
	} = $props();
</script>

<section class="rounded-xl border border-primary/20 bg-card p-4 shadow-sm">
	<div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_18rem] lg:items-center">
		<div class="min-w-0">
			<div class="mb-2 flex flex-wrap items-center gap-2">
				<Badge class="border-primary/20 bg-primary/10 text-primary" variant="outline">{code}</Badge>
				<Badge class="border-border bg-muted text-muted-foreground" variant="outline">{badge}</Badge>
				{#if context}
					<Badge class="border-border bg-card text-muted-foreground" variant="outline">{context}</Badge>
				{/if}
			</div>
			<h1 class="text-2xl font-semibold tracking-tight text-foreground">{title}</h1>
			<p class="mt-2 max-w-2xl text-sm leading-6 text-muted-foreground">{description}</p>
		</div>
		{#if primaryAction || secondaryActions.length > 0}
			<div class="rounded-lg border border-primary/20 bg-primary/5 p-3">
				<p class="text-xs font-semibold uppercase tracking-[0.18em] text-primary">Aksi utama</p>
				{#if primaryAction}
					<Button href={primaryAction.href} class="mt-3 w-full" variant={primaryAction.variant ?? 'default'}>
						{primaryAction.label}
					</Button>
				{/if}
				{#if secondaryActions.length > 0}
					<div class="mt-2 flex flex-wrap gap-2">
						{#each secondaryActions as action (action.href)}
							<Button href={action.href} size="sm" variant={action.variant ?? 'outline'} class="flex-1">
								{action.label}
							</Button>
						{/each}
					</div>
				{/if}
			</div>
		{/if}
	</div>
</section>
