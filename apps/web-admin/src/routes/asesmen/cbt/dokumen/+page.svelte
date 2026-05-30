<svelte:head>
  <title>Dokumen & Cetak</title>
  <meta
    name="description"
    content="Pusat dokumen mengikuti CBT lama: kartu peserta, denah ruang, daftar hadir, berita acara, lembar pengawas, dan rekap."
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
          Dokumen & Cetak
        </h1>
        <p class="mt-2 max-w-3xl text-sm leading-6 text-slate-600">
          Pusat dokumen mengikuti CBT lama: kartu peserta, denah ruang, daftar
          hadir, berita acara, lembar pengawas, dan rekap.
        </p>
      </div>
      <div class="flex flex-wrap gap-2">
        <a
          href="/asesmen/dokumen"
          class="rounded-2xl bg-emerald-700 text-white hover:bg-emerald-800 px-4 py-2 text-sm font-black shadow-sm"
          >Dokumen Ringkas</a
        >
        <a
          href="/asesmen"
          class="rounded-2xl border border-slate-300 bg-white text-slate-700 hover:border-emerald-300 hover:text-emerald-800 px-4 py-2 text-sm font-black shadow-sm"
          >Command Center</a
        >
      </div>
    </div>
  </header>

  <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <p
        class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-400"
      >
        Kartu
      </p>
      <p class="mt-2 text-3xl font-black text-slate-950">QR+PIN</p>
      <p class="mt-1 text-xs font-semibold text-slate-500">Aksi khusus</p>
    </div>
    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <p
        class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-400"
      >
        Denah
      </p>
      <p class="mt-2 text-3xl font-black text-slate-950">Ruang</p>
      <p class="mt-1 text-xs font-semibold text-slate-500">R01–R08</p>
    </div>
    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <p
        class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-400"
      >
        Daftar Hadir
      </p>
      <p class="mt-2 text-3xl font-black text-slate-950">Per ruang</p>
      <p class="mt-1 text-xs font-semibold text-slate-500">Cetak</p>
    </div>
    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <p
        class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-400"
      >
        BA
      </p>
      <p class="mt-2 text-3xl font-black text-slate-950">Arsip</p>
      <p class="mt-1 text-xs font-semibold text-slate-500">Final</p>
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
          <p class="font-black text-slate-950">Kartu Peserta</p>
          <p class="mt-1 text-slate-600">
            QR+PIN diterbitkan hanya saat panitia menekan tombol khusus dan siap
            cetak.
          </p>
        </div>
        <span
          class="w-fit rounded-full bg-emerald-50 px-3 py-1 text-xs font-black text-emerald-800"
          >Rahasia</span
        >
      </div>
      <div
        class="grid gap-3 p-4 text-sm md:grid-cols-[1fr_auto] md:items-center"
      >
        <div>
          <p class="font-black text-slate-950">Lembar Pengawas</p>
          <p class="mt-1 text-slate-600">
            Dokumen per ruang, bisa dipakai pengawas pengganti, tidak membuat
            PIN peserta.
          </p>
        </div>
        <span
          class="w-fit rounded-full bg-emerald-50 px-3 py-1 text-xs font-black text-emerald-800"
          >Aman</span
        >
      </div>
      <div
        class="grid gap-3 p-4 text-sm md:grid-cols-[1fr_auto] md:items-center"
      >
        <div>
          <p class="font-black text-slate-950">Berita Acara</p>
          <p class="mt-1 text-slate-600">
            Catatan pelaksanaan dan tanda tangan untuk arsip madrasah.
          </p>
        </div>
        <span
          class="w-fit rounded-full bg-emerald-50 px-3 py-1 text-xs font-black text-emerald-800"
          >Arsip</span
        >
      </div>
    </div>
  </section>
</div>
