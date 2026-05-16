<script lang="ts">
	import { maintenanceModuleLabel } from '$lib/maintenance/modules';

	let { data } = $props();
	const status = $derived(data.maintenanceStatus);
	const window = $derived(status?.window);
	const modules = $derived((window?.affected_modules ?? []).map(maintenanceModuleLabel).join(', '));

	function formatDate(value?: string) {
		if (!value) return 'Belum ditentukan';
		return new Intl.DateTimeFormat('id-ID', {
			dateStyle: 'full',
			timeStyle: 'short',
			timeZone: 'Asia/Makassar'
		}).format(new Date(value));
	}
</script>

<svelte:head>
	<title>Maintenance Sistem</title>
</svelte:head>

<div class="min-h-screen bg-slate-950 px-4 py-10 text-slate-100 sm:px-6 lg:px-8">
	<div class="mx-auto flex min-h-[80vh] max-w-3xl flex-col justify-center">
		<div class="rounded-3xl border border-white/10 bg-white/10 p-8 shadow-2xl backdrop-blur">
			<div class="mb-6 inline-flex rounded-full border border-amber-300/30 bg-amber-400/10 px-3 py-1 text-sm font-semibold text-amber-100">
				Maintenance Center
			</div>
			<h1 class="text-3xl font-bold tracking-tight sm:text-4xl">
				{window?.title ?? 'Sistem sedang dalam pemeliharaan'}
			</h1>
			<p class="mt-4 text-base leading-7 text-slate-200">
				{window?.message ?? 'Layanan sedang disiapkan kembali. Silakan coba beberapa saat lagi.'}
			</p>

			<div class="mt-8 grid gap-4 sm:grid-cols-2">
				<div class="rounded-2xl border border-white/10 bg-slate-900/70 p-4">
					<p class="text-xs uppercase tracking-wide text-slate-400">Mode</p>
					<p class="mt-1 font-semibold capitalize">{window?.mode?.replace('_', ' ') ?? status?.mode ?? 'maintenance'}</p>
				</div>
				<div class="rounded-2xl border border-white/10 bg-slate-900/70 p-4">
					<p class="text-xs uppercase tracking-wide text-slate-400">Estimasi selesai</p>
					<p class="mt-1 font-semibold">{formatDate(window?.ends_at)}</p>
				</div>
				<div class="rounded-2xl border border-white/10 bg-slate-900/70 p-4 sm:col-span-2">
					<p class="text-xs uppercase tracking-wide text-slate-400">Modul terdampak</p>
					<p class="mt-1 font-semibold">{modules || 'Seluruh sistem'}</p>
				</div>
			</div>

			<p class="mt-8 text-sm text-slate-400">
				Admin yang diberi bypass tetap dapat masuk untuk menyelesaikan maintenance.
			</p>
		</div>
	</div>
</div>
