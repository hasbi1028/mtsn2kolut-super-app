<script lang="ts">
  let { children } = $props();

  const sections = [
    {
      group: "Pusat Data",
      items: [
        { href: "/asesmen/cbt", label: "Dashboard Utama", icon: "📊" },
        { href: "/akademik/mapel", label: "Mata Pelajaran", icon: "📚" },
        { href: "/akademik/rombel", label: "Data Kelas", icon: "🏫" },
        { href: "/students", label: "Daftar Siswa", icon: "🎓" },
        { href: "/employees", label: "Staf & Pengawas", icon: "👨‍🏫" },
        { href: "/asesmen/cbt/ruang", label: "Ruang Ujian", icon: "🏢" },
      ],
    },
    {
      group: "Manajemen Soal",
      items: [
        { href: "/asesmen/cbt/bank-soal", label: "Bank Soal", icon: "🗄️" },
        { href: "/asesmen/cbt/paket", label: "Paket Ujian", icon: "📦" },
      ],
    },
    {
      group: "Pelaksanaan",
      items: [
        { href: "/asesmen/cbt/sesi", label: "Jadwal Sesi", icon: "📅" },
        {
          href: "/asesmen/cbt/proctoring",
          label: "Proctoring Live",
          icon: "👁️",
        },
      ],
    },
    {
      group: "Laporan",
      items: [
        { href: "/asesmen/cbt/rekap", label: "Rekap Nilai", icon: "📝" },
        { href: "/asesmen/cbt/dokumen", label: "Dokumen & Cetak", icon: "🖨️" },
      ],
    },
    {
      group: "Pengaturan",
      items: [
        { href: "/settings/account", label: "Profil Saya", icon: "👤" },
        { href: "/settings/backups", label: "Backup & Restore", icon: "💾" },
      ],
    },
  ];
</script>

<div class="min-h-screen bg-slate-100 text-slate-950">
  <div class="mx-auto flex w-full max-w-7xl gap-4 px-3 py-4 lg:px-5">
    <aside
      class="hidden w-72 shrink-0 rounded-3xl border border-slate-200 bg-white p-4 shadow-sm lg:block"
    >
      <div class="mb-5 rounded-2xl bg-emerald-700 p-4 text-white">
        <p
          class="text-[10px] font-black uppercase tracking-[0.18em] text-emerald-100"
        >
          Mode CBT Familiar
        </p>
        <h2 class="mt-1 text-xl font-black">CBT Command</h2>
        <p class="mt-1 text-xs font-semibold leading-5 text-emerald-50">
          Tampilan mengikuti web app CBT lama, data tetap Super App.
        </p>
      </div>
      <nav class="space-y-4">
        {#each sections as section (section.group)}
          <div>
            <p
              class="mb-2 px-2 text-[10px] font-black uppercase tracking-[0.18em] text-slate-400"
            >
              {section.group}
            </p>
            <div class="space-y-1">
              {#each section.items as item (item.href)}
                <a
                  href={item.href}
                  class="flex items-center gap-3 rounded-2xl px-3 py-2 text-sm font-bold text-slate-700 hover:bg-emerald-50 hover:text-emerald-800"
                >
                  <span>{item.icon}</span><span>{item.label}</span>
                </a>
              {/each}
            </div>
          </div>
        {/each}
      </nav>
    </aside>
    <div class="min-w-0 flex-1">
      <div
        class="mb-4 rounded-3xl border border-slate-200 bg-white p-3 shadow-sm lg:hidden"
      >
        <p
          class="text-xs font-black uppercase tracking-[0.18em] text-emerald-700"
        >
          Mode CBT Familiar
        </p>
        <div class="mt-2 flex gap-2 overflow-x-auto pb-1">
          {#each sections.flatMap((section) => section.items) as item (item.href)}
            <a
              href={item.href}
              class="shrink-0 rounded-full border border-slate-200 px-3 py-1.5 text-xs font-black text-slate-700"
              >{item.label}</a
            >
          {/each}
        </div>
      </div>
      {@render children()}
    </div>
  </div>
</div>
