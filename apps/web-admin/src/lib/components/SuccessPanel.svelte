<script lang="ts">
	type Props = {
		title?: string;
		message: string;
		compact?: boolean;
	};

	let {
		title = 'Perubahan Berhasil Disimpan',
		message,
		compact = false,
	}: Props = $props();
</script>

<div class={`rounded-2xl border border-primary/20 bg-primary/10 ${compact ? 'px-4 py-4' : 'px-5 py-5'}`}>
	<p class="text-sm font-semibold text-primary">{title}</p>
	<p class="mt-1 text-sm leading-6 text-primary">{message}</p>
</div>
