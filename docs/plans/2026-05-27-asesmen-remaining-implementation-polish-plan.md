# Asesmen Remaining Implementation & Polish Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** menuntaskan sisa implementasi penyederhanaan modul Asesmen/CBT agar alur utama benar-benar rapi, konsisten, dan aman dipakai operator/panitia/pengawas tanpa kebocoran surface teknis yang membingungkan.

**Architecture:** pertahankan backend, route lama, dan surface teknis yang masih dibutuhkan, tetapi rapikan lapisan web-admin menjadi dua jalur jelas: **alur utama sederhana** dan **Mode Lengkap Panitia**. Fokus tahap ini bukan penambahan fitur besar baru, tetapi penyelesaian kontrak route, alias, copy, CTA, MicroActionTable rollout, dan smoke/QA visual pada surface yang masih setengah jadi.

**Tech Stack:** SvelteKit web-admin, route-access helpers, Vitest sidebar/route-access tests, PM2 deploy untuk `mtsn2kolut-web-admin`.

---

## Outcome akhir yang diinginkan

Setelah plan ini selesai:
- operator/panitia melihat alur utama yang konsisten: `Ringkasan Ujian → Persiapan → Pelaksanaan → Ruang Saya/Pantau Ruang → Hasil`
- pengawas hanya melihat surface lapangan yang relevan
- surface teknis tetap hidup, tetapi masuk melalui `Mode Lengkap Panitia`
- CTA, judul, helper text, dan breadcrumb antar halaman tidak saling bertabrakan
- alias route seperti `/asesmen/ruang-saya` benar-benar stabil
- tidak ada lagi “menu muncul tapi permission beda”, “halaman sederhana tapi copy masih teknis”, atau “jalur utama macet lalu user dilempar ke halaman lama yang ramai”

## Batasan

### In scope
- finalisasi route/permission/copy/CTA/alias untuk Asesmen web-admin
- penyelesaian Micro Table compact admin pada sisa surface utama
- browser smoke dan QA visual terfokus
- commit cleanup yang fokus

### Out of scope
- redesign total backend/API
- hapus route lama atau hapus fitur teknis secara destruktif
- perubahan DB/migration besar
- rewrite public `/ujian` atau `/pengawas-ujian` lagi kecuali bug minor copy/CTA yang sangat terkait

---

## Task 1: Audit gap implementasi yang masih tersisa

**Objective:** menghasilkan daftar gap nyata sebelum edit lanjutan, supaya finishing tidak berdasarkan asumsi.

**Files:**
- Read: `apps/web-admin/src/routes/asesmen/+page.svelte`
- Read: `apps/web-admin/src/routes/asesmen/ringkas/+page.svelte`
- Read: `apps/web-admin/src/routes/asesmen/persiapan/+page.svelte`
- Read: `apps/web-admin/src/routes/asesmen/pelaksanaan/+page.svelte`
- Read: `apps/web-admin/src/routes/asesmen/pengawasan/+page.svelte`
- Read: `apps/web-admin/src/routes/asesmen/aplikasi-siswa/+page.svelte`
- Read: `apps/web-admin/src/routes/asesmen/hasil/+page.svelte`
- Read: `apps/web-admin/src/routes/asesmen/panitia/+page.svelte`
- Read: `apps/web-admin/src/routes/asesmen/kegiatan/+page.svelte`
- Read: `apps/web-admin/src/routes/asesmen/sesi/+page.svelte`

**Step 1: Audit route map dan CTA lintas halaman**

Cek setiap halaman untuk pertanyaan ini:
- Apakah judulnya sudah selaras dengan IA baru?
- Apakah tombol balik/lanjut mengarah ke jalur utama, bukan ke surface lama yang ramai?
- Apakah ada copy lama seperti `Asesmen CBT`, `Beranda Asesmen CBT`, `Portal Pengawasan`, atau istilah teknis lain?
- Apakah `Mode Lengkap Panitia` konsisten sebagai pintu fitur teknis?

**Step 2: Audit permission/route contract**

Baca:
- `apps/web-admin/src/lib/server/route-access.ts`
- `apps/web-admin/src/lib/server/route-access.test.ts`
- `apps/web-admin/src/lib/components/sidebar/sidebar-config.ts`
- `apps/web-admin/src/lib/components/sidebar/sidebar-config.test.ts`

