<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog';
	import { Input } from '$lib/components/ui/input';
	import { onMount } from 'svelte';

	type CommandItem = {
		href: string;
		label: string;
		group: string;
		pinned: boolean;
	};

	let {
		open = $bindable(false),
		items,
		recentHrefs,
		runCommand,
		clearRecent,
		isActive
	}: {
		open?: boolean;
		items: CommandItem[];
		recentHrefs: string[];
		runCommand: (href: string) => void | Promise<void>;
		clearRecent: () => void;
		isActive: (href: string) => boolean;
	} = $props();

	let query = $state('');
	let commandInputRef = $state<HTMLInputElement | null>(null);

	const filteredItems = $derived.by(() => {
		const normalizedQuery = query.trim().toLowerCase();
		const base = items.filter((item) => {
			if (!normalizedQuery) return true;
			return (
				item.label.toLowerCase().includes(normalizedQuery) ||
				item.group.toLowerCase().includes(normalizedQuery) ||
				item.href.toLowerCase().includes(normalizedQuery)
			);
		});
		return [...base].sort((left, right) => {
			const leftScore = Number(left.pinned) + Number(isActive(left.href)) * 3;
			const rightScore = Number(right.pinned) + Number(isActive(right.href)) * 3;
			if (leftScore !== rightScore) return rightScore - leftScore;
			return left.label.localeCompare(right.label, 'id');
		});
	});

	const recentItems = $derived.by(() => {
		return recentHrefs
			.map((href) => items.find((item) => item.href === href))
			.filter((item): item is CommandItem => !!item && !isActive(item.href));
	});

	const pinnedItems = $derived.by(() => filteredItems.filter((item) => item.pinned));

	const recentMatches = $derived.by(() => {
		const recentSet = new Set(recentItems.map((item) => item.href));
		return filteredItems.filter((item) => recentSet.has(item.href) && !item.pinned);
	});

	const allMenuItems = $derived.by(() => {
		const excluded = new Set<string>([
			...pinnedItems.map((item) => item.href),
			...recentMatches.map((item) => item.href)
		]);
		return filteredItems.filter((item) => !excluded.has(item.href));
	});

	function badgeToneClasses(kind: 'quick' | 'recent' | 'active') {
		if (kind === 'quick') {
			return 'bg-amber-100 text-amber-700';
		}
		if (kind === 'recent') {
			return 'bg-slate-100 text-slate-600';
		}
		return 'bg-green-100 text-green-700';
	}

	function cardClasses(href: string) {
		return isActive(href)
			? 'border-green-200 bg-green-50 text-green-900'
			: 'border-slate-200 bg-white hover:bg-slate-50';
	}

	function openPalette() {
		query = '';
		open = true;
	}

	onMount(() => {
		if (open) {
			openPalette();
		}
	});

	$effect(() => {
		if (!open || !commandInputRef) return;
		queueMicrotask(() => commandInputRef?.focus());
	});

	$effect(() => {
		if (!open) {
			query = '';
		}
	});
</script>

