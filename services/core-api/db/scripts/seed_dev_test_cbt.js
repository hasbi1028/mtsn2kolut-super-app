/**
 * Seed data for development testing.
 *
 * Creates:
 * - 1 kegiatan/event
 * - 1 paket soal
 * - 1 sesi ujian dengan durasi aktif 30 hari
 * - 1 ruang ujian
 * - 30 peserta ujian
 * - 1 pengawas ruang
 *
 * Usage:
 *   DATABASE_URL=postgresql://... npm run seed:dev-test-cbt
 */

import pg from 'pg';
import { fileURLToPath } from 'url';
import path from 'path';

const { Pool } = pg;
const __dir = path.dirname(fileURLToPath(import.meta.url));
void __dir;

const DATABASE_URL = process.env.DATABASE_URL ?? 'postgresql://pusaka:pusaka_dev@localhost:5432/pusaka';
const EVENT_DAYS = parseInt(process.env.SEED_DEV_CBT_EVENT_DAYS ?? '30', 10);
const PARTICIPANT_COUNT = parseInt(process.env.SEED_DEV_CBT_PARTICIPANT_COUNT ?? '30', 10);
const ROOM_CAPACITY = parseInt(process.env.SEED_DEV_CBT_ROOM_CAPACITY ?? String(PARTICIPANT_COUNT), 10);
const EXAM_DURATION_MINUTES = parseInt(process.env.SEED_DEV_CBT_DURATION_MINUTES ?? '120', 10);

const cfg = {
  academicYear: {
    name: '2026/2027 - Dev Test CBT',
    startDate: '2026-07-01',
    endDate: '2027-06-30',
  },
  schoolClass: {
    code: 'CBT-DEV-7A',
    name: 'Kelas VII A - Dev Test CBT',
    level: '7',
  },
  subject: {
    code: 'CBT-DEV',
    name: 'Simulasi CBT Development Test',
  },
  teacher: {
    nip: '199901012026011001',
    nama: 'Guru Dev Test CBT',
    pegawaiUid: '9990000001001',
  },
  proctor: {
    nip: '199901012026011002',
    nama: 'Pengawas Dev Test CBT',
    pegawaiUid: '9990000001002',
  },
  event: {
    title: 'Kegiatan Dev Test CBT',
    examType: 'tryout',
    scope: 'grade',
    targetLevels: ['7'],
    targetPg: 3,
    targetEssay: 1,
  },
  package: {
    title: 'Paket Dev Test CBT',
    description: 'Data seed untuk development test CBT.',
    durationMinutes: EXAM_DURATION_MINUTES,
  },
  session: {
    title: 'Sesi Dev Test CBT',
    roomName: 'Ruang Dev Test 01',
  },
};