Checklist:
- `ringkas/persiapan` = operator lane
- `pelaksanaan` = operator **dan** proctor lane sesuai desain final
- `ruang-saya/pengawasan` = proctor lane
- `hasil` = result_read
- `aplikasi-siswa` = jalur panduan, bukan sidebar utama

**Step 3: Catat gap dalam 4 kategori**
- Contract gap
- Navigation gap
- Copy/CTA gap
- Visual density/consistency gap

**Step 4: Simpan audit note singkat di bagian atas plan ini atau file kerja sementara**

Tidak perlu laporan panjang; cukup bullet actionable.

---

## Task 2: Finalkan contract route sederhana vs route teknis

**Objective:** memastikan semua entry point Asesmen masuk akal dan tidak ada alias setengah jadi.

**Files:**
- Modify: `apps/web-admin/src/routes/asesmen/+page.server.ts`
- Modify: `apps/web-admin/src/routes/asesmen/ruang-saya/+page.server.ts`
- Modify if needed: `apps/web-admin/src/lib/server/route-access.ts`
- Test: `apps/web-admin/src/lib/server/route-access.test.ts`

**Step 1: Verifikasi perilaku `/asesmen`**

Target final:
- operator lane → `/asesmen/ringkas`
- proctor lane → `/asesmen/ruang-saya`
- result lane → `/asesmen/hasil`
- lainnya → `/`

**Step 2: Verifikasi alias `/asesmen/ruang-saya`**

Target final:
- route ini menjadi alias friendly yang stabil ke `/asesmen/pengawasan`
- jangan duplicate UI; cukup redirect server-side

**Step 3: Samakan permission `pelaksanaan` dengan desain final**

Keputusan final yang harus konsisten di semua tempat:
- jika `Pelaksanaan` memang pintu hari-H untuk operator **dan** pengawas, maka route-access, sidebar, dan teks helper harus mencerminkan itu
- jika ternyata mau proctor-only, rollback sidebar dan copy agar konsisten

**Step 4: Tambah/rapikan test permission**

Tambahkan assert untuk:
- operator bisa akses `pelaksanaan`
- proctor bisa akses `pelaksanaan`
- proctor bisa akses `ruang-saya`
- user biasa tanpa permission tetap tertolak

**Step 5: Run focused tests**

Run:
```bash
npm --prefix apps/web-admin run test:unit -- src/lib/server/route-access.test.ts
```

Expected: PASS

---

## Task 3: Rapikan jalur utama page-to-page

**Objective:** membuat halaman utama terasa satu workflow, bukan kumpulan halaman terpisah.

**Files:**
- Modify: `apps/web-admin/src/routes/asesmen/ringkas/+page.svelte`
- Modify: `apps/web-admin/src/routes/asesmen/persiapan/+page.svelte`
- Modify: `apps/web-admin/src/routes/asesmen/pelaksanaan/+page.svelte`
- Modify: `apps/web-admin/src/routes/asesmen/hasil/+page.svelte`
- Modify: `apps/web-admin/src/routes/asesmen/aplikasi-siswa/+page.svelte`
- Modify if needed: `apps/web-admin/src/routes/asesmen/+page.svelte`

**Step 1: Samakan header language**

Gunakan pola konsisten:
- label kecil: `Asesmen/Ujian Digital` atau `Ujian Digital`
- judul besar: sesuai fungsi halaman, mis. `Ringkasan Ujian`, `Persiapan Ujian`, `Pelaksanaan Ujian`, `Hasil & Penutupan`
- helper text singkat, operasional, tidak terlalu teknis

**Step 2: Rapikan CTA utama**

Target CTA lintas halaman:
- `Ringkasan` → `Persiapan`, `Pelaksanaan`, `Hasil`, `Mode Lengkap Panitia`
- `Persiapan` → `Kembali ke Ringkasan`, `Mode Lengkap Panitia`
- `Pelaksanaan` → `Ruang Saya`, `Persiapan`, `Hasil` sesuai role
- `Hasil` → `Kegiatan`, `Sesi`, `Pengawasan`, `Ringkasan`
- `Aplikasi Siswa` → `Ruang Saya`, `Latihan Lokal`, `Mode Lengkap Panitia`

**Step 3: Hapus copy lama yang bentrok**

Cari dan hapus/ubah istilah berikut bila masih ada:
- `Asesmen CBT`
- `Beranda Asesmen CBT`
- `Portal Pengawasan`
- `Browser Darurat` jika konteks user-facing sederhana bisa diganti `Mode Cadangan`
- istilah teknis internal yang tidak perlu untuk guru/pengawas

