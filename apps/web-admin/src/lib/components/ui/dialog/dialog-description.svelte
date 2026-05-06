<script lang="ts">
  let { children } = $props();
</script>

<p class="desc">
  {@render children?.()}
</p>

<style>
  .desc {
    font-size: 0.85rem;
    color: var(--muted-foreground);
    margin: 0;
  }
</style>
