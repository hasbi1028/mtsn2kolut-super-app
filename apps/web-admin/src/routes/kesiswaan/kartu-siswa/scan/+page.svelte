<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import * as Card from '$lib/components/ui/card';
	import { readClientJson } from '$lib/client/api';

	let qrToken = $state('');
	let activityCode = $state('general');
	let scanType = $state('present');
	let mode = $state<'attendance' | 'library' | 'cbt'>('attendance');
	let busy = $state(false);
	let result = $state<unknown>(null);
	let error = $state('');

	async function scan() {
		if (!qrToken.trim()) { error = 'Tempel URL/token QR hasil scan kartu.'; return; }
		busy = true; error = ''; result = null;
		try {
			const body = mode === 'attendance'
				? { qr_token: qrToken, activity_code: activityCode || 'general', scan_type: scanType || 'present' }
				: { qr_token: qrToken };
			result = await readClientJson(await fetch(`/api/kesiswaan/kartu-siswa/scan/${mode}`, {
				method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body)
			}));
		} catch (e) { error = (e as Error).message || 'Scan gagal'; }
		finally { busy = false; }
	}
</script>

<div class="space-y-6 p-4 md:p-6">
	<div>
		<p class="text-sm text-muted-foreground">Kesiswaan</p>
		<h1 class="text-2xl font-semibold tracking-tight">Scanner Kartu Siswa</h1>
		<p class="text-sm text-muted-foreground">Scan kartu untuk presensi kegiatan, lookup perpustakaan, atau validasi peserta CBT. Scan CBT tidak membuka token ujian.</p>
	</div>

	<Card.Root>
		<Card.Header><Card.Title>Mode scan</Card.Title><Card.Description>Pilih konteks layanan sebelum scan.</Card.Description></Card.Header>
		<Card.Content class="space-y-4">
			<div class="grid gap-3 md:grid-cols-3">
				<Button variant={mode === 'attendance' ? 'default' : 'outline'} onclick={() => mode = 'attendance'}>Presensi</Button>
				<Button variant={mode === 'library' ? 'default' : 'outline'} onclick={() => mode = 'library'}>Perpustakaan</Button>
				<Button variant={mode === 'cbt' ? 'default' : 'outline'} onclick={() => mode = 'cbt'}>Validasi CBT</Button>
			</div>
			<textarea class="min-h-28 w-full rounded-md border bg-background p-3 text-sm" bind:value={qrToken} placeholder="Tempel URL/token QR kartu siswa di sini"></textarea>
			{#if mode === 'attendance'}
				<div class="grid gap-3 md:grid-cols-2">
					<Input bind:value={activityCode} placeholder="Kode kegiatan, contoh: upacara-2026-05-22" />
					<select bind:value={scanType} class="h-10 rounded-md border bg-background px-3 text-sm">
						<option value="present">Hadir</option>
						<option value="late">Terlambat</option>
						<option value="out">Pulang/Keluar</option>
					</select>
				</div>
			{/if}
			<Button onclick={scan} disabled={busy}>{busy ? 'Memproses...' : 'Scan kartu'}</Button>
		</Card.Content>
	</Card.Root>

	{#if error}<div class="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700">{error}</div>{/if}
	{#if result}
		<Card.Root>
			<Card.Header><Card.Title>Hasil scan</Card.Title></Card.Header>
			<Card.Content><pre class="overflow-auto rounded-lg bg-muted p-4 text-xs">{JSON.stringify(result, null, 2)}</pre></Card.Content>
		</Card.Root>
	{/if}
</div>
