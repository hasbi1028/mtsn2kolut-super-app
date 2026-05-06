<script lang="ts">
  let { children } = $props();
</script>

<div class="content">
  {@render children?.()}
</div>

<style>
  .content {
    background: var(--popover);
    border: 1px solid var(--border);
    border-radius: 0.75rem;
    padding: 24px;
    min-width: min(90vw, 500px);
    max-width: 96vw;
    width: fit-content;
    max-height: 85vh;
    overflow-y: auto;
    color: var(--popover-foreground);
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
  }

  :global(.dark) .content {
    box-shadow: 0 20px 70px rgba(0, 0, 0, 0.45);
  }
</style>
