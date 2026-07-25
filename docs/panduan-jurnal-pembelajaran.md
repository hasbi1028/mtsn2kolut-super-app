# 📘 Panduan Penggunaan Jurnal Pembelajaran (Class Journal)

## Ringkasan Sistem

Modul **Jurnal Pembelajaran** adalah sistem pencatatan kegiatan belajar mengajar harian dan kehadiran murid per pertemuan. Terintegrasi dengan modul **Assign Guru** (penugasan guru ke mapel) dan **Jadwal Pelajaran** (slot waktu mingguan).

---

## ✅ Status Kesiapan

| Komponen | Status | Data Saat Ini |
|----------|--------|:------------:|
| Guru (Pegawai) | ✅ Siap | **45 guru** |
| Kelas/Rombel | ✅ Siap | **10 kelas aktif** |
| Mata Pelajaran | ✅ Siap | **21 mapel** |
| Tahun Ajaran | ✅ Siap | **2026/2027 aktif** |
| Data Murid | ❌ **Perlu diisi** | **0 murid** |
| Assign Guru | ❌ **Perlu diisi** | **0 penugasan** |
| Jadwal Pelajaran | ❌ **Perlu diisi** | **0 slot** |
| Jurnal Harian | ❌ **Perlu diisi** | **0 sesi** |

---

## 📋 Prasyarat & Alur Penggunaan

### Langkah 1 — Isi Data Murid
**Menu:** `Kesiswaan → Data Murid`

Klik tombol **"+ Tambah Murid"** lalu isi form:
| Field | Wajib | Keterangan |
|-------|:-----:|------------|
| NIS | ✅ | Nomor Induk Siswa |
| Nama Lengkap | ✅ | Nama sesuai ijazah |
| Jenis Kelamin | ✅ | Laki-laki / Perempuan |
| NISN | ❌ | Nomor Induk Siswa Nasional |
| Rombel | ❌ | Pilih kelas dari dropdown |
| Status | ❌ | Aktif (default), Alumni, Mutasi, Dropout |
| NIK, Tempat Lahir, Tgl Lahir, Alamat, Agama, HP, Orang Tua | ❌ | Data pelengkap |

> **Data Murid** diperlukan sebagai daftar peserta didik yang akan diisi kehadirannya di jurnal.

### Langkah 2 — Assign Guru
**Menu:** `Akademik → Assign Guru`

1. Lihat tabel matrix **Mapel × Kelas**
2. Klik sel yang bertanda **`?`** (belum assign)
3. Pilih guru dari dialog radio
4. Klik **Simpan**
5. Status sel berubah dari `?` menjadi nama guru (Lengkap ✓)

> Syarat: **Murid** tidak diperlukan untuk assign guru. Namun **Guru, Kelas, Mapel** harus sudah ada.

### Langkah 3 — Buat Jadwal Pelajaran
**Menu:** `Akademik → Jadwal Pelajaran`

1. Setelah ada assign guru, halaman akan menampilkan **kartu per kelas**
2. Klik tombol **`+`** di sel kosong pada jadwal
3. Isi form:
   - **Mata Pelajaran** — pilih dari dropdown (assignment yang sudah dibuat)
   - **Hari** — Senin s.d. Sabtu
   - **Jam Mulai / Selesai**
   - **Ruang** — opsional
4. Klik **Simpan**
5. Slot jadwal tampil dengan informasi mapel & guru pengampu

### Langkah 4 — Catat Jurnal Harian
**Menu:** `Akademik → Jurnal Harian`

1. **Pilih Mata Pelajaran** dari dropdown
2. Klik tombol **"+ Catat Pertemuan"**
3. Isi form:
   - **Tanggal** — tanggal pertemuan
   - **Materi** — materi yang diajarkan
   - **Kegiatan** — deskripsi kegiatan
   - **Catatan** — catatan tambahan
   - **Guru Hadir** — toggle ya/tidak
4. Klik **Simpan**
5. Sesi baru muncul di tabel/kartu dengan nomor pertemuan otomatis

### Langkah 5 — Isi Kehadiran Murid
1. Klik tombol **"Absensi"** pada sesi jurnal
2. Dialog absensi menampilkan daftar murid per rombel
3. Klik status untuk setiap murid:
   - **Hadir** (hijau)
   - **Sakit**
   - **Izin**
   - **Alpha**
4. Klik **"Simpan Kehadiran"**

### Langkah 6 — Edit/Hapus
| Aksi | Cara |
|------|------|
| **Edit slot jadwal** | Klik **Edit** pada slot → ubah data → Simpan |
| **Hapus slot jadwal** | Klik **Hapus** pada slot → konfirmasi |
| **Edit kehadiran** | Klik **Absensi** → ubah status → Simpan Kehadiran |
| **Hapus sesi jurnal** | Klik **Hapus** pada sesi → konfirmasi |

---

## 🔗 Integrasi Antar Modul

```
Data Murid ──────────────┐
                         ├──→ Jurnal Harian (daftar hadir)
Assign Guru ─────────────┼──→ Jadwal Pelajaran (slot otomatis)
                         │
Guru ─── Kelas ─── Mapel ─┘
```

## ❓ Troubleshooting

### "Belum ada data jadwal"
> **Penyebab:** Belum ada **Assign Guru** atau kelas tidak memiliki penugasan.
> **Solusi:** Buka Assign Guru → klik sel `?` → pilih guru → Simpan.

### "Belum ada data kehadiran"
> **Penyebab:** Tidak ada murid terdaftar di rombel yang dipilih.
> **Solusi:** Tambah murid via Data Murid → pastikan rombel sesuai.

### "Assignment selector kosong"
> **Penyebab:** Belum ada **Assign Guru** yang dibuat.
> **Solusi:** Buat penugasan guru dulu di menu Assign Guru.

---

## 🧪 Cakupan E2E Tests

```
33 Tests ALL PASS ✅
Desktop: 27 tests — semua modul tercover CRUD
Mobile:   6 tests — responsive, no overflow, screenshots
```

Dengan panduan ini, operator/admin madrasah bisa langsung menggunakan modul Jurnal Pembelajaran secara bertahap.
