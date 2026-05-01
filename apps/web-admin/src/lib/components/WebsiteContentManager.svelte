<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import * as Card from '$lib/components/ui/card';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import { toast } from '$lib/components/ui/sonner';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import EmptyStatePanel from '$lib/components/EmptyStatePanel.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import SuccessPanel from '$lib/components/SuccessPanel.svelte';
	import { confirmAction } from '$lib/confirm-dialog';

	type ContentStatus = 'draft' | 'published';
	type ContentPresentationStatus = 'draft' | 'published' | 'scheduled';
	type ContentKind = 'page' | 'post' | 'announcement';
	type WebsiteContent = {
		id: string;
		kind: ContentKind;
		title: string;
		slug: string;
		excerpt: string;
		content_html: string;
		cover_image_url: string;
		is_featured: boolean;
		meta_title: string;
		meta_description: string;
		status: ContentStatus;
		published_at: string | null;
		updated_at: string;
	};
	type ContentListPayload = {
		items?: WebsiteContent[];
		data?: WebsiteContent[];
		error?: string;
		message?: string;
	};
	type ApiEnvelope<T> = {
		data?: T;
		error?: string;
		message?: string;
	};

	let {
		kind,
		title,
		description,
		publicBasePath,
	}: {
		kind: ContentKind;
		title: string;
		description: string;
		publicBasePath: string;
	} = $props();

	let contentPromise = $state<Promise<WebsiteContent[]> | null>(null);
	let items = $state<WebsiteContent[]>([]);
	let success = $state('');
	let search = $state('');
	let showDialog = $state(false);
	let saving = $state(false);
	let uploadingCover = $state(false);
	let toggleBusyId = $state<string | null>(null);
	let deleteBusyId = $state<string | null>(null);
	let editingId = $state<string | null>(null);
	let form = $state({
		title: '',
		slug: '',
		excerpt: '',
		content_html: '',
		cover_image_url: '',
		is_featured: false,
		meta_title: '',
		meta_description: '',
		status: 'draft' as ContentStatus,
		published_at: '',
	});

	let filteredItems = $derived.by(() => {
		const query = search.trim().toLowerCase();
		if (!query) return items;
		return items.filter((item) =>
			item.title.toLowerCase().includes(query) ||
			item.slug.toLowerCase().includes(query) ||
			item.excerpt.toLowerCase().includes(query)
		);
	});

	function resetForm() {
		editingId = null;
		form = {
			title: '',
			slug: '',
			excerpt: '',
			content_html: '',
			cover_image_url: '',
			is_featured: false,
			meta_title: '',
			meta_description: '',
			status: 'draft',
			published_at: '',
		};
	}

	function openCreate() {
		resetForm();
		showDialog = true;
	}

	function openEdit(item: WebsiteContent) {
		editingId = item.id;
		form = {
			title: item.title,
			slug: item.slug,
			excerpt: item.excerpt,
			content_html: item.content_html,
			cover_image_url: item.cover_image_url,
			is_featured: item.is_featured ?? false,
			meta_title: item.meta_title ?? '',
			meta_description: item.meta_description ?? '',
			status: item.status,
			published_at: toInputDateTime(item.published_at),
		};
		showDialog = true;
	}

	function fmtDate(value: string | null) {
		if (!value) return 'Draft';
		return new Date(value).toLocaleString('id-ID', {
			timeZone: 'Asia/Makassar',
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
		});
	}

	function toInputDateTime(value: string | null) {
		if (!value) return '';
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return '';
		const year = date.getFullYear();
		const month = `${date.getMonth() + 1}`.padStart(2, '0');
		const day = `${date.getDate()}`.padStart(2, '0');
		const hours = `${date.getHours()}`.padStart(2, '0');
		const minutes = `${date.getMinutes()}`.padStart(2, '0');
		return `${year}-${month}-${day}T${hours}:${minutes}`;
	}

	function publishAtPayloadValue(value: string) {
		const trimmed = value.trim();
		if (!trimmed) return '';
		const date = new Date(trimmed);
		return Number.isNaN(date.getTime()) ? trimmed : date.toISOString();
	}

	function itemPresentationStatus(item: WebsiteContent): ContentPresentationStatus {
		if (item.status !== 'published') return 'draft';
		if (item.published_at && new Date(item.published_at).getTime() > Date.now()) return 'scheduled';
		return 'published';
	}

	function statusLabel(item: WebsiteContent) {
		const state = itemPresentationStatus(item);
		if (state === 'scheduled') return 'Terjadwal';
		return state === 'published' ? 'Terbit' : 'Draft';
	}

	function statusBadgeClass(item: WebsiteContent) {
		const state = itemPresentationStatus(item);
		if (state === 'published') return 'border-emerald-300 text-emerald-700';
		if (state === 'scheduled') return 'border-sky-300 text-sky-700';
		return '';
	}

	function publicPath(item: WebsiteContent) {
		if (kind === 'page') {
			if (item.slug === 'ppdb-info') return '/ppdb';
			return `/${item.slug}`;
		}
		return `${publicBasePath}/${item.slug}`;
	}

	const publishedCount = $derived(items.filter((item) => itemPresentationStatus(item) === 'published').length);
	const scheduledCount = $derived(items.filter((item) => itemPresentationStatus(item) === 'scheduled').length);
	const draftCount = $derived(items.filter((item) => item.status === 'draft').length);
	const featuredCount = $derived(items.filter((item) => item.is_featured).length);

	function isRecord(value: unknown): value is Record<string, unknown> {
		return typeof value === 'object' && value !== null;
	}

	function apiErrorMessage(payload: unknown) {
		if (!isRecord(payload)) return '';
		const error = payload.error;
		if (typeof error === 'string' && error.trim()) return error;
		const message = payload.message;
		if (typeof message === 'string' && message.trim()) return message;
		return '';
	}

	async function readApi<T>(response: Response, fallbackMessage: string): Promise<T> {
		const payload = (await response.json().catch(() => null)) as ApiEnvelope<T> | T | null;
		const message = apiErrorMessage(payload);
		if (!response.ok) throw new Error(message || fallbackMessage);
		if (isRecord(payload) && typeof payload.error === 'string' && payload.error.trim()) {
			throw new Error(payload.error);
		}
		if (isRecord(payload) && 'data' in payload) {
			const envelope = payload as ApiEnvelope<T>;
			if (envelope.data === undefined) throw new Error(fallbackMessage);
			return envelope.data;
		}
		if (payload === null) throw new Error(fallbackMessage);
		return payload as T;
	}

	function normalizeItems(value: ContentListPayload | WebsiteContent[] | null | undefined): WebsiteContent[] {
		if (Array.isArray(value)) return value;
		if (!value) return [];
		if (Array.isArray(value.items)) return value.items;
		if (Array.isArray(value.data)) return value.data;
		return [];
	}

	function contentErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		if (typeof error === 'string' && error.trim()) return error;
		return `Gagal memuat ${title.toLowerCase()}. Coba lagi untuk mengambil daftar konten terbaru.`;
	}

	function mutationErrorMessage(error: unknown, fallbackMessage: string) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return fallbackMessage;
	}

	async function fetchContentItems() {
		const payload = await fetch(`/api/website/content?kind=${kind}`).then((response) =>
			readApi<ContentListPayload | WebsiteContent[]>(response, `Gagal memuat ${title.toLowerCase()}`)
		);
		return normalizeItems(payload);
	}

	function loadContent() {
		contentPromise = fetchContentItems().then((nextItems) => {
			items = nextItems;
			return nextItems;
		});
		return contentPromise;
	}

	async function refreshContent() {
		if (!contentPromise) {
			await loadContent();
			return;
		}
		try {
			const nextItems = await fetchContentItems();
			items = nextItems;
			contentPromise = Promise.resolve(nextItems);
		} catch (error) {
			contentPromise = Promise.resolve(items);
			toast.error(contentErrorMessage(error));
		}
	}

	function retryContent(reset?: () => void) {
		reset?.();
		loadContent();
	}

	function handleContentRenderError(error: unknown, reset: () => void) {
		console.error('Website content manager render failed', error);
		reset();
	}

	async function uploadCoverImage(e: Event) {
		const input = e.target as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;
		uploadingCover = true;
		try {
			const fd = new FormData();
			fd.append('file', file);
			const res = await fetch('/api/website/media', { method: 'POST', body: fd });
			const data = await res.json().catch(() => null);
			if (!res.ok) throw new Error(apiErrorMessage(data) || 'Gagal upload gambar.');
			const url = isRecord(data) && typeof data.url === 'string' ? data.url : '';
			if (!url) throw new Error('Upload berhasil tetapi URL gambar tidak dikembalikan.');
			form.cover_image_url = url;
			toast.success('Gambar berhasil diunggah.');
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Gagal upload gambar.'));
		} finally {
			uploadingCover = false;
			input.value = '';
		}
	}

	async function save() {
		saving = true;
		try {
			success = '';
			const payload = { kind, ...form, published_at: publishAtPayloadValue(form.published_at) };
			const res = await fetch(editingId ? `/api/website/content/${editingId}` : '/api/website/content', {
				method: editingId ? 'PUT' : 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify(payload),
			});
			const data = await res.json().catch(() => null);
			if (!res.ok) throw new Error(apiErrorMessage(data) || 'Gagal menyimpan konten.');
			toast.success(editingId ? 'Konten berhasil diperbarui.' : 'Konten berhasil dibuat.');
			const scheduled = form.status === 'published' && !!form.published_at.trim();
			success = editingId
				? `Konten "${form.title}" berhasil diperbarui. ${scheduled ? 'Jadwal tayangnya juga sudah diperbarui.' : 'Periksa status publish-nya sebelum menutup sesi editorial ini.'}`
				: `Konten "${form.title}" berhasil dibuat sebagai ${form.status === 'published' ? (scheduled ? 'konten terjadwal' : 'terbit') : 'draft'}.`;
			showDialog = false;
			resetForm();
			await refreshContent();
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Gagal menyimpan konten.'));
		} finally {
			saving = false;
		}
	}

	async function remove(item: WebsiteContent) {
		if (!(await confirmAction({
			title: 'Hapus Konten Website',
			message: `Hapus konten "${item.title}"?`,
			confirmLabel: 'Hapus Konten',
			tone: 'danger'
		}))) return;
		deleteBusyId = item.id;
		try {
			const res = await fetch(`/api/website/content/${item.id}`, { method: 'DELETE' });
			const data = await res.json().catch(() => null);
			if (!res.ok) throw new Error(apiErrorMessage(data) || 'Gagal menghapus konten.');
			toast.success('Konten berhasil dihapus.');
			success = `Konten "${item.title}" berhasil dihapus dari area editorial.`;
			await refreshContent();
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Gagal menghapus konten.'));
		} finally {
			deleteBusyId = null;
		}
	}

	async function toggleStatus(item: WebsiteContent) {
		const nextStatus: ContentStatus = item.status === 'published' ? 'draft' : 'published';
		toggleBusyId = item.id;
		try {
			const res = await fetch(`/api/website/content/${item.id}`, {
				method: 'PUT',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({
					kind,
					title: item.title,
					slug: item.slug,
					excerpt: item.excerpt,
					content_html: item.content_html,
					cover_image_url: item.cover_image_url,
					is_featured: item.is_featured ?? false,
					meta_title: item.meta_title ?? '',
					meta_description: item.meta_description ?? '',
					status: nextStatus,
					published_at: nextStatus === 'published' ? '' : '',
				}),
			});
			const data = await res.json().catch(() => null);
			if (!res.ok) throw new Error(apiErrorMessage(data) || 'Gagal mengubah status konten.');
			toast.success(nextStatus === 'published' ? 'Konten diterbitkan.' : 'Konten dikembalikan ke draft.');
			success = nextStatus === 'published'
				? `Konten "${item.title}" berhasil dipublikasikan dan sekarang tersedia di website publik.`
				: `Konten "${item.title}" berhasil dikembalikan ke draft untuk revisi lebih lanjut.`;
			await refreshContent();
		} catch (error) {
			toast.error(mutationErrorMessage(error, 'Gagal mengubah status konten.'));
		} finally {
			toggleBusyId = null;
		}
	}

	onMount(() => {
		void loadContent();
	});
