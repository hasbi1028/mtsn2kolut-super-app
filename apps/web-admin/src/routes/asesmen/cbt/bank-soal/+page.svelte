<script lang="ts">
	import CbtFeatureSurface from '../CbtFeatureSurface.svelte';

	const actions = [{"label": "Daftar Soal", "href": "/bank-soal", "tone": "primary"}, {"label": "Tambah Soal", "href": "/bank-soal/tambah", "tone": "soft"}, {"label": "Verifikasi", "href": "/bank-soal/verifikasi", "tone": "soft"}];
	const stats = [{"label": "Sumber", "value": "Super App", "detail": "Bukan SQLite legacy"}, {"label": "Review", "value": "Layak/Revisi", "detail": "Workflow resmi"}, {"label": "Import", "value": "CSV", "detail": "Format Super App"}, {"label": "Paket", "value": "Builder", "detail": "Dipakai sesi CBT"}];
	const steps = [{"label": "Tulis/Import", "detail": "Guru/admin membuat soal atau import CSV.", "badge": "Input"}, {"label": "Review", "detail": "Admin menandai Layak, Revisi, atau Tolak.", "badge": "QC"}, {"label": "Publish", "detail": "Soal siap dipakai paket ujian.", "badge": "Rilis"}, {"label": "Paket", "detail": "Paket ujian mengambil pool dari Bank Soal.", "badge": "CBT"}];
	const features = [{"title": "Daftar & filter soal", "detail": "Menggunakan daftar Bank Soal Super App dengan filter mapel, pembuat, status, dan workflow.", "status": "aktif", "actions": [{"label": "Buka daftar", "href": "/bank-soal", "tone": "primary"}]}, {"title": "Tambah/import soal", "detail": "Pintasan tambah manual dan import CSV; metadata AI tetap mengikuti aturan Super App.", "status": "aktif", "actions": [{"label": "Tambah", "href": "/bank-soal/tambah", "tone": "soft"}, {"label": "Import", "href": "/bank-soal/impor", "tone": "soft"}]}, {"title": "Verifikasi soal", "detail": "Workflow Layak/Revisi/Tolak tetap di modul Super App agar tidak menurunkan kualitas Bank Soal.", "status": "aktif", "actions": [{"label": "Verifikasi", "href": "/bank-soal/verifikasi", "tone": "soft"}]}, {"title": "Cetak & analisis", "detail": "Cetak soal dan analisis butir diarahkan ke fitur Super App yang sudah ada.", "status": "aktif", "actions": [{"label": "Cetak", "href": "/bank-soal/cetak", "tone": "soft"}, {"label": "Analisis", "href": "/bank-soal/analisis-butir", "tone": "soft"}]}];
</script>

<svelte:head>
	<title>Bank Soal</title>
	<meta name="description" content="Bank Soal legacy tidak dicopy mentah. Fitur dalamnya dipetakan ke Bank Soal Super App: daftar, tambah, import CSV, verifikasi Layak/Revisi/Tolak, cetak, dan paket builder." />
</svelte:head>

<CbtFeatureSurface
	title="Bank Soal"
	eyebrow="Manajemen Soal"
	description="Bank Soal legacy tidak dicopy mentah. Fitur dalamnya dipetakan ke Bank Soal Super App: daftar, tambah, import CSV, verifikasi Layak/Revisi/Tolak, cetak, dan paket builder."
	{actions}
	{stats}
	{steps}
	{features}
/>
