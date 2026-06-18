<script lang="ts">
	import { Toaster as Sonner, type ToasterProps } from 'svelte-sonner';

	let {
		theme = 'system',
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
				toast: 'border border-base-300 bg-base-100 text-base-content shadow-lg',
				title: 'text-base-content',
				description: 'text-base-content/70',
				actionButton: 'btn btn-primary btn-sm',
				cancelButton: 'btn btn-soft btn-sm'
			}
		}}
/>
