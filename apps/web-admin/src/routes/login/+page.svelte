<script lang="ts">
	import type { ActionData } from './$types';
	import { appAttribution } from '$lib/branding';
	import { navigating } from '$app/state';

	let { form }: { form: ActionData } = $props();
	let pending = $derived(!!navigating.to);
	let passwordShown = $state(false);
</script>

<svelte:head><title>Masuk — MTs Negeri 2 Kolaka Utara</title></svelte:head>

<!--
  Skeleton UI 2-tone layout: HIJAU MADRASAH kiri, PUTIH BERSIH kanan
  - Left panel: bg-primary-700 (hijau Kemenag) with text-primary-50
  - Right panel: bg-surface-50 (putih) with proper Skeleton card
  - Button: btn preset-filled-primary-500 (hijau madrasah)
  - Inputs: Skeleton .input class with proper surface colors
-->
<div class="min-h-screen bg-surface-100-900 px-4 py-6 sm:px-8 flex items-center justify-center">
	<div class="grid w-full max-w-5xl overflow-hidden rounded-xl border border-primary-800 bg-surface-50-950 shadow-2xl lg:grid-cols-[1.05fr_0.95fr]">

		<!-- ═══ Left: Brand Panel (HARD-CODED Madrasah Green, NOT depending on theme) ═══ -->
		<section class="relative hidden p-10 lg:flex lg:flex-col lg:justify-between"
			style="background: linear-gradient(135deg, oklch(0.32 0.13 145) 0%, oklch(0.22 0.09 145) 100%); color: oklch(0.96 0.025 145);"
		>
			<!-- Subtle pattern overlay -->
			<div class="absolute inset-0 opacity-10" style="background-image: radial-gradient(circle at 1px 1px, white 1px, transparent 0); background-size: 24px 24px;"></div>

			<div class="z-10 flex h-full flex-col justify-between gap-10">
				<div>
					<a href="/" class="inline-flex items-center gap-2.5 rounded-base border px-3 py-2 text-sm font-semibold transition"
						style="background-color: rgba(255,255,255,0.1); border-color: rgba(255,255,255,0.2); color: oklch(0.96 0.025 145);"
					>
						<img src="/brand/logo-kemenag-icon-64.png" alt="Logo Kemenag" class="h-6 w-6 rounded bg-white p-0.5" />
						<span>Kembali ke website publik</span>
					</a>
					<div class="mt-14 max-w-md">
						<p class="text-xs font-bold uppercase tracking-[0.3em]" style="color: oklch(0.85 0.12 145);">Sistem Administrasi Sekolah</p>
						<h1 class="mt-4 text-[2.4rem] font-bold leading-[1.1]" style="color: oklch(0.98 0.01 145);">
							Ruang kerja digital MTs Negeri 2 Kolaka Utara.
						</h1>
						<p class="mt-5 text-base leading-7" style="color: oklch(0.88 0.03 145);">
							Akses internal untuk akademik, CBT, tata usaha, Kehadiran, kesiswaan, dan publikasi website resmi madrasah.
						</p>
					</div>
				</div>

				<!-- Feature chips -->
				<div class="grid gap-3 sm:grid-cols-3">
					<div class="rounded-base p-4" style="background-color: rgba(255,255,255,0.08); border: 1px solid rgba(255,255,255,0.15);">
						<p class="text-[10px] font-bold uppercase tracking-[0.2em]" style="color: oklch(0.82 0.13 145);">Akses</p>
						<p class="mt-1.5 text-sm font-bold" style="color: oklch(0.98 0.01 145);">Admin &amp; Guru</p>
					</div>
					<div class="rounded-base p-4" style="background-color: rgba(255,255,255,0.08); border: 1px solid rgba(255,255,255,0.15);">
						<p class="text-[10px] font-bold uppercase tracking-[0.2em]" style="color: oklch(0.82 0.13 145);">Sesi</p>
						<p class="mt-1.5 text-sm font-bold" style="color: oklch(0.98 0.01 145);">Cookie aman</p>
					</div>
					<div class="rounded-base p-4" style="background-color: rgba(255,255,255,0.08); border: 1px solid rgba(255,255,255,0.15);">
						<p class="text-[10px] font-bold uppercase tracking-[0.2em]" style="color: oklch(0.82 0.13 145);">Audit</p>
						<p class="mt-1.5 text-sm font-bold" style="color: oklch(0.98 0.01 145);">Tercatat</p>
					</div>
				</div>
			</div>
		</section>

		<!-- ═══ Right: Login Form (PUTIH BERSIH) ═══ -->
		<section class="flex items-center justify-center p-6 sm:p-8 lg:p-10 bg-surface-50-950">
			<div class="w-full max-w-sm">

				<!-- Mobile brand header -->
				<div class="mb-6 flex items-center gap-3 lg:hidden">
					<img src="/brand/logo-kemenag-icon-64.png" alt="Logo Kemenag" class="h-10 w-10 rounded-base border border-primary-200" />
					<div>
						<p class="text-sm font-bold text-surface-950-50">MTs Negeri 2 Kolaka Utara</p>
						<p class="text-xs text-primary-600-400">Sistem Administrasi Sekolah</p>
					</div>
				</div>

				<!-- Heading -->
				<div class="mb-5">
					<p class="text-xs font-bold uppercase tracking-[0.25em] text-primary-600-400">Login Internal</p>
					<h2 class="mt-2 text-2xl font-bold text-surface-950-50">Masuk ke akun kerja</h2>
					<p class="mt-1.5 text-sm text-surface-600-400">
						Gunakan akun resmi yang diterbitkan sekolah. Hubungi admin jika akses bermasalah.
					</p>
				</div>

				<!-- Skeleton Card -->
				<div class="card border border-surface-200-800 bg-surface-50-950 p-5 shadow-lg">
					<header class="mb-4">
						<h3 class="text-base font-bold text-surface-950-50">Autentikasi pengguna</h3>
						<p class="text-xs text-surface-500-400">Akses disesuaikan dengan peran akun.</p>
					</header>

					<form method="POST" class="space-y-4">
						{#if form?.error}
							<div id="login-error" class="rounded-base border border-error-500/30 bg-error-500/10 px-3 py-2.5 text-sm font-semibold text-error-300">
								{form.error}
							</div>
						{/if}

						<!-- Username -->
						<label class="label">
							<span class="text-sm font-semibold text-surface-700-300">Username</span>
							<input
								id="username"
								name="username"
								type="text"
								autocomplete="username"
								placeholder="nama_pengguna"
								aria-describedby={form?.error ? 'login-error' : undefined}
								class="input border border-surface-300-700 bg-surface-50-950 text-base h-11 mt-1.5 focus:border-primary-500 focus:ring-2 focus:ring-primary-500/30"
								required
							/>
						</label>

						<!-- Password -->
						<label class="label">
							<span class="text-sm font-semibold text-surface-700-300">Password</span>
							<div class="input-group grid-cols-[1fr_auto] mt-1.5">
								<input
									id="password"
									name="password"
									type={passwordShown ? 'text' : 'password'}
									autocomplete="current-password"
									placeholder="••••••••"
									aria-describedby={form?.error ? 'login-error' : undefined}
									class="input border border-surface-300-700 bg-surface-50-950 text-base h-11 focus:border-primary-500 focus:ring-2 focus:ring-primary-500/30"
									required
								/>
								<button
									type="button"
									onclick={() => (passwordShown = !passwordShown)}
									class="btn preset-tonal-surface border border-surface-300-700 border-l-0 px-3"
									aria-label={passwordShown ? 'Sembunyikan password' : 'Lihat password'}
									aria-pressed={passwordShown}
								>
									<span class="text-sm font-medium">{passwordShown ? 'Sembunyi' : 'Lihat'}</span>
								</button>
							</div>
						</label>

						<!-- Submit (Skeleton btn preset-filled = HIJAU MADRASAH) -->
						<button
							type="submit"
							disabled={pending}
							class="btn preset-filled-primary-500 w-full h-11 text-base font-bold border border-primary-500 disabled:opacity-60 hover:preset-filled-primary-600 transition"
						>
							<span>{pending ? '⏳ Memproses…' : 'Masuk'}</span>
						</button>
					</form>
				</div>

				<!-- Security notice -->
				<div class="mt-4 alert border border-warning-500/30 bg-warning-500/10 p-3 text-xs leading-5 text-warning-700-300">
					<span>Jangan gunakan perangkat bersama tanpa logout. Aktivitas login/logout/refresh dicatat untuk audit.</span>
				</div>

				<p class="mt-4 text-center text-[11px] text-surface-500-400" aria-label={appAttribution.loginLabel}>
					{appAttribution.loginLabel}
				</p>
			</div>
		</section>
	</div>
</div>
