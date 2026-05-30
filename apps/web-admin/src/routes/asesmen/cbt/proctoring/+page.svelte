<script lang="ts">
	import CbtFeatureSurface from '../CbtFeatureSurface.svelte';

	const actions = [{"label": "Portal Pengawas", "href": "/pengawas-ujian?demo=1", "tone": "primary"}, {"label": "Pelaksanaan", "href": "/asesmen/pelaksanaan", "tone": "soft"}];
	const stats = [{"label": "Tab", "value": "Ruang/Peringatan/Peserta", "detail": "Portal pengawas"}, {"label": "Alert", "value": "Violation", "detail": "Keluar halaman/koneksi"}, {"label": "Aksi aman", "value": "3", "detail": "Cek, peringatkan, hubungi admin"}, {"label": "Admin", "value": "Terpisah", "detail": "Unlock/reset tidak di portal"}];
	const steps = [{"label": "Pilih ruang", "detail": "Pengawas membuka ruang yang dijaga.", "badge": "Ruang"}, {"label": "Pantau peserta", "detail": "Lihat siap masuk, mengerjakan, selesai, offline.", "badge": "Live"}, {"label": "Tindak alert", "detail": "Sudah Dicek atau Beri Peringatan.", "badge": "Aksi"}, {"label": "Eskalasi", "detail": "Hubungi Admin untuk reset/unlock sensitif.", "badge": "Aman"}];
	const features = [{"title": "Status peserta", "detail": "Menampilkan Siap Masuk, Mengerjakan, Offline, Minta Selesai, dan Selesai.", "status": "fitur", "actions": [{"label": "Portal", "href": "/pengawas-ujian?demo=1", "tone": "primary"}]}, {"title": "Peringatan anti-cheat", "detail": "Event seperti keluar halaman, shortcut terlarang, copy/paste, dan koneksi buruk masuk daftar peringatan.", "status": "fitur", "actions": []}, {"title": "Antrian bantuan admin", "detail": "Pengawas tidak diberi tombol reset massal; hanya mengirim bantuan ke admin/panitia.", "status": "aman", "actions": []}, {"title": "Monitor admin", "detail": "Panitia tetap punya akses mode lengkap di /asesmen/pelaksanaan bila perlu.", "status": "mode lengkap", "actions": [{"label": "Pelaksanaan", "href": "/asesmen/pelaksanaan", "tone": "soft"}]}];
</script>

<svelte:head>
	<title>Proctoring Live</title>
	<meta name="description" content="Fitur proctoring: status peserta, filter ruang, peringatan, antrian bantuan admin, aksi aman pengawas, dan eskalasi tanpa membuka kontrol sensitif." />
</svelte:head>

<CbtFeatureSurface
	title="Proctoring Live"
	eyebrow="Pelaksanaan"
	description="Fitur proctoring: status peserta, filter ruang, peringatan, antrian bantuan admin, aksi aman pengawas, dan eskalasi tanpa membuka kontrol sensitif."
	{actions}
	{stats}
	{steps}
	{features}
/>
