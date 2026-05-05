<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';

	const userRoles = $derived(page.data.user?.roles ?? (page.data.user?.role ? [page.data.user.role] : []));
	const canAccess = $derived(userRoles.includes('admin'));

	type KegiatanRoute = '/cbt/events' | '/cbt/sessions';

	type KegiatanCard = {
		title: string;
		description: string;
		href: KegiatanRoute;
		label: string;
		tone: 'operation' | 'session';
	};

	const cards: KegiatanCard[] = [
		{
			title: 'Daftar Kegiatan',
			label: 'Event',
			description: 'Lihat dan kelola kegiatan ujian. Setiap kegiatan menampung satu atau lebih sesi dengan paket soal yang sudah ditentukan.',
			href: '/cbt/events',
			tone: 'operation'
		},
		{
			title: 'Daftar Sesi',
			label: 'Sesi',
			description: 'Atur jadwal, peserta, ruang, token, dan kesiapan operasional untuk setiap sesi ujian dalam satu kegiatan.',
			href: '/cbt/sessions',
			tone: 'session'
		}
	];

	function toneClass(tone: KegiatanCard['tone']): string {
		switch (tone) {
			case 'operation':
				return 'border-teal-200 bg-teal-50/70 text-teal-800';
			case 'session':
				return 'border-sky-200 bg-sky-50/70 text-sky-800';
		}
	}
</script>

<svelte:head>
	<title>Kegiatan & Sesi — MTsN 2 Kolaka Utara</title>
</svelte:head>

{#if canAccess}
	<div class="space-y-6">
	<section class="rounded-3xl border border-emerald-100 bg-gradient-to-br from-emerald-50 via-white to-lime-50 p-6 shadow-sm">
		<div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
			<div class="max-w-3xl space-y-3">
				<p class="text-xs font-semibold uppercase tracking-[0.24em] text-emerald-700">CBT · Kegiatan & Sesi</p>
				<h1 class="text-3xl font-semibold tracking-tight text-slate-900">Kegiatan & Sesi</h1>
				<p class="max-w-2xl text-sm leading-6 text-slate-600">
					Atur kegiatan ujian dan sesi pelaksanaannya. Kegiatan membungkus rangkaian sesi, sementara sesi menentukan
					jadwal, peserta, ruang, dan token ujian.
				</p>
			</div>
			<div class="flex flex-wrap gap-3">
				<Button href={resolve('/cbt')} variant="outline">Beranda CBT</Button>
				<Button href={resolve('/cbt/events')}>Daftar Kegiatan</Button>
				<Button href={resolve('/cbt/sessions')} variant="outline">Daftar Sesi</Button>
			</div>
		</div>
	</section>

	<section class="grid gap-4 lg:grid-cols-2" aria-label="Pilihan utama Kegiatan & Sesi">
		{#each cards as card (card.title)}
			<Card.Root class="flex h-full flex-col border-slate-200 bg-white shadow-sm transition hover:-translate-y-0.5 hover:border-emerald-200 hover:shadow-md">
				<Card.Header class="space-y-3">
					<Badge class={toneClass(card.tone)} variant="outline">{card.label}</Badge>
					<div>
						<Card.Title class="text-lg text-slate-900">{card.title}</Card.Title>
						<Card.Description class="mt-2 leading-6">{card.description}</Card.Description>
					</div>
				</Card.Header>
				<Card.Footer>
					<Button href={resolve(card.href)} variant="outline" class="border-emerald-200 text-emerald-800 hover:bg-emerald-50">Buka</Button>
				</Card.Footer>
			</Card.Root>
		{/each}
	</section>

	<Card.Root class="border-emerald-200 bg-emerald-50/60 shadow-sm">
		<Card.Header>
			<Card.Title class="text-lg text-slate-900">Alur Kegiatan & Sesi</Card.Title>
			<Card.Description>
				Buat kegiatan terlebih dahulu, tetapkan paket soal, lalu tambahkan sesi dengan jadwal, peserta, dan ruang.
				Token ujian diterbitkan dari sesi setelah seluruh konfigurasi dianggap siap.
			</Card.Description>
		</Card.Header>
	</Card.Root>
	</div>
{:else}
	<div class="flex min-h-[60vh] flex-col items-center justify-center gap-4 px-4 py-16 text-center">
		<div class="rounded-2xl border border-slate-200 bg-white p-8 shadow-sm max-w-lg">
			<h2 class="text-xl font-semibold text-slate-900">Akses terbatas</h2>
			<p class="mt-3 text-sm leading-6 text-slate-600">
				Halaman Kegiatan & Sesi hanya tersedia untuk admin. Silakan hubungi administrator jika Anda memerlukan akses.
			</p>
			<div class="mt-6">
				<a href={resolve('/cbt')} class="inline-flex rounded-md border border-emerald-200 bg-white px-4 py-2 text-sm font-semibold text-emerald-800 transition hover:bg-emerald-50">Kembali ke Beranda CBT</a>
			</div>
		</div>
	</div>
{/if}
