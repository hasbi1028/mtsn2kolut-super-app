<script lang="ts">
	import { onMount } from 'svelte';
	import StudentIdCardTemplate from '$lib/components/student-id-card/StudentIdCardTemplate.svelte';
	import { readClientApiData, readClientJson } from '$lib/client/api';

	let { params }: { params: { token: string } } = $props();

	type VerifyResult = {
		valid: boolean;
		message: string;
		card?: { id?: string; card_no?: string; status?: string };
		student?: { id?: string; name?: string; nis?: string; class_name?: string; photo_url?: string };
	};

	let loading = $state(true);
	let result = $state<VerifyResult | null>(null);
	let error = $state('');
	let pin = $state('');
	let challengeId = $state('');
	let loginMessage = $state('');
	let busy = $state('');

	async function loadVerify() {
		loading = true;
		error = '';
		try {
			result = await readClientApiData<VerifyResult>(await fetch(`/api/public/student-cards/verify/${encodeURIComponent(params.token)}`), 'Kartu tidak ditemukan');
		} catch (e) {
			error = (e as Error).message || 'Kartu tidak valid';
		} finally {
			loading = false;
		}
	}

	async function startLogin() {
		busy = 'start';
		loginMessage = '';
		try {
			const payload = await readClientApiData<{ challenge_id: string; student_hint?: { name?: string; class_name?: string } }>(await fetch('/api/public/student-cards/portal-login/start', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ qr_token: params.token })
			}), 'Gagal memulai login QR');
			challengeId = payload.challenge_id;
			loginMessage = 'QR valid. Masukkan PIN siswa untuk melanjutkan.';
		} catch (e) {
			loginMessage = (e as Error).message || 'Gagal memulai login QR';
		} finally {
			busy = '';
		}
	}

	async function completeLogin() {
		if (!challengeId || !pin.trim()) {
			loginMessage = 'Challenge dan PIN wajib diisi.';
			return;
		}
		busy = 'complete';
		try {
			const payload = await readClientJson<{ data?: { message?: string; student_id?: string }; error?: string }>(await fetch('/api/public/student-cards/portal-login/complete', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ challenge_id: challengeId, pin })
			}));
			loginMessage = payload.data?.message || 'PIN valid. Integrasi session portal siswa siap disambungkan ke auth portal.';
		} catch (e) {
			loginMessage = (e as Error).message || 'PIN tidak valid';
		} finally {
			busy = '';
		}
	}

	onMount(loadVerify);
</script>

<svelte:head><title>Verifikasi Kartu Siswa</title></svelte:head>

<div class="min-h-screen bg-[#f6f0dc] px-4 py-8 text-[#132217]">
	<main class="mx-auto max-w-4xl space-y-6">
		<section class="rounded-3xl bg-white/85 p-6 shadow-xl ring-1 ring-emerald-900/10">
			<p class="text-sm font-semibold uppercase tracking-[0.2em] text-emerald-800">MTsN 2 Kolaka Utara</p>
			<h1 class="mt-2 text-3xl font-bold">Verifikasi ID Card Siswa Terpadu</h1>
			<p class="mt-2 text-sm text-slate-600">Halaman ini hanya menampilkan informasi terbatas. QR bukan password dan tidak membuka akun tanpa PIN/token layanan.</p>
		</section>

		{#if loading}
			<div class="rounded-2xl bg-white p-6 text-sm text-slate-600 shadow">Memeriksa kartu...</div>
		{:else if error || !result?.valid}
			<div class="rounded-2xl border border-red-200 bg-red-50 p-6 text-red-800 shadow">{error || result?.message || 'Kartu tidak valid atau tidak aktif.'}</div>
		{:else}
			<div class="grid gap-6 lg:grid-cols-[1fr_360px]">
				<section class="rounded-3xl bg-white p-6 shadow-xl">
					<div class="mb-4 inline-flex rounded-full bg-emerald-100 px-3 py-1 text-sm font-semibold text-emerald-800">{result.message}</div>
					<h2 class="text-2xl font-bold">{result.student?.name}</h2>
					<div class="mt-4 grid gap-3 text-sm sm:grid-cols-2">
						<div><p class="text-slate-500">NIS</p><p class="font-semibold">{result.student?.nis || '-'}</p></div>
						<div><p class="text-slate-500">Kelas</p><p class="font-semibold">{result.student?.class_name || '-'}</p></div>
						<div><p class="text-slate-500">Nomor Kartu</p><p class="font-semibold">{result.card?.card_no || '-'}</p></div>
						<div><p class="text-slate-500">Status</p><p class="font-semibold">{result.card?.status || '-'}</p></div>
					</div>

					<div class="mt-6 rounded-2xl border p-4">
						<h3 class="font-semibold">Login Portal Siswa via QR + PIN</h3>
						<p class="mt-1 text-sm text-slate-600">Fitur ini memvalidasi QR dan PIN. Pembuatan session portal final tetap mengikuti auth siswa di backend.</p>
						<div class="mt-4 flex flex-col gap-2 sm:flex-row">
							<button class="rounded-lg bg-emerald-700 px-4 py-2 text-sm font-semibold text-white disabled:opacity-60" disabled={busy === 'start'} onclick={startLogin}>Mulai login QR</button>
							<input class="rounded-lg border px-3 py-2 text-sm" type="password" placeholder="PIN siswa" bind:value={pin} />
							<button class="rounded-lg bg-amber-500 px-4 py-2 text-sm font-semibold text-emerald-950 disabled:opacity-60" disabled={busy === 'complete'} onclick={completeLogin}>Validasi PIN</button>
						</div>
						{#if loginMessage}<p class="mt-3 text-sm text-slate-700">{loginMessage}</p>{/if}
					</div>
				</section>
				<section class="overflow-hidden rounded-3xl bg-white p-4 shadow-xl">
					<StudentIdCardTemplate card={{ card_no: result.card?.card_no, nama: result.student?.name, nis: result.student?.nis, class_name: result.student?.class_name, photo_url: result.student?.photo_url, status: result.card?.status, qr_token: params.token }} />
				</section>
			</div>
		{/if}
	</main>
</div>
