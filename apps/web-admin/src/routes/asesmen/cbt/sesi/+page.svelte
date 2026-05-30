<script lang="ts">
	import CbtFeatureSurface from '../CbtFeatureSurface.svelte';

	const actions = [{"label": "Kelola sesi", "href": "/asesmen", "tone": "primary"}, {"label": "Pelaksanaan", "href": "/asesmen/pelaksanaan", "tone": "soft"}];
	const stats = [{"label": "Status", "value": "Draft/Aktif", "detail": "Sesi dikontrol panitia"}, {"label": "Durasi", "value": "90 menit", "detail": "Default dapat diubah"}, {"label": "Token", "value": "Opsional", "detail": "Bukan password utama"}, {"label": "Dokumen", "value": "Absen+BA", "detail": "Cetak per sesi"}];
	const steps = [{"label": "Draft sesi", "detail": "Tentukan paket, tanggal, jam, dan durasi.", "badge": "Draft"}, {"label": "Siapkan ruang", "detail": "Apply peserta ke ruang/kursi.", "badge": "Ruang"}, {"label": "Buka sesi", "detail": "Peserta masuk saat jadwal/panitia siap.", "badge": "Mulai"}, {"label": "Tutup rekap", "detail": "Finalisasi hasil dan dokumen.", "badge": "Selesai"}];
	const features = [{"title": "Jadwal dan durasi", "detail": "Form di Command Center menyiapkan tanggal mulai/selesai dan durasi sesi.", "status": "fitur", "actions": [{"label": "Command Center", "href": "/asesmen", "tone": "primary"}]}, {"title": "Kontrol buka/tutup", "detail": "Status sesi dipakai untuk membedakan draft, berlangsung, selesai, dan arsip.", "status": "fitur", "actions": []}, {"title": "Daftar hadir", "detail": "Dokumen sesi diarahkan ke Dokumen & Cetak agar pengawas tidak mencari banyak menu.", "status": "dokumen", "actions": [{"label": "Dokumen", "href": "/asesmen/cbt/dokumen", "tone": "soft"}]}, {"title": "Berita acara", "detail": "BA pelaksanaan disiapkan sebagai arsip per kegiatan/sesi.", "status": "dokumen", "actions": [{"label": "BA/Cetak", "href": "/asesmen/dokumen", "tone": "soft"}]}];
</script>

<svelte:head>
	<title>Jadwal Sesi</title>
	<meta name="description" content="Fitur sesi: jadwal, durasi, ruang target, paket aktif, status buka/tutup, daftar hadir, berita acara, dan token kontrol hari-H." />
</svelte:head>

<CbtFeatureSurface
	title="Jadwal Sesi"
	eyebrow="Pelaksanaan"
	description="Fitur sesi: jadwal, durasi, ruang target, paket aktif, status buka/tutup, daftar hadir, berita acara, dan token kontrol hari-H."
	{actions}
	{stats}
	{steps}
	{features}
/>
