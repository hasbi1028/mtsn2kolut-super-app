<script lang="ts">
  let username = $state('');
  let password = $state('');
  let loading = $state(false);
  let error = $state('');
  let showPassword = $state(false);

  async function handleSubmit(e: Event) {
    e.preventDefault();
    if (!username || !password) {
      error = 'Username dan password harus diisi';
      return;
    }
    loading = true;
    error = '';

    try {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
      });

      let data: Record<string, unknown> = {};
      const text = await res.text();
      try { data = JSON.parse(text); } catch { /* non-JSON response */ }

      if (!res.ok) {
        if (res.status === 401) {
          error = 'Username atau password salah';
        } else if (res.status === 403) {
          error = (data.error as string) || 'Akun telah ditangguhkan';
        } else if (res.status === 503) {
          error = 'Layanan backend tidak tersedia. Silakan coba beberapa saat lagi.';
        } else {
          error = (data.error as string) || (data.message as string) || `Kesalahan server (${res.status})`;
        }
        loading = false;
        return;
      }

      window.location.href = '/';
    } catch (e: unknown) {
      const msg = e instanceof TypeError ? e.message : String(e);
      if (msg.includes('Failed to fetch') || msg.includes('NetworkError') || msg.includes('ECONNREFUSED')) {
        error = 'Tidak dapat terhubung ke server. Pastikan server backend berjalan.';
      } else {
        error = 'Terjadi kesalahan. Silakan coba lagi.';
      }
      loading = false;
    }
  }
</script>

<div class="login-shell container min-vh-100 d-flex align-items-center justify-content-center py-4">
  <div class="login-card bg-white p-4 p-md-5">
    <div class="text-center mb-4">
      <div class="bg-success text-white rounded-3 d-inline-flex align-items-center justify-content-center mb-3 shadow-sm" style="width:64px;height:64px;">
        <span class="fw-black fs-3">M</span>
      </div>
      <h5 class="fw-black mb-1">MTsN 2 Kolut</h5>
      <p class="text-secondary" style="font-size:0.75rem;font-weight:600;letter-spacing:0.1em;">SISTEM INFORMASI MADRASAH</p>
    </div>

    {#if error}
      <div class="alert alert-danger d-flex align-items-center py-2" role="alert">
        <i class="bi bi-exclamation-circle me-2"></i>
        <small>{error}</small>
      </div>
    {/if}

    <form onsubmit={handleSubmit}>
      <div class="mb-3">
        <label for="username" class="form-label small fw-semibold">Username</label>
        <div class="input-group">
          <span class="input-group-text"><i class="bi bi-person"></i></span>
          <input
            type="text"
            class="form-control"
            id="username"
            bind:value={username}
            placeholder="Masukkan username"
            autocomplete="username"
          />
        </div>
      </div>

      <div class="mb-4">
        <label for="password" class="form-label small fw-semibold">Password</label>
        <div class="input-group">
          <span class="input-group-text"><i class="bi bi-lock"></i></span>
          <input
            type={showPassword ? 'text' : 'password'}
            class="form-control"
            id="password"
            bind:value={password}
            placeholder="Masukkan password"
            autocomplete="current-password"
          />
          <button class="btn btn-outline-secondary" type="button" onclick={() => showPassword = !showPassword} aria-label="Toggle password visibility">
            <i class="bi {showPassword ? 'bi-eye-slash' : 'bi-eye'}"></i>
          </button>
        </div>
      </div>

      <button type="submit" class="btn btn-primary w-100 py-2 fw-bold login-submit-btn" disabled={loading}>
        {#if loading}
          <span class="spinner-border spinner-border-sm me-2" role="status"></span>
          Memproses...
        {:else}
          <i class="bi bi-box-arrow-in-right me-2"></i>Masuk
        {/if}
      </button>
    </form>

    <p class="text-center text-secondary mt-4" style="font-size:0.7rem;">
      &copy; {new Date().getFullYear()} MTsN 2 Kolaka Utara
    </p>
  </div>
</div>
