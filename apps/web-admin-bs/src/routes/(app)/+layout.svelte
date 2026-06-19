<script lang="ts">
  import { page } from '$app/state';
  import Sidebar from '$lib/components/Sidebar.svelte';

  let { children } = $props();
  const activeRoute = $derived(page.url.pathname);

  const days = ['Minggu', 'Senin', 'Selasa', 'Rabu', 'Kamis', 'Jumat', 'Sabtu'];
  const months = ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'];
  const now = new Date();
  const dateStr = `${days[now.getDay()]}, ${now.getDate()} ${months[now.getMonth()]} ${now.getFullYear()}`;
</script>

<div class="app-shell d-flex">
  <Sidebar {activeRoute} />

  <div class="content-area d-flex flex-column">
    <nav class="page-header d-flex align-items-center px-3 px-md-4">
      <button class="btn btn-light border mobile-menu-btn d-lg-none me-2" type="button" data-bs-toggle="offcanvas" data-bs-target="#sidebarOffcanvas" aria-label="Buka menu navigasi">
        <i class="bi bi-list fs-4"></i>
      </button>

      <div class="min-w-0">
        <h5 class="mb-0 fw-bold app-title">Super App — MTsN 2 Kolut</h5>
        <div class="d-md-none text-secondary" style="font-size:0.7rem;">Sistem Manajemen Madrasah</div>
      </div>

      <div class="ms-auto d-flex align-items-center gap-2 gap-md-3">
        <span class="text-secondary small fw-semibold d-none d-sm-inline">{dateStr}</span>
        <div class="dropdown">
          <button class="btn btn-outline-light border-0 dropdown-toggle d-flex align-items-center gap-2 user-menu-btn" data-bs-toggle="dropdown" aria-label="User menu">
            <i class="bi bi-person-circle fs-5 text-primary"></i>
            <span class="small fw-semibold text-dark d-none d-sm-inline">Admin</span>
          </button>
          <ul class="dropdown-menu dropdown-menu-end shadow-sm">
            <li><a class="dropdown-item" href="/settings/account"><i class="bi bi-person me-2"></i>Akun Saya</a></li>
            <li><hr class="dropdown-divider"></li>
            <li><a class="dropdown-item text-danger" href="/logout"><i class="bi bi-box-arrow-right me-2"></i>Keluar</a></li>
          </ul>
        </div>
      </div>
    </nav>

    <main class="app-main flex-grow-1 p-3 p-md-4">
      {@render children()}
    </main>
  </div>
</div>
