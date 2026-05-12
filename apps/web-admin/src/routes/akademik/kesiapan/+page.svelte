<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { readClientApiData } from '$lib/client/api';

	type ReadinessPayload = {
		active_academic_year: string;
		active_semester: string;
		total_classes: number;
		total_active_students: number;
		classes_without_students: number;
		classes_without_homeroom: number;
		classes_without_curriculum_profile: number;
		report_subjects_missing_teacher: number;
		timetable_conflicts: number;
		teachers_under_24_hours: number;
		teachers_over_40_hours: number;
		report_assignments_without_components: number;
		students_with_incomplete_grades: number;
		report_assignments_not_finalized: number;
		report_descriptions_missing: number;
		report_settings_count: number;
	};

	type CheckItem = {
		group: string;
		label: string;
		count: number;
		goodWhenZero?: boolean;
		action: string;
		href: string;
	};

	let payload = $state<ReadinessPayload | null>(null);
	let loading = $state(true);
	let errorMessage = $state('');

	const checks = $derived.by<CheckItem[]>(() => {
		if (!payload) return [];
		return [
			{ group: 'Data Dasar', label: 'Rombel tanpa siswa', count: payload.classes_without_students, action: 'Lengkapi siswa pada rombel', href: '/students' },
			{ group: 'Data Dasar', label: 'Rombel tanpa wali kelas', count: payload.classes_without_homeroom, action: 'Tetapkan wali kelas', href: '/akademik/rombel' },
			{ group: 'Kurikulum', label: 'Rombel tanpa struktur kurikulum', count: payload.classes_without_curriculum_profile, action: 'Tetapkan struktur kurikulum', href: '/akademik/kurikulum' },
			{ group: 'Kurikulum', label: 'Mapel rapor tanpa guru', count: payload.report_subjects_missing_teacher, action: 'Lengkapi guru mapel', href: '/akademik/guru-mapel' },
			{ group: 'Jadwal', label: 'Jadwal perlu peninjauan', count: payload.timetable_conflicts, action: 'Periksa jadwal mingguan', href: '/akademik/jadwal' },
			{ group: 'Beban Guru', label: 'Guru kurang dari 24 JP', count: payload.teachers_under_24_hours, action: 'Tinjau beban guru', href: '/akademik/beban-guru' },
			{ group: 'Beban Guru', label: 'Guru perlu peninjauan JP', count: payload.teachers_over_40_hours, action: 'Tinjau beban guru', href: '/akademik/beban-guru' },
			{ group: 'Nilai', label: 'Mapel rapor belum punya komponen nilai', count: payload.report_assignments_without_components, action: 'Lengkapi komponen nilai', href: '/grades' },
			{ group: 'Nilai', label: 'Siswa dengan nilai belum lengkap', count: payload.students_with_incomplete_grades, action: 'Lengkapi nilai siswa', href: '/grades' },
			{ group: 'Rapor', label: 'Mapel rapor belum dikunci', count: payload.report_assignments_not_finalized, action: 'Periksa cetak rapor', href: '/grades/rapor' },
			{ group: 'Rapor', label: 'Deskripsi capaian belum lengkap', count: payload.report_descriptions_missing, action: 'Isi deskripsi capaian', href: '/grades/rapor' },
			{ group: 'Rapor', label: 'Pengaturan rapor belum tersedia', count: payload.report_settings_count > 0 ? 0 : 1, action: 'Simpan pengaturan rapor', href: '/grades/rapor' }
		];
	});

	const totalIssues = $derived(checks.reduce((sum, item) => sum + Math.max(0, item.count), 0));
	const readyCount = $derived(checks.filter((item) => item.count === 0).length);
	const percentReady = $derived(checks.length === 0 ? 0 : Math.round((readyCount / checks.length) * 100));
	const priorityChecks = $derived(checks.filter((item) => item.count > 0).slice(0, 5));

	function badgeVariant(count: number) {
		return count === 0 ? 'default' : 'outline';
	}

	async function load() {
		loading = true;
		errorMessage = '';
		try {
			const response = await fetch('/api/academic/readiness');
			payload = await readClientApiData<ReadinessPayload>(response, 'Gagal memuat kesiapan akademik');
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'Gagal memuat kesiapan akademik';
		} finally {
			loading = false;
		}
	}

	onMount(load);
</script>

<svelte:head><title>Kesiapan Akademik & Rapor — Akademik</title></svelte:head>

