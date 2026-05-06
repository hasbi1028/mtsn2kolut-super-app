<script lang="ts">
	import * as Card from '$lib/components/ui/card';
</script>

<svelte:head><title>Website Publik — MTSN 2 Kolaka Utara</title></svelte:head>

<div class="space-y-6">
	<div>
		<h1 class="text-2xl font-semibold text-foreground">Website Publik Sekolah</h1>
		<p class="mt-1 text-sm text-muted-foreground">Kelola wajah publik madrasah: berita, pengumuman, halaman profil, dan informasi PPDB.</p>
	</div>

	<div class="grid gap-4 md:grid-cols-3">
		<a href="/website/posts">
			<Card.Root class="h-full border-border transition hover:border-primary/20 hover:shadow-sm">
				<Card.Header>
					<Card.Title>Berita</Card.Title>
					<Card.Description>Artikel kegiatan, prestasi, dan kabar terbaru madrasah.</Card.Description>
				</Card.Header>
			</Card.Root>
		</a>
		<a href="/website/announcements">
			<Card.Root class="h-full border-border transition hover:border-primary/20 hover:shadow-sm">
				<Card.Header>
					<Card.Title>Pengumuman</Card.Title>
					<Card.Description>Informasi resmi seperti jadwal, hasil seleksi, dan pemberitahuan penting.</Card.Description>
				</Card.Header>
			</Card.Root>
		</a>
		<a href="/website/pages">
			<Card.Root class="h-full border-border transition hover:border-primary/20 hover:shadow-sm">
				<Card.Header>
					<Card.Title>Halaman Publik</Card.Title>
					<Card.Description>Profil, kontak, dan informasi PPDB yang tampil di website sekolah.</Card.Description>
				</Card.Header>
			</Card.Root>
		</a>
	</div>
</div>
