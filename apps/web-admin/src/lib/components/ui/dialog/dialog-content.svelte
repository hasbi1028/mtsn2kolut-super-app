<script lang="ts">
  let { children } = $props();
</script>

<div class="content">
  {@render children?.()}
</div>

<style>
  .content {
    background: #ffffff;
    border: 1px solid oklch(0.88 0.012 145);
    border-radius: 0.75rem;
    padding: 24px;
    max-width: 500px;
    width: 90%;
    max-height: 85vh;
    overflow-y: auto;
    color: oklch(0.18 0.02 145);
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
  }
</style>
