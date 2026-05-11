# Academic Editable UI Contracts

Dokumen ini mendefinisikan kontrak data awal untuk Sprint Akademik 1 dan menjadi acuan Sprint 2+ agar UI editable tetap konsisten antara SvelteKit web-admin dan Go core-api.

## Dashboard Summary

Endpoint backend:

```http
GET /api/academic/dashboard
```

Endpoint SvelteKit BFF:

```http
GET /api/academic/dashboard
```

Response:

```ts
export type AcademicDashboardSummary = {
  active_academic_year: string;
  active_semester: string;
  total_classes: number;
  total_active_students: number;
  students_without_class: number;
  classes_without_homeroom: number;
  subject_assignments_missing_teacher: number;
  timetable_conflicts: number;
  student_accounts_missing: number;
  parent_accounts_missing: number;
};
```

Field notes:

- `active_academic_year`: nama tahun ajaran aktif; string kosong jika belum ada.
- `active_semester`: label semester turunan awal, `Ganjil` untuk Juli–Desember dan `Genap` untuk Januari–Juni.
- `total_classes`: jumlah rombel aktif pada tahun ajaran aktif.
- `total_active_students`: jumlah siswa aktif.
- `students_without_class`: siswa aktif yang belum terhubung ke rombel aktif tahun berjalan.
- `classes_without_homeroom`: rombel aktif tanpa wali kelas aktif.
- `subject_assignments_missing_teacher`: assignment guru mapel yang gurunya hilang/tidak aktif.
- `timetable_conflicts`: jumlah slot jadwal yang punya potensi bentrok kelas, guru, atau ruang.
- `student_accounts_missing`: siswa aktif tanpa akun user aktif.
- `parent_accounts_missing`: orang tua terkait siswa aktif tanpa akun user aktif.

## Editable Rombel Row

Dipakai untuk Sprint 2 ketika inline edit rombel diterapkan.

```ts
export type EditableRombelRow = {
  id: string;
  code: string;
  name: string;
  level: string;
  is_active: boolean;
  homeroom_teacher_id: string | null;
  homeroom_teacher_name: string;
  total_students: number;
  total_subject_assignments: number;
  total_timetable_slots: number;
};
```

Editable fields awal:

- `code`
- `name`
- `level`
- `is_active`
- `homeroom_teacher_id` lewat drawer/selector, bukan inline text bebas.

Validation awal:

- `code` wajib dan unik dalam tahun ajaran.
- `name` wajib.
- `level` harus sesuai level yang didukung sekolah.
- Deaktivasi rombel dengan siswa aktif harus dikonfirmasi atau ditolak oleh backend.

## Editable Subject / Mapel

Dipakai mulai Sprint 3.

Endpoint backend:

```http
GET /api/academic
POST /api/academic/subjects
PUT /api/academic/subjects/{id}
```

Endpoint SvelteKit BFF:

```http
GET /api/academic/subjects
POST /api/academic/subjects
PUT /api/academic/subjects?id={id}
```

```ts
export type AcademicSubject = {
  id: string;
  code: string;
  name: string;
  category: 'intrakurikuler' | 'muatan_lokal' | 'kokurikuler' | 'kegiatan' | 'lainnya' | string;
  is_assessment_subject: boolean;
  is_report_subject: boolean;
  is_schedule_activity: boolean;
  default_weekly_hours: number;
  display_order: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
};
```

Validation awal:

- `code` dan `name` wajib.
- `category` harus salah satu kategori UI yang didukung.
- `default_weekly_hours` harus 0–60.
- `display_order` tidak boleh negatif.
- `code` unik global antar mapel.

## Subject Assignment Matrix

Dipakai mulai Sprint 3.

Endpoint backend dan BFF:

```http
GET /api/academic/subject-assignment-matrix
PUT /api/academic/subject-assignment-matrix
```

```ts
export type SubjectMatrixCell = {
  class_id: string;
  subject_id: string;
  assignment_id: string | null;
  teacher_employee_id: string | null;
  teacher_name: string;
  status: 'complete' | 'missing_teacher' | 'missing_assignment';
};

export type SubjectAssignmentMatrix = {
  academic_year_id: string;
  academic_year_name: string;
  classes: Array<{ id: string; code: string; name: string; level: string }>;
  subjects: Array<Pick<AcademicSubject, 'id' | 'code' | 'name' | 'category' | 'is_assessment_subject' | 'is_report_subject' | 'is_schedule_activity' | 'default_weekly_hours' | 'display_order'>>;
  teachers: Array<{ id: string; nip: string; nama: string; unit_kerja: string }>;
  cells: SubjectMatrixCell[];
};

export type UpdateSubjectMatrixCellRequest = {
  class_id: string;
  subject_id: string;
  teacher_employee_id: string; // empty string clears assignment
};
```

Validation awal:

- Cell kosong diberi status `missing_assignment`; assignment dengan guru hilang/tidak aktif diberi status `missing_teacher`.
- Update hanya menerima rombel aktif tahun ajaran aktif, mapel aktif, dan guru aktif.
- Mengosongkan guru akan menghapus assignment melalui guard service existing, sehingga assignment yang sudah punya dependent tetap ditolak oleh backend.
- Perubahan matrix disimpan per changed cell dari UI dengan dirty-change bar; bulk preview lebih besar tetap ditunda.

## Weekly Timetable

Dipakai mulai Sprint 4.

Endpoint backend:

```http
GET /api/academic/timetable/weekly
GET /api/academic/timetable/conflicts
```

Endpoint SvelteKit BFF:

```http
GET /api/academic/timetable/weekly
GET /api/academic/timetable/conflicts
```

Mutasi slot tetap memakai endpoint rombel yang sudah ada:

