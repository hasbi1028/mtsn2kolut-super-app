<script lang="ts">
  import { page } from '$app/stores';
</script>

<div class="container py-5">
  <div class="row justify-content-center">
    <div class="col-md-6 text-center">
      <div class="text-danger mb-4">
        <i class="bi bi-exclamation-triangle" style="font-size:4rem;"></i>
      </div>
      <h1 class="fw-black mb-3">Terjadi Kesalahan</h1>
      <p class="text-secondary mb-4">
        Sistem sedang mengalami kendala. Silakan coba beberapa saat lagi.
      </p>
      <div class="alert alert-warning d-inline-block text-start" role="alert">
        <i class="bi bi-info-circle me-2"></i>
        Kode: {$page.status}
      </div>
      <div class="mt-4">
        <a href="/" class="btn btn-primary">
          <i class="bi bi-house me-2"></i>Kembali ke Beranda
        </a>
      </div>
    </div>
  </div>
</div>
