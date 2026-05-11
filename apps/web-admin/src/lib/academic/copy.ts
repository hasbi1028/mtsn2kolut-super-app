export const academicCopy = {
  actions: {
    activate: 'Aktifkan',
    add: 'Tambah',
    applyPromotion: 'Terapkan kenaikan kelas',
    cancel: 'Batalkan',
    checkImport: 'Cek data sebelum impor',
    clearSelection: 'Bersihkan pilihan',
    close: 'Tutup',
    continue: 'Lanjutkan',
    deactivate: 'Nonaktifkan',
    delete: 'Hapus',
    download: 'Unduh',
    downloadTemplate: 'Unduh template',
    edit: 'Ubah',
    exportData: 'Unduh data',
    importData: 'Impor data',
    preview: 'Pratinjau',
    previewPromotion: 'Pratinjau kenaikan kelas',
    refresh: 'Muat ulang',
    reset: 'Atur ulang',
    save: 'Simpan',
    saveAll: 'Simpan semua perubahan',
    selectAll: 'Pilih semua',
    viewDetail: 'Lihat detail'
  },
  labels: {
    academicDataCompleteness: 'Kelengkapan data akademik',
    academicYear: 'Tahun ajaran',
    activeAcademicYear: 'Tahun ajaran aktif',
    assessmentSubject: 'Dipakai untuk asesmen',
    classGroup: 'Rombel',
    classLevel: 'Tingkat',
    confirmationSentence: 'Kalimat konfirmasi',
    dataToCheck: 'Data yang perlu diperiksa',
    displayOrder: 'Urutan tampil',
    homeroomTeacher: 'Wali kelas',
    importCheck: 'Cek data sebelum impor',
    learningHour: 'Jam pelajaran',
    reportSubject: 'Masuk rapor',
    resultSummary: 'Ringkasan hasil',
    scheduleActivity: 'Aktivitas jadwal',
    selectedRows: 'Data dipilih',
    semester: 'Semester',
    subject: 'Mapel',
    subjectAssignment: 'Penugasan guru mapel',
    subjectSettings: 'Pengaturan mapel',
    subjectTeacher: 'Guru mapel',
    teacherAssignmentTable: 'Tabel penugasan guru mapel',
    unsavedChanges: 'Perubahan belum disimpan',
    warningNotes: 'Catatan yang perlu diperiksa',
    weeklyHours: 'JP per minggu',
    weeklySchedule: 'Jadwal mingguan'
  },
  helper: {
    academicDataCompleteness:
      'Periksa bagian yang belum lengkap agar data akademik siap dipakai untuk pembelajaran, asesmen, dan pelaporan.',
    confirmationSentence:
      'Ketik kalimat konfirmasi sesuai petunjuk sebelum menjalankan tindakan berisiko.',
    importCheck:
      'Unggah data untuk diperiksa terlebih dahulu. Sistem belum menyimpan perubahan sebelum Bapak/Ibu menekan tombol simpan atau impor.',
    promotionApply:
      'Tindakan ini menyiapkan data tahun ajaran tujuan berdasarkan pratinjau. Data tahun ajaran lama tidak dihapus.',
    promotionImpact:
      'Data tahun ajaran lama tidak dihapus. Sistem akan menyiapkan rombel tujuan, menyalin wali kelas, guru mapel, dan jadwal bila belum ada, lalu memindahkan siswa sesuai pratinjau.',
    promotionPreview:
      'Periksa daftar rombel, siswa, wali kelas, guru mapel, dan jadwal yang akan disiapkan untuk tahun ajaran tujuan.',
    teacherAssignmentTable:
      'Pilih guru pengampu untuk setiap mapel dan rombel. Perubahan dapat disimpan setelah diperiksa.',
    unsavedChanges:
      'Ada perubahan yang belum disimpan. Simpan perubahan atau batalkan sebelum meninggalkan halaman.'
  },
  empty: {
    noAcademicYear: 'Belum ada tahun ajaran yang tersedia.',
    noClassGroups: 'Belum ada rombel yang tersedia.',
    noSchedule: 'Belum ada jadwal yang dibuat.',
    noStudents: 'Belum ada siswa yang sesuai filter.',
    noSubjects: 'Belum ada mapel yang tersedia.',
    noTeachers: 'Belum ada guru yang tersedia.'
  },
  status: {
    active: 'Aktif',
    inactive: 'Tidak aktif',
    incomplete: 'Belum lengkap',
    needsReview: 'Perlu diperiksa',
    ready: 'Siap',
    saved: 'Tersimpan'
  },
  terms: {
    assignmentsCopied: 'Guru mapel disalin',
    classesCreated: 'Rombel baru dibuat',
    classesReused: 'Rombel yang sudah ada digunakan',
    homeroomsCopied: 'Wali kelas disalin',
    studentsSkipped: 'Siswa perlu tindak lanjut manual',
    timetableSlotsCopied: 'Jadwal disalin'
  }
} as const;

export type AcademicCopy = typeof academicCopy;
