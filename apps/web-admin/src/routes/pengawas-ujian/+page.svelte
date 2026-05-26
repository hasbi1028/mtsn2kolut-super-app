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

	let queryCard = $derived($page.url.searchParams.get('card') ?? $page.url.searchParams.get('token') ?? '');
	let dashboardHref = $derived(card?.session_id && card?.room_id ? `/asesmen/sesi/${card.session_id}/rooms/${card.room_id}/proctoring?mode=simple` : '/asesmen/pengawasan');

	$effect(() => {
		if (queryCard && !token) token = queryCard;
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
			if (!response.ok) throw new Error(String(body.error ?? body.message ?? 'Kartu pengawas tidak cocok atau PIN salah.'));
			card = (body.data?.card ?? body.card ?? {}) as ProctorCard;
			if (browser && card.session_id && card.room_id) {
				sessionStorage.setItem('cbt_proctor_card_room', JSON.stringify({ token: token.trim(), verified_at: new Date().toISOString(), card }));
			}
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'Kartu pengawas tidak cocok atau PIN salah.';
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
	<title>Portal Pengawas Ujian — MTsN 2 Kolaka Utara</title>
</svelte:head>

<main class="min-h-screen bg-emerald-950 px-4 py-6 text-white">
	<section class="mx-auto max-w-2xl space-y-4">
		<header class="rounded-2xl border border-white/15 bg-white/10 p-5">
			<p class="text-xs font-semibold uppercase tracking-[0.3em] text-amber-200">MTsN 2 Kolaka Utara</p>
			<h1 class="mt-2 text-3xl font-bold">Portal Pengawasan</h1>
			<p class="mt-2 text-sm text-emerald-50/85">Scan QR pada Kartu Pengawas, masukkan PIN, lalu buka ruang tugas. Untuk pengawas, gunakan tombol besar di halaman ruang.</p>
		</header>

		{#if errorMessage}
			<div class="rounded-xl border border-red-200/30 bg-red-500/20 p-3 text-sm text-red-50">{errorMessage}</div>
		{/if}

		{#if card}
			<section class="rounded-2xl bg-white p-5 text-slate-950">
				<p class="text-xs font-semibold uppercase tracking-[0.2em] text-emerald-700">Kartu Pengawas Valid</p>
				<h2 class="mt-2 text-2xl font-bold">{card.room_name ?? 'Ruang Ujian'}</h2>
				<div class="mt-4 grid gap-2 rounded-xl bg-emerald-50 p-4 text-sm sm:grid-cols-2">
					<p><b>Pengawas:</b> {card.proctor_name ?? 'Pengawas Ruang'}</p>
					<p><b>Tugas:</b> {card.proctor_role ?? 'utama'}</p>
					<p><b>Sesi:</b> {card.session_title ?? 'Sesi CBT'}</p>
					<p><b>Mapel/Paket:</b> {card.package_title ?? '—'}</p>
					<p><b>Jadwal:</b> {fmtDt(card.scheduled_start)}</p>
					<p><b>Status:</b> {card.session_status ?? '—'}</p>
				</div>
				<a class="mt-5 block rounded-xl bg-emerald-700 px-4 py-4 text-center text-lg font-bold text-white" href={dashboardHref}>Buka Ruang Pengawasan</a>
				<p class="mt-3 text-center text-xs text-slate-500">Jika diminta login admin/operator, gunakan akun pengawas/operator yang ditugaskan pada web-admin.</p>
			</section>
		{:else}
			<section class="rounded-2xl bg-white p-5 text-slate-950">
				<h2 class="text-xl font-bold">Masuk dengan Kartu Pengawas</h2>
				<p class="mt-1 text-sm text-slate-600">Kode QR biasanya terisi otomatis setelah scan. Jika tidak, ketik kode kartu secara manual.</p>
				<div class="mt-4 grid gap-3 sm:grid-cols-2">
					<label class="space-y-1 text-sm font-medium">Kode Kartu / QR Token<input class="w-full rounded-lg border px-3 py-2" bind:value={token} autocomplete="off" /></label>
					<label class="space-y-1 text-sm font-medium">PIN<input class="w-full rounded-lg border px-3 py-2 text-center text-xl tracking-[0.4em]" bind:value={pin} inputmode="numeric" autocomplete="one-time-code" maxlength="8" placeholder="••••" /></label>
				</div>
				<button class="mt-4 w-full rounded-xl bg-emerald-700 px-4 py-3 font-semibold text-white disabled:opacity-60" disabled={loading} onclick={verifyCard}>{loading ? 'Memeriksa...' : 'Masuk Portal Pengawasan'}</button>
			</section>
		{/if}
	</section>
</main>
