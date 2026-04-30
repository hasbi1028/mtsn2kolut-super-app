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
		success: 'border-emerald-200 bg-emerald-50 text-emerald-900',
		error: 'border-rose-200 bg-rose-50 text-rose-900',
		warning: 'border-amber-200 bg-amber-50 text-amber-900',
		info: 'border-sky-200 bg-sky-50 text-sky-900',
	};

	const messageClass: Record<Tone, string> = {
		success: 'text-emerald-800',
		error: 'text-rose-700',
		warning: 'text-amber-800',
		info: 'text-sky-800',
	};
</script>

<div class={`rounded-2xl border ${toneClass[tone]} ${compact ? 'px-4 py-4' : 'px-5 py-5'}`}>
	<p class="text-sm font-semibold">{title}</p>
	<p class={`mt-1 text-sm leading-6 ${messageClass[tone]}`}>{message}</p>
</div>
