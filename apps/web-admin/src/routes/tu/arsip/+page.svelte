<script lang="ts">
	import ArchiveIcon from '@lucide/svelte/icons/archive';
	import DownloadIcon from '@lucide/svelte/icons/download';
	import PencilIcon from '@lucide/svelte/icons/pencil';
	import RefreshCcwIcon from '@lucide/svelte/icons/refresh-ccw';
	import SaveIcon from '@lucide/svelte/icons/save';
	import SearchIcon from '@lucide/svelte/icons/search';
	import Trash2Icon from '@lucide/svelte/icons/trash-2';
	import UploadIcon from '@lucide/svelte/icons/upload';
	import { onMount } from 'svelte';
	import AsyncContent from '$lib/components/AsyncContent.svelte';
	import LoadingButton from '$lib/components/LoadingButton.svelte';
	import RecoveryPanel from '$lib/components/RecoveryPanel.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import * as Table from '$lib/components/ui/table';
	import { Textarea } from '$lib/components/ui/textarea';
	import { toast } from '$lib/components/ui/sonner';
	import { confirmAction } from '$lib/confirm-dialog';
	import { readClientApiData, readClientJson } from '$lib/client/api';

	type ArchiveStats = {
		active_categories: number;
		total_documents: number;
		active_documents: number;
		borrowed_documents: number;
		disposed_documents: number;
		documents_this_year: number;
		total_file_size: number;
	};

	type ArchiveCategory = {
		id: string;
		code: string;
		name: string;
		classification_code: string | null;
		classification_name: string;
		description: string;
		retention_years: number;
		is_active: boolean;
		document_count: number;
	};

	type ArchiveDocument = {
		id: string;
		category_id: string;
		category_code: string;
		category_name: string;
		classification_code: string;
		title: string;
		archive_number: string;
		document_date: string | null;
		received_date: string;
		summary: string;
		tags: string;
		status: string;
		storage_location: string;
		retention_until: string | null;
		original_name: string;
		mime_type: string;
		file_size: number;
		uploaded_by_username: string;
		created_at: string;
	};

	type ArchiveOverview = {
		stats: ArchiveStats;
		categories: ArchiveCategory[];
		documents: ArchiveDocument[];
	};

	type CategoryForm = {
		code: string;
		name: string;
		classification_code: string;
		description: string;
		retention_years: number;
		is_active: boolean;
	};

	type DocumentForm = {
		category_id: string;
		title: string;
		archive_number: string;
		document_date: string;
		received_date: string;
		summary: string;
		tags: string;
		status: string;
		storage_location: string;
		retention_until: string;
	};

	const CLASSIFICATION_OPTIONS = [
		['', 'Tanpa klasifikasi'],
		['OT.00', 'OT.00 Organisasi & Tata Usaha'],
		['PP.00.2', 'PP.00.2 Peserta Didik'],
		['KP.00', 'KP.00 Kepegawaian'],
		['KU.00', 'KU.00 Keuangan'],
		['KS.00', 'KS.00 Sarana Prasarana'],
		['HM.00', 'HM.00 Hubungan Masyarakat'],
		['HK.00', 'HK.00 Hukum']
	] as const;

	const STATUS_OPTIONS = [
		['', 'Semua status'],
		['active', 'Aktif'],
		['borrowed', 'Dipinjam'],
		['disposed', 'Dimusnahkan']
	] as const;

	const today = new Date().toISOString().slice(0, 10);
	const emptyStats: ArchiveStats = {
		active_categories: 0,
		total_documents: 0,
		active_documents: 0,
		borrowed_documents: 0,
		disposed_documents: 0,
		documents_this_year: 0,
		total_file_size: 0
	};

	let archivePromise = $state<Promise<ArchiveOverview> | null>(null);
	let archiveRequestId = 0;
	let categories = $state<ArchiveCategory[]>([]);
	let documents = $state<ArchiveDocument[]>([]);
	let search = $state('');
	let status = $state('');
	let categoryFilter = $state('');
	let classificationFilter = $state('');
	let refreshBusy = $state(false);
	let filterBusy = $state(false);
	let categoryBusy = $state(false);
	let documentBusy = $state(false);
	let deleteBusy = $state<Record<string, boolean>>({});
	let editingCategoryId = $state('');
	let editingDocumentId = $state('');
	let selectedFile = $state<File | null>(null);
	let fileInput = $state<HTMLInputElement | null>(null);
	let categoryForm = $state<CategoryForm>({
		code: '',
		name: '',
		classification_code: '',
		description: '',
		retention_years: 5,
		is_active: true
	});
	let documentForm = $state<DocumentForm>({
		category_id: '',
		title: '',
		archive_number: '',
		document_date: '',
		received_date: today,
		summary: '',
		tags: '',
		status: 'active',
		storage_location: '',
		retention_until: ''
	});

	let activeCategories = $derived(categories.filter((category) => category.is_active || category.id === documentForm.category_id));

	async function fetchArchiveOverview(): Promise<ArchiveOverview> {
		const params = new URLSearchParams();
		if (search.trim()) params.set('search', search.trim());
		if (categoryFilter) params.set('category_id', categoryFilter);
		if (status) params.set('status', status);
		if (classificationFilter) params.set('classification_code', classificationFilter);
		const suffix = params.toString() ? `?${params.toString()}` : '';
		const [statsRes, categoriesRes, documentsRes] = await Promise.all([
			fetch('/api/tu/archives/stats'),
			fetch('/api/tu/archives/categories'),
			fetch(`/api/tu/archives/documents${suffix}`)
		]);
		const [stats, categoryRows, documentRows] = await Promise.all([
			readClientApiData<ArchiveStats>(statsRes, 'Gagal memuat statistik arsip.'),
			readClientApiData<ArchiveCategory[]>(categoriesRes, 'Gagal memuat kategori arsip.'),
			readClientApiData<ArchiveDocument[]>(documentsRes, 'Gagal memuat register arsip.')
		]);
		return { stats, categories: categoryRows ?? [], documents: documentRows ?? [] };
	}

	function applyOverview(overview: ArchiveOverview) {
		categories = overview.categories ?? [];
		documents = overview.documents ?? [];
		const firstActiveCategory = categories.find((category) => category.is_active);
		if (!documentForm.category_id && firstActiveCategory) {
			documentForm.category_id = firstActiveCategory.id;
		}
	}

	function load() {
		const requestId = ++archiveRequestId;
		archivePromise = fetchArchiveOverview()
			.then((overview) => {
				if (requestId === archiveRequestId) {
					applyOverview(overview);
					return overview;
				}
				return { stats: emptyStats, categories, documents };
			})
			.catch((error: unknown) => {
				if (requestId === archiveRequestId) throw error;
				return { stats: emptyStats, categories, documents };
			});
		return archivePromise;
	}

	async function refreshArchive() {
		refreshBusy = true;
		const promise = load();
		const requestId = archiveRequestId;
		try {
			await promise;
		} catch (error) {
			toast.error(archiveErrorMessage(error));
		} finally {
			if (requestId === archiveRequestId) refreshBusy = false;
		}
	}

	async function applyArchiveFilters() {
		filterBusy = true;
		const promise = load();
		const requestId = archiveRequestId;
		try {
			await promise;
		} catch (error) {
			toast.error(archiveErrorMessage(error));
		} finally {
			if (requestId === archiveRequestId) filterBusy = false;
		}
	}

	function retryArchive(reset?: () => void) {
		reset?.();
		load();
	}

	function archiveErrorMessage(error: unknown) {
		if (error instanceof Error && error.message.trim()) return error.message;
		return 'Gagal memuat data arsip. Coba lagi untuk mengambil register terbaru.';
	}

	function resetCategoryForm() {
		editingCategoryId = '';
		categoryForm = {
			code: '',
			name: '',
			classification_code: '',
			description: '',
			retention_years: 5,
			is_active: true
		};
	}

	function editCategory(category: ArchiveCategory) {
		editingCategoryId = category.id;
		categoryForm = {
			code: category.code,
			name: category.name,
			classification_code: category.classification_code ?? '',
			description: category.description,
			retention_years: category.retention_years,
			is_active: category.is_active
		};
	}

	function resetDocumentForm() {
		editingDocumentId = '';
		selectedFile = null;
		if (fileInput) fileInput.value = '';
		const firstActiveCategory = categories.find((category) => category.is_active);
		documentForm = {
			category_id: firstActiveCategory?.id ?? '',
			title: '',
			archive_number: '',
			document_date: '',
			received_date: today,
			summary: '',
			tags: '',
			status: 'active',
			storage_location: '',
			retention_until: ''
		};
	}

	function editDocument(document: ArchiveDocument) {
		editingDocumentId = document.id;
		selectedFile = null;
		if (fileInput) fileInput.value = '';
		documentForm = {
			category_id: document.category_id,
			title: document.title,
			archive_number: document.archive_number,
			document_date: toInputDate(document.document_date),
			received_date: toInputDate(document.received_date) || today,
			summary: document.summary,
			tags: document.tags,
			status: document.status,
			storage_location: document.storage_location,
			retention_until: toInputDate(document.retention_until)
		};
	}

	function handleFileChange(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		selectedFile = input.files?.[0] ?? null;
	}

	function categoryPayload() {
		return {
			code: categoryForm.code.trim(),
			name: categoryForm.name.trim(),
			classification_code: categoryForm.classification_code,
			description: categoryForm.description.trim(),
			retention_years: Number(categoryForm.retention_years) || 0,
			is_active: categoryForm.is_active
		};
	}

	async function saveCategory() {
		if (!categoryForm.code.trim()) {
			toast.error('Kode kategori wajib diisi');
			return;
		}
		if (!categoryForm.name.trim()) {
			toast.error('Nama kategori wajib diisi');
			return;
		}
		categoryBusy = true;
		try {
			const response = await fetch(
				editingCategoryId ? `/api/tu/archives/categories/${editingCategoryId}` : '/api/tu/archives/categories',
				{
					method: editingCategoryId ? 'PUT' : 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify(categoryPayload())
				}
			);
			await readClientApiData<ArchiveCategory>(response, 'Gagal menyimpan kategori arsip.');
			toast.success(editingCategoryId ? 'Kategori arsip diperbarui' : 'Kategori arsip ditambahkan');
			resetCategoryForm();
			load();
		} catch (error) {
			toast.error(error instanceof Error ? error.message : 'Gagal menyimpan kategori arsip');
		} finally {
			categoryBusy = false;
		}
	}

	function appendDocumentForm(formData: FormData) {
		formData.set('category_id', documentForm.category_id);
		formData.set('title', documentForm.title.trim());
		formData.set('archive_number', documentForm.archive_number.trim());
		formData.set('document_date', documentForm.document_date);
		formData.set('received_date', documentForm.received_date);
		formData.set('summary', documentForm.summary.trim());
		formData.set('tags', documentForm.tags.trim());
		formData.set('status', documentForm.status);
		formData.set('storage_location', documentForm.storage_location.trim());
		formData.set('retention_until', documentForm.retention_until);
	}

	function documentPayload() {
		return {
			category_id: documentForm.category_id,
			title: documentForm.title.trim(),
			archive_number: documentForm.archive_number.trim(),
			document_date: documentForm.document_date,
			received_date: documentForm.received_date,
			summary: documentForm.summary.trim(),
			tags: documentForm.tags.trim(),
			status: documentForm.status,
			storage_location: documentForm.storage_location.trim(),
			retention_until: documentForm.retention_until
		};
	}

	async function saveDocument() {
		if (!documentForm.category_id) {
			toast.error('Kategori arsip wajib dipilih');
			return;
		}
		if (!documentForm.title.trim()) {
			toast.error('Judul arsip wajib diisi');
			return;
		}
		if (!editingDocumentId && !selectedFile) {
			toast.error('Pilih file arsip yang akan diunggah');
			return;
		}
		documentBusy = true;
		try {
			let response: Response;
			if (editingDocumentId) {
				response = await fetch(`/api/tu/archives/documents/${editingDocumentId}`, {
					method: 'PUT',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify(documentPayload())
				});
			} else {
				const formData = new FormData();
				appendDocumentForm(formData);
				if (selectedFile) formData.set('file', selectedFile);
				response = await fetch('/api/tu/archives/documents', {
					method: 'POST',
					body: formData
				});
			}
			await readClientApiData<ArchiveDocument>(response, 'Gagal menyimpan arsip.');
			toast.success(editingDocumentId ? 'Metadata arsip diperbarui' : 'Dokumen arsip diunggah');
			resetDocumentForm();
			load();
		} catch (error) {
			toast.error(error instanceof Error ? error.message : 'Gagal menyimpan arsip');
		} finally {
			documentBusy = false;
		}
	}

	async function deleteCategory(id: string) {
		if (!(await confirmAction({
			title: 'Hapus Kategori Arsip',
			message: 'Hapus kategori arsip ini? Kategori yang masih memiliki dokumen tidak dapat dihapus.',
			confirmLabel: 'Hapus Kategori',
			tone: 'danger'
		}))) return;
		deleteBusy[id] = true;
		try {
			const response = await fetch(`/api/tu/archives/categories/${id}`, { method: 'DELETE' });
			await readClientJson<unknown>(response);
			toast.success('Kategori arsip dihapus');
			load();
		} catch (error) {
			toast.error(error instanceof Error ? error.message : 'Gagal menghapus kategori arsip');
		} finally {
			deleteBusy[id] = false;
		}
	}

	async function deleteDocument(id: string) {
		if (!(await confirmAction({
			title: 'Hapus Arsip',
			message: 'Hapus arsip dan file tersimpan ini?',
			confirmLabel: 'Hapus Arsip',
			tone: 'danger'
		}))) return;
		deleteBusy[id] = true;
		try {
			const response = await fetch(`/api/tu/archives/documents/${id}`, { method: 'DELETE' });
			await readClientJson<unknown>(response);
			toast.success('Arsip dihapus');
			load();
		} catch (error) {
			toast.error(error instanceof Error ? error.message : 'Gagal menghapus arsip');
		} finally {
			deleteBusy[id] = false;
		}
	}

	function statusLabel(value: string) {
		if (value === 'borrowed') return 'Dipinjam';
		if (value === 'disposed') return 'Dimusnahkan';
		return 'Aktif';
	}

	function statusVariant(value: string): 'default' | 'secondary' | 'destructive' | 'outline' {
		if (value === 'borrowed') return 'secondary';
		if (value === 'disposed') return 'destructive';
		return 'default';
	}

	function formatBytes(value: number) {
		if (!Number.isFinite(value) || value <= 0) return '0 B';
		if (value < 1024) return `${value} B`;
		if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`;
		return `${(value / (1024 * 1024)).toFixed(1)} MB`;
	}

	function toInputDate(value: string | null | undefined) {
		if (!value) return '';
		return value.slice(0, 10);
	}

	function displayDate(value: string | null | undefined) {
		return toInputDate(value) || '-';
	}

	function handleArchiveRenderError(error: unknown) {
		console.error('Archive render failed', error);
	}

	onMount(() => {
		void load();
	});
</script>

<svelte:head><title>Arsip TU - MTsN 2 Kolaka Utara</title></svelte:head>

<div class="space-y-6">
	<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-lg font-semibold text-slate-800">Arsip Tata Usaha</h1>
			<p class="text-sm text-slate-500">Register digital arsip madrasah, kategori klasifikasi, lokasi simpan, dan masa retensi.</p>
		</div>
		<LoadingButton variant="outline" size="sm" loading={refreshBusy} loadingLabel="Memuat..." onclick={() => void refreshArchive()}>
			<RefreshCcwIcon class="mr-2 size-4" />
			Refresh
		</LoadingButton>
	</div>

	<AsyncContent promise={archivePromise} onerror={handleArchiveRenderError}>
		{#snippet pending()}
			<div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
				{#each Array.from({ length: 4 }) as _, index (`archive-stat-skeleton-${index}`)}
					<Card.Root class="border-slate-200">
						<Card.Content class="p-4">
							<Skeleton class="h-4 w-28" />
							<Skeleton class="mt-3 h-8 w-16" />
						</Card.Content>
					</Card.Root>
				{/each}
			</div>
			<div class="grid gap-6 lg:grid-cols-[1fr_360px]">
				<Card.Root class="border-slate-200">
					<Card.Header>
						<Skeleton class="h-5 w-40" />
					</Card.Header>
					<Card.Content class="space-y-3">
						{#each Array.from({ length: 6 }) as _, index (`archive-row-skeleton-${index}`)}
							<Skeleton class="h-9 w-full" />
						{/each}
					</Card.Content>
				</Card.Root>
				<Card.Root class="border-slate-200">
					<Card.Header>
						<Skeleton class="h-5 w-40" />
					</Card.Header>
					<Card.Content class="space-y-3">
						{#each Array.from({ length: 8 }) as _, index (`archive-form-skeleton-${index}`)}
							<Skeleton class="h-9 w-full" />
						{/each}
					</Card.Content>
				</Card.Root>
			</div>
		{/snippet}

		{#snippet failed(error, reset)}
			<RecoveryPanel message={archiveErrorMessage(error)} onRetry={() => retryArchive(reset)} />
		{/snippet}

		{#snippet children(value)}
			{@const overview = value as ArchiveOverview}
			<div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
				<Card.Root class="border-slate-200">
					<Card.Content class="p-4">
						<p class="text-xs text-slate-500">Total Dokumen</p>
						<p class="mt-2 text-2xl font-semibold text-slate-900">{overview.stats.total_documents}</p>
						<p class="mt-1 text-xs text-slate-500">{overview.stats.documents_this_year} dokumen tahun ini</p>
					</Card.Content>
				</Card.Root>
				<Card.Root class="border-slate-200">
					<Card.Content class="p-4">
						<p class="text-xs text-slate-500">Kategori Aktif</p>
						<p class="mt-2 text-2xl font-semibold text-slate-900">{overview.stats.active_categories}</p>
						<p class="mt-1 text-xs text-slate-500">Berbasis klasifikasi TU</p>
					</Card.Content>
				</Card.Root>
				<Card.Root class="border-slate-200">
					<Card.Content class="p-4">
						<p class="text-xs text-slate-500">Status Arsip</p>
						<p class="mt-2 text-2xl font-semibold text-slate-900">{overview.stats.active_documents}</p>
						<p class="mt-1 text-xs text-slate-500">{overview.stats.borrowed_documents} dipinjam, {overview.stats.disposed_documents} dimusnahkan</p>
					</Card.Content>
				</Card.Root>
				<Card.Root class="border-slate-200">
					<Card.Content class="p-4">
						<p class="text-xs text-slate-500">Ukuran File</p>
						<p class="mt-2 text-2xl font-semibold text-slate-900">{formatBytes(overview.stats.total_file_size)}</p>
						<p class="mt-1 text-xs text-slate-500">Tersimpan di backend</p>
					</Card.Content>
				</Card.Root>
			</div>

			<div class="grid gap-6 xl:grid-cols-[1fr_420px]">
				<div class="space-y-6">
					<Card.Root class="border-slate-200">
						<Card.Header class="space-y-4">
							<div class="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
								<div>
									<Card.Title class="text-base">Register Arsip</Card.Title>
									<Card.Description>Dokumen yang sudah diunggah beserta metadata lokasi dan retensinya.</Card.Description>
								</div>
								<div class="grid gap-2 sm:grid-cols-2 lg:grid-cols-[220px_150px_170px_150px_auto]">
									<div>
										<label for="archive-search" class="text-xs font-medium text-slate-600">Cari arsip</label>
										<div class="relative mt-1">
											<SearchIcon class="absolute left-2.5 top-2.5 size-4 text-slate-400" />
											<Input id="archive-search" class="pl-8" placeholder="Judul, nomor, tag" bind:value={search} />
										</div>
									</div>
									<div>
										<label for="archive-status-filter" class="text-xs font-medium text-slate-600">Status</label>
										<select
											id="archive-status-filter"
											class="mt-1 h-10 w-full rounded-md border border-slate-200 bg-white px-3 text-sm text-slate-800"
											bind:value={status}
										>
											{#each STATUS_OPTIONS as option (option[0])}
												<option value={option[0]}>{option[1]}</option>
											{/each}
										</select>
									</div>
									<div>
										<label for="archive-category-filter" class="text-xs font-medium text-slate-600">Kategori</label>
										<select
											id="archive-category-filter"
											class="mt-1 h-10 w-full rounded-md border border-slate-200 bg-white px-3 text-sm text-slate-800"
											bind:value={categoryFilter}
										>
											<option value="">Semua kategori</option>
											{#each overview.categories as category (category.id)}
												<option value={category.id}>{category.code}</option>
											{/each}
										</select>
									</div>
									<div>
										<label for="archive-classification-filter" class="text-xs font-medium text-slate-600">Klasifikasi</label>
										<select
											id="archive-classification-filter"
											class="mt-1 h-10 w-full rounded-md border border-slate-200 bg-white px-3 text-sm text-slate-800"
											bind:value={classificationFilter}
										>
											<option value="">Semua kode</option>
											{#each CLASSIFICATION_OPTIONS.filter((option) => option[0]) as option (option[0])}
												<option value={option[0]}>{option[0]}</option>
											{/each}
										</select>
									</div>
									<LoadingButton class="self-end" size="sm" loading={filterBusy} loadingLabel="Menerapkan..." onclick={() => void applyArchiveFilters()}>
										<SearchIcon class="mr-2 size-4" />
										Terapkan
									</LoadingButton>
								</div>
							</div>
						</Card.Header>
						<Card.Content class="p-0">
							{#if overview.documents.length === 0}
								<div class="border-t border-slate-100 px-6 py-10 text-center">
									<ArchiveIcon class="mx-auto size-8 text-emerald-700" />
									<p class="mt-3 text-sm font-semibold text-slate-800">Belum ada arsip sesuai filter.</p>
									<p class="mt-1 text-sm text-slate-500">Unggah dokumen pertama atau ubah filter register.</p>
								</div>
							{:else}
								<div class="overflow-x-auto">
									<Table.Root>
										<Table.Header>
											<Table.Row>
												<Table.Head>Dokumen</Table.Head>
												<Table.Head>Kategori</Table.Head>
												<Table.Head>Tanggal</Table.Head>
												<Table.Head>Status</Table.Head>
												<Table.Head>File</Table.Head>
												<Table.Head class="text-right">Aksi</Table.Head>
											</Table.Row>
										</Table.Header>
										<Table.Body>
											{#each overview.documents as document (document.id)}
												<Table.Row>
													<Table.Cell class="min-w-[260px]">
														<p class="font-medium text-slate-900">{document.title}</p>
														<p class="mt-1 text-xs text-slate-500">{document.archive_number || 'Tanpa nomor arsip'}</p>
														{#if document.summary}
															<p class="mt-1 max-w-xl text-xs text-slate-500">{document.summary}</p>
														{/if}
													</Table.Cell>
													<Table.Cell>
														<p class="text-sm text-slate-800">{document.category_code}</p>
														<p class="text-xs text-slate-500">{document.classification_code || '-'}</p>
													</Table.Cell>
													<Table.Cell>
														<p class="text-sm text-slate-800">{displayDate(document.received_date)}</p>
														<p class="text-xs text-slate-500">Retensi {displayDate(document.retention_until)}</p>
													</Table.Cell>
													<Table.Cell>
														<Badge variant={statusVariant(document.status)}>{statusLabel(document.status)}</Badge>
														<p class="mt-1 text-xs text-slate-500">{document.storage_location || 'Lokasi belum diisi'}</p>
													</Table.Cell>
													<Table.Cell>
														<p class="text-sm text-slate-800">{formatBytes(document.file_size)}</p>
														<p class="text-xs text-slate-500">{document.original_name}</p>
													</Table.Cell>
													<Table.Cell class="text-right">
														<div class="flex justify-end gap-2">
															<Button href={`/api/tu/archives/documents/${document.id}/file`} target="_blank" variant="outline" size="icon" title="Lihat file arsip">
																<DownloadIcon class="size-4" />
															</Button>
															<Button variant="outline" size="icon" title="Edit metadata" onclick={() => editDocument(document)}>
																<PencilIcon class="size-4" />
															</Button>
															<LoadingButton
																variant="outline"
																size="icon"
																title="Hapus arsip"
																loading={deleteBusy[document.id] === true}
																onclick={() => void deleteDocument(document.id)}
															>
																<Trash2Icon class="size-4 text-red-600" />
															</LoadingButton>
														</div>
													</Table.Cell>
												</Table.Row>
											{/each}
										</Table.Body>
									</Table.Root>
								</div>
							{/if}
						</Card.Content>
					</Card.Root>

					<Card.Root class="border-slate-200">
						<Card.Header>
							<Card.Title class="text-base">Kategori Arsip</Card.Title>
							<Card.Description>Kategori mengikat dokumen ke klasifikasi, retensi, dan kelompok administrasi TU.</Card.Description>
						</Card.Header>
						<Card.Content class="p-0">
							<div class="overflow-x-auto">
								<Table.Root>
									<Table.Header>
										<Table.Row>
											<Table.Head>Kode</Table.Head>
											<Table.Head>Nama</Table.Head>
											<Table.Head>Klasifikasi</Table.Head>
											<Table.Head>Retensi</Table.Head>
											<Table.Head>Dokumen</Table.Head>
											<Table.Head class="text-right">Aksi</Table.Head>
										</Table.Row>
									</Table.Header>
									<Table.Body>
										{#each overview.categories as category (category.id)}
											<Table.Row>
												<Table.Cell>
													<p class="font-medium text-slate-900">{category.code}</p>
													<Badge variant={category.is_active ? 'default' : 'secondary'}>{category.is_active ? 'Aktif' : 'Nonaktif'}</Badge>
												</Table.Cell>
												<Table.Cell>
													<p class="text-sm text-slate-800">{category.name}</p>
													<p class="text-xs text-slate-500">{category.description || '-'}</p>
												</Table.Cell>
												<Table.Cell>{category.classification_code || '-'}</Table.Cell>
												<Table.Cell>{category.retention_years} tahun</Table.Cell>
												<Table.Cell>{category.document_count}</Table.Cell>
												<Table.Cell class="text-right">
													<div class="flex justify-end gap-2">
														<Button variant="outline" size="icon" title="Edit kategori" onclick={() => editCategory(category)}>
															<PencilIcon class="size-4" />
														</Button>
														<LoadingButton
															variant="outline"
															size="icon"
															title="Hapus kategori"
															loading={deleteBusy[category.id] === true}
															onclick={() => void deleteCategory(category.id)}
														>
															<Trash2Icon class="size-4 text-red-600" />
														</LoadingButton>
													</div>
												</Table.Cell>
											</Table.Row>
										{/each}
									</Table.Body>
								</Table.Root>
							</div>
						</Card.Content>
					</Card.Root>
				</div>

				<div class="space-y-6">
					<Card.Root class="border-slate-200">
						<Card.Header>
							<Card.Title class="text-base">{editingDocumentId ? 'Edit Metadata Arsip' : 'Unggah Arsip'}</Card.Title>
							<Card.Description>{editingDocumentId ? 'Ubah metadata tanpa mengganti file tersimpan.' : 'Unggah file PDF, gambar, dokumen Office, teks, atau ZIP maksimal 25MB.'}</Card.Description>
						</Card.Header>
						<Card.Content>
							<form
								class="space-y-4"
								onsubmit={(event) => {
									event.preventDefault();
									void saveDocument();
								}}
							>
								<div>
									<label for="archive-category" class="text-sm font-medium text-slate-700">Kategori</label>
									<select
										id="archive-category"
										class="mt-1 h-10 w-full rounded-md border border-slate-200 bg-white px-3 text-sm text-slate-800"
										bind:value={documentForm.category_id}
									>
										<option value="">Pilih kategori</option>
										{#each activeCategories as category (category.id)}
											<option value={category.id}>{category.code} - {category.name}</option>
										{/each}
									</select>
								</div>
								<div>
									<label for="archive-title" class="text-sm font-medium text-slate-700">Judul Arsip</label>
									<Input id="archive-title" class="mt-1" bind:value={documentForm.title} placeholder="Contoh: Surat undangan rapat komite" />
								</div>
								<div class="grid gap-3 sm:grid-cols-2">
									<div>
										<label for="archive-number" class="text-sm font-medium text-slate-700">Nomor Arsip</label>
										<Input id="archive-number" class="mt-1" bind:value={documentForm.archive_number} placeholder="Opsional" />
									</div>
									<div>
										<label for="archive-status" class="text-sm font-medium text-slate-700">Status</label>
										<select
											id="archive-status"
											class="mt-1 h-10 w-full rounded-md border border-slate-200 bg-white px-3 text-sm text-slate-800"
											bind:value={documentForm.status}
										>
											{#each STATUS_OPTIONS.filter((option) => option[0]) as option (option[0])}
												<option value={option[0]}>{option[1]}</option>
											{/each}
										</select>
									</div>
								</div>
								<div class="grid gap-3 sm:grid-cols-2">
									<div>
										<label for="archive-document-date" class="text-sm font-medium text-slate-700">Tanggal Dokumen</label>
										<Input id="archive-document-date" class="mt-1" type="date" bind:value={documentForm.document_date} />
									</div>
									<div>
										<label for="archive-received-date" class="text-sm font-medium text-slate-700">Tanggal Terima/Simpan</label>
										<Input id="archive-received-date" class="mt-1" type="date" bind:value={documentForm.received_date} />
									</div>
								</div>
								<div class="grid gap-3 sm:grid-cols-2">
									<div>
										<label for="archive-retention-until" class="text-sm font-medium text-slate-700">Retensi Sampai</label>
										<Input id="archive-retention-until" class="mt-1" type="date" bind:value={documentForm.retention_until} />
									</div>
									<div>
										<label for="archive-storage-location" class="text-sm font-medium text-slate-700">Lokasi Fisik</label>
										<Input id="archive-storage-location" class="mt-1" bind:value={documentForm.storage_location} placeholder="Lemari A / Rak 2" />
									</div>
								</div>
								<div>
									<label for="archive-tags" class="text-sm font-medium text-slate-700">Tag</label>
									<Input id="archive-tags" class="mt-1" bind:value={documentForm.tags} placeholder="surat, komite, 2026" />
								</div>
								<div>
									<label for="archive-summary" class="text-sm font-medium text-slate-700">Ringkasan</label>
									<Textarea id="archive-summary" class="mt-1 min-h-20" bind:value={documentForm.summary} placeholder="Catatan singkat isi arsip" />
								</div>
								{#if !editingDocumentId}
									<div>
										<label for="archive-file" class="text-sm font-medium text-slate-700">File Arsip</label>
										<Input
											id="archive-file"
											bind:ref={fileInput}
											class="mt-1"
											type="file"
											accept=".pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.txt,.zip,image/*"
											onchange={handleFileChange}
										/>
										{#if selectedFile}
											<p class="mt-1 text-xs text-slate-500">{selectedFile.name} - {formatBytes(selectedFile.size)}</p>
										{/if}
									</div>
								{/if}
								<div class="flex flex-wrap justify-end gap-2">
									{#if editingDocumentId}
										<Button type="button" variant="outline" onclick={resetDocumentForm}>Batal Edit</Button>
									{/if}
									<LoadingButton type="submit" loading={documentBusy} loadingLabel="Menyimpan...">
										{#if editingDocumentId}
											<SaveIcon class="mr-2 size-4" />
											Simpan Metadata
										{:else}
											<UploadIcon class="mr-2 size-4" />
											Unggah Arsip
										{/if}
									</LoadingButton>
								</div>
							</form>
						</Card.Content>
					</Card.Root>

					<Card.Root class="border-slate-200">
						<Card.Header>
							<Card.Title class="text-base">{editingCategoryId ? 'Edit Kategori' : 'Tambah Kategori'}</Card.Title>
							<Card.Description>Kategori membantu arsip mengikuti klasifikasi dan jadwal retensi administrasi.</Card.Description>
						</Card.Header>
						<Card.Content>
							<form
								class="space-y-4"
								onsubmit={(event) => {
									event.preventDefault();
									void saveCategory();
								}}
							>
								<div class="grid gap-3 sm:grid-cols-2">
									<div>
										<label for="archive-category-code" class="text-sm font-medium text-slate-700">Kode</label>
										<Input id="archive-category-code" class="mt-1" bind:value={categoryForm.code} placeholder="ARS-SM" />
									</div>
									<div>
										<label for="archive-category-retention" class="text-sm font-medium text-slate-700">Retensi Tahun</label>
										<Input id="archive-category-retention" class="mt-1" type="number" min="0" bind:value={categoryForm.retention_years} />
									</div>
								</div>
								<div>
									<label for="archive-category-name" class="text-sm font-medium text-slate-700">Nama Kategori</label>
									<Input id="archive-category-name" class="mt-1" bind:value={categoryForm.name} placeholder="Surat Masuk" />
								</div>
								<div>
									<label for="archive-category-classification" class="text-sm font-medium text-slate-700">Kode Klasifikasi</label>
									<select
										id="archive-category-classification"
										class="mt-1 h-10 w-full rounded-md border border-slate-200 bg-white px-3 text-sm text-slate-800"
										bind:value={categoryForm.classification_code}
									>
										{#each CLASSIFICATION_OPTIONS as option (option[0])}
											<option value={option[0]}>{option[1]}</option>
										{/each}
									</select>
								</div>
								<div>
									<label for="archive-category-description" class="text-sm font-medium text-slate-700">Deskripsi</label>
									<Textarea id="archive-category-description" class="mt-1 min-h-20" bind:value={categoryForm.description} />
								</div>
								<label class="flex items-center gap-2 text-sm text-slate-700" for="archive-category-active">
									<input id="archive-category-active" type="checkbox" class="size-4 rounded border-slate-300" bind:checked={categoryForm.is_active} />
									Kategori aktif
								</label>
								<div class="flex flex-wrap justify-end gap-2">
									{#if editingCategoryId}
										<Button type="button" variant="outline" onclick={resetCategoryForm}>Batal Edit</Button>
									{/if}
									<LoadingButton type="submit" loading={categoryBusy} loadingLabel="Menyimpan...">
										<SaveIcon class="mr-2 size-4" />
										Simpan Kategori
									</LoadingButton>
								</div>
							</form>
						</Card.Content>
					</Card.Root>
				</div>
			</div>
		{/snippet}
	</AsyncContent>
</div>
