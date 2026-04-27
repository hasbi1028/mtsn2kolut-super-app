<script lang="ts">
  let { children } = $props();
</script>

<p class="desc">
  {@render children?.()}
</p>

<style>
  .desc {
    font-size: 0.85rem;
    color: oklch(0.5 0.02 145);
    margin: 0;
  }
</style>
