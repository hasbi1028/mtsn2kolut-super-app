<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/state';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from '$lib/components/ui/sonner';

	type SessionInfo = {
		id: string; title: string; package_title: string; duration_minutes: number;
		class_name: string; class_code: string; event_id: string | null;
		scope_type?: string; scope_ref?: string;
		scheduled_start: string; scheduled_end: string; status: string;
	};
	type ResultRow = {
		participant_id: string; nis: string; nama: string; gender: string;
		submitted_at: string | null; score: string | null;
		room_name?: string; seat_no?: number | null;
		total_answers: number; correct_answers: number;
	};
	type Participant = {
		id: string; nis: string; nama: string; gender: string;
		token: string; room_name: string; room_id: string | null;
		seat_no?: number | null;
		submitted_at: string | null; score: string | null;
		app_switch_count: number; screenshot_attempt: number; suspicious_flag: boolean;
		last_heartbeat: string | null;
	};
	type Room = { id: string; room_name: string; capacity: number; participant_count: number; };
	type ProctoringRow = {
		participant_id: string; nis: string; nama: string;
		token: string; room_name: string; seat_no?: number | null;
		submitted_at: string | null; last_heartbeat: string | null;
		app_switch_count: number; screenshot_attempt: number;
		suspicious_flag: boolean; answered_count: number; score: string | null;
	};
	type UngradedEssay = {
		id: string; nis: string; nama: string;
		question_text: string; answer: string;
	};

	const sessionId = page.params.id;

	let activeTab = $state<'hasil' | 'peserta' | 'ruangan' | 'proctoring' | 'essay'>('hasil');
	let session = $state<SessionInfo | null>(null);
	let results = $state<ResultRow[]>([]);
	let participants = $state<Participant[]>([]);
	let rooms = $state<Room[]>([]);
	let proctoring = $state<ProctoringRow[]>([]);
	let essays = $state<UngradedEssay[]>([]);
	let loading = $state(true);
	let error = $state('');
	let scoreBusy = $state(false);
	let shuffleBusy = $state(false);
	let newRoomName = $state('');
	let newRoomCap = $state(30);
	let roomBusy = $state(false);
	let seatBusy = $state(false);
	let procInterval: ReturnType<typeof setInterval> | null = null;
	let gradeInput = $state<Record<string, number>>({});
	let seatInput = $state<Record<string, number>>({});
	let roomInput = $state<Record<string, string>>({});

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

	function fmtDt(iso: string | null) {
		if (!iso) return '—';
		return new Date(iso).toLocaleString('id-ID', {
			timeZone: 'Asia/Makassar', year: 'numeric', month: 'short',
			day: 'numeric', hour: '2-digit', minute: '2-digit',
		}) + ' WITA';
	}

	function fmtScore(score: string | null) {
		if (score === null || score === undefined || score === '') return '—';
		const n = parseFloat(score);
		return isNaN(n) ? '—' : n.toFixed(1);
	}

	function scoreClass(score: string | null) {
		if (!score) return 'text-slate-400';
		const n = parseFloat(score);
		if (n >= 75) return 'text-emerald-600 font-semibold';
		if (n >= 60) return 'text-amber-600 font-semibold';
		return 'text-red-600 font-semibold';
	}

	function heartbeatStatus(hb: string | null): { label: string; cls: string } {
		if (!hb) return { label: 'Belum login', cls: 'text-slate-400' };
		const diff = (Date.now() - new Date(hb).getTime()) / 1000;
		if (diff < 60) return { label: '🟢 Online', cls: 'text-emerald-600' };
		if (diff < 180) return { label: '🟡 Lambat', cls: 'text-amber-600' };
		return { label: '🔴 Offline', cls: 'text-red-600' };
	}

	let stats = $derived({
		total: results.length,
		submitted: results.filter(r => r.submitted_at).length,
		avgScore: results.length > 0
			? results.reduce((sum, r) => sum + (r.score ? parseFloat(r.score) : 0), 0) / results.length
			: 0,
		passing: results.filter(r => r.score && parseFloat(r.score) >= 75).length,
	});

	function showToast(msg: string, ok = true) {
		if (ok) toast.success(msg);
		else toast.error(msg);
	}

	async function load() {
		loading = true; error = '';
		try {
			const res = await fetch(`/api/cbt/sessions/${sessionId}/results`);
			const json = await res.json();
			const d = json.data ?? json;
			session = d.session ?? null;
			results = d.results ?? [];
		} catch {
			error = 'Gagal memuat hasil ujian';
		} finally {
			loading = false;
		}
	}

	async function loadParticipants() {
		const res = await fetch(`/api/cbt/sessions/${sessionId}/participants`);
		if (res.ok) {
			participants = await res.json();
			seatInput = Object.fromEntries(participants.map((participant) => [participant.id, participant.seat_no ?? 0]));
			roomInput = Object.fromEntries(participants.map((participant) => [participant.id, participant.room_id ?? '']));
		}
	}

	async function loadRooms() {
		const res = await fetch(`/api/cbt/sessions/${sessionId}/rooms`);
		if (res.ok) rooms = await res.json();
	}

	async function loadProctoring() {
		const res = await fetch(`/api/cbt/sessions/${sessionId}/proctoring`);
		if (res.ok) proctoring = await res.json();
	}

	async function loadEssays() {
		const res = await fetch(`/api/cbt/sessions/${sessionId}/ungraded-essays`);
		if (res.ok) {
			const data = await res.json();
			essays = data.data ?? data;
		}
	}

	async function switchTab(tab: typeof activeTab) {
		activeTab = tab;
		if (tab === 'peserta') { await loadRooms(); await loadParticipants(); }
		if (tab === 'ruangan') { await loadRooms(); await loadParticipants(); }
		if (tab === 'essay') await loadEssays();
		if (tab === 'proctoring') {
			await loadProctoring();
			if (!procInterval) {
				procInterval = setInterval(loadProctoring, 15000);
			}
		} else {
			if (procInterval) { clearInterval(procInterval); procInterval = null; }
		}
	}

	async function triggerScoring() {
		if (!confirm('Hitung ulang skor semua peserta?')) return;
		scoreBusy = true;
		try {
			const res = await fetch(`/api/cbt/sessions/${sessionId}/score`, { method: 'POST' });
			if (!res.ok) { showToast('Gagal menghitung skor', false); return; }
			showToast('Penilaian selesai — skor diperbarui');
			await load();
		} finally { scoreBusy = false; }
	}

	async function generateTokens() {
		const res = await fetch(`/api/cbt/sessions/${sessionId}/generate-tokens`, { method: 'POST' });
		if (res.ok) { showToast('Token berhasil digenerate'); await loadParticipants(); }
		else showToast('Gagal generate token', false);
	}

	async function regenerateToken(pid: string) {
		const res = await fetch(`/api/cbt/sessions/${sessionId}/participants/${pid}/regenerate-token`, { method: 'POST' });
		if (res.ok) { showToast('Token diperbarui'); await loadParticipants(); }
		else showToast('Gagal regenerate token', false);
	}

	function copyToken(token: string) {
		navigator.clipboard.writeText(token).then(() => showToast('Token disalin'));
	}

	async function createRoom() {
		if (!newRoomName.trim()) return;
		roomBusy = true;
		const res = await fetch(`/api/cbt/sessions/${sessionId}/rooms`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ room_name: newRoomName.trim(), capacity: newRoomCap }),
		});
		if (res.ok) {
			showToast(`Ruangan "${newRoomName}" ditambahkan`);
			newRoomName = ''; newRoomCap = 30;
			await loadRooms();
		} else showToast('Gagal tambah ruangan', false);
		roomBusy = false;
	}

	async function deleteRoom(rid: string) {
		if (!confirm('Hapus ruangan ini? Peserta di ruangan ini akan dilepas.')) return;
		await fetch(`/api/cbt/sessions/${sessionId}/rooms/${rid}`, { method: 'DELETE' });
		await loadRooms(); await loadParticipants();
	}

	async function shuffleRooms() {
		if (!confirm('Acak peserta ke ruangan secara random? Assignment ruangan sebelumnya akan direset.')) return;
		shuffleBusy = true;
		const res = await fetch(`/api/cbt/sessions/${sessionId}/shuffle-rooms`, { method: 'POST' });
		if (res.ok) { showToast('Peserta berhasil diacak ke ruangan'); await loadRooms(); await loadParticipants(); }
		else showToast('Gagal mengacak ruangan', false);
		shuffleBusy = false;
	}

	async function autoAssignSeats() {
		seatBusy = true;
		try {
			const res = await fetch(`/api/cbt/sessions/${sessionId}/seats/auto`, { method: 'POST' });
			if (res.ok) {
				showToast('Nomor meja berhasil diurutkan otomatis');
				await loadParticipants();
			} else {
				showToast('Gagal mengatur nomor meja otomatis', false);
			}
		} finally {
			seatBusy = false;
		}
	}

	async function assignSeat(pid: string) {
		const roomId = roomInput[pid];
		const seatNo = seatInput[pid];
		if (!roomId || !seatNo || seatNo <= 0) {
			showToast('Pilih ruangan dan isi nomor meja yang valid', false);
			return;
		}
		seatBusy = true;
		try {
			const res = await fetch(`/api/cbt/sessions/${sessionId}/participants/${pid}/seat`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ room_id: roomId, seat_no: seatNo }),
			});
			if (res.ok) {
				showToast('No meja peserta diperbarui');
				await loadParticipants();
			} else {
				showToast('Gagal menyimpan nomor meja', false);
			}
		} finally {
			seatBusy = false;
		}
	}

	async function flagParticipant(pid: string, flag: boolean) {
		await fetch(`/api/cbt/sessions/${sessionId}/participants/${pid}/flag`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ flag }),
		});
		await loadProctoring();
	}

	async function submitGrade(aid: string) {
		const score = gradeInput[aid];
		if (score === undefined || score < 0 || score > 100) {
			showToast('Nilai harus 0-100', false);
			return;
		}
		const res = await fetch(`/api/cbt/sessions/${sessionId}/answers/${aid}/grade-essay`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ manual_score: score }),
		});
		if (res.ok) {
			showToast('Nilai berhasil disimpan');
			await loadEssays();
		} else {
			showToast('Gagal menyimpan nilai', false);
		}
	}

	function exportCSV() {
		if (!session || results.length === 0) return;
		const header = 'NIS,Nama,L/P,Jawaban Masuk,Benar,Skor,Waktu Submit';
		const rows = results.map(r =>
			[r.nis, `"${r.nama}"`, r.gender, r.total_answers, r.correct_answers,
			 fmtScore(r.score), r.submitted_at ? fmtDt(r.submitted_at) : ''].join(',')
		);
		const csv = [header, ...rows].join('\n');
		const blob = new Blob([csv], { type: 'text/csv' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `hasil_${session.title.replace(/\s+/g, '_')}.csv`;
		a.click();
		URL.revokeObjectURL(url);
	}

	onMount(load);
	onDestroy(() => { if (procInterval) clearInterval(procInterval); });
</script>

<svelte:head>
	<title>{session?.title ?? 'Detail Sesi'} — MTSN 2 Kolut</title>
</svelte:head>

<div class="space-y-5 p-6">
	<!-- Breadcrumb -->
	<div class="flex items-center gap-2 text-sm text-slate-500">
		<a href="/cbt/sessions" class="hover:text-slate-700">Sesi Ujian</a>
		<span>/</span>
		<span class="text-slate-700 font-medium truncate max-w-xs">{session?.title ?? '...'}</span>
	</div>

	{#if error}<div class="rounded-md bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-800">{error}</div>{/if}

	{#if loading}
		<p class="text-sm text-slate-500">Memuat data...</p>
	{:else if session}
		<!-- Session header -->
		<div class="flex items-start justify-between gap-4 flex-wrap">
			<div>
				<h1 class="text-2xl font-semibold text-[oklch(0.38_0.13_145)]">{session.title}</h1>
				<div class="flex flex-wrap gap-2 mt-2 text-sm text-slate-500">
					<span>{session.package_title}</span>
					{#if session.class_code}<span>· Kelas {session.class_code}</span>{/if}
					<span>· {session.duration_minutes} menit</span>
					<span>· {fmtDt(session.scheduled_start)}</span>
				</div>
			</div>
				<div class="flex items-center gap-2 flex-wrap">
					<Badge class={statusClass(session.status)}>{statusLabel[session.status] ?? session.status}</Badge>
				{#if session.status === 'finished' || session.status === 'active'}
					<Button size="sm" variant="outline" disabled={scoreBusy} onclick={triggerScoring}>
						{scoreBusy ? 'Menghitung...' : '⟳ Hitung Skor'}
					</Button>
				{/if}
					{#if results.length > 0}
						<Button size="sm" variant="outline" onclick={exportCSV}>↓ CSV</Button>
					{/if}
					<a href={`/cbt/sessions/${sessionId}/minutes`} class="inline-flex items-center rounded-md border border-input bg-background px-3 py-2 text-sm font-medium text-slate-700 hover:bg-muted">
						Berita Acara
					</a>
					{#if session.event_id}
						<a href={`/cbt/events/${session.event_id}/exam-cards`} class="inline-flex items-center rounded-md border border-input bg-background px-3 py-2 text-sm font-medium text-slate-700 hover:bg-muted">
							Kartu Ujian Event
						</a>
					{/if}
				</div>
		</div>

		<!-- Stats -->
		<div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
			{#each [
				{ label: 'Total Peserta', val: stats.total.toString() },
				{ label: 'Sudah Submit', val: stats.submitted.toString() },
				{ label: 'Rata-rata Skor', val: stats.total > 0 ? stats.avgScore.toFixed(1) : '—' },
				{ label: 'Lulus (≥75)', val: `${stats.passing} / ${stats.submitted}` },
			] as s}
				<Card.Root class="border-green-100">
					<Card.Content class="pt-4 pb-3 px-4">
						<p class="text-xs text-slate-500 mb-1">{s.label}</p>
						<p class="text-2xl font-bold text-[oklch(0.38_0.13_145)]">{s.val}</p>
					</Card.Content>
				</Card.Root>
			{/each}
		</div>

		<!-- Tabs -->
		<div class="border-b border-green-100">
			<nav class="flex gap-1">
				{#each [
					{ id: 'hasil', label: 'Hasil Ujian' },
					{ id: 'peserta', label: 'Peserta & Token' },
					{ id: 'ruangan', label: 'Ruangan' },
					{ id: 'proctoring', label: 'Proctoring' },
					{ id: 'essay', label: 'Koreksi Essay' },
				] as tab}
					<button
						onclick={() => switchTab(tab.id as any)}
						class="px-4 py-2 text-sm font-medium border-b-2 transition-colors {activeTab === tab.id
							? 'border-[oklch(0.38_0.13_145)] text-[oklch(0.38_0.13_145)]'
							: 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'}"
					>
						{tab.label}
					</button>
				{/each}
			</nav>
		</div>

		<!-- Tab: Hasil -->
		{#if activeTab === 'hasil'}
			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Title class="text-base">Daftar Nilai ({results.length} peserta)</Card.Title>
				</Card.Header>
				<Card.Content class="p-0 overflow-x-auto">
					<Table.Root>
						<Table.Header>
							<Table.Row class="bg-green-50">
								<Table.Head class="w-8">#</Table.Head>
								<Table.Head>NIS</Table.Head>
								<Table.Head>Nama</Table.Head>
								<Table.Head>L/P</Table.Head>
								<Table.Head class="text-center">Jawaban</Table.Head>
								<Table.Head class="text-center">Benar</Table.Head>
								<Table.Head class="text-center">Skor</Table.Head>
								<Table.Head>Waktu Submit</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each results as r, i}
								<Table.Row>
									<Table.Cell class="text-slate-400 text-xs">{i + 1}</Table.Cell>
									<Table.Cell class="font-mono text-sm">{r.nis}</Table.Cell>
									<Table.Cell class="font-medium">{r.nama}</Table.Cell>
									<Table.Cell><Badge variant="outline" class="text-xs">{r.gender}</Badge></Table.Cell>
									<Table.Cell class="text-center text-sm">{r.total_answers}</Table.Cell>
									<Table.Cell class="text-center text-sm">{r.correct_answers}</Table.Cell>
									<Table.Cell class="text-center"><span class={scoreClass(r.score)}>{fmtScore(r.score)}</span></Table.Cell>
									<Table.Cell class="text-slate-500 text-xs whitespace-nowrap">
										{r.submitted_at ? fmtDt(r.submitted_at) : '<span class="text-slate-300">Belum submit</span>'}
									</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={8} class="text-center text-slate-400 py-10">Belum ada data nilai</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</Card.Content>
			</Card.Root>

		<!-- Tab: Peserta & Token -->
		{:else if activeTab === 'peserta'}
			<div class="flex gap-2 flex-wrap">
				<Button variant="outline" size="sm" onclick={generateTokens}>⚡ Generate Token Massal</Button>
				<Button variant="outline" size="sm" onclick={loadParticipants}>↻ Refresh</Button>
			</div>
			<Card.Root>
				<Card.Content class="p-0 overflow-x-auto">
					<Table.Root>
						<Table.Header>
							<Table.Row class="bg-green-50">
								<Table.Head>NIS</Table.Head>
								<Table.Head>Nama</Table.Head>
								<Table.Head>L/P</Table.Head>
								<Table.Head>Ruangan</Table.Head>
								<Table.Head>No Meja</Table.Head>
								<Table.Head>Token</Table.Head>
								<Table.Head>Status</Table.Head>
								<Table.Head class="text-right">Aksi</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
								{#each participants as p}
									<Table.Row class={p.suspicious_flag ? 'bg-red-50' : 'hover:bg-green-50/30'}>
									<Table.Cell class="font-mono text-sm">{p.nis}</Table.Cell>
									<Table.Cell class="font-medium">
										{p.nama}
										{#if p.suspicious_flag}<span class="ml-1 text-red-500 text-xs">⚑ Dicurigai</span>{/if}
									</Table.Cell>
										<Table.Cell><Badge variant="outline" class="text-xs">{p.gender}</Badge></Table.Cell>
										<Table.Cell class="text-sm text-muted-foreground">{p.room_name || '—'}</Table.Cell>
										<Table.Cell class="text-sm text-muted-foreground">{p.seat_no ?? '—'}</Table.Cell>
										<Table.Cell>
										{#if p.token}
											<code class="bg-green-50 text-green-800 px-2 py-0.5 rounded text-xs font-mono border border-green-200">
												{p.token}
											</code>
										{:else}
											<span class="text-slate-400 text-xs">—</span>
										{/if}
									</Table.Cell>
									<Table.Cell>
										{#if p.submitted_at}
											<Badge variant="outline" class="text-xs bg-slate-100 text-slate-500">Submit</Badge>
										{:else}
											<Badge variant="outline" class="text-xs bg-amber-50 text-amber-700 border-amber-200">Belum</Badge>
										{/if}
									</Table.Cell>
										<Table.Cell class="text-right">
											<div class="flex items-center justify-end gap-1">
												<select bind:value={roomInput[p.id]} class="h-8 rounded-md border border-input bg-background px-2 text-xs">
													<option value="">Ruangan</option>
													{#each rooms as room}
														<option value={room.id}>{room.room_name}</option>
													{/each}
												</select>
												<Input bind:value={seatInput[p.id]} type="number" min="1" class="h-8 w-16" />
												<Button variant="outline" size="sm" onclick={() => assignSeat(p.id)} disabled={seatBusy}>Simpan</Button>
												{#if p.token}
													<Button variant="outline" size="sm" onclick={() => copyToken(p.token)}>Salin</Button>
												{/if}
											<Button variant="outline" size="sm" onclick={() => regenerateToken(p.id)}>Regenerate</Button>
										</div>
									</Table.Cell>
								</Table.Row>
								{:else}
									<Table.Row>
										<Table.Cell colspan={8} class="text-center text-slate-400 py-8">Belum ada peserta</Table.Cell>
									</Table.Row>
								{/each}
						</Table.Body>
					</Table.Root>
				</Card.Content>
			</Card.Root>

		<!-- Tab: Ruangan -->
		{:else if activeTab === 'ruangan'}
			<Card.Root class="border-green-200">
				<Card.Header class="pb-3">
					<Card.Title class="text-base">Tambah Ruangan Baru</Card.Title>
				</Card.Header>
				<Card.Content>
					<div class="flex gap-3 flex-wrap items-end">
						<div>
							<label for="r-name" class="block text-sm font-medium mb-1">Nama Ruangan</label>
							<Input id="r-name" bind:value={newRoomName} placeholder="Ruang 1 / Lab Komputer A" class="w-48" />
						</div>
						<div>
							<label for="r-cap" class="block text-sm font-medium mb-1">Kapasitas</label>
							<Input id="r-cap" type="number" bind:value={newRoomCap} min={1} max={100} class="w-24" />
						</div>
						<Button onclick={createRoom} disabled={roomBusy || !newRoomName.trim()}>
							{roomBusy ? 'Menyimpan...' : '+ Tambah Ruangan'}
						</Button>
							{#if rooms.length > 0}
								<Button variant="outline" disabled={shuffleBusy} onclick={shuffleRooms}
									class="border-amber-300 text-amber-700 hover:bg-amber-50">
									{shuffleBusy ? 'Mengacak...' : '🔀 Acak Peserta ke Ruangan'}
								</Button>
								<Button variant="outline" disabled={seatBusy} onclick={autoAssignSeats}>
									{seatBusy ? 'Mengatur...' : '🪑 Atur No Meja'}
								</Button>
							{/if}
					</div>
				</Card.Content>
			</Card.Root>

			{#if rooms.length > 0}
				<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 mt-4">
					{#each rooms as room}
						<Card.Root class="border-green-100">
							<Card.Content class="p-4">
								<div class="flex items-start justify-between">
									<div>
										<div class="font-semibold text-[oklch(0.38_0.13_145)]">{room.room_name}</div>
										<div class="text-sm text-muted-foreground mt-1">
											Kapasitas: {room.capacity} · Terisi: {room.participant_count}
										</div>
										<div class="mt-2 h-2 rounded-full bg-green-100 overflow-hidden">
											<div class="h-full bg-[oklch(0.38_0.13_145)] rounded-full transition-all"
												style="width: {Math.min(100, (room.participant_count / room.capacity) * 100)}%">
											</div>
										</div>
									</div>
									<Button variant="outline" size="sm"
										class="border-red-200 text-red-600 hover:bg-red-50 ml-3"
										onclick={() => deleteRoom(room.id)}>Hapus</Button>
								</div>
							</Card.Content>
						</Card.Root>
					{/each}
				</div>

				<!-- Participants by room -->
				<Card.Root class="mt-4">
					<Card.Header class="pb-2">
						<Card.Title class="text-base">Peserta per Ruangan</Card.Title>
					</Card.Header>
					<Card.Content class="p-0 overflow-x-auto">
						<Table.Root>
							<Table.Header>
								<Table.Row class="bg-green-50">
									<Table.Head>NIS</Table.Head>
									<Table.Head>Nama</Table.Head>
										<Table.Head>L/P</Table.Head>
										<Table.Head>Ruangan</Table.Head>
										<Table.Head>No Meja</Table.Head>
										<Table.Head>Token</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each participants as p}
									<Table.Row>
										<Table.Cell class="font-mono text-sm">{p.nis}</Table.Cell>
										<Table.Cell class="font-medium">{p.nama}</Table.Cell>
										<Table.Cell><Badge variant="outline" class="text-xs">{p.gender}</Badge></Table.Cell>
										{#if p.room_name}
											<Table.Cell class="text-sm">{p.room_name}</Table.Cell>
											{:else}
												<Table.Cell class="text-sm text-slate-400">Belum ditentukan</Table.Cell>
											{/if}
											<Table.Cell class="text-sm text-slate-600">{p.seat_no ?? '—'}</Table.Cell>
											<Table.Cell>
											{#if p.token}
												<code class="bg-green-50 text-green-800 px-2 py-0.5 rounded text-xs font-mono border border-green-200">{p.token}</code>
											{:else}
												<span class="text-slate-400 text-xs">—</span>
											{/if}
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					</Card.Content>
				</Card.Root>
			{:else}
				<div class="rounded-lg border border-dashed border-green-200 p-8 text-center text-muted-foreground text-sm mt-4">
					Belum ada ruangan. Tambah ruangan di atas, lalu klik "Acak Peserta ke Ruangan".
				</div>
			{/if}

		<!-- Tab: Proctoring -->
		{:else if activeTab === 'proctoring'}
			<div class="flex items-center justify-between mb-4">
				<p class="text-sm text-muted-foreground">Auto-refresh setiap 15 detik</p>
				<Button variant="outline" size="sm" onclick={loadProctoring}>↻ Refresh Sekarang</Button>
			</div>
			<Card.Root>
				<Card.Content class="p-0 overflow-x-auto">
					<Table.Root>
						<Table.Header>
							<Table.Row class="bg-green-50">
								<Table.Head>Nama</Table.Head>
								<Table.Head>Ruangan</Table.Head>
								<Table.Head>Status</Table.Head>
								<Table.Head class="text-center">Dijawab</Table.Head>
								<Table.Head class="text-center">App Switch</Table.Head>
								<Table.Head class="text-center">Screenshot</Table.Head>
								<Table.Head>Submit</Table.Head>
								<Table.Head class="text-right">Flag</Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
								{#each proctoring as p}
									{@const hb = heartbeatStatus(p.last_heartbeat)}
									<Table.Row class={p.suspicious_flag ? 'bg-red-50' : p.app_switch_count >= 3 ? 'bg-amber-50/50' : 'hover:bg-green-50/30'}>
									<Table.Cell class="font-medium">
										{p.nama}
										<div class="text-xs text-muted-foreground font-mono">{p.nis}</div>
									</Table.Cell>
									<Table.Cell class="text-sm text-muted-foreground">{p.room_name || '—'}</Table.Cell>
									<Table.Cell>
										<span class="text-sm font-medium {hb.cls}">{p.submitted_at ? '✅ Submit' : hb.label}</span>
									</Table.Cell>
									<Table.Cell class="text-center font-mono text-sm">{p.answered_count}</Table.Cell>
									<Table.Cell class="text-center">
										<span class="font-mono text-sm {p.app_switch_count >= 3 ? 'text-red-600 font-bold' : 'text-slate-600'}">
											{p.app_switch_count}x
										</span>
									</Table.Cell>
									<Table.Cell class="text-center">
										<span class="font-mono text-sm {p.screenshot_attempt > 0 ? 'text-amber-600 font-semibold' : 'text-slate-400'}">
											{p.screenshot_attempt}x
										</span>
									</Table.Cell>
									<Table.Cell class="text-xs text-muted-foreground whitespace-nowrap">
										{p.submitted_at ? fmtDt(p.submitted_at) : '—'}
									</Table.Cell>
									<Table.Cell class="text-right">
											<Button
												variant="outline" size="sm"
												class={p.suspicious_flag ? 'border-red-400 text-red-700 bg-red-50' : 'border-slate-200 text-slate-500'}
												onclick={() => flagParticipant(p.participant_id, !p.suspicious_flag)}>
											{p.suspicious_flag ? '⚑ Unflag' : '⚐ Flag'}
										</Button>
									</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={8} class="text-center text-slate-400 py-8">Belum ada data proctoring</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</Card.Content>
			</Card.Root>

		<!-- Tab: Essay Grading -->
		{:else if activeTab === 'essay'}
			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Title class="text-base">Koreksi Jawaban Essay ({essays.length} belum dinilai)</Card.Title>
				</Card.Header>
				<Card.Content class="p-0 overflow-x-auto">
					<Table.Root>
						<Table.Header>
							<Table.Row class="bg-green-50">
								<Table.Head>Siswa</Table.Head>
								<Table.Head>Pertanyaan</Table.Head>
								<Table.Head>Jawaban Siswa</Table.Head>
								<Table.Head class="w-32">Nilai (0-100)</Table.Head>
								<Table.Head></Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each essays as e}
								<Table.Row>
									<Table.Cell>
										<div class="font-medium">{e.nama}</div>
										<div class="text-xs text-slate-500 font-mono">{e.nis}</div>
									</Table.Cell>
									<Table.Cell class="max-w-xs text-sm">{e.question_text}</Table.Cell>
									<Table.Cell class="max-w-sm">
										<div class="rounded bg-slate-50 p-2 text-sm border border-slate-200 whitespace-pre-wrap">{e.answer}</div>
									</Table.Cell>
									<Table.Cell>
										<Input type="number" min="0" max="100" bind:value={gradeInput[e.id]} placeholder="0-100" class="w-24 h-8" />
									</Table.Cell>
									<Table.Cell>
										<Button size="sm" onclick={() => submitGrade(e.id)}>Simpan</Button>
									</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row>
									<Table.Cell colspan={5} class="text-center text-slate-400 py-12">
										Tidak ada jawaban essay yang perlu dikoreksi.
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</Card.Content>
			</Card.Root>
		{/if}
	{/if}
</div>
