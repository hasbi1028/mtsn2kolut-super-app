<script lang="ts">
	import { Toaster as Sonner, type ToasterProps } from 'svelte-sonner';

	let {
		theme = 'light',
		position = 'top-right',
		richColors = true,
		closeButton = true
	}: ToasterProps = $props();
</script>

<Sonner
	{theme}
	{position}
	{richColors}
	{closeButton}
	toastOptions={{
		classes: {
			toast: 'border border-emerald-200 bg-white text-slate-900 shadow-lg',
			title: 'text-slate-900',
			description: 'text-slate-600',
			actionButton: 'bg-emerald-700 text-white',
			cancelButton: 'bg-slate-100 text-slate-700'
		}
	}}
/>