**Step 4: Pastikan breadcrumb mental konsisten**

User harus bisa paham:
- `Ringkasan` = pintu masuk operator
- `Pelaksanaan` = hari-H control lane
- `Ruang Saya` = ruang/pengawas
- `Mode Lengkap Panitia` = fitur teknis/lanjutan

**Step 5: Build setelah edit batch ini**

Run:
```bash
rm -rf apps/web-admin/build && npm --prefix apps/web-admin run build
```

Expected: PASS

---

## Task 4: Selesaikan rollout Micro Table pada surface yang masih setengah rapi

**Objective:** memastikan halaman operasional memakai pola compact admin yang konsisten, bukan campuran card lama dan table baru.

**Files:**
- Modify: `apps/web-admin/src/routes/asesmen/kegiatan/+page.svelte`
- Modify: `apps/web-admin/src/routes/asesmen/sesi/+page.svelte`
- Modify: `apps/web-admin/src/routes/asesmen/kegiatan/[id]/cetak/+page.svelte`
- Modify: `apps/web-admin/src/routes/asesmen/kegiatan/[id]/archive/+page.svelte`
- Test if needed: `apps/web-admin/src/lib/asesmen/document-print-readiness.test.ts`

**Step 1: Audit apakah action row sudah workflow-first**

Checklist:
- `Kegiatan`: `Buka Alur`, `Arsip`, `Hapus` masuk akal
- `Sesi`: status-shaped actions (`Daftarkan`, `Jadwalkan`, `Mulai`, `Pantau`, `BA Sesi`, `Detail`) konsisten desktop + mobile
- `Cetak`: tiga jalur utama dokumen sudah jelas
- `Archive`: checklist/closing flow tidak terlalu verbose

**Step 2: Rapikan text density dan tombol sekunder**

Aturan:
- aksi primer 1 saja per row
- aksi sekunder outline/ghost
- hindari 4+ tombol setara yang saling berebut perhatian

**Step 3: Audit mobile snippet**

Pastikan mobile view tidak lebih ramai daripada desktop.
Kalau perlu, kurangi helper text yang berulang pada mode mobile.

**Step 4: Verifikasi tidak ada token/code sensitif tampil berlebihan di tabel**

Jika ada raw token/kode ruang terlalu telanjang di screen mass list, mask atau pindahkan ke surface cetak/fallback yang eksplisit.

---

## Task 5: Rapikan sidebar, hidden surface, dan Mode Lengkap Panitia

**Objective:** memastikan user biasa tidak “nyasar” ke surface teknis dari navigasi utama, tapi operator tetap punya pintu lengkap yang jelas.

**Files:**
- Modify if needed: `apps/web-admin/src/lib/components/sidebar/sidebar-config.ts`
- Modify if needed: `apps/web-admin/src/lib/components/sidebar/sidebar-config.test.ts`
- Modify: `apps/web-admin/src/routes/asesmen/panitia/+page.svelte`
- Modify: `apps/web-admin/src/routes/asesmen/aplikasi-siswa/+page.svelte`

**Step 1: Re-verify five-leaf sidebar contract**

Visible leaves only:
- `/asesmen/ringkas`
- `/asesmen/persiapan`
- `/asesmen/pelaksanaan`
- `/asesmen/ruang-saya`
- `/asesmen/hasil`

**Step 2: Pastikan `Mode Lengkap Panitia` cukup kuat sebagai hub**

Hub ini harus memuat link jelas ke:
- `kegiatan`
- `paket`
- `sesi`
- `pengawasan`
- `aplikasi-siswa`
- `matrix`
- `release`
- `hasil`
- `non-tes` bila memang tetap dipertahankan

**Step 3: Pastikan halaman panduan siswa tidak terasa seperti release center**

`/asesmen/aplikasi-siswa` harus tetap jadi panduan sederhana, dengan link teknis hanya sebagai jalur admin kecil atau lewat `Mode Lengkap Panitia`.

**Step 4: Run sidebar tests**

Run:
```bash
npm --prefix apps/web-admin run test:unit -- src/lib/components/sidebar/sidebar-config.test.ts
```

Expected: PASS

---

## Task 6: QA visual dan browser smoke rute utama

**Objective:** memastikan implementasi yang secara code benar juga terasa benar saat dibuka.

**Files:**
- No source file required by default
- Optional fixes in any route audited during smoke

**Step 1: Smoke route auth boundaries**

