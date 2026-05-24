<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import RefreshCcwIcon from '@lucide/svelte/icons/refresh-ccw';
	import SaveIcon from '@lucide/svelte/icons/save';
	import UploadIcon from '@lucide/svelte/icons/upload';
	import RotateCcwIcon from '@lucide/svelte/icons/rotate-ccw';
	import { onMount } from 'svelte';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { toast } from '$lib/components/ui/sonner';
	import { readClientApiData } from '$lib/client/api';
	import { appAttribution, defaultBranding, normalizeBranding, versionedAsset, type BrandingSettings } from '$lib/branding';

	type Purpose = 'logo' | 'mark' | 'favicon' | 'apple_touch_icon' | 'pwa_icon_192' | 'pwa_icon_512' | 'formal_logo';
	type AssetCard = { purpose: Purpose; title: string; hint: string; field: keyof BrandingSettings; min: string };

	const assets: AssetCard[] = [
		{ purpose: 'logo', title: 'Logo Aplikasi', hint: 'Logo untuk header/login. PNG/JPEG maksimal 2 MB.', field: 'logo_url', min: 'min 128×48' },
		{ purpose: 'mark', title: 'Mark Madrasah', hint: 'Ikon persegi untuk sidebar kecil dan tab.', field: 'mark_url', min: 'persegi min 64×64' },
		{ purpose: 'favicon', title: 'Favicon', hint: 'Ikon browser. Disarankan PNG persegi.', field: 'favicon_url', min: 'persegi min 32×32' },
		{ purpose: 'apple_touch_icon', title: 'Apple Touch Icon', hint: 'Ikon iOS/home screen.', field: 'apple_touch_icon_url', min: 'persegi min 180×180' },
		{ purpose: 'pwa_icon_192', title: 'PWA Icon 192', hint: 'Ikon PWA Android ukuran menengah.', field: 'pwa_icon_192_url', min: 'persegi min 192×192' },
		{ purpose: 'pwa_icon_512', title: 'PWA Icon 512', hint: 'Ikon PWA resolusi besar.', field: 'pwa_icon_512_url', min: 'persegi min 512×512' },
		{ purpose: 'formal_logo', title: 'Logo Formal', hint: 'Logo untuk konteks formal; default tetap Kemenag.', field: 'formal_logo_url', min: 'min 128×128' }
	];

	let brandingPromise = $state<Promise<BrandingSettings> | null>(null);
	let branding = $state<BrandingSettings>({ ...defaultBranding });
	let saving = $state(false);
	let refreshBusy = $state(false);
	let busyPurpose = $state<string | null>(null);

	function errorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Branding belum dapat dimuat.';
	}

	async function fetchBranding() {
		const res = await fetch('/api/branding');
		return normalizeBranding(await readClientApiData<Partial<BrandingSettings>>(res, 'Branding gagal dimuat.'));
	}

	function loadBranding() {
		brandingPromise = fetchBranding().then((data) => {
			branding = data;
			return data;
		});
		return brandingPromise;
	}

	async function refreshBranding(showToast = false) {
		refreshBusy = true;
		try {
			await loadBranding();
			if (showToast) toast.success('Branding dimuat ulang.');
		} catch (error) {
			if (showToast) toast.error(errorMessage(error));
		} finally {
			refreshBusy = false;
		}
	}

	async function saveBranding() {
		saving = true;
		try {
			const res = await fetch('/api/branding', {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(branding)
			});
			branding = normalizeBranding(await readClientApiData<Partial<BrandingSettings>>(res, 'Branding gagal disimpan.'));
			brandingPromise = Promise.resolve(branding);
			await invalidateAll();
			toast.success('Branding aplikasi disimpan.');
		} catch (error) {
			toast.error(errorMessage(error));
		} finally {
			saving = false;
		}
	}

	async function uploadAsset(purpose: Purpose, input: HTMLInputElement) {
		const file = input.files?.[0];
		if (!file) return;
		if (!['image/png', 'image/jpeg'].includes(file.type)) {
			toast.error('Gunakan PNG atau JPEG agar dimensi bisa divalidasi aman.');
			input.value = '';
			return;
		}
		if (file.size > 2 * 1024 * 1024) {
			toast.error('Ukuran file maksimal 2 MB.');
			input.value = '';
			return;
		}
		busyPurpose = purpose;
		try {
			const form = new FormData();
			form.set('file', file);
			const res = await fetch(`/api/branding/assets/${purpose}`, { method: 'POST', body: form });
			const data = await readClientApiData<{ branding?: BrandingSettings }>(res, 'Upload aset branding gagal.');
			branding = normalizeBranding(data.branding ?? branding);
			brandingPromise = Promise.resolve(branding);
			await invalidateAll();
			toast.success('Aset branding diperbarui.');
		} catch (error) {
			toast.error(errorMessage(error));
		} finally {
			busyPurpose = null;
			input.value = '';
		}
	}

	async function resetAsset(purpose: Purpose) {
		if (!confirm('Reset aset ini ke default aplikasi?')) return;
		busyPurpose = purpose;
		try {
			const res = await fetch(`/api/branding/assets/${purpose}`, { method: 'DELETE' });
			branding = normalizeBranding(await readClientApiData<Partial<BrandingSettings>>(res, 'Reset aset branding gagal.'));
			brandingPromise = Promise.resolve(branding);
			await invalidateAll();
			toast.success('Aset branding direset ke default.');
		} catch (error) {
			toast.error(errorMessage(error));
		} finally {
			busyPurpose = null;
		}
	}

	onMount(() => void loadBranding());
