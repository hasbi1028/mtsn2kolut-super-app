<script lang="ts">
	import { onMount } from 'svelte';
	import type Quill from 'quill';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { toast } from '$lib/components/ui/sonner';
	import katex from 'katex';
	import 'quill/dist/quill.snow.css';
	import 'katex/dist/katex.min.css';

	let {
		value = $bindable(''),
		id = '',
		placeholder = 'Ketik di sini...',
		minRows = 5,
		compact = false,
		resizable = true,
		onImageUpload,
	}: {
		value: string;
		id?: string;
		placeholder?: string;
		minRows?: number;
		compact?: boolean;
		resizable?: boolean;
		onImageUpload?: (file: File) => Promise<string>;
	} = $props();

	type QuillToolbarModule = { addHandler: (name: string, handler: () => void) => void };
	type ImageSizePreset = { label: string; width: number };
	type QuillLeafWithDom = { domNode?: Node };

	const imageSizePresets: ImageSizePreset[] = [
		{ label: 'Kecil', width: 240 },
		{ label: 'Sedang', width: 320 },
		{ label: 'Besar', width: 480 },
		{ label: 'Full', width: 640 },
	];

	let editorElement = $state<HTMLDivElement | null>(null);
	let controlsElement = $state<HTMLDivElement | null>(null);
	let quill: Quill | null = null;
	let selectedImage = $state<HTMLImageElement | null>(null);
	let manualWidth = $state('320');
	let activePresetWidth = $state<number | null>(320);
	let isAdjustingImage = $state(false);
	let tableDialogOpen = $state(false);
	let tableRows = $state('4');
	let tableColumns = $state('2');
	let tableHasHeader = $state(true);
	let tableCaption = $state('');
	let syncingFromEditor = false;
	const minHeight = $derived(`${Math.max(compact ? 2 : 3, minRows) * (compact ? 2 : 2.5)}rem`);

	function editorHTML() {
		if (!quill) return '';
		return quill.root.innerHTML === '<p><br></p>' ? '' : quill.root.innerHTML;
	}

	function syncValueFromEditor() {
		if (!quill) return;
		syncingFromEditor = true;
		value = editorHTML();
		queueMicrotask(() => {
			syncingFromEditor = false;
		});
	}

	function hasDomNode(value: unknown): value is QuillLeafWithDom {
		return typeof value === 'object' && value !== null && 'domNode' in value;
	}

	function setImageSelection(image: HTMLImageElement | null) {
		selectedImage = image;
		if (!image) {
			activePresetWidth = null;
			return;
		}
		const widthValue = Number(image.getAttribute('width') || image.width || 320);
		manualWidth = String(widthValue || 320);
		activePresetWidth = imageSizePresets.some((preset) => preset.width === widthValue)
			? widthValue
			: null;
	}

	function getImageNodeFromLeaf(editor: Quill, index: number) {
		const [leaf] = editor.getLeaf(index) as [unknown, number];
		const domNode = hasDomNode(leaf) ? leaf.domNode : undefined;
		return domNode instanceof HTMLImageElement ? domNode : null;
	}

	function applyImageWidth(width: number) {
		if (!selectedImage || !Number.isFinite(width) || width < 16) return;
		selectedImage.setAttribute('width', String(width));
		selectedImage.removeAttribute('height');
		manualWidth = String(width);
		activePresetWidth = imageSizePresets.some((preset) => preset.width === width) ? width : null;
		syncValueFromEditor();
	}

	function applyManualWidth() {
		const parsed = Number(manualWidth);
		if (!Number.isFinite(parsed)) return;
		applyImageWidth(Math.max(16, Math.min(1200, parsed)));
	}

	function shouldKeepImageSelection() {
		if (isAdjustingImage) return true;
		const active = document.activeElement;
		return Boolean(active && controlsElement?.contains(active));
	}

	function uploadErrorMessage(error: unknown) {
		return error instanceof Error && error.message ? error.message : 'Upload gambar gagal';
	}

	async function fileToDataURL(file: File) {
		return new Promise<string>((resolve, reject) => {
			const reader = new FileReader();
			reader.onload = () => resolve(String(reader.result ?? ''));
			reader.onerror = reject;
			reader.readAsDataURL(file);
		});
	}

	async function uploadImage(file: File) {
		if (onImageUpload) return onImageUpload(file);
		return fileToDataURL(file);
	}

	function clampTableDimension(value: string, min: number, max: number, fallback: number) {
		const parsed = Number.parseInt(value, 10);
		if (!Number.isFinite(parsed)) return fallback;
		return Math.max(min, Math.min(max, parsed));
	}

	function escapeTableText(value: string) {
		return value
			.replaceAll('&', '&amp;')
			.replaceAll('<', '&lt;')
			.replaceAll('>', '&gt;')
			.replaceAll('"', '&quot;')
			.replaceAll("'", '&#39;');
	}

	function tableCellPlaceholder(rowIndex: number, columnIndex: number) {
		if (tableHasHeader && rowIndex === 0) return columnIndex === 0 ? 'Kolom 1' : `Kolom ${columnIndex + 1}`;
		return columnIndex === 0 ? `Baris ${tableHasHeader ? rowIndex : rowIndex + 1}` : '';
	}

	function buildTableHTML() {
		const rows = clampTableDimension(tableRows, 1, 20, 4);
		const columns = clampTableDimension(tableColumns, 1, 8, 2);
		const caption = tableCaption.trim();
		const bodyRows = tableHasHeader ? Math.max(1, rows - 1) : rows;
		let html = '<table class="bank-soal-table"><tbody>';
		if (caption) {
			html += '<tr>';
			for (let col = 0; col < columns; col += 1) {
				html += `<td>${col === 0 ? `<strong>${escapeTableText(caption)}</strong>` : ''}</td>`;
			}
			html += '</tr>';
		}
		if (tableHasHeader) {
			html += '<tr>';
			for (let col = 0; col < columns; col += 1) {
				html += `<td><strong>${escapeTableText(tableCellPlaceholder(0, col))}</strong></td>`;
			}
			html += '</tr>';
		}
		for (let row = 0; row < bodyRows; row += 1) {
			html += '<tr>';
			for (let col = 0; col < columns; col += 1) {
				html += `<td>${escapeTableText(tableCellPlaceholder(tableHasHeader ? row + 1 : row, col))}</td>`;
			}
			html += '</tr>';
		}
		html += '</tbody></table><p><br></p>';
		return html;
	}

	function insertTable() {
		if (!quill) return;
		const range = quill.getSelection(true) ?? { index: quill.getLength(), length: 0 };
		quill.clipboard.dangerouslyPasteHTML(range.index, buildTableHTML(), 'user');
		syncValueFromEditor();
		quill.setSelection(range.index + 1, 0);
		tableDialogOpen = false;
		toast.success('Tabel ditambahkan. Klik isi sel untuk mengubah teks tabel.');
	}

	function installImageHandler(editor: Quill) {
		const toolbar = editor.getModule('toolbar') as QuillToolbarModule;
		toolbar.addHandler('image', () => {
			const input = document.createElement('input');
			input.type = 'file';
			input.accept = 'image/*';
			input.click();

			input.onchange = async () => {
				const file = input.files?.[0];
				if (!file) return;
				try {
					const url = await uploadImage(file);
					if (!url) return;
					const range = editor.getSelection() ?? { index: editor.getLength(), length: 0 };
					editor.insertEmbed(range.index, 'image', url);
					const imageNode = getImageNodeFromLeaf(editor, range.index);
					if (imageNode) {
						imageNode.setAttribute('width', '320');
						imageNode.removeAttribute('height');
						setImageSelection(imageNode);
					}
					syncValueFromEditor();
					editor.setSelection(range.index + 1, 0);
				} catch (error) {
					toast.error(uploadErrorMessage(error));
				}
			};
		});
	}

	onMount(async () => {
		if (!editorElement) return;
		(globalThis as { katex?: typeof katex }).katex = katex;
		const QuillModule = (await import('quill')).default;
		const editor = new QuillModule(editorElement, {
			modules: {
				toolbar: [
					[{ header: [1, 2, false] }],
					[{ align: ['', 'center', 'right', 'justify'] }],
					['bold', 'italic', 'underline'],
					['formula', 'image'],
					[{ list: 'ordered' }, { list: 'bullet' }],
					['clean'],
				],
			},
			placeholder,
			theme: 'snow',
		});

		quill = editor;
		if (value) editor.root.innerHTML = value;

		editor.on('text-change', syncValueFromEditor);
		editor.on('selection-change', (range) => {
			if (!range) {
				if (shouldKeepImageSelection()) return;
				setImageSelection(null);
				return;
			}
			setImageSelection(getImageNodeFromLeaf(editor, range.index));
		});
		editor.root.addEventListener('click', (event) => {
			const target = event.target;
			setImageSelection(target instanceof HTMLImageElement ? target : null);
		});
		installImageHandler(editor);
	});

	$effect(() => {
		if (!quill || syncingFromEditor) return;
		const incoming = value || '';
		if (incoming !== editorHTML()) {
			quill.root.innerHTML = incoming;
		}
	});
