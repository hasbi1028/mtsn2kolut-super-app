<svelte:head>
  <title>Proctoring Live</title>
  <meta
    name="description"
    content="Monitoring familiar CBT lama: status peserta, peringatan, request selesai, dan aksi pengawas yang sederhana."
  />
</svelte:head>

<div class="space-y-4 pb-10">
  <header class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
    <p class="text-xs font-black uppercase tracking-[0.2em] text-emerald-700">
      Pelaksanaan
    </p>
    <div
      class="mt-2 flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between"
    >
      <div>
        <h1 class="text-3xl font-black tracking-tight md:text-4xl">
          Proctoring Live
        </h1>
        <p class="mt-2 max-w-3xl text-sm leading-6 text-slate-600">
          Monitoring familiar CBT lama: status peserta, peringatan, request
          selesai, dan aksi pengawas yang sederhana.
        </p>
      </div>
      <div class="flex flex-wrap gap-2">
        <a
          href="/pengawas-ujian?demo=1"
          class="rounded-2xl bg-emerald-700 text-white hover:bg-emerald-800 px-4 py-2 text-sm font-black shadow-sm"
          >Portal Pengawas</a
        >
        <a
          href="/asesmen/pelaksanaan"
          class="rounded-2xl border border-slate-300 bg-white text-slate-700 hover:border-emerald-300 hover:text-emerald-800 px-4 py-2 text-sm font-black shadow-sm"
          >Pelaksanaan</a
        >
      </div>
    </div>
  </header>

  <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <p
        class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-400"
      >
        Mengerjakan
      </p>
      <p class="mt-2 text-3xl font-black text-slate-950">—</p>
      <p class="mt-1 text-xs font-semibold text-slate-500">Live</p>
    </div>
    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <p
        class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-400"
      >
        Siap Masuk
      </p>
      <p class="mt-2 text-3xl font-black text-slate-950">—</p>
      <p class="mt-1 text-xs font-semibold text-slate-500">Menunggu</p>
    </div>
    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <p
        class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-400"
      >
        Perlu Dicek
      </p>
      <p class="mt-2 text-3xl font-black text-slate-950">0</p>
      <p class="mt-1 text-xs font-semibold text-slate-500">Peringatan</p>
    </div>
    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <p
        class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-400"
      >
        Selesai
      </p>
      <p class="mt-2 text-3xl font-black text-slate-950">—</p>
      <p class="mt-1 text-xs font-semibold text-slate-500">Terkirim</p>
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
          <p class="font-black text-slate-950">Status peserta</p>
          <p class="mt-1 text-slate-600">
            Siap Masuk, Mengerjakan, Waktu Habis, Selesai, Offline.
          </p>
        </div>
        <span
          class="w-fit rounded-full bg-emerald-50 px-3 py-1 text-xs font-black text-emerald-800"
          >Ringkas</span
        >
      </div>
      <div
        class="grid gap-3 p-4 text-sm md:grid-cols-[1fr_auto] md:items-center"
      >
        <div>
          <p class="font-black text-slate-950">Peringatan</p>
          <p class="mt-1 text-slate-600">
            Keluar halaman, koneksi tidak stabil, dan minta selesai masuk daftar
            perhatian.
          </p>
        </div>
        <span
          class="w-fit rounded-full bg-emerald-50 px-3 py-1 text-xs font-black text-emerald-800"
          >Prioritas</span
        >
      </div>
      <div
        class="grid gap-3 p-4 text-sm md:grid-cols-[1fr_auto] md:items-center"
      >
        <div>
          <p class="font-black text-slate-950">Aksi sensitif</p>
          <p class="mt-1 text-slate-600">
            Reset login dan finalisasi tetap di mode panitia/admin, bukan portal
            sederhana.
          </p>
        </div>
        <span
          class="w-fit rounded-full bg-emerald-50 px-3 py-1 text-xs font-black text-emerald-800"
          >Dibatasi</span
        >
      </div>
    </div>
  </section>
</div>
