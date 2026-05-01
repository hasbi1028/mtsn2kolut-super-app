<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import PublicHome from '$lib/components/PublicHome.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';

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
			user?: { role?: string; roles?: string[]; employee_id?: string; student_id?: string; parent_id?: string };
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
	}
	interface ParentPortalData {
		parent: { nama: string; phone: string; address: string };
		children: Array<{ id: string; nama: string; nis: string; class_name: string }>;
	}

	let academicStats = $state<AcademicStats | null>(null);
	let guruStats = $state<GuruStats | null>(null);
	let studentPortal = $state<StudentPortalData | null>(null);
	let parentPortal = $state<ParentPortalData | null>(null);

	const roles = $derived(data.user?.roles || (data.user?.role ? [data.user.role] : []));
	const isGuru = $derived(roles.includes('guru'));
	const isSiswa = $derived(roles.includes('siswa'));
	const isParent = $derived(roles.includes('ortu'));
	const isAdmin = $derived(roles.includes('admin'));
	const isStaff = $derived(roles.includes('staf'));
	const dashboardLoading = $derived(
		(isGuru && !guruStats) ||
		(isSiswa && !studentPortal) ||
		(isParent && !parentPortal) ||
		((isAdmin || isStaff) && !academicStats)
	);

	function parseData<T>(raw: unknown): T | null {
		if (!raw || typeof raw !== 'object') return null;
		const wrapper = raw as { data?: T };
		return (wrapper.data ?? raw) as T;
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

	async function load() {
		try {
			if (isGuru) {
				const [sessRes, essaysRes, studentsRes] = await Promise.all([
					fetch('/api/cbt/sessions'),
					fetch('/api/cbt/sessions/my-essays'),
					fetch('/api/students'),
				]);
				const sessions = parseData<any[]>(await sessRes.json().catch(() => null)) ?? [];
				const essays = parseData<any[]>(await essaysRes.json().catch(() => null)) ?? [];
				const students = parseData<any[]>(await studentsRes.json().catch(() => null)) ?? [];
				const activeSessions = sessions.filter((session) => session.status === 'active' || session.status === 'scheduled');
				const subjects = new Set(activeSessions.map((session) => session.package_title));
				guruStats = {
					active_sessions: activeSessions.length,
					ungraded_essays: essays.length,
					my_students: students.length,
					my_subjects: subjects.size,
				};
				return;
			}

			if (isSiswa) {
				const res = await fetch('/api/portal/student/me');
				studentPortal = parseData<StudentPortalData>(await res.json().catch(() => null));
				return;
			}

			if (isParent) {
				const res = await fetch('/api/portal/parent/me');
				parentPortal = parseData<ParentPortalData>(await res.json().catch(() => null));
				return;
			}

			if (isAdmin || isStaff) {
				const res = await fetch('/api/academic/stats');
				academicStats = parseData<AcademicStats>(await res.json().catch(() => null));
			}
		} catch {
			// Keep dashboard resilient; individual cards can stay empty.
		}
	}

	onMount(() => {
		if (data.user) {
			void load();
		}
	});
</script>

<svelte:head><title>{data.user ? 'Dashboard — MTSN 2 Kolut' : 'MTs Negeri 2 Kolaka Utara — Website Resmi'}</title></svelte:head>

{#if !data.user}
	<PublicHome home={data.publicHome} />
{:else}
<div class="space-y-6">
	<div>
		{#if isGuru}
			<h1 class="text-2xl font-semibold text-slate-800">Dasbor Guru</h1>
			<p class="mt-1 text-sm text-muted-foreground">Ringkasan kelas, aktivitas CBT, dan pekerjaan koreksi yang perlu diperhatikan hari ini.</p>
		{:else if isSiswa}
			<h1 class="text-2xl font-semibold text-slate-800">Dasbor Siswa</h1>
			<p class="mt-1 text-sm text-muted-foreground">Lihat identitas akademik, sesi ujian yang terdaftar, dan informasi wali yang terhubung.</p>
		{:else if isParent}
			<h1 class="text-2xl font-semibold text-slate-800">Dasbor Orang Tua</h1>
			<p class="mt-1 text-sm text-muted-foreground">Pantau data putra-putri yang terhubung dan informasi dasar wali dari satu tempat.</p>
		{:else if isStaff}
			<h1 class="text-2xl font-semibold text-slate-800">Dasbor Staf</h1>
			<p class="mt-1 text-sm text-muted-foreground">Akses cepat ke data operasional yang paling sering dipakai untuk layanan sekolah.</p>
		{:else}
			<h1 class="text-2xl font-semibold text-slate-800">Dasbor Utama</h1>
			<p class="mt-1 text-sm text-muted-foreground">Ringkasan akademik dan operasional MTs Negeri 2 Kolaka Utara.</p>
		{/if}
	</div>

	{#if isSiswa && dashboardLoading}
		<div class="grid gap-4 lg:grid-cols-[1.2fr,0.8fr]">
			{#each Array.from({ length: 2 }) as _, index (`student-dashboard-skeleton-${index}`)}
				<Card.Root class="border-slate-200">
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
	{:else if isSiswa && studentPortal}
		<div class="grid gap-4 lg:grid-cols-[1.2fr,0.8fr]">
			<Card.Root class="border-emerald-100">
				<Card.Header>
					<Card.Title class="text-base">Profil Akademik</Card.Title>
					<Card.Description>{studentPortal.student.nama} · NIS {studentPortal.student.nis}</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-3">
					<div class="flex flex-wrap gap-2">
						<Badge variant="outline">{studentPortal.student.class_code || studentPortal.student.class_name || 'Belum ada kelas'}</Badge>
						<Badge class={studentPortal.student.status === 'active' ? 'bg-emerald-100 text-emerald-700 border-emerald-200' : 'bg-slate-100 text-slate-600'}>{studentPortal.student.status}</Badge>
					</div>
					<p class="text-sm text-slate-600">Kontak wali utama: {studentPortal.student.parent_phone || 'belum diisi'}</p>
					{#if studentPortal.parents.length > 0}
						<div class="rounded-xl border border-slate-200 bg-slate-50 p-3">
							<p class="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">Wali Terhubung</p>
							<div class="mt-2 space-y-2">
								{#each studentPortal.parents as parent (parent.id)}
									<div class="flex items-center justify-between gap-3 text-sm">
										<span class="font-medium text-slate-900">{parent.nama}</span>
										<span class="text-slate-500">{parent.phone || '—'}</span>
									</div>
								{/each}
							</div>
						</div>
					{/if}
				</Card.Content>
			</Card.Root>

			<Card.Root class="border-slate-200">
				<Card.Header>
					<Card.Title class="text-base">Ujian Saya</Card.Title>
					<Card.Description>Sesi terbaru yang terdaftar untuk akun ini.</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-3">
					{#if studentPortal.sessions.length > 0}
						{#each studentPortal.sessions.slice(0, 4) as session (session.participant_id)}
							<div class="rounded-xl border border-slate-200 bg-white p-3">
								<div class="flex items-start justify-between gap-3">
									<p class="text-sm font-medium text-slate-900">{session.session_title}</p>
									<Badge variant="outline">{session.session_status}</Badge>
								</div>
								<p class="mt-2 text-xs text-slate-500">{fmtDateTime(session.scheduled_start)}</p>
								<p class="mt-1 text-xs text-slate-500">Ruang {session.room_name || '—'} · Meja {session.seat_no ?? '—'}</p>
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
	{/if}

	{#if isParent && dashboardLoading}
		<div class="grid gap-4 lg:grid-cols-[0.9fr,1.1fr]">
			{#each Array.from({ length: 2 }) as _, index (`parent-dashboard-skeleton-${index}`)}
				<Card.Root class="border-slate-200">
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
	{:else if isParent && parentPortal}
		<div class="grid gap-4 lg:grid-cols-[0.9fr,1.1fr]">
			<Card.Root class="border-emerald-100">
				<Card.Header>
					<Card.Title class="text-base">Profil Wali</Card.Title>
					<Card.Description>{parentPortal.parent.nama}</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-2 text-sm text-slate-600">
					<p>Nomor HP: {parentPortal.parent.phone || 'belum diisi'}</p>
					<p>Alamat: {parentPortal.parent.address || 'belum diisi'}</p>
				</Card.Content>
			</Card.Root>

			<Card.Root class="border-slate-200">
				<Card.Header>
					<Card.Title class="text-base">Anak Terhubung</Card.Title>
					<Card.Description>Data siswa yang sudah ditautkan ke akun orang tua ini.</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-3">
					{#if parentPortal.children.length > 0}
						{#each parentPortal.children as child (child.id)}
							<div class="rounded-xl border border-slate-200 bg-white p-3">
								<div class="flex items-start justify-between gap-3">
									<div>
										<p class="text-sm font-medium text-slate-900">{child.nama}</p>
										<p class="mt-1 text-xs text-slate-500">NIS {child.nis}</p>
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
	{/if}

	{#if isGuru && dashboardLoading}
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
			{#each Array.from({ length: 4 }) as _, index (`guru-stat-skeleton-${index}`)}
				<Card.Root class="border-slate-200">
					<Card.Content class="space-y-2 pt-4">
						<Skeleton class="h-4 w-28" />
						<Skeleton class="h-8 w-16" />
					</Card.Content>
				</Card.Root>
			{/each}
		</div>
	{:else if isGuru && guruStats}
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
			<Card.Root class="border-green-100"><Card.Content class="pt-4"><p class="text-xs font-medium uppercase tracking-wider text-muted-foreground">Sesi CBT Berjalan</p><p class="mt-1 text-3xl font-bold text-green-800">{guruStats.active_sessions}</p></Card.Content></Card.Root>
			<Card.Root class="border-amber-100"><Card.Content class="pt-4"><p class="text-xs font-medium uppercase tracking-wider text-muted-foreground">Esai Belum Dikoreksi</p><p class="mt-1 text-3xl font-bold text-amber-700">{guruStats.ungraded_essays}</p></Card.Content></Card.Root>
			<Card.Root class="border-green-100"><Card.Content class="pt-4"><p class="text-xs font-medium uppercase tracking-wider text-muted-foreground">Siswa Terpantau</p><p class="mt-1 text-3xl font-bold text-green-800">{guruStats.my_students}</p></Card.Content></Card.Root>
			<Card.Root class="border-green-100"><Card.Content class="pt-4"><p class="text-xs font-medium uppercase tracking-wider text-muted-foreground">Mapel Diampu</p><p class="mt-1 text-3xl font-bold text-green-800">{guruStats.my_subjects}</p></Card.Content></Card.Root>
		</div>
	{/if}

	{#if (isAdmin || isStaff) && dashboardLoading}
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
			{#each Array.from({ length: 4 }) as _, index (`admin-stat-skeleton-${index}`)}
				<Card.Root class="border-slate-200">
					<Card.Content class="space-y-2 pt-4">
						<Skeleton class="h-4 w-28" />
						<Skeleton class="h-8 w-16" />
					</Card.Content>
				</Card.Root>
			{/each}
		</div>
	{:else if isAdmin || isStaff}
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
			<Card.Root class="border-green-100"><Card.Content class="pt-4"><p class="text-xs font-medium uppercase tracking-wider text-muted-foreground">Total Siswa</p><p class="mt-1 text-3xl font-bold text-green-800">{academicStats?.total_students ?? '—'}</p></Card.Content></Card.Root>
			<Card.Root class="border-green-100"><Card.Content class="pt-4"><p class="text-xs font-medium uppercase tracking-wider text-muted-foreground">Kelas Aktif</p><p class="mt-1 text-3xl font-bold text-green-800">{academicStats?.total_classes ?? '—'}</p></Card.Content></Card.Root>
			<Card.Root class="border-green-100"><Card.Content class="pt-4"><p class="text-xs font-medium uppercase tracking-wider text-muted-foreground">Mapel Aktif</p><p class="mt-1 text-3xl font-bold text-green-800">{academicStats?.total_subjects ?? '—'}</p></Card.Content></Card.Root>
			<Card.Root class="border-green-100"><Card.Content class="pt-4"><p class="text-xs font-medium uppercase tracking-wider text-muted-foreground">Tahun Ajaran</p><p class="mt-1 text-3xl font-bold text-green-800">{academicStats?.total_years ?? '—'}</p></Card.Content></Card.Root>
		</div>
	{/if}

	{#if isStaff}
		<Card.Root class="border-slate-200">
			<Card.Header>
				<Card.Title class="text-base">Akses Cepat Staf</Card.Title>
				<Card.Description>Menu yang paling sering dipakai untuk layanan data orang tua dan siswa.</Card.Description>
			</Card.Header>
			<Card.Content>
				<div class="flex flex-wrap gap-2">
					<Button variant="default" size="sm" href="/parents">Manajemen Orang Tua</Button>
					<Button variant="outline" size="sm" href="/students">Data Siswa</Button>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	{#if isAdmin}
		<Card.Root class="border-slate-200">
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
					<Button variant="default" size="sm" href="/pusaka">Kontrol & Monitor</Button>
					<Button variant="outline" size="sm" href="/pusaka/kehadiran">Data Kehadiran</Button>
					<Button variant="outline" size="sm" href="/pusaka/summary">Ringkasan</Button>
					<Button variant="outline" size="sm" href="/pusaka/antrian">Antrian Job</Button>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}
</div>
{/if}
