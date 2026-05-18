-- Seed soal simulasi Bank Soal: Informatika VIII (20 PG + 5 Essay)
-- Idempotent: kode SIM-INF-VIII-* akan diupdate jika sudah ada.

BEGIN;

WITH subject_pick AS (
  SELECT id AS subject_id
  FROM subjects
  WHERE code = 'INF'
  ORDER BY created_at DESC
  LIMIT 1
), seed_rows AS (
  SELECT
    s.subject_id, v.code, v.question_type, v.question_text,
    v.option_a, v.option_b, v.option_c, v.option_d, v.option_e,
    v.answer_key, v.explanation, v.difficulty::cbt_question_difficulty_enum AS difficulty, v.options,
    v.material_topic, v.cognitive_level, v.hots_flag
  FROM subject_pick s
  CROSS JOIN (VALUES
    ('SIM-INF-VIII-PG-0001', 'multiple_choice', 'Apa fungsi utama sistem operasi pada komputer?', 'Mengatur perangkat keras dan perangkat lunak agar dapat digunakan pengguna', 'Menghapus seluruh data pengguna secara otomatis', 'Membuat komputer tidak membutuhkan listrik', 'Mengubah monitor menjadi printer', '', 'A', 'Sistem operasi bertugas mengelola sumber daya komputer dan menyediakan layanan bagi aplikasi/pengguna.', 'medium', 'C2', false, 'Literasi digital dan perangkat komputer', '[{"key":"A","text":"Mengatur perangkat keras dan perangkat lunak agar dapat digunakan pengguna"},{"key":"B","text":"Menghapus seluruh data pengguna secara otomatis"},{"key":"C","text":"Membuat komputer tidak membutuhkan listrik"},{"key":"D","text":"Mengubah monitor menjadi printer"}]'::jsonb),
    ('SIM-INF-VIII-PG-0002', 'multiple_choice', 'Perangkat berikut yang termasuk perangkat input adalah ....', 'Keyboard', 'Monitor', 'Speaker', 'Proyektor', '', 'A', 'Keyboard digunakan untuk memasukkan data/perintah ke komputer.', 'medium', 'C2', false, 'Literasi digital dan perangkat komputer', '[{"key":"A","text":"Keyboard"},{"key":"B","text":"Monitor"},{"key":"C","text":"Speaker"},{"key":"D","text":"Proyektor"}]'::jsonb),
    ('SIM-INF-VIII-PG-0003', 'multiple_choice', 'Contoh perangkat output pada komputer adalah ....', 'Mouse', 'Scanner', 'Monitor', 'Keyboard', '', 'C', 'Monitor menampilkan hasil pemrosesan komputer kepada pengguna.', 'medium', 'C2', false, 'Literasi digital dan perangkat komputer', '[{"key":"A","text":"Mouse"},{"key":"B","text":"Scanner"},{"key":"C","text":"Monitor"},{"key":"D","text":"Keyboard"}]'::jsonb),
    ('SIM-INF-VIII-PG-0004', 'multiple_choice', 'Apa yang dimaksud dengan internet?', 'Jaringan komputer global yang saling terhubung', 'Aplikasi untuk mengetik dokumen', 'Perangkat penyimpan data', 'Bahasa pemrograman khusus', '', 'A', 'Internet adalah jaringan global yang menghubungkan banyak jaringan komputer.', 'medium', 'C2', false, 'Literasi digital dan perangkat komputer', '[{"key":"A","text":"Jaringan komputer global yang saling terhubung"},{"key":"B","text":"Aplikasi untuk mengetik dokumen"},{"key":"C","text":"Perangkat penyimpan data"},{"key":"D","text":"Bahasa pemrograman khusus"}]'::jsonb),
    ('SIM-INF-VIII-PG-0005', 'multiple_choice', 'Sikap aman saat menerima tautan mencurigakan di pesan adalah ....', 'Langsung membuka tautan tersebut', 'Membagikannya ke semua teman', 'Memeriksa sumber dan tidak memasukkan data pribadi', 'Mengabaikan semua pesan dari guru', '', 'C', 'Tautan mencurigakan perlu dicek sumbernya agar terhindar dari phishing.', 'medium', 'C2', false, 'Literasi digital dan perangkat komputer', '[{"key":"A","text":"Langsung membuka tautan tersebut"},{"key":"B","text":"Membagikannya ke semua teman"},{"key":"C","text":"Memeriksa sumber dan tidak memasukkan data pribadi"},{"key":"D","text":"Mengabaikan semua pesan dari guru"}]'::jsonb),
    ('SIM-INF-VIII-PG-0006', 'multiple_choice', 'Kata sandi yang paling kuat adalah ....', '12345678', 'nama sendiri', 'kombinasi huruf besar-kecil, angka, dan simbol', 'tanggal lahir saja', '', 'C', 'Password kuat menggunakan kombinasi karakter dan tidak mudah ditebak.', 'medium', 'C2', false, 'Literasi digital dan perangkat komputer', '[{"key":"A","text":"12345678"},{"key":"B","text":"nama sendiri"},{"key":"C","text":"kombinasi huruf besar-kecil, angka, dan simbol"},{"key":"D","text":"tanggal lahir saja"}]'::jsonb),
    ('SIM-INF-VIII-PG-0007', 'multiple_choice', 'Dalam pengolah kata, fitur untuk menebalkan tulisan disebut ....', 'Italic', 'Bold', 'Underline', 'Align', '', 'B', 'Bold digunakan untuk membuat teks menjadi tebal.', 'medium', 'C2', false, 'Literasi digital dan perangkat komputer', '[{"key":"A","text":"Italic"},{"key":"B","text":"Bold"},{"key":"C","text":"Underline"},{"key":"D","text":"Align"}]'::jsonb),
    ('SIM-INF-VIII-PG-0008', 'multiple_choice', 'Ekstensi file gambar yang umum digunakan adalah ....', '.jpg', '.mp3', '.docx', '.xlsx', '', 'A', '.jpg adalah format berkas gambar.', 'medium', 'C2', false, 'Literasi digital dan perangkat komputer', '[{"key":"A","text":".jpg"},{"key":"B","text":".mp3"},{"key":"C","text":".docx"},{"key":"D","text":".xlsx"}]'::jsonb),
    ('SIM-INF-VIII-PG-0009', 'multiple_choice', 'Apa tujuan membuat folder pada komputer?', 'Mengelompokkan dan merapikan file', 'Menghapus sistem operasi', 'Mempercepat listrik', 'Mengubah jenis monitor', '', 'A', 'Folder membantu menyusun file agar lebih rapi dan mudah ditemukan.', 'medium', 'C2', false, 'Literasi digital dan perangkat komputer', '[{"key":"A","text":"Mengelompokkan dan merapikan file"},{"key":"B","text":"Menghapus sistem operasi"},{"key":"C","text":"Mempercepat listrik"},{"key":"D","text":"Mengubah jenis monitor"}]'::jsonb),
    ('SIM-INF-VIII-PG-0010', 'multiple_choice', 'Perangkat lunak untuk menjelajah internet disebut ....', 'Browser', 'Printer', 'Router', 'Keyboard', '', 'A', 'Browser seperti Chrome/Firefox digunakan untuk mengakses halaman web.', 'medium', 'C2', false, 'Literasi digital dan perangkat komputer', '[{"key":"A","text":"Browser"},{"key":"B","text":"Printer"},{"key":"C","text":"Router"},{"key":"D","text":"Keyboard"}]'::jsonb),
    ('SIM-INF-VIII-PG-0011', 'multiple_choice', 'Contoh alamat email yang benar adalah ....', 'siswa.example.com', 'siswa@example.com', 'siswa@example', '@siswa.example', '', 'B', 'Format email umum adalah nama@domain.', 'medium', 'C2', false, 'Literasi digital dan perangkat komputer', '[{"key":"A","text":"siswa.example.com"},{"key":"B","text":"siswa@example.com"},{"key":"C","text":"siswa@example"},{"key":"D","text":"@siswa.example"}]'::jsonb),
    ('SIM-INF-VIII-PG-0012', 'multiple_choice', 'Apa fungsi tombol Ctrl + C pada banyak aplikasi?', 'Menyalin data yang dipilih', 'Menutup komputer', 'Mencetak dokumen', 'Membuka kamera', '', 'A', 'Ctrl+C umumnya digunakan untuk copy/menyalin.', 'medium', 'C2', false, 'Literasi digital dan perangkat komputer', '[{"key":"A","text":"Menyalin data yang dipilih"},{"key":"B","text":"Menutup komputer"},{"key":"C","text":"Mencetak dokumen"},{"key":"D","text":"Membuka kamera"}]'::jsonb),
    ('SIM-INF-VIII-PG-0013', 'multiple_choice', 'Apa fungsi tombol Ctrl + V pada banyak aplikasi?', 'Memotong data', 'Menempelkan data yang sudah disalin', 'Menghapus permanen', 'Membuka pengaturan BIOS', '', 'B', 'Ctrl+V digunakan untuk paste/menempelkan data.', 'medium', 'C2', false, 'Literasi digital dan perangkat komputer', '[{"key":"A","text":"Memotong data"},{"key":"B","text":"Menempelkan data yang sudah disalin"},{"key":"C","text":"Menghapus permanen"},{"key":"D","text":"Membuka pengaturan BIOS"}]'::jsonb),
    ('SIM-INF-VIII-PG-0014', 'multiple_choice', 'Data pribadi yang sebaiknya tidak dibagikan sembarangan di internet adalah ....', 'Nomor identitas dan kata sandi', 'Judul buku pelajaran', 'Nama mata pelajaran', 'Warna favorit tanpa konteks akun', '', 'A', 'Nomor identitas dan password termasuk data sensitif.', 'medium', 'C2', false, 'Literasi digital dan perangkat komputer', '[{"key":"A","text":"Nomor identitas dan kata sandi"},{"key":"B","text":"Judul buku pelajaran"},{"key":"C","text":"Nama mata pelajaran"},{"key":"D","text":"Warna favorit tanpa konteks akun"}]'::jsonb),
    ('SIM-INF-VIII-PG-0015', 'multiple_choice', 'Apa yang dimaksud dengan perangkat lunak aplikasi?', 'Program yang membantu pengguna melakukan tugas tertentu', 'Kabel penghubung monitor', 'Bagian fisik komputer', 'Sumber listrik cadangan', '', 'A', 'Aplikasi adalah software untuk kebutuhan pengguna, misalnya mengetik, presentasi, atau belajar.', 'medium', 'C2', false, 'Literasi digital dan perangkat komputer', '[{"key":"A","text":"Program yang membantu pengguna melakukan tugas tertentu"},{"key":"B","text":"Kabel penghubung monitor"},{"key":"C","text":"Bagian fisik komputer"},{"key":"D","text":"Sumber listrik cadangan"}]'::jsonb),
    ('SIM-INF-VIII-PG-0016', 'multiple_choice', 'Contoh layanan penyimpanan awan adalah ....', 'Google Drive', 'Mouse', 'CPU', 'Kertas HVS', '', 'A', 'Google Drive menyimpan file di layanan cloud/awan.', 'medium', 'C2', false, 'Literasi digital dan perangkat komputer', '[{"key":"A","text":"Google Drive"},{"key":"B","text":"Mouse"},{"key":"C","text":"CPU"},{"key":"D","text":"Kertas HVS"}]'::jsonb),
    ('SIM-INF-VIII-PG-0017', 'multiple_choice', 'Apa manfaat membuat cadangan data?', 'Agar data tetap bisa dipulihkan jika hilang', 'Agar komputer pasti rusak', 'Agar file tidak bisa dibuka', 'Agar internet mati', '', 'A', 'Backup membantu memulihkan data saat perangkat bermasalah atau file terhapus.', 'medium', 'C2', false, 'Literasi digital dan perangkat komputer', '[{"key":"A","text":"Agar data tetap bisa dipulihkan jika hilang"},{"key":"B","text":"Agar komputer pasti rusak"},{"key":"C","text":"Agar file tidak bisa dibuka"},{"key":"D","text":"Agar internet mati"}]'::jsonb),
    ('SIM-INF-VIII-PG-0018', 'multiple_choice', 'Dalam etika digital, menulis komentar dengan bahasa sopan termasuk ....', 'Netiket yang baik', 'Peretasan', 'Spam', 'Phishing', '', 'A', 'Netiket adalah etika berkomunikasi di dunia digital.', 'medium', 'C2', false, 'Literasi digital dan perangkat komputer', '[{"key":"A","text":"Netiket yang baik"},{"key":"B","text":"Peretasan"},{"key":"C","text":"Spam"},{"key":"D","text":"Phishing"}]'::jsonb),
    ('SIM-INF-VIII-PG-0019', 'multiple_choice', 'Apa fungsi antivirus?', 'Membantu mendeteksi dan mencegah perangkat lunak berbahaya', 'Membuat baterai menjadi penuh', 'Mengganti monitor', 'Mengubah file menjadi buku', '', 'A', 'Antivirus membantu melindungi komputer dari malware/virus.', 'medium', 'C2', false, 'Literasi digital dan perangkat komputer', '[{"key":"A","text":"Membantu mendeteksi dan mencegah perangkat lunak berbahaya"},{"key":"B","text":"Membuat baterai menjadi penuh"},{"key":"C","text":"Mengganti monitor"},{"key":"D","text":"Mengubah file menjadi buku"}]'::jsonb),
    ('SIM-INF-VIII-PG-0020', 'multiple_choice', 'Jika komputer terasa lambat, tindakan awal yang wajar adalah ....', 'Memeriksa aplikasi yang berjalan dan menutup yang tidak diperlukan', 'Membanting perangkat', 'Membagikan password', 'Menghapus semua file sistem', '', 'A', 'Aplikasi terlalu banyak dapat membebani memori/prosesor.', 'medium', 'C2', false, 'Literasi digital dan perangkat komputer', '[{"key":"A","text":"Memeriksa aplikasi yang berjalan dan menutup yang tidak diperlukan"},{"key":"B","text":"Membanting perangkat"},{"key":"C","text":"Membagikan password"},{"key":"D","text":"Menghapus semua file sistem"}]'::jsonb),
    ('SIM-INF-VIII-ES-0001', 'essay', 'Jelaskan perbedaan perangkat keras dan perangkat lunak beserta masing-masing dua contohnya.', '', '', '', '', '', 'A', 'Perangkat keras adalah bagian fisik komputer seperti monitor dan keyboard. Perangkat lunak adalah program seperti sistem operasi dan aplikasi pengolah kata.', 'medium', 'C3', true, 'Literasi digital dan perangkat komputer', '[]'::jsonb),
    ('SIM-INF-VIII-ES-0002', 'essay', 'Uraikan langkah aman saat membuat kata sandi untuk akun belajar online.', '', '', '', '', '', 'A', 'Gunakan kombinasi huruf besar/kecil, angka, simbol, panjang cukup, tidak memakai tanggal lahir, dan tidak membagikan password kepada orang lain.', 'medium', 'C3', true, 'Literasi digital dan perangkat komputer', '[]'::jsonb),
    ('SIM-INF-VIII-ES-0003', 'essay', 'Jelaskan manfaat internet untuk kegiatan belajar siswa madrasah.', '', '', '', '', '', 'A', 'Internet membantu mencari referensi, mengakses materi digital, mengirim tugas, mengikuti pembelajaran daring, dan berkomunikasi dengan guru secara bijak.', 'medium', 'C3', true, 'Literasi digital dan perangkat komputer', '[]'::jsonb),
    ('SIM-INF-VIII-ES-0004', 'essay', 'Mengapa kita perlu berhati-hati saat mengunduh file dari internet?', '', '', '', '', '', 'A', 'Karena file dapat berisi virus/malware atau konten tidak aman. Pengguna perlu mengecek sumber, ekstensi file, dan memakai perlindungan keamanan.', 'medium', 'C3', true, 'Literasi digital dan perangkat komputer', '[]'::jsonb),
    ('SIM-INF-VIII-ES-0005', 'essay', 'Jelaskan apa yang dimaksud dengan jejak digital dan berikan contoh perilaku yang baik.', '', '', '', '', '', 'A', 'Jejak digital adalah rekaman aktivitas kita di internet. Contoh perilaku baik: berkomentar sopan, tidak menyebarkan hoaks, dan tidak membagikan data pribadi.', 'medium', 'C3', true, 'Literasi digital dan perangkat komputer', '[]'::jsonb)
  ) AS v(code, question_type, question_text, option_a, option_b, option_c, option_d, option_e, answer_key, explanation, difficulty, cognitive_level, hots_flag, material_topic, options)
), updated AS (
  UPDATE cbt_questions q
  SET subject_id = r.subject_id,
      question_type = r.question_type,
      question_text = r.question_text,
      option_a = r.option_a,
      option_b = r.option_b,
      option_c = r.option_c,
      option_d = r.option_d,
      option_e = r.option_e,
      answer_key = r.answer_key,
      explanation = r.explanation,
      difficulty = r.difficulty,
      status = 'published'::cbt_question_status_enum,
      options = r.options,
      stem_html = r.question_text,
      explanation_html = r.explanation,
      academic_phase = 'Fase D',
      target_level = 'VIII',
      cp_ref = 'SIM-CP-INF',
      tp_ref = 'SIM-TP-INF',
      kd_ref = 'SIM-KD-INF',
      indicator_ref = 'Simulasi bank soal 20 PG dan 5 Essay',
      material_topic = r.material_topic,
      cognitive_level = r.cognitive_level,
      hots_flag = r.hots_flag,
      workflow_status = 'published',
      author_username = 'system',
      writer_notes = 'Seed simulasi dibuat atas permintaan admin untuk uji coba paket soal.',
      updated_at = NOW()
  FROM seed_rows r
  WHERE q.code = r.code
  RETURNING q.id, q.code
), inserted AS (
  INSERT INTO cbt_questions (
    subject_id, code, question_type, question_text,
    option_a, option_b, option_c, option_d, option_e,
    answer_key, explanation, difficulty, status, options,
    stem_html, explanation_html, academic_phase, target_level,
    cp_ref, tp_ref, kd_ref, indicator_ref, material_topic, cognitive_level, hots_flag,
    workflow_status, author_username, writer_notes, updated_at
  )
  SELECT
    r.subject_id, r.code, r.question_type, r.question_text,
    r.option_a, r.option_b, r.option_c, r.option_d, r.option_e,
    r.answer_key, r.explanation, r.difficulty, 'published'::cbt_question_status_enum, r.options,
    r.question_text, r.explanation, 'Fase D', 'VIII',
    'SIM-CP-INF', 'SIM-TP-INF', 'SIM-KD-INF', 'Simulasi bank soal 20 PG dan 5 Essay', r.material_topic, r.cognitive_level, r.hots_flag,
    'published', 'system', 'Seed simulasi dibuat atas permintaan admin untuk uji coba paket soal.', NOW()
  FROM seed_rows r
  WHERE NOT EXISTS (SELECT 1 FROM updated u WHERE u.code = r.code)
    AND NOT EXISTS (SELECT 1 FROM cbt_questions q WHERE q.code = r.code)
  RETURNING id, code
), upserted AS (
  SELECT id, code FROM updated
  UNION ALL
  SELECT id, code FROM inserted), version_fix AS (
  UPDATE cbt_questions q
  SET version_group_id = q.id
  FROM upserted u
  WHERE q.id = u.id AND q.version_group_id IS NULL
  RETURNING q.id
), package_pick AS (
  SELECT p.id
  FROM cbt_packages p
  JOIN subject_pick s ON s.subject_id = p.subject_id
  WHERE p.title = 'SIMULASI - Informatika VIII (20 PG + 5 Essay)'
  LIMIT 1
), package_insert AS (
  INSERT INTO cbt_packages (subject_id, title, description, duration_minutes, randomize_questions, randomize_options, is_active, source_mode, draw_pg_count, draw_essay_count, composition_log, updated_at)
  SELECT s.subject_id, 'SIMULASI - Informatika VIII (20 PG + 5 Essay)', 'Paket simulasi otomatis berisi 20 pilihan ganda dan 5 essay untuk uji coba CBT/Bank Soal.', 60, true, true, true, 'teacher_class', 20, 5, '{"seed":"simulasi-bank-soal-20pg-5essay","subject":"Informatika","target_level":"VIII"}'::jsonb, NOW()
  FROM subject_pick s
  WHERE NOT EXISTS (SELECT 1 FROM package_pick)
  RETURNING id
), package_final AS (
  SELECT id FROM package_pick
  UNION ALL
  SELECT id FROM package_insert
), package_update AS (
  UPDATE cbt_packages p
  SET description = 'Paket simulasi otomatis berisi 20 pilihan ganda dan 5 essay untuk uji coba CBT/Bank Soal.',
      duration_minutes = 60,
      randomize_questions = true,
      randomize_options = true,
      is_active = true,
      source_mode = 'teacher_class',
      draw_pg_count = 20,
      draw_essay_count = 5,
      composition_log = '{"seed":"simulasi-bank-soal-20pg-5essay","subject":"Informatika","target_level":"VIII"}'::jsonb,
      updated_at = NOW()
  FROM package_final pf
  WHERE p.id = pf.id
  RETURNING p.id
), linked AS (
  INSERT INTO cbt_package_questions (package_id, question_id, position, points)
  SELECT pf.id, q.id,
         ROW_NUMBER() OVER (ORDER BY CASE WHEN q.question_type = 'multiple_choice' THEN 1 ELSE 2 END, q.code)::int AS position,
         CASE WHEN q.question_type = 'essay' THEN 5 ELSE 1 END AS points
  FROM package_final pf
  JOIN upserted u ON TRUE
  JOIN cbt_questions q ON q.id = u.id
  ON CONFLICT (package_id, question_id) DO UPDATE SET
    position = EXCLUDED.position,
    points = EXCLUDED.points
  RETURNING package_id, question_id
)
SELECT
  (SELECT COUNT(*) FROM upserted WHERE code LIKE 'SIM-INF-VIII-PG-%') AS pg_count,
  (SELECT COUNT(*) FROM upserted WHERE code LIKE 'SIM-INF-VIII-ES-%') AS essay_count,
  (SELECT id FROM package_final LIMIT 1) AS package_id,
  (SELECT COUNT(*) FROM linked) AS linked_count;

COMMIT;