const questions = [
  {
    code: 'DEV-PG-001',
    questionText: 'Hasil dari 12 + 18 adalah ....',
    questionType: 'multiple_choice',
    options: [
      { label: 'A', text: '20' },
      { label: 'B', text: '28' },
      { label: 'C', text: '30' },
      { label: 'D', text: '32' },
    ],
    answerKey: 'C',
    explanation: '12 + 18 = 30.',
    difficulty: 'easy',
    stemHtml: '<p>Hitung penjumlahan bilangan bulat berikut.</p>',
    stimulusHtml: '',
    explanationHtml: '<p>12 ditambah 18 sama dengan 30.</p>',
    rubricHtml: '',
    materialTopic: 'Operasi bilangan bulat',
    cognitiveLevel: 'C1',
    hotsFlag: false,
  },
  {
    code: 'DEV-PG-002',
    questionText: 'Berapakah hasil dari 7 x 9?',
    questionType: 'multiple_choice',
    options: [
      { label: 'A', text: '54' },
      { label: 'B', text: '56' },
      { label: 'C', text: '63' },
      { label: 'D', text: '72' },
    ],
    answerKey: 'C',
    explanation: '7 x 9 = 63.',
    difficulty: 'easy',
    stemHtml: '<p>Hitung perkalian berikut.</p>',
    stimulusHtml: '',
    explanationHtml: '<p>7 dikali 9 menghasilkan 63.</p>',
    rubricHtml: '',
    materialTopic: 'Operasi bilangan bulat',
    cognitiveLevel: 'C1',
    hotsFlag: false,
  },
  {
    code: 'DEV-PG-003',
    questionText: 'Pernyataan "Siswa harus tetap di aplikasi saat ujian berlangsung" adalah ....',
    questionType: 'multiple_choice',
    options: [
      { label: 'A', text: 'Benar' },
      { label: 'B', text: 'Salah' },
    ],
    answerKey: 'A',
    explanation: 'Tetap di aplikasi membantu mencegah gangguan ujian.',
    difficulty: 'easy',
    stemHtml: '<p>Pilih jawaban yang paling tepat.</p>',
    stimulusHtml: '',
    explanationHtml: '<p>Pernyataan tersebut benar.</p>',
    rubricHtml: '',
    materialTopic: 'Kesiapan teknis CBT',
    cognitiveLevel: 'C2',
    hotsFlag: false,
  },
  {
    code: 'DEV-ESSAY-004',
    questionText: 'Jelaskan dua langkah yang perlu dilakukan jika koneksi ujian menurun.',
    questionType: 'essay',
    options: [],
    answerKey: 'A',
    explanation: 'Tetap di aplikasi, lapor ke pengawas, dan tunggu sinkronisasi pulih.',
    difficulty: 'medium',
    stemHtml: '<p>Tulis jawaban singkat dengan bahasa sendiri.</p>',
    stimulusHtml: '<p>Situasi: indikator koneksi berubah menjadi <strong>Menurun</strong>.</p>',
    explanationHtml: '',
    rubricHtml: '<ul><li>Menyebut tetap di aplikasi.</li><li>Menyebut lapor ke pengawas.</li><li>Menyebut menunggu sinkron pulih.</li></ul>',
    materialTopic: 'Prosedur CBT',
    cognitiveLevel: 'C4',
    hotsFlag: true,
  },
];

function addDays(date, days) {
  return new Date(date.getTime() + days * 24 * 60 * 60 * 1000);
}

