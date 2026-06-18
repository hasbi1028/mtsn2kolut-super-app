<script lang="ts">
	import type { ActionData } from './$types';
	import { appAttribution } from '$lib/branding';
	import { navigating } from '$app/state';

	let { form }: { form: ActionData } = $props();
	let pending = $derived(!!navigating.to);
	let passwordShown = $state(false);
</script>

<svelte:head>
	<title>Masuk — MTs Negeri 2 Kolaka Utara</title>
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
	<link href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:ital,wght@0,400;0,600;0,700;0,800&display=swap" rel="stylesheet" />
</svelte:head>

<div class="login-root">
	<!-- ═══ Left: Brand Panel (Desktop) ═══ -->
	<aside class="brand-panel">
		<div class="brand-panel-bg"></div>
		<div class="brand-panel-content">
			<a href="/" class="brand-link">
				<div class="brand-icon-wrap">
					<img src="/brand/logo-kemenag-icon-64.png" alt="Logo Kemenag" class="brand-icon-img" />
				</div>
				<span class="brand-link-text">Kembali ke Beranda</span>
			</a>

			<div class="brand-headline">
				<div class="brand-badge">
					<span class="brand-badge-dot"></span>
					Portal Administrasi Terintegrasi
				</div>
				<h1 class="brand-title">Ruang Kerja Digital <span class="brand-title-accent">MTsN 2.</span></h1>
				<p class="brand-desc">Platform tata kelola mandiri untuk akademik, CBT, kesiswaan, dan publikasi resmi madrasah.</p>
			</div>

			<div class="brand-stats">
				<div class="brand-stat">
					<span class="brand-stat-label">Otoritas</span>
					<span class="brand-stat-value">Admin & Guru</span>
				</div>
				<div class="brand-stat-divider"></div>
				<div class="brand-stat">
					<span class="brand-stat-label">Enkripsi</span>
					<span class="brand-stat-value">AES-256 Bit</span>
				</div>
				<div class="brand-stat-divider"></div>
				<div class="brand-stat">
					<span class="brand-stat-label">Logging</span>
					<span class="brand-stat-value">Audit Aktif</span>
				</div>
			</div>
		</div>
	</aside>

	<!-- ═══ Right: Login Form ═══ -->
	<section class="form-panel">
		<div class="form-inner">

			<!-- Mobile Header -->
			<div class="mobile-header">
				<div class="mobile-logo">
					<img src="/brand/logo-kemenag-icon-64.png" alt="Logo Kemenag" class="mobile-logo-img" />
				</div>
				<h2 class="mobile-title">MTsN 2 Kolaka Utara</h2>
				<p class="mobile-subtitle">Sistem Administrasi Sekolah</p>
			</div>

			<!-- Heading -->
			<div class="form-heading">
				<h2 class="form-title">Selamat Datang</h2>
				<p class="form-desc">Silakan masuk ke akun kerja Anda untuk mengakses layanan internal madrasah.</p>
			</div>

			<!-- Form -->
			<form method="POST" class="login-form">
				{#if form?.error}
					<div class="error-banner">
						<svg class="error-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						<p class="error-text">{form.error}</p>
					</div>
				{/if}

				<!-- Username -->
				<div class="field">
					<label for="username" class="field-label">Nama Pengguna</label>
					<div class="input-wrap">
						<input
							id="username"
							name="username"
							type="text"
							autocomplete="username"
							placeholder="Masukkan username"
							required
							class="input-field"
						/>
						<div class="input-icon">
							<svg class="icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
							</svg>
						</div>
					</div>
				</div>

				<!-- Password -->
				<div class="field">
					<label for="password" class="field-label">Kata Sandi</label>
					<div class="input-wrap">
						<input
							id="password"
							name="password"
							type={passwordShown ? 'text' : 'password'}
							autocomplete="current-password"
							placeholder="••••••••"
							required
							class="input-field input-field--password"
						/>
						<div class="input-icon">
							<svg class="icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
							</svg>
						</div>
						<button
							type="button"
							onclick={() => (passwordShown = !passwordShown)}
							class="toggle-password"
						>
							{passwordShown ? 'Sembunyi' : 'Lihat'}
						</button>
					</div>
				</div>

				<!-- Submit -->
				<button
					type="submit"
					disabled={pending}
					class="submit-btn"
				>
					{#if pending}
						<svg class="spinner" fill="none" viewBox="0 0 24 24">
							<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
							<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
						</svg>
						Memproses...
					{:else}
						Masuk Ke Sistem
						<svg class="arrow-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7l5 5m0 0l-5 5m5-5H6" />
						</svg>
					{/if}
				</button>

				<!-- Security Note -->
				<div class="security-note">
					<svg class="security-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
					</svg>
					<p class="security-text">Keamanan terjamin. Seluruh sesi diaudit secara real-time untuk integritas data.</p>
				</div>
			</form>

			<!-- Attribution -->
			<div class="attribution">
				<p>{appAttribution.loginLabel}</p>
			</div>
		</div>
	</section>
</div>

<style>
	/* ── Reset & Base ── */
	.login-root {
		min-height: 100vh;
		display: grid;
		place-items: stretch;
		font-family: 'Plus Jakarta Sans', system-ui, sans-serif;
	}

	@media (min-width: 1024px) {
		.login-root {
			grid-template-columns: 1.1fr 1fr;
		}
	}

	/* ── Brand Panel (Desktop) ── */
	.brand-panel {
		display: none;
		position: relative;
		overflow: hidden;
		color: white;
	}

	.brand-panel-bg {
		position: absolute;
		inset: 0;
		background: linear-gradient(165deg, #166534 0%, #14532d 100%);
	}

	.brand-panel-bg::before {
		content: '';
		position: absolute;
		inset: 0;
		background:
			radial-gradient(circle at 20% 80%, rgba(74,222,128,0.12) 0%, transparent 50%),
			radial-gradient(circle at 80% 20%, rgba(74,222,128,0.08) 0%, transparent 50%);
	}

	@media (min-width: 1024px) {
		.brand-panel {
			display: flex;
		}
	}

	.brand-panel-content {
		position: relative;
		z-index: 1;
		display: flex;
		flex-direction: column;
		justify-content: space-between;
		padding: 3rem;
	}

	/* ── Brand Link ── */
	.brand-link {
		display: inline-flex;
		align-items: center;
		gap: 0.75rem;
		text-decoration: none;
		color: white;
		transition: opacity 0.2s;
	}

	.brand-link:hover {
		opacity: 0.9;
	}

	.brand-icon-wrap {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 2.75rem;
		height: 2.75rem;
		border-radius: 14px;
		background: rgba(255,255,255,0.12);
		border: 1px solid rgba(255,255,255,0.2);
		backdrop-filter: blur(12px);
	}

	.brand-icon-img {
		width: 1.5rem;
		height: 1.5rem;
	}

	.brand-link-text {
		font-size: 0.8rem;
		font-weight: 700;
	}

	/* ── Brand Headline ── */
	.brand-headline {
		margin-top: 5rem;
		max-width: 28rem;
	}

	.brand-badge {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.35rem 0.85rem;
		border-radius: 100px;
		background: rgba(74,222,128,0.1);
		border: 1px solid rgba(74,222,128,0.2);
		margin-bottom: 1.5rem;
		font-size: 0.6rem;
		font-weight: 900;
		letter-spacing: 0.18em;
		text-transform: uppercase;
		color: #86efac;
	}

	.brand-badge-dot {
		width: 0.375rem;
		height: 0.375rem;
		border-radius: 50%;
		background: #4ade80;
		animation: pulse-dot 2s ease-in-out infinite;
	}

	@keyframes pulse-dot {
		0%, 100% { opacity: 1; }
		50% { opacity: 0.4; }
	}

	.brand-title {
		margin: 0;
		font-size: 3.2rem;
		font-weight: 900;
		line-height: 1.05;
		letter-spacing: -0.04em;
	}

	.brand-title-accent {
		color: #4ade80;
	}

	.brand-desc {
		margin-top: 1.25rem;
		font-size: 1.05rem;
		color: rgba(255,255,255,0.7);
		line-height: 1.6;
		font-weight: 500;
	}

	/* ── Brand Stats ── */
	.brand-stats {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 1.25rem 1.75rem;
		border-radius: 18px;
		background: rgba(0,0,0,0.15);
		backdrop-filter: blur(8px);
		border: 1px solid rgba(255,255,255,0.08);
	}

	.brand-stat {
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
		flex: 1;
	}

	.brand-stat:last-child {
		text-align: right;
	}

	.brand-stat-label {
		font-size: 0.55rem;
		font-weight: 900;
		letter-spacing: 0.16em;
		text-transform: uppercase;
		color: #4ade80;
	}

	.brand-stat-value {
		font-size: 0.75rem;
		font-weight: 700;
	}

	.brand-stat-divider {
		width: 1px;
		height: 2rem;
		background: rgba(255,255,255,0.12);
		margin: 0 1rem;
	}

	/* ── Form Panel ── */
	.form-panel {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 2rem 1.25rem;
		background: white;
	}

	@media (min-width: 1024px) {
		.form-panel {
			padding: 3rem;
		}
	}

	.form-inner {
		width: 100%;
		max-width: 24rem;
		display: flex;
		flex-direction: column;
		gap: 2.5rem;
	}

	/* ── Mobile Header ── */
	.mobile-header {
		display: flex;
		flex-direction: column;
		align-items: center;
		text-align: center;
		gap: 0.75rem;
	}

	@media (min-width: 1024px) {
		.mobile-header {
			display: none;
		}
	}

	.mobile-logo {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 4rem;
		height: 4rem;
		border-radius: 18px;
		background: #052e16;
		box-shadow: 0 16px 32px rgba(5,46,22,0.25);
	}

	.mobile-logo-img {
		width: 2.25rem;
		height: 2.25rem;
	}

	.mobile-title {
		margin: 0;
		font-size: 1.3rem;
		font-weight: 900;
		color: #0f172a;
		letter-spacing: -0.03em;
	}

	.mobile-subtitle {
		margin: 0;
		font-size: 0.7rem;
		font-weight: 800;
		text-transform: uppercase;
		letter-spacing: 0.16em;
		color: #16a34a;
	}

	/* ── Form Heading ── */
	.form-heading {
		space-y: 0.75rem;
	}

	.form-title {
		margin: 0;
		font-size: 2rem;
		font-weight: 900;
		color: #0f172a;
		letter-spacing: -0.04em;
	}

	.form-desc {
		margin: 0.5rem 0 0;
		color: #64748b;
		font-size: 0.9rem;
		line-height: 1.55;
		font-weight: 500;
	}

	/* ── Login Form ── */
	.login-form {
		display: flex;
		flex-direction: column;
		gap: 1.25rem;
	}

	/* ── Error Banner ── */
	.error-banner {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 1rem;
		border-radius: 14px;
		background: #fef2f2;
		border: 1px solid #fecaca;
		animation: slideIn 0.3s ease;
	}

	@keyframes slideIn {
		from { opacity: 0; transform: translateY(-8px); }
		to { opacity: 1; transform: translateY(0); }
	}

	.error-icon {
		width: 1.25rem;
		height: 1.25rem;
		color: #dc2626;
		flex-shrink: 0;
	}

	.error-text {
		margin: 0;
		font-size: 0.85rem;
		font-weight: 700;
		color: #dc2626;
		line-height: 1.4;
	}

	/* ── Fields ── */
	.field {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.field-label {
		font-size: 0.65rem;
		font-weight: 900;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: #94a3b8;
		margin-left: 0.25rem;
	}

	.input-wrap {
		position: relative;
	}

	.input-field {
		width: 100%;
		height: 3.25rem;
		padding: 0 3rem;
		padding-left: 3rem;
		border-radius: 14px;
		background: #f8fafc;
		border: 1.5px solid #e2e8f0;
		color: #0f172a;
		font-size: 0.95rem;
		font-weight: 700;
		outline: none;
		transition: all 0.2s ease;
		font-family: inherit;
	}

	.input-field--password {
		padding-right: 5rem;
	}

	.input-field::placeholder {
		color: #cbd5e1;
	}

	.input-field:focus {
		border-color: #16a34a;
		box-shadow: 0 0 0 4px rgba(22,163,74,0.1);
		background: white;
	}

	.input-icon {
		position: absolute;
		left: 1rem;
		top: 50%;
		transform: translateY(-50%);
		color: #94a3b8;
		transition: color 0.2s;
		pointer-events: none;
	}

	.input-wrap:focus-within .input-icon {
		color: #16a34a;
	}

	.icon {
		width: 1.2rem;
		height: 1.2rem;
	}

	.toggle-password {
		position: absolute;
		right: 1rem;
		top: 50%;
		transform: translateY(-50%);
		padding: 0.25rem 0.5rem;
		border: 0;
		background: none;
		font-size: 0.65rem;
		font-weight: 900;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		color: #16a34a;
		cursor: pointer;
		transition: color 0.15s;
		font-family: inherit;
	}

	.toggle-password:hover {
		color: #15803d;
	}

	/* ── Submit Button ── */
	.submit-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		width: 100%;
		height: 3.25rem;
		border: 0;
		border-radius: 14px;
		background: #16a34a;
		color: white;
		font-size: 0.8rem;
		font-weight: 900;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		box-shadow: 0 12px 28px rgba(22,163,74,0.3);
		cursor: pointer;
		transition: all 0.2s ease;
		font-family: inherit;
	}

	.submit-btn:hover:not(:disabled) {
		background: #15803d;
		box-shadow: 0 16px 36px rgba(22,163,74,0.35);
		transform: translateY(-1px);
	}

	.submit-btn:active:not(:disabled) {
		transform: translateY(0);
	}

	.submit-btn:disabled {
		opacity: 0.55;
		cursor: not-allowed;
		transform: none;
	}

	.spinner {
		width: 1rem;
		height: 1rem;
		animation: spin 0.75s linear infinite;
	}

	@keyframes spin {
		100% { transform: rotate(360deg); }
	}

	.arrow-icon {
		width: 1rem;
		height: 1rem;
		transition: transform 0.2s;
	}

	.submit-btn:hover:not(:disabled) .arrow-icon {
		transform: translateX(3px);
	}

	/* ── Security Note ── */
	.security-note {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		color: #94a3b8;
	}

	.security-icon {
		width: 0.875rem;
		height: 0.875rem;
		flex-shrink: 0;
	}

	.security-text {
		margin: 0;
		font-size: 0.65rem;
		font-weight: 600;
		max-width: 15rem;
		text-align: center;
		line-height: 1.5;
	}

	/* ── Attribution ── */
	.attribution {
		padding-top: 1rem;
		border-top: 1px solid #f1f5f9;
		text-align: center;
	}

	.attribution p {
		margin: 0;
		font-size: 0.55rem;
		font-weight: 900;
		text-transform: uppercase;
		letter-spacing: 0.2em;
		color: #cbd5e1;
	}

	/* ── Mobile Adjustments ── */
	@media (max-width: 380px) {
		.form-panel {
			padding: 1.5rem 1rem;
		}

		.form-inner {
			gap: 2rem;
		}

		.form-title {
			font-size: 1.6rem;
		}
	}
</style>
