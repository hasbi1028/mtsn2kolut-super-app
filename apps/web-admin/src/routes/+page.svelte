<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import PublicHome from '$lib/components/PublicHome.svelte';
	import { readClientApiData } from '$lib/client/api';

	type WebsiteContent = {
		id: string;
		title: string;
		slug: string;
		excerpt: string;
		content_html: string;
		cover_image_url: string;
		published_at: string | null;
	};

	let { data }: {
		data: {
			user?: { name?: string; role?: string; roles?: string[]; permissions?: string[] };
			publicHome?: { posts: WebsiteContent[]; featuredPosts: WebsiteContent[]; announcements: WebsiteContent[]; profil: WebsiteContent | null; ppdbInfo: WebsiteContent | null };
		};
	} = $props();

	const roles = $derived(data.user?.roles || (data.user?.role ? [data.user.role] : []));
	const isAdmin = $derived(roles.includes('admin'));
	const namaUser = $derived(data.user?.name || 'Pengguna');

	let now = $state(new Date());

	$effect(() => {
		const interval = setInterval(() => {
			now = new Date();
		}, 60000);
		return () => clearInterval(interval);
	});

	function sapaan(): string {
		const jam = now.getHours();
		if (jam < 11) return 'Selamat pagi';
		if (jam < 15) return 'Selamat siang';
		if (jam < 18) return 'Selamat sore';
		return 'Selamat malam';
	}

	function hariIni(): string {
		return now.toLocaleDateString('id-ID', {
			weekday: 'long', year: 'numeric', month: 'long', day: 'numeric',
			timeZone: 'Asia/Makassar'
		});
	}

	const aksiCepat = $derived.by(() => {
		const items: Array<{ label: string; href: string; icon: string }> = [];
		if (isAdmin) {
			items.push(
				{ label: 'Monitor Kehadiran', href: '/pusaka', icon: '📋' },
				{ label: 'Daftar Pegawai', href: '/employees', icon: '👥' },
				{ label: 'Backup Data', href: '/settings/backups', icon: '💾' },
				{ label: 'Profil Madrasah', href: '/settings/school-profile', icon: '🏫' },
			);
		} else {
			items.push(
				{ label: 'Kehadiran Saya', href: '/pusaka', icon: '📋' },
				{ label: 'Akun Saya', href: '/settings/account', icon: '⚙️' },
			);
		}
		return items;
	});
</script>

<svelte:head>
	<title>{data.user ? 'Beranda — MTsN 2 Kolut' : 'MTs Negeri 2 Kolaka Utara'}</title>
</svelte:head>

{#if !data.user}
	<PublicHome home={data.publicHome} />
{:else}
	<div class="space-y-6">
		<!-- Sapaan -->
		<div class="rounded-xl border border-[var(--border)] bg-[var(--card)] p-6">
			<h1 class="text-2xl font-bold text-[var(--foreground)]">
				{sapaan()}, {namaUser} 👋
			</h1>
			<p class="mt-1 text-[var(--muted-foreground)]">
				{hariIni()}
			</p>
		</div>

		<!-- Aksi Cepat -->
		<div>
			<h2 class="mb-3 text-sm font-semibold uppercase tracking-wider text-[var(--muted-foreground)]">Aksi Cepat</h2>
			<div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
				{#each aksiCepat as aksi (aksi.href)}
					<a href={aksi.href} class="flex items-center gap-3 rounded-xl border border-[var(--border)] bg-[var(--card)] p-4 transition-colors hover:border-[var(--primary)] hover:bg-[var(--primary)]/5">
						<span class="text-2xl">{aksi.icon}</span>
						<span class="text-sm font-medium text-[var(--foreground)]">{aksi.label}</span>
					</a>
				{/each}
			</div>
		</div>

		<!-- Info -->
		<div class="rounded-xl border border-[var(--border)] bg-[var(--card)] p-5">
			<h2 class="text-sm font-semibold uppercase tracking-wider text-[var(--muted-foreground)]">Tentang Sistem</h2>
			<p class="mt-2 text-sm leading-relaxed text-[var(--muted-foreground)]">
				Sistem Manajemen Madrasah — MTsN 2 Kolaka Utara. Kelola kehadiran, data pegawai, dan pengaturan dari satu tempat.
			</p>
			<div class="mt-4 flex flex-wrap gap-2">
				<span class="rounded-full bg-[var(--primary)]/10 px-3 py-1 text-xs font-medium text-[var(--primary)]">WITA</span>
				<span class="rounded-full bg-[var(--primary)]/10 px-3 py-1 text-xs font-medium text-[var(--primary)]">Kemenag</span>
				{#if isAdmin}
					<span class="rounded-full bg-[var(--primary)]/10 px-3 py-1 text-xs font-medium text-[var(--primary)]">Admin</span>
				{/if}
			</div>
		</div>
	</div>
{/if}
