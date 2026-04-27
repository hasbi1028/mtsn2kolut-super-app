<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';

	type ExamSession = {
		id: string; package_id: string; package_title: string;
		class_id: string; class_name: string; class_code: string;
		title: string; scheduled_start: string; scheduled_end: string;
		status: string; participant_count: number; created_at: string;
	};
	type CbtPackage = { id: string; title: string; subject_code: string; subject_name: string; };
	type SchoolClass = { id: string; name: string; code: string; level: string; };

	let sessions = $state<ExamSession[]>([]);
	let packages = $state<CbtPackage[]>([]);
	let classes = $state<SchoolClass[]>([]);
	let loading = $state(true);
	let error = $state('');
	let toast = $state({ msg: '', ok: true });
	let showForm = $state(false);

	let fPackageId = $state('');
	let fClassId = $state('');
	let fTitle = $state('');
	let fStart = $state('');
	let fEnd = $state('');
	let fBusy = $state(false);

	// Enroll modal
	let enrollSession = $state<ExamSession | null>(null);
	let enrollClassId = $state('');
	let enrollBusy = $state(false);

	const statusLabel: Record<string, string> = {
		draft: 'Draft', scheduled: 'Terjadwal', active: 'Berlangsung',
		finished: 'Selesai', cancelled: 'Dibatalkan',
	};

	function statusClass(s: string) {
		if (s === 'active') return 'bg-emerald-100 text-emerald-700 border-emerald-200';
		if (s === 'finished') return 'bg-slate-100 text-slate-500 border-slate-200';
		if (s === 'cancelled') return 'bg-red-100 text-red-700 border-red-200';
		if (s === 'scheduled') return 'bg-blue-100 text-blue-700 border-blue-200';
		return 'bg-amber-100 text-amber-700 border-amber-200';
	}

	function fmtDt(iso: string) {
		if (!iso) return '—';
		return new Date(iso).toLocaleString('id-ID', {
			timeZone: 'Asia/Makassar', year: 'numeric', month: 'short',
			day: 'numeric', hour: '2-digit', minute: '2-digit',
		}) + ' WITA';
	}

	function toRFC3339(localDt: string): string {
		if (!localDt) return '';
		return new Date(localDt).toISOString();
	}

	async function load() {
		try {
			const [sRes, pRes, aRes] = await Promise.all([
				fetch('/api/cbt/sessions'),
				fetch('/api/cbt/packages'),
				fetch('/api/academic'),
			]);
			const sJson = await sRes.json();
			const pJson = await pRes.json();
			const aJson = await aRes.json();
			if (sJson.error) { error = sJson.error; return; }
			sessions = sJson.data ?? sJson ?? [];
			const pd = pJson.data ?? pJson;
			packages = pd.packages ?? [];
			classes = (aJson.data ?? aJson)?.classes ?? [];
		} catch {
			error = 'Gagal memuat data sesi';
		} finally {
			loading = false;
		}
	}

	function showToast(msg: string, ok = true) {
		toast = { msg, ok };
		setTimeout(() => (toast = { msg: '', ok: true }), 3500);
	}

	async function createSession() {
		if (!fPackageId || !fClassId || !fTitle || !fStart || !fEnd) return;
		fBusy = true;
		try {
			const res = await fetch('/api/cbt/sessions', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					package_id: fPackageId, class_id: fClassId, title: fTitle,
					scheduled_start: toRFC3339(fStart), scheduled_end: toRFC3339(fEnd),
					status: 'draft',
				}),
			});
			if (!res.ok) { const j = await res.json(); showToast(j.error ?? 'Gagal', false); return; }
			fPackageId = ''; fClassId = ''; fTitle = ''; fStart = ''; fEnd = '';
			showForm = false;
			showToast('Sesi ujian berhasil dibuat');
			await load();
		} finally { fBusy = false; }
	}

	async function deleteSession(id: string, title: string) {
		if (!confirm(`Hapus sesi "${title}"? Hanya sesi berstatus Draft yang dapat dihapus.`)) return;
		const res = await fetch(`/api/cbt/sessions?id=${id}`, { method: 'DELETE' });
		if (!res.ok && res.status !== 204) {
			showToast('Gagal menghapus — hanya sesi Draft yang dapat dihapus', false);
		} else {
			showToast('Sesi dihapus');
			await load();
		}
	}

	async function updateStatus(id: string, status: string) {
		const res = await fetch(`/api/cbt/sessions/${id}/status`, {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ status }),
		});
		if (!res.ok) { showToast('Gagal mengubah status', false); return; }
		showToast('Status diperbarui');
		await load();
	}

	async function enrollClass() {
		if (!enrollSession || !enrollClassId) return;
		enrollBusy = true;
		try {
			const res = await fetch(`/api/cbt/sessions/${enrollSession.id}/enroll`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ class_id: enrollClassId }),
			});
			if (!res.ok) { showToast('Gagal mendaftarkan siswa', false); return; }
			showToast(`Siswa kelas berhasil didaftarkan ke sesi "${enrollSession.title}"`);
			enrollSession = null;
			enrollClassId = '';
			await load();
		} finally { enrollBusy = false; }
	}

	onMount(load);
