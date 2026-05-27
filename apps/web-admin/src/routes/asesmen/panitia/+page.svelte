<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { MicroActionTable } from '$lib/components/ops';

	type PanitiaRoute =
		| '/asesmen/ringkas'
		| '/asesmen/persiapan'
		| '/asesmen/pelaksanaan'
		| '/asesmen/kegiatan'
		| '/asesmen/paket'
		| '/asesmen/sesi'
		| '/asesmen/pengawasan'
		| '/asesmen/aplikasi-siswa'
		| '/asesmen/aplikasi-siswa/matrix'
		| '/asesmen/aplikasi-siswa/release'
		| '/asesmen/hasil'
		| '/asesmen/non-tes';

	type AdminTool = {
		phase: string;
		title: string;
		description: string;
		href: PanitiaRoute;
		cta: string;
		level: 'utama' | 'teknis' | 'lanjutan';
	};

	const roles = $derived(page.data.user?.roles ?? (page.data.user?.role ? [page.data.user.role] : []));
	const permissions = $derived(page.data.user?.permissions ?? []);
	const canAccess = $derived(
		roles.includes('admin')
			|| permissions.includes('asesmen.operator')
			|| permissions.includes('asesmen.event_manage')
			|| permissions.includes('asesmen.package_manage')
	);
	const canOpenResults = $derived(roles.includes('admin') || permissions.includes('asesmen.result_read'));

	const baseTools: AdminTool[] = [
		{ phase: 'Ringkas', title: 'Meja Kerja Panitia', description: 'Ringkasan sesi, paket, dan pembagian ruang.', href: '/asesmen/ringkas', cta: 'Buka Ringkas', level: 'utama' },
		{ phase: 'Pra', title: 'Persiapan Ujian', description: 'Checklist kegiatan, paket, jadwal, peserta, ruang, token, dan kartu.', href: '/asesmen/persiapan', cta: 'Buka Persiapan', level: 'utama' },
		{ phase: 'Data', title: 'Kegiatan Ujian', description: 'Kelola identitas kegiatan, anggota, kartu, dan arsip kegiatan.', href: '/asesmen/kegiatan', cta: 'Kelola Kegiatan', level: 'teknis' },
		{ phase: 'Data', title: 'Paket Ujian', description: 'Pilih dan kelola paket soal siap ujian.', href: '/asesmen/paket', cta: 'Kelola Paket', level: 'teknis' },
		{ phase: 'Jadwal', title: 'Jadwal/Sesi Ujian', description: 'Atur sesi, peserta, ruang, kursi, token, dan status pelaksanaan.', href: '/asesmen/sesi', cta: 'Kelola Sesi', level: 'teknis' },
		{ phase: 'Hari-H', title: 'Pelaksanaan & Pantau Ruang', description: 'Masuk ke alur hari-H, ruang berjalan, dan atensi peserta.', href: '/asesmen/pelaksanaan', cta: 'Buka Pelaksanaan', level: 'utama' },
		{ phase: 'Perangkat', title: 'Panduan Perangkat Siswa', description: 'Panduan Portal Ujian Web, latihan lokal, dan mode cadangan.', href: '/asesmen/aplikasi-siswa', cta: 'Buka Panduan', level: 'utama' },
		{ phase: 'Perangkat', title: 'Uji Perangkat', description: 'Tabel uji perangkat untuk operator saat simulasi.', href: '/asesmen/aplikasi-siswa/matrix', cta: 'Buka Uji Perangkat', level: 'lanjutan' },
		{ phase: 'Perangkat', title: 'Arsip Rilis Aplikasi', description: 'Arsip rilis aplikasi siswa untuk panitia teknis.', href: '/asesmen/aplikasi-siswa/release', cta: 'Buka Arsip', level: 'lanjutan' },
		{ phase: 'Lainnya', title: 'Penilaian Non-Tes', description: 'Workflow terpisah dari ujian digital; gunakan hanya bila panitia membutuhkan.', href: '/asesmen/non-tes', cta: 'Buka Non-Tes', level: 'lanjutan' }
	];

	const tools = $derived<AdminTool[]>(
		canOpenResults
			? [
				...baseTools,
				{ phase: 'Akhir', title: 'Hasil & Arsip', description: 'Rekap nilai, hasil, berita acara, dan unduhan akhir.', href: '/asesmen/hasil', cta: 'Buka Hasil', level: 'utama' }
			]
			: baseTools
	);

	const columns = [
		{ key: 'tool', label: 'Fitur', class: 'min-w-[16rem]' },
		{ key: 'phase', label: 'Fase', class: 'w-28' },
		{ key: 'description', label: 'Keterangan', class: 'min-w-[22rem]' },
		{ key: 'level', label: 'Mode', headClass: 'text-right', class: 'text-right' }
	];

	function levelClass(level: AdminTool['level']) {
		if (level === 'utama') return 'border-primary/20 bg-primary/10 text-primary';
		if (level === 'teknis') return 'border-warning/30 bg-warning/10 text-warning';
		return 'border-border bg-muted text-muted-foreground';
	}
