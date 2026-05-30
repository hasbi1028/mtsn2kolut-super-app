<svelte:head>
  <title>Bank Soal</title>
  <meta
    name="description"
    content="Bank Soal tidak disalin dari CBT lama. Modul ini menjadi fasad familiar yang mengarahkan operator ke Bank Soal Super App yang sudah lebih lengkap."
  />
</svelte:head>

<div class="space-y-4 pb-10">
  <header class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
    <p class="text-xs font-black uppercase tracking-[0.2em] text-emerald-700">
      Manajemen Soal
    </p>
    <div
      class="mt-2 flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between"
    >
      <div>
        <h1 class="text-3xl font-black tracking-tight md:text-4xl">
          Bank Soal
        </h1>
        <p class="mt-2 max-w-3xl text-sm leading-6 text-slate-600">
          Bank Soal tidak disalin dari CBT lama. Modul ini menjadi fasad
          familiar yang mengarahkan operator ke Bank Soal Super App yang sudah
          lebih lengkap.
        </p>
      </div>
      <div class="flex flex-wrap gap-2">
        <a
          href="/bank-soal"
          class="rounded-2xl bg-emerald-700 text-white hover:bg-emerald-800 px-4 py-2 text-sm font-black shadow-sm"
          >Buka Daftar Soal</a
        >
        <a
          href="/bank-soal/tambah"
          class="rounded-2xl border border-slate-300 bg-white text-slate-700 hover:border-emerald-300 hover:text-emerald-800 px-4 py-2 text-sm font-black shadow-sm"
          >Tambah Soal</a
        >
        <a
          href="/bank-soal/verifikasi"
          class="rounded-2xl border border-slate-300 bg-white text-slate-700 hover:border-emerald-300 hover:text-emerald-800 px-4 py-2 text-sm font-black shadow-sm"
          >Verifikasi</a
        >
      </div>
    </div>
  </header>

  <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <p
        class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-400"
      >
        Soal
      </p>
      <p class="mt-2 text-3xl font-black text-slate-950">Super App</p>
      <p class="mt-1 text-xs font-semibold text-slate-500">
        Workflow tetap dipakai
      </p>
    </div>
    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <p
        class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-400"
      >
        Import
      </p>
      <p class="mt-2 text-3xl font-black text-slate-950">CSV</p>
      <p class="mt-1 text-xs font-semibold text-slate-500">Versi Super App</p>
    </div>
    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <p
        class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-400"
      >
        Review
      </p>
      <p class="mt-2 text-3xl font-black text-slate-950">RBAC</p>
      <p class="mt-1 text-xs font-semibold text-slate-500">
        Layak/Revisi/Tolak
      </p>
    </div>
    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <p
        class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-400"
      >
        Paket
      </p>
      <p class="mt-2 text-3xl font-black text-slate-950">Tersambung</p>
      <p class="mt-1 text-xs font-semibold text-slate-500">
        Dipakai Paket Ujian
      </p>
    </div>
  </section>

  <section class="rounded-3xl border border-slate-200 bg-white shadow-sm">
    <div class="border-b border-slate-200 p-4">
      <h2 class="text-lg font-black text-slate-950">Daftar operasional</h2>
      <p class="text-sm font-semibold text-slate-500">
        Pola tabel familiar: filter kecil, aksi jelas, dan status ringkas.
      </p>
    </div>
    <div class="divide-y divide-slate-200">
      <div
        class="grid gap-3 p-4 text-sm md:grid-cols-[1fr_auto] md:items-center"
      >
        <div>
          <p class="font-black text-slate-950">Daftar Soal</p>
          <p class="mt-1 text-slate-600">
            Filter, tambah, import, review, publish, dan arsip tetap memakai
            modul Bank Soal Super App.
          </p>
        </div>
        <span
          class="w-fit rounded-full bg-emerald-50 px-3 py-1 text-xs font-black text-emerald-800"
          >Disesuaikan</span
        >
      </div>
      <div
        class="grid gap-3 p-4 text-sm md:grid-cols-[1fr_auto] md:items-center"
      >
        <div>
          <p class="font-black text-slate-950">Paket Builder</p>
          <p class="mt-1 text-slate-600">
            Paket ujian mengambil soal dari Bank Soal Super App, bukan SQLite
            CBT lama.
          </p>
        </div>
        <span
          class="w-fit rounded-full bg-emerald-50 px-3 py-1 text-xs font-black text-emerald-800"
          >Tersambung</span
        >
      </div>
      <div
        class="grid gap-3 p-4 text-sm md:grid-cols-[1fr_auto] md:items-center"
      >
        <div>
          <p class="font-black text-slate-950">Cetak Soal</p>
          <p class="mt-1 text-slate-600">
            Gunakan halaman cetak Bank Soal Super App untuk kebutuhan
            naskah/arsip.
          </p>
        </div>
        <span
          class="w-fit rounded-full bg-emerald-50 px-3 py-1 text-xs font-black text-emerald-800"
          >Cetak</span
        >
      </div>
    </div>
  </section>

  <p
    class="rounded-3xl border border-amber-200 bg-amber-50 p-4 text-sm font-bold leading-6 text-amber-900"
  >
    Penyesuaian penting: data Bank Soal lama tidak dicopy mentah. Yang diambil
    hanya tampilan familiar dan pintasan aksi; data dan workflow tetap milik
    Super App.
  </p>
</div>