</script>

<div
	class="legacy-rich-editor overflow-hidden rounded-md border border-border bg-card text-foreground shadow-sm"
	class:legacy-rich-editor--compact={compact}
	class:legacy-rich-editor--resizable={resizable}
>
	<div class="flex flex-wrap items-center justify-between gap-2 border-b border-border bg-muted/60 px-3 py-2">
		<div class="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Alat cepat</div>
		<Button type="button" variant="outline" size="sm" class="h-7 px-2 text-[10px] font-semibold uppercase tracking-wider" onclick={() => (tableDialogOpen = true)}>
			+ Tabel
		</Button>
		<Dialog.Root bind:open={tableDialogOpen}>
			<Dialog.Content>
				<Dialog.Header>
					<Dialog.Title>Tambah Tabel</Dialog.Title>
					<Dialog.Description>
						Sisipkan tabel sederhana untuk soal data, statistika, IPA, IPS, atau bacaan. Setelah masuk editor, klik isi sel untuk mengubah teksnya.
					</Dialog.Description>
				</Dialog.Header>
				<div class="space-y-4 py-2">
					<div class="grid grid-cols-2 gap-3">
						<label class="space-y-1 text-xs font-semibold text-foreground">
							<span>Jumlah baris</span>
							<Input type="number" min="1" max="20" bind:value={tableRows} />
						</label>
						<label class="space-y-1 text-xs font-semibold text-foreground">
							<span>Jumlah kolom</span>
							<Input type="number" min="1" max="8" bind:value={tableColumns} />
						</label>
					</div>
					<label class="space-y-1 text-xs font-semibold text-foreground">
						<span>Judul tabel opsional</span>
						<Input bind:value={tableCaption} placeholder="Contoh: Data berat badan siswa" />
					</label>
					<label class="flex items-start gap-2 rounded-lg border border-border bg-muted/40 p-3 text-xs text-foreground">
						<input type="checkbox" bind:checked={tableHasHeader} class="mt-0.5" />
						<span>
							<span class="block font-semibold">Gunakan baris header</span>
							<span class="text-muted-foreground">Cocok untuk tabel seperti “Berat badan (kg)” dan “Banyak orang”.</span>
						</span>
					</label>
				</div>
				<Dialog.Footer>
					<Button type="button" variant="outline" onclick={() => (tableDialogOpen = false)}>Batal</Button>
					<Button type="button" onclick={insertTable}>Masukkan Tabel</Button>
				</Dialog.Footer>
			</Dialog.Content>
		</Dialog.Root>
	</div>
	{#if selectedImage}
		<div
			bind:this={controlsElement}
			class="border-b border-success/20 bg-success/10 px-3 py-2"
			onfocusin={() => (isAdjustingImage = true)}
			onfocusout={(event) => {
				const nextTarget = event.relatedTarget;
				if (!(nextTarget instanceof Node) || !controlsElement?.contains(nextTarget)) {
					isAdjustingImage = false;
				}
			}}
		>
			<div class="flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between">
				<div class="flex flex-wrap items-center gap-1.5">
					<span class="text-[10px] font-semibold uppercase text-success">Ukuran Gambar</span>
					{#each imageSizePresets as preset (preset.width)}
						<Button
							type="button"
							size="sm"
							variant={activePresetWidth === preset.width ? 'default' : 'outline'}
							class="h-7 px-2 text-[10px]"
							onmousedown={(event) => event.preventDefault()}
							onclick={() => applyImageWidth(preset.width)}
						>
							{preset.label}
						</Button>
					{/each}
				</div>
				<div class="flex items-center gap-2">
					<Input
						type="number"
						min="16"
						max="1200"
						step="10"
						bind:value={manualWidth}
						class="h-7 w-24 bg-card text-xs"
						onkeydown={(event) => {
							if (event.key === 'Enter') {
								event.preventDefault();
								applyManualWidth();
							}
						}}
					/>
					<span class="text-[10px] font-semibold uppercase text-success">px</span>
					<Button
						type="button"
						size="sm"
						variant="outline"
						class="h-7 px-2 text-[10px]"
						onmousedown={(event) => event.preventDefault()}
						onclick={applyManualWidth}
					>
						Terapkan
					</Button>
				</div>
			</div>
		</div>
	{/if}
	<div bind:this={editorElement} {id} style:min-height={minHeight}></div>
</div>

<style>
	:global(.legacy-rich-editor .ql-toolbar.ql-snow) {
		border: none !important;
		border-bottom: 1px solid var(--border) !important;
		background: var(--muted) !important;
		padding: 8px !important;
	}
	:global(.legacy-rich-editor .ql-container.ql-snow) {
		border: none !important;
		background: var(--card) !important;
		color: var(--foreground) !important;
		font-family: inherit !important;
		font-size: 14px !important;
	}
	:global(.legacy-rich-editor--resizable .ql-container.ql-snow) {
		resize: vertical;
		overflow: auto !important;
	}
	:global(.legacy-rich-editor .ql-editor) {
		min-height: inherit;
		padding: 14px !important;
		line-height: 1.65;
	}
	:global(.legacy-rich-editor--compact .ql-toolbar.ql-snow) {
		padding: 4px !important;
	}
	:global(.legacy-rich-editor--compact .ql-toolbar.ql-snow .ql-formats) {
		margin-right: 6px !important;
	}
	:global(.legacy-rich-editor--compact .ql-toolbar.ql-snow button) {
		height: 24px !important;
		width: 24px !important;
		padding: 3px !important;
	}
	:global(.legacy-rich-editor--compact .ql-toolbar.ql-snow .ql-picker) {
		height: 24px !important;
		font-size: 12px !important;
	}
	:global(.legacy-rich-editor--compact .ql-toolbar.ql-snow .ql-picker-label) {
		padding-left: 4px !important;
		padding-right: 14px !important;
	}
	:global(.legacy-rich-editor--compact .ql-container.ql-snow) {
		font-size: 13px !important;
	}
	:global(.legacy-rich-editor--compact .ql-editor) {
		padding: 9px !important;
		line-height: 1.45;
	}
	:global(.legacy-rich-editor .ql-editor img) {
		display: block;
		max-width: 100%;
		height: auto;
		margin: 0.75rem 0;
		border-radius: 0.375rem;
	}
	:global(.legacy-rich-editor .ql-editor p) {
		line-height: 1.65;
	}
	:global(.legacy-rich-editor .ql-editor .ql-formula) {
		display: inline-flex;
		max-width: 100%;
		overflow-x: auto;
		vertical-align: middle;
	}
	:global(.legacy-rich-editor .ql-editor table) {
		width: 100%;
		max-width: 100%;
		margin: 0.75rem 0;
		border-collapse: collapse;
		overflow-x: auto;
	}
	:global(.legacy-rich-editor .ql-editor td) {
		border: 1px solid var(--border);
		padding: 0.45rem 0.55rem;
		vertical-align: top;
	}
	:global(.legacy-rich-editor .ql-editor tr:first-child td) {
		background: var(--muted);
	}
	:global(.legacy-rich-editor .ql-snow .ql-stroke) {
		stroke: var(--primary) !important;
	}
	:global(.legacy-rich-editor .ql-snow .ql-fill) {
		fill: var(--primary) !important;
	}
	:global(.legacy-rich-editor .ql-snow .ql-picker) {
		color: var(--primary) !important;
	}
	:global(.legacy-rich-editor .ql-editor.ql-blank::before) {
		color: var(--muted-foreground) !important;
		font-style: normal !important;
	}
</style>
