<script lang="ts">
	import TiptapEditor from '$lib/components/TiptapEditor.svelte';

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
</script>

<TiptapEditor {id} {placeholder} {minRows} {onImageUpload} bind:value />
