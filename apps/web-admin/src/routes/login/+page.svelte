<script lang="ts">
	import type { ActionData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import PasswordInput from '$lib/components/PasswordInput.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import { appAttribution } from '$lib/branding';
	import { navigating } from '$app/state';

	let { form }: { form: ActionData } = $props();
	let pending = $derived(!!navigating.to);
</script>

<svelte:head><title>Masuk — MTs Negeri 2 Kolaka Utara</title></svelte:head>

<div class="min-h-screen overflow-hidden bg-[var(--background)] px-4 py-6 text-[var(--foreground)] sm:px-6 lg:px-10">
	<div class="mx-auto flex min-h-[calc(100vh-3rem)] w-full max-w-6xl items-center">
		<div class="grid w-full overflow-hidden rounded-2xl border border-[var(--border)] bg-[var(--card)]/90 shadow-lg lg:grid-cols-[1.08fr_0.92fr]">

			<!-- Left panel — branding / info (desktop only) -->
			<section class="relative hidden min-h-[40rem] overflow-hidden bg-[oklch(0.25_0.08_145)] p-10 text-white lg:block">
				<div class="relative z-10 flex h-full flex-col justify-between">
					<div>
						<a href="/" class="inline-flex items-center gap-3 rounded-full border border-white/15 bg-white/10 px-4 py-2 text-sm font-medium text-white transition hover:bg-white/15">
							<span class="flex h-8 w-8 items-center justify-center overflow-hidden rounded-full bg-white p-1 shadow-sm">
								<img src="/brand/logo-kemenag-icon-64.png" alt="Logo Kemenag" class="h-full w-full object-contain" />
							</span>
							Kembali ke website publik
						</a>
						<div class="mt-16 max-w-xl">
							<p class="text-xs font-semibold uppercase tracking-[0.34em] text-green-200">Sistem Administrasi Sekolah</p>
							<h1 class="mt-5 text-5xl font-semibold leading-tight tracking-tight">
								Ruang kerja digital MTs Negeri 2 Kolaka Utara.
							</h1>
							<p class="mt-5 max-w-lg text-base leading-7 text-white/80">
								Akses internal untuk akademik, CBT, tata usaha, Kehadiran, kesiswaan, dan publikasi website resmi madrasah.
							</p>
						</div>
					</div>

					<div class="grid gap-3 sm:grid-cols-3">
						<div class="rounded-2xl border border-white/10 bg-white/10 p-4">
							<p class="text-[10px] font-semibold uppercase tracking-[0.24em] text-green-200">Akses</p>
							<p class="mt-2 text-sm font-semibold text-white">Admin & Guru</p>
						</div>
						<div class="rounded-2xl border border-white/10 bg-white/10 p-4">
							<p class="text-[10px] font-semibold uppercase tracking-[0.24em] text-green-200">Sesi</p>
							<p class="mt-2 text-sm font-semibold text-white">Cookie aman</p>
						</div>
						<div class="rounded-2xl border border-white/10 bg-white/10 p-4">
							<p class="text-[10px] font-semibold uppercase tracking-[0.24em] text-green-200">Audit</p>
							<p class="mt-2 text-sm font-semibold text-white">Tercatat</p>
						</div>
					</div>
				</div>
			</section>

			<!-- Right panel — login form -->
			<section class="flex min-h-[calc(100vh-3rem)] items-center justify-center p-5 sm:p-8 lg:min-h-[40rem] lg:p-12">
				<div class="w-full max-w-md">
					<!-- Mobile brand header -->
					<div class="mb-8 flex items-center justify-between gap-4 lg:hidden">
						<a href="/" class="flex items-center gap-3">
							<span class="flex h-11 w-11 items-center justify-center overflow-hidden rounded-2xl bg-white p-1.5 shadow-sm ring-1 ring-[var(--border)]">
								<img src="/brand/logo-kemenag-icon-64.png" alt="Logo Kemenag" class="h-full w-full object-contain" />
							</span>
							<div>
								<p class="text-sm font-semibold text-[var(--foreground)]">MTs Negeri 2 Kolaka Utara</p>
								<p class="text-xs text-[var(--primary)]">Sistem Administrasi Sekolah</p>
							</div>
						</a>
					</div>

					<!-- Heading -->
					<div class="mb-6">
						<p class="text-xs font-semibold uppercase tracking-[0.28em] text-[var(--primary)]">Login Internal</p>
						<h2 class="mt-3 text-3xl font-semibold tracking-tight text-[var(--foreground)]">Masuk ke akun kerja</h2>
						<p class="mt-2 text-sm leading-6 text-[var(--muted-foreground)]">
							Gunakan akun resmi yang diterbitkan sekolah. Jika akses bermasalah, hubungi admin madrasah.
						</p>
					</div>

					<!-- Card -->
					<Card.Root class="border-[var(--border)] bg-[var(--card)]/92 shadow-sm">
						<Card.Header class="pb-3">
							<Card.Title class="text-base">Autentikasi pengguna</Card.Title>
							<Card.Description>Akses disesuaikan dengan peran akun.</Card.Description>
						</Card.Header>
						<Card.Content>
							<form method="POST" class="space-y-4">
								{#if form?.error}
									<div id="login-error" class="rounded-xl border border-[var(--error)]/30 bg-[var(--error)]/10 px-4 py-3 text-sm font-medium text-[var(--error)]">
										{form.error}
									</div>
								{/if}

								<div class="space-y-1.5">
									<label for="username" class="block text-sm font-medium text-[var(--foreground)]">Username</label>
									<Input
										id="username"
										name="username"
										type="text"
										autocomplete="username"
										placeholder="nama_pengguna"
										aria-describedby={form?.error ? 'login-error' : undefined}
										class="h-11 text-base"
										required
									/>
								</div>

								<div class="space-y-1.5">
									<label for="password" class="block text-sm font-medium text-[var(--foreground)]">Password</label>
									<PasswordInput
										id="password"
										name="password"
										autocomplete="current-password"
										placeholder="••••••••"
										aria-describedby={form?.error ? 'login-error' : undefined}
										class="h-11 text-base"
										required
									/>
								</div>

								<LoadingButton
									type="submit"
									class="w-full h-11 text-base font-semibold"
									loading={pending}
									loadingLabel="Memproses..."
									label="Masuk"
								/>
							</form>
						</Card.Content>
					</Card.Root>

					<!-- Security notice -->
					<div class="mt-5 rounded-2xl border border-[var(--warning)]/30 bg-[var(--warning)]/10 px-4 py-3 text-sm leading-6 text-[var(--warning)]">
						Jangan gunakan perangkat bersama tanpa logout. Aktivitas penting seperti login, refresh, logout, dan revoke sesi dicatat untuk audit keamanan.
					</div>

					<p class="mt-4 text-center text-xs leading-5 text-[var(--muted-foreground)]" aria-label={`Atribusi aplikasi ${appAttribution.loginLabel}`}>
						{appAttribution.loginLabel}
					</p>
				</div>
			</section>
		</div>
	</div>
</div>
