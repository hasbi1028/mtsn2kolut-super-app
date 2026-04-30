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

<div class={`rounded-2xl border border-emerald-200 bg-emerald-50 ${compact ? 'px-4 py-4' : 'px-5 py-5'}`}>
	<p class="text-sm font-semibold text-emerald-900">{title}</p>
	<p class="mt-1 text-sm leading-6 text-emerald-800">{message}</p>
</div>