<Dialog.Root bind:open={open}>
	{#if open}
		<Dialog.Content>
			<Dialog.Header>
				<Dialog.Title>Cari Menu</Dialog.Title>
				<Dialog.Description>Ketik nama menu, grup, atau path. Pintasan: Ctrl/Cmd + K atau /</Dialog.Description>
			</Dialog.Header>

			<div class="space-y-3">
				<Input
					bind:ref={commandInputRef}
					bind:value={query}
					placeholder="Mis. Bank Soal, Hasil & Analisis, Inventaris, atau PUSAKA"
				/>

				{#if pinnedItems.length > 0}
					<div class="space-y-2">
						<p class="text-[11px] font-semibold uppercase tracking-[0.18em] text-slate-400">Akses Cepat</p>
						<div class="space-y-2">
							{#each pinnedItems as item (item.href)}
								<button
									type="button"
									class={`flex w-full items-start justify-between rounded-xl border px-3 py-3 text-left transition-colors ${cardClasses(item.href)}`}
									onclick={() => runCommand(item.href)}
								>
									<div class="min-w-0">
										<div class="flex items-center gap-2">
											<span class="text-sm font-semibold">{item.label}</span>
											<span class={`rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.14em] ${badgeToneClasses('quick')}`}>
												Cepat
											</span>
										</div>
										<p class="mt-1 text-xs text-slate-500">{item.group} · {item.href}</p>
									</div>
									<svg class="mt-0.5 h-4 w-4 shrink-0 text-slate-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
									</svg>
								</button>
							{/each}
						</div>
					</div>
				{/if}

				{#if recentItems.length > 0 && !query.trim()}
					<div class="space-y-2">
						<div class="flex items-center justify-between">
							<p class="text-[11px] font-semibold uppercase tracking-[0.18em] text-slate-400">Terakhir Dibuka</p>
							<button
								type="button"
								class="text-xs font-medium text-slate-500 hover:text-slate-700"
								onclick={clearRecent}
							>
								Bersihkan
							</button>
						</div>
						<div class="space-y-2">
							{#each recentItems as item (item.href)}
								<button
									type="button"
									class={`flex w-full items-start justify-between rounded-xl border px-3 py-3 text-left transition-colors ${cardClasses(item.href)}`}
									onclick={() => runCommand(item.href)}
								>
									<div class="min-w-0">
										<div class="flex items-center gap-2">
											<span class="text-sm font-semibold">{item.label}</span>
											<span class={`rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.14em] ${badgeToneClasses('recent')}`}>
												Baru
											</span>
										</div>
										<p class="mt-1 text-xs text-slate-500">{item.group} · {item.href}</p>
									</div>
									<svg class="mt-0.5 h-4 w-4 shrink-0 text-slate-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
									</svg>
								</button>
							{/each}
						</div>
					</div>
				{/if}

				{#if filteredItems.length === 0}
					<div class="rounded-xl border border-slate-200 bg-slate-50 px-4 py-6 text-center text-sm text-slate-500">
						Tidak ada menu yang cocok dengan pencarian.
					</div>
				{:else if allMenuItems.length > 0}
					<div class="space-y-2">
						<p class="text-[11px] font-semibold uppercase tracking-[0.18em] text-slate-400">Semua Menu</p>
						<div class="max-h-[420px] space-y-2 overflow-y-auto pr-1">
							{#each allMenuItems as item (item.href)}
								<button
									type="button"
									class={`flex w-full items-start justify-between rounded-xl border px-3 py-3 text-left transition-colors ${cardClasses(item.href)}`}
									onclick={() => runCommand(item.href)}
								>
									<div class="min-w-0">
										<div class="flex items-center gap-2">
											<span class="text-sm font-semibold">{item.label}</span>
											{#if isActive(item.href)}
												<span class={`rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.14em] ${badgeToneClasses('active')}`}>
													Aktif
												</span>
											{/if}
										</div>
										<p class="mt-1 text-xs text-slate-500">{item.group} · {item.href}</p>
									</div>
									<svg class="mt-0.5 h-4 w-4 shrink-0 text-slate-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
									</svg>
								</button>
							{/each}
						</div>
					</div>
				{:else}
					<div class="max-h-[420px] space-y-2 overflow-y-auto pr-1">
						{#each filteredItems as item (item.href)}
							<button
								type="button"
								class={`flex w-full items-start justify-between rounded-xl border px-3 py-3 text-left transition-colors ${cardClasses(item.href)}`}
								onclick={() => runCommand(item.href)}
							>
								<div class="min-w-0">
									<div class="flex items-center gap-2">
										<span class="text-sm font-semibold">{item.label}</span>
										{#if item.pinned}
											<span class={`rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.14em] ${badgeToneClasses('quick')}`}>
												Cepat
											</span>
										{/if}
										{#if isActive(item.href)}
											<span class={`rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.14em] ${badgeToneClasses('active')}`}>
												Aktif
											</span>
										{/if}
									</div>
									<p class="mt-1 text-xs text-slate-500">{item.group} · {item.href}</p>
								</div>
								<svg class="mt-0.5 h-4 w-4 shrink-0 text-slate-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
								</svg>
							</button>
						{/each}
					</div>
				{/if}
			</div>
		</Dialog.Content>
	{/if}
</Dialog.Root>
