<script lang="ts">
	import { readClientApiData, readClientJson } from '$lib/client/api';

	let token = $state('');
	let pin = $state('');
	let challengeId = $state('');
	let message = $state('');
	let busy = $state('');

	async function start() {
		if (!token.trim()) { message = 'Tempel URL/token QR kartu terlebih dahulu.'; return; }
		busy = 'start';
		try {
			const res = await readClientApiData<{ challenge_id: string; student_hint?: { name?: string; class_name?: string } }>(await fetch('/api/public/student-cards/portal-login/start', {
				method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ qr_token: token })
			}), 'QR kartu tidak valid');
			challengeId = res.challenge_id;
			message = `QR valid untuk ${res.student_hint?.name || 'siswa'} ${res.student_hint?.class_name ? '(' + res.student_hint.class_name + ')' : ''}. Masukkan PIN.`;
		} catch (e) { message = (e as Error).message || 'QR kartu tidak valid'; }
		finally { busy = ''; }
	}

	async function complete() {
		if (!challengeId || !pin.trim()) { message = 'Klik Mulai dan isi PIN terlebih dahulu.'; return; }
		busy = 'complete';
		try {
			const res = await readClientJson<{ data?: { message?: string } }>(await fetch('/api/public/student-cards/portal-login/complete', {
				method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ challenge_id: challengeId, pin })
			}));
			message = res.data?.message || 'PIN valid. Session portal siswa siap disambungkan.';
		} catch (e) { message = (e as Error).message || 'PIN tidak valid'; }
		finally { busy = ''; }
	}
</script>

<svelte:head><title>Login Portal Siswa via Kartu</title></svelte:head>

<div class="min-h-screen bg-gradient-to-br from-emerald-950 via-emerald-800 to-amber-600 px-4 py-10 text-white">
	<main class="mx-auto max-w-xl rounded-3xl bg-white/95 p-6 text-slate-900 shadow-2xl">
		<p class="text-sm font-bold uppercase tracking-[0.2em] text-emerald-700">Portal Siswa</p>
		<h1 class="mt-2 text-3xl font-bold">Login QR + PIN</h1>
		<p class="mt-2 text-sm text-slate-600">Scan kartu akan mengisi token/URL QR. QR bukan password; akun tetap perlu PIN siswa.</p>
		<div class="mt-6 space-y-3">
			<label class="block text-sm font-medium" for="qr-token-input">URL/token QR kartu</label>
			<textarea id="qr-token-input" class="min-h-24 w-full rounded-xl border p-3 text-sm" bind:value={token} placeholder="Tempel hasil scan QR kartu di sini"></textarea>
			<button class="w-full rounded-xl bg-emerald-700 px-4 py-3 font-semibold text-white disabled:opacity-60" disabled={busy === 'start'} onclick={start}>Mulai validasi QR</button>
			<label class="block text-sm font-medium" for="student-pin-input">PIN siswa</label>
			<input id="student-pin-input" class="w-full rounded-xl border p-3 text-sm" type="password" bind:value={pin} placeholder="Masukkan PIN" />
			<button class="w-full rounded-xl bg-amber-500 px-4 py-3 font-semibold text-emerald-950 disabled:opacity-60" disabled={busy === 'complete'} onclick={complete}>Validasi PIN</button>
			{#if message}<div class="rounded-xl bg-slate-100 p-3 text-sm text-slate-700">{message}</div>{/if}
		</div>
	</main>
</div>
