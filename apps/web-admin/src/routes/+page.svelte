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
			<div class="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-primary text-primary-content text-xl font-black shadow-lg shadow-primary/25">
				M
			</div>
			<h1 class="text-xl font-black tracking-tight text-base-content">
				MTs Negeri 2 Kolaka Utara
			</h1>
			<p class="text-sm font-medium text-base-content/70">
				Sistem Manajemen Madrasah
			</p>
			<a
				href={resolve('/login')}
				class="inline-flex items-center gap-2 rounded-xl bg-primary px-5 py-2.5 text-sm font-bold text-primary-content transition-all hover:bg-primary/90 hover:shadow-lg hover:shadow-primary/25"
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
			<h1 class="text-2xl font-black tracking-tight text-base-content">Dashboard</h1>
			<p class="text-sm text-base-content/70 mt-1">
				Selamat datang, <span class="font-semibold text-base-content">{data.user?.username || 'Pengguna'}</span>
			</p>
		</div>

		<!-- Stat Cards (daisyui card + badge, compact & informative) -->
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
			<div class="card bg-base-100 border border-base-300 shadow-sm">
				<div class="card-body p-4 gap-1">
					<div class="flex items-center gap-2">
						<div class="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10 text-primary">
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
							</svg>
						</div>
						<div class="badge badge-sm badge-outline">{isAdmin ? 'Admin' : 'User'}</div>
					</div>
					<p class="card-title text-lg font-black mt-1">{data.user?.username || 'Pengguna'}</p>
					<p class="text-xs text-base-content/70">{isAdmin ? 'Administrator' : 'Pengguna'}</p>
				</div>
			</div>
			<div class="card bg-base-100 border border-base-300 shadow-sm">
				<div class="card-body p-4 gap-1">
					<div class="flex items-center gap-2">
						<div class="flex h-8 w-8 items-center justify-center rounded-lg bg-emerald-50 text-emerald-600">
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
							</svg>
						</div>
						<div class="badge badge-sm badge-success">Aktif</div>
					</div>
					<p class="card-title text-lg font-black mt-1">{hariIni()}</p>
					<p class="text-xs text-base-content/70">{jamSekarang()} WITA</p>
				</div>
			</div>
			<div class="card bg-base-100 border border-base-300 shadow-sm">
				<div class="card-body p-4 gap-1">
					<div class="flex items-center gap-2">
						<div class="flex h-8 w-8 items-center justify-center rounded-lg bg-sky-50 text-sky-600">
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0H5m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
							</svg>
						</div>
						<div class="badge badge-sm badge-info">Instansi</div>
					</div>
					<p class="card-title text-lg font-black mt-1 truncate">MTsN 2 Kolut</p>
					<p class="text-xs text-base-content/70">MTs Negeri 2 Kolaka Utara</p>
				</div>
			</div>
			<div class="card bg-base-100 border border-base-300 shadow-sm">
				<div class="card-body p-4 gap-1">
					<div class="flex items-center gap-2">
						<div class="flex h-8 w-8 items-center justify-center rounded-lg bg-amber-50 text-amber-600">
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
							</svg>
						</div>
						<div class="badge badge-sm badge-warning">Sistem</div>
					</div>
					<p class="card-title text-lg font-black mt-1">SMM v2.0</p>
					<p class="text-xs text-base-content/70">Sistem Manajemen Madrasah</p>
				</div>
			</div>
		</div>

		<!-- Quick Access / Modul Grid (daisyui card style) -->
		<div class="card bg-base-100 border border-base-300 shadow-sm">
			<div class="card-body p-0">
				<h2 class="card-title px-5 py-4 border-b border-base-300 bg-base-200/30 text-sm m-0 rounded-t-box">Modul Aplikasi</h2>
				<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
					{#each actions as action, i (action.href)}
						{@const icon = iconMap[action.icon]}
						<a
							href={resolve(action.href as '/')}
							class="group flex items-center gap-4 p-5 text-base-content transition-all hover:bg-base-200/50 border-r border-b border-base-300"
						>
							<div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-base-200/60 text-base-content/70 group-hover:text-primary transition-colors">
								{#if icon}
									<svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="1.8" viewBox={icon.viewBox}>
										<path stroke-linecap="round" stroke-linejoin="round" d={icon.path} />
									</svg>
								{/if}
							</div>
							<div class="min-w-0 flex-1">
								<div class="text-sm font-semibold text-base-content">{action.label}</div>
								<div class="text-xs text-base-content/70 mt-0.5">{action.desc}</div>
							</div>
							<svg class="h-4 w-4 shrink-0 text-base-content/70/40 group-hover:text-primary transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
								<path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
							</svg>
						</a>
					{/each}
				</div>
			</div>
		</div>

		<!-- Info Bar (daisyui card) -->
		<div class="card bg-base-100 border border-base-300 shadow-sm">
			<div class="card-body px-5 py-3 flex-row items-center gap-3">
				<div class="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
					<svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
				</div>
				<div class="min-w-0 flex-1">
					<p class="text-xs font-bold text-base-content">Sistem Manajemen Madrasah</p>
					<p class="text-[11px] font-medium text-base-content/70">MTsN 2 Kolaka Utara &middot; WITA &middot; by Hasbi Awal</p>
				</div>
			</div>
		</div>
	</div>
{/if}