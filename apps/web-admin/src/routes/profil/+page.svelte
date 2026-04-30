<script lang="ts">
	import PublicContentPage from '$lib/components/public/PublicContentPage.svelte';
	type PageContent = {
		title: string;
		excerpt: string;
		content_html: string;
		cover_image_url: string;
		published_at: string | null;
	};
	let { data }: { data: { page: PageContent } } = $props();
</script>

<svelte:head><title>{data.page.title} — MTsN 2 Kolaka Utara</title></svelte:head>

<PublicContentPage
	content={data.page}
	eyebrow="Profil Madrasah"
	sidePanels={[
		{
			eyebrow: 'Tentang Madrasah',
			title: 'Gambaran Singkat',
			lines: [
				'Halaman ini memuat profil resmi MTsN 2 Kolaka Utara.',
				'Gunakan informasi ini untuk mengenal arah, identitas, dan karakter madrasah.',
				'Untuk kebutuhan pendaftaran, lanjutkan ke halaman PPDB resmi.',
			],
		},
		{
			eyebrow: 'Jalur Lanjut',
			title: 'Informasi Terkait',
			tone: 'emerald',
			lines: [
				'Lihat halaman berita untuk kegiatan dan capaian terbaru madrasah.',
				'Gunakan halaman kontak resmi untuk komunikasi lanjutan dengan sekolah.',
				'PPDB tersedia bagi calon siswa yang ingin memulai proses pendaftaran.',
			],
		},
	]}
/>
