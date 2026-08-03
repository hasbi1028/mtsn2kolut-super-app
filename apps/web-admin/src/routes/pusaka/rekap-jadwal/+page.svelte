<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { toast } from '$lib/components/ui/sonner';
	import { Input } from '$lib/components/ui/input';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { readClientApiData } from '$lib/client/api';

	type PusakaSchedule = {
		id: string; label: string; run_type: string;
		run_time: string; is_enabled: boolean; send_telegram_after?: boolean;
		last_enqueued_for_date?: string | null;
		created_at?: string; updated_at?: string;
	};

	const runTypeLabel: Record<string, string> = {
		morning: 'Rekap Pagi',
		afternoon: 'Rekap Sore',
		checkin: 'Absen Masuk',
		checkout: 'Absen Pulang'
	};

	const runTypeOptions = [
		{ value: 'morning', label: 'Rekap Pagi (rekap absen)' },
		{ value: 'afternoon', label: 'Rekap Sore (rekap absen)' },
	];

	let schedules = $state<PusakaSchedule[]>([]);
	let schedulesPromise = $state<Promise<PusakaSchedule[]> | null>(null);
	let saving = $state(false);
	let scheduleRequestId = 0;

	// ── Rekap & Kirim Laporan (1 tombol, sama seperti di Monitor Kehadiran) ──
	type JobStats = { queued: number; running: number; success: number; failed: number };
	let rekapKirimPhase = $state<'idle' | 'rekaping' | 'mengirim' | 'done' | 'error'>('idle');
	let rekapKirimProgress = $state({ done: 0, total: 0 });
	const rekapKirimBusy = $derived(rekapKirimPhase === 'rekaping' || rekapKirimPhase === 'mengirim');
	const rekapKirimPct = $derived(rekapKirimProgress.total > 0 ? Math.round((rekapKirimProgress.done / rekapKirimProgress.total) * 100) : 0);

	function sleep(ms: number) { return new Promise((r) => setTimeout(r, ms)); }

	function todayWita() {
		return new Intl.DateTimeFormat('en-CA', { timeZone: 'Asia/Makassar', year: 'numeric', month: '2-digit', day: '2-digit' }).format(new Date());
	}

	async function fetchJobStats(): Promise<JobStats> {
		const res = await fetch('/api/pusaka/jobs/stats');
		return readClientApiData<JobStats>(res, 'Gagal memuat status antrian');
	}

	async function runRekapDanKirim() {
		if (rekapKirimBusy) return;
		rekapKirimPhase = 'rekaping';
		rekapKirimProgress = { done: 0, total: 0 };
		try {
			const before = await fetchJobStats();
			const runRes = await fetch('/api/pusaka/jobs/run-all', {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ run_type: 'morning' })
			});
			const runData = await readClientApiData<{ inserted?: number; skipped?: number }>(runRes, 'Gagal menjalankan rekap massal');
			const inserted = runData.inserted ?? 0;

			if (inserted > 0) {
				rekapKirimProgress = { done: 0, total: inserted };
				const deadline = Date.now() + 5 * 60 * 1000;
				while (Date.now() < deadline) {
					await sleep(5000);
					const stats = await fetchJobStats();
					const done = Math.max(0, (stats.success + stats.failed) - (before.success + before.failed));
					rekapKirimProgress = { done: Math.min(done, inserted), total: inserted };
					if (stats.queued === 0 && stats.running === 0) break;
					if (done >= inserted) break;
				}
			} else {
				rekapKirimProgress = { done: 1, total: 1 };
			}

			rekapKirimPhase = 'mengirim';
			const sendRes = await fetch('/api/pusaka/attendance-telegram/send', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ date: todayWita(), include_caption: true, include_image: true })
			});
			const sendData = await readClientApiData<{ target_chat_id_masked?: string }>(sendRes, 'Gagal mengirim laporan Telegram');
			const target = sendData?.target_chat_id_masked ?? 'Telegram';
			rekapKirimPhase = 'done';
			toast.success(`Rekap selesai (${rekapKirimProgress.done}/${rekapKirimProgress.total}) — laporan terkirim ke ${target}`);
		} catch (e) {
			rekapKirimPhase = 'error';
			toast.error(e instanceof Error ? e.message : 'Rekap & kirim gagal');
		}
	}

	// Form jadwal baru
	let newLabel = $state('');
	let newTime = $state('23:00');
	let newType = $state('morning');
	let showForm = $state(false);

	function fmtDt(iso: string | null | undefined) {
		if (!iso) return '—';
		const d = new Date(iso);
		return d.toLocaleString('id-ID', {
			timeZone: 'Asia/Makassar',
			day: 'numeric', month: 'short',
			hour: '2-digit', minute: '2-digit'
		});
	}

	async function fetchSchedules(): Promise<PusakaSchedule[]> {
		const res = await fetch('/api/pusaka/schedules');
		return readClientApiData<PusakaSchedule[]>(res, 'Gagal memuat jadwal');
	}

	function loadSchedules() {
		const requestId = ++scheduleRequestId;
		schedulesPromise = fetchSchedules().then((data) => {
			if (requestId === scheduleRequestId) {
				schedules = Array.isArray(data) ? data : [];
				return schedules;
			}
			return schedules;
		}).catch((error: unknown) => {
			if (requestId === scheduleRequestId) throw error;
			return schedules;
		});
		return schedulesPromise;
	}

	async function refreshSchedules() {
		try {
			const data = await fetchSchedules();
			schedules = Array.isArray(data) ? data : [];
			schedulesPromise = Promise.resolve(schedules);
		} catch (e) {
			schedulesPromise = Promise.resolve(schedules);
			toast.error(e instanceof Error ? e.message : 'Gagal refresh');
		}
	}

	function retry(ref?: () => void) { ref?.(); loadSchedules(); }

	function errorMsg(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat jadwal. Coba lagi.';
	}

	function handleError(error: unknown, reset: () => void) {
		console.error('Jadwal gagal dimuat', error);
		reset();
	}

	async function createSchedule() {
		if (!newLabel.trim()) { toast.error('Nama jadwal harus diisi'); return; }
		if (!newTime) { toast.error('Waktu harus diisi'); return; }
		saving = true;
		try {
			const res = await fetch('/api/pusaka/schedules', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					label: newLabel.trim(),
					run_time: newTime,
					run_type: newType,
					is_enabled: true,
					send_telegram_after: false
				})
			});
			if (!res.ok) { const err = await res.json(); throw new Error(err.error || 'Gagal membuat'); }
			toast.success(`Jadwal "${newLabel.trim()}" ditambahkan (jam ${newTime} WITA)`);
			newLabel = '';
			newTime = '23:00';
			showForm = false;
			await refreshSchedules();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Gagal membuat jadwal');
		} finally { saving = false; }
	}

	async function toggleSchedule(sched: PusakaSchedule) {
		sched.is_enabled = !sched.is_enabled;
		try {
			const res = await fetch(`/api/pusaka/schedules/${sched.id}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ label: sched.label, run_time: sched.run_time, is_enabled: sched.is_enabled, send_telegram_after: sched.send_telegram_after ?? false })
			});
			if (!res.ok) { const err = await res.json(); throw new Error(err.error || 'Gagal'); }
		} catch (e) {
			sched.is_enabled = !sched.is_enabled;
			toast.error(e instanceof Error ? e.message : 'Gagal mengubah jadwal');
		}
	}

	async function updateSchedule(sched: PusakaSchedule) {
		try {
			const res = await fetch(`/api/pusaka/schedules/${sched.id}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ label: sched.label, run_time: sched.run_time, is_enabled: sched.is_enabled, send_telegram_after: sched.send_telegram_after ?? false })
			});
			if (!res.ok) { const err = await res.json(); throw new Error(err.error || 'Gagal'); }
			toast.success(`Jadwal "${sched.label}" diubah ke jam ${sched.run_time} WITA`);
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Gagal menyimpan jadwal');
			await refreshSchedules();
		}
	}

	async function toggleTelegramAfter(sched: PusakaSchedule) {
		sched.send_telegram_after = !(sched.send_telegram_after ?? false);
		try {
			const res = await fetch(`/api/pusaka/schedules/${sched.id}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ label: sched.label, run_time: sched.run_time, is_enabled: sched.is_enabled, send_telegram_after: sched.send_telegram_after })
			});
			if (!res.ok) { const err = await res.json(); throw new Error(err.error || 'Gagal'); }
			toast.success(sched.send_telegram_after
				? `Jadwal "${sched.label}" akan kirim Telegram otomatis setelah rekap selesai`
				: `Jadwal "${sched.label}" tidak kirim Telegram otomatis`);
		} catch (e) {
			sched.send_telegram_after = !(sched.send_telegram_after ?? false);
			toast.error(e instanceof Error ? e.message : 'Gagal mengubah pengaturan Telegram');
		}
	}

	async function deleteSchedule(sched: PusakaSchedule) {
		if (!confirm(`Hapus jadwal "${sched.label}"?`)) return;
		try {
			const res = await fetch(`/api/pusaka/schedules/${sched.id}`, { method: 'DELETE' });
			if (!res.ok) { const err = await res.json(); throw new Error(err.error || 'Gagal'); }
			toast.success(`Jadwal "${sched.label}" dihapus`);
			await refreshSchedules();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Gagal menghapus jadwal');
		}
	}

	const sortedSchedules = $derived(
		[...schedules].sort((a, b) => a.run_time.localeCompare(b.run_time))
	);

	onMount(() => { loadSchedules(); });
</script>

<svelte:head><title>Jadwal Otomatis Rekap — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-4">

	<!-- Header -->
	<div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-lg font-black tracking-tight text-base-content lg:text-2xl">⏰ Jadwal Otomatis Rekap</h1>
			<div class="flex items-center gap-2 text-xs text-base-content/70">
				<a href={resolve('/pusaka')} class="hover:text-base-content">PUSAKA</a>
				<span>/</span>
				<span class="font-medium text-base-content">Jadwal Otomatis</span>
			</div>
		</div>
		<div class="flex flex-wrap items-center gap-2">
			<LoadingButton size="sm" variant="outline" onclick={() => void runRekapDanKirim()} loading={rekapKirimBusy} loadingLabel="Memproses..." label="" class="shrink-0 text-primary border-primary/50 hover:bg-primary/10" disabled={rekapKirimBusy}>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" /></svg>
				{rekapKirimPhase === 'done' ? 'Selesai ✓' : rekapKirimPhase === 'error' ? 'Coba Lagi' : 'Rekap & Kirim Laporan'}
			</LoadingButton>
			<Button size="sm" onclick={() => (showForm = !showForm)} class="shrink-0 gap-1">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
				{showForm ? 'Batal' : 'Tambah Jadwal'}
			</Button>
		</div>
	</div>

	<!-- Progress rekap & kirim -->
	{#if rekapKirimBusy}
		<div class="rounded-xl border border-primary/30 bg-primary/5 p-3">
			<div class="flex items-center justify-between text-xs">
				<span class="font-semibold text-base-content">
					{#if rekapKirimPhase === 'rekaping'}
						Merekap absen semua pegawai... {rekapKirimProgress.done}/{rekapKirimProgress.total} selesai
					{:else}
						Mengirim laporan ke Telegram...
					{/if}
				</span>
				{#if rekapKirimPhase === 'rekaping'}
					<span class="font-bold text-primary">{rekapKirimPct}%</span>
				{/if}
			</div>
			{#if rekapKirimPhase === 'rekaping'}
				<div class="mt-2 h-2 w-full overflow-hidden rounded-full bg-base-300">
					<div class="h-full rounded-full bg-primary transition-all duration-500" style="width: {rekapKirimPct}%"></div>
				</div>
			{/if}
		</div>
	{/if}

	<!-- Form tambah -->
	{#if showForm}
		<Card.Root class="border-primary/30 shadow-sm">
			<Card.Content class="p-4 space-y-3">
				<h3 class="text-sm font-bold">Tambah Jadwal Baru</h3>
				<div class="grid gap-3 sm:grid-cols-4">
					<div class="sm:col-span-2">
						<label class="block text-xs font-medium text-base-content/70 mb-1">Nama Jadwal</label>
						<Input type="text" bind:value={newLabel} placeholder="Contoh: Rekap Harian" />
					</div>
					<div>
						<label class="block text-xs font-medium text-base-content/70 mb-1">Jam (WITA)</label>
						<Input type="time" bind:value={newTime} />
					</div>
					<div>
						<label class="block text-xs font-medium text-base-content/70 mb-1">Tipe</label>
						<select bind:value={newType}
							class="h-10 w-full rounded-lg border border-input bg-background px-3 text-sm">
							{#each runTypeOptions as opt}
								<option value={opt.value}>{opt.label}</option>
							{/each}
						</select>
					</div>
				</div>
				<div class="flex justify-end gap-2 pt-1">
					<Button size="sm" variant="ghost" onclick={() => (showForm = false)}>Batal</Button>
					<LoadingButton size="sm" onclick={() => void createSchedule()} loading={saving} loadingLabel="Menyimpan..." label="Simpan" />
				</div>
			</Card.Content>
		</Card.Root>
	{/if}

	<!-- Info -->
	<p class="text-xs text-base-content/60 leading-relaxed">
		Scheduler berjalan otomatis setiap 30 detik. Jadwal yang aktif (<Badge variant="default" class="text-[10px]">AKTIF</Badge>) akan dijalankan sesuai jam yang ditentukan.
		Rekap akan memproses absensi <strong>seluruh pegawai PUSAKA</strong>.
	</p>

	<!-- Tabel jadwal -->
	{#await schedulesPromise}
		<div class="space-y-3 rounded-xl border border-base-300 bg-base-100 p-4">
			{#each Array(5) as _, i}
				<Skeleton class="h-12 w-full" />
			{/each}
		</div>
	{:then _}
		<Card.Root class="overflow-hidden border-base-300 shadow-sm">
			<div class="hidden overflow-x-auto lg:block">
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>Nama Jadwal</Table.Head>
							<Table.Head>Jam</Table.Head>
							<Table.Head>Tipe</Table.Head>
							<Table.Head>Status</Table.Head>
							<Table.Head>Kirim Telegram</Table.Head>
							<Table.Head>Terakhir Jalan</Table.Head>
							<Table.Head class="text-right">Aksi</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each sortedSchedules as sched (sched.id)}
							<Table.Row>
								<Table.Cell class="font-medium">{sched.label}</Table.Cell>
								<Table.Cell>
									<input type="time" bind:value={sched.run_time}
										onchange={() => void updateSchedule(sched)}
										class="h-8 rounded-lg border border-input bg-background px-2 text-xs w-28" />
								</Table.Cell>
								<Table.Cell>
									<Badge variant="outline">{runTypeLabel[sched.run_type] || sched.run_type}</Badge>
								</Table.Cell>
								<Table.Cell>
									<button type="button" role="switch" aria-checked={sched.is_enabled}
										onclick={() => void toggleSchedule(sched)}
										class="relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors {sched.is_enabled ? 'bg-primary' : 'bg-base-300'}">
										<span class="pointer-events-none inline-block h-4 w-4 rounded-full bg-white shadow transform ring-0 transition {sched.is_enabled ? 'translate-x-4' : 'translate-x-0'}" />
									</button>
								</Table.Cell>
								<Table.Cell>
									<button type="button" role="switch" aria-checked={sched.send_telegram_after ?? false}
										onclick={() => void toggleTelegramAfter(sched)}
										title={sched.send_telegram_after ? 'Kirim Telegram otomatis setelah rekap selesai' : 'Aktifkan kirim Telegram otomatis setelah rekap selesai'}
										class="relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors {sched.send_telegram_after ? 'bg-emerald-500' : 'bg-base-300'}">
										<span class="pointer-events-none inline-block h-4 w-4 rounded-full bg-white shadow transform ring-0 transition {sched.send_telegram_after ? 'translate-x-4' : 'translate-x-0'}" />
									</button>
								</Table.Cell>
								<Table.Cell class="text-xs text-base-content/60">{fmtDt(sched.last_enqueued_for_date)}</Table.Cell>
								<Table.Cell class="text-right">
									<button type="button" onclick={() => void deleteSchedule(sched)}
										class="text-xs text-destructive hover:underline">Hapus</button>
								</Table.Cell>
							</Table.Row>
						{:else}
							<Table.Row>
								<Table.Cell colspan={7} class="p-6">
									<EmptyStatePanel compact title="Belum ada jadwal" description="Klik 'Tambah Jadwal' untuk membuat jadwal otomatis rekap." />
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			</div>
			<!-- Mobile list -->
			<div class="lg:hidden">
				{#if sortedSchedules.length === 0}
					<div class="p-4"><EmptyStatePanel compact title="Belum ada jadwal" description="Klik 'Tambah Jadwal' untuk membuat jadwal otomatis rekap." /></div>
				{:else}
					<ul class="divide-y divide-border">
						{#each sortedSchedules as sched (sched.id)}
							<li class="flex flex-col gap-2 px-4 py-3">
								<div class="flex items-center justify-between">
									<div>
										<p class="text-sm font-semibold">{sched.label}</p>
										<p class="text-[11px] text-base-content/60">{runTypeLabel[sched.run_type] || sched.run_type}</p>
									</div>
									<button type="button" role="switch" aria-checked={sched.is_enabled}
										onclick={() => void toggleSchedule(sched)}
										class="relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors {sched.is_enabled ? 'bg-primary' : 'bg-base-300'}">
										<span class="pointer-events-none inline-block h-4 w-4 rounded-full bg-white shadow transform ring-0 transition {sched.is_enabled ? 'translate-x-4' : 'translate-x-0'}" />
									</button>
								</div>
								<div class="flex items-center gap-3">
									<input type="time" bind:value={sched.run_time}
										onchange={() => void updateSchedule(sched)}
										class="h-8 rounded-lg border border-input bg-background px-2 text-xs w-28" />
									<span class="text-[10px] text-base-content/50">Terakhir: {fmtDt(sched.last_enqueued_for_date)}</span>
									<button type="button" onclick={() => void deleteSchedule(sched)}
										class="ml-auto text-xs text-destructive hover:underline">Hapus</button>
								</div>
								<div class="flex items-center gap-2">
									<button type="button" role="switch" aria-checked={sched.send_telegram_after ?? false}
										onclick={() => void toggleTelegramAfter(sched)}
										class="relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors {sched.send_telegram_after ? 'bg-emerald-500' : 'bg-base-300'}">
										<span class="pointer-events-none inline-block h-4 w-4 rounded-full bg-white shadow transform ring-0 transition {sched.send_telegram_after ? 'translate-x-4' : 'translate-x-0'}" />
									</button>
									<span class="text-[10px] text-base-content/60">Kirim Telegram otomatis setelah rekap selesai</span>
								</div>
							</li>
						{/each}
					</ul>
				{/if}
			</div>
		</Card.Root>
	{:catch error}
		<RecoveryPanel title="Gagal Memuat Jadwal" message={errorMsg(error)} onRetry={() => retry()} />
	{/await}
</div>