```http
POST /api/academic/rombel/{class_id}/timetable-slots
PUT /api/academic/rombel/{class_id}/timetable-slots/{slot_id}
DELETE /api/academic/rombel/{class_id}/timetable-slots/{slot_id}
```

Response ringkas:

```ts
export type WeeklyTimetable = {
  active_academic_year_id: string;
  active_academic_year_name: string;
  classes: Array<{ id: string; code: string; name: string; level: string }>;
  teachers: Array<{ id: string; nip: string; nama: string; unit_kerja: string }>;
  subjects: Array<Pick<AcademicSubject, 'id' | 'code' | 'name' | 'category' | 'is_schedule_activity' | 'default_weekly_hours' | 'display_order'>>;
  assignments: Array<{
    id: string;
    class_id: string;
    subject_id: string;
    teacher_employee_id: string;
    class_code: string;
    subject_name: string;
    teacher_name: string;
  }>;
  slots: Array<{
    id: string;
    assignment_id: string;
    class_id: string;
    teacher_employee_id: string;
    day_of_week: number;
    start_time: string;
    end_time: string;
    room_label: string;
    conflict_status: 'ok' | 'conflict' | 'invalid_time_range';
    conflict_label: string;
    conflict_count: number;
  }>;
  conflicts: TimetableConflict[];
};
```

Conflict detection awal:

- `same_class`: slot rombel yang waktunya tumpang tindih.
- `same_teacher`: guru mengajar pada slot yang waktunya tumpang tindih.
- `same_room`: `room_label` sama dan tidak kosong pada slot yang waktunya tumpang tindih.
- `invalid_time_range`: `start_time >= end_time`, untuk menjaga data lama bila pernah melewati constraint.

## Academic Year, Rollover Preview, and Import Dry-Run

Dipakai mulai Sprint 5.

Endpoint backend:

```http
POST /api/academic/years/{id}/activate
POST /api/academic/year-rollover/preview
POST /api/academic/import-export/dry-run
```

Endpoint SvelteKit BFF:

```http
POST /api/academic/years/{id}/activate
POST /api/academic/year-rollover/preview
GET /api/academic/import-export/templates/{kind}
POST /api/academic/import-export/dry-run
```

Tahun ajaran masih dibaca/dibuat lewat kontrak akademik existing:

```http
GET /api/academic
POST /api/academic?entity=years
```

```ts
export type AcademicYear = {
  id: string;
  name: string; // format YYYY/YYYY
  start_date: string;
  end_date: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
};

export type YearRolloverPreview = {
  source_academic_year_id: string;
  source_academic_year_name: string;
  target_academic_year_id: string;
  target_academic_year_name: string;
  counts: {
    classes_to_create: number;
    students_to_promote: number;
    students_without_next_class: number;
    homeroom_assignments_to_copy: number;
    subject_assignments_to_copy: number;
    timetable_slots_to_copy: number;
  };
  classes_to_create: Array<{
    source_class_id: string;
    source_code: string;
    source_name: string;
    source_level: string;
    target_code: string;
    target_name: string;
    target_level: string;
  }>;
  students_to_promote: Array<{
    student_id: string;
    nis: string;
    nisn: string;
    nama: string;
    from_class_id: string;
    from_class_code: string;
    to_class_id: string;
    to_class_code: string;
    to_class_name: string;
  }>;
  students_without_next_class: Array<{
    student_id: string;
    nis: string;
    nisn: string;
    nama: string;
    from_class_id: string;
    from_class_code: string;
    reason: string;
  }>;
  warnings: string[];
};

export type AcademicImportDryRunResult = {
  kind: 'siswa' | 'rombel' | 'guru_mapel' | 'jadwal' | string;
  total_rows: number;
  add_count: number;
  update_count: number;
  skip_count: number;
  error_count: number;
  rows: Array<{ row: number; action: 'add' | 'update' | 'skip' | 'error' | string; summary: string }>;
  row_errors: Array<{ row: number; field: string; message: string }>;
};
```

Validation dan guardrails Sprint 5:

- Tahun ajaran baru wajib format `YYYY/YYYY`, rentang tahun berurutan, tanggal mulai < tanggal selesai, dan selalu dibuat nonaktif.
- Aktivasi tahun ajaran wajib challenge `AKTIFKAN`; aktivasi hanya mengubah flag aktif/nonaktif, tidak menghapus data lama.
- Rollover baru **preview-only**: tidak ada endpoint apply yang menjalankan mutasi massal sampai safety token/backup/QA disiapkan.
- Template CSV tersedia untuk `siswa`, `rombel`, `guru_mapel`, dan `jadwal` dengan header Bahasa Indonesia.
- Import dry-run hanya membaca CSV, memvalidasi baris, dan mengembalikan rencana tambah/ubah/skip/error; dry-run tidak mutate DB.
- Audit write-event akademik belum diaktifkan karena service audit existing hanya mendukung cleanup/read lifecycle; Sprint 5 mendokumentasikan deferred audit integration agar tidak membuat pola audit ad-hoc.

## Shared Editable Component Types

Komponen foundation Sprint 1 menggunakan tipe berikut:

```ts
export type EditableOption = {
  value: string;
  label: string;
  description?: string;
  disabled?: boolean;
};

export type EditableCellCommit<T = string> = {
  value: T;
  previousValue: T;
};

export type EditableCellCancel<T = string> = {
  value: T;
  originalValue: T;
};
```

## Guardrails

- Web-admin tidak boleh query DB langsung.
- Semua mutasi data akademik lewat Go API.
- Bulk mutation wajib menampilkan jumlah data terdampak.
- Operasi besar seperti rollover tahun ajaran wajib punya preview/dry-run.
