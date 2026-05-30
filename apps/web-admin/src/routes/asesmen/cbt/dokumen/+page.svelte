<script lang="ts">
	import CbtFeatureSurface from '../CbtFeatureSurface.svelte';

	const actions = [{"label": "Cetak produksi", "href": "/asesmen/dokumen", "tone": "primary"}, {"label": "Command Center", "href": "/asesmen", "tone": "soft"}];
	const stats = [{"label": "Kartu", "value": "QR+PIN", "detail": "Terbit eksplisit"}, {"label": "Pengawas", "value": "Lembar ruang", "detail": "Tanpa bocor PIN peserta"}, {"label": "Denah", "value": "Kursi", "detail": "Per ruang"}, {"label": "Arsip", "value": "BA+Rekap", "detail": "Final"}];
	const steps = [{"label": "Cek ruang", "detail": "Pastikan peserta sudah punya ruang/kursi.", "badge": "1"}, {"label": "Terbitkan QR+PIN", "detail": "Hanya setelah konfirmasi panitia.", "badge": "2"}, {"label": "Cetak paket", "detail": "Kartu, denah, hadir, pengawas, BA.", "badge": "3"}, {"label": "Arsipkan", "detail": "Simpan hasil dan dokumen final.", "badge": "4"}];
	const features = [{"title": "Kartu peserta QR+PIN", "detail": "Aksi sensitif tetap terpisah: QR+PIN tidak dibuat otomatis dari preview ruang.", "status": "aman", "actions": [{"label": "Dokumen produksi", "href": "/asesmen/dokumen", "tone": "primary"}]}, {"title": "Lembar pengawas ruang", "detail": "Pengawas mendapat daftar peserta/ruang tanpa PIN rahasia peserta.", "status": "fitur", "actions": []}, {"title": "Denah & daftar hadir", "detail": "Cetak per ruang untuk ditempel/dipegang pengawas saat ujian.", "status": "fitur", "actions": []}, {"title": "Berita Acara & arsip", "detail": "BA, rekap, dan checklist final ditempatkan dalam satu pusat dokumen.", "status": "arsip", "actions": []}];
</script>

<svelte:head>
	<title>Dokumen & Cetak</title>
	<meta name="description" content="Fitur dokumen: kartu peserta QR+PIN, lembar pengawas ruang, denah ruang, daftar hadir, berita acara, rekap nilai, dan checklist arsip." />
</svelte:head>

<CbtFeatureSurface
	title="Dokumen & Cetak"
	eyebrow="Laporan"
	description="Fitur dokumen: kartu peserta QR+PIN, lembar pengawas ruang, denah ruang, daftar hadir, berita acara, rekap nilai, dan checklist arsip."
	{actions}
	{stats}
	{steps}
	{features}
/>
