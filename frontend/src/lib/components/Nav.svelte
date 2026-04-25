<script lang="ts">
  import { page } from '$app/state';

  let { user }: { user?: { id: string } } = $props();

  const links = [
    { href: '/',           label: 'Dashboard' },
    { href: '/employees',  label: 'Pegawai'   },
    { href: '/jobs',       label: 'Jobs'       },
    { href: '/attendance', label: 'Absensi'    },
    { href: '/settings',   label: 'Pengaturan' }
  ];

  function active(href: string) {
    if (href === '/') return page.url.pathname === '/';
    return page.url.pathname.startsWith(href);
  }

  async function logout() {
    await fetch('/api/auth/logout', { method: 'POST' });
    location.href = '/login';
  }
</script>

<nav class="topnav">
  <a class="brand" href="/">Pusaka Worker</a>
  <div class="nav-links">
    {#each links as l}
      <a href={l.href} class:active={active(l.href)}>{l.label}</a>
    {/each}
  </div>
  {#if user}
    <button class="logout-btn" onclick={logout}>Keluar</button>
  {/if}
</nav>

<style>
  .topnav {
    position: sticky; top: 0; z-index: 100;
    display: flex; align-items: center; gap: 8px;
    padding: 0 24px;
    background: rgba(7, 11, 19, 0.92);
    border-bottom: 1px solid rgba(130, 157, 204, 0.15);
    backdrop-filter: blur(8px);
    height: 52px;
  }

  .brand {
    font-weight: 700; font-size: 1rem;
    color: #58a6ff; text-decoration: none;
    margin-right: 8px; white-space: nowrap;
  }

  .nav-links {
    display: flex; gap: 2px; flex: 1;
  }

  .nav-links a {
    color: #9db2d1; text-decoration: none;
    padding: 6px 14px; border-radius: 8px;
    font-size: 0.9rem; font-weight: 500;
    transition: background 0.15s, color 0.15s;
  }

  .nav-links a:hover { background: rgba(88,166,255,0.1); color: #e7edf7; }
  .nav-links a.active { background: rgba(88,166,255,0.18); color: #58a6ff; font-weight: 600; }

  .logout-btn {
    background: none;
    border: 1px solid rgba(130, 157, 204, 0.25);
    color: #9db2d1; border-radius: 7px;
    padding: 5px 12px; font-size: 0.82rem;
    cursor: pointer; white-space: nowrap;
    transition: background 0.15s, color 0.15s;
    flex-shrink: 0;
  }
  .logout-btn:hover {
    background: rgba(248,81,73,0.12);
    color: #ff7b72;
    border-color: rgba(248,81,73,0.3);
  }

  @media (max-width: 600px) {
    .topnav { padding: 0 12px; }
    .brand { font-size: 0.9rem; }
    .nav-links a { padding: 6px 8px; font-size: 0.82rem; }
    .logout-btn { padding: 4px 8px; font-size: 0.78rem; }
  }
</style>
