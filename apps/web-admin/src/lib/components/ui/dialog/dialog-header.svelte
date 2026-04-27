<script lang="ts">
  let { children } = $props();
</script>

<div class="header">
  {@render children?.()}
</div>

<style>
  .header {
    margin-bottom: 16px;
  }
</style>
