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
			toast: 'border border-primary/20 bg-card text-foreground shadow-lg',
			title: 'text-foreground',
			description: 'text-muted-foreground',
			actionButton: 'bg-primary text-primary-foreground',
			cancelButton: 'bg-muted text-foreground'
		}
	}}
/>
