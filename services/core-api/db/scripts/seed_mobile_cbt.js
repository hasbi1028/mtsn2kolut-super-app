/**
 * Seed data for manual Flutter CBT mobile testing.
 *
 * Usage:
 *   DATABASE_URL=postgresql://... npm run seed:mobile-cbt
 *
 * The seed is intentionally a script, not a migration. It creates or updates
 * a small active exam session and resets only the seeded participants so the
 * same tokens can be reused during manual app testing.
 */

import fs from 'fs/promises';
import path from 'path';
import zlib from 'zlib';
import pg from 'pg';
import { fileURLToPath } from 'url';

const { Pool } = pg;

const __dir = path.dirname(fileURLToPath(import.meta.url));
const CORE_API_DIR = path.resolve(__dir, '../..');
const DATABASE_URL = process.env.DATABASE_URL
  ?? 'postgresql://pusaka:pusaka_dev@localhost:5432/pusaka';
const API_BASE_URL = process.env.API_BASE_URL ?? 'http://localhost:8080';
const RESET_PARTICIPANT_STATE = !isFalse(process.env.RESET_CBT_MOBILE_SEED ?? 'true');
const SEED_ASSETS = !isFalse(process.env.SEED_CBT_MOBILE_ASSETS ?? 'true');
const DURATION_MINUTES = parseInt(process.env.SEED_CBT_DURATION_MINUTES ?? '120', 10);

const seedConfig = {
  academicYear: {
    name: '2026/2027 - Seed CBT Mobile',
    startDate: '2026-07-01',
    endDate: '2027-06-30',
  },
  schoolClass: {
    code: 'CBT-MOBILE-7A',
    name: 'Kelas VII A - Seed CBT Mobile',
    level: '7',
  },
  subject: {
    code: 'CBT-MOBILE',
    name: 'Simulasi CBT Mobile',
  },
  package: {
    title: 'Paket Simulasi Flutter CBT Mobile',
    description: 'Data seed untuk uji manual aplikasi Flutter CBT mobile.',
    durationMinutes: DURATION_MINUTES,
  },
  session: {
    title: 'Simulasi Manual Flutter CBT Mobile',
    roomName: 'Ruang Simulasi Mobile 01',
  },
  participants: [
    {
      nis: 'CBTMOB001',
      nisn: '9990000001',
      nama: 'Ahmad Fadli Seed',
      gender: 'L',
      parentName: 'Orang Tua Ahmad',
      parentPhone: '081200000001',
      token: 'a1b2c3d4',
      seatNo: 1,
    },
    {
      nis: 'CBTMOB002',
      nisn: '9990000002',
      nama: 'Siti Rahma Seed',
      gender: 'P',
      parentName: 'Orang Tua Siti',
      parentPhone: '081200000002',
      token: 'b2c3d4e5',
      seatNo: 2,
    },
  ],
};

