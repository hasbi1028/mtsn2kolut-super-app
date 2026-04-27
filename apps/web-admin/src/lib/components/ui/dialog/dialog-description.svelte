<script lang="ts">
  let { children } = $props();
</script>

<p class="desc">
  {@render children?.()}
</p>

<style>
  .desc {
    font-size: 0.85rem;
    color: #8aa0c0;
    margin: 0;
  }
</style>
