<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { toast } from '$lib/components/ui/sonner';
	import LoadingButton from '$lib/components/LoadingButton.svelte';

	let nama = $state('');
	let nis = $state('');
	let gender = $state('L');
	let parentName = $state('');
	let parentPhone = $state('');
	let busy = $state(false);

	async function submitRegistration() {
		if (!nama || !nis || !gender) {
			toast.error('Nama, NIS, dan jenis kelamin wajib diisi');
			return;
		}
		busy = true;
		try {
			const res = await fetch('/api/public/register-student', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					nama,
					nis,
					gender,
					parent_name: parentName,
					parent_phone: parentPhone,
				}),
			});
			const payload = await res.json().catch(() => ({}));
			if (!res.ok) {
				toast.error(payload.error ?? 'Pendaftaran gagal');
				return;
			}
			toast.success('Pendaftaran berhasil dikirim. Status awal sebagai calon siswa.');
			nama = '';
			nis = '';
			gender = 'L';
			parentName = '';
			parentPhone = '';
		} finally {
			busy = false;
		}
	}
</script>

<svelte:head><title>PPDB Awal — MTsN 2 Kolaka Utara</title></svelte:head>

<div class="min-h-screen bg-[linear-gradient(180deg,rgba(236,253,245,0.95),rgba(255,255,255,1))] px-4 py-10">
	<div class="mx-auto max-w-3xl space-y-6">
		<div class="space-y-2 text-center">
			<p class="text-sm font-semibold uppercase tracking-[0.28em] text-emerald-700">PPDB Awal</p>
			<h1 class="text-3xl font-semibold text-slate-900">Pendaftaran Calon Siswa MTs Negeri 2 Kolaka Utara</h1>
			<p class="mx-auto max-w-2xl text-sm text-slate-600">
				Form ini untuk pendaftaran awal. Data akan masuk sebagai <span class="font-medium">calon siswa</span> dan diverifikasi admin sebelum menjadi siswa aktif.
			</p>
		</div>

		<Card.Root class="border-emerald-200 shadow-sm">
			<Card.Header class="border-b bg-emerald-50/60">
				<Card.Title class="text-lg text-slate-900">Formulir Pendaftaran Awal</Card.Title>
				<Card.Description>Isi data dasar terlebih dahulu. Detail lanjutan bisa dilengkapi setelah diverifikasi.</Card.Description>
			</Card.Header>
			<Card.Content class="space-y-5 pt-6">
				<div class="grid gap-4 sm:grid-cols-2">
					<div class="sm:col-span-2">
						<label for="ppdb-nama" class="mb-1 block text-xs font-medium text-slate-600">Nama Lengkap</label>
						<Input id="ppdb-nama" bind:value={nama} placeholder="Nama calon siswa" />
					</div>
					<div>
						<label for="ppdb-nis" class="mb-1 block text-xs font-medium text-slate-600">NIS / Nomor Pendaftaran</label>
						<Input id="ppdb-nis" bind:value={nis} placeholder="Nomor identitas calon siswa" />
					</div>
					<div>
						<label for="ppdb-gender" class="mb-1 block text-xs font-medium text-slate-600">Jenis Kelamin</label>
						<select id="ppdb-gender" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={gender}>
							<option value="L">Laki-laki</option>
							<option value="P">Perempuan</option>
						</select>
					</div>
					<div>
						<label for="ppdb-parent" class="mb-1 block text-xs font-medium text-slate-600">Nama Orang Tua / Wali</label>
						<Input id="ppdb-parent" bind:value={parentName} placeholder="Nama orang tua atau wali" />
					</div>
					<div>
						<label for="ppdb-phone" class="mb-1 block text-xs font-medium text-slate-600">Nomor HP Orang Tua / Wali</label>
						<Input id="ppdb-phone" bind:value={parentPhone} placeholder="08xxxxxxxxxx" />
					</div>
				</div>

				<div class="rounded-xl border border-slate-200 bg-slate-50 p-4 text-sm text-slate-600">
					Setelah pendaftaran dikirim, admin akan meninjau data ini dan mengubah status siswa menjadi aktif saat proses PPDB selesai.
				</div>

				<div class="flex flex-wrap gap-2">
					<LoadingButton onclick={submitRegistration} loading={busy} loadingLabel="Mengirim..." label="Kirim Pendaftaran" />
					<Button variant="outline" href="/login">Masuk Admin</Button>
				</div>
			</Card.Content>
		</Card.Root>
	</div>
</div>
