<script lang="ts">
  let { children } = $props();
</script>

<div class="footer">
  {@render children?.()}
</div>

<style>
  .footer {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    margin-top: 20px;
    padding-top: 16px;
    border-top: 1px solid var(--border);
  }
</style>
