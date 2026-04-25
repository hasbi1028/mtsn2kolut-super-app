<script lang="ts">
  import type { ActionData } from './$types';
  let { form }: { form: ActionData } = $props();
</script>

<svelte:head><title>Login — Pusaka Worker</title></svelte:head>

<div class="wrap">
  <div class="box">
    <h1>Pusaka Worker</h1>
    <form method="POST">
      {#if form?.error}
        <div class="error">{form.error}</div>
      {/if}
      <label>
        Username
        <input name="username" type="text" autocomplete="username" required />
      </label>
      <label>
        Password
        <input name="password" type="password" autocomplete="current-password" required />
      </label>
      <button type="submit">Masuk</button>
    </form>
  </div>
</div>

<style>
  .wrap {
    position: fixed; inset: 0;
    display: flex; align-items: center; justify-content: center;
    background: #070b13;
    z-index: 999;
  }
  .box {
    width: 100%; max-width: 360px;
    background: rgba(22, 30, 46, 0.95);
    border: 1px solid rgba(130, 157, 204, 0.18);
    border-radius: 14px;
    padding: 36px 32px;
    margin: 16px;
  }
  h1 {
    color: #58a6ff; font-size: 1.25rem;
    margin: 0 0 24px; text-align: center; font-weight: 700;
  }
  form { display: flex; flex-direction: column; gap: 14px; }
  label { display: flex; flex-direction: column; gap: 5px; font-size: 0.85rem; color: #9db2d1; }
  input {
    background: rgba(7, 11, 19, 0.6);
    border: 1px solid rgba(130, 157, 204, 0.2);
    border-radius: 8px; color: #e7edf7;
    padding: 9px 12px; font-size: 0.95rem; outline: none;
  }
  input:focus { border-color: #58a6ff; }
  button {
    background: #1f6feb; color: #fff; border: none;
    border-radius: 8px; padding: 10px;
    font-size: 0.95rem; font-weight: 600; cursor: pointer; margin-top: 4px;
  }
  button:hover { background: #388bfd; }
  .error {
    background: rgba(248, 81, 73, 0.15);
    border: 1px solid rgba(248, 81, 73, 0.4);
    color: #ff7b72; border-radius: 8px;
    padding: 8px 12px; font-size: 0.85rem;
  }
</style>