Run:
```bash
curl -s -o /tmp/asesmen-root.html -w 'asesmen:%{http_code} redirect:%{redirect_url}\n' http://127.0.0.1:8021/asesmen
curl -s -o /tmp/asesmen-ringkas.html -w 'ringkas:%{http_code} redirect:%{redirect_url}\n' http://127.0.0.1:8021/asesmen/ringkas
curl -s -o /tmp/asesmen-persiapan.html -w 'persiapan:%{http_code} redirect:%{redirect_url}\n' http://127.0.0.1:8021/asesmen/persiapan
curl -s -o /tmp/asesmen-pelaksanaan.html -w 'pelaksanaan:%{http_code} redirect:%{redirect_url}\n' http://127.0.0.1:8021/asesmen/pelaksanaan
curl -s -o /tmp/asesmen-ruang-saya.html -w 'ruang-saya:%{http_code} redirect:%{redirect_url}\n' http://127.0.0.1:8021/asesmen/ruang-saya
curl -s -o /tmp/asesmen-hasil.html -w 'hasil:%{http_code} redirect:%{redirect_url}\n' http://127.0.0.1:8021/asesmen/hasil
```

Expected when logged out: protected routes return `302` to login.

**Step 2: Browser QA logged-in**

Gunakan browser automation untuk cek minimal:
- sidebar label
- header/title per halaman utama
- CTA utama tidak duplikatif
- tidak ada halaman yang terasa “kembali ke versi lama”

Prioritas visual:
- `/asesmen/ringkas`
- `/asesmen/persiapan`
- `/asesmen/pelaksanaan`
- `/asesmen/pengawasan` via alias `/asesmen/ruang-saya`
- `/asesmen/hasil`

**Step 3: Catat bug minor visual**

Kelompokkan bug jadi:
- must-fix before close
- can-follow-up later

---

## Task 7: Final verification, deploy, dan commit bersih

**Objective:** menutup wave ini dengan repo bersih dan deploy stabil.

**Files:**
- Only intended touched files
- Do not add unrelated docs/artifacts automatically

**Step 1: Run full verification batch**

```bash
npm --prefix apps/web-admin run test:unit -- src/lib/components/sidebar/sidebar-config.test.ts src/lib/server/route-access.test.ts
rm -rf apps/web-admin/build && npm --prefix apps/web-admin run build
git diff --check
npm run security:scan-secrets:staged
npm run security:scan-secrets
```

Expected: all pass

**Step 2: Restart frontend**

```bash
pm2 restart mtsn2kolut-web-admin --update-env
```

**Step 3: Final smoke**

```bash
curl -fsS -o /tmp/web-login.html -w 'login:%{http_code}\n' http://127.0.0.1:8021/login
curl -s -o /tmp/asesmen-ringkas.html -w 'ringkas:%{http_code}\n' http://127.0.0.1:8021/asesmen/ringkas
```

Expected:
- `/login` = `200`
- protected route logged-out = `302`

**Step 4: Inspect git status**

Run:
```bash
git status --short
```

Make sure no accidental files ikut staging.

**Step 5: Commit fokus**

Contoh:
```bash
git add [explicit file list]
git commit -m "polish asesmen simplified workflow"
```

---

## Risiko utama yang harus dijaga

- jangan sampai `pelaksanaan` punya contract berbeda antara sidebar, route-access, dan real use-case
- jangan sampai `/asesmen/ruang-saya` alias hidup tapi tidak tercermin di CTA utama
- jangan sampai `Mode Lengkap Panitia` kosong/lemot sehingga user dipaksa balik ke route teknis tersebar
- jangan sampai QA hanya build-pass tetapi alur visual masih membingungkan
- jangan commit file plan, backup, screenshot, atau artifact lain kecuali memang diminta

## Definition of done

Wave penyelesaian dianggap selesai jika:
- route contract final stabil dan dites
- jalur utama sederhana konsisten di semua halaman utama
- `Mode Lengkap Panitia` jadi satu pintu teknis yang cukup jelas
- alias `Ruang Saya` stabil
- CTA/operator copy tidak saling bertabrakan
- build, tests, smoke, restart sukses
- commit fokus sudah dibuat

## Rekomendasi urutan eksekusi nyata

1. Audit gap
2. Finalkan contract route/permission
3. Rapikan jalur utama page-to-page
4. Selesaikan Micro Table rollout yang masih setengah
5. Rapikan sidebar + Mode Lengkap Panitia
6. QA visual/browser smoke
7. Verify + deploy + commit
