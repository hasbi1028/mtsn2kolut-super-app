<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import {
		fetchStudentPortalCbtSchedule,
		revealStudentPortalCbtToken,
		type StudentPortalCbtScheduleItem
	} from '$lib/client/student-portal';

	const participantId = page.params.participant_id ?? '';
	let item = $state<StudentPortalCbtScheduleItem | null>(null);
	let loading = $state(true);
	let errorMessage = $state('');
	let roomToken = $state('');
	let revealedToken = $state('');
	let revealError = $state('');
	let revealing = $state(false);

	onMount(loadCard);

	async function loadCard() {
		loading = true;
		errorMessage = '';
		try {
			const payload = await fetchStudentPortalCbtSchedule();
			item = payload.schedule.find((entry) => entry.participant_id === participantId) ?? null;
			if (!item) errorMessage = 'Kartu ujian CBT tidak ditemukan untuk akun ini.';
		} catch (error) {
			errorMessage = error instanceof Error ? error.message : 'Kartu ujian CBT belum dapat dimuat.';
		} finally {
			loading = false;
		}
	}

	async function revealToken() {
		if (!item || revealing) return;
		revealing = true;
		revealError = '';
		revealedToken = '';
		try {
			const payload = await revealStudentPortalCbtToken(item.participant_id, roomToken);
			revealedToken = payload.token;
		} catch (error) {
			revealError = error instanceof Error ? error.message : 'Token ujian belum dapat dibuka.';
		} finally {
			revealing = false;
		}
	}

	function fmtDate(value: string) {
		if (!value) return '—';
		return new Intl.DateTimeFormat('id-ID', {
			dateStyle: 'full',
			timeStyle: 'short',
			timeZone: 'Asia/Makassar'
		}).format(new Date(value));
	}

	function statusLabel(status: string) {
		return {
			upcoming: 'Belum dibuka',
			token_window: 'Token dapat dibuka',
			active: 'Sedang berlangsung',
			submitted: 'Selesai/submit',
			closed: 'Ditutup',
			locked: 'Dikunci pengawas'
		}[status] ?? status;
	}
</script>

<svelte:head><title>Kartu Ujian CBT — Portal Siswa</title></svelte:head>

<div class="mx-auto max-w-3xl space-y-4 p-6 print:p-0">
	<div class="flex items-center justify-between print:hidden">
		<a class="text-sm text-primary hover:underline" href={resolve('/portal/siswa')}>← Kembali ke Portal Siswa</a>
		<Button variant="outline" onclick={() => window.print()}>Cetak</Button>
	</div>

	{#if loading}
		<Card.Root><Card.Content class="p-6 text-sm text-muted-foreground">Memuat kartu ujian...</Card.Content></Card.Root>
	{:else if errorMessage}
		<RecoveryPanel title="Kartu ujian belum tersedia" message={errorMessage} onRetry={loadCard} />
	{:else if item}
		<Card.Root class="border-primary/30 shadow-sm print:shadow-none">
			<Card.Header class="text-center">
				<p class="text-xs font-semibold uppercase tracking-[0.2em] text-primary">Kartu Ujian CBT</p>
				<Card.Title class="text-2xl">{item.session_title}</Card.Title>
				<div><Badge variant="outline">{statusLabel(item.status)}</Badge></div>
			</Card.Header>
			<Card.Content class="space-y-5">
				<div class="grid gap-3 rounded-xl border border-border p-4 text-sm sm:grid-cols-2">
					<p><span class="font-semibold">Paket/Mapel:</span> {item.package_title || '—'}</p>
					<p><span class="font-semibold">Ruang:</span> {item.room_name || '—'}</p>
					<p><span class="font-semibold">Nomor meja:</span> {item.seat_no ?? '—'}</p>
					<p><span class="font-semibold">Durasi:</span> {item.duration_minutes} menit</p>
					<p class="sm:col-span-2"><span class="font-semibold">Mulai:</span> {fmtDate(item.scheduled_start)}</p>
					<p class="sm:col-span-2"><span class="font-semibold">Selesai:</span> {fmtDate(item.scheduled_end)}</p>
				</div>

				<div class="rounded-xl border border-warning/30 bg-warning/10 p-4 text-sm leading-6">
					<p class="font-semibold text-warning">Instruksi token</p>
					<p>Token ruang diberikan oleh pengawas saat peserta sudah berada di ruang ujian. Token siswa hanya dibuka pada waktu yang diizinkan dan jangan dibagikan ke perangkat lain.</p>
					<p class="mt-2">Token tersamarkan: <span class="font-mono font-semibold">{item.token_masked || 'Belum dibuka'}</span></p>
				</div>

				{#if item.can_reveal_token}
					<div class="space-y-3 rounded-xl border border-border p-4 print:hidden">
						<label class="text-sm font-semibold" for="room-token">Masukkan Token Ruang</label>
						<input id="room-token" class="w-full rounded-md border border-input bg-background px-3 py-2 font-mono text-sm" bind:value={roomToken} placeholder="Token dari pengawas" />
						<Button onclick={revealToken} disabled={revealing || roomToken.trim().length < 4}>{revealing ? 'Membuka...' : 'Buka Token Ujian'}</Button>
						{#if revealError}<p class="text-sm text-destructive">{revealError}</p>{/if}
						{#if revealedToken}
							<div class="rounded-lg border border-primary/30 bg-primary/10 p-3">
								<p class="text-xs font-semibold uppercase tracking-wide text-primary">Token ujian Anda</p>
								<p class="mt-1 font-mono text-xl font-bold tracking-wider">{revealedToken}</p>
							</div>
						{/if}
					</div>
				{:else}
					<p class="rounded-xl border border-border bg-muted/40 p-4 text-sm text-muted-foreground">Token belum dapat dibuka. Tunggu arahan pengawas saat ujian dimulai.</p>
				{/if}
			</Card.Content>
		</Card.Root>
	{/if}
</div>
