<script lang="ts">
  let {
    page = $bindable(1),
    total = 0,
    perPage = 12,
    onpagechange,
    onloadall,
  }: {
    page?: number;
    total: number;
    perPage?: number;
    onpagechange?: (p: number) => void;
    onloadall?: () => void;
  } = $props();

  const totalPages = $derived(Math.max(1, Math.ceil(total / perPage)));
  const isAllShown = $derived(perPage >= total);

  function go(p: number) {
    if (p < 1 || p > totalPages || p === page) return;
    page = p;
    onpagechange?.(p);
  }

  const start = $derived(total === 0 ? 0 : (page - 1) * perPage + 1);
  const end = $derived(Math.min(page * perPage, total));

  // Page numbers to show (with ellipsis)
  const pages = $derived.by(() => {
    const p: (number | '...')[] = [];
    if (totalPages <= 7) {
      for (let i = 1; i <= totalPages; i++) p.push(i);
      return p;
    }
    p.push(1);
    if (page > 3) p.push('...');
    const startP = Math.max(2, page - 1);
    const endP = Math.min(totalPages - 1, page + 1);
    for (let i = startP; i <= endP; i++) p.push(i);
    if (page < totalPages - 2) p.push('...');
    p.push(totalPages);
    return p;
  });
</script>

{#if total > 0}
  <div class="flex flex-col sm:flex-row items-center justify-between gap-2">
    <span class="text-xs text-muted-foreground">
      {#if isAllShown}
        Menampilkan semua {total} rekaman
      {:else}
        {perPage} dari {total} rekaman &middot; halaman {page} dari {totalPages}
      {/if}
    </span>

    <div class="flex items-center gap-1">
      {#if !isAllShown}
        <!-- Prev -->
        <button
          onclick={() => go(page - 1)}
          disabled={page <= 1}
          class="inline-flex items-center justify-center rounded-lg border border-input bg-background px-2.5 h-8 text-xs font-medium text-foreground hover:bg-accent transition-colors disabled:opacity-30 disabled:pointer-events-none"
          aria-label="Halaman sebelumnya"
        >
          <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
          </svg>
        </button>

        <!-- Pages -->
        {#each pages as p (p === '...' ? `ellipsis-${pages.indexOf(p)}` : p)}
          {#if p === '...'}
            <span class="inline-flex items-center justify-center w-7 h-8 text-xs text-muted-foreground">&hellip;</span>
          {:else}
            <button
              onclick={() => go(p)}
              class="inline-flex items-center justify-center rounded-lg min-w-[32px] h-8 px-2 text-xs font-medium transition-colors {p === page
                ? 'bg-primary text-primary-foreground shadow-sm'
                : 'border border-input bg-background text-foreground hover:bg-accent'}"
              aria-label={'Halaman ' + p}
              aria-current={p === page ? 'page' : undefined}
            >
              {p}
            </button>
          {/if}
        {/each}

        <!-- Next -->
        <button
          onclick={() => go(page + 1)}
          disabled={page >= totalPages}
          class="inline-flex items-center justify-center rounded-lg border border-input bg-background px-2.5 h-8 text-xs font-medium text-foreground hover:bg-accent transition-colors disabled:opacity-30 disabled:pointer-events-none"
          aria-label="Halaman berikutnya"
        >
          <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
          </svg>
        </button>

        <!-- Load all -->
        {#if onloadall}
          <span class="mx-1.5 text-[10px] text-border" aria-hidden="true">|</span>
          <button
            onclick={() => onloadall()}
            class="inline-flex items-center justify-center rounded-lg h-8 px-2.5 text-[10px] font-semibold text-primary hover:bg-primary/5 border border-dashed border-primary/30 transition-colors"
          >
            Muat semua {total}
          </button>
        {/if}
      {:else if onloadall}
        <!-- Back to paginated mode -->
        <button
          onclick={() => onloadall()}
          class="inline-flex items-center justify-center rounded-lg h-8 px-2.5 text-xs font-medium border border-input bg-background text-foreground hover:bg-accent transition-colors"
        >
          Kembali ke halaman
        </button>
      {/if}
    </div>
  </div>
{/if}
