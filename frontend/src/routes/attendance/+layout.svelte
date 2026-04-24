<script>
  let { children } = $props();
</script>

{@render children()}

<style>
  :global(.page) {
    max-width: 420px !important;
    margin: 0 auto !important;
    padding: 6px 6px 6px !important;
    gap: 4px !important;
  }
</style>
