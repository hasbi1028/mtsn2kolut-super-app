<script lang="ts">
	import type { ActionData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import { navigating } from '$app/state';

	let { form }: { form: ActionData } = $props();
	let pending = $derived(!!navigating.to);
</script>

<svelte:head><title>Masuk — MTs Negeri 2 Kolaka Utara</title></svelte:head>

<div class="min-h-screen overflow-hidden bg-[radial-gradient(circle_at_top_left,_#d8f7e6,_transparent_34%),linear-gradient(135deg,_#f8fafc_0%,_#eefdf4_48%,_#fff7ed_100%)] px-4 py-6 text-slate-900 sm:px-6 lg:px-10">
	<div class="mx-auto flex min-h-[calc(100vh-3rem)] w-full max-w-6xl items-center">
		<div class="grid w-full overflow-hidden rounded-[2rem] border border-emerald-100 bg-white/82 shadow-[0_24px_80px_rgba(15,23,42,0.10)] backdrop-blur lg:grid-cols-[1.08fr_0.92fr]">
			<section class="relative hidden min-h-[40rem] overflow-hidden bg-emerald-950 p-10 text-white lg:block">
				<div class="absolute -left-24 top-10 h-72 w-72 rounded-full bg-emerald-500/25 blur-3xl"></div>
				<div class="absolute bottom-0 right-0 h-96 w-96 rounded-full bg-lime-300/20 blur-3xl"></div>
				<div class="absolute inset-x-10 bottom-10 top-36 rounded-[2rem] border border-white/10 bg-[linear-gradient(135deg,_rgba(255,255,255,0.14),_rgba(255,255,255,0.03))]"></div>

				<div class="relative z-10 flex h-full flex-col justify-between">
					<div>
						<a href="/" class="inline-flex items-center gap-3 rounded-full border border-white/15 bg-white/10 px-4 py-2 text-sm font-medium text-emerald-50 transition hover:bg-white/15">
							<span class="flex h-8 w-8 items-center justify-center rounded-full bg-white text-xs font-bold text-emerald-800">MTs</span>
							Kembali ke website publik
						</a>
						<div class="mt-16 max-w-xl">
							<p class="text-xs font-semibold uppercase tracking-[0.34em] text-emerald-200">Sistem Administrasi Sekolah</p>
							<h1 class="mt-5 text-5xl font-semibold leading-tight tracking-tight">
								Ruang kerja digital MTs Negeri 2 Kolaka Utara.
							</h1>
							<p class="mt-5 max-w-lg text-base leading-7 text-emerald-50/80">
								Akses internal untuk akademik, CBT, tata usaha, PUSAKA, kesiswaan, dan publikasi website resmi madrasah.
							</p>
						</div>
					</div>

					<div class="grid gap-3 sm:grid-cols-3">
						<div class="rounded-2xl border border-white/10 bg-white/10 p-4">
							<p class="text-[10px] font-semibold uppercase tracking-[0.24em] text-emerald-200">Akses</p>
							<p class="mt-2 text-sm font-semibold text-white">Admin & Guru</p>
						</div>
						<div class="rounded-2xl border border-white/10 bg-white/10 p-4">
							<p class="text-[10px] font-semibold uppercase tracking-[0.24em] text-emerald-200">Sesi</p>
							<p class="mt-2 text-sm font-semibold text-white">Cookie aman</p>
						</div>
						<div class="rounded-2xl border border-white/10 bg-white/10 p-4">
							<p class="text-[10px] font-semibold uppercase tracking-[0.24em] text-emerald-200">Audit</p>
							<p class="mt-2 text-sm font-semibold text-white">Tercatat</p>
						</div>
					</div>
				</div>
			</section>

			<section class="flex min-h-[calc(100vh-3rem)] items-center justify-center p-5 sm:p-8 lg:min-h-[40rem] lg:p-12">
				<div class="w-full max-w-md">
					<div class="mb-8 flex items-center justify-between gap-4 lg:hidden">
						<a href="/" class="flex items-center gap-3">
							<div class="flex h-11 w-11 items-center justify-center rounded-2xl bg-emerald-800 text-sm font-bold text-white shadow-sm">
								MTs
							</div>
							<div>
								<p class="text-sm font-semibold text-slate-900">MTs Negeri 2 Kolaka Utara</p>
								<p class="text-xs text-emerald-700">Sistem Administrasi Sekolah</p>
							</div>
						</a>
					</div>

					<div class="mb-6">
						<p class="text-xs font-semibold uppercase tracking-[0.28em] text-emerald-700">Login Internal</p>
						<h2 class="mt-3 text-3xl font-semibold tracking-tight text-slate-950">Masuk ke akun kerja</h2>
						<p class="mt-2 text-sm leading-6 text-slate-600">
							Gunakan akun resmi yang diterbitkan sekolah. Jika akses bermasalah, hubungi admin madrasah.
						</p>
					</div>

					<Card.Root class="border-slate-200 bg-white/92 shadow-sm">
						<Card.Header class="pb-3">
							<Card.Title class="text-base">Autentikasi pengguna</Card.Title>
							<Card.Description>Akses disesuaikan dengan peran akun.</Card.Description>
						</Card.Header>
						<Card.Content>
							<form method="POST" class="space-y-4">
								{#if form?.error}
									<div id="login-error" class="rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
										{form.error}
									</div>
								{/if}

								<div class="space-y-1.5">
									<label for="username" class="block text-sm font-medium text-slate-700">Username</label>
									<Input
										id="username"
										name="username"
										type="text"
										autocomplete="username"
										placeholder="nama_pengguna"
										aria-describedby={form?.error ? 'login-error' : undefined}
										required
									/>
								</div>

								<div class="space-y-1.5">
									<label for="password" class="block text-sm font-medium text-slate-700">Password</label>
									<Input
										id="password"
										name="password"
										type="password"
										autocomplete="current-password"
										placeholder="••••••••"
										aria-describedby={form?.error ? 'login-error' : undefined}
										required
									/>
								</div>

								<LoadingButton type="submit" class="w-full bg-emerald-800 hover:bg-emerald-900" loading={pending} loadingLabel="Memproses..." label="Masuk" />
							</form>
						</Card.Content>
					</Card.Root>

					<div class="mt-5 rounded-2xl border border-amber-100 bg-amber-50/80 px-4 py-3 text-sm leading-6 text-amber-900">
						Jangan gunakan perangkat bersama tanpa logout. Aktivitas penting seperti login, refresh, logout, dan revoke sesi dicatat untuk audit keamanan.
					</div>
				</div>
			</section>
		</div>
	</div>
</div>
