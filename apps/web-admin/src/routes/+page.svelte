<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { defaultBranding, versionedAsset, type BrandingSettings } from '$lib/branding';

	let { data }: {
		data: {
			user?: { username?: string; role?: string; roles?: string[] };
			branding?: BrandingSettings;
		};
	} = $props();

	const branding = $derived(data.branding ?? defaultBranding);

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
	<!-- ═══ Public Landing Page ═══ -->
	<div class="landing">
		<!-- Hero Section -->
		<div class="hero">
			<div class="hero-bg"></div>
			<div class="hero-content">
				<div class="hero-logo-wrap">
					{#if branding?.mark_url}
						<img src={versionedAsset(branding.mark_url, branding.version)} alt="Logo Kemenag" class="hero-logo" />
					{:else}
						<div class="hero-logo-placeholder">M</div>
					{/if}
				</div>
				<h1 class="hero-title">{branding?.app_name || 'MTs Negeri 2 Kolaka Utara'}</h1>
				<p class="hero-subtitle">{branding?.tagline || 'Sistem Informasi Madrasah'}</p>
				<p class="hero-desc">
					Platform tata kelola mandiri untuk pengelolaan pegawai, kehadiran, dan pengaturan sistem madrasah.
				</p>
				<a href={resolve('/login')} class="hero-cta">
					Masuk ke Sistem
					<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7l5 5m0 0l-5 5m5-5H6" />
					</svg>
				</a>
			</div>
		</div>

		<!-- Features Section -->
		<div class="features">
			<div class="features-header">
				<p class="features-kicker">Modul Utama</p>
				<h2 class="features-title">Fitur Sistem</h2>
				<p class="features-desc">Kelola seluruh aspek operasional madrasah dalam satu platform.</p>
			</div>
			<div class="features-grid">
				<div class="feature-card">
					<div class="feature-icon feature-icon--primary">
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
					<h3 class="feature-label">Kehadiran</h3>
					<p class="feature-desc">Pantau absensi pegawai melalui integrasi PUSAKA Kemenag secara real-time.</p>
				</div>
				<div class="feature-card">
					<div class="feature-icon feature-icon--success">
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
						</svg>
					</div>
					<h3 class="feature-label">Pegawai</h3>
					<p class="feature-desc">Kelola data pegawai, jadwal kerja, dan akun PUSAKA dalam satu tempat.</p>
				</div>
				<div class="feature-card">
					<div class="feature-icon feature-icon--info">
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
						</svg>
					</div>
					<h3 class="feature-label">Pengaturan</h3>
					<p class="feature-desc">Konfigurasi sistem, backup data, audit log, dan manajemen pengguna.</p>
				</div>
			</div>
		</div>
	</div>
{:else}
	<!-- ═══ Dashboard ═══ -->
	<div class="dashboard">

		<!-- Welcome Header -->
		<div class="welcome-section">
			<div class="welcome-text">
				<h1 class="welcome-title">Dashboard</h1>
				<p class="welcome-subtitle">
					Selamat datang, <span class="welcome-name">{data.user?.username || 'Pengguna'}</span>
				</p>
			</div>
			<div class="welcome-time">
				<div class="time-badge">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
					<span class="time-value">{jamSekarang()} WITA</span>
				</div>
				<p class="time-date">{hariIni()}</p>
			</div>
		</div>

		<!-- Stat Cards -->
		<div class="stats-grid">
			<!-- User Card -->
			<div class="stat-card stat-card--primary">
				<div class="stat-icon-wrap stat-icon--primary">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
					</svg>
				</div>
				<div class="stat-label">Pengguna</div>
				<div class="stat-value">{data.user?.username || 'Pengguna'}</div>
				<div class="stat-sublabel">{isAdmin ? 'Administrator' : 'Pengguna'}</div>
			</div>

			<!-- Status Card -->
			<div class="stat-card stat-card--success">
				<div class="stat-icon-wrap stat-icon--success">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
				</div>
				<div class="stat-label">Status</div>
				<div class="stat-value">Aktif</div>
				<div class="stat-sublabel">Sistem Online</div>
			</div>

			<!-- Madrasah Card -->
			<div class="stat-card stat-card--info">
				<div class="stat-icon-wrap stat-icon--info">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0H5m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
					</svg>
				</div>
				<div class="stat-label">Madrasah</div>
				<div class="stat-value">MTsN 2 Kolut</div>
				<div class="stat-sublabel">MTs Negeri 2 Kolaka Utara</div>
			</div>

			<!-- App Card -->
			<div class="stat-card stat-card--warning">
				<div class="stat-icon-wrap stat-icon--warning">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
					</svg>
				</div>
				<div class="stat-label">Aplikasi</div>
				<div class="stat-value">SMM v2.0</div>
				<div class="stat-sublabel">Sistem Manajemen Madrasah</div>
			</div>
		</div>

		<!-- Modul Aplikasi -->
		<div class="modules-card">
			<div class="modules-header">
				<h2 class="modules-title">Modul Aplikasi</h2>
				<span class="modules-count">{actions.length} modul</span>
			</div>
			<div class="modules-grid">
				{#each actions as action, i (action.href)}
					{@const icon = iconMap[action.icon]}
					<a
						href={resolve(action.href as '/')}
						class="module-item"
					>
						<div class="module-icon">
							{#if icon}
								<svg class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="1.8" viewBox={icon.viewBox}>
									<path stroke-linecap="round" stroke-linejoin="round" d={icon.path} />
								</svg>
							{/if}
						</div>
						<div class="module-info">
							<div class="module-label">{action.label}</div>
							<div class="module-desc">{action.desc}</div>
						</div>
						<svg class="module-arrow" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
							<path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
						</svg>
					</a>
				{/each}
			</div>
		</div>

		<!-- Quick Access -->
		<div class="quick-access-grid">
			<a href="/pusaka" class="quick-card quick-card--primary">
				<div class="quick-icon quick-icon--primary">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
				</div>
				<h3 class="quick-label">Monitor Kehadiran</h3>
				<p class="quick-desc">Pantau absensi pegawai via PUSAKA</p>
				<span class="quick-link">Buka Modul →</span>
			</a>

			<a href="/pusaka/kehadiran" class="quick-card quick-card--success">
				<div class="quick-icon quick-icon--success">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
					</svg>
				</div>
				<h3 class="quick-label">Data Kehadiran</h3>
				<p class="quick-desc">Rekap harian dari PUSAKA Kemenag</p>
				<span class="quick-link quick-link--success">Lihat Rekap →</span>
			</a>

			<a href="/pusaka/summary" class="quick-card quick-card--warning">
				<div class="quick-icon quick-icon--warning">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
					</svg>
				</div>
				<h3 class="quick-label">Ringkasan Kehadiran</h3>
				<p class="quick-desc">Akumulasi per periode / bulan</p>
				<span class="quick-link quick-link--warning">Lihat Ringkasan →</span>
			</a>
		</div>

		<!-- Info Footer -->
		<div class="info-footer">
			<div class="info-icon">
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
			</div>
			<div class="info-text">
				<p class="info-label">Sistem Manajemen Madrasah</p>
				<p class="info-sublabel">MTsN 2 Kolaka Utara · WITA · by Hasbi Awal</p>
			</div>
		</div>
	</div>
{/if}

<style>
	/* ── Fonts ── */
	:global(html) {
		font-family: 'Plus Jakarta Sans', system-ui, sans-serif;
	}

	/* ── Landing Page ── */
	.landing {
		display: flex;
		flex-direction: column;
	}

	/* ── Hero Section ── */
	.hero {
		position: relative;
		overflow: hidden;
		padding: 4rem 1.25rem;
		background: linear-gradient(165deg, #166534 0%, #14532d 100%);
		color: white;
		text-align: center;
	}

	.hero-bg {
		position: absolute;
		inset: 0;
		background:
			radial-gradient(circle at 20% 80%, rgba(74,222,128,0.15) 0%, transparent 50%),
			radial-gradient(circle at 80% 20%, rgba(74,222,128,0.1) 0%, transparent 50%);
	}

	.hero-content {
		position: relative;
		z-index: 1;
		max-width: 32rem;
		margin: 0 auto;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1rem;
	}

	.hero-logo-wrap {
		margin-bottom: 0.5rem;
	}

	.hero-logo {
		width: 5rem;
		height: 5rem;
		object-fit: contain;
		border-radius: 22px;
		background: rgba(255,255,255,0.12);
		padding: 0.75rem;
		border: 1px solid rgba(255,255,255,0.2);
		backdrop-filter: blur(12px);
		box-shadow: 0 20px 40px rgba(0,0,0,0.2);
	}

	.hero-logo-placeholder {
		width: 5rem;
		height: 5rem;
		border-radius: 22px;
		background: rgba(255,255,255,0.12);
		border: 1px solid rgba(255,255,255,0.2);
		backdrop-filter: blur(12px);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 2rem;
		font-weight: 900;
		box-shadow: 0 20px 40px rgba(0,0,0,0.2);
	}

	.hero-title {
		margin: 0;
		font-size: 2.2rem;
		font-weight: 900;
		line-height: 1.1;
		letter-spacing: -0.04em;
	}

	@media (min-width: 640px) {
		.hero-title {
			font-size: 3rem;
		}
	}

	.hero-subtitle {
		margin: 0;
		font-size: 0.7rem;
		font-weight: 900;
		text-transform: uppercase;
		letter-spacing: 0.18em;
		color: #4ade80;
	}

	.hero-desc {
		margin: 0;
		font-size: 0.95rem;
		color: rgba(255,255,255,0.7);
		line-height: 1.6;
		font-weight: 500;
	}

	.hero-cta {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		margin-top: 1rem;
		padding: 0.85rem 2rem;
		border-radius: 14px;
		background: white;
		color: #166534;
		font-size: 0.85rem;
		font-weight: 900;
		text-decoration: none;
		text-transform: uppercase;
		letter-spacing: 0.04em;
		box-shadow: 0 12px 28px rgba(0,0,0,0.2);
		transition: all 0.2s;
	}

	.hero-cta:hover {
		box-shadow: 0 16px 36px rgba(0,0,0,0.25);
		transform: translateY(-2px);
	}

	/* ── Features Section ── */
	.features {
		padding: 3rem 1.25rem;
		background: white;
	}

	.features-header {
		text-align: center;
		max-width: 32rem;
		margin: 0 auto 2rem;
	}

	.features-kicker {
		margin: 0;
		font-size: 0.65rem;
		font-weight: 900;
		text-transform: uppercase;
		letter-spacing: 0.18em;
		color: #16a34a;
	}

	.features-title {
		margin: 0.5rem 0 0;
		font-size: 1.75rem;
		font-weight: 900;
		color: #0f172a;
		letter-spacing: -0.04em;
	}

	.features-desc {
		margin: 0.5rem 0 0;
		font-size: 0.9rem;
		color: #64748b;
		font-weight: 500;
	}

	.features-grid {
		display: grid;
		grid-template-columns: 1fr;
		gap: 1rem;
		max-width: 40rem;
		margin: 0 auto;
	}

	@media (min-width: 640px) {
		.features-grid {
			grid-template-columns: repeat(3, 1fr);
		}
	}

	.feature-card {
		padding: 1.5rem;
		border-radius: 18px;
		background: #f8fafc;
		border: 1.5px solid #f1f5f9;
		transition: all 0.2s;
	}

	.feature-card:hover {
		border-color: #dcfce7;
		box-shadow: 0 8px 24px rgba(22,163,74,0.08);
	}

	.feature-icon {
		width: 2.75rem;
		height: 2.75rem;
		border-radius: 14px;
		display: flex;
		align-items: center;
		justify-content: center;
		margin-bottom: 1rem;
	}

	.feature-icon--primary {
		background: rgba(22,163,74,0.1);
		color: #16a34a;
	}

	.feature-icon--success {
		background: rgba(22,163,74,0.1);
		color: #16a34a;
	}

	.feature-icon--info {
		background: rgba(14,116,144,0.1);
		color: #0e7490;
	}

	.feature-label {
		margin: 0;
		font-size: 1rem;
		font-weight: 900;
		color: #0f172a;
		letter-spacing: -0.02em;
	}

	.feature-desc {
		margin: 0.5rem 0 0;
		font-size: 0.8rem;
		color: #64748b;
		line-height: 1.5;
		font-weight: 500;
	}

	/* ── Logged Out View (minimal) ── */
	.logged-out-view {
		min-height: 60vh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 1rem;
	}

	.logged-out-card {
		max-width: 24rem;
		text-align: center;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1.25rem;
	}

	.logged-out-logo {
		width: 3.5rem;
		height: 3.5rem;
		border-radius: 18px;
		background: #16a34a;
		color: white;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 1.5rem;
		font-weight: 900;
		box-shadow: 0 16px 32px rgba(22,163,74,0.3);
	}

	.logged-out-title {
		margin: 0;
		font-size: 1.5rem;
		font-weight: 900;
		color: #0f172a;
		letter-spacing: -0.03em;
	}

	.logged-out-desc {
		margin: 0;
		font-size: 0.85rem;
		color: #64748b;
		font-weight: 600;
	}

	.logged-out-button {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1.5rem;
		border-radius: 14px;
		background: #16a34a;
		color: white;
		font-size: 0.85rem;
		font-weight: 800;
		text-decoration: none;
		transition: all 0.2s;
		box-shadow: 0 12px 24px rgba(22,163,74,0.25);
	}

	.logged-out-button:hover {
		background: #15803d;
		box-shadow: 0 16px 32px rgba(22,163,74,0.3);
		transform: translateY(-1px);
	}

	/* ── Dashboard ── */
	.dashboard {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	/* ── Welcome Section ── */
	.welcome-section {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
	}

	.welcome-text {
		flex: 1;
	}

	.welcome-title {
		margin: 0;
		font-size: 1.75rem;
		font-weight: 900;
		color: #0f172a;
		letter-spacing: -0.04em;
	}

	.welcome-subtitle {
		margin: 0.25rem 0 0;
		font-size: 0.9rem;
		color: #64748b;
		font-weight: 500;
	}

	.welcome-name {
		font-weight: 800;
		color: #0f172a;
	}

	.welcome-time {
		text-align: right;
		flex-shrink: 0;
	}

	.time-badge {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.35rem 0.75rem;
		border-radius: 10px;
		background: #f0fdf4;
		border: 1px solid #dcfce7;
		color: #16a34a;
		font-size: 0.75rem;
		font-weight: 800;
	}

	.time-value {
		font-variant-numeric: tabular-nums;
	}

	.time-date {
		margin: 0.35rem 0 0;
		font-size: 0.7rem;
		color: #94a3b8;
		font-weight: 600;
	}

	/* ── Stats Grid ── */
	.stats-grid {
		display: grid;
		grid-template-columns: repeat(1, 1fr);
		gap: 0.75rem;
	}

	@media (min-width: 640px) {
		.stats-grid {
			grid-template-columns: repeat(2, 1fr);
		}
	}

	@media (min-width: 1024px) {
		.stats-grid {
			grid-template-columns: repeat(4, 1fr);
		}
	}

	.stat-card {
		padding: 1.25rem;
		border-radius: 18px;
		background: white;
		border: 1.5px solid #f1f5f9;
		transition: all 0.2s;
	}

	.stat-card:hover {
		border-color: #e2e8f0;
		box-shadow: 0 8px 24px rgba(0,0,0,0.04);
	}

	.stat-icon-wrap {
		width: 2.5rem;
		height: 2.5rem;
		border-radius: 12px;
		display: flex;
		align-items: center;
		justify-content: center;
		margin-bottom: 0.75rem;
	}

	.stat-icon--primary {
		background: rgba(22,163,74,0.1);
		color: #16a34a;
	}

	.stat-icon--success {
		background: rgba(22,163,74,0.1);
		color: #16a34a;
	}

	.stat-icon--info {
		background: rgba(14,116,144,0.1);
		color: #0e7490;
	}

	.stat-icon--warning {
		background: rgba(217,119,6,0.1);
		color: #d97706;
	}

	.stat-label {
		font-size: 0.6rem;
		font-weight: 900;
		text-transform: uppercase;
		letter-spacing: 0.16em;
		color: #94a3b8;
		margin-bottom: 0.25rem;
	}

	.stat-value {
		font-size: 1.25rem;
		font-weight: 900;
		color: #0f172a;
		letter-spacing: -0.03em;
		line-height: 1.2;
	}

	.stat-sublabel {
		font-size: 0.7rem;
		color: #64748b;
		font-weight: 600;
		margin-top: 0.25rem;
	}

	/* ── Modules Card ── */
	.modules-card {
		border-radius: 18px;
		background: white;
		border: 1.5px solid #f1f5f9;
		overflow: hidden;
	}

	.modules-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 1.25rem 1.5rem;
		border-bottom: 1px solid #f1f5f9;
	}

	.modules-title {
		margin: 0;
		font-size: 0.95rem;
		font-weight: 900;
		color: #0f172a;
		letter-spacing: -0.02em;
	}

	.modules-count {
		font-size: 0.65rem;
		font-weight: 800;
		color: #94a3b8;
		text-transform: uppercase;
		letter-spacing: 0.1em;
	}

	.modules-grid {
		display: grid;
		grid-template-columns: 1fr;
	}

	@media (min-width: 640px) {
		.modules-grid {
			grid-template-columns: repeat(2, 1fr);
		}
	}

	@media (min-width: 1024px) {
		.modules-grid {
			grid-template-columns: repeat(3, 1fr);
		}
	}

	.module-item {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1rem 1.5rem;
		border-right: 1px solid #f1f5f9;
		border-bottom: 1px solid #f1f5f9;
		text-decoration: none;
		color: inherit;
		transition: all 0.15s;
	}

	@media (min-width: 1024px) {
		.module-item:nth-child(3n) {
			border-right: 0;
		}
	}

	.module-item:hover {
		background: #f0fdf4;
	}

	.module-icon {
		width: 2.5rem;
		height: 2.5rem;
		border-radius: 12px;
		background: #f8fafc;
		display: flex;
		align-items: center;
		justify-content: center;
		color: #64748b;
		flex-shrink: 0;
		transition: all 0.15s;
	}

	.module-item:hover .module-icon {
		background: rgba(22,163,74,0.1);
		color: #16a34a;
	}

	.module-info {
		flex: 1;
		min-width: 0;
	}

	.module-label {
		font-size: 0.85rem;
		font-weight: 800;
		color: #0f172a;
	}

	.module-desc {
		font-size: 0.7rem;
		color: #64748b;
		font-weight: 500;
		margin-top: 0.1rem;
	}

	.module-arrow {
		width: 1rem;
		height: 1rem;
		color: #cbd5e1;
		flex-shrink: 0;
		transition: all 0.15s;
	}

	.module-item:hover .module-arrow {
		color: #16a34a;
		transform: translateX(2px);
	}

	/* ── Quick Access ── */
	.quick-access-grid {
		display: grid;
		grid-template-columns: 1fr;
		gap: 0.75rem;
	}

	@media (min-width: 640px) {
		.quick-access-grid {
			grid-template-columns: repeat(3, 1fr);
		}
	}

	.quick-card {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		padding: 1.5rem;
		border-radius: 18px;
		text-decoration: none;
		color: inherit;
		transition: all 0.2s;
	}

	.quick-card--primary {
		background: linear-gradient(135deg, #f0fdf4 0%, #dcfce7 100%);
		border: 1px solid #bbf7d0;
	}

	.quick-card--primary:hover {
		box-shadow: 0 12px 28px rgba(22,163,74,0.15);
		transform: translateY(-2px);
	}

	.quick-card--success {
		background: linear-gradient(135deg, #f0fdf4 0%, #dcfce7 100%);
		border: 1px solid #bbf7d0;
	}

	.quick-card--success:hover {
		box-shadow: 0 12px 28px rgba(22,163,74,0.15);
		transform: translateY(-2px);
	}

	.quick-card--warning {
		background: linear-gradient(135deg, #fffbeb 0%, #fef3c7 100%);
		border: 1px solid #fde68a;
	}

	.quick-card--warning:hover {
		box-shadow: 0 12px 28px rgba(217,119,6,0.15);
		transform: translateY(-2px);
	}

	.quick-icon {
		width: 2.5rem;
		height: 2.5rem;
		border-radius: 12px;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.quick-icon--primary {
		background: rgba(22,163,74,0.15);
		color: #16a34a;
	}

	.quick-icon--success {
		background: rgba(22,163,74,0.15);
		color: #16a34a;
	}

	.quick-icon--warning {
		background: rgba(217,119,6,0.15);
		color: #d97706;
	}

	.quick-label {
		margin: 0;
		font-size: 0.95rem;
		font-weight: 900;
		color: #0f172a;
		letter-spacing: -0.02em;
	}

	.quick-desc {
		margin: 0;
		font-size: 0.75rem;
		color: #64748b;
		font-weight: 500;
		line-height: 1.4;
	}

	.quick-link {
		font-size: 0.7rem;
		font-weight: 800;
		color: #16a34a;
		transition: color 0.15s;
	}

	.quick-link--success {
		color: #16a34a;
	}

	.quick-link--warning {
		color: #d97706;
	}

	.quick-card:hover .quick-link {
		color: #15803d;
	}

	/* ── Info Footer ── */
	.info-footer {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 1rem 1.25rem;
		border-radius: 14px;
		background: #f8fafc;
		border: 1px solid #f1f5f9;
	}

	.info-icon {
		width: 2rem;
		height: 2rem;
		border-radius: 10px;
		background: rgba(22,163,74,0.1);
		color: #16a34a;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.info-label {
		margin: 0;
		font-size: 0.75rem;
		font-weight: 800;
		color: #0f172a;
	}

	.info-sublabel {
		margin: 0.1rem 0 0;
		font-size: 0.65rem;
		color: #94a3b8;
		font-weight: 600;
	}
</style>
