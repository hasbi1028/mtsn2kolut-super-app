<svelte:head>
  <title>Jadwal Sesi</title>
  <meta
    name="description"
    content="Jadwal Sesi mengikuti istilah CBT lama: paket, tanggal, jam, durasi, kelas target, ruang, dan pengawas."
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
          Jadwal Sesi
        </h1>
        <p class="mt-2 max-w-3xl text-sm leading-6 text-slate-600">
          Jadwal Sesi mengikuti istilah CBT lama: paket, tanggal, jam, durasi,
          kelas target, ruang, dan pengawas.
        </p>
      </div>
      <div class="flex flex-wrap gap-2">
        <a
          href="/asesmen"
          class="rounded-2xl bg-emerald-700 text-white hover:bg-emerald-800 px-4 py-2 text-sm font-black shadow-sm"
          >Buat Sesi</a
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
        Sesi
      </p>
      <p class="mt-2 text-3xl font-black text-slate-950">—</p>
      <p class="mt-1 text-xs font-semibold text-slate-500">Terjadwal</p>
    </div>
    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <p
        class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-400"
      >
        Durasi
      </p>
      <p class="mt-2 text-3xl font-black text-slate-950">90m</p>
      <p class="mt-1 text-xs font-semibold text-slate-500">
        Default dapat diubah
      </p>
    </div>
    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <p
        class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-400"
      >
        Kelas
      </p>
      <p class="mt-2 text-3xl font-black text-slate-950">VII–IX</p>
      <p class="mt-1 text-xs font-semibold text-slate-500">Target rombel</p>
    </div>
    <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <p
        class="text-[10px] font-black uppercase tracking-[0.16em] text-slate-400"
      >
        Status
      </p>
      <p class="mt-2 text-3xl font-black text-slate-950">Draft</p>
      <p class="mt-1 text-xs font-semibold text-slate-500">Siap dibuka</p>
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
          <p class="font-black text-slate-950">Sesi 1</p>
          <p class="mt-1 text-slate-600">
            Paket, jadwal, ruang, dan pengawas ditentukan sebelum portal peserta
            dibuka.
          </p>
        </div>
        <span
          class="w-fit rounded-full bg-emerald-50 px-3 py-1 text-xs font-black text-emerald-800"
          >Draft</span
        >
      </div>
      <div
        class="grid gap-3 p-4 text-sm md:grid-cols-[1fr_auto] md:items-center"
      >
        <div>
          <p class="font-black text-slate-950">Token sesi</p>
          <p class="mt-1 text-slate-600">
            Dipakai bila panitia ingin kontrol tambahan selain QR/PIN peserta.
          </p>
        </div>
        <span
          class="w-fit rounded-full bg-emerald-50 px-3 py-1 text-xs font-black text-emerald-800"
          >Opsional</span
        >
      </div>
      <div
        class="grid gap-3 p-4 text-sm md:grid-cols-[1fr_auto] md:items-center"
      >
        <div>
          <p class="font-black text-slate-950">Kunci sesi</p>
          <p class="mt-1 text-slate-600">
            Buka/tutup ujian tetap dari alur panitia, bukan dari siswa.
          </p>
        </div>
        <span
          class="w-fit rounded-full bg-emerald-50 px-3 py-1 text-xs font-black text-emerald-800"
          >Aman</span
        >
      </div>
    </div>
  </section>
</div>
