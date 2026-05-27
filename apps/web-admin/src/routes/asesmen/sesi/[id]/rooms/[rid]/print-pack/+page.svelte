<script lang="ts">
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import PrinterIcon from '@lucide/svelte/icons/printer';
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { Button } from '$lib/components/ui/button';
	import { fetchSchoolProfile, schoolAddressLine, type SchoolProfile } from '$lib/school-profile';
	import { clientApiPath, readClientApiData } from '$lib/client/api';

	type RoomDashboard = {
		id: string;
		session_id: string;
		room_name: string;
		capacity: number;
		room_token: string;
		status: string;
		is_locked: boolean;
		session_title: string;
		session_status: string;
		scheduled_start: string;
		scheduled_end: string;
		package_title: string;
		duration_minutes: number;
		school_room_code: string;
		school_room_name: string;
		school_room_building: string;
		school_room_location_note: string;
		participant_count: number;
		submitted_count: number;
		online_count: number;
		suspicious_count: number;
		missing_seat_count: number;
	};
	type RoomProctor = {
		id: string;
		exam_room_id: string;
		employee_id: string;
		nip: string;
		nama: string;
		role: string;
		assigned_at: string;
	};
	type ProctoringRow = {
		participant_id: string;
		student_id: string;
		nis: string;
		nama: string;
		token: string;
		room_id: string | null;
		room_name: string;
		seat_no?: number | null;
		submitted_at: string | null;
		last_heartbeat: string | null;
		app_switch_count: number;
		screenshot_attempt: number;
		suspicious_flag: boolean;
		answered_count: number;
		score: string | null;
	};
	type PrintPackPayload = {
		room?: RoomDashboard | null;
		proctors?: RoomProctor[];
		participants?: ProctoringRow[];
	};
	type PrintPackDetail = {
		room: RoomDashboard;
		proctors: RoomProctor[];
		participants: ProctoringRow[];
	};
	type PrintPackData = {
		schoolProfile: SchoolProfile;
		detail: PrintPackDetail;
	};

	const sessionId = page.params.id ?? '';
	const roomId = page.params.rid ?? '';

	const checklistItems = [
		'Ruang, kursi, dan nomor meja sesuai daftar peserta',
		'Jaringan internet dan listrik sudah dicek sebelum sesi dimulai',
		'Kode ruang diumumkan hanya setelah peserta siap di ruang dan dipakai sebagai pemeriksaan masuk APK/portal',
		'Perangkat cadangan / prosedur masuk ulang sudah diketahui pengawas',
		'Pengawas menyampaikan tata tertib, waktu, dan prosedur kirim ujian',
		'Semua peserta akhir sesi sudah kirim ujian atau diberi catatan kejadian'
	];

	let printPackPromise = $state<Promise<PrintPackData> | null>(null);

	async function fetchPrintPack(): Promise<PrintPackDetail> {
		const response = await fetch(clientApiPath`/api/asesmen/sessions/${sessionId}/rooms/${roomId}/print-pack`);
		const payload = await readClientApiData<PrintPackPayload>(response, 'Gagal memuat paket pengawas ruang');
		if (!payload.room) throw new Error('Data ruang ujian tidak ditemukan');
		return {
			room: payload.room,
			proctors: payload.proctors ?? [],
			participants: sortParticipants(payload.participants ?? []),
		};
	}

	async function fetchPrintPackData(): Promise<PrintPackData> {
		const [schoolProfile, detail] = await Promise.all([fetchSchoolProfile(), fetchPrintPack()]);
		return { schoolProfile, detail };
	}

	function loadPrintPack() {
		printPackPromise = fetchPrintPackData();
	}

	function retryPrintPack(reset?: () => void) {
		reset?.();
		loadPrintPack();
	}

	function printPackErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat paket pengawas ruang';
	}

	function handlePrintPackRenderError(error: unknown) {
		console.error('Paket cetak pengawas ruang belum dapat ditampilkan', error);
	}

	function sortParticipants(rows: ProctoringRow[]) {
		return [...rows].sort((a, b) => {
			const seatA = a.seat_no ?? Number.MAX_SAFE_INTEGER;
			const seatB = b.seat_no ?? Number.MAX_SAFE_INTEGER;
			if (seatA !== seatB) return seatA - seatB;
			return a.nama.localeCompare(b.nama, 'id-ID');
		});
	}

	function fmtDate(value: string | null | undefined) {
		if (!value) return '—';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return value;
		return date.toLocaleString('id-ID', {
			timeZone: 'Asia/Makassar',
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
		}) + ' WITA';
	}

	function submittedLabel(value: string | null) {
		return value ? 'Sudah kirim' : 'Belum kirim';
	}

	function proctorRoleLabel(role: string) {
		const labels: Record<string, string> = {
			primary: 'Utama',
			assistant: 'Pendamping',
			backup: 'Cadangan',
		};
		return labels[role] ?? role;
	}

	function roomLocation(room: RoomDashboard) {
		const parts = [room.school_room_code, room.school_room_name, room.school_room_building, room.school_room_location_note]
			.map((item) => item?.trim())
			.filter(Boolean);
		return parts.length > 0 ? parts.join(' · ') : 'Ruang manual sesi';
	}

	onMount(() => {
		loadPrintPack();
	});
