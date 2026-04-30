<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import { toast } from '$lib/components/ui/sonner';

	type ContentStatus = 'draft' | 'published';
	type ContentKind = 'page' | 'post' | 'announcement';
	type WebsiteContent = {
		id: string;
		kind: ContentKind;
		title: string;
		slug: string;
		excerpt: string;
		content_html: string;
		cover_image_url: string;
		status: ContentStatus;
		published_at: string | null;
		updated_at: string;
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

	let items = $state<WebsiteContent[]>([]);
	let loading = $state(true);
	let search = $state('');
	let showDialog = $state(false);
	let saving = $state(false);
	let editingId = $state<string | null>(null);
	let form = $state({
		title: '',
		slug: '',
		excerpt: '',
		content_html: '',
		cover_image_url: '',
		status: 'draft' as ContentStatus,
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
			status: 'draft',
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
			status: item.status,
		};
		showDialog = true;
	}

	function fmtDate(value: string | null) {
		if (!value) return 'Draft';
		return new Date(value).toLocaleDateString('id-ID', {
			timeZone: 'Asia/Makassar',
			year: 'numeric',
			month: 'short',
			day: 'numeric',
		});
	}

	function publicHref(item: WebsiteContent) {
		if (kind === 'page') {
			if (item.slug === 'ppdb-info') return '/ppdb';
			return `/${item.slug}`;
		}
		return `${publicBasePath}/${item.slug}`;
	}

	async function load() {
		loading = true;
		try {
			const res = await fetch(`/api/website/content?kind=${kind}`);
			const data = await res.json();
			items = data.items ?? [];
		} catch {
			toast.error(`Gagal memuat ${title.toLowerCase()}.`);
		} finally {
			loading = false;
		}
	}

	async function save() {
		saving = true;
		try {
			const payload = { kind, ...form };
			const res = await fetch(editingId ? `/api/website/content/${editingId}` : '/api/website/content', {
				method: editingId ? 'PUT' : 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify(payload),
			});
			const data = await res.json().catch(() => ({}));
			if (!res.ok) {
				toast.error((data as { error?: string }).error || 'Gagal menyimpan konten.');
				return;
			}
			toast.success(editingId ? 'Konten berhasil diperbarui.' : 'Konten berhasil dibuat.');
			showDialog = false;
			resetForm();
			await load();
		} finally {
			saving = false;
		}
	}

	async function remove(item: WebsiteContent) {
		if (!confirm(`Hapus konten "${item.title}"?`)) return;
		const res = await fetch(`/api/website/content/${item.id}`, { method: 'DELETE' });
		if (!res.ok) {
			toast.error('Gagal menghapus konten.');
			return;
		}
		toast.success('Konten berhasil dihapus.');
		await load();
	}

	async function toggleStatus(item: WebsiteContent) {
		const nextStatus: ContentStatus = item.status === 'published' ? 'draft' : 'published';
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
				status: nextStatus,
			}),
		});
		const data = await res.json().catch(() => ({}));
		if (!res.ok) {
			toast.error((data as { error?: string }).error || 'Gagal mengubah status konten.');
			return;
		}
		toast.success(nextStatus === 'published' ? 'Konten dipublikasikan.' : 'Konten dikembalikan ke draft.');
		await load();
	}

	onMount(() => {
		void load();
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

	<Card.Root class="overflow-hidden border-slate-200 shadow-sm">
		<Card.Content class="p-0">
			{#if loading}
				<div class="px-5 py-10 text-center text-sm text-slate-500">Memuat konten...</div>
			{:else if filteredItems.length === 0}
				<div class="px-5 py-10 text-center text-sm text-slate-500">Belum ada konten untuk kategori ini.</div>
			{:else}
				<div class="divide-y divide-slate-200">
					{#each filteredItems as item (item.id)}
						<div class="grid gap-4 px-5 py-5 lg:grid-cols-[1.3fr,0.7fr,0.8fr] lg:items-start">
							<div class="space-y-2">
								<div class="flex flex-wrap items-center gap-2">
									<h2 class="text-lg font-semibold text-slate-900">{item.title}</h2>
									<Badge variant={item.status === 'published' ? 'outline' : 'secondary'} class={item.status === 'published' ? 'border-emerald-300 text-emerald-700' : ''}>
										{item.status === 'published' ? 'Published' : 'Draft'}
									</Badge>
								</div>
								<p class="font-mono text-xs text-slate-500">{publicHref(item)}</p>
								<p class="text-sm leading-7 text-slate-600">{item.excerpt || 'Belum ada ringkasan.'}</p>
							</div>
							<div class="space-y-2 text-sm text-slate-600">
								<p><span class="font-medium text-slate-900">Tayang:</span> {fmtDate(item.published_at)}</p>
								<p><span class="font-medium text-slate-900">Update:</span> {fmtDate(item.updated_at)}</p>
							</div>
							<div class="flex flex-wrap justify-start gap-2 lg:justify-end">
								{#if item.status === 'published'}
									<a href={publicHref(item)} target="_blank" rel="noreferrer">
										<Button variant="outline">Lihat</Button>
									</a>
								{/if}
								<Button variant="outline" onclick={() => openEdit(item)}>Edit</Button>
								<Button variant="outline" onclick={() => toggleStatus(item)}>
									{item.status === 'published' ? 'Jadikan Draft' : 'Publish'}
								</Button>
								<Button variant="ghost" class="text-destructive hover:text-destructive" onclick={() => remove(item)}>
									Hapus
								</Button>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</Card.Content>
	</Card.Root>
</div>

<Dialog.Root bind:open={showDialog}>
	<Dialog.Content>
		<div class="space-y-4">
			<div>
				<h2 class="text-base font-semibold text-slate-900">{editingId ? 'Edit Konten' : 'Tambah Konten'}</h2>
				<p class="mt-1 text-sm text-slate-500">Kelola konten {title.toLowerCase()} dengan workflow draft ke publish.</p>
			</div>

			<div class="grid gap-3 sm:grid-cols-2">
				<div class="sm:col-span-2">
					<label for="website-title" class="mb-1 block text-xs font-medium text-slate-600">Judul</label>
					<Input id="website-title" bind:value={form.title} />
				</div>
				<div>
					<label for="website-slug" class="mb-1 block text-xs font-medium text-slate-600">Slug</label>
					<Input id="website-slug" bind:value={form.slug} placeholder="otomatis jika dikosongkan" />
				</div>
				<div>
					<label for="website-status" class="mb-1 block text-xs font-medium text-slate-600">Status</label>
					<select id="website-status" bind:value={form.status} class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm">
						<option value="draft">Draft</option>
						<option value="published">Published</option>
					</select>
				</div>
				<div class="sm:col-span-2">
					<label for="website-cover" class="mb-1 block text-xs font-medium text-slate-600">Cover Image URL</label>
					<Input id="website-cover" bind:value={form.cover_image_url} placeholder="https://..." />
				</div>
				<div class="sm:col-span-2">
					<label for="website-excerpt" class="mb-1 block text-xs font-medium text-slate-600">Ringkasan</label>
					<Textarea id="website-excerpt" rows={3} bind:value={form.excerpt} />
				</div>
				<div class="sm:col-span-2">
					<label for="website-content" class="mb-1 block text-xs font-medium text-slate-600">Konten HTML</label>
					<Textarea id="website-content" rows={12} bind:value={form.content_html} />
				</div>
			</div>

			<div class="flex justify-end gap-2">
				<Button variant="outline" onclick={() => (showDialog = false)}>Batal</Button>
				<Button onclick={save} disabled={saving}>{saving ? 'Menyimpan...' : 'Simpan'}</Button>
			</div>
		</div>
	</Dialog.Content>
</Dialog.Root>
