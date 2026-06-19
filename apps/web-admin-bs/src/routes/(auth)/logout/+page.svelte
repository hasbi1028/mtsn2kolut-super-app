<script lang="ts">
  import { onMount } from 'svelte';

  onMount(async () => {
    try {
      await fetch('/api/auth/logout', { method: 'POST' });
    } catch (_) {}
    document.cookie = 'access_token=; path=/; max-age=0';
    document.cookie = 'refresh_token=; path=/; max-age=0';
    window.location.href = '/login';
  });
</script>

<div class="container vh-100 d-flex align-items-center justify-content-center">
  <div class="text-center">
    <div class="spinner-border text-success mb-3" role="status"></div>
    <p class="text-secondary">Sedang keluar...</p>
  </div>
</div>