const questions = [
  {
    code: 'MOB-PG-001',
    questionText: 'Hasil dari 12 x 8 adalah ....',
    questionType: 'multiple_choice',
    options: [
      { label: 'A', text: '86' },
      { label: 'B', text: '96' },
      { label: 'C', text: '108' },
      { label: 'D', text: '128' },
    ],
    answerKey: 'B',
    explanation: '12 x 8 = 96.',
    difficulty: 'easy',
    stemHtml: '<p>Hitung perkalian bilangan bulat berikut.</p>',
    stimulusHtml: '',
    explanationHtml: '<p>Perkalian 12 x 8 menghasilkan 96.</p>',
    rubricHtml: '',
    materialTopic: 'Operasi bilangan bulat',
    cognitiveLevel: 'C2',
    hotsFlag: false,
  },
  {
    code: 'MOB-RICH-002',
    questionText: 'Berdasarkan stimulus, berapa total buku yang dibawa petugas perpustakaan?',
    questionType: 'multiple_choice',
    options: [
      { label: 'A', text: '72 buku' },
      { label: 'B', text: '84 buku' },
      { label: 'C', text: '96 buku' },
      { label: 'D', text: '108 buku' },
    ],
    answerKey: 'C',
    explanation: '3 rak x 32 buku = 96 buku.',
    difficulty: 'medium',
    stemHtml: '<p>Setiap rak berisi <strong>32 buku</strong>. Petugas membawa 3 rak.</p>',
    stimulusHtml: '<p>Perhatikan informasi perpustakaan berikut:</p><ul><li>Rak pertama penuh.</li><li>Rak kedua penuh.</li><li>Rak ketiga penuh.</li></ul>',
    explanationHtml: '<p>Total buku dihitung dengan 3 x 32 = 96.</p>',
    rubricHtml: '',
    materialTopic: 'Literasi numerasi',
    cognitiveLevel: 'C3',
    hotsFlag: false,
  },
  {
    code: 'MOB-AUDIO-003',
    questionText: 'Pernyataan "Jaringan internet yang stabil membantu sinkronisasi jawaban" adalah ....',
    questionType: 'multiple_choice',
    options: [
      { label: 'A', text: 'Benar' },
      { label: 'B', text: 'Salah' },
    ],
    answerKey: 'A',
    explanation: 'Koneksi stabil membantu proses sinkronisasi jawaban ke server.',
    difficulty: 'easy',
    stemHtml: '<p>Dengarkan audio petunjuk jika tersedia, lalu jawab pernyataan berikut.</p>',
    stimulusHtml: '',
    explanationHtml: '<p>Jawaban benar karena koneksi stabil menurunkan risiko jawaban tertahan lokal.</p>',
    rubricHtml: '',
    materialTopic: 'Kesiapan teknis CBT',
    cognitiveLevel: 'C1',
    hotsFlag: false,
  },
  {
    code: 'MOB-ESSAY-004',
    questionText: 'Jelaskan dua hal yang harus dilakukan siswa bila koneksi aplikasi ujian menurun.',
    questionType: 'essay',
    options: [],
    // Existing schema still has the original answer_key CHECK for A-E.
    // Essay grading ignores answer_key, so A is used only to satisfy that legacy constraint.
    answerKey: 'A',
    explanation: 'Siswa perlu tetap di aplikasi, memberi tahu pengawas, dan menunggu sinkronisasi pulih.',
    difficulty: 'medium',
    stemHtml: '<p>Tulis jawaban singkat dengan bahasa sendiri.</p>',
    stimulusHtml: '<p>Situasi: indikator koneksi berubah menjadi <strong>Menurun</strong> saat ujian berjalan.</p>',
    explanationHtml: '',
    rubricHtml: '<ul><li>Menyebut tetap di aplikasi.</li><li>Menyebut lapor ke pengawas.</li><li>Menyebut menunggu status sinkron pulih.</li></ul>',
    materialTopic: 'Prosedur BYOD CBT',
    cognitiveLevel: 'C4',
    hotsFlag: true,
  },
];

function isFalse(value) {
  return ['0', 'false', 'no', 'off'].includes(String(value).trim().toLowerCase());
}

function maskDatabaseUrl(value) {
  return value.replace(/:([^:@/]+)@/, ':***@');
}

function toAssetDir() {
  const value = process.env.CBT_ASSET_DIR ?? 'data/cbt-assets';
  return path.isAbsolute(value) ? value : path.resolve(CORE_API_DIR, value);
}

function addMinutes(date, minutes) {
  return new Date(date.getTime() + minutes * 60_000);
}

function formatWita(date) {
  return new Intl.DateTimeFormat('id-ID', {
    timeZone: 'Asia/Makassar',
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date);
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
    [
      seedConfig.academicYear.name,
      seedConfig.academicYear.startDate,
      seedConfig.academicYear.endDate,
    ],
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
    [
      academicYearId,
      seedConfig.schoolClass.code,
      seedConfig.schoolClass.name,
      seedConfig.schoolClass.level,
    ],
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
    [seedConfig.subject.code, seedConfig.subject.name],
  );
}

