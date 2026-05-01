<script lang="ts">
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import PrinterIcon from '@lucide/svelte/icons/printer';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { schoolAddressLine, type SchoolProfile } from '$lib/school-profile';

	type CertificateDetail = {
		id: string;
		template_name: string;
		template_body: string;
		student_nis: string;
		student_nisn: string;
		student_name: string;
		student_gender: string;
		student_status: string;
		class_name: string;
		parent_name: string;
		nik: string;
		tempat_lahir: string;
		tanggal_lahir: string | null;
		alamat: string;
		agama: string;
		nomor_surat: string;
		classification_code: string;
		tanggal_surat: string;
		purpose: string;
		recipient: string;
		remarks: string;
		status: string;
		created_by_username: string;
	};

	let { data }: { data: { certificate: CertificateDetail; schoolProfile: SchoolProfile } } = $props();
	let certificate = $derived(data.certificate);
	let schoolProfile = $derived(data.schoolProfile);

	function formatDate(value: string | null | undefined) {
		if (!value) return '-';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return value.slice(0, 10);
		return new Intl.DateTimeFormat('id-ID', { day: 'numeric', month: 'long', year: 'numeric' }).format(date);
	}

	function genderLabel(value: string) {
		if (value === 'male' || value === 'L') return 'Laki-laki';
		if (value === 'female' || value === 'P') return 'Perempuan';
		return value || '-';
	}

	function statusLabel(value: string) {
		if (value === 'active') return 'Aktif';
		if (value === 'alumni') return 'Alumni';
		if (value === 'mutated') return 'Pindah';
		if (value === 'prospective') return 'Calon siswa';
		return value || '-';
	}

</script>

<svelte:head>
	<title>Cetak {certificate.nomor_surat} | MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="mx-auto max-w-4xl space-y-4 py-4">
	<div class="no-print flex flex-wrap items-center justify-between gap-3 rounded-lg border bg-background px-4 py-3">
		<div>
			<p class="text-sm font-semibold text-slate-950">Cetak Surat Keterangan</p>
			<p class="text-sm text-muted-foreground">{certificate.nomor_surat}</p>
		</div>
		<div class="flex gap-2">
			<Button variant="outline" href="/tu/surat-keterangan">
				<ArrowLeftIcon class="size-4" />
				Kembali
			</Button>
			<Button onclick={() => window.print()}>
				<PrinterIcon class="size-4" />
				Cetak
			</Button>
		</div>
	</div>

	<main class="print-page bg-white px-10 py-10 text-slate-950 shadow-sm ring-1 ring-slate-200">
		<header class="border-b-4 border-slate-900 pb-4 text-center">
			<p class="text-sm font-semibold uppercase">{schoolProfile.ministry_line}</p>
			<p class="text-sm font-semibold uppercase">{schoolProfile.office_line}</p>
			<h1 class="mt-1 text-xl font-bold uppercase">{schoolProfile.name}</h1>
			<p class="mt-1 text-xs leading-5">{schoolAddressLine(schoolProfile) || 'Alamat madrasah belum diisi'}</p>
			{#if schoolProfile.nsm || schoolProfile.npsn}
				<p class="mt-1 text-xs leading-5">
					{#if schoolProfile.nsm}NSM {schoolProfile.nsm}{/if}
					{#if schoolProfile.nsm && schoolProfile.npsn} · {/if}
					{#if schoolProfile.npsn}NPSN {schoolProfile.npsn}{/if}
				</p>
			{/if}
		</header>

		<section class="mt-8 text-center">
			<h2 class="text-base font-bold uppercase underline">{certificate.template_name}</h2>
			<p class="mt-1 text-sm">Nomor: {certificate.nomor_surat}</p>
			{#if certificate.status === 'canceled'}
				<div class="mt-3">
					<Badge variant="secondary">Dibatalkan</Badge>
				</div>
			{/if}
		</section>

		<section class="mt-8 space-y-5 text-sm leading-7">
			<p>
				Yang bertanda tangan di bawah ini, Kepala {schoolProfile.name}, menerangkan bahwa:
			</p>

			<div class="mx-auto max-w-2xl">
				<div class="grid grid-cols-[150px_12px_1fr] gap-y-2">
					<span>Nama</span><span>:</span><span class="font-semibold">{certificate.student_name}</span>
					<span>NIS / NISN</span><span>:</span><span>{certificate.student_nis || '-'} / {certificate.student_nisn || '-'}</span>
					<span>NIK</span><span>:</span><span>{certificate.nik || '-'}</span>
					<span>Tempat, Tgl Lahir</span><span>:</span><span>{certificate.tempat_lahir || '-'}, {formatDate(certificate.tanggal_lahir)}</span>
					<span>Jenis Kelamin</span><span>:</span><span>{genderLabel(certificate.student_gender)}</span>
					<span>Kelas</span><span>:</span><span>{certificate.class_name || '-'}</span>
					<span>Status Siswa</span><span>:</span><span>{statusLabel(certificate.student_status)}</span>
					<span>Nama Orang Tua</span><span>:</span><span>{certificate.parent_name || '-'}</span>
					<span>Alamat</span><span>:</span><span>{certificate.alamat || '-'}</span>
				</div>
			</div>

			<p>{certificate.template_body}</p>
			<p>
				Surat keterangan ini diterbitkan untuk keperluan: <span class="font-semibold">{certificate.purpose}</span>.
			</p>
			<p>
				Demikian surat keterangan ini dibuat dengan sebenarnya untuk dipergunakan sebagaimana mestinya.
			</p>
		</section>

		<section class="mt-12 flex justify-end text-sm leading-7">
			<div class="w-64">
				<p>Kolaka Utara, {formatDate(certificate.tanggal_surat)}</p>
				<p>Kepala Madrasah,</p>
				<div class="h-20"></div>
				<p class="font-semibold underline">{schoolProfile.head_name || '........................................'}</p>
				<p>NIP. {schoolProfile.head_nip || '................................'}</p>
			</div>
		</section>

		<footer class="mt-8 border-t pt-3 text-xs text-muted-foreground">
			<p>Dicatat oleh TU: {certificate.created_by_username || '-'} · Klasifikasi: {certificate.classification_code}</p>
		</footer>
	</main>
</div>

<style>
	@media print {
		:global(body) {
			background: white;
		}

		.no-print {
			display: none !important;
		}

		.print-page {
			box-shadow: none !important;
			width: 100%;
			min-height: 100vh;
			padding: 0;
		}
	}
</style>
