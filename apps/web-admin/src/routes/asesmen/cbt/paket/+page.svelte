<script lang="ts">
	import CbtFeatureSurface from '../CbtFeatureSurface.svelte';

	const actions = [{"label": "Mapping paket", "href": "/asesmen", "tone": "primary"}, {"label": "Bank Soal", "href": "/bank-soal", "tone": "soft"}];
	const stats = [{"label": "Mapping", "value": "Per rombel", "detail": "class + subject"}, {"label": "Durasi", "value": "Dari paket", "detail": "Menit ujian"}, {"label": "Validasi", "value": "Jumlah soal", "detail": "Cek sebelum sesi"}, {"label": "Kunci", "value": "Opsional", "detail": "Saat siap ujian"}];
	const steps = [{"label": "Pilih mapel", "detail": "Ambil mapel dari Akademik/Super App.", "badge": "1"}, {"label": "Pilih paket", "detail": "Ambil paket Bank Soal yang siap digunakan.", "badge": "2"}, {"label": "Mapping rombel", "detail": "Pasangkan paket ke kelas/rombel target.", "badge": "3"}, {"label": "Validasi sesi", "detail": "Pastikan paket, jadwal, ruang, dan peserta lengkap.", "badge": "4"}];
	const features = [{"title": "Package options", "detail": "Endpoint /api/asesmen/packages/options dipakai di Command Center untuk menampilkan paket siap pakai.", "status": "backend", "actions": [{"label": "Buka mapping", "href": "/asesmen", "tone": "primary"}]}, {"title": "Pemetaan paket", "detail": "PUT package-maps menyimpan paket per rombel/mapel agar sesi tidak manual berulang.", "status": "backend", "actions": []}, {"title": "Validasi readiness", "detail": "Operator bisa melihat kelas belum punya paket dan paket kurang soal sebelum ujian dibuka.", "status": "fitur", "actions": []}, {"title": "Jembatan Bank Soal", "detail": "Paket tetap mengambil data dari Bank Soal Super App, bukan data legacy SQLite.", "status": "aman", "actions": [{"label": "Daftar Soal", "href": "/bank-soal", "tone": "soft"}]}];
</script>

<svelte:head>
	<title>Paket Ujian</title>
	<meta name="description" content="Fitur paket: pilih paket dari Bank Soal, mapping ke rombel/mapel, validasi jumlah soal, kunci paket, lalu jadwalkan sesi." />
</svelte:head>

<CbtFeatureSurface
	title="Paket Ujian"
	eyebrow="Manajemen Soal"
	description="Fitur paket: pilih paket dari Bank Soal, mapping ke rombel/mapel, validasi jumlah soal, kunci paket, lalu jadwalkan sesi."
	{actions}
	{stats}
	{steps}
	{features}
/>