<div class="space-y-6 p-4 md:p-6">
	<div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
		<div>
			<p class="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">Akademik</p>
			<h1 class="text-2xl font-semibold text-foreground">Kesiapan Akademik & Rapor</h1>
			<p class="text-sm text-muted-foreground">Pantau kelengkapan data akademik sebelum jadwal berjalan dan rapor dicetak.</p>
		</div>
		<div class="flex gap-2">
			<Badge variant="outline">{payload?.active_academic_year || 'Tahun ajaran aktif'}</Badge>
			<Badge variant="outline">{payload?.active_semester || 'Semester aktif'}</Badge>
			<Button variant="outline" onclick={load}>Muat ulang</Button>
		</div>
	</div>

	{#if loading}
		<Card.Root><Card.Content class="p-6 text-sm text-muted-foreground">Memuat kesiapan akademik...</Card.Content></Card.Root>
	{:else if errorMessage}
		<Card.Root><Card.Content class="p-6 text-sm text-destructive">{errorMessage}</Card.Content></Card.Root>
	{:else if payload}
		<div class="grid gap-3 md:grid-cols-4">
			<Card.Root><Card.Content class="p-4"><p class="text-xs text-muted-foreground">Kesiapan</p><p class="text-2xl font-semibold">{percentReady}%</p><p class="text-xs text-muted-foreground">{readyCount}/{checks.length} pemeriksaan aman</p></Card.Content></Card.Root>
			<Card.Root><Card.Content class="p-4"><p class="text-xs text-muted-foreground">Hal perlu ditindaklanjuti</p><p class="text-2xl font-semibold">{totalIssues}</p></Card.Content></Card.Root>
			<Card.Root><Card.Content class="p-4"><p class="text-xs text-muted-foreground">Rombel aktif</p><p class="text-2xl font-semibold">{payload.total_classes}</p></Card.Content></Card.Root>
			<Card.Root><Card.Content class="p-4"><p class="text-xs text-muted-foreground">Siswa aktif</p><p class="text-2xl font-semibold">{payload.total_active_students}</p></Card.Content></Card.Root>
		</div>

		{#if priorityChecks.length > 0}
			<Card.Root class="border-amber-200 bg-amber-50/60 dark:border-amber-900 dark:bg-amber-950/20">
				<Card.Header><Card.Title class="text-base">Prioritas Tindak Lanjut</Card.Title><Card.Description>Mulai dari daftar ini agar data akademik dan rapor lebih siap.</Card.Description></Card.Header>
				<Card.Content class="grid gap-3 md:grid-cols-2">
					{#each priorityChecks as item}
						<a class="rounded-lg border bg-background p-4 transition hover:border-primary" href={item.href}>
							<div class="flex items-start justify-between gap-3"><div><p class="font-medium">{item.label}</p><p class="text-sm text-muted-foreground">{item.action}</p></div><Badge variant="outline">{item.count}</Badge></div>
						</a>
					{/each}
				</Card.Content>
			</Card.Root>
		{:else}
			<Card.Root class="border-emerald-200 bg-emerald-50/60 dark:border-emerald-900 dark:bg-emerald-950/20">
				<Card.Content class="p-6"><p class="font-medium text-emerald-700 dark:text-emerald-300">Semua pemeriksaan utama sudah aman.</p><p class="text-sm text-muted-foreground">Data akademik dan rapor siap digunakan sesuai pemeriksaan sistem.</p></Card.Content>
			</Card.Root>
		{/if}

		<Card.Root>
			<Card.Header><Card.Title class="text-base">Daftar Pemeriksaan</Card.Title><Card.Description>Setiap baris menunjukkan bagian yang perlu dilengkapi oleh operator madrasah.</Card.Description></Card.Header>
			<Card.Content>
				<Table.Root>
					<Table.Header><Table.Row><Table.Head>Bagian</Table.Head><Table.Head>Pemeriksaan</Table.Head><Table.Head>Jumlah</Table.Head><Table.Head>Keterangan</Table.Head></Table.Row></Table.Header>
					<Table.Body>
						{#each checks as item}
							<Table.Row>
								<Table.Cell class="font-medium">{item.group}</Table.Cell>
								<Table.Cell>{item.label}</Table.Cell>
								<Table.Cell><Badge variant={badgeVariant(item.count)}>{item.count === 0 ? 'Aman' : item.count}</Badge></Table.Cell>
								<Table.Cell><a class="text-primary hover:underline" href={item.href}>{item.count === 0 ? 'Tidak perlu tindakan' : item.action}</a></Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			</Card.Content>
		</Card.Root>
	{/if}
</div>
