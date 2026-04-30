<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from '$lib/components/ui/sonner';

	type ExamSession = {
		id: string; package_id: string; package_title: string;
		class_id: string; class_name: string; class_code: string;
		scope_type: string; scope_ref: string; mix_policy: string; assignment_mode: string;
		allow_cross_grade: boolean; is_special_event: boolean;
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
	let showForm = $state(false);

	let fPackageId = $state('');
	let fScopeType = $state('class');
	let fClassId = $state('');
	let fGradeLevel = $state('VII');
	let fMixPolicy = $state('same_grade');
	let fAssignmentMode = $state('random_balanced');
	let fAllowCrossGrade = $state(false);
	let fIsSpecialEvent = $state(false);
	let fTitle = $state('');
	let fStart = $state('');
	let fEnd = $state('');
	let fBusy = $state(false);

	// Enroll modal
	let enrollSession = $state<ExamSession | null>(null);
	let enrollScopeType = $state('class');
	let enrollClassId = $state('');
	let enrollGradeLevel = $state('VII');
	let enrollBusy = $state(false);

	const statusLabel: Record<string, string> = {
		draft: 'Draft', scheduled: 'Terjadwal', active: 'Berlangsung',
		finished: 'Selesai', cancelled: 'Dibatalkan',
	};

	function statusClass(s: string) {
		if (s === 'active') return 'bg-emerald-100 text-emerald-700 border-emerald-200';
		if (s === 'finished') return 'bg-slate-100 text-slate-500 border-slate-200';
		if (s === 'cancelled') return 'bg-red-100 text-red-700 border-red-200';
		if (s === 'scheduled') return 'bg-green-100 text-green-800 border-green-200';
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

	function scopeSummary(session: ExamSession) {
		if (session.scope_type === 'grade') return `Tingkat ${session.scope_ref || '—'}`;
		if (session.scope_type === 'school') return 'Seluruh sekolah';
		if (session.scope_type === 'custom') return session.scope_ref || 'Cohort khusus';
		return session.class_code || session.class_name || 'Per kelas';
	}

	function mixPolicyLabel(value: string) {
		if (value === 'same_class') return 'Tetap per kelas';
		if (value === 'mixed_scope') return 'Campur lintas scope';
		return 'Campur dalam tingkat';
	}

	function adaptiveMixPolicy(scopeType: string) {
		if (scopeType === 'class') return 'same_class';
		if (scopeType === 'grade') return 'same_grade';
		return 'mixed_scope';
	}

	function sessionActionLabel(status: string) {
		if (status === 'draft') return 'Draft';
		if (status === 'scheduled') return 'Terjadwal';
		if (status === 'active') return 'Berlangsung';
		if (status === 'finished') return 'Selesai';
		return 'Dibatalkan';
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
		if (ok) toast.success(msg);
		else toast.error(msg);
	}

	async function createSession() {
		if (!fPackageId || !fTitle || !fStart || !fEnd) return;
		if (fScopeType === 'class' && !fClassId) return;
		if (fScopeType === 'grade' && !fGradeLevel) return;
		if (fAllowCrossGrade && !fIsSpecialEvent) {
			showToast('Lintas tingkat hanya boleh untuk special event', false);
			return;
		}
		fBusy = true;
		try {
			const scopeRef = fScopeType === 'class' ? fClassId : fScopeType === 'grade' ? fGradeLevel : '';
			const res = await fetch('/api/cbt/sessions', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					package_id: fPackageId,
					class_id: fScopeType === 'class' ? fClassId : '',
					scope_type: fScopeType,
					scope_ref: scopeRef,
					mix_policy: adaptiveMixPolicy(fScopeType),
					assignment_mode: fAssignmentMode,
					allow_cross_grade: fAllowCrossGrade,
					is_special_event: fIsSpecialEvent,
					title: fTitle,
					scheduled_start: toRFC3339(fStart), scheduled_end: toRFC3339(fEnd),
					status: 'draft',
				}),
			});
			if (!res.ok) { const j = await res.json(); showToast(j.error ?? 'Gagal', false); return; }
			fPackageId = ''; fScopeType = 'class'; fClassId = ''; fGradeLevel = 'VII';
			fMixPolicy = 'same_class'; fAssignmentMode = 'random_balanced';
			fAllowCrossGrade = false; fIsSpecialEvent = false;
			fTitle = ''; fStart = ''; fEnd = '';
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

	async function enrollParticipants() {
		if (!enrollSession) return;
		if (enrollScopeType === 'class' && !enrollClassId) return;
		if (enrollScopeType === 'grade' && !enrollGradeLevel) return;
		enrollBusy = true;
		try {
			const payload =
				enrollScopeType === 'class'
					? { scope_type: 'class', class_id: enrollClassId }
					: enrollScopeType === 'grade'
						? { scope_type: 'grade', level: enrollGradeLevel }
						: { scope_type: 'school' };
			const res = await fetch(`/api/cbt/sessions/${enrollSession.id}/enroll`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(payload),
			});
			if (!res.ok) { showToast('Gagal mendaftarkan siswa', false); return; }
			showToast(`Peserta berhasil didaftarkan ke sesi "${enrollSession.title}"`);
			enrollSession = null;
			enrollScopeType = 'class';
			enrollClassId = '';
			enrollGradeLevel = 'VII';
			await load();
		} finally { enrollBusy = false; }
	}

	onMount(load);
