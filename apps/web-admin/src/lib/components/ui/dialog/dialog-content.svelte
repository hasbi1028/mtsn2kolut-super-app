<script lang="ts">
  let { children } = $props();
</script>

<div class="content">
  {@render children?.()}
</div>

<style>
  .content {
    background: #1a2742;
    border: 1px solid #2f4668;
    border-radius: 10px;
    padding: 24px;
    max-width: 500px;
    width: 90%;
    max-height: 85vh;
    overflow-y: auto;
    color: #e2eaf6;
  }
</style>
