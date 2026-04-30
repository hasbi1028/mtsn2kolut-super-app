<script lang="ts">
	import { goto } from '$app/navigation';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import { toast } from '$lib/components/ui/sonner';

	type Assignment = {
		id: string;
		class_name: string;
		class_code: string;
		subject_name: string;
		subject_code: string;
		teacher_name: string;
	};

	type Session = {
		id: string;
		assignment_id: string;
		tanggal: string;
		pertemuan_ke: number;
		materi: string;
		kegiatan: string;
		catatan: string;
		guru_hadir: boolean;
		class_name: string;
		subject_name: string;
		teacher_name: string;
	};

	type AttendanceSummary = {
		student_id: string;
		nis: string;
		nisn: string;
		nama: string;
		total_pertemuan: number;
		hadir: number;
		sakit: number;
		izin: number;
		alpha: number;
	};

	type Overview = {
		assignments: Assignment[];
		sessions: Session[];
		summary: AttendanceSummary[];
	};

	let loading = $state(true);
	let error = $state('');
	let assignments = $state<Assignment[]>([]);
	let sessions = $state<Session[]>([]);
	let summary = $state<AttendanceSummary[]>([]);
	let assignmentId = $state('');
	let activeTab = $state<'sessions' | 'rekap'>('sessions');

	let createOpen = $state(false);
	let createBusy = $state(false);
	let deleteBusy = $state<Record<string, boolean>>({});

	let newTanggal = $state('');
	let newMateri = $state('');
	let newKegiatan = $state('');
	let newCatatan = $state('');
	let newGuruHadir = $state(true);

	async function loadOverview() {
		loading = true;
		error = '';
		try {
			const params = assignmentId ? `?assignment_id=${assignmentId}` : '';
			const res = await fetch(`/api/journal${params}`);
			if (!res.ok) throw new Error((await res.json()).error ?? `HTTP ${res.status}`);
			const data = (await res.json()) as Overview;
			assignments = data.assignments ?? [];
			sessions = data.sessions ?? [];
			summary = data.summary ?? [];
		} catch (e) {
			error = e instanceof Error ? e.message : 'Gagal memuat data jurnal';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		// Track assignmentId to re-fetch on change
		const _ = assignmentId;
		loadOverview();
	});

	async function createSession() {
		if (!assignmentId) { toast.error('Pilih kelas-mapel terlebih dahulu'); return; }
		if (!newTanggal) { toast.error('Tanggal wajib diisi'); return; }
		createBusy = true;
		try {
			const res = await fetch('/api/journal/sessions', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					assignment_id: assignmentId,
					tanggal: newTanggal,
					materi: newMateri,
					kegiatan: newKegiatan,
					catatan: newCatatan,
					guru_hadir: newGuruHadir
				})
			});
			const json = await res.json();
			if (!res.ok) {
				if (res.status === 409) toast.error('Pertemuan pada tanggal ini sudah ada');
				else toast.error(json.error ?? 'Gagal membuat pertemuan');
				return;
			}
			createOpen = false;
			newTanggal = '';
			newMateri = '';
			newKegiatan = '';
			newCatatan = '';
			newGuruHadir = true;
			toast.success('Sesi berhasil dibuat');
			goto(`/journal/${json.session.id}`);
		} catch {
			toast.error('Terjadi kesalahan jaringan');
		} finally {
			createBusy = false;
		}
	}

	async function deleteSession(id: string) {
		if (!confirm('Hapus pertemuan ini? Semua data kehadiran akan ikut terhapus.')) return;
		deleteBusy = { ...deleteBusy, [id]: true };
		try {
			const res = await fetch(`/api/journal/sessions/${id}`, { method: 'DELETE' });
			if (!res.ok && res.status !== 204) {
				const json = await res.json();
				toast.error(json.error ?? 'Gagal menghapus sesi');
				return;
			}
			toast.success('Sesi dihapus');
			await loadOverview();
		} catch {
			toast.error('Terjadi kesalahan jaringan');
		} finally {
			deleteBusy = { ...deleteBusy, [id]: false };
		}
	}

	function formatTanggal(raw: string) {
		if (!raw) return '-';
		const d = new Date(raw);
		return d.toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' });
	}
</script>

