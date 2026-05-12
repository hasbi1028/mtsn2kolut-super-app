<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import {
		AlertTriangle,
		ArrowRight,
		BookOpenCheck,
		CalendarClock,
		CheckCircle2,
		GraduationCap,
		Layers,
		RefreshCw,
		School,
		UserCheck,
		Users
	} from '@lucide/svelte';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { academicCopy } from '$lib/academic/copy';
	import { readClientApiData } from '$lib/client/api';

	type AcademicDashboardSummary = {
		active_academic_year: string;
		active_semester: string;
		total_classes: number;
		total_active_students: number;
		students_without_class: number;
		classes_without_homeroom: number;
		subject_assignments_missing_teacher: number;
		timetable_conflicts: number;
		student_accounts_missing: number;
		parent_accounts_missing: number;
		classes_without_curriculum_profile: number;
		classes_weekly_hours_under_42: number;
		classes_weekly_hours_over_48: number;
		required_subjects_missing_teacher: number;
		teachers_under_24_hours: number;
		timetable_hours_mismatch: number;
		non_ranking_subjects_in_ranking: number;
	};

	type StatusCard = {
		label: string;
		value: number;
		description: string;
		tone: 'default' | 'warning' | 'danger';
	};

	type ChecklistItem = {
		label: string;
		count: number;
		description: string;
		href: string;
	};

	const emptySummary: AcademicDashboardSummary = {
		active_academic_year: '',
		active_semester: '',
		total_classes: 0,
		total_active_students: 0,
		students_without_class: 0,
		classes_without_homeroom: 0,
		subject_assignments_missing_teacher: 0,
		timetable_conflicts: 0,
		student_accounts_missing: 0,
		parent_accounts_missing: 0,
		classes_without_curriculum_profile: 0,
		classes_weekly_hours_under_42: 0,
		classes_weekly_hours_over_48: 0,
		required_subjects_missing_teacher: 0,
		teachers_under_24_hours: 0,
		timetable_hours_mismatch: 0,
		non_ranking_subjects_in_ranking: 0
	};

	let summary = $state<AcademicDashboardSummary>(emptySummary);
	let dashboardPromise = $state<Promise<AcademicDashboardSummary> | null>(null);
	let refreshBusy = $state(false);
	let requestId = 0;

	const activeYearLabel = $derived(summary.active_academic_year || 'Belum ditetapkan');
	const semesterLabel = $derived(summary.active_semester || 'Belum ditetapkan');
	const readinessIssues = $derived(
		summary.students_without_class +
			summary.classes_without_homeroom +
			summary.subject_assignments_missing_teacher +
			summary.timetable_conflicts +
			summary.classes_without_curriculum_profile +
			summary.classes_weekly_hours_under_42 +
			summary.classes_weekly_hours_over_48 +
			summary.required_subjects_missing_teacher +
			summary.timetable_hours_mismatch
	);
	const readinessLabel = $derived(readinessIssues === 0 ? 'Data siap dipakai' : `${readinessIssues} data perlu diperiksa`);

	const statusCards = $derived.by<StatusCard[]>(() => [
		{
			label: 'Rombel aktif',
			value: summary.total_classes,
			description: 'Rombel aktif pada tahun ajaran berjalan.',
			tone: 'default'
		},
		{
			label: 'Siswa aktif',
			value: summary.total_active_students,
			description: 'Siswa aktif yang tercatat di data induk.',
			tone: 'default'
		},
		{
			label: 'Siswa belum memiliki rombel',
			value: summary.students_without_class,
			description: 'Perlu ditempatkan ke rombel aktif.',
			tone: summary.students_without_class > 0 ? 'warning' : 'default'
		},
		{
			label: 'Rombel belum memiliki wali kelas',
			value: summary.classes_without_homeroom,
			description: 'Rombel aktif belum memiliki wali kelas aktif.',
			tone: summary.classes_without_homeroom > 0 ? 'warning' : 'default'
		},
		{
			label: 'Penugasan guru mapel belum lengkap',
			value: summary.subject_assignments_missing_teacher,
			description: 'Ada mapel yang belum memiliki guru aktif.',
			tone: summary.subject_assignments_missing_teacher > 0 ? 'warning' : 'default'
		},
		{
			label: 'Jadwal perlu diperiksa',
			value: summary.timetable_conflicts,
			description: 'Ada jadwal kelas, guru, atau ruang yang bertabrakan.',
			tone: summary.timetable_conflicts > 0 ? 'danger' : 'default'
		},
		{
			label: 'Rombel tanpa profil kurikulum',
			value: summary.classes_without_curriculum_profile,
			description: 'Rombel perlu profil Kurikulum Merdeka/KMA 1503 aktif.',
			tone: summary.classes_without_curriculum_profile > 0 ? 'warning' : 'default'
		},
		{
			label: 'JP mingguan di luar 42–48',
			value: summary.classes_weekly_hours_under_42 + summary.classes_weekly_hours_over_48,
			description: 'Validasi batas JP KMA 1503 termasuk tambahan maksimal +6 JP.',
			tone: summary.classes_weekly_hours_under_42 + summary.classes_weekly_hours_over_48 > 0 ? 'warning' : 'default'
		},
		{
			label: 'Guru kurang 24 JP',
			value: summary.teachers_under_24_hours,
			description: 'Beban mengajar guru aktif perlu dipantau.',
			tone: summary.teachers_under_24_hours > 0 ? 'warning' : 'default'
		}
	]);

	const checklist = $derived.by<ChecklistItem[]>(() => [
		{
			label: 'Tempatkan semua siswa aktif ke rombel',
			count: summary.students_without_class,
			description: 'Siswa tanpa rombel akan sulit dipakai untuk jadwal, asesmen, dan akses layanan siswa.',
			href: resolve('/students')
		},
		{
			label: 'Lengkapi wali kelas',
			count: summary.classes_without_homeroom,
			description: 'Wali kelas menjadi penanggung jawab operasional rombel.',
			href: resolve('/akademik/rombel')
		},
		{
			label: 'Periksa guru mapel',
			count: summary.subject_assignments_missing_teacher,
			description: 'Penugasan mapel harus memiliki guru aktif.',
			href: resolve('/akademik/guru-mapel')
		},
		{
			label: 'Periksa jadwal yang bertabrakan',
			count: summary.timetable_conflicts,
			description: 'Jadwal yang bertabrakan perlu dibenahi sebelum dipakai harian.',
			href: resolve('/akademik/jadwal')
		},
		{
			label: 'Tetapkan profil kurikulum rombel',
			count: summary.classes_without_curriculum_profile,
			description: 'Profil kurikulum menjadi dasar alokasi JP KMA 1503.',
			href: resolve('/akademik/kurikulum')
		},
		{
			label: 'Lengkapi guru mapel wajib KMA',
			count: summary.required_subjects_missing_teacher,
			description: 'Mapel wajib yang masuk jadwal harus memiliki guru aktif.',
			href: resolve('/akademik/guru-mapel')
		},
		{
			label: 'Periksa JP mingguan rombel',
			count: summary.classes_weekly_hours_under_42 + summary.classes_weekly_hours_over_48,
			description: 'JP efektif sebaiknya berada pada rentang 42–48 sesuai batas tambahan.',
			href: resolve('/akademik/guru-mapel')
		}
	]);

	async function fetchDashboard(): Promise<AcademicDashboardSummary> {
		return await fetch('/api/academic/dashboard').then((response) =>
			readClientApiData<AcademicDashboardSummary>(response, 'Gagal memuat ringkasan akademik')
		);
	}

	function loadDashboard() {
		const current = ++requestId;
		dashboardPromise = fetchDashboard().then((nextSummary) => {
			if (current === requestId) {
				summary = nextSummary;
			}
			return nextSummary;
		});
	}

	async function refreshDashboard() {
		refreshBusy = true;
		try {
			loadDashboard();
			await dashboardPromise;
		} finally {
			refreshBusy = false;
		}
	}

	function dashboardErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Ringkasan akademik belum dapat dimuat. Periksa koneksi layanan sistem lalu coba lagi.';
	}

	function handleRenderError(error: unknown) {
		console.error('Academic dashboard render failed', error);
	}

	function statusToneClass(tone: StatusCard['tone']) {
		if (tone === 'danger') return 'border-destructive/30 bg-destructive/5';
		if (tone === 'warning') return 'border-amber-300 bg-amber-50';
		return '';
	}

	onMount(loadDashboard);
