<script lang="ts">
	import BarChart3Icon from '@lucide/svelte/icons/bar-chart-3';
	import ArchiveIcon from '@lucide/svelte/icons/archive';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import MonitorPlayIcon from '@lucide/svelte/icons/monitor-play';
	import ClipboardCheckIcon from '@lucide/svelte/icons/clipboard-check';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import MicroActionTable from '$lib/components/ops/MicroActionTable.svelte';

	type ResultRoute =
		| '/asesmen/ringkas'
		| '/asesmen/kegiatan'
		| '/asesmen/sesi'
		| '/asesmen/pelaksanaan'
		| '/asesmen/ruang-saya'
		| '/asesmen/aplikasi-siswa'
		| '/asesmen'
		| '/asesmen/pengawasan'
		| '/asesmen/persiapan';

	type ResultRow = {
		id: string;
		title: string;
		helper: string;
		status: string;
		coverage: string;
		href: ResultRoute;
		action: string;
		secondaryHref?: ResultRoute;
		secondaryAction?: string;
		icon: typeof BarChart3Icon;
	};

	const rowColumns = [
		{ key: 'task', label: 'Alur Hasil', class: 'min-w-[18rem]' },
		{ key: 'status', label: 'Status', class: 'min-w-[12rem]' },
		{ key: 'coverage', label: 'Cakupan', class: 'min-w-[14rem]' }
	];

	const userRoles = $derived(page.data.user?.roles ?? (page.data.user?.role ? [page.data.user.role] : []));
	const userPermissions = $derived((page.data.user?.permissions ?? []).map((permission) => permission.trim()).filter(Boolean));
	const isOperator = $derived(
		userPermissions.includes('asesmen.operator') ||
		userPermissions.includes('asesmen.event_manage') ||
		userPermissions.includes('asesmen.package_manage')
	);
	const isResultReader = $derived(userPermissions.includes('asesmen.result_read'));
	const isProctor = $derived(userPermissions.includes('asesmen.proctor'));
	const canAccess = $derived(isResultReader || userRoles.includes('admin') || userRoles.includes('guru'));
	const canOpenRingkasan = $derived(
		userRoles.includes('admin')
			|| userPermissions.includes('asesmen.operator')
			|| userPermissions.includes('asesmen.event_manage')
			|| userPermissions.includes('asesmen.package_manage')
	);
	const resultRows = $derived<ResultRow[]>(
		isOperator
			? [
				{
					id: 'event-results',
					title: 'Rekap Kegiatan',
					helper: 'Masuk dari daftar kegiatan untuk membaca hasil, detail sesi, dan jejak kesiapan penutupan.',
					status: 'Pintu utama operator',
					coverage: 'Per kegiatan · hasil gabungan · finalisasi',
					href: '/asesmen/kegiatan',
					action: 'Buka Daftar Kegiatan',
					secondaryHref: '/asesmen/ringkas',
					secondaryAction: 'Kembali ke Ringkasan',
					icon: BarChart3Icon
				},
				{
					id: 'session-results',
					title: 'Hasil Sesi & BA',
					helper: 'Pilih sesi untuk melihat kiriman peserta, BA sesi, nilai, dan analisis butir.',
					status: 'Dipakai saat verifikasi',
					coverage: 'Per sesi · BA · analisis soal',
					href: '/asesmen/sesi',
					action: 'Buka Daftar Sesi',
					secondaryHref: '/asesmen/ruang-saya',
					secondaryAction: 'Buka Ruang Saya',
					icon: FileTextIcon
				},
				{
					id: 'archive-results',
					title: 'Arsip Final',
					helper: 'Gunakan saat paket, hasil, dokumen cetak, dan pengesahan sudah siap untuk ditutup.',
					status: 'Tahap penutupan kegiatan',
					coverage: 'Checklist arsip · pengesahan · dokumen BA',
					href: '/asesmen/kegiatan',
					action: 'Buka Kegiatan & Arsip',
					secondaryHref: '/asesmen/persiapan',
					secondaryAction: 'Kembali ke Persiapan',
					icon: ArchiveIcon
				}
			]
			: [
				{
					id: 'reader-results',
					title: 'Rekap Hasil Kegiatan',
					helper: 'Buka daftar kegiatan untuk melihat hasil yang sudah dibuka oleh operator/panitia.',
					status: isResultReader ? 'Baca hasil tersedia' : 'Ikuti arahan operator',
					coverage: 'Per kegiatan · rekap akhir',
					href: '/asesmen/kegiatan',
					action: 'Buka Kegiatan',
					secondaryHref: '/asesmen/ringkas',
					secondaryAction: 'Kembali ke Ringkasan',
					icon: BarChart3Icon
				},
				{
					id: 'reader-monitoring',
					title: 'Pantau Pelaksanaan',
					helper: 'Gunakan bila hasil perlu dicocokkan dengan ruang aktif, status kiriman, atau kejadian pengawasan.',
					status: isProctor ? 'Mode ruang aktif' : 'Pendukung monitoring',
					coverage: 'Ruang berjalan · kejadian · status kiriman',
					href: '/asesmen/pelaksanaan',
					action: 'Buka Pelaksanaan',
					secondaryHref: '/asesmen/aplikasi-siswa',
					secondaryAction: 'Panduan Siswa',
					icon: MonitorPlayIcon
				},
				{
					id: 'reader-followup',
					title: 'Tindak Lanjut Operator',
					helper: 'Jika butuh BA sesi, analisis butir, atau arsip final, lanjutkan lewat operator/panitia.',
					status: 'Mode lengkap tetap hidup',
					coverage: 'BA sesi · analisis · arsip final',
					href: '/asesmen/sesi',
					action: 'Buka Sesi',
					secondaryHref: '/asesmen/ruang-saya',
					secondaryAction: 'Buka Ruang Saya',
					icon: ClipboardCheckIcon
				}
			]
	);

	function badgeClass(label: string) {
		if (label.includes('utama') || label.includes('tersedia') || label.includes('aktif')) {
			return 'border-success/20 bg-success/10 text-success';
		}
		if (label.includes('penutupan') || label.includes('verifikasi')) {
			return 'border-warning/30 bg-warning/10 text-warning';
		}
		return 'border-primary/20 bg-primary/10 text-primary';
	}
