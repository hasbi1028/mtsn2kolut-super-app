<svelte:head>
  <title>Ruang Ujian</title>
  <meta
    name="description"
    content="Mengikuti menu Ruang Ujian CBT lama: daftar ruang, kapasitas, token ruang, denah, dan monitor ruang."
  />
</svelte:head>

<div class="space-y-4 pb-10">
  <header class="rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
    <p class="text-xs font-black uppercase tracking-[0.2em] text-emerald-700">
      Pusat Data
    </p>
    <div
      class="mt-2 flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between"
    >
      <div>
        <h1 class="text-3xl font-black tracking-tight md:text-4xl">
          Ruang Ujian
        </h1>
        <p class="mt-2 max-w-3xl text-sm leading-6 text-slate-600">
          Mengikuti menu Ruang Ujian CBT lama: daftar ruang, kapasitas, token
          ruang, denah, dan monitor ruang.
        </p>
      </div>
      <div class="flex flex-wrap gap-2">
        <a
          href="/asesmen"
          class="rounded-2xl bg-emerald-700 text-white hover:bg-emerald-800 px-4 py-2 text-sm font-black shadow-sm"
          >Atur Ruang</a
        >
        <a
          href="/pengawas-ujian?demo=1"
          class="rounded-2xl border border-slate-300 bg-white text-slate-700 hover:border-emerald-300 hover:text-emerald-800 px-4 py-2 text-sm font-black shadow-sm"
          >Portal Pengawas</a
        >
      </div>
    </div>
  </header>

  <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <p
        class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-400"
      >
        Ruang
      </p>
      <p class="mt-2 text-3xl font-black text-slate-950">8</p>
      <p class="mt-1 text-xs font-semibold text-slate-500">R01–R08</p>
    </div>
    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <p
        class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-400"
      >
        Kapasitas
      </p>
      <p class="mt-2 text-3xl font-black text-slate-950">240</p>
      <p class="mt-1 text-xs font-semibold text-slate-500">8 × 30</p>
    </div>
    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <p
        class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-400"
      >
        Token Ruang
      </p>
      <p class="mt-2 text-3xl font-black text-slate-950">Aktif</p>
      <p class="mt-1 text-xs font-semibold text-slate-500">Untuk pengawas</p>
    </div>
    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <p
        class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-400"
      >
        Denah
      </p>
      <p class="mt-2 text-3xl font-black text-slate-950">Siap</p>
      <p class="mt-1 text-xs font-semibold text-slate-500">Cetak/validasi</p>
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
          <p class="font-black text-slate-950">R01</p>
          <p class="mt-1 text-slate-600">
            Kapasitas 30, token ruang, monitor dan denah siap dipakai.
          </p>
        </div>
        <span
          class="w-fit rounded-full bg-emerald-50 px-3 py-1 text-xs font-black text-emerald-800"
          >Monitor</span
        >
      </div>
      <div
        class="grid gap-3 p-4 text-sm md:grid-cols-[1fr_auto] md:items-center"
      >
        <div>
          <p class="font-black text-slate-950">Validasi kursi</p>
          <p class="mt-1 text-slate-600">
            Cek nomor meja ganda, peserta tanpa ruang, dan overload kapasitas.
          </p>
        </div>
        <span
          class="w-fit rounded-full bg-emerald-50 px-3 py-1 text-xs font-black text-emerald-800"
          >Denah</span
        >
      </div>
      <div
        class="grid gap-3 p-4 text-sm md:grid-cols-[1fr_auto] md:items-center"
      >
        <div>
          <p class="font-black text-slate-950">Lembar Pengawas</p>
          <p class="mt-1 text-slate-600">
            Dokumen per ruang tanpa menerbitkan PIN peserta.
          </p>
        </div>
        <span
          class="w-fit rounded-full bg-emerald-50 px-3 py-1 text-xs font-black text-emerald-800"
          >Cetak</span
        >
      </div>
    </div>
  </section>
</div>
