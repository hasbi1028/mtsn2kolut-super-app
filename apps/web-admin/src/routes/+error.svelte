<script lang="ts">
	import { resolve } from '$app/paths';

	let { status, error } = $props<{
		status: number;
		error: { message?: string };
	}>();

	let title = $derived.by(() => status === 404
		? 'Halaman tidak ditemukan'
		: status === 403
			? 'Akses ditolak'
			: status === 503
				? 'Layanan sementara bermasalah'
				: status === 500
					? 'Halaman sedang bermasalah'
					: 'Terjadi kesalahan');

	let description = $derived.by(() => status === 404
		? 'Halaman yang Anda cari tidak tersedia atau sudah dipindahkan.'
		: status === 403
			? 'Anda tidak memiliki izin untuk membuka halaman ini.'
			: status === 503
				? error?.message || 'Layanan sesi sedang tidak dapat diverifikasi. Coba muat ulang beberapa saat lagi tanpa melakukan login ulang terlebih dahulu.'
				: status === 500
					? 'Halaman ini gagal dimuat karena kendala sistem. Data Bapak/Ibu tetap aman. Silakan coba muat ulang; jika masih terjadi, laporkan halaman ini ke admin.'
					: error?.message || 'Sistem sedang mengalami kendala. Coba muat ulang beberapa saat lagi.');

	let reference = $derived.by(() => `ERR-${status}-${new Date().toISOString().slice(0, 10).replaceAll('-', '')}`);
</script>

<svelte:head><title>{status} — {title}</title></svelte:head>

<div class="flex min-h-[70vh] items-center justify-center">
	<div class="w-full max-w-2xl rounded-[2rem] border border-border bg-card px-6 py-10 text-center shadow-sm sm:px-10">
		<p class="text-xs font-semibold uppercase tracking-[0.24em] text-primary">Kode {status}</p>
		<h1 class="mt-4 text-4xl font-bold text-foreground sm:text-5xl">{title}</h1>
		<p class="mx-auto mt-4 max-w-xl text-base leading-8 text-muted-foreground">{description}</p>
		{#if status >= 500}
			<div class="mx-auto mt-5 max-w-md rounded-2xl border border-warning/30 bg-warning/10 px-4 py-3 text-left text-sm text-warning">
				<p class="font-semibold">Yang bisa dilakukan:</p>
				<ul class="mt-2 list-disc space-y-1 pl-5">
					<li>Tekan <strong>Coba Muat Ulang</strong>.</li>
					<li>Jika sedang mengisi data, jangan ulangi aksi simpan berkali-kali.</li>
					<li>Sampaikan kode referensi <strong>{reference}</strong> ke admin.</li>
				</ul>
			</div>
		{/if}
		<div class="mt-8 flex flex-wrap justify-center gap-3">
			<a href={resolve('/')} class="rounded-full bg-primary px-5 py-3 text-sm font-semibold text-primary-foreground shadow-sm hover:brightness-105">
				Kembali ke Beranda
			</a>
			{#if status >= 500}
				<button type="button" onclick={() => window.location.reload()} class="rounded-full border border-primary/20 px-5 py-3 text-sm font-semibold text-primary hover:bg-primary/10">
					Coba Muat Ulang
				</button>
			{/if}
			<a href={resolve('/login')} class="rounded-full border border-border px-5 py-3 text-sm font-semibold text-foreground hover:bg-muted/50">
				Masuk
			</a>
		</div>
	</div>
</div>