function formatWita(date) {
  return new Intl.DateTimeFormat('id-ID', {
    timeZone: 'Asia/Makassar',
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date);
}

function maskDatabaseUrl(value) {
  return value.replace(/:([^:@/]+)@/, ':***@');
}

async function one(client, text, params = []) {
  const result = await client.query(text, params);
  if (result.rowCount < 1) {
    throw new Error('expected one row, got none');
  }
  return result.rows[0];
}

async function maybeOne(client, text, params = []) {
  const result = await client.query(text, params);
  return result.rows[0] ?? null;
}

function buildParticipants(count) {
  return Array.from({ length: count }, (_, index) => {
    const n = String(index + 1).padStart(3, '0');
    return {
      nis: `DEV${n}`,
      nisn: `9980000${String(index + 1).padStart(4, '0')}`,
      nama: `Siswa Dev Test ${String(index + 1).padStart(2, '0')}`,
      gender: index % 2 === 0 ? 'L' : 'P',
      parentName: `Orang Tua Dev ${String(index + 1).padStart(2, '0')}`,
      parentPhone: `0813000${String(index + 1).padStart(4, '0')}`,
      token: `devtest${n}`,
      seatNo: index + 1,
    };
  });
}

async function upsertAcademicYear(client) {
  return one(
    client,
    `
      INSERT INTO academic_years (name, start_date, end_date, is_active)
      VALUES ($1, $2, $3, FALSE)
      ON CONFLICT (name) DO UPDATE
      SET start_date = EXCLUDED.start_date,
          end_date = EXCLUDED.end_date,
          updated_at = NOW()
      RETURNING id
    `,
    [cfg.academicYear.name, cfg.academicYear.startDate, cfg.academicYear.endDate],
  );
}

async function upsertSchoolClass(client, academicYearId) {
  return one(
    client,
    `
      INSERT INTO school_classes (academic_year_id, code, name, level, is_active)
      VALUES ($1, $2, $3, $4, TRUE)
      ON CONFLICT (academic_year_id, code) DO UPDATE
      SET name = EXCLUDED.name,
          level = EXCLUDED.level,
          is_active = TRUE,
          updated_at = NOW()
      RETURNING id
    `,
    [academicYearId, cfg.schoolClass.code, cfg.schoolClass.name, cfg.schoolClass.level],
  );
}

async function upsertSubject(client) {
  return one(
    client,
    `
      INSERT INTO subjects (code, name, is_active)
      VALUES ($1, $2, TRUE)
      ON CONFLICT (code) DO UPDATE
      SET name = EXCLUDED.name,
          is_active = TRUE,
          updated_at = NOW()
      RETURNING id
    `,
    [cfg.subject.code, cfg.subject.name],
  );
}

async function upsertEmployee(client, employee) {
  return one(
    client,
    `
      INSERT INTO employees (nip, nama, unit_kerja, is_active, employment_type, pegawai_uid, jenis_kelamin)
      VALUES ($1, $2, 'MTsN 2 Kolaka Utara', TRUE, 'pns', $3, 'L')
      ON CONFLICT (pegawai_uid) DO UPDATE
      SET nama = EXCLUDED.nama,
          unit_kerja = EXCLUDED.unit_kerja,
          is_active = TRUE,
          employment_type = EXCLUDED.employment_type,
          updated_at = NOW()
      RETURNING id
    `,
    [employee.nip, employee.nama, employee.pegawaiUid],
  );
}

async function upsertClassSubjectAssignment(client, classId, subjectId, teacherEmployeeId) {
  return one(
    client,
    `
      INSERT INTO class_subject_assignments (class_id, subject_id, teacher_employee_id)
      VALUES ($1, $2, $3)
      ON CONFLICT (class_id, subject_id) DO UPDATE
      SET teacher_employee_id = EXCLUDED.teacher_employee_id,
          updated_at = NOW()
      RETURNING id
    `,
    [classId, subjectId, teacherEmployeeId],
  );
}

async function upsertEvent(client, academicYearId) {
  const existing = await maybeOne(
    client,
    `
      SELECT id
      FROM cbt_exam_events
      WHERE title = $1
      ORDER BY created_at DESC
      LIMIT 1
    `,
    [cfg.event.title],
  );

  if (existing) {
    return one(
      client,
      `
        UPDATE cbt_exam_events
        SET exam_type = $2,
            scope = $3,
            target_levels = $4::text[],
            academic_year_id = $5,
            status = 'active',
            updated_at = NOW()
        WHERE id = $1
        RETURNING id
      `,
      [existing.id, cfg.event.examType, cfg.event.scope, cfg.event.targetLevels, academicYearId],
    );
  }

  return one(
    client,
    `
      INSERT INTO cbt_exam_events (title, exam_type, scope, target_levels, academic_year_id, status)
      VALUES ($1, $2, $3, $4::text[], $5, 'active')
      RETURNING id
    `,
    [cfg.event.title, cfg.event.examType, cfg.event.scope, cfg.event.targetLevels, academicYearId],
  );
}

async function upsertQuestionRequirement(client, eventId, subjectId) {
  const existing = await maybeOne(
    client,
    `
      SELECT id
      FROM cbt_event_question_requirements
      WHERE event_id = $1
        AND level = $2
        AND subject_id = $3
      ORDER BY created_at DESC
      LIMIT 1
    `,
    [eventId, cfg.event.targetLevels[0], subjectId],
  );

  if (existing) {
    return one(
      client,
      `
        UPDATE cbt_event_question_requirements
        SET scope_mode = 'pool_level_subject',
            class_id = NULL,
            target_pg = $2,
            target_essay = $3,
            status_filter = 'published_only',
            updated_at = NOW()
        WHERE id = $1
        RETURNING id
      `,
      [existing.id, cfg.event.targetPg, cfg.event.targetEssay],
    );
  }

  return one(
    client,
    `
      INSERT INTO cbt_event_question_requirements (
        event_id, scope_mode, level, class_id, subject_id, target_pg, target_essay, status_filter
      )
      VALUES ($1, 'pool_level_subject', $2, NULL, $3, $4, $5, 'published_only')
      RETURNING id
    `,
    [eventId, cfg.event.targetLevels[0], subjectId, cfg.event.targetPg, cfg.event.targetEssay],
  );
}

async function upsertQuestion(client, subjectId, question) {
  const existing = await maybeOne(
    client,
    `
      SELECT id
      FROM cbt_questions
      WHERE subject_id = $1 AND code = $2
      ORDER BY created_at DESC
      LIMIT 1
    `,
    [subjectId, question.code],
  );

  const values = [
    subjectId,
    question.code,
    question.questionText,
    question.questionType,
    JSON.stringify(question.options),
    question.options.find((item) => item.label === 'A')?.text ?? '',
    question.options.find((item) => item.label === 'B')?.text ?? '',
    question.options.find((item) => item.label === 'C')?.text ?? '',
    question.options.find((item) => item.label === 'D')?.text ?? '',
    question.options.find((item) => item.label === 'E')?.text ?? '',
    question.answerKey,
    question.explanation,
    question.difficulty,
    question.stemHtml,
    question.stimulusHtml,
    question.explanationHtml,
    question.rubricHtml,
    question.materialTopic,
    question.cognitiveLevel,
    question.hotsFlag,
  ];

  if (existing) {
    return one(
      client,
      `
        UPDATE cbt_questions
        SET subject_id = $2,
            code = $3,
            question_text = $4,
            question_type = $5,
            options = $6::jsonb,
            option_a = $7,
            option_b = $8,
            option_c = $9,
            option_d = $10,
            option_e = $11,
            answer_key = $12,
            explanation = $13,
            difficulty = $14,
            status = 'published',
            stem_html = $15,
            stimulus_html = $16,
            explanation_html = $17,
            rubric_html = $18,
            academic_phase = 'D',
            target_level = 'VII',
            cp_ref = 'CP-DEV-TEST',
            tp_ref = 'TP-DEV-TEST',
            kd_ref = 'KD-DEV-TEST',
            indicator_ref = 'IND-DEV-TEST',
            material_topic = $19,
            cognitive_level = $20,
            hots_flag = $21,
            media_asset_ids = '[]'::jsonb,
            workflow_status = 'siap_pakai',
            author_username = 'dev-test-seeder',
            approver_username = 'dev-test-seeder',
            approved_at = NOW(),
            writer_notes = 'Seed development test CBT',
            updated_at = NOW()
        WHERE id = $1
        RETURNING id, code
      `,
      [existing.id, ...values],
    );
  }

  return one(
    client,
    `
      INSERT INTO cbt_questions (
        subject_id, code, question_text, question_type, options,
        option_a, option_b, option_c, option_d, option_e,
        answer_key, explanation, difficulty, status,
        stem_html, stimulus_html, explanation_html, rubric_html,
        academic_phase, target_level, cp_ref, tp_ref, kd_ref, indicator_ref,
        material_topic, cognitive_level, hots_flag, media_asset_ids,
        workflow_status, version, author_username, approver_username,
        approved_at, writer_notes
      )
      VALUES (
        $1, $2, $3, $4, $5::jsonb,
        $6, $7, $8, $9, $10,
        $11, $12, $13, 'published',
        $14, $15, $16, $17,
        'D', 'VII', 'CP-DEV-TEST', 'TP-DEV-TEST',
        'KD-DEV-TEST', 'IND-DEV-TEST',
        $18, $19, $20, '[]'::jsonb,
        'siap_pakai', 1, 'dev-test-seeder', 'dev-test-seeder',
        NOW(), 'Seed development test CBT'
      )
      RETURNING id, code
    `,
    values,
  );
}

async function linkEventQuestions(client, eventId, questionRows) {
  await client.query(
    `
      UPDATE cbt_questions
      SET event_id = $1,
          updated_at = NOW()
      WHERE id = ANY($2::uuid[])
    `,
    [eventId, questionRows.map((row) => row.id)],
  );
}

async function upsertPackage(client, subjectId, eventId) {
  const existing = await maybeOne(
    client,
    `
      SELECT id
      FROM cbt_packages
      WHERE subject_id = $1 AND title = $2
      ORDER BY created_at DESC
      LIMIT 1
    `,
    [subjectId, cfg.package.title],
  );

  if (existing) {
    return one(
      client,
      `
        UPDATE cbt_packages
        SET event_id = $2,
            description = $3,
            duration_minutes = $4,
            randomize_questions = TRUE,
            randomize_options = TRUE,
            source_mode = 'event_pool',
            draw_pg_count = $5,
            draw_essay_count = $6,
            random_seed = 'seed-dev-test-cbt',
            is_active = TRUE,
            updated_at = NOW()
        WHERE id = $1
        RETURNING id
      `,
      [existing.id, eventId, cfg.package.description, cfg.package.durationMinutes, cfg.event.targetPg, cfg.event.targetEssay],
    );
  }

  return one(
    client,
    `
      INSERT INTO cbt_packages (
        subject_id, event_id, title, description, duration_minutes, randomize_questions,
        randomize_options, source_mode, draw_pg_count, draw_essay_count, random_seed, is_active
      )
      VALUES ($1, $2, $3, $4, $5, TRUE, TRUE, 'event_pool', $6, $7, 'seed-dev-test-cbt', TRUE)
      RETURNING id
    `,
    [subjectId, eventId, cfg.package.title, cfg.package.description, cfg.package.durationMinutes, cfg.event.targetPg, cfg.event.targetEssay],
  );
}

async function linkPackageQuestions(client, packageId, questionRows) {
  for (let index = 0; index < questionRows.length; index += 1) {
    await client.query(
      `
        INSERT INTO cbt_package_questions (package_id, question_id, position, points)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (package_id, question_id) DO UPDATE
        SET position = EXCLUDED.position,
            points = EXCLUDED.points
      `,
      [packageId, questionRows[index].id, index + 1, 1],
    );
  }
}

async function upsertSession(client, packageId, classId, eventId, startAt, endAt) {
  const existing = await maybeOne(
    client,
    `
      SELECT id
      FROM cbt_exam_sessions
      WHERE package_id = $1 AND class_id = $2 AND title = $3
      ORDER BY created_at DESC
      LIMIT 1
    `,
    [packageId, classId, cfg.session.title],
  );

  if (existing) {
    return one(
      client,
      `
        UPDATE cbt_exam_sessions
        SET event_id = $2,
            scheduled_start = $3,
            scheduled_end = $4,
            status = 'active',
            scope_type = 'class',
            scope_ref = $5,
            mix_policy = 'same_class',
            assignment_mode = 'manual',
            allow_cross_grade = FALSE,
            is_special_event = FALSE,
            access_mode = 'simulation',
            student_portal_direct_login_enabled = TRUE,
            require_room_token_for_web = FALSE,
            nisn_direct_login_enabled = TRUE,
            updated_at = NOW()
        WHERE id = $1
        RETURNING id
      `,
      [existing.id, eventId, startAt, endAt, classId],
    );
  }

  return one(
    client,
    `
      INSERT INTO cbt_exam_sessions (
        package_id, event_id, class_id, scope_type, scope_ref, mix_policy, assignment_mode,
        allow_cross_grade, is_special_event, access_mode, student_portal_direct_login_enabled,
        require_room_token_for_web, nisn_direct_login_enabled, title, scheduled_start, scheduled_end, status
      )
      VALUES (
        $1, $2, $3, 'class', $4, 'same_class', 'manual',
        FALSE, FALSE, 'simulation', TRUE, FALSE, TRUE, $5, $6, $7, 'active'
      )
      RETURNING id
    `,
    [packageId, eventId, classId, classId, cfg.session.title, startAt, endAt],
  );
}

async function upsertRoom(client, sessionId) {
  return one(
    client,
    `
      INSERT INTO cbt_exam_rooms (session_id, room_name, capacity, room_token, status, allow_web_fallback)
      VALUES ($1, $2, $3, 'ROOMDEV01', 'active', TRUE)
      ON CONFLICT (session_id, room_name) DO UPDATE
      SET capacity = EXCLUDED.capacity,
          room_token = EXCLUDED.room_token,
          status = 'active',
          allow_web_fallback = TRUE
      RETURNING id, room_name
    `,
    [sessionId, cfg.session.roomName, ROOM_CAPACITY],
  );
}

async function upsertRoomProctor(client, roomId, proctorEmployeeId, assignedByUserId) {
  await client.query('DELETE FROM cbt_room_proctors WHERE exam_room_id = $1', [roomId]);
  return one(
    client,
    `
      INSERT INTO cbt_room_proctors (exam_room_id, employee_id, role, assigned_by)
      VALUES ($1, $2, 'utama', $3)
      RETURNING id
    `,
    [roomId, proctorEmployeeId, assignedByUserId],
  );
}

async function upsertStudents(client, classId) {
  const students = buildParticipants(PARTICIPANT_COUNT);
  const rows = [];
  for (const student of students) {
    const row = await one(
      client,
      `
        INSERT INTO students (
          nis, nisn, nama, gender, parent_name, parent_phone, class_id, is_active, status
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, TRUE, 'active')
        ON CONFLICT (nis) DO UPDATE
        SET nisn = EXCLUDED.nisn,
            nama = EXCLUDED.nama,
            gender = EXCLUDED.gender,
            parent_name = EXCLUDED.parent_name,
            parent_phone = EXCLUDED.parent_phone,
            class_id = EXCLUDED.class_id,
            is_active = TRUE,
            status = 'active',
            updated_at = NOW()
        RETURNING id, nis, nama
      `,
      [student.nis, student.nisn, student.nama, student.gender, student.parentName, student.parentPhone, classId],
    );
    rows.push({ ...row, token: student.token, seatNo: student.seatNo });
  }
  return rows;
}

async function upsertParticipants(client, sessionId, roomId, students) {
  await client.query('UPDATE cbt_exam_participants SET token = $1 WHERE token = ANY($2::text[])', ['', students.map((student) => student.token)]);

  const participants = [];
  for (const student of students) {
    const row = await one(
      client,
      `
        INSERT INTO cbt_exam_participants (session_id, student_id, token, room_id, seat_no)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (session_id, student_id) DO UPDATE
        SET token = EXCLUDED.token,
            room_id = EXCLUDED.room_id,
            seat_no = EXCLUDED.seat_no,
            joined_at = NULL,
            submitted_at = NULL,
            score = NULL,
            device_fingerprint = NULL,
            question_order = NULL,
            last_heartbeat = NULL,
            app_switch_count = 0,
            screenshot_attempt = 0,
            login_ip = NULL,
            suspicious_flag = FALSE
        RETURNING id, token
      `,
      [sessionId, student.id, student.token, roomId, student.seatNo],
    );
    participants.push({ ...row, nis: student.nis, nama: student.nama, seatNo: student.seatNo });
  }

  await client.query('DELETE FROM cbt_student_answers WHERE participant_id = ANY($1::uuid[])', [participants.map((participant) => participant.id)]);
  await client.query('DELETE FROM cbt_participant_events WHERE participant_id = ANY($1::uuid[])', [participants.map((participant) => participant.id)]);

  return participants;
}

async function main() {
  if (process.env.NODE_ENV === 'production' && process.env.ALLOW_CBT_DEV_TEST_SEED !== 'true') {
    throw new Error('refusing to seed in NODE_ENV=production without ALLOW_CBT_DEV_TEST_SEED=true');
  }

  const now = new Date();
  const startAt = new Date(now.getTime() - 10 * 60 * 1000);
  const endAt = addDays(now, EVENT_DAYS);
  const pool = new Pool({ connectionString: DATABASE_URL });
  const client = await pool.connect();

  console.log(`Postgres: ${maskDatabaseUrl(DATABASE_URL)}`);
  console.log(`Peserta : ${PARTICIPANT_COUNT}`);
  console.log(`Ruang   : 1`);
  console.log(`Pengawas: 1`);
  console.log('');

  try {
    await client.query('BEGIN');

    const academicYear = await upsertAcademicYear(client);
    const schoolClass = await upsertSchoolClass(client, academicYear.id);
    const subject = await upsertSubject(client);
    const teacher = await upsertEmployee(client, cfg.teacher);
    const proctor = await upsertEmployee(client, cfg.proctor);
    const adminUser = await one(
      client,
      `
        SELECT id
        FROM users
        WHERE username = 'admin'
          AND deleted_at IS NULL
        ORDER BY created_at ASC
        LIMIT 1
      `,
    );
    await upsertClassSubjectAssignment(client, schoolClass.id, subject.id, teacher.id);
    const event = await upsertEvent(client, academicYear.id);
    await upsertQuestionRequirement(client, event.id, subject.id);

    const questionRows = [];
    for (const question of questions) {
      questionRows.push(await upsertQuestion(client, subject.id, question));
    }
    await linkEventQuestions(client, event.id, questionRows);

    const cbtPackage = await upsertPackage(client, subject.id, event.id);
    await linkPackageQuestions(client, cbtPackage.id, questionRows);

    const session = await upsertSession(client, cbtPackage.id, schoolClass.id, event.id, startAt, endAt);
    const room = await upsertRoom(client, session.id);
    await upsertRoomProctor(client, room.id, proctor.id, adminUser.id);

    const students = await upsertStudents(client, schoolClass.id);
    const participants = await upsertParticipants(client, session.id, room.id, students);

    await client.query('COMMIT');

    console.log('Seed dev test CBT siap dipakai.');
    console.log('');
    console.log(`Kegiatan : ${cfg.event.title}`);
    console.log(`Sesi     : ${cfg.session.title}`);
    console.log(`Ruang    : ${room.room_name}`);
    console.log(`Jadwal   : ${formatWita(startAt)} - ${formatWita(endAt)} WITA`);
    console.log(`Durasi   : ${EVENT_DAYS} hari kegiatan / ${cfg.package.durationMinutes} menit ujian`);
    console.log(`Soal     : ${questionRows.length}`);
    console.log(`Peserta  : ${participants.length}`);
    console.log('');
    console.log('Token siswa:');
    for (const participant of participants) {
      console.log(`- ${participant.token} | ${participant.nis} | ${participant.nama} | kursi ${participant.seatNo}`);
    }
    console.log('');
    console.log('Pengawas:');
    console.log(`- ${cfg.proctor.nama}`);
  } catch (err) {
    await client.query('ROLLBACK');
    throw err;
  } finally {
    client.release();
    await pool.end();
  }
}

main().catch((err) => {
  console.error('seed dev test CBT failed:', err.message);
  process.exit(1);
});