</script>

<svelte:head>
	<title>Paket Pengawas Ruang Ujian Digital</title>
</svelte:head>

<AsyncContent promise={printPackPromise} onerror={handlePrintPackRenderError}>
	{#snippet pending()}
		<div class="mx-auto max-w-6xl space-y-5 p-6">
			<div class="space-y-2">
				<Skeleton class="h-8 w-72" />
				<Skeleton class="h-4 w-96" />
			</div>
			<Skeleton class="h-32 rounded-lg" />
			<div class="grid gap-4 md:grid-cols-3">
				{#each Array.from({ length: 3 }) as _, index (`print-pack-stat-${index}`)}
					<Skeleton class="h-28 rounded-lg" />
				{/each}
			</div>
			<Skeleton class="h-72 rounded-lg" />
			<Skeleton class="h-64 rounded-lg" />
		</div>
	{/snippet}

	{#snippet failed(error, reset)}
		<div class="mx-auto max-w-6xl p-6">
			<RecoveryPanel
				title="Paket Pengawas Belum Tersaji"
				message={printPackErrorMessage(error)}
				onRetry={() => retryPrintPack(reset)}
			/>
		</div>
	{/snippet}

	{#snippet children(value)}
		{@const printData = value as PrintPackData}
		{@const detail = printData.detail}
		{@const schoolProfile = printData.schoolProfile}
		{@const room = detail.room}
		{@const participants = detail.participants}
		{@const proctors = detail.proctors}
		<div class="mx-auto max-w-6xl space-y-5 p-6 text-foreground print:max-w-none print:space-y-4 print:p-0">
			<div class="flex items-center justify-between gap-3 print:hidden">
				<div>
					<a href={resolve(`/asesmen/sesi/${sessionId}/rooms/${roomId}/proctoring`)} class="inline-flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.18em] text-primary">
						<ArrowLeftIcon class="size-3.5" />
						Kembali ke panel ruang
					</a>
					<h1 class="mt-2 text-2xl font-semibold tracking-tight">Paket Pengawas Ruang Ujian Digital</h1>
					<p class="text-sm text-muted-foreground">Daftar hadir, kode ujian, denah meja, kontak operator, dan daftar pemeriksaan kesiapan.</p>
				</div>
				<Button onclick={() => window.print()}>
					<PrinterIcon class="mr-2 size-4" />
					Cetak
				</Button>
			</div>

			<section class="border border-primary/20 bg-card p-5 text-center shadow-sm print:border-border print:shadow-none">
				<p class="text-xs font-semibold uppercase text-foreground">{schoolProfile.ministry_line}</p>
				<p class="text-xs font-semibold uppercase text-foreground">{schoolProfile.office_line}</p>
				<h2 class="mt-1 text-lg font-bold uppercase text-foreground">{schoolProfile.name}</h2>
				<p class="mt-1 text-[11px] leading-5 text-muted-foreground">{schoolAddressLine(schoolProfile) || 'Alamat madrasah belum diisi'}</p>
				{#if schoolProfile.nsm || schoolProfile.npsn}
					<p class="text-[11px] text-muted-foreground">
						{#if schoolProfile.nsm}NSM {schoolProfile.nsm}{/if}
						{#if schoolProfile.nsm && schoolProfile.npsn} · {/if}
						{#if schoolProfile.npsn}NPSN {schoolProfile.npsn}{/if}
					</p>
				{/if}
			</section>

			<section class="grid gap-4 border border-border bg-card p-5 print:grid-cols-[minmax(0,1fr)_220px]">
				<div>
					<p class="text-xs font-semibold uppercase tracking-[0.2em] text-primary">Paket Pengawas Ruang Ujian Digital</p>
					<h2 class="mt-2 text-xl font-bold uppercase">{room.session_title}</h2>
					<div class="mt-4 grid gap-2 text-sm md:grid-cols-2">
						<p><span class="font-semibold">Paket:</span> {room.package_title}</p>
						<p><span class="font-semibold">Ruang:</span> {room.room_name}</p>
						<p><span class="font-semibold">Mulai:</span> {fmtDate(room.scheduled_start)}</p>
						<p><span class="font-semibold">Selesai:</span> {fmtDate(room.scheduled_end)}</p>
						<p><span class="font-semibold">Durasi:</span> {room.duration_minutes} menit</p>
						<p><span class="font-semibold">Lokasi:</span> {roomLocation(room)}</p>
					</div>
				</div>
				<div class="border border-primary/20 bg-primary/10 p-4 text-center">
					<p class="text-xs font-semibold uppercase tracking-[0.18em] text-primary">Kode Ruang</p>
					<p class="mt-3 font-mono text-3xl font-bold tracking-[0.18em] text-primary">{room.room_token || '—'}</p>
					<p class="mt-3 text-[11px] leading-4 text-primary">Dipakai sebagai kunci ruang untuk membuka akses kode ujian siswa di portal dan/atau pemeriksaan masuk APK. Jangan diberikan sebelum peserta berada di ruang ujian.</p>
				</div>
			</section>

			<section class="grid gap-3 md:grid-cols-4 print:grid-cols-4">
				<div class="border border-border bg-card p-3">
					<p class="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">Peserta</p>
					<p class="mt-1 text-2xl font-bold">{participants.length}</p>
				</div>
				<div class="border border-border bg-card p-3">
					<p class="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">Kapasitas</p>
					<p class="mt-1 text-2xl font-bold">{room.capacity}</p>
				</div>
				<div class="border border-border bg-card p-3">
					<p class="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">Kirim</p>
					<p class="mt-1 text-2xl font-bold">{room.submitted_count}</p>
				</div>
				<div class="border border-border bg-card p-3">
					<p class="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">Atensi</p>
					<p class="mt-1 text-2xl font-bold">{room.suspicious_count}</p>
				</div>
			</section>

			<section class="grid gap-4 md:grid-cols-[minmax(0,1fr)_320px] print:grid-cols-[minmax(0,1fr)_300px]">
				<div class="border border-border bg-card p-4">
					<h3 class="text-sm font-bold uppercase tracking-[0.16em]">Pengawas Ruang</h3>
					<div class="mt-3 overflow-x-auto">
						<table class="min-w-full text-left text-xs">
							<thead class="bg-muted text-muted-foreground">
								<tr>
									<th class="border px-2 py-1.5">Nama</th>
									<th class="border px-2 py-1.5">NIP</th>
									<th class="border px-2 py-1.5">Peran</th>
									<th class="border px-2 py-1.5">Paraf</th>
								</tr>
							</thead>
							<tbody>
								{#each proctors as proctor (proctor.id)}
									<tr>
										<td class="border px-2 py-1.5 font-medium">{proctor.nama}</td>
										<td class="border px-2 py-1.5 font-mono">{proctor.nip || '—'}</td>
										<td class="border px-2 py-1.5">{proctorRoleLabel(proctor.role)}</td>
										<td class="border px-2 py-1.5">&nbsp;</td>
									</tr>
								{:else}
									<tr>
										<td colspan="4" class="border px-2 py-4 text-center text-warning">Pengawas ruang belum ditugaskan</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</div>

				<div class="border border-border bg-card p-4">
					<h3 class="text-sm font-bold uppercase tracking-[0.16em]">Kontak Operator</h3>
					<div class="mt-3 space-y-3 text-xs">
						<p class="border-b border-border pb-2">Operator ujian digital: ........................................</p>
						<p class="border-b border-border pb-2">Nomor HP: ............................................</p>
						<p class="border-b border-border pb-2">Waktu eskalasi: ......................................</p>
						<p class="min-h-16 border border-border p-2 text-muted-foreground">Catatan gangguan / arahan operator</p>
					</div>
				</div>
			</section>

			<section class="border border-border bg-card p-4">
					<h3 class="text-sm font-bold uppercase tracking-[0.16em]">Daftar Hadir dan Kode Ujian Siswa</h3>
				<div class="mt-3 overflow-x-auto">
					<table class="min-w-full text-left text-[11px]">
						<thead class="bg-muted text-muted-foreground">
							<tr>
								<th class="border px-2 py-1.5 text-center">No</th>
								<th class="border px-2 py-1.5 text-center">Meja</th>
								<th class="border px-2 py-1.5">NIS</th>
								<th class="border px-2 py-1.5">Nama Peserta</th>
								<th class="border px-2 py-1.5">Kode Ujian</th>
								<th class="border px-2 py-1.5">Status</th>
								<th class="border px-2 py-1.5">Paraf Masuk</th>
								<th class="border px-2 py-1.5">Paraf Keluar / Catatan</th>
							</tr>
						</thead>
						<tbody>
							{#each participants as participant, index (participant.participant_id)}
								<tr class={participant.suspicious_flag ? 'bg-warning/10' : ''}>
									<td class="border px-2 py-1.5 text-center">{index + 1}</td>
									<td class="border px-2 py-1.5 text-center font-mono">{participant.seat_no ?? '—'}</td>
									<td class="border px-2 py-1.5 font-mono">{participant.nis}</td>
									<td class="border px-2 py-1.5 font-medium">{participant.nama}</td>
									<td class="border px-2 py-1.5 font-mono">{participant.token || '—'}</td>
									<td class="border px-2 py-1.5">{submittedLabel(participant.submitted_at)}</td>
									<td class="border px-2 py-1.5">&nbsp;</td>
									<td class="border px-2 py-1.5">&nbsp;</td>
								</tr>
							{:else}
								<tr>
									<td colspan="8" class="border px-2 py-6 text-center text-muted-foreground">Belum ada peserta di ruang ini</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</section>

			<section class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_360px] print:grid-cols-[minmax(0,1fr)_330px]">
				<div class="border border-border bg-card p-4">
					<h3 class="text-sm font-bold uppercase tracking-[0.16em]">Denah Meja Sederhana</h3>
					<div class="mt-3 grid grid-cols-2 gap-2 sm:grid-cols-3 lg:grid-cols-4 print:grid-cols-4">
						{#each participants as participant (participant.participant_id)}
							<div class="min-h-20 border border-border bg-muted/50 p-2 text-[11px] break-inside-avoid">
								<p class="font-bold">Meja {participant.seat_no ?? '—'}</p>
								<p class="mt-1 line-clamp-2 font-medium">{participant.nama}</p>
								<p class="font-mono text-muted-foreground">{participant.nis}</p>
							</div>
						{:else}
							<p class="col-span-full border border-dashed border-border p-4 text-center text-xs text-muted-foreground">Denah belum tersedia karena peserta belum ditempatkan.</p>
						{/each}
					</div>
				</div>

				<div class="border border-border bg-card p-4">
					<h3 class="text-sm font-bold uppercase tracking-[0.16em]">Daftar Pemeriksaan Kesiapan</h3>
					<div class="mt-3 space-y-2 text-xs">
						{#each checklistItems as item (item)}
							<div class="flex items-start gap-2">
								<span class="mt-0.5 inline-block size-4 border border-border"></span>
								<span>{item}</span>
							</div>
						{/each}
					</div>
				</div>
			</section>

			<section class="grid gap-8 border border-border bg-card p-5 text-sm md:grid-cols-2 print:grid-cols-2">
				<div class="text-center">
					<p>Mengetahui,</p>
					<p>Kepala Madrasah</p>
					<div class="mt-16 border-t border-foreground pt-2">
						<p class="font-semibold">{schoolProfile.head_name || '........................................'}</p>
						<p>NIP. {schoolProfile.head_nip || '................................'}</p>
					</div>
				</div>
				<div class="text-center">
					<p>Pengawas Ruang,</p>
					<p>{room.room_name}</p>
					<div class="mt-16 border-t border-foreground pt-2">
						<p class="font-semibold">........................................</p>
						<p>NIP. ................................</p>
					</div>
				</div>
			</section>
		</div>
	{/snippet}
</AsyncContent>
