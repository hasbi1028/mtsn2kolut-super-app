<script lang="ts">
  let { children, onclick } = $props();
</script>

<button type="button" {onclick}>
  {@render children?.()}
</button>
