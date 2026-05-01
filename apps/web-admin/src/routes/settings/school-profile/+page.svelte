<script lang="ts">
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import RefreshCcwIcon from '@lucide/svelte/icons/refresh-ccw';
	import SaveIcon from '@lucide/svelte/icons/save';
	import { onMount } from 'svelte';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { Textarea } from '$lib/components/ui/textarea';
	import { toast } from '$lib/components/ui/sonner';
	import { readClientApiData } from '$lib/client/api';
	import {
		defaultSchoolProfile,
		fetchSchoolProfile,
		schoolAddressLine,
		type SchoolProfile,
	} from '$lib/school-profile';

	function emptyProfile(): SchoolProfile {
		return { ...defaultSchoolProfile };
	}

	let profilePromise = $state<Promise<SchoolProfile> | null>(null);
	let profile = $state<SchoolProfile>(emptyProfile());
	let saving = $state(false);
	let refreshBusy = $state(false);

	function normalizeProfile(value: Partial<SchoolProfile> | null | undefined): SchoolProfile {
		return { ...emptyProfile(), ...(value ?? {}) };
	}

	function profileErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		if (typeof error === 'string' && error.trim()) return error;
		return 'Profil madrasah belum dapat dimuat.';
	}

	async function fetchProfile(): Promise<SchoolProfile> {
		return fetchSchoolProfile();
	}

	function loadProfile() {
		profilePromise = fetchProfile().then((nextProfile) => {
			profile = nextProfile;
			return nextProfile;
		});
		return profilePromise;
	}

	async function refreshProfile(showFailureToast = false) {
		if (!profilePromise) {
			await loadProfile();
			return;
		}
		try {
			const nextProfile = await fetchProfile();
			profile = nextProfile;
			profilePromise = Promise.resolve(nextProfile);
		} catch (error) {
			profilePromise = Promise.resolve(profile);
			if (showFailureToast) toast.error(profileErrorMessage(error));
		}
	}

	async function refreshProfileAction() {
		refreshBusy = true;
		try {
			await refreshProfile(true);
		} catch (error) {
			toast.error(profileErrorMessage(error));
		} finally {
			refreshBusy = false;
		}
	}

	function retryProfile(reset?: () => void) {
		reset?.();
		loadProfile();
	}

	function handleProfileRenderError(error: unknown, reset: () => void) {
		console.error('School profile render failed', error);
		reset();
	}

	async function saveProfile() {
		saving = true;
		try {
			const res = await fetch('/api/school-profile', {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(profile),
			});
			const data = await readClientApiData<Partial<SchoolProfile>>(res, 'Profil madrasah gagal disimpan.');
			profile = normalizeProfile(data);
			profilePromise = Promise.resolve(profile);
			toast.success('Profil madrasah disimpan.');
		} catch (error) {
			toast.error(profileErrorMessage(error));
		} finally {
			saving = false;
		}
	}

	onMount(() => {
		void loadProfile();
	});
</script>

