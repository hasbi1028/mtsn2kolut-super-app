<script lang="ts">
	type Tone = 'success' | 'error' | 'warning' | 'info';

	type Props = {
		title: string;
		message: string;
		tone?: Tone;
		compact?: boolean;
	};

	let {
		title,
		message,
		tone = 'info',
		compact = false,
	}: Props = $props();

	const toneClass: Record<Tone, string> = {
		success: 'border-primary/20 bg-primary/10 text-primary',
		error: 'border-destructive/30 bg-destructive/10 text-destructive',
		warning: 'border-warning/30 bg-warning/10 text-warning',
		info: 'border-accent bg-accent/60 text-accent-foreground',
	};

	const messageClass: Record<Tone, string> = {
		success: 'text-primary',
		error: 'text-destructive',
		warning: 'text-warning',
		info: 'text-accent-foreground',
	};
</script>

<div class={`rounded-2xl border ${toneClass[tone]} ${compact ? 'px-4 py-4' : 'px-5 py-5'}`}>
	<p class="text-sm font-semibold">{title}</p>
	<p class={`mt-1 text-sm leading-6 ${messageClass[tone]}`}>{message}</p>
</div>
