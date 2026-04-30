<script lang="ts">
	let { status, error } = $props<{
		status: number;
		error: { message?: string };
	}>();

	let title = $derived.by(() => status === 404
		? 'Halaman tidak ditemukan'
		: status === 403
			? 'Akses ditolak'
			: 'Terjadi kesalahan');

	let description = $derived.by(() => status === 404
		? 'Halaman yang Anda cari tidak tersedia atau sudah dipindahkan.'
		: status === 403
			? 'Anda tidak memiliki izin untuk membuka halaman ini.'
			: error?.message || 'Sistem sedang mengalami kendala. Coba muat ulang beberapa saat lagi.');
</script>

<svelte:head><title>{status} — {title}</title></svelte:head>

<div class="flex min-h-[70vh] items-center justify-center">
	<div class="w-full max-w-2xl rounded-[2rem] border border-slate-200 bg-white px-6 py-10 text-center shadow-sm sm:px-10">
		<p class="text-xs font-semibold uppercase tracking-[0.24em] text-emerald-700">Error {status}</p>
		<h1 class="mt-4 text-4xl font-bold text-slate-900 sm:text-5xl">{title}</h1>
		<p class="mx-auto mt-4 max-w-xl text-base leading-8 text-slate-600">{description}</p>
		<div class="mt-8 flex flex-wrap justify-center gap-3">
			<a href="/" class="rounded-full bg-[oklch(0.38_0.13_145)] px-5 py-3 text-sm font-semibold text-white shadow-sm hover:brightness-105">
				Kembali ke Beranda
			</a>
			<a href="/login" class="rounded-full border border-slate-200 px-5 py-3 text-sm font-semibold text-slate-700 hover:bg-slate-50">
				Login Admin
			</a>
		</div>
	</div>
</div>
