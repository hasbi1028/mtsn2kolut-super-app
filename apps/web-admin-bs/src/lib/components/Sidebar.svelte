<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';

  let { activeRoute = '' } = $props();

  type NavItem = { label: string; icon: string; route: string };
  type NavSection = { title: string; items: NavItem[] };

  function closeOffcanvas() {
    if (!browser) return;
    const el = document.getElementById('sidebarOffcanvas');
    if (!el) return;
    // Use Bootstrap 5 Offcanvas API if available, otherwise manually remove show class
    const w = window as any;
    if (w.bootstrap?.Offcanvas) {
      w.bootstrap.Offcanvas.getInstance(el)?.hide();
    } else {
      el.classList.remove('show');
      el.setAttribute('aria-hidden', 'true');
      el.style.visibility = 'hidden';
      const backdrop = document.querySelector('.offcanvas-backdrop');
      if (backdrop) backdrop.remove();
      document.body.classList.remove('offcanvas-open');
      document.body.style.overflow = '';
    }
  }

  const sections: NavSection[] = [
    {
      title: '',
      items: [{ label: 'Dashboard', icon: 'bi-speedometer2', route: '/' }]
    },
    {
      title: 'KEHADIRAN',
      items: [
        { label: 'Monitor Kehadiran', icon: 'bi-clock', route: '/pusaka' },
        { label: 'Pegawai PUSAKA', icon: 'bi-people', route: '/pusaka/employees' },
        { label: 'Data Kehadiran', icon: 'bi-calendar-month', route: '/pusaka/kehadiran' },
        { label: 'Ringkasan Kehadiran', icon: 'bi-file-text', route: '/pusaka/summary' },
        { label: 'Laporan Telegram', icon: 'bi-send', route: '/pusaka/telegram-laporan' },
        { label: 'Antrian Sinkronisasi', icon: 'bi-arrow-repeat', route: '/pusaka/antrian' }
      ]
    },
    {
      title: 'PEGAWAI',
      items: [
        { label: 'Daftar Pegawai', icon: 'bi-person-badge', route: '/employees' }
      ]
    },
    {
      title: 'PENGATURAN',
      items: [
        { label: 'Akun Saya', icon: 'bi-person', route: '/settings/account' },
        { label: 'Profil Madrasah', icon: 'bi-building', route: '/settings/school-profile' },
        { label: 'Pengguna', icon: 'bi-shield', route: '/settings/users' },
        { label: 'Hak Akses', icon: 'bi-lock', route: '/settings/rbac' },
        { label: 'Logo & Branding', icon: 'bi-image', route: '/settings/branding' },
        { label: 'Backup Data', icon: 'bi-cloud-arrow-up', route: '/settings/backups' },
        { label: 'Audit Aktivitas', icon: 'bi-file-earmark-text', route: '/settings/audit-logs' },
        { label: 'Statistik', icon: 'bi-bar-chart', route: '/settings/analytics' }
      ]
    }
  ];
</script>

<!-- Desktop Sidebar -->
<aside class="sidebar bg-white border-end d-none d-lg-flex flex-column">
  <!-- Brand -->
  <div class="text-center py-3 border-bottom">
    <div class="bg-success text-white rounded-3 d-inline-flex align-items-center justify-content-center" style="width:44px;height:44px;">
      <span class="fw-black fs-5">M</span>
    </div>
    <div class="mt-1">
      <div class="fw-bold small">MTsN 2 Kolut</div>
      <div class="text-secondary" style="font-size:0.65rem;letter-spacing:0.15em;font-weight:700;">SUPER APP</div>
    </div>
  </div>

  <!-- Navigation -->
  <nav class="flex-grow-1 overflow-auto px-2 py-2">
    {#each sections as section}
      {#if section.title}
        <div class="nav-section">{section.title}</div>
      {/if}
      {#each section.items as item}
        <a
          href={item.route}
          class="nav-link d-flex align-items-center gap-2"
          class:active={activeRoute === item.route}
        >
          <i class="bi {item.icon}"></i>
          <span>{item.label}</span>
        </a>
      {/each}
    {/each}
  </nav>

  <!-- User footer -->
  <div class="border-top p-3">
    <div class="d-flex align-items-center gap-2 mb-2">
      <i class="bi bi-person-circle text-success"></i>
      <div>
        <div class="text-secondary" style="font-size:0.625rem;font-weight:700;">Masuk sebagai</div>
        <div class="fw-bold" style="font-size:0.8rem;">Admin</div>
      </div>
    </div>
    <a href="/logout" class="btn btn-outline-danger btn-sm w-100">
      <i class="bi bi-box-arrow-right me-1"></i>Keluar
    </a>
  </div>
</aside>

<!-- Mobile Offcanvas Sidebar -->
<div class="offcanvas offcanvas-start d-lg-none mobile-sidebar" tabindex="-1" id="sidebarOffcanvas" aria-labelledby="mobileSidebarTitle">
  <div class="offcanvas-header border-bottom">
    <div class="d-flex align-items-center gap-2">
      <div class="bg-success text-white rounded-3 d-inline-flex align-items-center justify-content-center" style="width:36px;height:36px;">
        <span class="fw-black">M</span>
      </div>
      <div>
        <div id="mobileSidebarTitle" class="fw-bold" style="font-size:0.85rem;">MTsN 2 Kolut</div>
        <div class="text-secondary" style="font-size:0.6rem;letter-spacing:0.15em;font-weight:700;">SUPER APP</div>
      </div>
    </div>
    <button type="button" class="btn-close" data-bs-dismiss="offcanvas" aria-label="Tutup menu navigasi"></button>
  </div>

  <div class="offcanvas-body p-2">
    {#each sections as section}
      {#if section.title}
        <div class="nav-section">{section.title}</div>
      {/if}
      {#each section.items as item}
        <a
          href={item.route}
          class="nav-link d-flex align-items-center gap-2"
          class:active={activeRoute === item.route}
          onclick={closeOffcanvas}
        >
          <i class="bi {item.icon}"></i>
          <span>{item.label}</span>
        </a>
      {/each}
    {/each}
  </div>

  <div class="offcanvas-footer border-top p-3 bg-white">
    <a href="/logout" class="btn btn-outline-danger btn-sm w-100">
      <i class="bi bi-box-arrow-right me-1"></i>Keluar
    </a>
  </div>
</div>
