<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { toast } from '$lib/components/ui/sonner';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import { readClientJson } from '$lib/client/api';
	import { trackPublicAnalyticsEvent } from '$lib/analytics/public-analytics';

	let nama = $state('');
	let nis = $state('');
	let gender = $state('L');
	let parentName = $state('');
	let parentPhone = $state('');
	let busy = $state(false);
	let registrationStarted = $state(false);

	function trackRegistrationStart() {
		if (registrationStarted) return;
		registrationStarted = true;
		void trackPublicAnalyticsEvent('public.form_start', {
			pathname: '/ppdb',
			metadata: { page_key: 'ppdb', form_key: 'ppdb_registration', form_step: 'identity' }
		});
	}

	async function submitRegistration() {
		if (!nama || !nis || !gender) {
			void trackPublicAnalyticsEvent('public.form_submit', {
				pathname: '/ppdb',
				metadata: { page_key: 'ppdb', form_key: 'ppdb_registration', result: 'validation_failed' },
				result: 'validation_failed'
			});
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
			await readClientJson<unknown>(res);
			void trackPublicAnalyticsEvent('public.form_submit', {
				pathname: '/ppdb',
				metadata: { page_key: 'ppdb', form_key: 'ppdb_registration', result: 'success' },
				result: 'success'
			});
			toast.success('Pendaftaran berhasil dikirim. Status awal sebagai calon siswa.');
			nama = '';
			nis = '';
			gender = 'L';
			parentName = '';
			parentPhone = '';
		} catch (error) {
			void trackPublicAnalyticsEvent('public.form_submit', {
				pathname: '/ppdb',
				metadata: { page_key: 'ppdb', form_key: 'ppdb_registration', result: 'failed' },
				result: 'failed'
			});
			toast.error(error instanceof Error && error.message.trim() ? error.message : 'Pendaftaran gagal dikirim. Periksa koneksi lalu coba lagi.');
		} finally {
			busy = false;
		}
	}
</script>

<svelte:head><title>PPDB Awal — MTsN 2 Kolaka Utara</title></svelte:head>

