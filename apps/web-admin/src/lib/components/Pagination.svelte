<script lang="ts">
  let {
    page = $bindable(1),
    total = 0,
    perPage = $bindable(12),
    onpagechange,
  }: {
    page?: number;
    total: number;
    perPage?: number;
    onpagechange?: (p: number) => void;
  } = $props();

  const PER_PAGE_OPTIONS = [10, 15, 25, 50, 0] as const;
  type PerPageOption = (typeof PER_PAGE_OPTIONS)[number];

  const effectivePerPage = $derived(perPage === 0 ? total : perPage);
  const totalPages = $derived(Math.max(1, Math.ceil(total / effectivePerPage)));

  function go(p: number) {
    if (p < 1 || p > totalPages || p === page) return;
    page = p;
    onpagechange?.(p);
  }

  function handlePerPageChange(e: Event) {
    const val = parseInt((e.target as HTMLSelectElement).value, 10);
    perPage = val as PerPageOption;
    page = 1;
  }

  const start = $derived(total === 0 ? 0 : (page - 1) * effectivePerPage + 1);
  const end = $derived(Math.min(page * effectivePerPage, total));

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
    <!-- Left: Rows per page selector + info -->
    <div class="flex items-center gap-3">
      <div class="flex items-center gap-1.5">
        <span class="text-[10px] text-muted-foreground font-medium whitespace-nowrap">Per halaman</span>
        <select
          onchange={handlePerPageChange}
          class="select select-bordered h-7 min-w-[68px] border border-input bg-background px-1.5 text-[11px] font-medium text-foreground"
          value={perPage}
        >
          {#each PER_PAGE_OPTIONS as opt}
            <option value={opt}>{opt === 0 ? 'Semua' : opt}</option>
          {/each}
        </select>
      </div>
      <span class="text-[11px] text-muted-foreground whitespace-nowrap">
        {start}&ndash;{end} dari {total}
      </span>
    </div>

    <!-- Right: Pagination buttons -->
    <div class="flex items-center gap-1">
      <!-- Prev -->
      <button
        onclick={() => go(page - 1)}
        disabled={page <= 1}
        class="inline-flex items-center justify-center rounded-lg border border-input bg-background px-2 h-7 text-xs font-medium text-foreground hover:bg-accent transition-colors disabled:opacity-30 disabled:pointer-events-none"
        aria-label="Halaman sebelumnya"
      >
        <svg class="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
        </svg>
      </button>

      <!-- Pages -->
      {#each pages as p (p === '...' ? `ellipsis-${pages.indexOf(p)}` : p)}
        {#if p === '...'}
          <span class="inline-flex items-center justify-center w-6 h-7 text-[10px] text-muted-foreground">&hellip;</span>
        {:else}
          <button
            onclick={() => go(p)}
            class="inline-flex items-center justify-center rounded-lg min-w-[28px] h-7 px-1.5 text-[11px] font-medium transition-colors {p === page
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
        class="inline-flex items-center justify-center rounded-lg border border-input bg-background px-2 h-7 text-xs font-medium text-foreground hover:bg-accent transition-colors disabled:opacity-30 disabled:pointer-events-none"
        aria-label="Halaman berikutnya"
      >
        <svg class="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
        </svg>
      </button>
    </div>
  </div>
{/if}
