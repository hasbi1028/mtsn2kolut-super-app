<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import katex from 'katex';

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
	let fileInputEl = $state<HTMLInputElement | null>(null);
	let internalUpdate = false;
	let imageUploading = $state(false);
	let uploadError = $state('');
	let showColorPicker = $state(false);

	// Math dialog
	let showMathDialog = $state(false);
	let mathInput = $state('');
	let mathMode = $state<'inline' | 'block'>('inline');
	let mathTargetEl = $state<HTMLElement | null>(null); // non-null = re-editing existing node

	// Toolbar state
	let stateBold = $state(false);
	let stateItalic = $state(false);
	let stateUnderline = $state(false);
	let stateStrike = $state(false);
	let stateBlock = $state('');
	let stateInTable = $state(false);

	const minHeight = $derived(`${minRows * 1.8}rem`);

	// ── Helpers ──────────────────────────────────────────────────────────────

	function findAncestor<T extends HTMLElement>(
		node: Node | null,
		predicate: (n: Node) => n is T,
	): T | null {
		let cur: Node | null = node;
		while (cur && cur !== editorEl) {
			if (predicate(cur)) return cur;
			cur = cur.parentNode;
		}
		return null;
	}

	function isCell(n: Node): n is HTMLTableCellElement {
		return n instanceof HTMLTableCellElement;
	}

	function currentCell(): HTMLTableCellElement | null {
		const sel = window.getSelection();
		if (!sel || sel.rangeCount === 0 || !editorEl) return null;
		return findAncestor(sel.anchorNode, isCell);
	}

	function escapeAttr(str: string) {
		return str.replace(/&/g, '&amp;').replace(/"/g, '&quot;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
	}

	// ── Toolbar state refresh ─────────────────────────────────────────────────

	function refreshToolbar() {
		stateBold      = document.queryCommandState('bold');
		stateItalic    = document.queryCommandState('italic');
		stateUnderline = document.queryCommandState('underline');
		stateStrike    = document.queryCommandState('strikeThrough');
		stateBlock     = document.queryCommandValue('formatBlock').toLowerCase();
		stateInTable   = currentCell() !== null;
	}

	function onSelectionChange() {
		if (editorEl?.contains(document.activeElement)) refreshToolbar();
	}

	// ── Core exec ────────────────────────────────────────────────────────────

	function exec(cmd: string, arg?: string) {
		editorEl?.focus();
		document.execCommand(cmd, false, arg ?? undefined);
		syncOut();
		refreshToolbar();
	}

	function syncOut() {
		if (!editorEl) return;
		internalUpdate = true;
		value = editorEl.innerHTML;
		internalUpdate = false;
	}

	$effect(() => {
		if (!internalUpdate && editorEl && editorEl.innerHTML !== value) {
			editorEl.innerHTML = value || '';
		}
	});

	// ── Table ────────────────────────────────────────────────────────────────

	function insertTable() {
		const html = `
<table class="ce-table">
  <thead><tr><th><br></th><th><br></th><th><br></th></tr></thead>
  <tbody>
    <tr><td><br></td><td><br></td><td><br></td></tr>
    <tr><td><br></td><td><br></td><td><br></td></tr>
  </tbody>
</table><p><br></p>`.trim();
		exec('insertHTML', html);
	}

	function tableInsertRow(below = true) {
		const td = currentCell();
		const tr = td?.closest('tr');
		if (!tr) return;
		const colCount = td!.closest('table')!.rows[0].cells.length;
		const newTr = document.createElement('tr');
		for (let i = 0; i < colCount; i++) {
			const cell = document.createElement('td');
			cell.innerHTML = '<br>';
			newTr.appendChild(cell);
		}
		if (below) tr.after(newTr); else tr.before(newTr);
		syncOut();
	}

	function tableInsertCol(right = true) {
		const td = currentCell();
		const tr = td?.closest('tr');
		const table = tr?.closest('table') as HTMLTableElement | null;
		if (!table || !tr || !td) return;
		const idx = Array.from(tr.cells).indexOf(td);
		const insertIdx = right ? idx + 1 : idx;
		for (const row of Array.from(table.rows)) {
			const cell = row.insertCell(insertIdx);
			cell.innerHTML = '<br>';
		}
		syncOut();
	}

	function tableDeleteRow() {
		const td = currentCell();
		const tr = td?.closest('tr') as HTMLTableRowElement | null;
		if (!tr) return;
		const table = tr.closest('table') as HTMLTableElement;
		if (table.rows.length <= 1) table.remove();
		else tr.remove();
		syncOut();
		stateInTable = false;
	}

	function tableDeleteCol() {
		const td = currentCell();
		const tr = td?.closest('tr');
		const table = tr?.closest('table') as HTMLTableElement | null;
		if (!table || !tr || !td) return;
		const idx = Array.from(tr.cells).indexOf(td);
		for (const row of Array.from(table.rows)) {
			if (row.cells[idx]) row.deleteCell(idx);
		}
		if ((table.rows[0]?.cells.length ?? 0) === 0) table.remove();
		syncOut();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key !== 'Tab') return;
		const td = currentCell();
		if (!td) return;
		e.preventDefault();
		const table = td.closest('table') as HTMLTableElement;
		const cells = Array.from(table.querySelectorAll('td, th')) as HTMLTableCellElement[];
		const idx = cells.indexOf(td);
		const next = e.shiftKey ? cells[idx - 1] : cells[idx + 1];
		if (next) {
			// Move cursor into next cell
			const range = document.createRange();
			range.selectNodeContents(next);
			range.collapse(false);
			const sel = window.getSelection();
			sel?.removeAllRanges();
			sel?.addRange(range);
		} else if (!e.shiftKey) {
			// Add new row when Tab past last cell
			tableInsertRow(true);
		}
	}

	// ── Math (LaTeX via KaTeX) ────────────────────────────────────────────────

	function openMathInline() { mathMode = 'inline'; mathInput = ''; mathTargetEl = null; showMathDialog = true; }
	function openMathBlock()  { mathMode = 'block';  mathInput = ''; mathTargetEl = null; showMathDialog = true; }

	function renderMathHtml(formula: string, block: boolean): string {
		const rendered = katex.renderToString(formula, {
			displayMode: block,
			throwOnError: false,
			output: 'html',
			trust: false,
		});
		if (block) {
			return `<div contenteditable="false" class="ce-math-block" data-latex="${escapeAttr(formula)}">${rendered}</div>`;
		}
		return `<span contenteditable="false" class="ce-math-inline" data-latex="${escapeAttr(formula)}">${rendered}</span>`;
	}

	function insertMath() {
		const formula = mathInput.trim();
		if (!formula || !editorEl) return;
		const html = renderMathHtml(formula, mathMode === 'block');

		if (mathTargetEl) {
			// Replace existing math node
			mathTargetEl.outerHTML = html;
			mathTargetEl = null;
			syncOut();
		} else {
			exec('insertHTML', html);
		}
		showMathDialog = false;
		mathInput = '';
	}

	// Click on a rendered math node to re-edit it
	function handleEditorClick(e: MouseEvent) {
		const target = e.target as HTMLElement;
		const mathEl = target.closest('.ce-math-inline, .ce-math-block') as HTMLElement | null;
		if (!mathEl) return;
		e.preventDefault();
		mathTargetEl = mathEl;
		mathInput = mathEl.dataset.latex ?? '';
		mathMode = mathEl.classList.contains('ce-math-block') ? 'block' : 'inline';
		showMathDialog = true;
	}

	// ── Image ────────────────────────────────────────────────────────────────

	async function handleImageFile(file: File) {
		if (!file || !editorEl) return;
		imageUploading = true;
		uploadError = '';
		try {
			let url = '';
			if (onImageUpload) {
				url = await onImageUpload(file);
			} else {
				url = await new Promise<string>((resolve, reject) => {
					const reader = new FileReader();
					reader.onload = () => resolve(reader.result as string);
					reader.onerror = reject;
					reader.readAsDataURL(file);
				});
			}
			if (url) exec('insertHTML', `<img src="${url}" alt="${escapeAttr(file.name)}" class="custom-editor-img" />`);
		} catch (err) {
			uploadError = (err as Error).message || 'Upload gambar gagal';
		} finally {
			imageUploading = false;
		}
	}

	function onFileChange(e: Event) {
		const file = (e.target as HTMLInputElement).files?.[0];
		uploadError = '';
		if (file) void handleImageFile(file);
		if (fileInputEl) fileInputEl.value = '';
	}

	onMount(() => {
		if (editorEl) editorEl.innerHTML = value || '';
		document.addEventListener('selectionchange', onSelectionChange);
	});

	onDestroy(() => {
		document.removeEventListener('selectionchange', onSelectionChange);
	});

	const colors = [
		'#000000', '#374151', '#1e40af', '#166534', '#9a3412', '#7c3aed',
		'#dc2626', '#d97706', '#16a34a', '#2563eb', '#7e22ce', '#db2777',
	];
