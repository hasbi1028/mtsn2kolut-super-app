<script lang="ts">
	import { onMount } from 'svelte';
	import { readClientApiData } from '$lib/client/api';
	type Exam = {
		id: string;
		title: string;
		status: string;
		participant_count?: number;
		room_count?: number;
		card_count?: number;
	};
	type AssignmentClassSummary = {
		class_code: string;
		class_name?: string;
		grade_level?: number;
		count: number;
	};
	type AssignmentRoom = {
		code: string;
		name: string;
		capacity: number;
		assigned_count: number;
		grade_levels?: number[];
		class_summary?: AssignmentClassSummary[];
	};
	type AssignmentResult = {
		exam_id: string;
		session_id?: string;
		mix_policy: string;
		room_count: number;
		capacity_per_room: number;
		total_participants: number;
		assigned_total: number;
		unassigned_total: number;
		rooms: AssignmentRoom[];
		applied?: boolean;
		message?: string;
	};
	type ParticipantPlacement = {
		participant_id: string;
		room_id?: string;
		student_name: string;
		nis?: string;
		nisn?: string;
		class_code: string;
		class_name: string;
		grade_level: number;
		room_code?: string;
		room_name?: string;
		room_capacity?: number;
		seat_no?: number;
		status: string;
	};
	type ManualRoom = { id: string; code: string; name: string; capacity: number };

	let exams = $state<Exam[]>([]);
	let selectedExamId = $state('');
	let roomCount = $state(8);
	let capacityPerRoom = $state(30);
	let mixPolicy = $state<'mixed' | 'class_grouped'>('mixed');
	let loading = $state(true);
	let busy = $state(false);
	let participantBusy = $state(false);
	let error = $state('');
	let notice = $state('');
	let preview = $state<AssignmentResult | null>(null);
	let participants = $state<ParticipantPlacement[]>([]);
	let selectedParticipantId = $state('');
	let manualRoomId = $state('');
	let manualSeatNo = $state(1);

	let selectedExam = $derived(exams.find((exam) => exam.id === selectedExamId) ?? null);
	let rooms = $derived(preview?.rooms ?? []);
	let totalCapacity = $derived(roomCount * capacityPerRoom);
	let canPreview = $derived(Boolean(selectedExamId) && !busy);
	let canApply = $derived(Boolean(preview && preview.exam_id === selectedExamId && !busy));
	let manualRooms = $derived(buildManualRooms(participants));
	let selectedParticipant = $derived(
		participants.find((participant) => participant.participant_id === selectedParticipantId) ?? null
	);
	let selectedManualRoom = $derived(manualRooms.find((room) => room.id === manualRoomId) ?? null);
	let canMoveParticipant = $derived(Boolean(selectedExamId && selectedParticipantId && manualRoomId && manualSeatNo > 0 && !participantBusy));

	onMount(() => {
		void loadExams();
	});

	function buildManualRooms(items: ParticipantPlacement[]): ManualRoom[] {
		return Array.from(
			new Map(
				items
					.filter((item) => item.room_id)
					.map((item) => [
						item.room_id,
						{
							id: item.room_id ?? '',
							code: item.room_code ?? 'Ruang',
							name: item.room_name ?? item.room_code ?? 'Ruang',
							capacity: item.room_capacity ?? capacityPerRoom
						}
					])
			).values()
		).sort((a, b) => a.code.localeCompare(b.code));
	}

	async function loadExams() {
		loading = true;
		error = '';
		try {
			exams = await fetch('/api/asesmen/exams?limit=20').then((response) =>
				readClientApiData<Exam[]>(response, 'Daftar ujian belum dapat dimuat')
			);
			if (exams.length > 0 && !selectedExamId) selectedExamId = exams[0].id;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Daftar ujian belum dapat dimuat';
			exams = [];
		} finally {
			loading = false;
		}
	}

	async function runPreview() {
		if (!selectedExamId) return;
		busy = true;
		error = '';
		notice = '';
		try {
			preview = await fetch(`/api/asesmen/exams/${encodeURIComponent(selectedExamId)}/assignment-preview`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ room_count: roomCount, capacity_per_room: capacityPerRoom, mix_policy: mixPolicy })
			}).then((response) => readClientApiData<AssignmentResult>(response, 'Preview ruang belum dapat dimuat'));
			notice = 'Preview pembagian ruang siap. Periksa ringkasan sebelum Simpan Ruang.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Preview ruang gagal';
		} finally {
			busy = false;
		}
	}

	async function applyAssignment() {
		if (!selectedExamId || !preview) return;
		const ok = window.confirm('Simpan pembagian ruang ini? Aksi ini hanya menyimpan ruang/kursi dan tidak menerbitkan QR+PIN.');
		if (!ok) return;
		busy = true;
		error = '';
		notice = '';
		try {
			preview = await fetch(`/api/asesmen/exams/${encodeURIComponent(selectedExamId)}/assignment-apply`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ room_count: roomCount, capacity_per_room: capacityPerRoom, mix_policy: mixPolicy })
			}).then((response) => readClientApiData<AssignmentResult>(response, 'Pembagian ruang belum dapat disimpan'));
			notice = 'Pembagian ruang tersimpan. QR+PIN tetap belum diterbitkan.';
			await loadParticipants();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Simpan ruang gagal';
		} finally {
			busy = false;
		}
	}

	async function loadParticipants() {
		if (!selectedExamId) return;
		participantBusy = true;
		error = '';
		try {
			participants = await fetch(`/api/asesmen/exams/${encodeURIComponent(selectedExamId)}/participants`).then((response) =>
				readClientApiData<ParticipantPlacement[]>(response, 'Daftar peserta belum dapat dimuat')
			);
			const firstWithRoom = participants.find((item) => item.room_id) ?? participants[0];
			if (firstWithRoom) chooseParticipant(firstWithRoom);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Daftar peserta belum dapat dimuat';
			participants = [];
		} finally {
			participantBusy = false;
		}
	}

	function chooseParticipant(item: ParticipantPlacement) {
		selectedParticipantId = item.participant_id;
		manualRoomId = item.room_id ?? '';
		manualSeatNo = item.seat_no ?? 1;
	}

	async function saveManualSeat() {
		if (!canMoveParticipant || !selectedExamId) return;
		participantBusy = true;
		error = '';
		notice = '';
		try {
			await fetch(`/api/asesmen/exams/${encodeURIComponent(selectedExamId)}/participants/seat`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ participant_id: selectedParticipantId, room_id: manualRoomId, seat_no: Number(manualSeatNo) })
			}).then((response) => readClientApiData<unknown>(response, 'Pindah kursi gagal'));
			notice = 'Pindah ruang/kursi tersimpan untuk satu peserta.';
			await loadParticipants();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Pindah kursi gagal';
		} finally {
			participantBusy = false;
		}
	}

	function roomComposition(room: AssignmentRoom) {
		const summary = room.class_summary ?? [];
		if (summary.length === 0) return 'Belum ada ringkasan rombel';
		return summary.map((item) => `${item.class_code || item.class_name}: ${item.count}`).join(' · ');
	}

	function printPage() {
		window.print();
	}
