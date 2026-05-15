<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { toast } from '$lib/components/ui/sonner';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import { readClientJson } from '$lib/client/api';

	type Settings = {
		is_enabled: boolean;
		send_time: string;
		timezone: string;
		target_chat_id: string;
		target_chat_id_masked?: string;
		include_caption: boolean;
		include_image: boolean;
		report_mode: string;
		bot_configured?: boolean;
	};
	type LogRow = { id: string; report_date: string; target_chat_id_masked: string; send_mode: string; status: string; telegram_message_id?: string; error_message?: string; sent_at: string };

	let settings = $state<Settings>({ is_enabled: false, send_time: '17:00', timezone: 'Asia/Makassar', target_chat_id: '1450267717', include_caption: true, include_image: true, report_mode: 'ringkas' });
	let logs = $state<LogRow[]>([]);
	let loading = $state(true);
	let saving = $state(false);
	let testing = $state(false);
	let testDate = $state('');

	function todayWita() {
		return new Intl.DateTimeFormat('en-CA', { timeZone: 'Asia/Makassar', year: 'numeric', month: '2-digit', day: '2-digit' }).format(new Date());
	}
	function dataOf(payload: unknown): unknown {
		return typeof payload === 'object' && payload !== null && 'data' in payload ? (payload as { data: unknown }).data : payload;
	}
	async function loadSettings() {
		loading = true;
		try {
			const [settingsRes, logsRes] = await Promise.all([
				fetch('/api/pusaka/attendance-telegram/settings'),
				fetch('/api/pusaka/attendance-telegram/logs?limit=20')
			]);
			settings = { ...settings, ...(dataOf(await readClientJson(settingsRes)) as Partial<Settings>) };
			const logData = dataOf(await readClientJson(logsRes));
			logs = Array.isArray(logData) ? logData as LogRow[] : [];
		} catch (error) {
			toast.error(error instanceof Error ? error.message : 'Gagal memuat pengaturan Telegram');
		} finally {
			loading = false;
		}
	}
	async function saveSettings() {
		saving = true;
		try {
			const res = await fetch('/api/pusaka/attendance-telegram/settings', { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(settings) });
			const payload = await readClientJson<Record<string, unknown>>(res);
			if (payload.error) throw new Error(String(payload.error));
			settings = { ...settings, ...(dataOf(payload) as Partial<Settings>) };
			toast.success('Pengaturan laporan Telegram tersimpan');
		} catch (error) {
			toast.error(error instanceof Error ? error.message : 'Gagal menyimpan pengaturan');
		} finally {
			saving = false;
		}
	}
	async function testSend() {
		testing = true;
		try {
			const res = await fetch('/api/pusaka/attendance-telegram/send', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ date: testDate || todayWita(), target_chat_id: settings.target_chat_id, include_caption: settings.include_caption, include_image: settings.include_image }) });
			const payload = await readClientJson<Record<string, unknown>>(res);
			if (payload.error) throw new Error(String(payload.error));
			toast.success('Test laporan Telegram berhasil dikirim');
			await loadSettings();
		} catch (error) {
			toast.error(error instanceof Error ? error.message : 'Gagal test kirim Telegram');
		} finally {
			testing = false;
		}
	}
	onMount(() => { testDate = todayWita(); void loadSettings(); });
</script>

<svelte:head><title>Laporan Telegram PUSAKA — MTsN 2 Kolut</title></svelte:head>

<div class="space-y-6">
	<div class="flex items-center gap-2 text-sm text-muted-foreground">
		<a href={resolve('/pusaka')} class="hover:text-foreground">PUSAKA</a><span>/</span>
		<span class="text-foreground font-medium">Laporan Telegram</span>
	</div>
	<div class="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
		<div>
			<h1 class="text-2xl font-semibold text-foreground">Laporan Telegram Daftar Hadir Ringkas</h1>
			<p class="mt-1 text-sm text-muted-foreground">Admin dapat mengatur jam kirim otomatis WITA dan melakukan test kirim manual.</p>
		</div>
		<LoadingButton variant="outline" href={resolve('/pusaka/kehadiran')} label="← Daftar Hadir" />
	</div>

	<Card.Root>
		<Card.Header><Card.Title>Pengaturan Jadwal</Card.Title><Card.Description>Token bot disimpan di backend environment, bukan di browser.</Card.Description></Card.Header>
		<Card.Content class="grid gap-4 md:grid-cols-2">
			<label class="flex items-center gap-3 rounded-lg border p-3 text-sm"><input aria-label="Aktifkan kirim otomatis harian" type="checkbox" bind:checked={settings.is_enabled} /> Aktifkan kirim otomatis harian</label>
			<div><label for="telegram-send-time" class="text-sm font-medium">Jam kirim WITA</label><Input id="telegram-send-time" type="time" bind:value={settings.send_time} /></div>
			<div><label for="telegram-timezone" class="text-sm font-medium">Zona waktu</label><Input id="telegram-timezone" value={settings.timezone} disabled /></div>
			<div><label for="telegram-chat-id" class="text-sm font-medium">Target Chat ID Telegram</label><Input id="telegram-chat-id" bind:value={settings.target_chat_id} placeholder="1450267717 atau ID grup/channel" /></div>
			<label class="flex items-center gap-3 rounded-lg border p-3 text-sm"><input aria-label="Sertakan caption ringkasan" type="checkbox" bind:checked={settings.include_caption} /> Sertakan caption ringkasan</label>
			<label class="flex items-center gap-3 rounded-lg border p-3 text-sm"><input aria-label="Sertakan gambar PNG laporan" type="checkbox" bind:checked={settings.include_image} /> Sertakan gambar PNG laporan</label>
			<div class="md:col-span-2 flex flex-wrap gap-2">
				<LoadingButton onclick={() => void saveSettings()} loading={saving} loadingLabel="Menyimpan..." label="Simpan Pengaturan" disabled={loading} />
				<Input type="date" bind:value={testDate} class="w-auto" />
				<LoadingButton variant="outline" onclick={() => void testSend()} loading={testing} loadingLabel="Mengirim..." label="Test Kirim" disabled={loading} />
			</div>
			{#if settings.bot_configured === false}
				<p class="md:col-span-2 rounded-lg border border-warning/40 bg-warning/10 p-3 text-sm text-warning">Bot Telegram belum dikonfigurasi di backend. Set environment <code>TELEGRAM_BOT_TOKEN</code>, lalu restart core-api.</p>
			{/if}
		</Card.Content>
	</Card.Root>

	<Card.Root>
		<Card.Header><Card.Title>Riwayat Pengiriman</Card.Title><Card.Description>Target chat ditampilkan masking untuk menjaga privasi.</Card.Description></Card.Header>
		<Card.Content class="overflow-x-auto">
			<table class="w-full text-sm">
				<thead><tr class="border-b text-left text-muted-foreground"><th class="py-2">Tanggal</th><th>Mode</th><th>Status</th><th>Target</th><th>Waktu</th><th>Error</th></tr></thead>
				<tbody>
					{#each logs as row (row.id)}
						<tr class="border-b"><td class="py-2">{row.report_date}</td><td>{row.send_mode}</td><td>{row.status}</td><td>{row.target_chat_id_masked}</td><td>{row.sent_at}</td><td class="text-destructive">{row.error_message ?? ''}</td></tr>
					{:else}
						<tr><td colspan="6" class="py-6 text-center text-muted-foreground">Belum ada riwayat pengiriman.</td></tr>
					{/each}
				</tbody>
			</table>
		</Card.Content>
	</Card.Root>
</div>
