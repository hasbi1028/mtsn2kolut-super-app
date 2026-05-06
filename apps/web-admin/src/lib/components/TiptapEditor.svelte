<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { Editor } from '@tiptap/core';
	import StarterKit from '@tiptap/starter-kit';
	import Underline from '@tiptap/extension-underline';
	import Image from '@tiptap/extension-image';
	import Placeholder from '@tiptap/extension-placeholder';
	import { Table } from '@tiptap/extension-table';
	import { TableRow } from '@tiptap/extension-table-row';
	import { TableHeader } from '@tiptap/extension-table-header';
	import { TableCell } from '@tiptap/extension-table-cell';
	import Subscript from '@tiptap/extension-subscript';
	import Superscript from '@tiptap/extension-superscript';
	import TextAlign from '@tiptap/extension-text-align';
	import Color from '@tiptap/extension-color';
	import { TextStyle } from '@tiptap/extension-text-style';
	import Mathematics from '@tiptap/extension-mathematics';
	import { toast } from '$lib/components/ui/sonner';

	let {
		value = $bindable(''),
		id = '',
		placeholder = 'Tulis konten di sini...',
		minRows = 5,
		onImageUpload,
	}: {
		value: string;
		id?: string;
		placeholder?: string;
		minRows?: number;
		onImageUpload?: (file: File) => Promise<string>;
	} = $props();

	let editorEl = $state<HTMLDivElement | null>(null);
	let editor = $state<Editor | null>(null);
	let internalUpdate = false;
	let mathInput = $state('');
	let showMathDialog = $state(false);
	let mathMode = $state<'inline' | 'block'>('inline');
	let imageUploading = $state(false);
	let fileInputEl = $state<HTMLInputElement | null>(null);
	let showColorPicker = $state(false);

	const minHeight = $derived(`${minRows * 1.8}rem`);

	onMount(() => {
		if (!editorEl) return;

		editor = new Editor({
			element: editorEl,
			extensions: [
				StarterKit.configure({ heading: { levels: [1, 2, 3] } }),
				Underline,
				TextStyle,
				Color,
				TextAlign.configure({ types: ['heading', 'paragraph'] }),
				Subscript,
				Superscript,
				Image.configure({ inline: true, allowBase64: false, HTMLAttributes: { class: 'max-w-full rounded' } }),
				Placeholder.configure({ placeholder }),
				Table.configure({ resizable: false }),
				TableRow,
				TableHeader,
				TableCell,
				Mathematics,
			],
			content: value,
			onUpdate({ editor: e }) {
				internalUpdate = true;
				value = e.getHTML();
				internalUpdate = false;
			},
		});

		return () => {
			editor?.destroy();
		};
	});

	onDestroy(() => {
		editor?.destroy();
	});

	// Sync external value changes (e.g. when parent resets the form)
	$effect(() => {
		if (!internalUpdate && editor && value !== editor.getHTML()) {
			editor.commands.setContent(value || '');
		}
	});

	function cmd(fn: () => void) {
		editor?.chain().focus();
		fn();
	}

	function isActive(name: string, attrs?: Record<string, unknown>) {
		return editor?.isActive(name, attrs) ?? false;
	}

	function insertTable() {
		editor?.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run();
	}

	function openMathInline() {
		mathMode = 'inline';
		mathInput = '';
		showMathDialog = true;
	}

	function openMathBlock() {
		mathMode = 'block';
		mathInput = '';
		showMathDialog = true;
	}

	function insertMath() {
		if (!mathInput.trim() || !editor) return;
		const wrap = mathMode === 'block' ? '$$' : '$';
		editor.chain().focus().insertContent(wrap + mathInput.trim() + wrap).run();
		showMathDialog = false;
		mathInput = '';
	}

	function uploadErrorMessage(error: unknown) {
		return error instanceof Error && error.message ? error.message : 'Upload gambar gagal';
	}

	async function handleImageFile(file: File) {
		if (!file || !editor) return;
		imageUploading = true;
		let url = '';
		try {
			if (onImageUpload) {
				url = await onImageUpload(file);
			} else {
				// base64 fallback (no upload handler)
				url = await new Promise<string>((resolve, reject) => {
					const reader = new FileReader();
					reader.onload = () => resolve(reader.result as string);
					reader.onerror = reject;
					reader.readAsDataURL(file);
				});
			}
		} catch (err) {
			toast.error(uploadErrorMessage(err));
			return;
		} finally {
			imageUploading = false;
		}
		if (url) {
			editor.chain().focus().setImage({ src: url, alt: file.name }).run();
		}
	}

	function onFileChange(e: Event) {
		const file = (e.target as HTMLInputElement).files?.[0];
		if (file) void handleImageFile(file);
		if (fileInputEl) fileInputEl.value = '';
	}

	const colors = [
		'#000000', '#374151', '#1e40af', '#166534', '#9a3412', '#7c3aed',
		'#dc2626', '#d97706', '#16a34a', '#2563eb', '#7e22ce', '#db2777',
	];
