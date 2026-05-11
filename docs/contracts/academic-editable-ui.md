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
