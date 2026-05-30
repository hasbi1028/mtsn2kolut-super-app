<script lang="ts">
	import CbtFeatureSurface from '../CbtFeatureSurface.svelte';

	const actions = [{"label": "Auto preview/simpan", "href": "/asesmen", "tone": "primary"}, {"label": "Portal Pengawas", "href": "/pengawas-ujian?demo=1", "tone": "soft"}];
	const stats = [{"label": "Ruang default", "value": "8", "detail": "R01–R08"}, {"label": "Kapasitas", "value": "30/ruang", "detail": "Dapat disesuaikan"}, {"label": "Mode", "value": "Campur", "detail": "Mixed rombel/kelas"}, {"label": "Manual", "value": "Aktif", "detail": "Pindah kursi satu peserta"}];
	const steps = [{"label": "Preview", "detail": "Simulasi pembagian tanpa menyimpan.", "badge": "Aman"}, {"label": "Simpan", "detail": "Apply assignment ke peserta/ruang.", "badge": "Backend"}, {"label": "Edit Manual", "detail": "Pindahkan peserta jika panitia perlu penyesuaian.", "badge": "Override"}, {"label": "Cetak Denah", "detail": "Cetak denah dan lembar pengawas setelah kursi final.", "badge": "Dokumen"}];
	const features = [{"title": "Auto placement mixed", "detail": "Mengikuti pola gladi: peserta dibagi round-robin antar rombel agar tiap ruang campur dan kapasitas terjaga.", "status": "aktif", "actions": [{"label": "Buka panel", "href": "/asesmen", "tone": "primary"}]}, {"title": "Validasi kapasitas", "detail": "Operator melihat total peserta, kapasitas total, sisa/tidak tertampung, dan ringkasan per ruang.", "status": "aktif", "actions": []}, {"title": "Denah ruang", "detail": "Menyiapkan tampilan kursi per ruang untuk dicetak atau dicek pengawas sebelum ujian.", "status": "fitur", "actions": [{"label": "Dokumen", "href": "/asesmen/cbt/dokumen", "tone": "soft"}]}, {"title": "Monitor ruang", "detail": "Pintasan ke portal pengawas dan proctoring untuk cek peserta per ruang.", "status": "hari-H", "actions": [{"label": "Proctoring", "href": "/asesmen/cbt/proctoring", "tone": "soft"}]}];
</script>

<svelte:head>
	<title>Ruang Ujian</title>
	<meta name="description" content="Fitur dalam ruang: auto-bagi peserta ke R01–R08, validasi kapasitas, denah kursi, token/lembar pengawas, monitor ruang, dan pindah kursi manual." />
</svelte:head>

<CbtFeatureSurface
	title="Ruang Ujian"
	eyebrow="Pusat Data"
	description="Fitur dalam ruang: auto-bagi peserta ke R01–R08, validasi kapasitas, denah kursi, token/lembar pengawas, monitor ruang, dan pindah kursi manual."
	{actions}
	{stats}
	{steps}
	{features}
/>