async function upsertStudents(client, classId) {
  const rows = [];
  for (const student of seedConfig.participants) {
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
      [
        student.nis,
        student.nisn,
        student.nama,
        student.gender,
        student.parentName,
        student.parentPhone,
        classId,
      ],
    );
    rows.push({ ...row, token: student.token, seatNo: student.seatNo });
  }
  return rows;
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

  const values = questionValues(subjectId, question);
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
            grade_level = 7,
            cp_ref = 'CP-SEED-CBT-MOBILE',
            tp_ref = 'TP-SEED-CBT-MOBILE',
            kd_ref = 'KD-SEED-CBT-MOBILE',
            indicator_ref = 'IND-SEED-CBT-MOBILE',
            material_topic = $19,
            cognitive_level = $20,
            hots_flag = $21,
            media_asset_ids = '[]'::jsonb,
            workflow_status = 'approved',
            author_username = 'mobile-cbt-seeder',
            approver_username = 'mobile-cbt-seeder',
            approved_at = NOW(),
            writer_notes = 'Seed manual Flutter CBT mobile',
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
        academic_phase, grade_level, cp_ref, tp_ref, kd_ref, indicator_ref,
        material_topic, cognitive_level, hots_flag, media_asset_ids,
        workflow_status, version, author_username, approver_username,
        approved_at, writer_notes
      )
      VALUES (
        $1, $2, $3, $4, $5::jsonb,
        $6, $7, $8, $9, $10,
        $11, $12, $13, 'published',
        $14, $15, $16, $17,
        'D', 7, 'CP-SEED-CBT-MOBILE', 'TP-SEED-CBT-MOBILE',
        'KD-SEED-CBT-MOBILE', 'IND-SEED-CBT-MOBILE',
        $18, $19, $20, '[]'::jsonb,
        'approved', 1, 'mobile-cbt-seeder', 'mobile-cbt-seeder',
        NOW(), 'Seed manual Flutter CBT mobile'
      )
      RETURNING id, code
    `,
    values,
  );
}

function questionValues(subjectId, question) {
  const option = (label) => question.options.find((item) => item.label === label)?.text ?? '';
  return [
    subjectId,
    question.code,
    question.questionText,
    question.questionType,
    JSON.stringify(question.options),
    option('A'),
    option('B'),
    option('C'),
    option('D'),
    option('E'),
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
}

async function upsertPackage(client, subjectId) {
  const existing = await maybeOne(
    client,
    `
      SELECT id
      FROM cbt_packages
      WHERE subject_id = $1 AND title = $2
      ORDER BY created_at DESC
      LIMIT 1
    `,
    [subjectId, seedConfig.package.title],
  );

  if (existing) {
    return one(
      client,
      `
        UPDATE cbt_packages
        SET description = $2,
            duration_minutes = $3,
            randomize_questions = TRUE,
            is_active = TRUE,
            updated_at = NOW()
        WHERE id = $1
        RETURNING id
      `,
      [
        existing.id,
        seedConfig.package.description,
        seedConfig.package.durationMinutes,
      ],
    );
  }

  return one(
    client,
    `
      INSERT INTO cbt_packages (
        subject_id, title, description, duration_minutes, randomize_questions, is_active
      )
      VALUES ($1, $2, $3, $4, TRUE, TRUE)
      RETURNING id
    `,
    [
      subjectId,
      seedConfig.package.title,
      seedConfig.package.description,
      seedConfig.package.durationMinutes,
    ],
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

async function upsertSession(client, packageId, classId, startAt, endAt) {
  const existing = await maybeOne(
    client,
    `
      SELECT id
      FROM cbt_exam_sessions
      WHERE package_id = $1 AND class_id = $2 AND title = $3
      ORDER BY created_at DESC
      LIMIT 1
    `,
    [packageId, classId, seedConfig.session.title],
  );

  if (existing) {
    return one(
      client,
      `
        UPDATE cbt_exam_sessions
        SET scheduled_start = $2,
            scheduled_end = $3,
            status = 'active',
            scope_type = 'class',
            scope_ref = $4,
            mix_policy = 'same_class',
            assignment_mode = 'manual',
            allow_cross_grade = FALSE,
            is_special_event = FALSE,
            updated_at = NOW()
        WHERE id = $1
        RETURNING id
      `,
      [existing.id, startAt, endAt, classId],
    );
  }

  return one(
    client,
    `
      INSERT INTO cbt_exam_sessions (
        package_id, class_id, scope_type, scope_ref, mix_policy, assignment_mode,
        allow_cross_grade, is_special_event, title, scheduled_start, scheduled_end, status
      )
      VALUES (
        $1, $2, 'class', $3, 'same_class', 'manual',
        FALSE, FALSE, $4, $5, $6, 'active'
      )
      RETURNING id
    `,
    [packageId, classId, classId, seedConfig.session.title, startAt, endAt],
  );
}

async function upsertRoom(client, sessionId) {
  return one(
    client,
    `
      INSERT INTO cbt_exam_rooms (session_id, room_name, capacity)
      VALUES ($1, $2, 30)
      ON CONFLICT (session_id, room_name) DO UPDATE
      SET capacity = EXCLUDED.capacity
      RETURNING id, room_name
    `,
    [sessionId, seedConfig.session.roomName],
  );
}

async function upsertParticipants(client, sessionId, roomId, students) {
  const tokens = students.map((student) => student.token);
  await client.query(
    'UPDATE cbt_exam_participants SET token = $1 WHERE token = ANY($2::text[])',
    ['', tokens],
  );

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

  if (RESET_PARTICIPANT_STATE && participants.length > 0) {
    const participantIds = participants.map((participant) => participant.id);
    await client.query(
      'DELETE FROM cbt_student_answers WHERE participant_id = ANY($1::uuid[])',
      [participantIds],
    );
    await client.query(
      'DELETE FROM cbt_participant_events WHERE participant_id = ANY($1::uuid[])',
      [participantIds],
    );
  }

  return participants;
}

async function seedAssets(client, questionRows) {
  if (!SEED_ASSETS) {
    return [];
  }

  const assetDir = toAssetDir();
  await fs.mkdir(assetDir, { recursive: true });

  const byCode = new Map(questionRows.map((row) => [row.code, row.id]));
  const assets = [
    {
      questionId: byCode.get('MOB-RICH-002'),
      originalName: 'stimulus-mobile-cbt.png',
      storedName: 'manual-mobile-cbt-stimulus.png',
      mimeType: 'image/png',
      purpose: 'stimulus',
      body: createSamplePngBuffer(),
    },
    {
      questionId: byCode.get('MOB-AUDIO-003'),
      originalName: 'audio-mobile-cbt.wav',
      storedName: 'manual-mobile-cbt-audio.wav',
      mimeType: 'audio/wav',
      purpose: 'general',
      body: createSampleWavBuffer(),
    },
  ].filter((asset) => asset.questionId);

  const questionIds = assets.map((asset) => asset.questionId);
  await client.query(
    `
      DELETE FROM cbt_question_assets
      WHERE uploaded_by = 'mobile-cbt-seeder'
        AND question_id = ANY($1::uuid[])
    `,
    [questionIds],
  );

  const inserted = [];
  for (const asset of assets) {
    const storagePath = path.join(assetDir, asset.storedName);
    await fs.writeFile(storagePath, asset.body);
    const row = await one(
      client,
      `
        INSERT INTO cbt_question_assets (
          question_id, original_name, stored_name, mime_type,
          file_size, storage_path, purpose, uploaded_by
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, 'mobile-cbt-seeder')
        RETURNING id, original_name, mime_type, purpose
      `,
      [
        asset.questionId,
        asset.originalName,
        asset.storedName,
        asset.mimeType,
        asset.body.length,
        storagePath,
        asset.purpose,
      ],
    );
    inserted.push(row);
  }
  return inserted;
}

function createSamplePngBuffer() {
  const width = 360;
  const height = 180;
  const bytesPerPixel = 4;
  const raw = Buffer.alloc((width * bytesPerPixel + 1) * height);

  for (let y = 0; y < height; y += 1) {
    const rowStart = y * (width * bytesPerPixel + 1);
    raw[rowStart] = 0;
    for (let x = 0; x < width; x += 1) {
      const offset = rowStart + 1 + x * bytesPerPixel;
      const inHeader = y < 42;
      const inCard = x > 44 && x < 316 && y > 66 && y < 136;
      const stripe = Math.floor((x + y) / 18) % 2 === 0;
      let color = inHeader ? [37, 110, 65] : [238, 245, 232];
      if (inCard) {
        color = stripe ? [178, 215, 190] : [250, 253, 247];
      }
      if (x < 8 || x >= width - 8 || y < 8 || y >= height - 8) {
        color = [23, 83, 48];
      }
      raw[offset] = color[0];
      raw[offset + 1] = color[1];
      raw[offset + 2] = color[2];
      raw[offset + 3] = 255;
    }
  }

  const signature = Buffer.from('89504e470d0a1a0a', 'hex');
  const ihdr = Buffer.alloc(13);
  ihdr.writeUInt32BE(width, 0);
  ihdr.writeUInt32BE(height, 4);
  ihdr[8] = 8;
  ihdr[9] = 6;
  ihdr[10] = 0;
  ihdr[11] = 0;
  ihdr[12] = 0;

  return Buffer.concat([
    signature,
    pngChunk('IHDR', ihdr),
    pngChunk('IDAT', zlib.deflateSync(raw)),
    pngChunk('IEND', Buffer.alloc(0)),
  ]);
}

function pngChunk(type, data) {
  const typeBuffer = Buffer.from(type, 'ascii');
  const length = Buffer.alloc(4);
  length.writeUInt32BE(data.length, 0);
  const crc = Buffer.alloc(4);
  crc.writeUInt32BE(crc32(Buffer.concat([typeBuffer, data])), 0);
  return Buffer.concat([length, typeBuffer, data, crc]);
}

function crc32(buffer) {
  let crc = 0xffffffff;
  for (const byte of buffer) {
    crc = (crc >>> 8) ^ crcTable[(crc ^ byte) & 0xff];
  }
  return (crc ^ 0xffffffff) >>> 0;
}

const crcTable = (() => {
  const table = new Uint32Array(256);
  for (let n = 0; n < 256; n += 1) {
    let value = n;
    for (let k = 0; k < 8; k += 1) {
      value = value & 1 ? 0xedb88320 ^ (value >>> 1) : value >>> 1;
    }
    table[n] = value >>> 0;
  }
  return table;
})();

function createSampleWavBuffer() {
  const sampleRate = 16000;
  const seconds = 1;
  const samples = sampleRate * seconds;
  const dataSize = samples * 2;
  const buffer = Buffer.alloc(44 + dataSize);

  buffer.write('RIFF', 0);
  buffer.writeUInt32LE(36 + dataSize, 4);
  buffer.write('WAVE', 8);
  buffer.write('fmt ', 12);
  buffer.writeUInt32LE(16, 16);
  buffer.writeUInt16LE(1, 20);
  buffer.writeUInt16LE(1, 22);
  buffer.writeUInt32LE(sampleRate, 24);
  buffer.writeUInt32LE(sampleRate * 2, 28);
  buffer.writeUInt16LE(2, 32);
  buffer.writeUInt16LE(16, 34);
  buffer.write('data', 36);
  buffer.writeUInt32LE(dataSize, 40);

  for (let i = 0; i < samples; i += 1) {
    const envelope = Math.min(1, i / 800, (samples - i) / 800);
    const value = Math.round(Math.sin((2 * Math.PI * 440 * i) / sampleRate) * 18000 * envelope);
    buffer.writeInt16LE(value, 44 + i * 2);
  }

  return buffer;
}

async function main() {
  if (process.env.NODE_ENV === 'production' && process.env.ALLOW_CBT_MOBILE_SEED !== 'true') {
    throw new Error('refusing to seed in NODE_ENV=production without ALLOW_CBT_MOBILE_SEED=true');
  }

  const now = new Date();
  const startAt = addMinutes(now, -10);
  const endAt = addMinutes(now, DURATION_MINUTES);
  const pool = new Pool({ connectionString: DATABASE_URL });
  const client = await pool.connect();

  console.log(`Postgres: ${maskDatabaseUrl(DATABASE_URL)}`);
  console.log(`API base : ${API_BASE_URL}`);
  console.log('');

  try {
    await client.query('BEGIN');

    const academicYear = await upsertAcademicYear(client);
    const schoolClass = await upsertSchoolClass(client, academicYear.id);
    const subject = await upsertSubject(client);
    const students = await upsertStudents(client, schoolClass.id);
    const questionRows = [];
    for (const question of questions) {
      questionRows.push(await upsertQuestion(client, subject.id, question));
    }
    const cbtPackage = await upsertPackage(client, subject.id);
    await linkPackageQuestions(client, cbtPackage.id, questionRows);
    const session = await upsertSession(client, cbtPackage.id, schoolClass.id, startAt, endAt);
    const room = await upsertRoom(client, session.id);
    const participants = await upsertParticipants(client, session.id, room.id, students);
    const assets = await seedAssets(client, questionRows);

    await client.query('COMMIT');

    console.log('Seed CBT mobile siap dipakai.');
    console.log('');
    console.log(`Sesi     : ${seedConfig.session.title}`);
    console.log(`Jadwal   : ${formatWita(startAt)} - ${formatWita(endAt)} WITA`);
    console.log(`Ruang    : ${room.room_name}`);
    console.log(`Soal     : ${questionRows.length}`);
    console.log(`Media    : ${assets.length} asset`);
    console.log('');
    console.log('Token siswa:');
    for (const participant of participants) {
      console.log(`- ${participant.token} | ${participant.nis} | ${participant.nama} | kursi ${participant.seatNo}`);
    }
    console.log('');
    console.log('Flutter login:');
    console.log(`- Base URL: ${API_BASE_URL}`);
    console.log('- Token   : gunakan salah satu token di atas');
    if (RESET_PARTICIPANT_STATE) {
      console.log('');
      console.log('State peserta seed sudah direset: device binding, jawaban, event, submit, dan urutan soal.');
    }
  } catch (err) {
    await client.query('ROLLBACK');
    throw err;
  } finally {
    client.release();
    await pool.end();
  }
}

main().catch((err) => {
  console.error('seed mobile CBT failed:', err.message);
  process.exit(1);
});