</script>

<div class="space-y-6">
	<div class="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
		<div>
			<h1 class="text-2xl font-semibold text-slate-800">{title}</h1>
			<p class="mt-1 text-sm text-muted-foreground">{description}</p>
		</div>
		<div class="flex flex-wrap items-center gap-2">
			<Input placeholder="Cari judul, slug, atau ringkasan..." bind:value={search} class="w-72" />
			<Button onclick={openCreate}>Tambah Konten</Button>
		</div>
	</div>

	{#if success}
		<SuccessPanel title="Aksi Editorial Berhasil" message={success} />
	{/if}

	<AsyncContent promise={contentPromise} onerror={handleContentRenderError}>
		{#snippet pending()}
			<div class="space-y-4">
				<div class="grid gap-3 md:grid-cols-5">
					{#each Array.from({ length: 5 }) as _, index (`website-stat-skeleton-${index}`)}
						<div class="rounded-2xl border border-slate-200 bg-white px-4 py-4">
							<Skeleton class="h-3 w-24" />
							<Skeleton class="mt-3 h-8 w-14" />
							<Skeleton class="mt-2 h-4 w-32" />
						</div>
					{/each}
				</div>
				<Card.Root class="border-slate-200 shadow-sm">
					<Card.Content class="grid gap-3 p-4 md:grid-cols-[1.2fr_auto]">
						<div class="space-y-2">
							<Skeleton class="h-3 w-28" />
							<Skeleton class="h-10 w-full" />
						</div>
						<div class="flex items-end">
							<Skeleton class="h-10 w-full md:w-36" />
						</div>
					</Card.Content>
				</Card.Root>
				<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
					<Card.Content class="space-y-4 px-5 py-6">
						{#each Array.from({ length: 4 }) as _, index (`website-content-skeleton-${index}`)}
							<div class="grid gap-4 lg:grid-cols-[1.3fr,0.7fr,0.8fr] lg:items-start">
								<div class="space-y-2">
									<Skeleton class="h-6 w-48" />
									<Skeleton class="h-4 w-40" />
									<Skeleton class="h-4 w-full max-w-xl" />
									<Skeleton class="h-4 w-full max-w-lg" />
								</div>
								<div class="space-y-2">
									<Skeleton class="h-4 w-28" />
									<Skeleton class="h-4 w-28" />
								</div>
								<div class="flex flex-wrap justify-start gap-2 lg:justify-end">
									<Skeleton class="h-9 w-20" />
									<Skeleton class="h-9 w-20" />
									<Skeleton class="h-9 w-28" />
								</div>
							</div>
						{/each}
					</Card.Content>
				</Card.Root>
			</div>
		{/snippet}
		{#snippet failed(error, reset)}
			<RecoveryPanel title="Konten Website Belum Tersaji" message={contentErrorMessage(error)} onRetry={() => retryContent(reset)} />
		{/snippet}
		{#snippet children(_items)}

	<div class="grid gap-3 md:grid-cols-5">
		<div class="rounded-2xl border border-emerald-100 bg-emerald-50 px-4 py-4">
			<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-emerald-700">Total Konten</p>
			<p class="mt-2 text-2xl font-semibold text-slate-900">{items.length}</p>
			<p class="text-sm text-slate-600">seluruh item pada kategori ini</p>
		</div>
		<div class="rounded-2xl border border-sky-100 bg-sky-50 px-4 py-4">
			<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-sky-700">Terbit</p>
			<p class="mt-2 text-2xl font-semibold text-slate-900">{publishedCount}</p>
			<p class="text-sm text-slate-600">konten yang sudah tampil di website publik</p>
		</div>
		<div class="rounded-2xl border border-cyan-100 bg-cyan-50 px-4 py-4">
			<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-cyan-700">Terjadwal</p>
			<p class="mt-2 text-2xl font-semibold text-slate-900">{scheduledCount}</p>
			<p class="text-sm text-slate-600">konten yang akan tayang otomatis sesuai jadwal</p>
		</div>
		<div class="rounded-2xl border border-amber-100 bg-amber-50 px-4 py-4">
			<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-amber-700">Draft</p>
			<p class="mt-2 text-2xl font-semibold text-slate-900">{draftCount}</p>
			<p class="text-sm text-slate-600">konten yang masih menunggu finalisasi</p>
		</div>
		<div class="rounded-2xl border border-sky-100 bg-sky-50 px-4 py-4">
			<p class="text-[11px] font-semibold uppercase tracking-[0.22em] text-sky-700">Unggulan</p>
			<p class="mt-2 text-2xl font-semibold text-slate-900">{featuredCount}</p>
			<p class="text-sm text-slate-600">konten yang ditandai untuk sorotan publik</p>
		</div>
	</div>

	<Card.Root class="border-slate-200 shadow-sm">
		<Card.Content class="grid gap-3 p-4 md:grid-cols-[1.2fr_auto]">
			<div>
				<p class="mb-1 text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">Cari Konten</p>
				<Input placeholder="Cari judul, slug, atau ringkasan..." bind:value={search} class="w-full" />
			</div>
			<div class="flex items-end">
				<Button variant="outline" class="w-full md:w-auto" onclick={() => (search = '')} disabled={!search.trim()}>
					Reset pencarian
				</Button>
			</div>
		</Card.Content>
	</Card.Root>

	<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
		<Card.Content class="p-0">
			{#if filteredItems.length === 0}
				<div class="p-4">
					<EmptyStatePanel
						title={search.trim() ? 'Tidak ada konten yang cocok' : 'Belum ada konten untuk kategori ini'}
						description={search.trim()
							? 'Ubah kata kunci pencarian atau reset filter untuk melihat item lain yang sudah tersedia.'
							: 'Tambahkan konten pertama agar area editorial ini mulai menampilkan berita, pengumuman, atau halaman publik sekolah.'}
						compact
					/>
				</div>
			{:else}
				<div class="divide-y divide-slate-200">
					{#each filteredItems as item (item.id)}
						<div class="grid gap-4 px-5 py-5 lg:grid-cols-[1.3fr,0.7fr,0.8fr] lg:items-start">
							<div class="space-y-2">
								<div class="flex flex-wrap items-center gap-2">
									<h2 class="text-lg font-semibold text-slate-900">{item.title}</h2>
									<Badge variant={item.status === 'draft' ? 'secondary' : 'outline'} class={statusBadgeClass(item)}>
										{statusLabel(item)}
									</Badge>
									{#if item.is_featured}
										<Badge variant="outline" class="border-amber-300 text-amber-700">Unggulan</Badge>
									{/if}
								</div>
								<p class="font-mono text-xs text-slate-500">{publicPath(item)}</p>
								<p class="text-sm leading-7 text-slate-600">{item.excerpt || 'Belum ada ringkasan.'}</p>
							</div>
							<div class="space-y-2 text-sm text-slate-600">
								<p><span class="font-medium text-slate-900">{itemPresentationStatus(item) === 'scheduled' ? 'Jadwal tayang:' : 'Tayang:'}</span> {fmtDate(item.published_at)}</p>
								<p><span class="font-medium text-slate-900">Update:</span> {fmtDate(item.updated_at)}</p>
							</div>
							<div class="flex flex-wrap justify-start gap-2 lg:justify-end">
									{#if item.status === 'published'}
										{#if item.kind === 'post'}
											<a href={resolve('/berita/[slug]', { slug: item.slug })} target="_blank" rel="noreferrer">
												<Button variant="outline">Lihat</Button>
											</a>
										{:else if item.kind === 'announcement'}
											<a href={resolve('/pengumuman/[slug]', { slug: item.slug })} target="_blank" rel="noreferrer">
												<Button variant="outline">Lihat</Button>
											</a>
										{:else if item.slug === 'profil'}
											<a href={resolve('/profil')} target="_blank" rel="noreferrer">
												<Button variant="outline">Lihat</Button>
											</a>
										{:else if item.slug === 'kontak'}
											<a href={resolve('/kontak')} target="_blank" rel="noreferrer">
												<Button variant="outline">Lihat</Button>
											</a>
										{:else if item.slug === 'ppdb-info'}
											<a href={resolve('/ppdb')} target="_blank" rel="noreferrer">
												<Button variant="outline">Lihat</Button>
											</a>
										{/if}
									{/if}
								<Button variant="outline" onclick={() => openEdit(item)}>Edit</Button>
								<LoadingButton
									variant="outline"
									onclick={() => toggleStatus(item)}
									loading={toggleBusyId === item.id}
									loadingLabel="Memproses..."
									disabled={(toggleBusyId !== null && toggleBusyId !== item.id) || deleteBusyId !== null}
								>
									{item.status === 'published' ? 'Kembalikan ke Draft' : 'Terbitkan'}
								</LoadingButton>
								<LoadingButton
									variant="ghost"
									class="text-destructive hover:text-destructive"
									onclick={() => remove(item)}
									loading={deleteBusyId === item.id}
									loadingLabel="Menghapus..."
									disabled={(deleteBusyId !== null && deleteBusyId !== item.id) || toggleBusyId !== null}
								>
									Hapus
								</LoadingButton>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</Card.Content>
	</Card.Root>
		{/snippet}
	</AsyncContent>
</div>

<Dialog.Root bind:open={showDialog}>
	<Dialog.Content>
		<div class="space-y-4">
			<div>
				<h2 class="text-base font-semibold text-slate-900">{editingId ? 'Edit Konten' : 'Tambah Konten'}</h2>
				<p class="mt-1 text-sm text-slate-500">Kelola konten {title.toLowerCase()} dengan alur draft hingga terbit.</p>
			</div>

			<div class="grid gap-3 sm:grid-cols-2">
				<div class="sm:col-span-2">
					<label for="website-title" class="mb-1 block text-xs font-medium text-slate-600">Judul</label>
					<Input id="website-title" bind:value={form.title} />
				</div>
				<div>
					<label for="website-slug" class="mb-1 block text-xs font-medium text-slate-600">Slug</label>
					<Input id="website-slug" bind:value={form.slug} placeholder="Biarkan kosong jika ingin dibuat otomatis" />
				</div>
				<div>
					<label for="website-status" class="mb-1 block text-xs font-medium text-slate-600">Status</label>
					<select id="website-status" bind:value={form.status} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
						<option value="draft">Draft</option>
						<option value="published">Terbit</option>
					</select>
				</div>
				<div class="sm:col-span-2">
					<label for="website-published-at" class="mb-1 block text-xs font-medium text-slate-600">Jadwal Tayang</label>
					<Input id="website-published-at" type="datetime-local" bind:value={form.published_at} />
					<p class="mt-1 text-xs text-slate-500">Kosongkan untuk tayang segera saat status diubah ke terbit. Isi jadwal jika konten ingin tayang otomatis di waktu tertentu.</p>
				</div>

				<!-- Cover Image -->
				<div class="sm:col-span-2">
					<label for="website-cover" class="mb-1 block text-xs font-medium text-slate-600">Cover Image</label>
					<div class="flex gap-2">
						<Input id="website-cover" bind:value={form.cover_image_url} placeholder="Tempel URL gambar atau unggah file di samping" class="flex-1" />
						<label class="flex cursor-pointer items-center gap-1.5 rounded-md border border-input bg-background px-3 py-2 text-sm text-slate-700 transition-colors hover:bg-slate-50 {uploadingCover ? 'opacity-60 pointer-events-none' : ''}">
							{#if uploadingCover}
								<span class="h-4 w-4 animate-spin rounded-full border-2 border-slate-400 border-t-transparent"></span>
								Mengunggah...
							{:else}
								<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
								Upload
							{/if}
							<input type="file" accept="image/*" class="sr-only" onchange={uploadCoverImage} />
						</label>
					</div>
					{#if form.cover_image_url}
						<div class="mt-2">
							<img src={form.cover_image_url} alt="Cover preview" class="h-24 rounded-md border border-slate-200 object-cover" onerror={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }} />
						</div>
					{/if}
				</div>

				<!-- Featured Flag -->
				<div class="sm:col-span-2 flex items-center gap-2">
					<input
						type="checkbox"
						id="website-featured"
						bind:checked={form.is_featured}
						class="h-4 w-4 rounded border-input accent-emerald-700"
					/>
					<label for="website-featured" class="text-sm text-slate-700 cursor-pointer">
						Tandai sebagai konten unggulan (ditampilkan di bagian utama homepage)
					</label>
				</div>

				<div class="sm:col-span-2">
					<label for="website-excerpt" class="mb-1 block text-xs font-medium text-slate-600">Ringkasan</label>
					<p class="mb-1 text-xs text-slate-500">Ringkasan singkat ini dipakai pada daftar konten dan pratinjau publik.</p>
					<Textarea id="website-excerpt" rows={3} bind:value={form.excerpt} />
				</div>
				<div class="sm:col-span-2">
					<label for="website-content" class="mb-1 block text-xs font-medium text-slate-600">Konten HTML</label>
					<p class="mb-1 text-xs text-slate-500">Gunakan HTML yang rapi agar tampilan halaman publik tetap nyaman dibaca.</p>
					<Textarea id="website-content" rows={12} bind:value={form.content_html} />
				</div>

				<!-- SEO Section -->
				<div class="sm:col-span-2">
					<p class="mb-2 text-xs font-semibold uppercase tracking-wide text-slate-500">Metadata SEO</p>
					<div class="grid gap-3">
						<div>
							<label for="website-meta-title" class="mb-1 block text-xs font-medium text-slate-600">
								Judul SEO <span class="text-slate-400">(kosongkan = pakai judul utama)</span>
							</label>
							<Input id="website-meta-title" bind:value={form.meta_title} placeholder="Judul untuk mesin pencari..." />
						</div>
						<div>
							<label for="website-meta-desc" class="mb-1 block text-xs font-medium text-slate-600">
								Deskripsi SEO <span class="text-slate-400">(150–160 karakter ideal)</span>
							</label>
							<Textarea id="website-meta-desc" rows={2} bind:value={form.meta_description} placeholder="Deskripsi singkat untuk mesin pencari dan media sosial..." />
							<p class="mt-1 text-right text-xs text-slate-400">{form.meta_description.length} karakter</p>
						</div>
					</div>
				</div>
			</div>

			<div class="flex justify-end gap-2">
				<Button variant="outline" onclick={() => (showDialog = false)}>Batal</Button>
				<LoadingButton onclick={() => void save()} loading={saving} loadingLabel="Menyimpan..." label="Simpan" />
			</div>
		</div>
	</Dialog.Content>
</Dialog.Root>