</script>

<svelte:head>
	<title>Ruang Ujian Operasional</title>
	<meta name="description" content="Panel operasional ruang ujian CBT familiar: preview, simpan ruang, ringkasan, dan edit manual peserta." />
</svelte:head>

<div class="space-y-4 pb-10">
	<header class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm print:hidden">
		<p class="text-xs font-black uppercase tracking-[0.2em] text-emerald-700">Pusat Data · Operasional Penuh</p>
		<div class="mt-2 flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
			<div>
				<h1 class="text-3xl font-black tracking-tight md:text-4xl">Ruang Ujian</h1>
				<p class="mt-2 max-w-3xl text-sm leading-6 text-slate-600">
					Pilih ujian, preview pembagian R01–R08, simpan ruang/kursi, cek ringkasan, lalu pindahkan peserta secara manual bila diperlukan.
				</p>
			</div>
			<div class="flex flex-wrap gap-2">
				<a href="/asesmen/cbt/dokumen" class="rounded-2xl border border-slate-300 bg-white px-4 py-2 text-sm font-black text-slate-700 shadow-sm hover:border-emerald-300 hover:text-emerald-800">Dokumen & Cetak</a>
				<button class="rounded-2xl border border-slate-300 bg-white px-4 py-2 text-sm font-black text-slate-700 shadow-sm hover:border-emerald-300 hover:text-emerald-800" onclick={printPage}>Cetak ringkasan</button>
			</div>
		</div>
	</header>

	{#if error}
		<p class="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm font-bold text-red-800 print:hidden" role="alert">{error}</p>
	{/if}
	{#if notice}
		<p class="rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm font-bold text-emerald-900 print:hidden">{notice}</p>
	{/if}

	<section class="grid gap-4 lg:grid-cols-[0.85fr_1.15fr] print:hidden">
		<div class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
			<h2 class="text-lg font-black text-slate-950">1. Pilih ujian & kebijakan ruang</h2>
			<label class="mt-4 block space-y-1.5 text-sm font-bold text-slate-700">
				<span>Ujian</span>
				<select class="w-full rounded-2xl border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-emerald-500" bind:value={selectedExamId} disabled={loading}>
					<option value="">{loading ? 'Memuat ujian...' : 'Pilih ujian'}</option>
					{#each exams as exam (exam.id)}
						<option value={exam.id}>{exam.title} · {exam.status}</option>
					{/each}
				</select>
			</label>
			<div class="mt-3 grid gap-3 sm:grid-cols-2">
				<label class="space-y-1.5 text-sm font-bold text-slate-700">
					<span>Jumlah ruang</span>
					<input type="number" min="1" max="20" bind:value={roomCount} class="w-full rounded-2xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-emerald-500" />
				</label>
				<label class="space-y-1.5 text-sm font-bold text-slate-700">
					<span>Kapasitas/ruang</span>
					<input type="number" min="1" max="60" bind:value={capacityPerRoom} class="w-full rounded-2xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-emerald-500" />
				</label>
			</div>
			<label class="mt-3 block space-y-1.5 text-sm font-bold text-slate-700">
				<span>Kebijakan campur</span>
				<select class="w-full rounded-2xl border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-emerald-500" bind:value={mixPolicy}>
					<option value="mixed">Campur rombel/kelas</option>
					<option value="class_grouped">Kelompok per kelas</option>
				</select>
			</label>
			<div class="mt-4 rounded-2xl bg-slate-50 p-3 text-sm font-bold text-slate-600">
				Total kapasitas: <span class="text-slate-950">{totalCapacity}</span>
				{#if selectedExam}
					· Peserta terdaftar: <span class="text-slate-950">{selectedExam.participant_count ?? '—'}</span>
				{/if}
			</div>
			<div class="mt-4 flex flex-wrap gap-2">
				<button class="rounded-2xl bg-emerald-700 px-4 py-2 text-sm font-black text-white hover:bg-emerald-800 disabled:cursor-not-allowed disabled:bg-slate-300" disabled={!canPreview} onclick={runPreview}>{busy ? 'Memproses...' : 'Preview Ruang'}</button>
				<button class="rounded-2xl border border-amber-300 bg-amber-50 px-4 py-2 text-sm font-black text-amber-900 hover:bg-amber-100 disabled:cursor-not-allowed disabled:opacity-50" disabled={!canApply} onclick={applyAssignment}>Simpan Ruang</button>
				<button class="rounded-2xl border border-slate-300 px-4 py-2 text-sm font-black text-slate-700 hover:border-emerald-300 hover:text-emerald-800 disabled:cursor-not-allowed disabled:opacity-50" disabled={!selectedExamId || participantBusy} onclick={loadParticipants}>Muat Peserta</button>
			</div>
			<p class="mt-3 text-xs font-bold leading-5 text-slate-500">Simpan Ruang tidak menerbitkan kartu, QR, atau PIN. Kredensial tetap dari Dokumen & Cetak.</p>
		</div>

		<div class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
			<h2 class="text-lg font-black text-slate-950">2. Ringkasan hasil penempatan</h2>
			<p class="text-sm font-semibold text-slate-500">Gunakan panel ini untuk cek cepat sebelum cetak atau pindah manual.</p>
			{#if !preview}
				<p class="mt-4 rounded-2xl border border-dashed border-slate-300 bg-slate-50 p-4 text-sm text-slate-600">Belum ada preview. Pilih ujian lalu klik Preview Ruang.</p>
			{:else}
				<div class="mt-4 grid gap-3 sm:grid-cols-3">
					<div class="rounded-2xl bg-emerald-50 p-3"><p class="text-xs font-black uppercase text-emerald-800">Terbagi</p><p class="text-2xl font-black text-emerald-950">{preview.assigned_total}</p></div>
					<div class="rounded-2xl bg-slate-50 p-3"><p class="text-xs font-black uppercase text-slate-500">Belum tertampung</p><p class="text-2xl font-black text-slate-950">{preview.unassigned_total}</p></div>
					<div class="rounded-2xl bg-slate-50 p-3"><p class="text-xs font-black uppercase text-slate-500">Mode</p><p class="text-2xl font-black text-slate-950">{preview.mix_policy}</p></div>
				</div>
				<div class="mt-4 grid gap-2 md:grid-cols-2 xl:grid-cols-4">
					{#each rooms as room (room.code)}
						<div class="rounded-2xl border border-slate-200 p-3">
							<div class="flex items-center justify-between gap-2"><p class="font-black text-slate-950">{room.code}</p><span class="rounded-full bg-slate-100 px-2 py-0.5 text-xs font-black text-slate-600">{room.assigned_count}/{room.capacity}</span></div>
							<p class="mt-1 text-xs font-bold leading-5 text-slate-500">{roomComposition(room)}</p>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	</section>

	<section class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm print:hidden">
		<div class="flex flex-col gap-3 border-b border-slate-200 pb-3 lg:flex-row lg:items-end lg:justify-between">
			<div>
				<h2 class="text-lg font-black text-slate-950">3. Edit manual peserta</h2>
				<p class="text-sm font-semibold text-slate-500">Pindahkan satu peserta tanpa mengulang pembagian otomatis. Backend akan menolak kursi ganda/di luar kapasitas.</p>
			</div>
			<span class="rounded-full bg-emerald-50 px-3 py-1 text-xs font-black text-emerald-800">{participants.length} peserta dimuat</span>
		</div>
		{#if participants.length === 0}
			<p class="mt-4 rounded-2xl border border-dashed border-slate-300 bg-slate-50 p-4 text-sm text-slate-600">Klik Muat Peserta setelah Simpan Ruang untuk membuka edit manual.</p>
		{:else}
			<div class="mt-4 grid gap-4 lg:grid-cols-[1fr_0.85fr]">
				<div class="max-h-96 overflow-y-auto rounded-2xl border border-slate-200">
					{#each participants as item (item.participant_id)}
						<button class={`grid w-full gap-2 border-b border-slate-100 p-3 text-left text-sm hover:bg-emerald-50 sm:grid-cols-[1fr_auto] ${selectedParticipantId === item.participant_id ? 'bg-emerald-50' : 'bg-white'}`} onclick={() => chooseParticipant(item)}>
							<span class="min-w-0"><span class="block truncate font-black text-slate-950">{item.student_name}</span><span class="text-xs font-bold text-slate-500">{item.class_code || item.class_name} · {item.nis || item.nisn || 'tanpa NIS'}</span></span>
							<span class="w-fit rounded-xl bg-slate-100 px-2 py-1 text-xs font-black text-slate-700">{item.room_code || '—'} / {item.seat_no || '-'}</span>
						</button>
					{/each}
				</div>
				<div class="rounded-2xl border border-slate-200 bg-slate-50 p-4">
					<p class="text-sm font-black text-slate-950">Form pindah ruang/kursi</p>
					{#if selectedParticipant}<p class="mt-1 text-xs font-bold text-slate-600">{selectedParticipant.student_name} · {selectedParticipant.class_code || selectedParticipant.class_name}</p>{/if}
					<label class="mt-3 block space-y-1.5 text-sm font-bold text-slate-700"><span>Ruang tujuan</span><select class="w-full rounded-2xl border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-emerald-500" bind:value={manualRoomId}><option value="">Pilih ruang</option>{#each manualRooms as room (room.id)}<option value={room.id}>{room.code} · {room.name}</option>{/each}</select></label>
					<label class="mt-3 block space-y-1.5 text-sm font-bold text-slate-700"><span>Nomor kursi {selectedManualRoom ? `(maks. ${selectedManualRoom.capacity})` : ''}</span><input type="number" min="1" max={selectedManualRoom?.capacity || 60} class="w-full rounded-2xl border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-emerald-500" bind:value={manualSeatNo} /></label>
					<button class="mt-4 rounded-2xl bg-emerald-700 px-4 py-2 text-sm font-black text-white hover:bg-emerald-800 disabled:cursor-not-allowed disabled:bg-slate-300" disabled={!canMoveParticipant} onclick={saveManualSeat}>{participantBusy ? 'Menyimpan...' : 'Simpan Pindahan'}</button>
				</div>
			</div>
		{/if}
	</section>

	<section class="hidden print:block">
		<h1 class="text-xl font-black">Ringkasan Ruang Ujian</h1>
		<p class="text-sm">{selectedExam?.title ?? 'Ujian belum dipilih'}</p>
		<div class="mt-4 grid grid-cols-2 gap-2">
			{#each rooms as room (room.code)}
				<div class="border border-slate-300 p-2 text-sm"><strong>{room.code}</strong> · {room.assigned_count}/{room.capacity}<br />{roomComposition(room)}</div>
			{/each}
		</div>
	</section>
</div>
