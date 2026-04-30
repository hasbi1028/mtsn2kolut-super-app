<script lang="ts">
	import type { ActionData } from './$types';
	import * as Card from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { navigating } from '$app/state';

	let { form }: { form: ActionData } = $props();
	let pending = $derived(!!navigating.to);
</script>

<svelte:head><title>Masuk — MTs Negeri 2 Kolaka Utara</title></svelte:head>

<div class="flex min-h-screen items-center justify-center bg-slate-50 px-4">
	<div class="w-full max-w-sm">
		<div class="mb-6 flex flex-col items-center gap-2">
			<div class="flex h-12 w-12 items-center justify-center rounded-xl bg-green-700 text-white text-sm font-bold shadow-sm">
				MTs
			</div>
			<div class="text-center">
				<p class="text-sm font-semibold text-slate-800">MTs Negeri 2 Kolaka Utara</p>
				<p class="text-xs text-slate-400">Sistem Administrasi Sekolah</p>
			</div>
		</div>

		<Card.Root class="border-slate-200 shadow-sm">
			<Card.Header class="pb-4">
				<Card.Title class="text-center text-base">Masuk ke Akun Anda</Card.Title>
			</Card.Header>
			<Card.Content>
				<form method="POST" class="space-y-4">
					{#if form?.error}
						<div class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
							{form.error}
						</div>
					{/if}

					<div class="space-y-1.5">
						<label for="username" class="block text-sm font-medium text-slate-700">Username</label>
						<Input id="username" name="username" type="text" autocomplete="username"
							placeholder="nama_pengguna" required />
					</div>

					<div class="space-y-1.5">
						<label for="password" class="block text-sm font-medium text-slate-700">Password</label>
						<Input id="password" name="password" type="password" autocomplete="current-password"
							placeholder="••••••••" required />
					</div>

					<Button type="submit" class="w-full" disabled={pending}>
						{pending ? 'Memproses…' : 'Masuk'}
					</Button>
				</form>
			</Card.Content>
		</Card.Root>
	</div>
</div>
