<script lang="ts">
  let { children } = $props();
</script>

<h3 class="title">
  {@render children?.()}
</h3>

<style>
  .title {
    font-size: 1.05rem;
    font-weight: 600;
    color: oklch(0.18 0.02 145);
    margin: 0 0 4px;
  }
</style>
