<svelte:head>
  <title>Rekap Nilai</title>
  <meta
    name="description"
    content="Rekap Nilai mengambil pola CBT lama tetapi tetap disiapkan untuk pipeline hasil Super App."
  />
</svelte:head>

<div class="space-y-4 pb-10">
  <header class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
    <p class="text-xs font-black uppercase tracking-[0.2em] text-emerald-700">
      Laporan
    </p>
    <div
      class="mt-2 flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between"
    >
      <div>
        <h1 class="text-3xl font-black tracking-tight md:text-4xl">
          Rekap Nilai
        </h1>
        <p class="mt-2 max-w-3xl text-sm leading-6 text-slate-600">
          Rekap Nilai mengambil pola CBT lama tetapi tetap disiapkan untuk
          pipeline hasil Super App.
        </p>
      </div>
      <div class="flex flex-wrap gap-2">
        <a
          href="/asesmen/hasil"
          class="rounded-2xl bg-emerald-700 text-white hover:bg-emerald-800 px-4 py-2 text-sm font-black shadow-sm"
          >Buka Hasil</a
        >
        <a
          href="/bank-soal"
          class="rounded-2xl border border-slate-300 bg-white text-slate-700 hover:border-emerald-300 hover:text-emerald-800 px-4 py-2 text-sm font-black shadow-sm"
          >Bank Soal</a
        >
      </div>
    </div>
  </header>

  <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <p
        class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-400"
      >
        Hasil
      </p>
      <p class="mt-2 text-3xl font-black text-slate-950">—</p>
      <p class="mt-1 text-xs font-semibold text-slate-500">Per sesi</p>
    </div>
    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <p
        class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-400"
      >
        PG
      </p>
      <p class="mt-2 text-3xl font-black text-slate-950">Otomatis</p>
      <p class="mt-1 text-xs font-semibold text-slate-500">Skor cepat</p>
    </div>
    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <p
        class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-400"
      >
        Essay
      </p>
      <p class="mt-2 text-3xl font-black text-slate-950">Manual</p>
      <p class="mt-1 text-xs font-semibold text-slate-500">Jika ada</p>
    </div>
    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <p
        class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-400"
      >
        Export
      </p>
      <p class="mt-2 text-3xl font-black text-slate-950">CSV/PDF</p>
      <p class="mt-1 text-xs font-semibold text-slate-500">Arsip</p>
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
          <p class="font-black text-slate-950">Rekap sesi</p>
          <p class="mt-1 text-slate-600">
            Nilai per peserta, status selesai, dan catatan pelaksanaan.
          </p>
        </div>
        <span
          class="w-fit rounded-full bg-emerald-50 px-3 py-1 text-xs font-black text-emerald-800"
          >Hasil</span
        >
      </div>
      <div
        class="grid gap-3 p-4 text-sm md:grid-cols-[1fr_auto] md:items-center"
      >
        <div>
          <p class="font-black text-slate-950">Analisis ringkas</p>
          <p class="mt-1 text-slate-600">
            Butir, rata-rata, dan peserta perlu tindak lanjut.
          </p>
        </div>
        <span
          class="w-fit rounded-full bg-emerald-50 px-3 py-1 text-xs font-black text-emerald-800"
          >Lanjutan</span
        >
      </div>
      <div
        class="grid gap-3 p-4 text-sm md:grid-cols-[1fr_auto] md:items-center"
      >
        <div>
          <p class="font-black text-slate-950">Arsip</p>
          <p class="mt-1 text-slate-600">
            Simpan rekap, daftar hadir, BA, dan dokumen ruang.
          </p>
        </div>
        <span
          class="w-fit rounded-full bg-emerald-50 px-3 py-1 text-xs font-black text-emerald-800"
          >Final</span
        >
      </div>
    </div>
  </section>
</div>
