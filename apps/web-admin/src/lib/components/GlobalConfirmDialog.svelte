<script lang="ts">
	import { onMount } from 'svelte';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { setConfirmHandler, type ConfirmTone } from '$lib/confirm-dialog';

	type ConfirmRequest = {
		title: string;
		message: string;
		confirmLabel: string;
		cancelLabel: string;
		tone: ConfirmTone;
		challenge?: string;
	};

	type PendingConfirm = {
		options: ConfirmRequest;
		resolve: (confirmed: boolean) => void;
	};

	let open = $state(false);
	let current = $state<PendingConfirm | null>(null);
	let queue = $state<PendingConfirm[]>([]);
	let challengeInput = $state('');
	let wasOpen = false;

	const canConfirm = $derived(
		!current?.options.challenge || challengeInput.trim() === current.options.challenge
	);

	function toneClasses(tone: ConfirmTone) {
		if (tone === 'danger') return 'border-destructive/30 bg-destructive/10 text-destructive';
		if (tone === 'warning') return 'border-warning/30 bg-warning/10 text-warning';
		if (tone === 'success') return 'border-primary/20 bg-primary/10 text-primary';
		return 'border-border bg-muted/50 text-foreground';
	}

	function confirmButtonVariant(tone: ConfirmTone) {
		return tone === 'danger' ? 'destructive' : 'default';
	}

	function showNext() {
		if (current || queue.length === 0) return;
		const [next, ...rest] = queue;
		queue = rest;
		current = next;
		challengeInput = '';
		open = true;
	}

	function finish(confirmed: boolean) {
		const active = current;
		current = null;
		open = false;
		challengeInput = '';
		active?.resolve(confirmed);
		queueMicrotask(showNext);
	}

	onMount(() => {
		setConfirmHandler((options) => {
			return new Promise<boolean>((resolve) => {
				queue = [...queue, { options, resolve }];
				showNext();
			});
		});

		return () => {
			setConfirmHandler(null);
		};
	});

	$effect(() => {
		if (wasOpen && !open && current) {
			finish(false);
		}
		wasOpen = open;
	});
</script>

<div class="global-confirm-dialog-layer">
	<Dialog.Root bind:open>
		{#if current}
			<Dialog.Content>
				<Dialog.Header>
					<Dialog.Title>{current.options.title}</Dialog.Title>
					<Dialog.Description>
						Tinjau dampak tindakan sebelum melanjutkan.
					</Dialog.Description>
				</Dialog.Header>

				<div class={`rounded-2xl border px-4 py-3 ${toneClasses(current.options.tone)}`}>
					<p class="whitespace-pre-line text-sm leading-6">{current.options.message}</p>
				</div>

				{#if current.options.challenge}
					<div class="mt-4 space-y-2">
						<label for="confirm-challenge" class="block text-xs font-medium text-muted-foreground">
							Ketik <span class="font-semibold text-foreground">{current.options.challenge}</span> untuk mengaktifkan tombol konfirmasi.
						</label>
						<Input
							id="confirm-challenge"
							bind:value={challengeInput}
							placeholder={current.options.challenge}
							autocomplete="off"
						/>
					</div>
				{/if}

				<Dialog.Footer>
					<Button variant="outline" onclick={() => finish(false)}>{current.options.cancelLabel}</Button>
					<Button
						variant={confirmButtonVariant(current.options.tone)}
						disabled={!canConfirm}
						onclick={() => finish(true)}
					>
						{current.options.confirmLabel}
					</Button>
				</Dialog.Footer>
			</Dialog.Content>
		{/if}
	</Dialog.Root>
</div>

<style>
	:global(.global-confirm-dialog-layer .overlay) {
		z-index: 90;
	}
</style>