</script>

<div class="tiptap-editor-wrapper overflow-hidden rounded-md border border-input bg-background">
	<!-- Toolbar -->
	<div class="flex flex-wrap items-center gap-0.5 border-b bg-muted/50 px-2 py-1">
		<!-- History -->
		<button
			type="button"
			title="Undo"
			onclick={() => editor?.chain().focus().undo().run()}
			class="toolbar-btn"
		>↩</button>
		<button
			type="button"
			title="Redo"
			onclick={() => editor?.chain().focus().redo().run()}
			class="toolbar-btn"
		>↪</button>

		<div class="divider"></div>

		<!-- Headings -->
		<button
			type="button"
			title="Heading 1"
			onclick={() => editor?.chain().focus().toggleHeading({ level: 1 }).run()}
			class="toolbar-btn font-bold {isActive('heading', { level: 1 }) ? 'active' : ''}"
		>H1</button>
		<button
			type="button"
			title="Heading 2"
			onclick={() => editor?.chain().focus().toggleHeading({ level: 2 }).run()}
			class="toolbar-btn font-semibold text-xs {isActive('heading', { level: 2 }) ? 'active' : ''}"
		>H2</button>
		<button
			type="button"
			title="Heading 3"
			onclick={() => editor?.chain().focus().toggleHeading({ level: 3 }).run()}
			class="toolbar-btn text-xs {isActive('heading', { level: 3 }) ? 'active' : ''}"
		>H3</button>
		<button
			type="button"
			title="Paragraf"
			onclick={() => editor?.chain().focus().setParagraph().run()}
			class="toolbar-btn text-xs {isActive('paragraph') ? 'active' : ''}"
		>P</button>

		<div class="divider"></div>

		<!-- Inline formatting -->
		<button
			type="button"
			title="Tebal (Ctrl+B)"
			onclick={() => editor?.chain().focus().toggleBold().run()}
			class="toolbar-btn font-bold {isActive('bold') ? 'active' : ''}"
		>B</button>
		<button
			type="button"
			title="Miring (Ctrl+I)"
			onclick={() => editor?.chain().focus().toggleItalic().run()}
			class="toolbar-btn italic {isActive('italic') ? 'active' : ''}"
		>I</button>
		<button
			type="button"
			title="Garis Bawah (Ctrl+U)"
			onclick={() => editor?.chain().focus().toggleUnderline().run()}
			class="toolbar-btn underline {isActive('underline') ? 'active' : ''}"
		>U</button>
		<button
			type="button"
			title="Coret"
			onclick={() => editor?.chain().focus().toggleStrike().run()}
			class="toolbar-btn line-through {isActive('strike') ? 'active' : ''}"
		>S</button>
		<button
			type="button"
			title="Superscript — untuk pangkat (e.g. x²)"
			onclick={() => editor?.chain().focus().toggleSuperscript().run()}
			class="toolbar-btn {isActive('superscript') ? 'active' : ''}"
		>x<sup>2</sup></button>
		<button
			type="button"
			title="Subscript — untuk indeks (e.g. H₂O)"
			onclick={() => editor?.chain().focus().toggleSubscript().run()}
			class="toolbar-btn {isActive('subscript') ? 'active' : ''}"
		>x<sub>2</sub></button>

		<div class="divider"></div>

		<!-- Lists -->
		<button
			type="button"
			title="Daftar Tidak Berurut"
			onclick={() => editor?.chain().focus().toggleBulletList().run()}
			class="toolbar-btn {isActive('bulletList') ? 'active' : ''}"
		>• List</button>
		<button
			type="button"
			title="Daftar Berurut"
			onclick={() => editor?.chain().focus().toggleOrderedList().run()}
			class="toolbar-btn {isActive('orderedList') ? 'active' : ''}"
		>1. List</button>

		<div class="divider"></div>

		<!-- Alignment -->
		<button
			type="button"
			title="Rata Kiri"
			onclick={() => editor?.chain().focus().setTextAlign('left').run()}
			class="toolbar-btn {isActive('paragraph', { textAlign: 'left' }) || isActive('heading', { textAlign: 'left' }) ? 'active' : ''}"
		>⬤◻</button>
		<button
			type="button"
			title="Rata Tengah"
			onclick={() => editor?.chain().focus().setTextAlign('center').run()}
			class="toolbar-btn {isActive('paragraph', { textAlign: 'center' }) || isActive('heading', { textAlign: 'center' }) ? 'active' : ''}"
		>◻⬤◻</button>
		<button
			type="button"
			title="Rata Kanan"
			onclick={() => editor?.chain().focus().setTextAlign('right').run()}
			class="toolbar-btn {isActive('paragraph', { textAlign: 'right' }) || isActive('heading', { textAlign: 'right' }) ? 'active' : ''}"
		>◻⬤</button>

		<div class="divider"></div>

		<!-- Color -->
		<div class="relative">
			<button
				type="button"
				title="Warna Teks"
				onclick={() => (showColorPicker = !showColorPicker)}
				class="toolbar-btn"
			>A<span class="ml-0.5 inline-block h-1.5 w-4 rounded-sm bg-current"></span></button>
			{#if showColorPicker}
				<div
					class="absolute left-0 top-full z-20 mt-1 flex flex-wrap gap-1 rounded-md border bg-card p-2 shadow-lg"
					style="width:130px"
				>
					{#each colors as color (color)}
						<button
							type="button"
							class="h-5 w-5 rounded border border-border"
							style="background:{color}"
							title={color}
							aria-label={`Pilih warna ${color}`}
							onclick={() => {
								editor?.chain().focus().setColor(color).run();
								showColorPicker = false;
							}}
						></button>
					{/each}
					<button
						type="button"
						class="mt-1 w-full rounded border border-border px-1 py-0.5 text-xs text-muted-foreground hover:bg-muted"
						onclick={() => {
							editor?.chain().focus().unsetColor().run();
							showColorPicker = false;
						}}
					>Reset</button>
				</div>
			{/if}
		</div>

		<div class="divider"></div>

		<!-- Table -->
		<button
			type="button"
			title="Sisipkan tabel 3×3"
			onclick={insertTable}
			class="toolbar-btn"
		>Tabel</button>

		<div class="divider"></div>

		<!-- Math -->
		<button
			type="button"
			title="Sisipkan rumus LaTeX inline (mis. x²+5x=0)"
			onclick={openMathInline}
			class="toolbar-btn font-serif"
		>∑ inline</button>
		<button
			type="button"
			title="Sisipkan rumus LaTeX blok (tampil di baris sendiri)"
			onclick={openMathBlock}
			class="toolbar-btn font-serif"
		>∑ blok</button>

		<div class="divider"></div>

		<!-- Image -->
		<button
			type="button"
			title="Upload gambar"
			onclick={() => fileInputEl?.click()}
			disabled={imageUploading}
			class="toolbar-btn"
		>{imageUploading ? '⏳' : '🖼'}</button>
		<input
			bind:this={fileInputEl}
			type="file"
			accept="image/*"
			class="hidden"
			onchange={onFileChange}
		/>
	</div>

	<!-- Editor area -->
	<div
		bind:this={editorEl}
		{id}
		class="tiptap-content px-4 py-3 focus-within:outline-none"
		style="min-height:{minHeight}"
	></div>
</div>

<!-- Math input dialog -->
{#if showMathDialog}
	<div
		role="presentation"
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
		onclick={(e) => { if (e.target === e.currentTarget) showMathDialog = false; }}
		onkeydown={(e) => { if (e.key === 'Escape') showMathDialog = false; }}
	>
		<div class="w-full max-w-md rounded-xl border bg-card p-5 shadow-xl">
			<h3 class="mb-3 text-sm font-semibold text-foreground">
				Masukkan LaTeX — {mathMode === 'inline' ? 'Inline (dalam kalimat)' : 'Blok (baris sendiri)'}
			</h3>
			<textarea
				class="block w-full rounded-md border border-input bg-background px-3 py-2 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-ring"
				rows={3}
				placeholder={mathMode === 'inline' ? 'x^2 + 5x + 6 = 0' : '\\frac{-b \\pm \\sqrt{b^2 - 4ac}}{2a}'}
				bind:value={mathInput}
				onkeydown={(e) => { if (e.key === 'Enter' && e.ctrlKey) { e.preventDefault(); insertMath(); } }}
			></textarea>
			<p class="mt-1 text-xs text-muted-foreground">Tekan Ctrl+Enter untuk sisipkan. Gunakan sintaks LaTeX standar.</p>
			<div class="mt-3 flex gap-2">
				<button
					type="button"
					onclick={insertMath}
					class="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
				>Sisipkan</button>
				<button
					type="button"
					onclick={() => (showMathDialog = false)}
					class="rounded-md border border-border px-4 py-2 text-sm text-foreground hover:bg-muted/50"
				>Batal</button>
			</div>
		</div>
	</div>
{/if}

<style>
	:global(.tiptap-content .ProseMirror) {
		outline: none;
		min-height: inherit;
	}

	:global(.tiptap-content .ProseMirror p.is-editor-empty:first-child::before) {
		content: attr(data-placeholder);
		float: left;
		color: var(--muted-foreground);
		pointer-events: none;
		height: 0;
	}

	/* Heading styles */
	:global(.tiptap-content h1) { font-size: 1.25rem; font-weight: 700; margin: 0.75rem 0 0.5rem; }
	:global(.tiptap-content h2) { font-size: 1.1rem; font-weight: 600; margin: 0.6rem 0 0.4rem; }
	:global(.tiptap-content h3) { font-size: 1rem; font-weight: 600; margin: 0.5rem 0 0.3rem; }

	/* Paragraph */
	:global(.tiptap-content p) { margin: 0.25rem 0; line-height: 1.6; font-size: 0.875rem; }

	/* Lists */
	:global(.tiptap-content ul) { list-style-type: disc; padding-left: 1.5rem; margin: 0.5rem 0; }
	:global(.tiptap-content ol) { list-style-type: decimal; padding-left: 1.5rem; margin: 0.5rem 0; }
	:global(.tiptap-content li) { font-size: 0.875rem; line-height: 1.6; }

	/* Table */
	:global(.tiptap-content table) { border-collapse: collapse; width: 100%; margin: 0.5rem 0; }
	:global(.tiptap-content td, .tiptap-content th) { border: 1px solid var(--border); padding: 6px 10px; font-size: 0.875rem; }
	:global(.tiptap-content th) { background: var(--muted); font-weight: 600; }
	:global(.tiptap-content .selectedCell) { background: var(--accent); }

	/* Image */
	:global(.tiptap-content img) { max-width: 100%; border-radius: 4px; margin: 4px 0; }
	:global(.tiptap-content img.ProseMirror-selectednode) { outline: 2px solid var(--ring); }

	/* Math */
	:global(.tiptap-content .math-inline, .tiptap-content .math-node) {
		display: inline-block;
		vertical-align: middle;
		cursor: pointer;
		border-radius: 2px;
		padding: 0 2px;
	}
	:global(.tiptap-content .math-inline:hover, .tiptap-content .math-node:hover) {
		background: var(--accent);
		outline: 1px solid var(--ring);
	}
	:global(.tiptap-content .math-block) {
		display: block;
		text-align: center;
		margin: 0.75rem 0;
		padding: 0.5rem;
		background: var(--muted);
		border-radius: 4px;
		overflow-x: auto;
	}

	/* Toolbar */
	:global(.toolbar-btn) {
		padding: 3px 7px;
		border-radius: 4px;
		font-size: 0.8rem;
		color: var(--foreground);
		line-height: 1.4;
		white-space: nowrap;
	}
	:global(.toolbar-btn:hover) { background: var(--muted); }
	:global(.toolbar-btn.active) { background: var(--primary); color: var(--primary-foreground); }
	:global(.divider) { width: 1px; height: 1rem; background: var(--border); margin: 0 2px; }
</style>
