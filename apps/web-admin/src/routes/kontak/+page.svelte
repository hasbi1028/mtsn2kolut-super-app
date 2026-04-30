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
	eyebrow="Kontak Resmi"
	sidePanels={[
		{
			eyebrow: 'Panduan Kontak',
			title: 'Sebelum Menghubungi',
			lines: [
				'Siapkan nama lengkap dan keperluan utama agar respons lebih cepat.',
				'Jika terkait pendaftaran, sertakan identitas calon siswa atau wali.',
				'Gunakan saluran resmi sekolah untuk pertanyaan administratif dan akademik.',
			],
		},
		{
			eyebrow: 'Informasi Tambahan',
			title: 'Rujukan Cepat',
			tone: 'emerald',
			lines: [
				'Pengumuman terbaru tersedia pada halaman pengumuman resmi.',
				'Kegiatan madrasah dan capaian siswa dapat dilihat di halaman berita.',
				'PPDB awal dapat diisi langsung melalui halaman pendaftaran publik.',
			],
		},
	]}
/>
