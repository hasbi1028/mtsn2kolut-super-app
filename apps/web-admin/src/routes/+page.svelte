<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';

	let { data }: {
		data: {
			user?: { username?: string; role?: string; roles?: string[] };
		};
	} = $props();

	const roles = $derived(data.user?.roles || (data.user?.role ? [data.user.role] : []));
	const isAdmin = $derived(roles.includes('admin'));

	let now = $state(new Date());

	$effect(() => {
		const interval = setInterval(() => { now = new Date(); }, 60000);
		return () => clearInterval(interval);
	});

	function hariIni(): string {
		return now.toLocaleDateString('id-ID', {
			weekday: 'long', year: 'numeric', month: 'long', day: 'numeric',
			timeZone: 'Asia/Makassar'
		});
	}

	function jamSekarang(): string {
		return now.toLocaleTimeString('id-ID', {
			hour: '2-digit', minute: '2-digit', timeZone: 'Asia/Makassar'
		});
	}

	const adminActions = [
		{ label: 'Monitor Kehadiran', desc: 'Pantau absensi pegawai', href: '/pusaka', icon: 'clock' },
		{ label: 'Data Pegawai', desc: 'Kelola daftar pegawai', href: '/employees', icon: 'users' },
		{ label: 'Manajemen Pengguna', desc: 'Kelola akun & hak akses', href: '/settings/users', icon: 'shield' },
		{ label: 'Profil Madrasah', desc: 'Pengaturan profil', href: '/settings/school-profile', icon: 'building' },
		{ label: 'Audit Log', desc: 'Jejak aktivitas sistem', href: '/settings/audit', icon: 'file-text' },
		{ label: 'Backup Data', desc: 'Cadangkan data sistem', href: '/settings/backups', icon: 'database' },
	];

	const userActions = [
		{ label: 'Kehadiran', desc: 'Lihat catatan kehadiran', href: '/pusaka', icon: 'clock' },
		{ label: 'Akun Saya', desc: 'Pengaturan akun pribadi', href: '/settings/account', icon: 'user' },
	];

	const actions = $derived(isAdmin ? adminActions : userActions);

	const iconMap: Record<string, { viewBox: string; path: string }> = {
		clock: { viewBox: '0 0 24 24', path: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z' },
		users: { viewBox: '0 0 24 24', path: 'M12 4.354a4 4 0 110 7.292 4 4 0 010-7.292zM15 21H9a2 2 0 01-2-2V12a2 2 0 012-2h6a2 2 0 012 2v7a2 2 0 01-2 2z' },
		shield: { viewBox: '0 0 24 24', path: 'M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z' },
		building: { viewBox: '0 0 24 24', path: 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0H5m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4' },
		'file-text': { viewBox: '0 0 24 24', path: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z' },
		database: { viewBox: '0 0 24 24', path: 'M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4' },
		user: { viewBox: '0 0 24 24', path: 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z' }
	};
</script>

<svelte:head>
	<title>Beranda — MTsN 2 Kolut</title>
</svelte:head>

{#if !data.user}
	<!-- Logged out: minimal login prompt -->
	<div class="flex min-h-[60vh] items-center justify-center p-4">
		<div class="w-full max-w-sm text-center space-y-5">
			<div class="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-primary text-primary-foreground text-xl font-black shadow-lg shadow-primary/25">
				M
			</div>
			<h1 class="text-xl font-black tracking-tight text-foreground">
				MTs Negeri 2 Kolaka Utara
			</h1>
			<p class="text-sm font-medium text-muted-foreground">
				Sistem Manajemen Madrasah
			</p>
			<a
				href={resolve('/login')}
				class="inline-flex items-center gap-2 rounded-xl bg-primary px-5 py-2.5 text-sm font-bold text-primary-foreground transition-all hover:bg-primary/90 hover:shadow-lg hover:shadow-primary/25"
			>
				Masuk ke Sistem
			</a>
		</div>
	</div>
{:else}
	<!-- ═══ CBT-style Dashboard ═══ -->
	<div class="space-y-8">

		<!-- Page Title (matches CBT admin header) -->
		<div>
			<h1 class="text-2xl font-black tracking-tight text-foreground">Dashboard</h1>
			<p class="text-sm text-muted-foreground mt-1">
				Selamat datang, <span class="font-semibold text-foreground">{data.user?.username || 'Pengguna'}</span>
			</p>
		</div>

		<!-- Stat Cards (CBT style: compact, bordered, white cards) -->
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
			<div class="rounded-xl border border-border bg-card p-5 shadow-sm">
				<div class="flex items-center gap-3">
					<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10 text-primary">
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
						</svg>
					</div>
					<div>
						<div class="text-[11px] font-bold tracking-wider text-muted-foreground uppercase">Role</div>
						<div class="text-lg font-black text-foreground">{isAdmin ? 'Administrator' : 'Pengguna'}</div>
					</div>
				</div>
			</div>
			<div class="rounded-xl border border-border bg-card p-5 shadow-sm">
				<div class="flex items-center gap-3">
					<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-emerald-50 text-emerald-600">
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
					<div>
						<div class="text-[11px] font-bold tracking-wider text-muted-foreground uppercase">Status</div>
						<div class="text-lg font-black text-emerald-600">Aktif</div>
					</div>
				</div>
			</div>
			<div class="rounded-xl border border-border bg-card p-5 shadow-sm">
				<div class="flex items-center gap-3">
					<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-sky-50 text-sky-600">
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0H5m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
						</svg>
					</div>
					<div>
						<div class="text-[11px] font-bold tracking-wider text-muted-foreground uppercase">Instansi</div>
						<div class="text-lg font-black text-foreground truncate">MTsN 2 Kolut</div>
					</div>
				</div>
			</div>
			<div class="rounded-xl border border-border bg-card p-5 shadow-sm">
				<div class="flex items-center gap-3">
					<div class="flex h-10 w-10 items-center justify-center rounded-lg bg-amber-50 text-amber-600">
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
					<div>
						<div class="text-[11px] font-bold tracking-wider text-muted-foreground uppercase">{hariIni()}</div>
						<div class="text-lg font-black text-foreground">{jamSekarang()} WITA</div>
					</div>
				</div>
			</div>
		</div>

		<!-- Quick Access / Modul Grid (CBT card-table style) -->
		<div class="rounded-xl border border-border bg-card shadow-sm overflow-hidden">
			<div class="px-5 py-4 border-b border-border bg-muted/30">
				<h2 class="text-sm font-bold text-foreground">Modul Aplikasi</h2>
			</div>
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
				{#each actions as action, i (action.href)}
					{@const icon = iconMap[action.icon]}
					<a
						href={resolve(action.href as '/')}
						class="group flex items-center gap-4 p-5 text-foreground transition-all hover:bg-muted/50 border-r border-b border-border"
					>
						<div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-muted/60 text-muted-foreground group-hover:text-primary transition-colors">
							{#if icon}
								<svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="1.8" viewBox={icon.viewBox}>
									<path stroke-linecap="round" stroke-linejoin="round" d={icon.path} />
								</svg>
							{/if}
						</div>
						<div class="min-w-0 flex-1">
							<div class="text-sm font-semibold text-foreground">{action.label}</div>
							<div class="text-xs text-muted-foreground mt-0.5">{action.desc}</div>
						</div>
						<svg class="h-4 w-4 shrink-0 text-muted-foreground/40 group-hover:text-primary transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
						</svg>
					</a>
				{/each}
			</div>
		</div>

		<!-- Info Bar (CBT footer-style) -->
		<div class="rounded-xl border border-border bg-card px-5 py-3 shadow-sm">
			<div class="flex items-center gap-3">
				<div class="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
					<svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
				</div>
				<div class="min-w-0 flex-1">
					<p class="text-xs font-bold text-foreground">Sistem Manajemen Madrasah</p>
					<p class="text-[11px] font-medium text-muted-foreground">MTsN 2 Kolaka Utara &middot; WITA &middot; by Hasbi Awal</p>
				</div>
			</div>
		</div>
	</div>
{/if}