<script lang="ts">
	import type { ActionData } from './$types';
	import { appAttribution } from '$lib/branding';
	import { navigating } from '$app/state';

	let { form }: { form: ActionData } = $props();
	let pending = $derived(!!navigating.to);
	let passwordShown = $state(false);
</script>

<svelte:head><title>Masuk — MTs Negeri 2 Kolaka Utara</title></svelte:head>

<div class="min-h-screen bg-base-200 flex items-center justify-center px-4 py-6 sm:px-8">
	<div class="grid w-full max-w-4xl overflow-hidden rounded-xl border border-base-300 bg-base-100 shadow-xl lg:grid-cols-[1fr_1fr]">

		<!-- ═══ Left: Brand Panel (Hijau Kemenag solid) ═══ -->
		<section class="relative hidden p-10 lg:flex lg:flex-col lg:justify-between"
			style="background: linear-gradient(135deg, oklch(0.32 0.13 145) 0%, oklch(0.22 0.09 145) 100%); color: oklch(0.96 0.025 145);"
		>
			<div class="absolute inset-0 opacity-10" style="background-image: radial-gradient(circle at 1px 1px, white 1px, transparent 0); background-size: 24px 24px;"></div>

			<div class="z-10 flex h-full flex-col justify-between gap-10">
				<div>
					<a href="/" class="inline-flex items-center gap-2.5 rounded-lg border px-3 py-2 text-sm font-semibold transition hover:bg-white/10"
						style="border-color: rgba(255,255,255,0.2);"
					>
						<img src="/brand/logo-kemenag-icon-64.png" alt="Logo Kemenag" class="h-6 w-6 rounded bg-white p-0.5" />
						<span>Kembali ke website publik</span>
					</a>
					<div class="mt-14 max-w-md">
						<p class="text-xs font-bold uppercase tracking-[0.3em]" style="color: oklch(0.85 0.12 145);">Sistem Administrasi Sekolah</p>
						<h1 class="mt-4 text-[2.4rem] font-bold leading-[1.1] text-white">
							Ruang kerja digital MTs Negeri 2 Kolaka Utara.
						</h1>
						<p class="mt-5 text-base leading-7" style="color: oklch(0.88 0.03 145);">
							Akses internal untuk akademik, CBT, tata usaha, Kehadiran, kesiswaan, dan publikasi website resmi madrasah.
						</p>
					</div>
				</div>

				<div class="grid gap-3 sm:grid-cols-3">
					<div class="rounded-lg p-4" style="background-color: rgba(255,255,255,0.08); border: 1px solid rgba(255,255,255,0.15);">
						<p class="text-[10px] font-bold uppercase tracking-[0.2em]" style="color: oklch(0.82 0.13 145);">Akses</p>
						<p class="mt-1.5 text-sm font-bold text-white">Admin &amp; Guru</p>
					</div>
					<div class="rounded-lg p-4" style="background-color: rgba(255,255,255,0.08); border: 1px solid rgba(255,255,255,0.15);">
						<p class="text-[10px] font-bold uppercase tracking-[0.2em]" style="color: oklch(0.82 0.13 145);">Sesi</p>
						<p class="mt-1.5 text-sm font-bold text-white">Cookie aman</p>
					</div>
					<div class="rounded-lg p-4" style="background-color: rgba(255,255,255,0.08); border: 1px solid rgba(255,255,255,0.15);">
						<p class="text-[10px] font-bold uppercase tracking-[0.2em]" style="color: oklch(0.82 0.13 145);">Audit</p>
						<p class="mt-1.5 text-sm font-bold text-white">Tercatat</p>
					</div>
				</div>
			</div>
		</section>

		<!-- ═══ Right: Login Form ═══ -->
		<section class="flex items-center justify-center p-6 sm:p-8 lg:p-10 bg-base-100">
			<div class="w-full max-w-sm">

				<!-- Mobile brand header -->
				<div class="mb-6 flex items-center gap-3 lg:hidden">
					<img src="/brand/logo-kemenag-icon-64.png" alt="Logo Kemenag" class="h-10 w-10 rounded-lg border border-base-300" />
					<div>
						<p class="text-sm font-bold text-base-content">MTs Negeri 2 Kolaka Utara</p>
						<p class="text-xs text-base-content/70">Sistem Administrasi Sekolah</p>
					</div>
				</div>

				<!-- Heading -->
				<div class="mb-5">
					<p class="text-xs font-bold uppercase tracking-[0.25em] text-primary">Login Internal</p>
					<h2 class="mt-2 text-2xl font-bold text-base-content">Masuk ke akun kerja</h2>
					<p class="mt-1.5 text-sm text-base-content/70">
						Gunakan akun resmi yang diterbitkan sekolah. Hubungi admin jika akses bermasalah.
					</p>
				</div>

				<!-- Form Card -->
				<div class="rounded-xl border border-base-300 bg-base-100 p-5 shadow-sm">
					<header class="mb-4">
						<h3 class="text-base font-bold text-base-content">Autentikasi pengguna</h3>
						<p class="text-xs text-base-content/70">Akses disesuaikan dengan peran akun.</p>
					</header>

					<form method="POST" class="space-y-4">
						{#if form?.error}
							<div class="rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2.5 text-sm font-semibold text-destructive">
								{form.error}
							</div>
						{/if}

						<!-- Username -->
						<label class="block">
							<span class="text-sm font-semibold text-base-content">Username</span>
							<input
								id="username"
								name="username"
								type="text"
								autocomplete="username"
								placeholder="nama_pengguna"
								aria-describedby={form?.error ? 'login-error' : undefined}
								class="mt-1.5 flex h-11 w-full rounded-lg border border-base-300 bg-base-200 px-3 py-2 text-base text-base-content placeholder:text-base-content/70 focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary"
								required
							/>
						</label>

						<!-- Password -->
						<label class="block">
							<span class="text-sm font-semibold text-base-content">Password</span>
							<div class="mt-1.5 flex">
								<input
									id="password"
									name="password"
									type={passwordShown ? 'text' : 'password'}
									autocomplete="current-password"
									placeholder="••••••••"
									aria-describedby={form?.error ? 'login-error' : undefined}
									class="flex h-11 w-full rounded-l-lg border border-base-300 bg-base-200 px-3 py-2 text-base text-base-content placeholder:text-base-content/70 focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary"
									required
								/>
								<button
									type="button"
									onclick={() => (passwordShown = !passwordShown)}
									class="flex h-11 items-center justify-center rounded-r-lg border border-l-0 border-base-300 bg-base-200 px-3 text-sm font-medium text-base-content/70 hover:bg-base-200 hover:text-base-content transition"
									aria-label={passwordShown ? 'Sembunyikan password' : 'Lihat password'}
									aria-pressed={passwordShown}
								>
									{passwordShown ? 'Sembunyi' : 'Lihat'}
								</button>
							</div>
						</label>

						<!-- Submit -->
						<button
							type="submit"
							disabled={pending}
							class="flex w-full h-11 items-center justify-center rounded-lg bg-primary text-primary-content text-base font-bold shadow-sm transition hover:bg-primary/90 disabled:opacity-60"
						>
							{pending ? '⏳ Memproses…' : 'Masuk'}
						</button>
					</form>
				</div>

				<!-- Security notice -->
				<div class="mt-4 rounded-lg border border-amber-500/30 bg-amber-50 px-3 py-2.5 text-xs leading-5 text-amber-700">
					Jangan gunakan perangkat bersama tanpa logout. Aktivitas login/logout/refresh dicatat untuk audit.
				</div>

				<p class="mt-4 text-center text-[11px] text-base-content/70" aria-label={appAttribution.loginLabel}>
					{appAttribution.loginLabel}
				</p>
			</div>
		</section>
	</div>
</div>