</script>

<svelte:head>
	<title>Mode Lengkap Panitia — MTsN 2 Kolut</title>
</svelte:head>

{#if canAccess}
	<div class="space-y-5">
		<section class="rounded-2xl border border-border bg-card p-4 shadow-sm">
			<div class="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
				<div class="max-w-3xl space-y-2">
					<Badge class="border-warning/30 bg-warning/10 text-warning" variant="outline">Mode Lengkap Panitia</Badge>
					<h1 class="text-2xl font-semibold tracking-tight text-foreground">Fitur Teknis Ujian Digital</h1>
					<p class="text-sm leading-6 text-muted-foreground">
						Halaman ini menampung fitur lengkap agar sidebar dan alur guru/pengawas tetap sederhana. Gunakan hanya untuk admin/panitia/operator.
					</p>
				</div>
				<div class="flex flex-wrap gap-2">
					<Button href={resolve('/asesmen/ringkas')} size="sm">Kembali ke Ringkasan</Button>
					<Button href={resolve('/asesmen/pelaksanaan')} variant="outline" size="sm">Buka Pelaksanaan</Button>
				</div>
			</div>
		</section>

		<MicroActionTable title="Pintu Mode Lengkap" description="Fitur utama tetap didahulukan; fitur teknis dan lanjutan tidak muncul di sidebar guru/pengawas." columns={columns} rows={tools} rowKey={(row) => (row as AdminTool).href} tableClass="min-w-[860px]">
			{#snippet cell(row, column)}
				{@const tool = row as AdminTool}
				{#if column.key === 'tool'}
					<p class="font-semibold text-foreground">{tool.title}</p>
				{:else if column.key === 'phase'}
					<Badge variant="outline">{tool.phase}</Badge>
				{:else if column.key === 'description'}
					<p class="max-w-2xl text-xs leading-5 text-muted-foreground">{tool.description}</p>
				{:else}
					<Badge class={levelClass(tool.level)} variant="outline">{tool.level}</Badge>
				{/if}
			{/snippet}
			{#snippet actions(row)}
				{@const tool = row as AdminTool}
				<Button href={resolve(tool.href)} size="xs" variant={tool.level === 'utama' ? 'default' : 'outline'}>{tool.cta}</Button>
			{/snippet}
			{#snippet mobile(row)}
				{@const tool = row as AdminTool}
				<div class="space-y-2 text-xs">
					<div class="flex items-start justify-between gap-2">
						<div>
							<p class="font-semibold text-foreground">{tool.title}</p>
							<p class="mt-1 leading-5 text-muted-foreground">{tool.description}</p>
						</div>
						<Badge class={levelClass(tool.level)} variant="outline">{tool.level}</Badge>
					</div>
					<Button href={resolve(tool.href)} size="sm" variant="outline" class="w-full">{tool.cta}</Button>
				</div>
			{/snippet}
		</MicroActionTable>
	</div>
{:else}
	<div class="flex min-h-[60vh] flex-col items-center justify-center px-4 py-16 text-center">
		<div class="max-w-lg rounded-2xl border border-border bg-card p-8 shadow-sm">
			<h1 class="text-xl font-semibold text-foreground">Mode lengkap hanya untuk panitia</h1>
			<p class="mt-3 text-sm leading-6 text-muted-foreground">Gunakan menu Ruang Saya atau Pelaksanaan bila Bapak/Ibu bertugas sebagai pengawas.</p>
			<Button href={resolve('/asesmen/ruang-saya')} variant="outline" class="mt-6">Buka Ruang Saya</Button>
		</div>
	</div>
{/if}
