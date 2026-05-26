<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import PublicHome from '$lib/components/PublicHome.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import { clientApiPath, readClientApiData } from '$lib/client/api';
	import { trackInternalAnalyticsEvent } from '$lib/analytics/internal-analytics';
	import { dashboardDataAccessForUser, visibleDashboardWidgetsForUser } from '$lib/rbac/dashboard-policy';

	type WebsiteContent = {
		id: string;
		title: string;
		slug: string;
		excerpt: string;
		content_html: string;
		cover_image_url: string;
		published_at: string | null;
	};

	let { data }: {
		data: {
			user?: { role?: string; roles?: string[]; permissions?: string[]; employee_id?: string; student_id?: string; parent_id?: string };
			publicHome?: {
				posts: WebsiteContent[];
				featuredPosts: WebsiteContent[];
				announcements: WebsiteContent[];
				profil: WebsiteContent | null;
				ppdbInfo: WebsiteContent | null;
			};
		};
	} = $props();

	interface AcademicStats {
		total_students: number;
		total_classes: number;
		total_subjects: number;
		total_years: number;
	}
	interface GuruStats {
		active_sessions: number;
		ungraded_essays: number;
		my_students: number;
		my_subjects: number;
	}
	interface TimetableEntry {
		id: string;
		day_of_week: number;
		start_time: string;
		end_time: string;
		room_label: string;
		notes: string;
		class_name: string;
		class_code: string;
		subject_name: string;
		subject_code: string;
		teacher_name: string;
	}
	interface StudentPortalData {
		student: {
			nama: string;
			nis: string;
			class_name: string;
			class_code: string;
			status: string;
			parent_phone: string;
		};
		parents: Array<{ id: string; nama: string; phone: string }>;
		sessions: Array<{
			participant_id: string;
			session_title: string;
			session_status: string;
			scheduled_start: string;
			room_name: string;
			seat_no: number | null;
			score: string | number | null;
		}>;
		timetable: TimetableEntry[];
	}
	interface ParentPortalData {
		parent: { nama: string; phone: string; address: string };
		children: Array<{ id: string; nama: string; nis: string; class_name: string }>;
		timetable: Array<TimetableEntry & { student_id: string; student_name: string }>;
	}

	interface CbtSessionSummary {
		id: string;
		status: string;
		package_title?: string;
	}

	interface EssayQueueItem {
		id?: string;
	}

	interface StudentSummary {
		id?: string;
	}

	interface DashboardPayload {
		academicStats: AcademicStats | null;
		bankSoal: BankSoalDashboard | null;
		guruStats: GuruStats | null;
		guruTimetable: TimetableEntry[];
		studentPortal: StudentPortalData | null;
		parentPortal: ParentPortalData | null;
	}

	interface RoleHomeAction {
		label: string;
		href: string;
		description: string;
		variant?: 'default' | 'outline';
	}

	interface RoleHomeSection {
		eyebrow: string;
		title: string;
		description: string;
		primary: RoleHomeAction[];
		secondary: RoleHomeAction[];
		watchlist: string[];
	}

	interface BankSoalDashboard {
		canRead: boolean;
		canCreate: boolean;
		canReview: boolean;
		canImport: boolean;
		canAnalytics: boolean;
		canSettings: boolean;
	}

	let dashboardPromise = $state<Promise<DashboardPayload> | null>(null);
	let dashboardRefreshBusy = $state(false);

	const roles = $derived(data.user?.roles || (data.user?.role ? [data.user.role] : []));
	const isGuru = $derived(roles.includes('guru'));
	const isSiswa = $derived(roles.includes('siswa'));
	const isParent = $derived(roles.includes('ortu'));
	const isAdmin = $derived(roles.includes('admin'));
	const isStaff = $derived(roles.includes('staf'));
	const dashboardAccess = $derived(dashboardDataAccessForUser(data.user));
	const dashboardWidgets = $derived(visibleDashboardWidgetsForUser(data.user));
	const hasTeacherDashboard = $derived(dashboardWidgets.some((widget) => widget.id.startsWith('teacher-')));
	const hasBankSoalDashboard = $derived(dashboardWidgets.some((widget) => widget.id.startsWith('bank-soal-')));
	const dashboardEyebrow = $derived.by(() => {
		if (isGuru) return 'Ruang Kerja Guru';
		if (isSiswa) return 'Portal Siswa';
		if (isParent) return 'Portal Orang Tua';
		if (isStaff) return 'Ruang Kerja Staf';
		if (hasBankSoalDashboard) return 'Bank Soal';
		return 'Pusat Operasi Madrasah';
	});
	const dashboardTitle = $derived.by(() => {
		if (isGuru && !hasTeacherDashboard && hasBankSoalDashboard) return 'Beranda Bank Soal';
		if (isGuru) return 'Beranda Guru';
		if (isSiswa) return 'Beranda Siswa';
		if (isParent) return 'Beranda Orang Tua';
		if (isStaff) return 'Beranda Staf';
		if (hasBankSoalDashboard && !dashboardAccess.academicStats) return 'Beranda Bank Soal';
		return 'Beranda Utama';
	});
	const dashboardDescription = $derived.by(() => {
		if (isGuru && !hasTeacherDashboard && hasBankSoalDashboard) return 'Pintasan penyusunan dan pengelolaan Bank Soal sesuai hak akses yang aktif pada akun ini.';
		if (isGuru) return 'Ringkasan kelas, aktivitas CBT, jadwal mengajar, dan pekerjaan koreksi yang perlu diperhatikan hari ini.';
		if (isSiswa) return 'Lihat identitas akademik, sesi ujian yang terdaftar, jadwal belajar, dan informasi wali yang terhubung.';
		if (isParent) return 'Pantau data putra-putri yang terhubung, jadwal anak, dan informasi dasar wali dari satu tempat.';
		if (isStaff) return 'Akses cepat ke layanan operasional sekolah, dokumen, arsip, perpustakaan, dan data akademik pendukung.';
		if (hasBankSoalDashboard && !dashboardAccess.academicStats) return 'Pintasan penyusunan dan pengelolaan Bank Soal sesuai hak akses yang aktif pada akun ini.';
		return 'Ringkasan akademik dan operasional MTs Negeri 2 Kolaka Utara untuk pengambilan keputusan harian.';
	});
	const dashboardRoleLabel = $derived(roles.length > 0 ? roles.join(' / ') : 'pengguna');

	const roleHome = $derived.by<RoleHomeSection>(() => {
		if (isSiswa) {
			return {
				eyebrow: 'Beranda Siswa',
				title: 'Mulai dari jadwal, ujian, dan data akademik pribadi.',
				description: 'Ruang ini memprioritaskan informasi yang langsung dibutuhkan siswa setelah login.',
				primary: [
					{ label: 'Portal Siswa', href: '/portal/siswa', description: 'Buka data profil, kelas, dan informasi akademik siswa.' },
					{ label: 'Portal Ujian Web', href: '/ujian', description: 'Jalur resmi ujian siswa tahun ini tanpa instalasi APK.', variant: 'outline' }
				],
				secondary: [
					{ label: 'Arsip APK CBT', href: '/asesmen/aplikasi-siswa/release', description: 'Arsip APK nonaktif/tahap lanjutan, bukan instruksi utama ujian.' }
				],
				watchlist: ['Cek jadwal belajar terbaru.', 'Pastikan sesi ujian dan ruang CBT sudah benar.', 'Hubungi wali kelas jika data profil belum sesuai.']
			};
		}
		if (isParent) {
			return {
				eyebrow: 'Beranda Orang Tua',
				title: 'Pantau informasi anak dari satu tempat.',
				description: 'Dirancang agar orang tua cepat melihat anak terhubung, jadwal, dan informasi penting madrasah.',
				primary: [
					{ label: 'Portal Orang Tua', href: '/portal/orang-tua', description: 'Pantau data anak dan jadwal yang terhubung.' },
					{ label: 'Portal Ujian Web', href: '/ujian', description: 'Jalur resmi ujian siswa tahun ini; tidak perlu instalasi APK.', variant: 'outline' }
				],
				secondary: [
					{ label: 'Pengumuman', href: '/pengumuman', description: 'Lihat pengumuman resmi madrasah.' }
				],
				watchlist: ['Pastikan semua anak sudah terhubung ke akun orang tua.', 'Cek jadwal anak secara berkala.', 'Simpan informasi Portal Ujian Web jika madrasah membuka asesmen.']
			};
		}
		if (isGuru) {
			return {
				eyebrow: 'Beranda Guru',
				title: 'Fokus ke jurnal, nilai, bank soal, dan asesmen.',
				description: 'Pintasan ini mengikuti pekerjaan harian guru agar tidak perlu mencari menu di sidebar panjang.',
				primary: [
					{ label: 'Jurnal Kelas', href: '/journal', description: 'Isi atau cek jurnal pembelajaran hari ini.' },
					{ label: 'Input Nilai', href: '/grades', description: 'Kelola nilai siswa sesuai mapel dan kelas.' },
					{ label: 'Bank Soal', href: '/bank-soal', description: 'Susun, cek, dan gunakan soal pembelajaran.' }
				],
				secondary: [
					{ label: 'Jadwal Mengajar', href: '/akademik/jadwal', description: 'Lihat slot jadwal kelas dan mapel.' },
					{ label: 'Rapor Siswa', href: '/grades/rapor', description: 'Cek nilai akhir dan deskripsi capaian.' },
					{ label: 'Persiapan Asesmen', href: '/asesmen/persiapan', description: 'Cek kesiapan paket dan sesi asesmen.' }
				],
				watchlist: ['Jurnal kelas yang belum diisi.', 'Nilai atau esai yang belum lengkap.', 'Jadwal mengajar dan asesmen aktif.']
			};
		}
		if (isStaff) {
			return {
				eyebrow: 'Beranda Staf',
				title: 'Prioritas layanan Tata Usaha, dokumen, arsip, dan aset.',
				description: 'Staf langsung diarahkan ke pekerjaan operasional yang paling sering digunakan.',
				primary: [
					{ label: 'Ringkasan Tata Usaha', href: '/tu', description: 'Ringkasan layanan surat dan administrasi.' },
					{ label: 'Monitoring Dokumen', href: '/document-cycles', description: 'Pantau siklus dan status dokumen.' },
					{ label: 'Inventaris', href: '/inventory', description: 'Cek aset dan daftar barang madrasah.' }
				],
				secondary: [
					{ label: 'Perpustakaan', href: '/library', description: 'Kelola layanan perpustakaan.' },
					{ label: 'Verifikasi Dokumen', href: '/document-cycles/verifikasi', description: 'Tindak lanjuti dokumen yang perlu verifikasi.' }
				],
				watchlist: ['Surat atau dokumen yang perlu ditindaklanjuti.', 'Barang inventaris yang perlu perhatian.', 'Arsip yang perlu dilengkapi.']
			};
		}
		return {
			eyebrow: 'Beranda Admin',
			title: 'Pantau kesiapan akademik, rapor, CBT, dan layanan madrasah.',
			description: 'Admin mendapat ringkasan prioritas untuk mengawasi operasional utama MTsN 2 Kolaka Utara.',
			primary: [
				{ label: 'Kesiapan Akademik & Rapor', href: '/akademik/kesiapan', description: 'Cek masalah wali kelas, jadwal, nilai, dan rapor.' },
				{ label: 'Rombel', href: '/akademik/rombel', description: 'Kelola kelas, wali kelas, dan siswa per rombel.' },
				{ label: 'Persiapan Asesmen', href: '/asesmen/persiapan', description: 'Pantau kesiapan kegiatan CBT.' }
			],
			secondary: [
				{ label: 'Jadwal', href: '/akademik/jadwal', description: 'Cek jadwal dan potensi bentrok.' },
				{ label: 'Rapor Siswa', href: '/grades/rapor', description: 'Kelola pengaturan dan cetak rapor.' },
				{ label: 'Monitor PUSAKA', href: '/pusaka', description: 'Pantau integrasi kehadiran pegawai.' }
			],
			watchlist: ['Kesiapan akademik dan rapor yang belum lengkap.', 'Kegiatan CBT mendekati pelaksanaan.', 'Sinkronisasi PUSAKA dan tindak lanjut Tata Usaha.']
		};
	});

	async function fetchJSON<T>(path: string): Promise<T> {
		const res = await fetch(path);
		return readClientApiData<T>(res, `Respons ${path} tidak valid.`);
	}

	async function loadUngradedEssaysForSessions(sessions: CbtSessionSummary[]): Promise<EssayQueueItem[]> {
		const essayLists = await Promise.all(
			sessions.map((session) => fetchJSON<EssayQueueItem[]>(clientApiPath`/api/asesmen/sessions/${session.id}/ungraded-essays`))
		);
		return essayLists.flat();
	}

	function fmtDateTime(iso: string) {
		if (!iso) return '—';
		return new Date(iso).toLocaleString('id-ID', {
			timeZone: 'Asia/Makassar',
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
		}) + ' WITA';
	}

	const dayLabels: Record<number, string> = {
		1: 'Senin',
		2: 'Selasa',
		3: 'Rabu',
		4: 'Kamis',
		5: 'Jumat',
		6: 'Sabtu',
	};

	function fmtTime(value: string) {
		return value.slice(0, 5);
	}

	function hasPermission(permission: string) {
		if (isAdmin) return true;
		return (data.user?.permissions ?? []).includes(permission);
	}

	function bankSoalDashboard(): BankSoalDashboard | null {
		if (!dashboardAccess.bankSoal) return null;
		return {
			canRead: hasPermission('bank_soal.read'),
			canCreate: hasPermission('bank_soal.create'),
			canReview: hasPermission('bank_soal.review') || hasPermission('bank_soal.publish'),
			canImport: hasPermission('bank_soal.import'),
			canAnalytics: hasPermission('bank_soal.analytics'),
			canSettings: hasPermission('bank_soal.settings')
		};
	}

	function emptyDashboardPayload(): DashboardPayload {
		return {
			academicStats: null,
			bankSoal: bankSoalDashboard(),
			guruStats: null,
			guruTimetable: [],
			studentPortal: null,
			parentPortal: null
		};
	}

	async function loadDashboard(): Promise<DashboardPayload> {
		const payload = emptyDashboardPayload();

		if (dashboardAccess.studentPortal) {
			payload.studentPortal = await fetchJSON<StudentPortalData>('/api/portal/student/me');
			return payload;
		}

		if (dashboardAccess.parentPortal) {
			payload.parentPortal = await fetchJSON<ParentPortalData>('/api/portal/parent/me');
			return payload;
		}

		const [sessions, students, timetable, academicStats] = await Promise.all([
			dashboardAccess.assessmentSessions ? fetchJSON<CbtSessionSummary[]>('/api/asesmen/sessions') : Promise.resolve([]),
			dashboardAccess.studentSummary ? fetchJSON<StudentSummary[]>('/api/students') : Promise.resolve([]),
			dashboardAccess.teacherTimetable ? fetchJSON<{ timetable: TimetableEntry[] }>('/api/portal/guru/timetable') : Promise.resolve({ timetable: [] }),
			dashboardAccess.academicStats ? fetchJSON<AcademicStats>('/api/academic/stats') : Promise.resolve(null)
		]);

		payload.academicStats = academicStats;
		payload.guruTimetable = timetable.timetable ?? [];

		if (dashboardAccess.assessmentSessions || dashboardAccess.studentSummary || dashboardAccess.teacherTimetable) {
			const essays = dashboardAccess.ungradedEssays ? await loadUngradedEssaysForSessions(sessions) : [];
			const activeSessions = sessions.filter((session) => session.status === 'active' || session.status === 'scheduled');
			const subjects = new Set(activeSessions.map((session) => session.package_title));
			payload.guruStats = {
				active_sessions: activeSessions.length,
				ungraded_essays: essays.length,
				my_students: students.length,
				my_subjects: subjects.size
			};
		}
		return payload;
	}

	function refreshDashboard() {
		dashboardPromise = loadDashboard();
		return dashboardPromise;
	}

	async function retryDashboard() {
		dashboardRefreshBusy = true;
		void trackInternalAnalyticsEvent('dashboard.refresh', {
			pathname: window.location.pathname,
			role: roles[0],
			metadata: { page_key: 'dashboard' }
		});
		try {
			await refreshDashboard();
		} finally {
			dashboardRefreshBusy = false;
		}
	}

	function dashboardErrorMessage(error: unknown) {
		return error instanceof Error ? error.message : 'Terjadi gangguan saat memuat beranda.';
	}

	function handleDashboardRenderError(error: unknown) {
		console.error('Beranda boundary error', error);
	}

	onMount(() => {
		if (data.user) {
			void trackInternalAnalyticsEvent('dashboard.view', {
				pathname: window.location.pathname,
				role: roles[0],
				metadata: { page_key: 'dashboard' }
			});
			refreshDashboard();
		}
	});
