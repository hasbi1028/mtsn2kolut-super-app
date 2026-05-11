export const operatorCopy = {
  commonActions: {
    save: 'Simpan',
    cancel: 'Batal',
    refresh: 'Muat ulang',
    retry: 'Coba ulang',
    checkData: 'Cek data',
    checkBeforeApply: 'Cek data sebelum diterapkan',
    applyChanges: 'Terapkan perubahan',
    download: 'Unduh data',
    upload: 'Unggah data',
    importData: 'Impor data',
    exportData: 'Unduh data',
    publish: 'Terbitkan',
    unpublish: 'Batalkan penerbitan',
    print: 'Cetak',
    viewDetails: 'Lihat detail',
    manage: 'Kelola'
  },
  commonStatus: {
    draft: 'Konsep',
    published: 'Terbit',
    active: 'Aktif',
    inactive: 'Tidak aktif',
    pending: 'Menunggu',
    processing: 'Sedang diproses',
    completed: 'Selesai',
    failed: 'Belum berhasil',
    ready: 'Siap digunakan',
    needsReview: 'Perlu diperiksa',
    missingData: 'Data belum lengkap'
  },
  commonMessages: {
    dataLoaded: 'Data berhasil dimuat.',
    dataSaved: 'Perubahan berhasil disimpan.',
    dataNotFound: 'Data tidak ditemukan.',
    dataInvalid: 'Data yang dikirim belum sesuai.',
    systemBusy: 'Layanan sistem belum dapat memproses permintaan. Coba lagi beberapa saat.',
    noPermission: 'Akun tidak memiliki izin untuk aksi ini.',
    confirmBeforeApply: 'Periksa kembali data sebelum menerapkan perubahan.',
    unsavedChanges: 'Ada perubahan yang belum disimpan.'
  },
  commonTerms: {
    systemService: 'layanan sistem',
    sentData: 'data yang dikirim',
    accessCode: 'kode akses',
    verificationCode: 'kode verifikasi',
    selected: 'dipilih',
    bulkAction: 'aksi massal',
    role: 'peran',
    permission: 'izin akses',
    auditLog: 'riwayat aktivitas',
    queue: 'antrian proses',
    sync: 'sinkronkan data',
    retry: 'coba ulang',
    fileFormat: 'format file',
    template: 'format isian',
    validation: 'pemeriksaan data',
    duplicate: 'data ganda',
    attachment: 'lampiran',
    evidence: 'bukti pendukung',
    workflow: 'alur kerja',
    cycle: 'siklus proses',
    version: 'versi aplikasi',
    release: 'rilis aplikasi',
    build: 'paket aplikasi',
    matrix: 'tabel pengaturan',
    debug: 'pemeriksaan masalah'
  },
  moduleTerms: {
    assessment: {
      cbt: 'Ujian Berbasis Komputer',
      event: 'kegiatan asesmen',
      session: 'sesi ujian',
      participant: 'peserta ujian',
      package: 'paket soal',
      room: 'ruang ujian',
      roomCode: 'kode ruang',
      examCode: 'kode ujian',
      proctoring: 'pengawasan ujian',
      readiness: 'kesiapan ujian',
      randomization: 'pengacakan soal',
      answerSheet: 'lembar jawaban',
      examCard: 'kartu ujian',
      minutes: 'berita acara'
    },
    questionBank: {
      questionBank: 'bank soal',
      item: 'butir soal',
      composer: 'penyusun soal',
      import: 'impor soal',
      parser: 'pembaca file',
      tag: 'penanda soal',
      difficulty: 'tingkat kesulitan',
      itemAnalysis: 'analisis butir',
      verification: 'verifikasi soal'
    },
    settings: {
      account: 'akun',
      user: 'pengguna',
      role: 'peran pengguna',
      rbac: 'hak akses pengguna',
      permission: 'izin akses',
      changeRequest: 'permintaan perubahan akun',
      analytics: 'ringkasan penggunaan',
      auditLog: 'riwayat aktivitas',
      schoolProfile: 'profil madrasah'
    },
    pusaka: {
      employee: 'pegawai',
      attendance: 'kehadiran',
      summary: 'ringkasan kehadiran',
      queue: 'antrian proses',
      sync: 'tarik data',
      retry: 'coba ulang',
      failed: 'belum berhasil'
    },
    office: {
      incomingLetter: 'surat masuk',
      outgoingLetter: 'surat keluar',
      disposition: 'disposisi',
      archive: 'arsip',
      certificateLetter: 'surat keterangan',
      compliancePack: 'paket kelengkapan dokumen',
      documentCycle: 'siklus dokumen',
      approval: 'persetujuan'
    },
    portal: {
      studentPortal: 'portal siswa',
      parentPortal: 'portal orang tua',
      guardian: 'orang tua/wali',
      studentAccount: 'akun siswa',
      parentAccount: 'akun orang tua',
      access: 'akses akun'
    },
    website: {
      page: 'halaman',
      post: 'berita/tulisan',
      announcement: 'pengumuman',
      slug: 'alamat halaman',
      seo: 'pengaturan pencarian',
      draft: 'konsep',
      publish: 'terbitkan',
      media: 'gambar/lampiran'
    },
    inventory: {
      item: 'barang',
      asset: 'aset/barang',
      stock: 'stok',
      loan: 'peminjaman',
      condition: 'kondisi barang'
    },
    library: {
      book: 'buku',
      loan: 'peminjaman',
      return: 'pengembalian',
      member: 'anggota perpustakaan'
    },
    governance: {
      governance: 'tata kelola',
      briefing: 'ringkasan arahan',
      evidence: 'bukti pendukung',
      owner: 'penanggung jawab',
      snp: 'Standar Nasional Pendidikan'
    }
  }
} as const;

export type OperatorCopy = typeof operatorCopy;