</script>

<svelte:head><title>Sesi Ujian CBT — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div>
			<h1 class="text-2xl font-semibold text-slate-800">Sesi Ujian CBT</h1>
			<p class="text-sm text-slate-500 mt-1">Jadwalkan sesi per kelas, tingkat, atau seluruh sekolah dengan rooming yang fleksibel</p>
		</div>
		<Button onclick={() => (showForm = !showForm)}>
			{showForm ? 'Batal' : '+ Buat Sesi'}
		</Button>
	</div>

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
						<label for="session-package" class="text-xs text-slate-500 mb-1 block">Paket Soal <span class="text-red-500">*</span></label>
						<select id="session-package" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fPackageId}>
							<option value="">-- Pilih Paket --</option>
							{#each packages as p}
								<option value={p.id}>{p.title} ({p.subject_code})</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="session-scope" class="text-xs text-slate-500 mb-1 block">Scope peserta <span class="text-red-500">*</span></label>
						<select id="session-scope" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fScopeType}>
							<option value="class">Per kelas</option>
							<option value="grade">Per tingkat</option>
							<option value="school">Seluruh sekolah</option>
						</select>
					</div>
					{#if fScopeType === 'class'}
						<div>
							<label for="session-class" class="text-xs text-slate-500 mb-1 block">Kelas <span class="text-red-500">*</span></label>
							<select id="session-class" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fClassId}>
								<option value="">-- Pilih Kelas --</option>
								{#each classes as c}
									<option value={c.id}>{c.code} — {c.name}</option>
								{/each}
							</select>
						</div>
					{:else if fScopeType === 'grade'}
						<div>
							<label for="session-grade" class="text-xs text-slate-500 mb-1 block">Tingkat <span class="text-red-500">*</span></label>
							<select id="session-grade" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fGradeLevel}>
								<option value="VII">VII</option>
								<option value="VIII">VIII</option>
								<option value="IX">IX</option>
							</select>
						</div>
					{:else}
						<div class="rounded-md border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-900">
							Semua siswa aktif di sekolah dapat menjadi peserta sesi ini.
						</div>
					{/if}
					<div>
						<label for="session-mix-policy" class="text-xs text-slate-500 mb-1 block">Mix policy</label>
						<select id="session-mix-policy" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fMixPolicy}>
							<option value="same_class">Tetap per kelas</option>
							<option value="same_grade">Campur dalam tingkat</option>
							<option value="mixed_scope">Campur lintas scope</option>
						</select>
					</div>
					<div>
						<label for="session-assignment-mode" class="text-xs text-slate-500 mb-1 block">Mode alokasi ruangan</label>
						<select id="session-assignment-mode" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={fAssignmentMode}>
							<option value="random_balanced">Acak seimbang</option>
							<option value="manual">Manual</option>
							<option value="random_by_gender">Acak per gender</option>
							<option value="random_by_accommodation">Acak akomodasi khusus</option>
						</select>
					</div>
					<div class="sm:col-span-2">
						<label for="session-title" class="text-xs text-slate-500 mb-1 block">Nama Sesi <span class="text-red-500">*</span></label>
						<Input id="session-title" placeholder="mis: UTS Matematika VII A - Semester 1 2025" bind:value={fTitle} />
					</div>
					<div>
						<label for="session-start" class="text-xs text-slate-500 mb-1 block">Mulai <span class="text-red-500">*</span></label>
						<Input id="session-start" type="datetime-local" bind:value={fStart} />
					</div>
					<div>
						<label for="session-end" class="text-xs text-slate-500 mb-1 block">Selesai <span class="text-red-500">*</span></label>
						<Input id="session-end" type="datetime-local" bind:value={fEnd} />
					</div>
					<div class="sm:col-span-2 grid gap-3 sm:grid-cols-2">
						<label class="flex items-center gap-2 rounded-md border border-input px-3 py-2 text-sm text-slate-700">
							<input type="checkbox" bind:checked={fIsSpecialEvent} class="size-4 accent-emerald-700" />
							Tandai sebagai special event
						</label>
						<label class="flex items-center gap-2 rounded-md border border-input px-3 py-2 text-sm text-slate-700">
							<input type="checkbox" bind:checked={fAllowCrossGrade} class="size-4 accent-emerald-700" />
							Izinkan lintas tingkat
						</label>
					</div>
				</div>
				<div class="flex gap-2">
					<Button
						disabled={fBusy || !fPackageId || !fTitle || !fStart || !fEnd || (fScopeType === 'class' && !fClassId) || (fScopeType === 'grade' && !fGradeLevel)}
						onclick={createSession}
					>
						{fBusy ? 'Menyimpan...' : 'Buat Sesi'}
					</Button>
					<Button variant="outline" onclick={() => (showForm = false)}>Batal</Button>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	<!-- Enroll modal -->
	{#if enrollSession}
		<Card.Root class="border-green-200 bg-green-50">
			<Card.Header class="pb-2">
				<Card.Title class="text-base text-green-900">Daftarkan Siswa ke Sesi</Card.Title>
				<p class="text-sm text-green-700 mt-0.5">{enrollSession.title}</p>
			</Card.Header>
			<Card.Content class="space-y-3">
				<p class="text-sm text-slate-600">Tentukan cohort peserta untuk sesi ini. Ruangan tetap bisa diacak terpisah setelah peserta terdaftar.</p>
				<div>
					<label for="enroll-scope" class="text-xs text-slate-500 mb-1 block">Scope cohort</label>
					<select id="enroll-scope" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={enrollScopeType}>
						<option value="class">Per kelas</option>
						<option value="grade">Per tingkat</option>
						<option value="school">Seluruh sekolah</option>
					</select>
				</div>
				{#if enrollScopeType === 'class'}
					<div>
						<label for="enroll-class" class="text-xs text-slate-500 mb-1 block">Pilih Kelas</label>
						<select id="enroll-class" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={enrollClassId}>
							<option value="">-- Pilih Kelas --</option>
							{#each classes as c}
								<option value={c.id}>{c.code} — {c.name}</option>
							{/each}
						</select>
					</div>
				{:else if enrollScopeType === 'grade'}
					<div>
						<label for="enroll-grade" class="text-xs text-slate-500 mb-1 block">Pilih Tingkat</label>
						<select id="enroll-grade" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={enrollGradeLevel}>
							<option value="VII">VII</option>
							<option value="VIII">VIII</option>
							<option value="IX">IX</option>
						</select>
					</div>
				{:else}
					<div class="rounded-md border border-emerald-200 bg-white px-3 py-2 text-sm text-slate-700">
						Semua siswa aktif di sekolah akan didaftarkan ke sesi ini.
					</div>
				{/if}
				<div class="flex gap-2">
					<Button
						disabled={enrollBusy || (enrollScopeType === 'class' && !enrollClassId) || (enrollScopeType === 'grade' && !enrollGradeLevel)}
						onclick={enrollParticipants}
					>
						{enrollBusy ? 'Mendaftarkan...' : 'Daftarkan Siswa'}
					</Button>
					<Button variant="outline" onclick={() => { enrollSession = null; enrollScopeType = 'class'; enrollClassId = ''; enrollGradeLevel = 'VII'; }}>Batal</Button>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	{#if loading}
		<p class="text-sm text-slate-500">Memuat data...</p>
	{:else}
		<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
			<Card.Header class="pb-2">
				<Card.Title class="text-base">Daftar Sesi ({sessions.length})</Card.Title>
			</Card.Header>
			<Card.Content class="p-0">
				<div class="hidden overflow-x-auto lg:block">
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>Nama Sesi</Table.Head>
							<Table.Head>Paket</Table.Head>
							<Table.Head>Scope</Table.Head>
							<Table.Head>Jadwal Mulai</Table.Head>
							<Table.Head>Peserta</Table.Head>
							<Table.Head>Status</Table.Head>
							<Table.Head></Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each sessions as s (s.id)}
							<Table.Row>
								<Table.Cell class="font-medium max-w-48">
									<p class="truncate">{s.title}</p>
								</Table.Cell>
								<Table.Cell class="text-slate-500 text-sm truncate max-w-32">{s.package_title}</Table.Cell>
								<Table.Cell>
									<div class="space-y-1">
										<Badge variant="outline" class="text-xs">{scopeSummary(s)}</Badge>
										<p class="text-[11px] text-slate-500">{mixPolicyLabel(s.mix_policy)}</p>
									</div>
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
											<Button
												size="xs"
												variant="outline"
												onclick={() => {
													enrollSession = s;
													enrollScopeType = s.scope_type || 'class';
													enrollClassId = s.class_id;
													enrollGradeLevel = s.scope_type === 'grade' ? s.scope_ref : 'VII';
												}}
											>
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
				</div>

				<div class="grid gap-3 p-4 lg:hidden">
					{#each sessions as s (s.id)}
						<div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
							<div class="flex items-start justify-between gap-3">
								<div class="min-w-0">
									<p class="text-sm font-semibold text-slate-900">{s.title}</p>
									<p class="mt-1 text-xs text-slate-500">{s.package_title}</p>
								</div>
								<Badge class={statusClass(s.status)}>{sessionActionLabel(s.status)}</Badge>
							</div>
							<div class="mt-3 flex flex-wrap items-center gap-2">
								<Badge variant="outline" class="text-xs">{scopeSummary(s)}</Badge>
								<Badge variant="outline" class="text-xs">{mixPolicyLabel(s.mix_policy)}</Badge>
								<Badge variant="secondary">{s.participant_count} peserta</Badge>
							</div>
							<p class="mt-3 text-xs text-slate-500">{fmtDt(s.scheduled_start)}</p>
							<div class="mt-4 flex flex-wrap gap-2">
								{#if s.status === 'draft'}
									<Button
										size="sm"
										variant="outline"
										onclick={() => {
											enrollSession = s;
											enrollScopeType = s.scope_type || 'class';
											enrollClassId = s.class_id;
											enrollGradeLevel = s.scope_type === 'grade' ? s.scope_ref : 'VII';
										}}
									>
										Daftarkan
									</Button>
									<Button size="sm" onclick={() => updateStatus(s.id, 'scheduled')}>
										Jadwalkan
									</Button>
									<Button size="sm" variant="destructive" onclick={() => deleteSession(s.id, s.title)}>
										Hapus
									</Button>
								{:else if s.status === 'scheduled'}
									<Button size="sm" onclick={() => updateStatus(s.id, 'active')}>Mulai</Button>
									<Button size="sm" variant="outline" onclick={() => updateStatus(s.id, 'cancelled')}>Batalkan</Button>
								{:else if s.status === 'active'}
									<Button size="sm" onclick={() => updateStatus(s.id, 'finished')}>Selesaikan</Button>
								{/if}
								{#if s.status === 'finished' || s.status === 'active'}
									<a href="/cbt/sessions/{s.id}" class="inline-flex items-center rounded-md px-3 py-1.5 text-sm font-medium border border-input bg-background hover:bg-muted text-slate-700 transition-colors">
										Lihat Hasil
									</a>
								{/if}
							</div>
						</div>
					{:else}
						<div class="rounded-2xl border border-dashed border-slate-300 bg-slate-50 px-4 py-10 text-center text-sm text-slate-500">
							Belum ada sesi ujian
						</div>
					{/each}
				</div>
			</Card.Content>
		</Card.Root>
	{/if}
</div>
