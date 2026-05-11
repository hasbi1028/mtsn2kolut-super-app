<script lang="ts">
	import { Check, Pencil, X } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import type { EditableCellCancel, EditableCellCommit } from './types';

	let {
		value = $bindable(''),
		displayValue,
		placeholder = '-',
		ariaLabel = 'Edit teks',
		disabled = false,
		dirty = false,
		invalid = false,
		invalidMessage = '',
		inputType = 'text',
		oncommit,
		oncancel,
		oninput
	}: {
		value?: string;
		displayValue?: string;
		placeholder?: string;
		ariaLabel?: string;
		disabled?: boolean;
		dirty?: boolean;
		invalid?: boolean;
		invalidMessage?: string;
		inputType?: 'text' | 'email' | 'tel' | 'url' | 'number';
		oncommit?: (detail: EditableCellCommit<string>) => void;
		oncancel?: (detail: EditableCellCancel<string>) => void;
		oninput?: (value: string) => void;
	} = $props();

	let editing = $state(false);
	let draft = $state(value);
	let original = $state(value);

	const shownValue = $derived((displayValue ?? value).trim() || placeholder);

	$effect(() => {
		if (!editing) {
			draft = value;
			original = value;
		}
	});

	function startEdit() {
		if (disabled) return;
		original = value;
		draft = value;
		editing = true;
	}

	function commit() {
		if (invalid) return;
		const previousValue = value;
		value = draft;
		oninput?.(draft);
		oncommit?.({ value: draft, previousValue });
		editing = false;
	}

	function cancel() {
		const current = draft;
		draft = original;
		oncancel?.({ value: current, originalValue: original });
		editing = false;
	}

	function handleDraftInput(nextValue: string) {
		draft = nextValue;
		oninput?.(nextValue);
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter') {
			event.preventDefault();
			commit();
		}
		if (event.key === 'Escape') {
			event.preventDefault();
			cancel();
		}
	}
</script>

<div class="min-w-36 space-y-1">
	{#if editing}
		<div class="flex items-center gap-1">
			<Input
				type={inputType}
				value={draft}
				aria-label={ariaLabel}
				class={`h-8 min-w-0 text-sm ${invalid ? 'border-destructive focus-visible:ring-destructive/30' : ''}`}
				aria-invalid={invalid}
				oninput={(event) => handleDraftInput(event.currentTarget.value)}
				onkeydown={handleKeydown}
			/>
			<Button type="button" size="icon" class="size-8 shrink-0" disabled={invalid} aria-label="Simpan perubahan" onclick={commit}>
				<Check class="size-4" />
			</Button>
			<Button type="button" variant="ghost" size="icon" class="size-8 shrink-0" aria-label="Batalkan perubahan" onclick={cancel}>
				<X class="size-4" />
			</Button>
		</div>
		{#if invalid && invalidMessage}
			<p class="text-xs text-destructive">{invalidMessage}</p>
		{/if}
	{:else}
		<button
			type="button"
			class={`group flex min-h-8 w-full items-center justify-between gap-2 rounded-md border px-2 py-1.5 text-left text-sm transition-colors ${
				disabled
					? 'cursor-default border-transparent text-muted-foreground'
					: 'border-transparent hover:border-border hover:bg-muted/60 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring'
			} ${dirty ? 'bg-primary/10 text-foreground' : 'text-foreground'}`}
			disabled={disabled}
			aria-label={ariaLabel}
			onclick={startEdit}
		>
			<span class={shownValue === placeholder ? 'truncate text-muted-foreground' : 'truncate'}>{shownValue}</span>
			<span class="flex shrink-0 items-center gap-1">
				{#if dirty}
					<span class="size-2 rounded-full bg-primary" aria-label="Berubah"></span>
				{/if}
				{#if !disabled}
					<Pencil class="size-3.5 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100 group-focus-visible:opacity-100" />
				{/if}
			</span>
		</button>
	{/if}
</div>

