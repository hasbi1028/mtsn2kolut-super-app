<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { cn } from '$lib/utils';

	type Blocker = {
		label: string;
		description?: string;
		href?: string;
		actionLabel?: string;
		tone?: 'warning' | 'danger' | 'info';
	};

	let {
		title = 'Perlu Tindakan',
		blockers = [],
		class: className = '',
	}: {
		title?: string;
		blockers?: Blocker[];
		class?: string;
	} = $props();

	function toneClass(tone: Blocker['tone']) {
		if (tone === 'danger') return 'border-destructive/30 bg-destructive/10 text-destructive';
		if (tone === 'info') return 'border-border bg-muted/50 text-muted-foreground';
		return 'border-warning/30 bg-warning/10 text-warning';
	}
</script>

{#if blockers.length > 0}
	<section class={cn('rounded-lg border border-warning/30 bg-warning/10 p-4 shadow-sm', className)} aria-labelledby="blocker-panel-title">
		<h2 id="blocker-panel-title" class="text-sm font-semibold text-foreground">{title}</h2>
		<div class="mt-3 grid gap-2">
			{#each blockers as blocker (blocker.label)}
				<div class={cn('rounded-md border p-3', toneClass(blocker.tone))}>
					<div class="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
						<div class="min-w-0">
							<p class="text-sm font-semibold">{blocker.label}</p>
							{#if blocker.description}
								<p class="mt-1 text-xs leading-5 opacity-80">{blocker.description}</p>
							{/if}
						</div>
						{#if blocker.href}
							<Button href={blocker.href} variant="outline" size="sm" class="bg-card">{blocker.actionLabel ?? 'Buka'}</Button>
						{/if}
					</div>
				</div>
			{/each}
		</div>
	</section>
{/if}
