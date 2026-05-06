<script lang="ts">
	import { onMount } from 'svelte';
	import type Quill from 'quill';
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
