<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';

	type StatusTone = 'good' | 'warning' | 'danger';

	type StatusGuide = {
		label: string;
		tone: StatusTone;
		meaning: string;
		intervention: string;
	};

	const statuses: StatusGuide[] = [
		{ label: 'Hijau', tone: 'good', meaning: 'Perangkat tersambung dan tidak ada jawaban lokal yang tertahan.', intervention: 'Siswa dapat lanjut mengerjakan seperti biasa.' },
		{ label: 'Kuning', tone: 'warning', meaning: 'Koneksi atau sinkronisasi perlu dipantau.', intervention: 'Minta siswa tetap di layar ujian dan tunggu status pulih.' },
		{ label: 'Merah', tone: 'danger', meaning: 'Perangkat perlu bantuan pengawas/panitia.', intervention: 'Jangan izinkan kirim ujian sebelum panitia memastikan data aman.' }
	];

	const steps = [
		'Buka Portal Ujian Web di alamat /ujian pada perangkat siswa.',
		'Masukkan token peserta dan kode ruang sesuai kartu/pengawas.',
		'Pastikan status perangkat Hijau sebelum siswa mulai mengerjakan.',
		'Jika status Kuning atau Merah, siswa tetap di layar ujian dan pengawas menghubungi panitia.',
		'Sebelum kirim ujian, pastikan tidak ada jawaban yang masih menunggu sinkron.'
	];

	const adminLinks = [
		{ label: 'Uji Perangkat', href: '/asesmen/aplikasi-siswa/matrix' },
		{ label: 'Arsip Rilis Aplikasi', href: '/asesmen/aplikasi-siswa/release' }
	];

	const roles = $derived(page.data.user?.roles ?? (page.data.user?.role ? [page.data.user.role] : []));
	const permissions = $derived((page.data.user?.permissions ?? []).map((permission) => permission.trim()).filter(Boolean));
	const canOpenPanitiaTools = $derived(
		roles.includes('admin')
			|| permissions.includes('asesmen.operator')
			|| permissions.includes('asesmen.event_manage')
			|| permissions.includes('asesmen.package_manage')
			|| permissions.includes('asesmen.session_manage')
	);

	function badgeClass(tone: StatusTone) {
		if (tone === 'good') return 'border-primary/20 bg-primary/10 text-primary';
		if (tone === 'warning') return 'border-warning/30 bg-warning/10 text-warning';
		return 'border-destructive/30 bg-destructive/10 text-destructive';
	}
</script>

<svelte:head>
	<title>Panduan Perangkat Siswa — MTsN 2 Kolut</title>
</svelte:head>

<div class="space-y-5">
	<section class="rounded-2xl border border-border bg-card p-4 shadow-sm">
		<div class="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
			<div class="max-w-3xl space-y-2">
				<Badge class="border-primary/20 bg-primary/10 text-primary" variant="outline">Panduan Pengawas</Badge>
				<h1 class="text-2xl font-semibold tracking-tight text-foreground">Panduan Perangkat Siswa</h1>
				<p class="text-sm leading-6 text-muted-foreground">
					Jalur utama siswa adalah Portal Ujian Web. Halaman ini sengaja ringkas agar guru/pengawas cukup tahu langkah masuk, arti status, dan kapan menghubungi panitia.
				</p>
			</div>
			<div class="flex flex-wrap gap-2">
				<Button href={resolve('/asesmen/pelaksanaan')} size="sm">Buka Pelaksanaan</Button>
				<Button href={resolve('/asesmen/ruang-saya')} variant="outline" size="sm">Buka Ruang Saya</Button>
				<Button href="/ujian?demo=1" variant="outline" size="sm">Latihan Lokal</Button>
			</div>
		</div>
	</section>

	<section class="grid gap-4 lg:grid-cols-[1fr_0.8fr]">
		<div class="rounded-2xl border border-border bg-card p-4 shadow-sm">
			<p class="text-xs font-semibold uppercase tracking-[0.2em] text-primary">Langkah Pengawas</p>
			<h2 class="mt-1 text-lg font-semibold text-foreground">Saat Siswa Masuk Ujian</h2>
			<ol class="mt-4 space-y-3">
				{#each steps as step, index (step)}
					<li class="flex gap-3 rounded-xl border border-border bg-muted/40 px-3 py-2 text-sm leading-6 text-foreground">
						<span class="mt-0.5 inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary/15 text-xs font-semibold text-primary">{index + 1}</span>
						<span>{step}</span>
					</li>
				{/each}
			</ol>
		</div>

		<div class="rounded-2xl border border-border bg-card p-4 shadow-sm">
			<p class="text-xs font-semibold uppercase tracking-[0.2em] text-primary">Status</p>
			<h2 class="mt-1 text-lg font-semibold text-foreground">Arti Warna Perangkat</h2>
			<div class="mt-4 space-y-3">
				{#each statuses as status (status.label)}
					<div class="rounded-xl border border-border bg-muted/30 p-3 text-sm">
						<Badge class={badgeClass(status.tone)} variant="outline">{status.label}</Badge>
						<p class="mt-2 font-medium leading-5 text-foreground">{status.meaning}</p>
						<p class="mt-1 text-xs leading-5 text-muted-foreground"><span class="font-semibold text-foreground">Tindakan:</span> {status.intervention}</p>
					</div>
				{/each}
			</div>
		</div>
	</section>

	<section class="rounded-2xl border border-dashed border-border bg-muted/30 p-4 text-sm text-muted-foreground">
		<div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
			<p>
				Fitur teknis seperti uji perangkat, arsip aplikasi, dan pengaturan lanjutan dipindahkan ke Mode Lengkap Panitia agar tidak membingungkan pengawas.
			</p>
			{#if canOpenPanitiaTools}
				<div class="flex flex-wrap gap-2">
					{#each adminLinks as link (link.href)}
						<a class="rounded-md border border-border bg-card px-3 py-1.5 text-xs font-semibold text-foreground hover:border-primary/30 hover:bg-primary/10" href={resolve(link.href as '/')}>{link.label}</a>
					{/each}
				</div>
			{/if}
		</div>
	</section>
</div>