<svelte:head><title>Profil Madrasah - MTsN 2 Kolaka Utara</title></svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
		<div>
			<h1 class="text-lg font-semibold text-slate-800">Profil Madrasah</h1>
			<p class="text-sm text-slate-500">Identitas resmi untuk kop surat, paket cetak, dan tanda tangan kepala madrasah.</p>
		</div>
		<div class="flex flex-wrap gap-2">
			<Button href="/settings" variant="outline" size="sm">
				<ArrowLeftIcon class="mr-2 size-4" />
				Pengaturan
			</Button>
			<LoadingButton variant="outline" size="sm" loading={refreshBusy} loadingLabel="Memuat..." onclick={() => void refreshProfileAction()}>
				<RefreshCcwIcon class="mr-2 size-4" />
				Refresh
			</LoadingButton>
			<LoadingButton onclick={() => void saveProfile()} loading={saving} loadingLabel="Menyimpan..." size="sm">
				<SaveIcon class="mr-2 size-4" />
				Simpan Profil
			</LoadingButton>
		</div>
	</div>

	<AsyncContent promise={profilePromise} onerror={handleProfileRenderError}>
		{#snippet pending()}
			<div class="grid gap-6 xl:grid-cols-[1fr_360px]">
				<Card.Root class="border-slate-200">
					<Card.Content class="space-y-3 p-4">
						{#each Array.from({ length: 10 }) as _, index (`school-profile-skeleton-${index}`)}
							<Skeleton class="h-10 w-full" />
						{/each}
					</Card.Content>
				</Card.Root>
				<Skeleton class="h-80 w-full" />
			</div>
		{/snippet}
		{#snippet failed(error, reset)}
			<RecoveryPanel title="Profil Madrasah Belum Tersaji" message={profileErrorMessage(error)} onRetry={() => retryProfile(reset)} />
		{/snippet}
		{#snippet children(_profile)}
		<div class="grid gap-6 xl:grid-cols-[1fr_380px]">
			<Card.Root class="border-slate-200">
				<Card.Header class="pb-2">
					<Card.Title class="text-base">Identitas Satuan Kerja</Card.Title>
					<Card.Description>Data ini dipakai sebagai sumber kop resmi pada permukaan cetak aplikasi.</Card.Description>
				</Card.Header>
				<Card.Content class="space-y-5">
					<div class="grid gap-3 md:grid-cols-2">
						<div class="md:col-span-2">
							<label for="school-name" class="text-sm font-medium">Nama Madrasah</label>
							<Input id="school-name" bind:value={profile.name} />
						</div>
						<div>
							<label for="school-nsm" class="text-sm font-medium">NSM</label>
							<Input id="school-nsm" bind:value={profile.nsm} />
						</div>
						<div>
							<label for="school-npsn" class="text-sm font-medium">NPSN</label>
							<Input id="school-npsn" bind:value={profile.npsn} />
						</div>
						<div>
							<label for="school-ministry" class="text-sm font-medium">Baris Kementerian</label>
							<Input id="school-ministry" bind:value={profile.ministry_line} />
						</div>
						<div>
							<label for="school-office" class="text-sm font-medium">Baris Kantor</label>
							<Input id="school-office" bind:value={profile.office_line} />
						</div>
						<div class="md:col-span-2">
							<label for="school-address" class="text-sm font-medium">Alamat Jalan</label>
							<Textarea id="school-address" bind:value={profile.address} rows={3} />
						</div>
						<div>
							<label for="school-village" class="text-sm font-medium">Desa/Kelurahan</label>
							<Input id="school-village" bind:value={profile.village} />
						</div>
						<div>
							<label for="school-district" class="text-sm font-medium">Kecamatan</label>
							<Input id="school-district" bind:value={profile.district} />
						</div>
						<div>
							<label for="school-regency" class="text-sm font-medium">Kabupaten/Kota</label>
							<Input id="school-regency" bind:value={profile.regency} />
						</div>
						<div>
							<label for="school-province" class="text-sm font-medium">Provinsi</label>
							<Input id="school-province" bind:value={profile.province} />
						</div>
						<div>
							<label for="school-postal" class="text-sm font-medium">Kode Pos</label>
							<Input id="school-postal" bind:value={profile.postal_code} />
						</div>
						<div>
							<label for="school-phone" class="text-sm font-medium">Telepon</label>
							<Input id="school-phone" bind:value={profile.phone} />
						</div>
						<div>
							<label for="school-email" class="text-sm font-medium">Email</label>
							<Input id="school-email" type="email" bind:value={profile.email} />
						</div>
						<div>
							<label for="school-website" class="text-sm font-medium">Website</label>
							<Input id="school-website" bind:value={profile.website} />
						</div>
					</div>
				</Card.Content>
			</Card.Root>

			<div class="space-y-6">
				<Card.Root class="border-slate-200">
					<Card.Header class="pb-2">
						<Card.Title class="text-base">Kepala Madrasah</Card.Title>
						<Card.Description>Digunakan untuk blok tanda tangan surat dan paket cetak.</Card.Description>
					</Card.Header>
					<Card.Content class="space-y-3">
						<div>
							<label for="head-name" class="text-sm font-medium">Nama Kepala Madrasah</label>
							<Input id="head-name" bind:value={profile.head_name} />
						</div>
						<div>
							<label for="head-nip" class="text-sm font-medium">NIP Kepala Madrasah</label>
							<Input id="head-nip" bind:value={profile.head_nip} />
						</div>
					</Card.Content>
				</Card.Root>

				<Card.Root class="border-slate-200">
					<Card.Header class="pb-2">
						<Card.Title class="text-base">Pratinjau Kop</Card.Title>
						<Card.Description>Tampilan ringkas yang akan dibawa ke halaman cetak.</Card.Description>
					</Card.Header>
					<Card.Content>
						<div class="rounded-md border border-slate-200 bg-white p-5 text-center text-slate-950">
							<p class="text-xs font-semibold uppercase">{profile.ministry_line || '-'}</p>
							<p class="text-xs font-semibold uppercase">{profile.office_line || '-'}</p>
							<p class="mt-1 text-lg font-bold uppercase">{profile.name || '-'}</p>
							<p class="mt-1 text-xs leading-5">{schoolAddressLine(profile) || 'Alamat belum diisi'}</p>
							<p class="mt-1 text-xs text-slate-600">
								{#if profile.nsm}NSM {profile.nsm}{/if}
								{#if profile.nsm && profile.npsn} · {/if}
								{#if profile.npsn}NPSN {profile.npsn}{/if}
							</p>
						</div>
					</Card.Content>
				</Card.Root>
			</div>
		</div>
		{/snippet}
	</AsyncContent>
</div>
