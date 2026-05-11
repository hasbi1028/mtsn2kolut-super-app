<script lang="ts">
	import { Check, ChevronDown, Loader2, Pencil, X } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button';
	import type { EditableCellCancel, EditableCellCommit, EditableOption } from './types';

	let {
		value = $bindable(''),
		options = [],
		placeholder = '-',
		emptyLabel = 'Tidak dipilih',
		ariaLabel = 'Pilih opsi',
		disabled = false,
		loading = false,
		dirty = false,
		invalid = false,
		invalidMessage = '',
		allowEmpty = true,
		oncommit,
		oncancel,
		oninput
	}: {
		value?: string;
		options?: EditableOption[];
		placeholder?: string;
		emptyLabel?: string;
		ariaLabel?: string;
		disabled?: boolean;
		loading?: boolean;
		dirty?: boolean;
		invalid?: boolean;
		invalidMessage?: string;
		allowEmpty?: boolean;
		oncommit?: (detail: EditableCellCommit<string>) => void;
		oncancel?: (detail: EditableCellCancel<string>) => void;
		oninput?: (value: string) => void;
	} = $props();

	let editing = $state(false);
	let draft = $state(value);
	let original = $state(value);

	const selectedOption = $derived(options.find((option) => option.value === value));
	const shownValue = $derived(selectedOption?.label ?? (value ? value : placeholder));

	$effect(() => {
		if (!editing) {
			draft = value;
			original = value;
		}
	});

	function startEdit() {
		if (disabled || loading) return;
		original = value;
		draft = value;
		editing = true;
	}

	function commit() {
		if (invalid || loading) return;
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

	function handleChange(nextValue: string) {
		draft = nextValue;
		oninput?.(nextValue);
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			event.preventDefault();
			cancel();
		}
	}
</script>

<div class="min-w-40 space-y-1">
	{#if editing}
		<div class="flex items-center gap-1">
			<select
				class={`h-8 min-w-0 flex-1 rounded-md border border-input bg-background px-2 text-sm text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring ${
					invalid ? 'border-destructive focus-visible:ring-destructive/30' : ''
				}`}
				value={draft}
				aria-label={ariaLabel}
				aria-invalid={invalid}
				disabled={loading}
				onchange={(event) => handleChange(event.currentTarget.value)}
				onkeydown={handleKeydown}
			>
				{#if allowEmpty}
					<option value="">{emptyLabel}</option>
				{/if}
				{#each options as option (option.value)}
					<option value={option.value} disabled={option.disabled}>{option.label}</option>
				{/each}
			</select>
			<Button type="button" size="icon" class="size-8 shrink-0" disabled={invalid || loading} aria-label="Simpan pilihan" onclick={commit}>
				{#if loading}
					<Loader2 class="size-4 animate-spin" />
				{:else}
					<Check class="size-4" />
				{/if}
			</Button>
			<Button type="button" variant="ghost" size="icon" class="size-8 shrink-0" aria-label="Batalkan pilihan" onclick={cancel}>
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
				disabled || loading
					? 'cursor-default border-transparent text-muted-foreground'
					: 'border-transparent hover:border-border hover:bg-muted/60 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring'
			} ${dirty ? 'bg-primary/10 text-foreground' : 'text-foreground'}`}
			disabled={disabled || loading}
			aria-label={ariaLabel}
			onclick={startEdit}
		>
			<span class={shownValue === placeholder ? 'truncate text-muted-foreground' : 'truncate'}>{shownValue}</span>
			<span class="flex shrink-0 items-center gap-1">
				{#if loading}
					<Loader2 class="size-3.5 animate-spin text-muted-foreground" />
				{:else if dirty}
					<span class="size-2 rounded-full bg-primary" aria-label="Berubah"></span>
				{/if}
				{#if !disabled && !loading}
					<Pencil class="size-3.5 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100 group-focus-visible:opacity-100" />
					<ChevronDown class="size-3.5 text-muted-foreground opacity-60" />
				{/if}
			</span>
		</button>
	{/if}
</div>

