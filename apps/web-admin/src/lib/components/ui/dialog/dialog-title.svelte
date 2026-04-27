<script lang="ts">
  let { children } = $props();
</script>

<h3 class="title">
  {@render children?.()}
</h3>

<style>
  .title {
    font-size: 1.1rem;
    font-weight: 600;
    color: #e2eaf6;
    margin: 0 0 4px;
  }
</style>
