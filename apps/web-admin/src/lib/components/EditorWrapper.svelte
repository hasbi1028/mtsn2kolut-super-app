<script lang="ts">
	import { onMount } from 'svelte';
	import TiptapEditor from '$lib/components/TiptapEditor.svelte';
	import CustomEditor from '$lib/components/CustomEditor.svelte';

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

	const STORAGE_KEY = 'mtsn2-editor-mode';
	type EditorMode = 'tiptap' | 'custom';

	let mode = $state<EditorMode>('tiptap');

	onMount(() => {
		const saved = localStorage.getItem(STORAGE_KEY);
		if (saved === 'custom' || saved === 'tiptap') mode = saved;
	});

	function toggle(next: EditorMode) {
		mode = next;
		localStorage.setItem(STORAGE_KEY, next);
	}
</script>

<div class="space-y-1">
	<!-- Mode selector — pinned to top-right above the editor -->
	<div class="flex items-center justify-end gap-1">
		<span class="text-xs text-slate-400">Editor:</span>
		<div class="flex overflow-hidden rounded-md border border-slate-200 text-xs">
			<button
				type="button"
				onclick={() => toggle('tiptap')}
				class="px-2.5 py-1 font-medium transition {mode === 'tiptap' ? 'bg-emerald-700 text-white' : 'bg-white text-slate-600 hover:bg-slate-50'}"
			>TipTap</button>
			<button
				type="button"
				onclick={() => toggle('custom')}
				class="px-2.5 py-1 font-medium transition {mode === 'custom' ? 'bg-emerald-700 text-white' : 'bg-white text-slate-600 hover:bg-slate-50'}"
			>Editor Sendiri</button>
		</div>
	</div>

	{#if mode === 'tiptap'}
		<TiptapEditor {id} {placeholder} {minRows} {onImageUpload} bind:value />
	{:else}
		<CustomEditor {id} {placeholder} {minRows} {onImageUpload} bind:value />
	{/if}
</div>