</script>

<svelte:head>
	<title>Ringkasan Akademik | MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
		<div>
			<p class="text-sm font-medium text-primary">Akademik</p>
			<h1 class="text-2xl font-semibold tracking-normal text-foreground">Ringkasan Akademik</h1>
			<p class="mt-1 max-w-2xl text-sm text-muted-foreground">
				Ringkasan kelengkapan data tahun ajaran, rombel, siswa, wali kelas, guru mapel, dan jadwal.
			</p>
		</div>
		<Button variant="outline" onclick={() => void refreshDashboard()} disabled={refreshBusy} aria-label="Muat ulang ringkasan akademik">
			<RefreshCw class={`mr-2 size-4 ${refreshBusy ? 'animate-spin' : ''}`} />
			Muat Ulang
		</Button>
	</div>

	<AsyncContent promise={dashboardPromise} onerror={handleRenderError}>
		{#snippet pending()}
			<div class="space-y-4">
				<Skeleton class="h-32 w-full" />
				<div class="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
					<Skeleton class="h-28 w-full" />
					<Skeleton class="h-28 w-full" />
					<Skeleton class="h-28 w-full" />
					<Skeleton class="h-28 w-full" />
					<Skeleton class="h-28 w-full" />
					<Skeleton class="h-28 w-full" />
				</div>
				<Skeleton class="h-72 w-full" />
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel
				title="Ringkasan Akademik Belum Tersaji"
				message={dashboardErrorMessage(error)}
				onRetry={() => {
					reset?.();
					loadDashboard();
				}}
			/>
		{/snippet}

		<div class="rounded-md border bg-card p-5">
			<div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
				<div class="flex items-start gap-3">
					<div class="rounded-md bg-primary/10 p-2 text-primary">
						<School class="size-6" />
					</div>
					<div class="space-y-1">
						<p class="text-sm text-muted-foreground">{academicCopy.labels.activeAcademicYear}</p>
						<h2 class="text-xl font-semibold text-foreground">{activeYearLabel}</h2>
						<p class="text-sm text-muted-foreground">Semester {semesterLabel}</p>
					</div>
				</div>
				<div class="flex flex-wrap items-center gap-2">
					<Badge class={readinessIssues === 0 ? 'border-primary/20 bg-primary/15 text-primary' : 'border-amber-300 bg-amber-50 text-amber-800'}>
						{#if readinessIssues === 0}
							<CheckCircle2 class="mr-1 size-3.5" />
						{:else}
							<AlertTriangle class="mr-1 size-3.5" />
						{/if}
						{readinessLabel}
					</Badge>
				</div>
			</div>
		</div>

		<div class="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
			{#each statusCards as card (card.label)}
				<Card.Root class={statusToneClass(card.tone)}>
					<Card.Header class="pb-2">
						<Card.Description>{card.label}</Card.Description>
						<Card.Title class="text-2xl tabular-nums">{card.value}</Card.Title>
					</Card.Header>
					<Card.Content>
						<p class="text-sm text-muted-foreground">{card.description}</p>
					</Card.Content>
				</Card.Root>
			{/each}
		</div>

		<div class="grid gap-4 xl:grid-cols-[1.2fr_0.8fr]">
			<Card.Root>
				<Card.Header>
					<Card.Title class="text-base">{academicCopy.labels.academicDataCompleteness}</Card.Title>
					<Card.Description>{academicCopy.helper.academicDataCompleteness}</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-3">
					{#each checklist as item (item.label)}
						<a href={item.href} class="flex items-start justify-between gap-4 rounded-md border p-3 transition-colors hover:bg-muted/60 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
							<div class="min-w-0 space-y-1">
								<div class="flex flex-wrap items-center gap-2">
									<p class="font-medium text-foreground">{item.label}</p>
									{#if item.count === 0}
										<Badge class="border-primary/20 bg-primary/15 text-primary">Lengkap</Badge>
									{:else}
										<Badge class="border-amber-300 bg-amber-50 text-amber-800">{item.count} perlu diperiksa</Badge>
									{/if}
								</div>
								<p class="text-sm text-muted-foreground">{item.description}</p>
							</div>
							<ArrowRight class="mt-1 size-4 shrink-0 text-muted-foreground" />
						</a>
					{/each}
				</Card.Content>
			</Card.Root>

			<Card.Root>
				<Card.Header>
					<Card.Title class="text-base">Aksi Cepat</Card.Title>
					<Card.Description>Masuk ke pekerjaan akademik yang paling sering dipakai operator.</Card.Description>
				</Card.Header>
				<Card.Content class="grid gap-2">
					<Button variant="outline" class="h-auto justify-start gap-3 py-3" href={resolve('/akademik/rombel')}>
						<Layers class="size-4 text-primary" />
						<span class="text-left">
							<span class="block font-medium">Kelola Rombel</span>
							<span class="block text-xs text-muted-foreground">Wali kelas, siswa, guru mapel, dan jadwal per rombel</span>
						</span>
					</Button>
					<Button variant="outline" class="h-auto justify-start gap-3 py-3" href={resolve('/students')}>
						<Users class="size-4 text-primary" />
						<span class="text-left">
							<span class="block font-medium">Kelola Siswa</span>
							<span class="block text-xs text-muted-foreground">Data siswa, rombel, status, dan akun</span>
						</span>
					</Button>
					<Button variant="outline" class="h-auto justify-start gap-3 py-3" href={resolve('/akademik/guru-mapel')}>
						<UserCheck class="size-4 text-primary" />
						<span class="text-left">
							<span class="block font-medium">Guru Mapel</span>
							<span class="block text-xs text-muted-foreground">Atur tabel penugasan guru per mapel dan rombel</span>
						</span>
					</Button>
					<Button variant="outline" class="h-auto justify-start gap-3 py-3" href={resolve('/akademik/jadwal')}>
						<CalendarClock class="size-4 text-primary" />
						<span class="text-left">
							<span class="block font-medium">Jadwal Pelajaran</span>
							<span class="block text-xs text-muted-foreground">Lihat jadwal kelas, guru, dan hari aktif</span>
						</span>
					</Button>
					<Button variant="outline" class="h-auto justify-start gap-3 py-3" href={resolve('/akademik/tahun-ajaran')}>
						<BookOpenCheck class="size-4 text-primary" />
						<span class="text-left">
							<span class="block font-medium">Tahun Ajaran</span>
							<span class="block text-xs text-muted-foreground">Aktivasi, pratinjau kenaikan kelas, dan impor/unduh data</span>
						</span>
					</Button>
				</Card.Content>
			</Card.Root>
		</div>

		<div class="rounded-md border bg-muted/30 p-4">
			<div class="flex items-start gap-3">
				<GraduationCap class="mt-0.5 size-5 text-primary" />
				<p class="text-sm text-muted-foreground">
					Akun siswa belum tersedia: <span class="font-medium text-foreground">{summary.student_accounts_missing}</span>.
					Akun orang tua belum tersedia: <span class="font-medium text-foreground">{summary.parent_accounts_missing}</span>.
				</p>
			</div>
		</div>
	</AsyncContent>
</div>
