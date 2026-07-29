<script lang="ts">
  import type { ActionData } from './$types';

  let { form }: { form: ActionData } = $props();
</script>

<svelte:head>
  <title>Masuk — MTsN 2 Kolut</title>
  <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1" />
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body { font-family: sans-serif; background: #f5f5f5; }
    .mobile-login-root { min-height: 100dvh; display: flex; align-items: center; justify-content: center; padding: 20px; }
    .mobile-login-card { width: 100%; max-width: 380px; background: white; border-radius: 20px; padding: 32px 24px; box-shadow: 0 4px 24px rgba(0,0,0,0.08); }
    .mobile-login-card h1 { font-size: 20px; font-weight: 800; text-align: center; margin-bottom: 4px; }
    .mobile-login-card .subtitle { font-size: 13px; color: #666; text-align: center; margin-bottom: 24px; }
    .mobile-login-card .logo { text-align: center; margin-bottom: 16px; }
    .mobile-login-card .logo img { width: 56px; height: 56px; border-radius: 50%; background: #166534; padding: 8px; }
    .error-msg { background: #fef2f2; border: 1px solid #fecaca; color: #dc2626; padding: 10px 14px; border-radius: 12px; font-size: 13px; margin-bottom: 16px; text-align: center; }
    .field { margin-bottom: 16px; }
    .field label { display: block; font-size: 12px; font-weight: 700; color: #333; margin-bottom: 6px; text-transform: uppercase; letter-spacing: 0.5px; }
    .field input { width: 100%; height: 48px; padding: 0 16px; border: 1.5px solid #e0e0e0; border-radius: 14px; font-size: 16px; background: #fafafa; outline: none; transition: border-color 0.2s; }
    .field input:focus { border-color: #16a34a; background: white; }
    .submit-btn { width: 100%; height: 48px; background: #16a34a; color: white; border: none; border-radius: 14px; font-size: 15px; font-weight: 800; text-transform: uppercase; cursor: pointer; transition: background 0.2s; }
    .submit-btn:hover { background: #15803d; }
    .submit-btn:disabled { opacity: 0.5; cursor: not-allowed; }
    .attribution { text-align: center; margin-top: 20px; font-size: 11px; color: #999; }
    .attribution a { color: #16a34a; text-decoration: none; }
    .links { text-align: center; margin-top: 16px; }
    .links a { font-size: 13px; color: #16a34a; text-decoration: none; }
  </style>
</svelte:head>

<div class="mobile-login-root">
  <div class="mobile-login-card">
    <div class="logo">
      <img src="/brand/logo-kemenag-icon-64.png" alt="Logo Kemenag" />
    </div>
    <h1>MTsN 2 Kolut</h1>
    <p class="subtitle">Sistem Administrasi Madrasah</p>

    {#if form?.error}
      <div class="error-msg">{form.error}</div>
    {/if}

    <form method="POST">
      <div class="field">
        <label for="username">Nama Pengguna</label>
        <input id="username" name="username" type="text" placeholder="Username atau NIP" required autocomplete="username" />
      </div>
      <div class="field">
        <label for="password">Kata Sandi</label>
        <input id="password" name="password" type="password" placeholder="Masukkan password" required autocomplete="current-password" />
      </div>
      <button type="submit" class="submit-btn">Masuk</button>
    </form>

    <div class="links">
      <a href="/login">← Versi Desktop</a>
    </div>

    <div class="attribution">
      <p>Super App MTsN 2 Kolaka Utara · dibuat oleh Hasbi Awal, S. Kom.</p>
    </div>
  </div>
</div>
