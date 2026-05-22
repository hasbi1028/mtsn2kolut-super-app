<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from '$lib/components/ui/sonner';
	import { clientApiPath, readClientApiData, readClientJson } from '$lib/client/api';
	import StudentIdCardTemplate from '$lib/components/student-id-card/StudentIdCardTemplate.svelte';

	type Student = { id: string; nama: string; nis?: string; nisn?: string; class_name?: string; class_code?: string; is_active?: boolean };
	type StudentCard = {
		id: string;
		student_id: string;
		card_no: string;
		status: string;
		token_hint?: string;
		created_at?: string;
		updated_at?: string;
		printed_at?: string | null;
		revoked_at?: string | null;
		revoked_reason?: string | null;
		nis?: string;
		nisn?: string;
		nama?: string;
		class_name?: string;
		class_code?: string;
	};
	type HistoryItem = { id: string; event_type?: string; action?: string; source?: string; target?: string; created_at?: string; metadata?: unknown };
	type IssueResult = { card: StudentCard; qr_token?: string; qr_url?: string };

	let cards = $state<StudentCard[]>([]);
	let students = $state<Student[]>([]);
	let loading = $state(true);
	let busy = $state('');
	let search = $state('');
	let statusFilter = $state('');
	let selectedStudentId = $state('');
	let selectedCard = $state<StudentCard | null>(null);
	let events = $state<HistoryItem[]>([]);
	let auditLogs = $state<HistoryItem[]>([]);
	let detailLoading = $state(false);
	let reissueReason = $state('');
	let lastIssued = $state<IssueResult | null>(null);
	let lastIssuedQrReady = $state(false);

	const normalizedSearch = $derived(search.trim().toLowerCase());
	const filteredCards = $derived(cards.filter((card) => {
		const matchesStatus = !statusFilter || card.status === statusFilter;
		if (!matchesStatus) return false;
		if (!normalizedSearch) return true;
		return [card.card_no, card.nama, card.nis, card.nisn, card.class_name, card.class_code]
			.some((value) => String(value ?? '').toLowerCase().includes(normalizedSearch));
	}));
	const activeCount = $derived(cards.filter((card) => card.status === 'active').length);
	const lastIssuedHasQr = $derived(Boolean(lastIssued?.qr_url || lastIssued?.qr_token));
	const lastIssuedCanPrint = $derived(lastIssuedHasQr && lastIssuedQrReady);

	function formatDate(value?: string | null) {
		if (!value) return '-';
		return new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium', timeStyle: 'short', timeZone: 'Asia/Makassar' }).format(new Date(value));
	}

	function statusLabel(status: string) {
		return ({ active: 'Aktif', printed: 'Tercetak', lost: 'Hilang', revoked: 'Dicabut', replaced: 'Diganti', expired: 'Kedaluwarsa', suspended: 'Ditahan' } as Record<string, string>)[status] ?? status;
	}

	function statusVariant(status: string) {
		return status === 'active' ? 'default' : status === 'printed' ? 'secondary' : 'destructive';
	}

	function printLastIssued() {
		if (!lastIssuedCanPrint) {
			toast.error('QR aman belum tersedia. Jangan cetak kartu ini.');
			return;
		}
		document.body.classList.add('student-card-print-mode');
		window.print();
		setTimeout(() => document.body.classList.remove('student-card-print-mode'), 500);
	}

	async function load() {
		loading = true;
		const [cardResult, studentResult] = await Promise.allSettled([
			fetch('/api/kesiswaan/kartu-siswa?limit=200').then((res) => readClientApiData<StudentCard[]>(res, 'Gagal memuat kartu siswa')),
			fetch('/api/students').then((res) => readClientApiData<Student[]>(res, 'Gagal memuat siswa'))
		]);
		if (cardResult.status === 'fulfilled') {
			cards = cardResult.value ?? [];
			if (selectedCard) selectedCard = cards.find((card) => card.id === selectedCard?.id) ?? selectedCard;
		} else {
			toast.error(cardResult.reason?.message || 'Gagal memuat kartu siswa');
		}
		if (studentResult.status === 'fulfilled') {
			students = studentResult.value ?? [];
		} else {
			toast.error(studentResult.reason?.message || 'Gagal memuat siswa');
		}
		loading = false;
	}

	async function generateCard() {
		if (!selectedStudentId) {
			toast.error('Pilih siswa terlebih dahulu.');
			return;
		}
		busy = 'generate';
		try {
			lastIssuedQrReady = false;
			lastIssued = await readClientApiData<IssueResult>(await fetch('/api/kesiswaan/kartu-siswa', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ student_id: selectedStudentId })
			}), 'Gagal membuat kartu siswa');
			toast.success('Kartu siswa berhasil dibuat. QR siap dipreview/cetak.');
			selectedStudentId = lastIssued?.card?.student_id ?? selectedStudentId;
			await load();
		} catch (error) {
			toast.error((error as Error).message || 'Gagal membuat kartu siswa');
		} finally {
			busy = '';
		}
	}

	async function updateStatus(card: StudentCard, status: string) {
		const reason = prompt(`Alasan perubahan status menjadi ${statusLabel(status)}:`, card.revoked_reason ?? '') ?? '';
		busy = `status:${card.id}`;
		try {
			await readClientJson(await fetch(clientApiPath`/api/kesiswaan/kartu-siswa/${card.id}/status`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ status, reason })
			}));
			toast.success('Status kartu diperbarui.');
			await load();
		} catch (error) {
			toast.error((error as Error).message || 'Gagal mengubah status kartu');
		} finally {
			busy = '';
		}
	}

	async function markPrinted(card: StudentCard) {
		busy = `print:${card.id}`;
		try {
			await readClientJson(await fetch(clientApiPath`/api/kesiswaan/kartu-siswa/${card.id}/print-event`, { method: 'POST' }));
			toast.success('Kartu ditandai sudah dicetak.');
			await load();
		} catch (error) {
			toast.error((error as Error).message || 'Gagal menandai cetak');
		} finally {
			busy = '';
		}
	}

	async function reissue(card: StudentCard) {
		if (!reissueReason.trim()) {
			toast.error('Isi alasan cetak ulang/penggantian kartu.');
			return;
		}
		busy = `reissue:${card.id}`;
		try {
			lastIssuedQrReady = false;
			lastIssued = await readClientApiData<IssueResult>(await fetch(clientApiPath`/api/kesiswaan/kartu-siswa/${card.id}/reissue`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ reason: reissueReason.trim() })
			}), 'Gagal membuat kartu pengganti');
			toast.success('Kartu pengganti berhasil dibuat. QR baru siap dipreview/cetak.');
			reissueReason = '';
			selectedCard = null;
			await load();
		} catch (error) {
			toast.error((error as Error).message || 'Gagal membuat kartu pengganti');
		} finally {
			busy = '';
		}
	}

	async function openDetail(card: StudentCard) {
		selectedCard = card;
		detailLoading = true;
		events = [];
		auditLogs = [];
		try {
			const [eventItems, auditItems] = await Promise.all([
				fetch(clientApiPath`/api/kesiswaan/kartu-siswa/${card.id}/events?limit=20`).then((res) => readClientApiData<HistoryItem[]>(res, 'Gagal memuat event kartu')),
				fetch(clientApiPath`/api/kesiswaan/kartu-siswa/${card.id}/audit-logs?limit=20`).then((res) => readClientApiData<HistoryItem[]>(res, 'Gagal memuat audit kartu'))
			]);
			events = eventItems ?? [];
			auditLogs = auditItems ?? [];
		} catch (error) {
			toast.error((error as Error).message || 'Gagal memuat detail kartu');
		} finally {
			detailLoading = false;
		}
	}

	onMount(load);
