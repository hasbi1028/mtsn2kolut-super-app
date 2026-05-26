<script lang="ts">
	import { browser } from '$app/environment';
	import { page } from '$app/stores';

	type ProctorCard = {
		card_id?: string;
		card_type?: string;
		event_title?: string;
		session_id?: string;
		session_title?: string;
		session_status?: string;
		scheduled_start?: string;
		scheduled_end?: string;
		room_id?: string;
		room_name?: string;
		proctor_name?: string;
		proctor_role?: string;
		package_title?: string;
	};

	let token = $state('');
	let pin = $state('');
	let loading = $state(false);
	let errorMessage = $state('');
	let card = $state<ProctorCard | null>(null);
	let demoRoomStatus = $state<'Menunggu' | 'Ujian Dibuka' | 'Bantuan Admin Diminta' | 'Ujian Ditutup'>('Menunggu');

	let queryCard = $derived($page.url.searchParams.get('card') ?? $page.url.searchParams.get('token') ?? '');
	let demoMode = $derived($page.url.searchParams.get('demo') === '1');
	let dashboardHref = $derived(card?.session_id && card?.room_id && !demoMode ? `/asesmen/sesi/${card.session_id}/rooms/${card.room_id}/proctoring?mode=simple` : '/asesmen/pengawasan');

	$effect(() => {
		if (queryCard && !token) token = queryCard;
		if (demoMode && !card) {
			card = {
				card_type: 'proctor',
				event_title: 'MODE DEMO Pengawasan CBT',
				session_id: 'demo-session',
				room_id: 'demo-room',
				session_title: 'Simulasi Pengawasan Portal Web',
				session_status: 'demo',
				scheduled_start: new Date().toISOString(),
				scheduled_end: new Date(Date.now() + 45 * 60_000).toISOString(),
				room_name: 'Ruang DEMO 01',
				proctor_name: 'Lembar Pengawas Ruang',
				proctor_role: 'pengawas_ruang',
				package_title: 'Informatika — Contoh lokal'
			};
		}
	});

	async function verifyCard() {
		loading = true;
		errorMessage = '';
		card = null;
		try {
			const response = await fetch('/api/exam/proctor/card/verify', {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ token: token.trim(), pin: pin.trim() })
			});
			const body = await response.json().catch(() => ({}));
			if (!response.ok) throw new Error(String(body.error ?? body.message ?? 'Lembar pengawas ruang tidak cocok atau PIN salah.'));
			card = (body.data?.card ?? body.card ?? {}) as ProctorCard;
			if (browser && card.session_id && card.room_id) {
				sessionStorage.setItem('cbt_proctor_card_room', JSON.stringify({ token: token.trim(), verified_at: new Date().toISOString(), card }));
			}
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'Lembar pengawas ruang tidak cocok atau PIN salah.';
		} finally {
			loading = false;
		}
	}

	function fmtDt(value: string | undefined) {
		if (!value) return '—';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return value;
		return date.toLocaleString('id-ID', { timeZone: 'Asia/Makassar', day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' }) + ' WITA';
	}
</script>

<svelte:head>
	<title>Portal Pengawasan Ruang — MTsN 2 Kolaka Utara</title>
</svelte:head>

<main class="min-h-screen bg-emerald-950 px-4 py-6 text-white">
	<section class="mx-auto max-w-2xl space-y-4">
		<header class="rounded-2xl border border-white/15 bg-white/10 p-5">
			<p class="text-xs font-semibold uppercase tracking-[0.3em] text-amber-200">MTsN 2 Kolaka Utara</p>
			<h1 class="mt-2 text-3xl font-bold">Portal Pengawasan Ruang</h1>
			<p class="mt-2 text-sm text-emerald-50/85">{demoMode ? 'MODE DEMO meniru alur pengawasan ruang ujian nyata tanpa API/database. Cocok untuk latihan pengawas sebelum Simulasi/Gladi.' : 'Scan QR pada Lembar Pengawas Ruang, masukkan PIN ruang, lalu buka ruang tugas. Lembar ini fleksibel: bila pengawas berhalangan, guru pengganti dapat memakai lembar ruang yang sama.'}</p>
		</header>

		{#if errorMessage}
			<div class="rounded-xl border border-red-200/30 bg-red-500/20 p-3 text-sm text-red-50">{errorMessage}</div>
		{/if}

		{#if card}
			<section class="rounded-2xl bg-white p-5 text-slate-950">
				<p class="text-xs font-semibold uppercase tracking-[0.2em] text-emerald-700">Lembar Pengawas Ruang Valid</p>
				<h2 class="mt-2 text-2xl font-bold">{card.room_name ?? 'Ruang Ujian'}</h2>
				<div class="mt-4 grid gap-2 rounded-xl bg-emerald-50 p-4 text-sm sm:grid-cols-2">
					<p><b>Akses:</b> {card.proctor_name ?? 'Lembar Pengawas Ruang'}</p>
					<p><b>Fleksibilitas:</b> Bisa dipakai guru pengganti yang ditunjuk</p>
					<p><b>Sesi:</b> {card.session_title ?? 'Sesi CBT'}</p>
					<p><b>Mapel/Paket:</b> {card.package_title ?? '—'}</p>
					<p><b>Jadwal:</b> {fmtDt(card.scheduled_start)}</p>
					<p><b>Status:</b> {card.session_status ?? '—'}</p>
				</div>
				{#if demoMode}
					<div class="mt-5 rounded-2xl border border-emerald-200 bg-emerald-50 p-4">
						<p class="text-xs font-semibold uppercase tracking-[0.2em] text-emerald-700">Simulasi tombol pengawas</p>
						<p class="mt-2 text-lg font-bold text-emerald-950">Status ruang: {demoRoomStatus}</p>
						<p class="mt-1 text-sm text-emerald-900">Tombol ini meniru Portal Pengawasan saat Simulasi/Gladi/Ujian nyata, tetapi tidak mengubah data server.</p>
						<div class="mt-4 grid gap-2 sm:grid-cols-3">
							<button class="rounded-xl bg-emerald-700 px-4 py-3 font-semibold text-white" onclick={() => (demoRoomStatus = 'Ujian Dibuka')}>Mulai Ujian</button>
							<button class="rounded-xl bg-amber-500 px-4 py-3 font-semibold text-amber-950" onclick={() => (demoRoomStatus = 'Bantuan Admin Diminta')}>Hubungi Admin</button>
							<button class="rounded-xl bg-slate-800 px-4 py-3 font-semibold text-white" onclick={() => (demoRoomStatus = 'Ujian Ditutup')}>Tutup Ujian</button>
						</div>
					</div>
				{:else}
					<a class="mt-5 block rounded-xl bg-emerald-700 px-4 py-4 text-center text-lg font-bold text-white" href={dashboardHref}>Buka Ruang Pengawasan</a>
				{/if}
				<p class="mt-3 text-center text-xs text-slate-500">{demoMode ? 'MODE DEMO: tidak ada perubahan data, tidak membuka/menutup sesi server, dan tidak mengirim bantuan ke admin.' : 'Jika diminta login admin/operator, gunakan akun pengawas/operator yang ditugaskan pada web-admin.'}</p>
			</section>
		{:else}
			<section class="rounded-2xl bg-white p-5 text-slate-950">
				<h2 class="text-xl font-bold">{demoMode ? 'MODE DEMO Portal Pengawasan' : 'Masuk dengan Lembar Pengawas Ruang'}</h2>
				<p class="mt-1 text-sm text-slate-600">{demoMode ? 'Demo langsung menampilkan ruang contoh agar pengawas melihat alur yang sama seperti Simulasi/Gladi/Ujian nyata.' : 'Kode QR biasanya terisi otomatis setelah scan. Jika tidak, ketik kode lembar ruang secara manual.'}</p>
				<div class="mt-4 grid gap-3 sm:grid-cols-2">
					<label class="space-y-1 text-sm font-medium">Kode Lembar / QR Token<input class="w-full rounded-lg border px-3 py-2" bind:value={token} autocomplete="off" /></label>
					<label class="space-y-1 text-sm font-medium">PIN<input class="w-full rounded-lg border px-3 py-2 text-center text-xl tracking-[0.4em]" bind:value={pin} inputmode="numeric" autocomplete="one-time-code" maxlength="8" placeholder="••••" /></label>
				</div>
				{#if !demoMode}
					<button class="mt-4 w-full rounded-xl bg-emerald-700 px-4 py-3 font-semibold text-white disabled:opacity-60" disabled={loading} onclick={verifyCard}>{loading ? 'Memeriksa...' : 'Masuk Portal Pengawasan'}</button>
				{:else}
					<button class="mt-4 w-full rounded-xl bg-emerald-700 px-4 py-3 font-semibold text-white" onclick={() => (card = { card_type: 'proctor', event_title: 'MODE DEMO Pengawasan CBT', session_id: 'demo-session', room_id: 'demo-room', session_title: 'Simulasi Pengawasan Portal Web', session_status: 'demo', scheduled_start: new Date().toISOString(), scheduled_end: new Date(Date.now() + 45 * 60_000).toISOString(), room_name: 'Ruang DEMO 01', proctor_name: 'Lembar Pengawas Ruang', proctor_role: 'pengawas_ruang', package_title: 'Informatika — Contoh lokal' })}>Mulai DEMO Pengawasan</button>
				{/if}
			</section>
		{/if}
	</section>
</main>
