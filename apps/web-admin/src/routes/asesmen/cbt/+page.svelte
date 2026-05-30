<svelte:head>
  <title>Dashboard Utama</title>
  <meta
    name="description"
    content="Panel familiar dari CBT lama: kesiapan data, sesi hari ini, ruang, proctoring, dan pintasan tindakan harian dalam satu layar."
  />
</svelte:head>

<div class="space-y-4 pb-10">
  <header class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
    <p class="text-xs font-black uppercase tracking-[0.2em] text-emerald-700">
      Command Center CBT
    </p>
    <div
      class="mt-2 flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between"
    >
      <div>
        <h1 class="text-3xl font-black tracking-tight md:text-4xl">
          Dashboard Utama
        </h1>
        <p class="mt-2 max-w-3xl text-sm leading-6 text-slate-600">
          Panel familiar dari CBT lama: kesiapan data, sesi hari ini, ruang,
          proctoring, dan pintasan tindakan harian dalam satu layar.
        </p>
      </div>
      <div class="flex flex-wrap gap-2">
        <a
          href="/asesmen/persiapan"
          class="rounded-2xl bg-emerald-700 text-white hover:bg-emerald-800 px-4 py-2 text-sm font-black shadow-sm"
          >Mulai Persiapan</a
        >
        <a
          href="/asesmen"
          class="rounded-2xl border border-slate-300 bg-white text-slate-700 hover:border-emerald-300 hover:text-emerald-800 px-4 py-2 text-sm font-black shadow-sm"
          >Mode Ringkas</a
        >
      </div>
    </div>
  </header>

  <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <p
        class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-400"
      >
        Siswa
      </p>
      <p class="mt-2 text-3xl font-black text-slate-950">—</p>
      <p class="mt-1 text-xs font-semibold text-slate-500">
        Ambil dari Data Siswa Super App
      </p>
    </div>
    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <p
        class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-400"
      >
        Ruang
      </p>
      <p class="mt-2 text-3xl font-black text-slate-950">8</p>
      <p class="mt-1 text-xs font-semibold text-slate-500">
        Target ruang operasional
      </p>
    </div>
    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <p
        class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-400"
      >
        Sesi Aktif
      </p>
      <p class="mt-2 text-3xl font-black text-slate-950">—</p>
      <p class="mt-1 text-xs font-semibold text-slate-500">
        Dari jadwal asesmen
      </p>
    </div>
    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <p
        class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-400"
      >
        Alert
      </p>
      <p class="mt-2 text-3xl font-black text-slate-950">0</p>
      <p class="mt-1 text-xs font-semibold text-slate-500">
        Tindak lanjut pengawas
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
          <p class="font-black text-slate-950">Kesiapan Peserta</p>
          <p class="mt-1 text-slate-600">
            Cek data siswa, kelas, ruang, dan kartu peserta sebelum sesi dibuka.
          </p>
        </div>
        <span
          class="w-fit rounded-full bg-emerald-50 px-3 py-1 text-xs font-black text-emerald-800"
          >Pusat Data</span
        >
      </div>
      <div
        class="grid gap-3 p-4 text-sm md:grid-cols-[1fr_auto] md:items-center"
      >
        <div>
          <p class="font-black text-slate-950">Agenda Hari Ini</p>
          <p class="mt-1 text-slate-600">
            Tinjau jadwal sesi dan ruang yang berjalan hari ini.
          </p>
        </div>
        <span
          class="w-fit rounded-full bg-emerald-50 px-3 py-1 text-xs font-black text-emerald-800"
          >Pelaksanaan</span
        >
      </div>
      <div
        class="grid gap-3 p-4 text-sm md:grid-cols-[1fr_auto] md:items-center"
      >
        <div>
          <p class="font-black text-slate-950">Kekurangan Operasional</p>
          <p class="mt-1 text-slate-600">
            Tampilkan peserta tanpa ruang/kartu dan paket yang belum siap.
          </p>
        </div>
        <span
          class="w-fit rounded-full bg-emerald-50 px-3 py-1 text-xs font-black text-emerald-800"
          >Perlu Dicek</span
        >
      </div>
    </div>
  </section>
</div>