</script>

<svelte:head><title>Branding Aplikasi - MTsN 2 Kolaka Utara</title></svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
		<div>
			<h1 class="text-lg font-semibold text-foreground">Branding Aplikasi</h1>
			<p class="text-sm text-muted-foreground">Atur logo, favicon, ikon PWA, nama singkat, dan warna tema aplikasi.</p>
		</div>
		<div class="flex flex-wrap gap-2">
			<Button href="/settings" variant="outline" size="sm"><ArrowLeftIcon class="mr-2 size-4" />Pengaturan</Button>
			<LoadingButton variant="outline" size="sm" loading={refreshBusy} loadingLabel="Memuat..." onclick={() => void refreshBranding(true)}><RefreshCcwIcon class="mr-2 size-4" />Refresh</LoadingButton>
			<LoadingButton size="sm" loading={saving} loadingLabel="Menyimpan..." onclick={() => void saveBranding()}><SaveIcon class="mr-2 size-4" />Simpan Identitas</LoadingButton>
		</div>
	</div>

	<AsyncContent promise={brandingPromise}>
		{#snippet pending()}
			<div class="grid gap-4 md:grid-cols-2"><Skeleton class="h-64" /><Skeleton class="h-64" /></div>
		{/snippet}
		{#snippet failed(error, reset)}
			<RecoveryPanel title="Branding belum tersaji" message={errorMessage(error)} onRetry={() => { reset?.(); loadBranding(); }} />
		{/snippet}
		{#snippet children()}
			<div class="grid gap-6 xl:grid-cols-[1fr_380px]">
				<div class="space-y-6">
					<Card.Root>
						<Card.Header><Card.Title class="text-base">Identitas Visual</Card.Title><Card.Description>Nama dan warna dipakai pada tab browser, manifest PWA, dan tampilan aplikasi.</Card.Description></Card.Header>
						<Card.Content class="grid gap-4 md:grid-cols-2">
							<label class="space-y-1 text-sm font-medium">Nama Aplikasi<Input bind:value={branding.app_name} /></label>
							<label class="space-y-1 text-sm font-medium">Nama Singkat<Input bind:value={branding.short_name} /></label>
							<label class="space-y-1 text-sm font-medium md:col-span-2">Tagline<Input bind:value={branding.tagline} /></label>
							<label class="space-y-1 text-sm font-medium">Warna Utama<Input type="color" bind:value={branding.primary_color} /></label>
							<label class="space-y-1 text-sm font-medium">Warna Theme Browser<Input type="color" bind:value={branding.theme_color} /></label>
						</Card.Content>
					</Card.Root>

					<div class="grid gap-4 md:grid-cols-2">
						{#each assets as asset (asset.purpose)}
							<Card.Root>
								<Card.Header class="pb-2"><Card.Title class="text-sm">{asset.title}</Card.Title><Card.Description>{asset.hint}</Card.Description></Card.Header>
								<Card.Content class="space-y-3">
									<div class="flex items-center gap-3 rounded-lg border bg-muted/30 p-3">
										<img src={versionedAsset(String(branding[asset.field]), branding.version)} alt={asset.title} class="size-14 rounded-md border bg-white object-contain p-1" />
										<div class="min-w-0 text-xs text-muted-foreground"><p>{asset.min}</p><p class="truncate">{branding[asset.field]}</p></div>
									</div>
									<div class="flex flex-wrap gap-2">
										<input id={`asset-${asset.purpose}`} type="file" accept="image/png,image/jpeg" class="hidden" onchange={(event) => void uploadAsset(asset.purpose, event.currentTarget)} />
										<LoadingButton variant="outline" size="sm" loading={busyPurpose === asset.purpose} onclick={() => document.getElementById(`asset-${asset.purpose}`)?.click()}><UploadIcon class="mr-2 size-4" />Upload</LoadingButton>
										<Button variant="ghost" size="sm" onclick={() => void resetAsset(asset.purpose)}><RotateCcwIcon class="mr-2 size-4" />Reset</Button>
									</div>
								</Card.Content>
							</Card.Root>
						{/each}
					</div>
				</div>

				<div class="space-y-4">
					<Card.Root class="h-fit">
						<Card.Header><Card.Title class="text-base">Pratinjau</Card.Title><Card.Description>Simulasi sidebar, browser tab, dan ikon PWA.</Card.Description></Card.Header>
						<Card.Content class="space-y-4">
							<div class="rounded-xl border p-4">
								<div class="flex items-center gap-3"><img src={versionedAsset(branding.mark_url, branding.version)} alt="Mark" class="size-10 rounded-lg bg-white object-contain p-1" /><div><p class="font-semibold">{branding.short_name}</p><p class="text-xs text-muted-foreground">{branding.tagline}</p></div></div>
							</div>
							<div class="rounded-xl border p-4">
								<p class="mb-2 text-xs font-medium text-muted-foreground">Tab Browser</p>
								<div class="flex items-center gap-2 rounded-full bg-muted px-3 py-2 text-sm"><img src={versionedAsset(branding.favicon_url, branding.version)} alt="Favicon" class="size-4" />{branding.short_name}</div>
							</div>
							<div class="rounded-xl border p-4 text-sm text-muted-foreground">Manifest PWA otomatis memakai nama singkat, warna theme, dan ikon 192/512 dengan cache-busting version <code>{branding.version}</code>.</div>
						</Card.Content>
					</Card.Root>

					<Card.Root class="h-fit border-border/80 bg-muted/20">
						<Card.Header>
							<Card.Title class="text-base">Tentang Aplikasi</Card.Title>
							<Card.Description>Atribusi internal yang tampil halus tanpa mengubah identitas resmi madrasah.</Card.Description>
						</Card.Header>
						<Card.Content class="space-y-3 text-sm">
							<div class="rounded-xl border bg-background/70 p-4">
								<p class="font-semibold text-foreground">{appAttribution.productName}</p>
								<p class="mt-1 text-xs leading-5 text-muted-foreground">{appAttribution.description}</p>
							</div>
							<div class="flex items-center justify-between gap-3 rounded-xl border bg-background/70 px-4 py-3">
								<span class="text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">Pengembang</span>
								<span class="text-sm font-semibold text-foreground">{appAttribution.developerName}</span>
							</div>
							<p class="text-xs leading-5 text-muted-foreground">{appAttribution.formalDeveloperLabel}. Teks atribusi hanya ditampilkan pada area UI non-dokumen resmi.</p>
						</Card.Content>
					</Card.Root>
				</div>
			</div>
		{/snippet}
	</AsyncContent>
</div>