</script>

<svelte:head><title>{data.user ? 'Beranda — MTSN 2 Kolut' : 'MTs Negeri 2 Kolaka Utara — Website Resmi'}</title></svelte:head>

{#if !data.user}
	<PublicHome home={data.publicHome} />
{:else}
	<div class="space-y-6">
		<div class="overflow-hidden rounded-[1.75rem] border border-border bg-card shadow-sm">
			<div class="grid gap-5 p-5 md:grid-cols-[1fr_auto] md:p-6">
				<div>
					<p class="text-xs font-semibold uppercase tracking-[0.26em] text-primary">{dashboardEyebrow}</p>
					<h1 class="mt-3 text-2xl font-semibold tracking-tight text-foreground sm:text-3xl">{dashboardTitle}</h1>
					<p class="mt-2 max-w-3xl text-sm leading-6 text-muted-foreground">{dashboardDescription}</p>
					<div class="mt-4 flex flex-wrap gap-2">
						<Badge class="border-primary/20 bg-primary/10 text-primary">Role: {dashboardRoleLabel}</Badge>
						<Badge variant="outline">MTsN 2 Kolaka Utara</Badge>
						<Badge variant="outline">WITA</Badge>
					</div>
				</div>
				<div class="flex items-end md:min-w-48 md:justify-end">
					<LoadingButton
						variant="outline"
						loading={dashboardRefreshBusy}
						loadingLabel="Memuat..."
						onclick={() => void retryDashboard()}
						label="Refresh Beranda"
					/>
				</div>
			</div>
		</div>

		<Card.Root class="overflow-hidden border-primary/20 bg-gradient-to-br from-primary/10 via-card to-card shadow-sm">
			<Card.Content class="grid gap-5 p-5 lg:grid-cols-[1.15fr,0.85fr] lg:items-start md:p-6">
				<div class="space-y-4">
					<div>
						<p class="text-xs font-semibold uppercase tracking-[0.24em] text-primary">{roleHome.eyebrow}</p>
						<h2 class="mt-2 text-xl font-semibold tracking-tight text-foreground">{roleHome.title}</h2>
						<p class="mt-2 text-sm leading-6 text-muted-foreground">{roleHome.description}</p>
					</div>
					<div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
						{#each roleHome.primary as action (`primary-${action.href}`)}
							<a href={action.href} class="rounded-2xl border border-primary/20 bg-background/80 p-4 text-sm transition hover:-translate-y-0.5 hover:border-primary/40 hover:shadow-sm">
								<span class="font-semibold text-foreground">{action.label}</span>
								<span class="mt-1 block text-xs leading-5 text-muted-foreground">{action.description}</span>
							</a>
						{/each}
					</div>
					<div class="flex flex-wrap gap-2">
						{#each roleHome.secondary as action (`secondary-${action.href}`)}
							<Button size="sm" variant={action.variant ?? 'outline'} href={action.href}>{action.label}</Button>
						{/each}
					</div>
				</div>
				<div class="rounded-2xl border border-border bg-background/80 p-4">
					<p class="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">Perhatian hari ini</p>
					<ul class="mt-3 space-y-3 text-sm text-muted-foreground">
						{#each roleHome.watchlist as item (`watch-${item}`)}
							<li class="flex gap-2"><span class="mt-1 h-2 w-2 shrink-0 rounded-full bg-primary"></span><span>{item}</span></li>
						{/each}
					</ul>
				</div>
			</Card.Content>
		</Card.Root>

		<Card.Root class="overflow-hidden border-primary/20 bg-gradient-to-br from-primary/10 via-card to-card shadow-sm">
			<Card.Content class="grid gap-4 p-5 md:grid-cols-[1fr_auto] md:items-center md:p-6">
				<div class="space-y-3">
					<div class="flex flex-wrap items-center gap-2">
						<Badge class="border-primary/20 bg-primary/10 text-primary">Portal Ujian Web</Badge>
						<Badge variant="outline">Jalur resmi tahun ini</Badge>
					</div>
					<div>
						<h2 class="text-lg font-semibold tracking-tight text-foreground">Portal Ujian Web Siswa</h2>
						<p class="mt-1 max-w-2xl text-sm leading-6 text-muted-foreground">
							Arahkan siswa ke Portal Ujian Web /ujian. APK Flutter tetap tersedia sebagai arsip nonaktif/tahap lanjutan, bukan instruksi utama ujian tahun ini.
						</p>
					</div>
					<div class="flex flex-wrap gap-2">
						<Button size="sm" href="/ujian">Buka Portal Ujian Web</Button>
						<Button size="sm" variant="outline" href="/ujian?demo=1">Demo Lokal</Button>
						<Button size="sm" variant="outline" href="/asesmen/aplikasi-siswa/release">Arsip APK</Button>
					</div>
				</div>
				<div class="flex items-center gap-3 rounded-2xl border border-border bg-background/80 p-3 text-sm text-muted-foreground">
					<img src="/releases/mobile/latest-qr.svg" alt="QR arsip APK CBT Mobile" class="h-20 w-20 rounded-lg bg-white p-1" />
					<div class="hidden max-w-48 sm:block">
						<p class="font-semibold text-foreground">Arsip APK nonaktif</p>
						<p class="mt-1 break-all text-xs leading-5">Bukan jalur masuk ujian resmi tahun ini.</p>
					</div>
				</div>
			</Card.Content>
		</Card.Root>

		<AsyncContent promise={dashboardPromise} onerror={handleDashboardRenderError}>
		{#snippet pending()}
			{#if isSiswa}
				<div class="grid gap-4 lg:grid-cols-[1.2fr,0.8fr]">
					{#each Array.from({ length: 2 }) as _, index (`student-dashboard-skeleton-${index}`)}
						<Card.Root class="border-border">
							<Card.Header class="space-y-2">
								<Skeleton class="h-6 w-40" />
								<Skeleton class="h-4 w-56" />
							</Card.Header>
							<Card.Content class="space-y-3">
								<Skeleton class="h-6 w-32" />
								<Skeleton class="h-4 w-full" />
								<Skeleton class="h-16 w-full rounded-xl" />
							</Card.Content>
						</Card.Root>
					{/each}
				</div>
			{:else if isParent}
				<div class="grid gap-4 lg:grid-cols-[0.9fr,1.1fr]">
					{#each Array.from({ length: 2 }) as _, index (`parent-dashboard-skeleton-${index}`)}
						<Card.Root class="border-border">
							<Card.Header class="space-y-2">
								<Skeleton class="h-6 w-36" />
								<Skeleton class="h-4 w-44" />
							</Card.Header>
							<Card.Content class="space-y-3">
								<Skeleton class="h-4 w-full" />
								<Skeleton class="h-4 w-4/5" />
								<Skeleton class="h-16 w-full rounded-xl" />
							</Card.Content>
						</Card.Root>
					{/each}
				</div>
			{:else if isGuru && hasTeacherDashboard}
				<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
					{#each Array.from({ length: 4 }) as _, index (`guru-stat-skeleton-${index}`)}
						<Card.Root class="border-border">
							<Card.Content class="space-y-2 pt-4">
								<Skeleton class="h-4 w-28" />
								<Skeleton class="h-8 w-16" />
							</Card.Content>
						</Card.Root>
					{/each}
				</div>
		{:else if isAdmin}
				<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
					{#each Array.from({ length: 4 }) as _, index (`admin-stat-skeleton-${index}`)}
						<Card.Root class="border-border">
							<Card.Content class="space-y-2 pt-4">
								<Skeleton class="h-4 w-28" />
								<Skeleton class="h-8 w-16" />
							</Card.Content>
						</Card.Root>
					{/each}
				</div>
			{/if}
		{/snippet}

		{#snippet failed(error, reset)}
			<Card.Root class="border-warning/30 bg-warning/10">
				<Card.Header>
					<Card.Title class="text-base text-warning">Beranda belum berhasil dimuat</Card.Title>
					<Card.Description class="text-warning">
						{dashboardErrorMessage(error)}
					</Card.Description>
				</Card.Header>
				<Card.Content class="flex flex-wrap gap-2">
					<LoadingButton onclick={() => void retryDashboard()} loading={dashboardRefreshBusy} loadingLabel="Memuat...">Coba Lagi</LoadingButton>
					{#if reset}
						<Button variant="outline" onclick={reset}>Muat Ulang Tampilan</Button>
					{/if}
				</Card.Content>
			</Card.Root>
		{/snippet}

		{#snippet children(value)}
			{@const dashboard = value as DashboardPayload}
			{@const studentPortal = dashboard.studentPortal}
			{@const parentPortal = dashboard.parentPortal}
			{@const guruStats = dashboard.guruStats}
			{@const guruTimetable = dashboard.guruTimetable}
			{@const academicStats = dashboard.academicStats}
			{@const bankSoal = dashboard.bankSoal}
			{@const parentTimetableByChild = parentPortal
				? parentPortal.children.map((child) => ({
						child,
						slots: parentPortal.timetable.filter((slot) => slot.student_id === child.id),
					}))
				: []}

			{#if isSiswa && studentPortal}
				<div class="grid gap-4 lg:grid-cols-[1.2fr,0.8fr]">
			<Card.Root class="border-primary/20">
				<Card.Header>
					<Card.Title class="text-base">Profil Akademik</Card.Title>
					<Card.Description>{studentPortal.student.nama} · NIS {studentPortal.student.nis}</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-3">
					<div class="flex flex-wrap gap-2">
						<Badge variant="outline">{studentPortal.student.class_code || studentPortal.student.class_name || 'Belum ada kelas'}</Badge>
						<Badge class={studentPortal.student.status === 'active' ? 'bg-primary/15 text-primary border-primary/20' : 'bg-muted text-muted-foreground'}>{studentPortal.student.status}</Badge>
					</div>
					<p class="text-sm text-muted-foreground">Kontak wali utama: {studentPortal.student.parent_phone || 'belum diisi'}</p>
					{#if studentPortal.parents.length > 0}
						<div class="rounded-xl border border-border bg-muted/50 p-3">
							<p class="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">Wali Terhubung</p>
							<div class="mt-2 space-y-2">
								{#each studentPortal.parents as parent (parent.id)}
									<div class="flex items-center justify-between gap-3 text-sm">
										<span class="font-medium text-foreground">{parent.nama}</span>
										<span class="text-muted-foreground">{parent.phone || '—'}</span>
									</div>
								{/each}
							</div>
						</div>
					{/if}
				</Card.Content>
			</Card.Root>

			<Card.Root class="border-border">
				<Card.Header>
					<Card.Title class="text-base">Ujian Saya</Card.Title>
					<Card.Description>Sesi terbaru yang terdaftar untuk akun ini.</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-3">
					{#if studentPortal.sessions.length > 0}
						{#each studentPortal.sessions.slice(0, 4) as session (session.participant_id)}
							<div class="rounded-xl border border-border bg-card p-3">
								<div class="flex items-start justify-between gap-3">
									<p class="text-sm font-medium text-foreground">{session.session_title}</p>
									<Badge variant="outline">{session.session_status}</Badge>
								</div>
								<p class="mt-2 text-xs text-muted-foreground">{fmtDateTime(session.scheduled_start)}</p>
								<p class="mt-1 text-xs text-muted-foreground">Ruang {session.room_name || '—'} · Meja {session.seat_no ?? '—'}</p>
							</div>
						{/each}
					{:else}
						<EmptyStatePanel
							compact
							title="Belum ada sesi ujian"
							description="Sesi ujian yang terhubung ke akun siswa akan tampil di sini setelah peserta didaftarkan."
						/>
					{/if}
				</Card.Content>
			</Card.Root>
				</div>

				<Card.Root class="border-border">
			<Card.Header>
				<Card.Title class="text-base">Jadwal Pelajaran Minggu Ini</Card.Title>
				<Card.Description>Slot belajar yang tersusun untuk kelas {studentPortal.student.class_code || studentPortal.student.class_name || 'aktif'}.</Card.Description>
			</Card.Header>
			<Card.Content class="space-y-3">
				{#if studentPortal.timetable.length > 0}
					<div class="grid gap-3 lg:grid-cols-2">
						{#each studentPortal.timetable as slot (slot.id)}
							<div class="rounded-xl border border-border bg-card p-3">
								<div class="flex items-start justify-between gap-3">
									<div>
										<p class="text-sm font-medium text-foreground">{slot.subject_name}</p>
										<p class="mt-1 text-xs text-muted-foreground">{slot.teacher_name}</p>
									</div>
									<Badge variant="outline">{dayLabels[slot.day_of_week] ?? `Hari ${slot.day_of_week}`}</Badge>
								</div>
								<p class="mt-2 text-xs text-muted-foreground">{fmtTime(slot.start_time)}–{fmtTime(slot.end_time)} · {slot.room_label || 'Ruang belum diisi'}</p>
								{#if slot.notes}
									<p class="mt-1 text-xs text-muted-foreground">{slot.notes}</p>
								{/if}
							</div>
						{/each}
					</div>
				{:else}
					<EmptyStatePanel
						compact
						title="Jadwal pelajaran belum tersedia"
						description="Jadwal mingguan akan tampil di sini setelah operator akademik menyusun slot kelas untuk siswa ini."
					/>
				{/if}
			</Card.Content>
				</Card.Root>
			{/if}

			{#if isParent && parentPortal}
				<div class="grid gap-4 lg:grid-cols-[0.9fr,1.1fr]">
			<Card.Root class="border-primary/20">
				<Card.Header>
					<Card.Title class="text-base">Profil Wali</Card.Title>
					<Card.Description>{parentPortal.parent.nama}</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-2 text-sm text-muted-foreground">
					<p>Nomor HP: {parentPortal.parent.phone || 'belum diisi'}</p>
					<p>Alamat: {parentPortal.parent.address || 'belum diisi'}</p>
				</Card.Content>
			</Card.Root>

			<Card.Root class="border-border">
				<Card.Header>
					<Card.Title class="text-base">Anak Terhubung</Card.Title>
					<Card.Description>Data siswa yang sudah ditautkan ke akun orang tua ini.</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-3">
					{#if parentPortal.children.length > 0}
						{#each parentPortal.children as child (child.id)}
							<div class="rounded-xl border border-border bg-card p-3">
								<div class="flex items-start justify-between gap-3">
									<div>
										<p class="text-sm font-medium text-foreground">{child.nama}</p>
										<p class="mt-1 text-xs text-muted-foreground">NIS {child.nis}</p>
									</div>
									<Badge variant="outline">{child.class_name || 'Belum ada kelas'}</Badge>
								</div>
							</div>
						{/each}
					{:else}
						<EmptyStatePanel
							compact
							title="Belum ada anak yang terhubung"
							description="Minta admin sekolah menautkan akun orang tua ini ke data siswa agar informasi anak bisa tampil di dasbor."
						/>
					{/if}
				</Card.Content>
			</Card.Root>
				</div>

				<Card.Root class="border-border">
			<Card.Header>
				<Card.Title class="text-base">Jadwal Anak</Card.Title>
				<Card.Description>Ringkasan slot pelajaran untuk setiap anak yang sudah terhubung ke akun orang tua ini.</Card.Description>
			</Card.Header>
			<Card.Content class="space-y-4">
				{#if parentTimetableByChild.length > 0}
					{#each parentTimetableByChild as item (`parent-timetable-${item.child.id}`)}
						<div class="rounded-2xl border border-border bg-muted/50 p-4">
							<div class="flex items-start justify-between gap-3">
								<div>
									<p class="text-sm font-semibold text-foreground">{item.child.nama}</p>
									<p class="mt-1 text-xs text-muted-foreground">{item.child.class_name || 'Belum ada kelas'}</p>
								</div>
								<Badge variant="outline">{item.slots.length} slot</Badge>
							</div>

							{#if item.slots.length > 0}
								<div class="mt-3 grid gap-3 lg:grid-cols-2">
									{#each item.slots as slot (slot.id)}
										<div class="rounded-xl border border-border bg-card p-3">
											<div class="flex items-start justify-between gap-3">
												<div>
													<p class="text-sm font-medium text-foreground">{slot.subject_name}</p>
													<p class="mt-1 text-xs text-muted-foreground">{slot.teacher_name}</p>
												</div>
												<Badge variant="outline">{dayLabels[slot.day_of_week] ?? `Hari ${slot.day_of_week}`}</Badge>
											</div>
											<p class="mt-2 text-xs text-muted-foreground">{fmtTime(slot.start_time)}–{fmtTime(slot.end_time)} · {slot.room_label || 'Ruang belum diisi'}</p>
											{#if slot.notes}
												<p class="mt-1 text-xs text-muted-foreground">{slot.notes}</p>
											{/if}
										</div>
									{/each}
								</div>
							{:else}
								<div class="mt-3 rounded-xl border border-dashed border-border bg-card px-3 py-4 text-sm text-muted-foreground">
									Jadwal untuk anak ini belum tersedia.
								</div>
							{/if}
						</div>
					{/each}
				{:else}
					<EmptyStatePanel
						compact
						title="Jadwal anak belum tersedia"
						description="Jadwal akan muncul di sini setelah data anak terhubung dan operator akademik menyusun slot kelasnya."
					/>
				{/if}
			</Card.Content>
				</Card.Root>
			{/if}

			{#if isGuru && guruStats}
				<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
			<Card.Root class="border-success/20"><Card.Content class="pt-4"><p class="text-xs font-medium uppercase tracking-wider text-muted-foreground">Sesi CBT Berjalan</p><p class="mt-1 text-3xl font-bold text-success">{guruStats.active_sessions}</p></Card.Content></Card.Root>
			<Card.Root class="border-warning/30"><Card.Content class="pt-4"><p class="text-xs font-medium uppercase tracking-wider text-muted-foreground">Esai Belum Dikoreksi</p><p class="mt-1 text-3xl font-bold text-warning">{guruStats.ungraded_essays}</p></Card.Content></Card.Root>
			<Card.Root class="border-success/20"><Card.Content class="pt-4"><p class="text-xs font-medium uppercase tracking-wider text-muted-foreground">Siswa Terpantau</p><p class="mt-1 text-3xl font-bold text-success">{guruStats.my_students}</p></Card.Content></Card.Root>
			<Card.Root class="border-success/20"><Card.Content class="pt-4"><p class="text-xs font-medium uppercase tracking-wider text-muted-foreground">Mapel Diampu</p><p class="mt-1 text-3xl font-bold text-success">{guruStats.my_subjects}</p></Card.Content></Card.Root>
				</div>

				<Card.Root class="border-border">
			<Card.Header>
				<Card.Title class="text-base">Jadwal Mengajar</Card.Title>
				<Card.Description>Ringkasan slot kelas-mapel yang sudah dijadwalkan untuk guru pada minggu berjalan.</Card.Description>
			</Card.Header>
			<Card.Content class="space-y-3">
				{#if guruTimetable.length > 0}
					<div class="grid gap-3 lg:grid-cols-2">
						{#each guruTimetable as slot (slot.id)}
							<div class="rounded-xl border border-border bg-card p-3">
								<div class="flex items-start justify-between gap-3">
									<div>
										<p class="text-sm font-medium text-foreground">{slot.subject_name}</p>
										<p class="mt-1 text-xs text-muted-foreground">{slot.class_code} · {slot.class_name}</p>
									</div>
									<Badge variant="outline">{dayLabels[slot.day_of_week] ?? `Hari ${slot.day_of_week}`}</Badge>
								</div>
								<p class="mt-2 text-xs text-muted-foreground">{fmtTime(slot.start_time)}–{fmtTime(slot.end_time)} · {slot.room_label || 'Ruang belum diisi'}</p>
								{#if slot.notes}
									<p class="mt-1 text-xs text-muted-foreground">{slot.notes}</p>
								{/if}
							</div>
						{/each}
					</div>
				{:else}
					<EmptyStatePanel
						compact
						title="Jadwal mengajar belum tersedia"
						description="Slot jadwal guru akan tampil di sini setelah operator akademik menyusun jadwal pada assignment kelas-mapel yang diampu."
					/>
				{/if}
			</Card.Content>
				</Card.Root>
			{/if}

			{#if bankSoal}
				<Card.Root class="border-primary/20">
					<Card.Header>
						<Card.Title class="text-base">Bank Soal</Card.Title>
						<Card.Description>Shortcut yang tersedia mengikuti permission Bank Soal pada akun ini.</Card.Description>
					</Card.Header>
					<Card.Content>
						<div class="flex flex-wrap gap-2">
							{#if bankSoal.canRead}
								<Button variant="default" size="sm" href="/bank-soal">Ringkasan Bank Soal</Button>
								<Button variant="outline" size="sm" href="/bank-soal/daftar">Daftar Soal</Button>
								<Button variant="outline" size="sm" href="/bank-soal/mapel-kd">Mapel & KD</Button>
							{/if}
							{#if bankSoal.canCreate}<Button variant="outline" size="sm" href="/bank-soal/tambah">Tambah Soal</Button>{/if}
							{#if bankSoal.canReview}<Button variant="outline" size="sm" href="/bank-soal/verifikasi">Verifikasi Soal</Button>{/if}
							{#if bankSoal.canImport}<Button variant="outline" size="sm" href="/bank-soal/impor">Impor Soal</Button>{/if}
							{#if bankSoal.canAnalytics}<Button variant="outline" size="sm" href="/bank-soal/analisis-butir">Analisis Butir</Button>{/if}
							{#if bankSoal.canSettings}<Button variant="outline" size="sm" href="/bank-soal/pengaturan">Pengaturan</Button>{/if}
						</div>
					</Card.Content>
				</Card.Root>
			{/if}

			{#if isAdmin}
				<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
			<Card.Root class="border-success/20"><Card.Content class="pt-4"><p class="text-xs font-medium uppercase tracking-wider text-muted-foreground">Total Siswa</p><p class="mt-1 text-3xl font-bold text-success">{academicStats?.total_students ?? '—'}</p></Card.Content></Card.Root>
			<Card.Root class="border-success/20"><Card.Content class="pt-4"><p class="text-xs font-medium uppercase tracking-wider text-muted-foreground">Kelas Aktif</p><p class="mt-1 text-3xl font-bold text-success">{academicStats?.total_classes ?? '—'}</p></Card.Content></Card.Root>
			<Card.Root class="border-success/20"><Card.Content class="pt-4"><p class="text-xs font-medium uppercase tracking-wider text-muted-foreground">Mapel Aktif</p><p class="mt-1 text-3xl font-bold text-success">{academicStats?.total_subjects ?? '—'}</p></Card.Content></Card.Root>
			<Card.Root class="border-success/20"><Card.Content class="pt-4"><p class="text-xs font-medium uppercase tracking-wider text-muted-foreground">Tahun Ajaran</p><p class="mt-1 text-3xl font-bold text-success">{academicStats?.total_years ?? '—'}</p></Card.Content></Card.Root>
				</div>
			{/if}

			{#if isStaff}
				<Card.Root class="border-border">
			<Card.Header>
				<Card.Title class="text-base">Akses Cepat Staf</Card.Title>
				<Card.Description>Menu yang paling sering dipakai untuk layanan data orang tua dan siswa.</Card.Description>
			</Card.Header>
			<Card.Content>
				<div class="flex flex-wrap gap-2">
					<Button variant="default" size="sm" href="/tu">Layanan Tata Usaha</Button>
					<Button variant="outline" size="sm" href="/library">Perpustakaan</Button>
					<Button variant="outline" size="sm" href="/inventory">Inventaris</Button>
				</div>
			</Card.Content>
				</Card.Root>
			{/if}

			{#if isAdmin}
				<Card.Root class="border-border">
			<Card.Header class="pb-3">
				<div class="flex items-center justify-between">
					<div>
						<Card.Title class="text-base">Integrasi PUSAKA Kemenag</Card.Title>
						<Card.Description>Sinkronisasi data kehadiran pegawai dari sistem pemerintah.</Card.Description>
					</div>
					<Badge variant="outline" class="text-xs">Eksternal</Badge>
				</div>
			</Card.Header>
			<Card.Content>
				<div class="flex flex-wrap gap-2">
					<Button variant="default" size="sm" href="/pusaka">Monitor PUSAKA</Button>
					<Button variant="outline" size="sm" href="/pusaka/kehadiran">Data Kehadiran</Button>
					<Button variant="outline" size="sm" href="/pusaka/summary">Ringkasan</Button>
					<Button variant="outline" size="sm" href="/pusaka/antrian">Antrian Sinkronisasi</Button>
				</div>
			</Card.Content>
				</Card.Root>
			{/if}
		{/snippet}
	</AsyncContent>
</div>
{/if}
