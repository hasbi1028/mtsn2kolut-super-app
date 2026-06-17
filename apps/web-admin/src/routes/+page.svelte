<script lang="ts">
	import { resolve } from '$app/paths';

	let { data }: {
		data: {
			user?: { name?: string; role?: string; roles?: string[]; permissions?: string[] };
		};
	} = $props();

	const roles = $derived(data.user?.roles || (data.user?.role ? [data.user.role] : []));
	const isAdmin = $derived(roles.includes('admin'));
	const namaUser = $derived(data.user?.name || 'Pengguna');

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
		{ label: 'Monitor Kehadiran', desc: 'Pantau absensi pegawai hari ini', href: '/pusaka', icon: 'clock' },
		{ label: 'Data Pegawai', desc: 'Kelola daftar pegawai madrasah', href: '/employees', icon: 'users' },
		{ label: 'Manajemen Pengguna', desc: 'Kelola akun dan hak akses', href: '/settings/users', icon: 'shield' },
		{ label: 'Profil Sekolah', desc: 'Pengaturan profil madrasah', href: '/settings/school-profile', icon: 'building' },
		{ label: 'Audit Log', desc: 'Jejak aktivitas sistem', href: '/settings/audit', icon: 'file-text' },
		{ label: 'Backup Data', desc: 'Cadangkan data sistem', href: '/settings/backups', icon: 'database' },
	];

	const userActions = [
		{ label: 'Kehadiran', desc: 'Lihat catatan kehadiran', href: '/pusaka', icon: 'clock' },
		{ label: 'Akun Saya', desc: 'Pengaturan akun pribadi', href: '/settings/account', icon: 'user' },
	];

	const actions = $derived(isAdmin ? adminActions : userActions);

	const iconPaths: Record<string, string> = {
		clock: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z',
		users: 'M12 4.354a4 4 0 110 7.292 4 4 0 010-7.292zM15 21H9a2 2 0 01-2-2V12a2 2 0 012-2h6a2 2 0 012 2v7a2 2 0 01-2 2z',
		shield: 'M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z',
		building: 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0H5m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4',
		'file-text': 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z',
		database: 'M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4',
		user: 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z'
	};
</script>

<svelte:head>
	<title>Beranda — MTsN 2 Kolut</title>
</svelte:head>

{#if !data.user}
	<div class="flex min-h-[60vh] items-center justify-center p-4">
		<div class="w-full max-w-sm text-center space-y-5">
			<div class="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-primary-500 text-white text-xl font-black shadow-lg shadow-primary-500/25">
				M
			</div>
			<h1 class="text-xl font-black tracking-tight text-surface-900">
				MTs Negeri 2 Kolaka Utara
			</h1>
			<p class="text-sm font-medium text-surface-500">
				Sistem Manajemen Madrasah
			</p>
			<a
				href={resolve('/login')}
				class="inline-flex items-center gap-2 rounded-xl bg-primary-500 px-5 py-2.5 text-sm font-bold text-white transition-all hover:bg-primary-600 hover:shadow-lg hover:shadow-primary-500/25"
			>
				Masuk ke Sistem
			</a>
		</div>
	</div>
{:else}
	<div class="space-y-6">
		<!-- Simple Information Header (Replacing huge hero card) -->
		<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-surface-200 pb-4">
			<div>
				<h1 class="text-2xl font-bold tracking-tight text-surface-900">Dashboard</h1>
				<p class="text-sm text-surface-500 mt-1">Selamat datang kembali, <span class="font-medium text-surface-900">{namaUser}</span></p>
			</div>
			<div class="text-right">
				<div class="text-sm font-semibold text-surface-900">{hariIni()}</div>
				<div class="text-xs text-surface-500 mt-0.5">{jamSekarang()} WITA</div>
			</div>
		</div>

		<!-- Info Banners (like CBT status info) -->
		<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
			<div class="bg-white border border-surface-200 rounded-lg p-4 shadow-sm flex items-start gap-4">
				<div class="bg-primary-50 text-primary-600 p-2.5 rounded-md">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"></path></svg>
				</div>
				<div>
					<div class="text-xs font-semibold text-surface-500 uppercase tracking-wider">Akses Login</div>
					<div class="text-lg font-bold text-surface-900 mt-0.5">{isAdmin ? 'Administrator' : 'Pengguna'}</div>
				</div>
			</div>
			<div class="bg-white border border-surface-200 rounded-lg p-4 shadow-sm flex items-start gap-4">
				<div class="bg-emerald-50 text-emerald-600 p-2.5 rounded-md">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
				</div>
				<div>
					<div class="text-xs font-semibold text-surface-500 uppercase tracking-wider">Status Sistem</div>
					<div class="text-lg font-bold text-surface-900 mt-0.5 text-emerald-600">Aktif Dan Berjalan</div>
				</div>
			</div>
			<div class="bg-white border border-surface-200 rounded-lg p-4 shadow-sm flex items-start gap-4">
				<div class="bg-sky-50 text-sky-600 p-2.5 rounded-md">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
				</div>
				<div>
					<div class="text-xs font-semibold text-surface-500 uppercase tracking-wider">Instansi</div>
					<div class="text-lg font-bold text-surface-900 mt-0.5 truncate">MTsN 2 Kolut</div>
				</div>
			</div>
		</div>

		<!-- Grid Modul / Quick Access -->
		<div class="bg-white border border-surface-200 rounded-lg shadow-sm overflow-hidden">
			<div class="px-5 py-4 border-b border-surface-200 bg-surface-50/50">
				<h2 class="text-base font-semibold text-surface-900">Modul Aplikasi</h2>
			</div>
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 divide-y sm:divide-y-0 sm:divide-x border-surface-200">
				{#each actions as action, i}
					<a
						href={resolve(action.href as '/')}
						class="flex items-start gap-4 p-5 hover:bg-surface-50 transition-colors {i < 3 ? 'border-b border-surface-200 lg:border-b-0' : ''} {i === 0 || i === 3 ? 'sm:border-l-0' : ''}"
					>
						<div class="text-primary-600 mt-0.5">
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={iconPaths[action.icon]}></path>
							</svg>
						</div>
						<div>
							<div class="text-sm font-semibold text-surface-900">{action.label}</div>
							<div class="text-xs text-surface-500 mt-1 line-clamp-2">{action.desc}</div>
						</div>
					</a>
				{/each}
			</div>
		</div>
	</div>
{/if}