</script>

<div class="space-y-6 p-4 md:p-6">
	<div class="flex flex-col gap-2 md:flex-row md:items-end md:justify-between">
		<div>
			<p class="text-sm text-muted-foreground">Kesiswaan</p>
			<h1 class="text-2xl font-semibold tracking-tight">Kartu Siswa</h1>
			<p class="text-sm text-muted-foreground">Kelola penerbitan, status, cetak ulang, event, dan audit kartu siswa terpadu.</p>
		</div>

		<Button onclick={load} disabled={loading}>Muat ulang</Button>
	</div>

	<div class="grid gap-4 md:grid-cols-3">
		<Card.Root><Card.Header><Card.Title>Total kartu</Card.Title><Card.Description>{cards.length} kartu terdata</Card.Description></Card.Header></Card.Root>
		<Card.Root><Card.Header><Card.Title>Aktif</Card.Title><Card.Description>{activeCount} kartu aktif</Card.Description></Card.Header></Card.Root>
		<Card.Root><Card.Header><Card.Title>Terfilter</Card.Title><Card.Description>{filteredCards.length} kartu sesuai filter</Card.Description></Card.Header></Card.Root>
	</div>

	<Card.Root>
		<Card.Header>
			<Card.Title>Terbitkan kartu</Card.Title>
			<Card.Description>Pilih siswa lalu buat kartu baru. Backend akan menolak jika aturan penerbitan tidak terpenuhi.</Card.Description>
		</Card.Header>
		<Card.Content>
			<div class="flex flex-col gap-3 md:flex-row">
				<select bind:value={selectedStudentId} class="h-10 flex-1 rounded-md border bg-background px-3 text-sm">
					<option value="">Pilih siswa...</option>
					{#each students as student}
						<option value={student.id}>{student.nama} — {student.nis || '-'} {student.class_name ? `(${student.class_name})` : ''}</option>
					{/each}
				</select>
				<Button onclick={generateCard} disabled={busy === 'generate'}>{busy === 'generate' ? 'Membuat...' : 'Generate kartu'}</Button>
			</div>
			{#if lastIssued}
				<div class="mt-5 rounded-xl border bg-muted/20 p-4">
					<div class="mb-3 flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
						<div>
							<h3 class="font-semibold">Preview kartu terakhir dibuat</h3>
							<p class="text-xs text-muted-foreground">QR hanya dicetak sebagai kode gambar. Token mentah tidak ditampilkan sebagai teks. Cetak hanya jika indikator QR siap.</p>
						</div>
						<div class="flex gap-2">
							<Button variant="outline" disabled={!lastIssuedCanPrint} onclick={printLastIssued}>{lastIssuedCanPrint ? 'Cetak kartu' : 'Menyiapkan QR...'}</Button>
							<Button variant="ghost" onclick={() => { lastIssued = null; lastIssuedQrReady = false; }}>Tutup preview</Button>
						</div>
					</div>
					{#if !lastIssuedHasQr}
						<div class="mb-3 rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-800">QR aman belum tersedia. Jangan cetak kartu ini; lakukan reissue/generate ulang jika perlu.</div>
					{/if}
					<div class="student-card-print-area">
						<StudentIdCardTemplate card={{ ...lastIssued.card, qr_url: lastIssued.qr_url, qr_token: lastIssued.qr_token }} print onQrReady={(ready) => lastIssuedQrReady = ready} />
					</div>
				</div>
			{/if}
		</Card.Content>
	</Card.Root>

	<Card.Root>
		<Card.Header>
			<Card.Title>Daftar kartu</Card.Title>
			<Card.Description>Filter cepat berdasarkan nama/NIS/nomor kartu dan status.</Card.Description>
		</Card.Header>
		<Card.Content>
			<div class="mb-4 grid gap-3 md:grid-cols-[1fr_220px]">
				<Input placeholder="Cari nama, NIS, kelas, atau nomor kartu" bind:value={search} />
				<select bind:value={statusFilter} class="h-10 rounded-md border bg-background px-3 text-sm">
					<option value="">Semua status</option>
					<option value="active">Aktif</option>
					<option value="printed">Tercetak</option>
					<option value="lost">Hilang</option>
					<option value="revoked">Dicabut</option>
					<option value="replaced">Diganti</option>
					<option value="suspended">Ditahan</option>
				</select>
			</div>

			{#if loading}
				<div class="rounded-lg border p-6 text-sm text-muted-foreground">Memuat kartu siswa...</div>
			{:else if filteredCards.length === 0}
				<div class="rounded-lg border p-6 text-sm text-muted-foreground">Belum ada kartu sesuai filter.</div>
			{:else}
				<div class="overflow-x-auto rounded-lg border">
					<table class="w-full text-sm">
						<thead class="bg-muted/50 text-left">
							<tr>
								<th class="p-3 font-medium">Siswa</th>
								<th class="p-3 font-medium">Nomor kartu</th>
								<th class="p-3 font-medium">Status</th>
								<th class="p-3 font-medium">Cetak</th>
								<th class="p-3 font-medium">Dibuat</th>
								<th class="p-3 font-medium">Aksi</th>
							</tr>
						</thead>
						<tbody>
							{#each filteredCards as card (card.id)}
								<tr class="border-t align-top">
									<td class="p-3"><div class="font-medium">{card.nama || '-'}</div><div class="text-xs text-muted-foreground">{card.nis || '-'} · {card.class_name || card.class_code || '-'}</div></td>
									<td class="p-3"><code>{card.card_no}</code><div class="text-xs text-muted-foreground">Hint: {card.token_hint || '-'}</div></td>
									<td class="p-3"><Badge variant={statusVariant(card.status)}>{statusLabel(card.status)}</Badge></td>
									<td class="p-3">{formatDate(card.printed_at)}</td>
									<td class="p-3">{formatDate(card.created_at)}</td>
									<td class="p-3">
										<div class="flex flex-wrap gap-2">
											<Button size="sm" variant="outline" onclick={() => openDetail(card)}>Detail</Button>
											<Button size="sm" variant="outline" disabled={busy === `print:${card.id}`} onclick={() => markPrinted(card)}>Tandai tercetak</Button>
											<select class="h-9 rounded-md border bg-background px-2 text-xs" disabled={busy === `status:${card.id}`} onchange={(event) => updateStatus(card, event.currentTarget.value)}>
												<option value="">Ubah status</option>
												<option value="active">Aktif</option>
												<option value="lost">Hilang</option>
												<option value="revoked">Dicabut</option>
												<option value="suspended">Ditahan</option>
											</select>
										</div>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</Card.Content>
	</Card.Root>

	{#if selectedCard}
		<Card.Root>
			<Card.Header>
				<div class="flex items-start justify-between gap-4">
					<div>
						<Card.Title>Detail kartu {selectedCard.card_no}</Card.Title>
						<Card.Description>{selectedCard.nama || '-'} · {statusLabel(selectedCard.status)}</Card.Description>
					</div>
					<Button variant="ghost" onclick={() => selectedCard = null}>Tutup</Button>
				</div>
			</Card.Header>
			<Card.Content class="space-y-5">
				<div class="grid gap-3 md:grid-cols-4 text-sm">
					<div><div class="text-muted-foreground">NIS/NISN</div><div>{selectedCard.nis || '-'} / {selectedCard.nisn || '-'}</div></div>
					<div><div class="text-muted-foreground">Rombel</div><div>{selectedCard.class_name || selectedCard.class_code || '-'}</div></div>
					<div><div class="text-muted-foreground">Printed at</div><div>{formatDate(selectedCard.printed_at)}</div></div>
					<div><div class="text-muted-foreground">Revoked reason</div><div>{selectedCard.revoked_reason || '-'}</div></div>
				</div>

				<div class="rounded-lg border p-4">
					<h3 class="mb-2 font-medium">Cetak ulang / reissue</h3>
					<div class="flex flex-col gap-2 md:flex-row">
						<Input placeholder="Alasan reissue, misalnya kartu hilang/rusak" bind:value={reissueReason} />
						<Button disabled={busy === `reissue:${selectedCard.id}`} onclick={() => reissue(selectedCard!)}>Reissue</Button>
					</div>
				</div>

				{#if detailLoading}
					<div class="text-sm text-muted-foreground">Memuat event dan audit...</div>
				{:else}
					<div class="grid gap-4 md:grid-cols-2">
						<div class="rounded-lg border p-4">
							<h3 class="mb-3 font-medium">Events</h3>
							{#if events.length === 0}<p class="text-sm text-muted-foreground">Belum ada event.</p>{/if}
							<ul class="space-y-2 text-sm">
								{#each events as item}
									<li class="rounded-md bg-muted/40 p-2"><div class="font-medium">{item.event_type || item.action || '-'}</div><div class="text-xs text-muted-foreground">{formatDate(item.created_at)} · {item.source || '-'}</div></li>
								{/each}
							</ul>
						</div>
						<div class="rounded-lg border p-4">
							<h3 class="mb-3 font-medium">Audit logs</h3>
							{#if auditLogs.length === 0}<p class="text-sm text-muted-foreground">Belum ada audit.</p>{/if}
							<ul class="space-y-2 text-sm">
								{#each auditLogs as item}
									<li class="rounded-md bg-muted/40 p-2"><div class="font-medium">{item.action || item.event_type || '-'}</div><div class="text-xs text-muted-foreground">{formatDate(item.created_at)} · {item.target || '-'}</div></li>
								{/each}
							</ul>
						</div>
					</div>
				{/if}
			</Card.Content>
		</Card.Root>
	{/if}
</div>

<style>
	@media print {
		:global(body.student-card-print-mode) { background: white !important; }
		:global(body.student-card-print-mode *) { visibility: hidden !important; }
		:global(body.student-card-print-mode .student-card-print-area),
		:global(body.student-card-print-mode .student-card-print-area *) { visibility: visible !important; }
		:global(body.student-card-print-mode .student-card-print-area) { position: fixed !important; left: 10mm !important; top: 10mm !important; }
	}
</style>