<div class="container mx-auto max-w-6xl space-y-6 p-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Jurnal Kelas</h1>
			<p class="text-sm text-gray-500">Catatan pertemuan dan kehadiran siswa per mata pelajaran</p>
		</div>
		<Button onclick={() => (createOpen = true)} disabled={!assignmentId}>
			+ Tambah Pertemuan
		</Button>
	</div>

	<!-- Assignment selector -->
	<Card.Root>
		<Card.Content class="pt-4">
			<div class="flex items-center gap-4">
				<label for="assignment-select" class="w-32 shrink-0 text-sm font-medium text-gray-700">Kelas – Mapel</label>
				{#if loading && assignments.length === 0}
					<Skeleton class="h-9 w-72" />
				{:else}
					<select
						id="assignment-select"
						bind:value={assignmentId}
						class="w-72 rounded-md border border-gray-300 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-green-600"
					>
						<option value="">-- Pilih kelas & mata pelajaran --</option>
						{#each assignments as a}
							<option value={a.id}>{a.class_name} – {a.subject_name} ({a.teacher_name})</option>
						{/each}
					</select>
				{/if}
			</div>
		</Card.Content>
	</Card.Root>

	{#if error}
		<div class="rounded-md border border-red-200 bg-red-50 p-4 text-sm text-red-700">{error}</div>
	{/if}

	{#if assignmentId}
		<!-- Tabs -->
		<div class="flex gap-1 border-b border-gray-200">
			<button
				class="px-4 py-2 text-sm font-medium transition-colors {activeTab === 'sessions'
					? 'border-b-2 border-green-600 text-green-700'
					: 'text-gray-500 hover:text-gray-700'}"
				onclick={() => (activeTab = 'sessions')}
			>
				Daftar Pertemuan
			</button>
			<button
				class="px-4 py-2 text-sm font-medium transition-colors {activeTab === 'rekap'
					? 'border-b-2 border-green-600 text-green-700'
					: 'text-gray-500 hover:text-gray-700'}"
				onclick={() => (activeTab = 'rekap')}
			>
				Rekap Kehadiran
			</button>
		</div>

		<!-- Daftar Pertemuan Tab -->
		{#if activeTab === 'sessions'}
			<Card.Root>
				<Card.Content class="p-0">
					{#if loading}
						<div class="space-y-2 p-4">
							{#each [1, 2, 3] as _}
								<Skeleton class="h-12 w-full" />
							{/each}
						</div>
					{:else if sessions.length === 0}
						<div class="p-8 text-center text-sm text-gray-500">
							Belum ada pertemuan. Klik "Tambah Pertemuan" untuk mulai.
						</div>
					{:else}
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head class="w-12">No</Table.Head>
									<Table.Head class="w-20">Ke-</Table.Head>
									<Table.Head>Tanggal</Table.Head>
									<Table.Head>Materi</Table.Head>
									<Table.Head class="w-28">Guru</Table.Head>
									<Table.Head class="w-32">Aksi</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each sessions as s, i}
									<Table.Row>
										<Table.Cell class="text-gray-500">{i + 1}</Table.Cell>
										<Table.Cell>
											<Badge variant="outline">P-{s.pertemuan_ke}</Badge>
										</Table.Cell>
										<Table.Cell class="text-sm">{formatTanggal(s.tanggal)}</Table.Cell>
										<Table.Cell class="max-w-xs truncate text-sm text-gray-700">
											{#if s.materi}
												{s.materi}
											{:else}
												<span class="italic text-gray-400">–</span>
											{/if}
										</Table.Cell>
										<Table.Cell>
											{#if s.guru_hadir}
												<Badge class="bg-green-100 text-green-800 hover:bg-green-100">Hadir</Badge>
											{:else}
												<Badge class="bg-red-100 text-red-800 hover:bg-red-100">Tidak Hadir</Badge>
											{/if}
										</Table.Cell>
										<Table.Cell>
											<div class="flex gap-1">
												<Button variant="outline" size="sm" onclick={() => goto(`/journal/${s.id}`)}>
													Lihat
												</Button>
												<LoadingButton
													variant="destructive"
													size="sm"
													loading={deleteBusy[s.id] ?? false}
													onclick={() => deleteSession(s.id)}
												>
													Hapus
												</LoadingButton>
											</div>
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					{/if}
				</Card.Content>
			</Card.Root>
		{/if}

		<!-- Rekap Kehadiran Tab -->
		{#if activeTab === 'rekap'}
			<Card.Root>
				<Card.Content class="p-0">
					{#if loading}
						<div class="space-y-2 p-4">
							{#each [1, 2, 3] as _}
								<Skeleton class="h-12 w-full" />
							{/each}
						</div>
					{:else if summary.length === 0}
						<div class="p-8 text-center text-sm text-gray-500">
							Belum ada data rekap kehadiran.
						</div>
					{:else}
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head class="w-12">No</Table.Head>
									<Table.Head>Nama Siswa</Table.Head>
									<Table.Head class="w-24">NIS</Table.Head>
									<Table.Head class="w-16 text-center">H</Table.Head>
									<Table.Head class="w-16 text-center">S</Table.Head>
									<Table.Head class="w-16 text-center">I</Table.Head>
									<Table.Head class="w-16 text-center">A</Table.Head>
									<Table.Head class="w-20 text-center">Total</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each summary as st, i}
									<Table.Row>
										<Table.Cell class="text-gray-500">{i + 1}</Table.Cell>
										<Table.Cell class="font-medium">{st.nama}</Table.Cell>
										<Table.Cell class="text-sm text-gray-500">{st.nis}</Table.Cell>
										<Table.Cell class="text-center">
											<Badge class="bg-green-100 text-green-800 hover:bg-green-100">{st.hadir}</Badge>
										</Table.Cell>
										<Table.Cell class="text-center">
											<Badge class="bg-yellow-100 text-yellow-800 hover:bg-yellow-100">{st.sakit}</Badge>
										</Table.Cell>
										<Table.Cell class="text-center">
											<Badge class="bg-sky-100 text-sky-800 hover:bg-sky-100">{st.izin}</Badge>
										</Table.Cell>
										<Table.Cell class="text-center">
											<Badge class="bg-red-100 text-red-800 hover:bg-red-100">{st.alpha}</Badge>
										</Table.Cell>
										<Table.Cell class="text-center text-sm text-gray-600">{st.total_pertemuan}</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					{/if}
				</Card.Content>
			</Card.Root>
		{/if}
	{/if}
</div>

<!-- Create Session Dialog -->
<Dialog.Root bind:open={createOpen}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>Tambah Pertemuan</Dialog.Title>
			<Dialog.Description>Isi detail pertemuan baru untuk kelas ini.</Dialog.Description>
		</Dialog.Header>
		<div class="space-y-4 py-2">
			<div class="space-y-1">
				<label for="new-tanggal" class="text-sm font-medium text-gray-700">
					Tanggal <span class="text-red-500">*</span>
				</label>
				<Input id="new-tanggal" type="date" bind:value={newTanggal} />
			</div>
			<div class="space-y-1">
				<label for="new-materi" class="text-sm font-medium text-gray-700">Materi</label>
				<Textarea id="new-materi" bind:value={newMateri} placeholder="Topik atau materi yang diajarkan" rows={2} />
			</div>
			<div class="space-y-1">
				<label for="new-kegiatan" class="text-sm font-medium text-gray-700">Kegiatan</label>
				<Textarea id="new-kegiatan" bind:value={newKegiatan} placeholder="Aktivitas pembelajaran" rows={2} />
			</div>
			<div class="space-y-1">
				<label for="new-catatan" class="text-sm font-medium text-gray-700">Catatan</label>
				<Textarea id="new-catatan" bind:value={newCatatan} placeholder="Catatan tambahan (opsional)" rows={2} />
			</div>
			<div class="flex items-center gap-3">
				<input
					id="new-guru-hadir"
					type="checkbox"
					bind:checked={newGuruHadir}
					class="h-4 w-4 rounded border-gray-300 text-green-600"
				/>
				<label for="new-guru-hadir" class="text-sm font-medium text-gray-700">Guru hadir mengajar</label>
			</div>
		</div>
		<Dialog.Footer>
			<Button variant="outline" onclick={() => (createOpen = false)}>Batal</Button>
			<LoadingButton loading={createBusy} onclick={createSession}>Simpan & Buka</LoadingButton>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
