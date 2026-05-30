<script lang="ts">
	import CbtFeatureSurface from './CbtFeatureSurface.svelte';

	const actions = [{"label": "Buat/kelola ujian", "href": "/asesmen", "tone": "primary"}, {"label": "Dokumen & Cetak", "href": "/asesmen/cbt/dokumen", "tone": "soft"}];
	const stats = [{"label": "Ujian", "value": "Draft/Aktif", "detail": "Dikelola dari /asesmen"}, {"label": "Ruang", "value": "R01–R08", "detail": "Preview dan simpan 8 ruang"}, {"label": "Portal", "value": "Web-first", "detail": "/ujian dan /pengawas-ujian"}, {"label": "Dokumen", "value": "Terpisah", "detail": "QR+PIN tidak otomatis"}];
	const steps = [{"label": "Persiapan", "detail": "Buat draft ujian, pilih rombel, mapel, dan pastikan Bank Soal siap.", "badge": "Tahap 1"}, {"label": "Paket & Jadwal", "detail": "Mapping paket per rombel/mapel, atur tanggal, durasi, dan status sesi.", "badge": "Tahap 2"}, {"label": "Ruang & Dokumen", "detail": "Preview pembagian R01–R08, simpan kursi, lalu cetak dokumen.", "badge": "Tahap 3"}, {"label": "Pelaksanaan & Hasil", "detail": "Pengawas monitor ruang, peserta ujian, lalu rekap hasil.", "badge": "Tahap 4"}];
	const features = [{"title": "Quick Start Ujian", "detail": "Membuka Command Center produksi untuk membuat draft ujian, load rombel/mapel/paket, dan menjalankan aksi backend yang sudah ada.", "status": "backend", "actions": [{"label": "Buka Command Center", "href": "/asesmen", "tone": "primary"}]}, {"title": "Kesiapan ruang", "detail": "Shortcut ke panel ruang untuk auto-placement, validasi kapasitas, dan manual pindah kursi.", "status": "fitur", "actions": [{"label": "Ruang", "href": "/asesmen/cbt/ruang", "tone": "soft"}]}, {"title": "Cetak aman", "detail": "Kartu peserta QR+PIN, lembar pengawas, denah, daftar hadir, dan BA dikelompokkan dalam satu hub.", "status": "dokumen", "actions": [{"label": "Cetak", "href": "/asesmen/cbt/dokumen", "tone": "soft"}]}, {"title": "Portal hari-H", "detail": "Peserta memakai /ujian, pengawas memakai /pengawas-ujian; admin tetap di Super App.", "status": "portal", "actions": [{"label": "Portal Peserta", "href": "/ujian?demo=1", "tone": "soft"}, {"label": "Portal Pengawas", "href": "/pengawas-ujian?demo=1", "tone": "soft"}]}];
</script>

<svelte:head>
	<title>Dashboard Utama</title>
	<meta name="description" content="Panel harian yang menggabungkan alur legacy CBT dengan aksi Super App: buat ujian, siapkan paket, bagi ruang, cetak kartu, monitor, dan rekap." />
</svelte:head>

<CbtFeatureSurface
	title="Dashboard Utama"
	eyebrow="Command Center CBT"
	description="Panel harian yang menggabungkan alur legacy CBT dengan aksi Super App: buat ujian, siapkan paket, bagi ruang, cetak kartu, monitor, dan rekap."
	{actions}
	{stats}
	{steps}
	{features}
/>