</script>

<div class="custom-editor-wrapper overflow-hidden rounded-md border border-input bg-background">
	<!-- Main toolbar -->
	<div class="flex flex-wrap items-center gap-0.5 border-b bg-slate-50 px-2 py-1">
		<button type="button" title="Undo" onclick={() => exec('undo')} class="ce-btn">↩</button>
		<button type="button" title="Redo" onclick={() => exec('redo')} class="ce-btn">↪</button>
		<div class="ce-div"></div>

		<button type="button" title="Heading 1" onclick={() => exec('formatBlock', 'h1')} class="ce-btn font-bold {stateBlock === 'h1' ? 'active' : ''}">H1</button>
		<button type="button" title="Heading 2" onclick={() => exec('formatBlock', 'h2')} class="ce-btn font-semibold text-xs {stateBlock === 'h2' ? 'active' : ''}">H2</button>
		<button type="button" title="Heading 3" onclick={() => exec('formatBlock', 'h3')} class="ce-btn text-xs {stateBlock === 'h3' ? 'active' : ''}">H3</button>
		<button type="button" title="Paragraf" onclick={() => exec('formatBlock', 'p')} class="ce-btn text-xs {stateBlock === 'p' || stateBlock === '' ? 'active' : ''}">P</button>
		<div class="ce-div"></div>

		<button type="button" title="Tebal (Ctrl+B)" onclick={() => exec('bold')} class="ce-btn font-bold {stateBold ? 'active' : ''}">B</button>
		<button type="button" title="Miring (Ctrl+I)" onclick={() => exec('italic')} class="ce-btn italic {stateItalic ? 'active' : ''}">I</button>
		<button type="button" title="Garis Bawah (Ctrl+U)" onclick={() => exec('underline')} class="ce-btn underline {stateUnderline ? 'active' : ''}">U</button>
		<button type="button" title="Coret" onclick={() => exec('strikeThrough')} class="ce-btn line-through {stateStrike ? 'active' : ''}">S</button>
		<div class="ce-div"></div>

		<button type="button" title="Daftar Tidak Berurut" onclick={() => exec('insertUnorderedList')} class="ce-btn">• List</button>
		<button type="button" title="Daftar Berurut" onclick={() => exec('insertOrderedList')} class="ce-btn">1. List</button>
		<div class="ce-div"></div>

		<button type="button" title="Rata Kiri" onclick={() => exec('justifyLeft')} class="ce-btn">⬤◻</button>
		<button type="button" title="Rata Tengah" onclick={() => exec('justifyCenter')} class="ce-btn">◻⬤◻</button>
		<button type="button" title="Rata Kanan" onclick={() => exec('justifyRight')} class="ce-btn">◻⬤</button>
		<div class="ce-div"></div>

		<!-- Color picker -->
		<div class="relative">
			<button type="button" title="Warna Teks" onclick={() => (showColorPicker = !showColorPicker)} class="ce-btn">
				A<span class="ml-0.5 inline-block h-1.5 w-4 rounded-sm bg-current"></span>
			</button>
			{#if showColorPicker}
				<!-- svelte-ignore a11y_no_static_element_interactions -->
				<div
					class="absolute left-0 top-full z-20 mt-1 flex flex-wrap gap-1 rounded-md border bg-white p-2 shadow-lg"
					style="width:130px"
					onmouseleave={() => (showColorPicker = false)}
				>
					{#each colors as color (color)}
						<button
							type="button"
							class="h-5 w-5 rounded border border-slate-200"
							style="background:{color}"
							title={color}
							onclick={() => { exec('foreColor', color); showColorPicker = false; }}
						></button>
					{/each}
					<button
						type="button"
						class="mt-1 w-full rounded border border-slate-200 px-1 py-0.5 text-xs text-slate-600 hover:bg-slate-100"
						onclick={() => { exec('removeFormat'); showColorPicker = false; }}
					>Reset</button>
				</div>
			{/if}
		</div>
		<div class="ce-div"></div>

		<!-- Table -->
		<button type="button" title="Sisipkan tabel 3×3" onclick={insertTable} class="ce-btn {stateInTable ? 'active' : ''}">Tabel</button>
		<div class="ce-div"></div>

		<!-- Math -->
		<button type="button" title="Sisipkan rumus LaTeX inline" onclick={openMathInline} class="ce-btn font-serif">∑ inline</button>
		<button type="button" title="Sisipkan rumus LaTeX blok" onclick={openMathBlock} class="ce-btn font-serif">∑ blok</button>
		<div class="ce-div"></div>

		<!-- Image -->
		<button type="button" title="Upload gambar" onclick={() => fileInputEl?.click()} disabled={imageUploading} class="ce-btn">
			{imageUploading ? '⏳' : '🖼'}
		</button>
		<input bind:this={fileInputEl} type="file" accept="image/*" class="hidden" onchange={onFileChange} />
	</div>

	<!-- Table context toolbar — only visible when cursor is inside a table cell -->
	{#if stateInTable}
		<div class="flex flex-wrap items-center gap-0.5 border-b bg-sky-50 px-2 py-1">
			<span class="mr-1 text-xs font-semibold text-sky-700">Tabel:</span>
			<button type="button" onclick={() => tableInsertRow(false)} class="ce-btn text-xs">+ Baris Atas</button>
			<button type="button" onclick={() => tableInsertRow(true)}  class="ce-btn text-xs">+ Baris Bawah</button>
			<button type="button" onclick={() => tableInsertCol(false)} class="ce-btn text-xs">+ Kolom Kiri</button>
			<button type="button" onclick={() => tableInsertCol(true)}  class="ce-btn text-xs">+ Kolom Kanan</button>
			<div class="ce-div"></div>
			<button type="button" onclick={tableDeleteRow} class="ce-btn text-xs text-rose-600 hover:bg-rose-50">Hapus Baris</button>
			<button type="button" onclick={tableDeleteCol} class="ce-btn text-xs text-rose-600 hover:bg-rose-50">Hapus Kolom</button>
			<span class="ml-1 text-xs text-sky-500">Tab = pindah sel</span>
		</div>
	{/if}

	<!-- Editable area -->
	<div
		bind:this={editorEl}
		{id}
		contenteditable="true"
		role="textbox"
		tabindex="0"
		aria-multiline="true"
		aria-label={placeholder}
		data-placeholder={placeholder}
		class="custom-editor-content px-4 py-3 focus:outline-none"
		style="min-height:{minHeight}"
		oninput={syncOut}
		onkeydown={handleKeydown}
		onkeyup={refreshToolbar}
		onmouseup={refreshToolbar}
		onclick={handleEditorClick}
	></div>
</div>

{#if uploadError}
	<p class="mt-1 text-xs text-rose-600">{uploadError}</p>
{/if}

<!-- Math input dialog -->
{#if showMathDialog}
	<div
		role="presentation"
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
		onclick={(e) => { if (e.target === e.currentTarget) { showMathDialog = false; mathTargetEl = null; } }}
		onkeydown={(e) => { if (e.key === 'Escape') { showMathDialog = false; mathTargetEl = null; } }}
	>
		<div class="w-full max-w-md rounded-xl border bg-white p-5 shadow-xl">
			<h3 class="mb-3 text-sm font-semibold text-slate-900">
				Rumus LaTeX — {mathMode === 'inline' ? 'Inline (dalam kalimat)' : 'Blok (baris sendiri)'}
				{#if mathTargetEl}<span class="ml-2 text-xs font-normal text-slate-400">(mengedit ulang)</span>{/if}
			</h3>
			<div class="mb-2 flex gap-2">
				<button
					type="button"
					onclick={() => (mathMode = 'inline')}
					class="rounded-md border px-3 py-1 text-xs font-medium transition {mathMode === 'inline' ? 'border-emerald-300 bg-emerald-50 text-emerald-800' : 'border-slate-200 text-slate-600 hover:bg-slate-50'}"
				>Inline</button>
				<button
					type="button"
					onclick={() => (mathMode = 'block')}
					class="rounded-md border px-3 py-1 text-xs font-medium transition {mathMode === 'block' ? 'border-emerald-300 bg-emerald-50 text-emerald-800' : 'border-slate-200 text-slate-600 hover:bg-slate-50'}"
				>Blok</button>
			</div>
			<textarea
				class="block w-full rounded-md border border-input bg-background px-3 py-2 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
				rows={3}
				placeholder={mathMode === 'inline' ? 'x^2 + 5x + 6 = 0' : '\\frac{-b \\pm \\sqrt{b^2-4ac}}{2a}'}
				bind:value={mathInput}
				onkeydown={(e) => { if (e.key === 'Enter' && e.ctrlKey) { e.preventDefault(); insertMath(); } }}
			></textarea>
			<p class="mt-1 text-xs text-slate-500">Tekan Ctrl+Enter untuk sisipkan.</p>
			<div class="mt-3 flex gap-2">
				<button
					type="button"
					onclick={insertMath}
					class="rounded-md bg-emerald-700 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-800"
				>Sisipkan</button>
				<button
					type="button"
					onclick={() => { showMathDialog = false; mathTargetEl = null; }}
					class="rounded-md border border-slate-300 px-4 py-2 text-sm text-slate-700 hover:bg-slate-50"
				>Batal</button>
			</div>
		</div>
	</div>
{/if}

<style>
	:global(.custom-editor-content) {
		outline: none;
		font-size: 0.875rem;
		line-height: 1.6;
	}
	:global(.custom-editor-content:empty::before) {
		content: attr(data-placeholder);
		color: #94a3b8;
		pointer-events: none;
	}

	:global(.custom-editor-content h1) { font-size: 1.25rem; font-weight: 700; margin: 0.75rem 0 0.5rem; }
	:global(.custom-editor-content h2) { font-size: 1.1rem; font-weight: 600; margin: 0.6rem 0 0.4rem; }
	:global(.custom-editor-content h3) { font-size: 1rem; font-weight: 600; margin: 0.5rem 0 0.3rem; }
	:global(.custom-editor-content p)  { margin: 0.25rem 0; }
	:global(.custom-editor-content ul) { list-style-type: disc; padding-left: 1.5rem; margin: 0.5rem 0; }
	:global(.custom-editor-content ol) { list-style-type: decimal; padding-left: 1.5rem; margin: 0.5rem 0; }
	:global(.custom-editor-content li) { line-height: 1.6; }

	/* Table */
	:global(.custom-editor-content table.ce-table) {
		border-collapse: collapse;
		width: 100%;
		margin: 0.75rem 0;
		font-size: 0.875rem;
	}
	:global(.custom-editor-content table.ce-table td,
	        .custom-editor-content table.ce-table th) {
		border: 1px solid #cbd5e1;
		padding: 6px 10px;
		min-width: 3rem;
		vertical-align: top;
	}
	:global(.custom-editor-content table.ce-table th) {
		background: #f8fafc;
		font-weight: 600;
	}
	:global(.custom-editor-content table.ce-table td:focus,
	        .custom-editor-content table.ce-table th:focus) {
		outline: 2px solid #059669;
		outline-offset: -2px;
	}

	/* Image */
	:global(.custom-editor-img) { max-width: 100%; border-radius: 4px; margin: 4px 0; display: block; }

	/* Math */
	:global(.ce-math-inline) {
		display: inline-block;
		vertical-align: middle;
		cursor: pointer;
		border-radius: 3px;
		padding: 0 2px;
		user-select: none;
	}
	:global(.ce-math-inline:hover) { background: #ecfdf5; outline: 1px solid #6ee7b7; }

	:global(.ce-math-block) {
		display: block;
		text-align: center;
		margin: 0.75rem 0;
		padding: 0.5rem;
		background: #f8fafc;
		border-radius: 4px;
		overflow-x: auto;
		cursor: pointer;
		user-select: none;
	}
	:global(.ce-math-block:hover) { background: #ecfdf5; outline: 1px solid #6ee7b7; }

	/* Toolbar */
	:global(.ce-btn) {
		padding: 3px 7px;
		border-radius: 4px;
		font-size: 0.8rem;
		color: #374151;
		line-height: 1.4;
		white-space: nowrap;
	}
	:global(.ce-btn:hover)    { background: #e2e8f0; }
	:global(.ce-btn.active)   { background: #d1fae5; color: #065f46; }
	:global(.ce-btn:disabled) { opacity: 0.5; cursor: not-allowed; }
	:global(.ce-div)          { width: 1px; height: 1rem; background: #cbd5e1; margin: 0 2px; flex-shrink: 0; }
</style>
