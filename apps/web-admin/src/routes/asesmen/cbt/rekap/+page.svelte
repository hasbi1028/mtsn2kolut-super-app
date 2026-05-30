<script lang="ts">
	import CbtFeatureSurface from '../CbtFeatureSurface.svelte';

	const actions = [{"label": "Buka hasil", "href": "/asesmen/hasil", "tone": "primary"}, {"label": "Analisis Bank Soal", "href": "/bank-soal/analisis-butir", "tone": "soft"}];
	const stats = [{"label": "PG", "value": "Otomatis", "detail": "Skor langsung"}, {"label": "Essay", "value": "Manual", "detail": "Perlu koreksi"}, {"label": "Export", "value": "CSV/PDF", "detail": "Arsip"}, {"label": "Analisis", "value": "Butir", "detail": "Kualitas soal"}];
	const steps = [{"label": "Kumpulkan hasil", "detail": "Cek peserta selesai dan jawaban tersimpan.", "badge": "1"}, {"label": "Koreksi essay", "detail": "Guru/admin menilai jawaban uraian jika ada.", "badge": "2"}, {"label": "Export rekap", "detail": "Unduh CSV/PDF untuk arsip madrasah.", "badge": "3"}, {"label": "Analisis", "detail": "Review butir sulit/mudah dan tindak lanjut.", "badge": "4"}];
	const features = [{"title": "Rekap per peserta", "detail": "Menampilkan nilai, status selesai, ruang, dan catatan pelaksanaan.", "status": "fitur", "actions": [{"label": "Hasil", "href": "/asesmen/hasil", "tone": "primary"}]}, {"title": "Koreksi essay", "detail": "Jika paket berisi essay, koreksi dilakukan sebelum hasil final.", "status": "fitur", "actions": []}, {"title": "Export hasil", "detail": "Rekap nilai dan arsip hasil disiapkan untuk CSV/PDF.", "status": "arsip", "actions": []}, {"title": "Analisis butir", "detail": "Arahkan ke analisis Bank Soal agar kualitas soal tetap dievaluasi.", "status": "analisis", "actions": [{"label": "Analisis", "href": "/bank-soal/analisis-butir", "tone": "soft"}]}];
</script>

<svelte:head>
	<title>Rekap Nilai</title>
	<meta name="description" content="Fitur rekap: nilai per sesi, status submit, koreksi essay, export CSV/PDF, analisis butir, dan arsip hasil." />
</svelte:head>

<CbtFeatureSurface
	title="Rekap Nilai"
	eyebrow="Laporan"
	description="Fitur rekap: nilai per sesi, status submit, koreksi essay, export CSV/PDF, analisis butir, dan arsip hasil."
	{actions}
	{stats}
	{steps}
	{features}
/>