</script>

<svelte:head><title>Sesi Ujian CBT — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6 p-6 max-w-5xl mx-auto">
	<div class="flex items-start justify-between">
		<div>
			<h1 class="text-2xl font-semibold text-slate-800">Sesi Ujian CBT</h1>
			<p class="text-sm text-slate-500 mt-1">Jadwalkan dan kelola pelaksanaan ujian per kelas</p>
		</div>
		<Button onclick={() => (showForm = !showForm)}>
			{showForm ? 'Batal' : '+ Buat Sesi'}
		</Button>
	</div>

	{#if toast.msg}
		<div class="rounded-md px-4 py-3 text-sm {toast.ok ? 'bg-emerald-50 border border-emerald-200 text-emerald-800' : 'bg-red-50 border border-red-200 text-red-800'}">
			{toast.msg}
		</div>
	{/if}
	{#if error}
		<div class="rounded-md bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-800">{error}</div>
	{/if}

	{#if showForm}
		<Card.Root>
			<Card.Header class="pb-2">
				<Card.Title class="text-base">Buat Sesi Ujian Baru</Card.Title>
			</Card.Header>
			<Card.Content class="space-y-4">
				<div class="grid gap-3 sm:grid-cols-2">
					<div>
						<label class="text-xs text-slate-500 mb-1 block">Paket Soal <span class="text-red-500">*</span></label>
						<select class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fPackageId}>
							<option value="">-- Pilih Paket --</option>
							{#each packages as p}
								<option value={p.id}>{p.title} ({p.subject_code})</option>
							{/each}
						</select>
					</div>
					<div>
						<label class="text-xs text-slate-500 mb-1 block">Kelas <span class="text-red-500">*</span></label>
						<select class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fClassId}>
							<option value="">-- Pilih Kelas --</option>
							{#each classes as c}
								<option value={c.id}>{c.code} — {c.name}</option>
							{/each}
						</select>
					</div>
					<div class="sm:col-span-2">
						<label class="text-xs text-slate-500 mb-1 block">Nama Sesi <span class="text-red-500">*</span></label>
						<Input placeholder="mis: UTS Matematika VII A - Semester 1 2025" bind:value={fTitle} />
					</div>
					<div>
						<label class="text-xs text-slate-500 mb-1 block">Mulai <span class="text-red-500">*</span></label>
						<Input type="datetime-local" bind:value={fStart} />
					</div>
					<div>
						<label class="text-xs text-slate-500 mb-1 block">Selesai <span class="text-red-500">*</span></label>
						<Input type="datetime-local" bind:value={fEnd} />
					</div>
				</div>
				<div class="flex gap-2">
					<Button disabled={fBusy || !fPackageId || !fClassId || !fTitle || !fStart || !fEnd} onclick={createSession}>
						{fBusy ? 'Menyimpan...' : 'Buat Sesi'}
					</Button>
					<Button variant="outline" onclick={() => (showForm = false)}>Batal</Button>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	<!-- Enroll modal -->
	{#if enrollSession}
		<Card.Root class="border-blue-200 bg-blue-50">
			<Card.Header class="pb-2">
				<Card.Title class="text-base text-blue-800">Daftarkan Siswa ke Sesi</Card.Title>
				<p class="text-sm text-blue-600 mt-0.5">{enrollSession.title}</p>
			</Card.Header>
			<Card.Content class="space-y-3">
				<p class="text-sm text-slate-600">Semua siswa aktif dari kelas yang dipilih akan didaftarkan ke sesi ini.</p>
				<div>
					<label class="text-xs text-slate-500 mb-1 block">Pilih Kelas</label>
					<select class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={enrollClassId}>
						<option value="">-- Pilih Kelas --</option>
						{#each classes as c}
							<option value={c.id}>{c.code} — {c.name}</option>
						{/each}
					</select>
				</div>
				<div class="flex gap-2">
					<Button disabled={enrollBusy || !enrollClassId} onclick={enrollClass}>
						{enrollBusy ? 'Mendaftarkan...' : 'Daftarkan Siswa'}
					</Button>
					<Button variant="outline" onclick={() => { enrollSession = null; enrollClassId = ''; }}>Batal</Button>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	{#if loading}
		<p class="text-sm text-slate-500">Memuat data...</p>
	{:else}
		<Card.Root>
			<Card.Header class="pb-2">
				<Card.Title class="text-base">Daftar Sesi ({sessions.length})</Card.Title>
			</Card.Header>
			<Card.Content class="p-0">
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>Nama Sesi</Table.Head>
							<Table.Head>Paket</Table.Head>
							<Table.Head>Kelas</Table.Head>
							<Table.Head>Jadwal Mulai</Table.Head>
							<Table.Head>Peserta</Table.Head>
							<Table.Head>Status</Table.Head>
							<Table.Head></Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each sessions as s}
							<Table.Row>
								<Table.Cell class="font-medium max-w-48">
									<p class="truncate">{s.title}</p>
								</Table.Cell>
								<Table.Cell class="text-slate-500 text-sm truncate max-w-32">{s.package_title}</Table.Cell>
								<Table.Cell>
									<Badge variant="outline" class="text-xs">{s.class_code}</Badge>
								</Table.Cell>
								<Table.Cell class="text-slate-500 text-xs whitespace-nowrap">{fmtDt(s.scheduled_start)}</Table.Cell>
								<Table.Cell>
									<span class="font-mono text-sm">{s.participant_count}</span>
								</Table.Cell>
								<Table.Cell>
									<Badge class={statusClass(s.status)}>{statusLabel[s.status] ?? s.status}</Badge>
								</Table.Cell>
								<Table.Cell>
									<div class="flex gap-1 flex-wrap">
										{#if s.status === 'draft'}
											<Button size="xs" variant="outline" onclick={() => { enrollSession = s; enrollClassId = s.class_id; }}>
												Daftarkan Siswa
											</Button>
											<Button size="xs" onclick={() => updateStatus(s.id, 'scheduled')}>
												Jadwalkan
											</Button>
											<Button size="xs" variant="destructive" onclick={() => deleteSession(s.id, s.title)}>
												Hapus
											</Button>
										{:else if s.status === 'scheduled'}
											<Button size="xs" onclick={() => updateStatus(s.id, 'active')}>Mulai</Button>
											<Button size="xs" variant="outline" onclick={() => updateStatus(s.id, 'cancelled')}>Batalkan</Button>
										{:else if s.status === 'active'}
											<Button size="xs" onclick={() => updateStatus(s.id, 'finished')}>Selesaikan</Button>
										{/if}
										{#if s.status === 'finished' || s.status === 'active'}
											<a href="/cbt/sessions/{s.id}" class="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium border border-input bg-background hover:bg-muted text-slate-700 transition-colors">
												Lihat Hasil
											</a>
										{/if}
									</div>
								</Table.Cell>
							</Table.Row>
						{:else}
							<Table.Row>
								<Table.Cell colspan={7} class="text-center text-slate-400 py-8">Belum ada sesi ujian</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			</Card.Content>
		</Card.Root>
	{/if}
</div>