</script>

<svelte:head>
	<title>Hasil & Penutupan Ujian — MTsN 2 Kolaka Utara</title>
</svelte:head>

{#if canAccess}
	<div class="mx-auto max-w-6xl space-y-6 p-6">
		<section class="space-y-4">
			<div class="space-y-2">
				<p class="text-xs font-semibold uppercase tracking-[0.18em] text-primary">Asesmen · Hasil</p>
				<h1 class="text-2xl font-semibold tracking-tight text-foreground">Hasil & Penutupan Kegiatan</h1>
				<p class="max-w-3xl text-sm leading-6 text-muted-foreground">
					Halaman ini disederhanakan menjadi pintu kerja hasil. Rekap detail tetap ada di halaman
					kegiatan, sesi, dan arsip; di sini operator cukup memilih alur berikutnya.
				</p>
			</div>
			<div class="flex flex-wrap gap-2">
				{#if canOpenRingkasan}
					<Button href={resolve('/asesmen/ringkas')} variant="outline">Kembali ke Ringkasan</Button>
				{/if}
				{#if isOperator}
					<Badge variant="outline" class="border-primary/20 bg-primary/10 text-primary">Mode operator/panitia</Badge>
				{:else}
					<Badge variant="outline" class="border-muted bg-muted text-muted-foreground">Mode baca hasil</Badge>
				{/if}
			</div>
		</section>

		<MicroActionTable
			title="Alur utama hasil"
			description="Gunakan jalur ini untuk memeriksa hasil, BA sesi, pengawasan, dan penutupan arsip tanpa membuka terlalu banyak menu."
			columns={rowColumns}
			rows={resultRows}
			rowKey={(row) => (row as ResultRow).id}
			tableClass="min-w-[900px]"
		>
			{#snippet cell(row, column)}
				{@const item = row as ResultRow}
				{#if column.key === 'task'}
					<div class="flex items-start gap-3">
						<div class="grid size-10 shrink-0 place-items-center rounded-xl bg-primary/10 text-primary">
							<item.icon class="size-4" />
						</div>
						<div class="space-y-1">
							<p class="font-semibold text-foreground">{item.title}</p>
							<p class="text-xs leading-5 text-muted-foreground">{item.helper}</p>
						</div>
					</div>
				{:else if column.key === 'status'}
					<Badge variant="outline" class={badgeClass(item.status)}>{item.status}</Badge>
				{:else}
					<p class="text-xs leading-5 text-muted-foreground">{item.coverage}</p>
				{/if}
			{/snippet}
			{#snippet actions(row)}
				{@const item = row as ResultRow}
				<a href={resolve(item.href)} class="inline-flex items-center rounded-md bg-primary px-3 py-2 text-xs font-semibold text-primary-foreground hover:bg-primary/90">{item.action}</a>
				{#if item.secondaryHref && item.secondaryAction}
					<a href={resolve(item.secondaryHref)} class="inline-flex items-center rounded-md border border-border bg-card px-3 py-2 text-xs font-semibold text-foreground hover:bg-muted">{item.secondaryAction}</a>
				{/if}
			{/snippet}
		</MicroActionTable>
	</div>
{:else}
	<div class="flex min-h-[60vh] flex-col items-center justify-center gap-4 px-4 py-16 text-center">
		<div class="max-w-lg rounded-2xl border border-border bg-card p-8 shadow-sm">
			<h2 class="text-xl font-semibold text-foreground">Akses terbatas</h2>
			<p class="mt-3 text-sm leading-6 text-muted-foreground">
				Halaman hasil hanya tersedia untuk akun yang diberi akses baca hasil asesmen.
			</p>
			<div class="mt-6">
				<a href={resolve('/asesmen/ringkas')} class="inline-flex rounded-md border border-primary/20 bg-card px-4 py-2 text-sm font-semibold text-primary transition hover:bg-primary/10">Kembali ke Ringkasan</a>
			</div>
		</div>
	</div>
{/if}