<div class="min-h-screen bg-[linear-gradient(180deg,rgba(236,253,245,0.95),rgba(255,255,255,1))] px-4 py-10">
	<div class="mx-auto max-w-6xl space-y-8">
		<div class="rounded-[2rem] border border-emerald-100 bg-white/80 px-6 py-8 shadow-sm backdrop-blur sm:px-8 sm:py-10">
			<div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_20rem] lg:items-end">
				<div class="space-y-3">
					<p class="text-sm font-semibold uppercase tracking-[0.28em] text-emerald-700">PPDB Awal</p>
					<h1 class="max-w-3xl text-3xl font-semibold leading-tight text-slate-900 sm:text-5xl">
						Pendaftaran Calon Siswa MTs Negeri 2 Kolaka Utara
					</h1>
					<p class="max-w-3xl text-sm leading-7 text-slate-600 sm:text-base">
						Form ini digunakan untuk pendaftaran awal. Data akan masuk sebagai <span class="font-medium text-slate-900">calon siswa</span> dan diverifikasi admin sebelum berubah menjadi siswa aktif.
					</p>
				</div>
				<div class="rounded-[1.75rem] border border-emerald-100 bg-emerald-50 px-5 py-5">
					<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-emerald-700">Alur Singkat</p>
					<div class="mt-4 space-y-3 text-sm leading-7 text-slate-700">
						<p>1. Isi data dasar calon siswa.</p>
						<p>2. Admin memeriksa kelengkapan awal.</p>
						<p>3. Status diperbarui setelah proses PPDB selesai.</p>
					</div>
				</div>
			</div>
		</div>

		<div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_22rem]">
			<Card.Root class="border-emerald-200 shadow-sm">
				<Card.Header class="border-b bg-emerald-50/60">
					<Card.Title class="text-lg text-slate-900">Formulir Pendaftaran Awal</Card.Title>
					<Card.Description>Isi data dasar terlebih dahulu. Detail lanjutan bisa dilengkapi setelah verifikasi awal.</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-5 pt-6">
					<form name="ppdb_registration" data-analytics-form="ppdb_registration" class="grid gap-4 sm:grid-cols-2" onsubmit={(event) => { event.preventDefault(); void submitRegistration(); }}>
						<div class="sm:col-span-2">
							<label for="ppdb-nama" class="mb-1 block text-xs font-medium text-slate-600">Nama Lengkap</label>
							<Input id="ppdb-nama" bind:value={nama} onfocus={trackRegistrationStart} placeholder="Tuliskan nama lengkap calon siswa" />
						</div>
						<div>
							<label for="ppdb-nis" class="mb-1 block text-xs font-medium text-slate-600">NIS / Nomor Pendaftaran</label>
							<Input id="ppdb-nis" bind:value={nis} onfocus={trackRegistrationStart} placeholder="Nomor identitas atau nomor pendaftaran" />
						</div>
						<div>
							<label for="ppdb-gender" class="mb-1 block text-xs font-medium text-slate-600">Jenis Kelamin</label>
							<select id="ppdb-gender" class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm" bind:value={gender} onfocus={trackRegistrationStart}>
								<option value="L">Laki-laki</option>
								<option value="P">Perempuan</option>
							</select>
						</div>
						<div>
							<label for="ppdb-parent" class="mb-1 block text-xs font-medium text-slate-600">Nama Orang Tua / Wali</label>
							<Input id="ppdb-parent" bind:value={parentName} onfocus={trackRegistrationStart} placeholder="Nama orang tua atau wali utama" />
						</div>
						<div>
							<label for="ppdb-phone" class="mb-1 block text-xs font-medium text-slate-600">Nomor HP Orang Tua / Wali</label>
							<Input id="ppdb-phone" bind:value={parentPhone} onfocus={trackRegistrationStart} placeholder="Gunakan nomor yang aktif dihubungi" />
						</div>
					</form>

					<div class="rounded-xl border border-slate-200 bg-slate-50 p-4 text-sm leading-7 text-slate-600">
						Setelah pendaftaran dikirim, admin akan meninjau data ini dan menghubungi calon siswa atau wali bila diperlukan untuk melengkapi berkas lanjutan.
					</div>

					<div class="flex flex-wrap gap-2">
						<LoadingButton onclick={() => void submitRegistration()} loading={busy} loadingLabel="Mengirim..." label="Kirim Pendaftaran" />
						<Button variant="outline" href="/login">Masuk Admin</Button>
					</div>
				</Card.Content>
			</Card.Root>

			<div class="space-y-4">
				<Card.Root class="border-slate-200 shadow-sm">
					<Card.Header class="pb-2">
						<Card.Title class="text-base text-slate-900">Yang Perlu Disiapkan</Card.Title>
					</Card.Header>
					<Card.Content class="space-y-3 text-sm leading-7 text-slate-600">
						<p>Pastikan nama calon siswa ditulis lengkap dan mudah dicocokkan dengan dokumen sekolah sebelumnya.</p>
						<p>Gunakan nomor HP wali yang aktif agar tim sekolah mudah melakukan konfirmasi lanjutan.</p>
						<p>Jika belum memiliki nomor pendaftaran, isi identitas yang paling mudah dikenali terlebih dahulu.</p>
					</Card.Content>
				</Card.Root>

				<Card.Root class="border-emerald-100 bg-emerald-50 shadow-sm">
					<Card.Header class="pb-2">
						<Card.Title class="text-base text-emerald-900">Setelah Mengirim Form</Card.Title>
					</Card.Header>
					<Card.Content class="space-y-3 text-sm leading-7 text-slate-700">
						<p>Data akan masuk ke antrean verifikasi admin.</p>
						<p>Status awal tersimpan sebagai <span class="font-medium text-slate-900">calon siswa</span>.</p>
						<p>Informasi lanjutan terkait jadwal dan kelengkapan akan disampaikan oleh pihak sekolah.</p>
					</Card.Content>
				</Card.Root>
			</div>
		</div>
	</div>
</div>
