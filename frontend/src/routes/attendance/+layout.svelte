<script>
  let { children } = $props();
</script>

{@render children()}

<style>
  /* Override .page padding/gap khusus halaman attendance */
  :global(.page) {
    padding-top: 10px !important;
    padding-bottom: 10px !important;
    gap: 6px !important;
  }
</style>
