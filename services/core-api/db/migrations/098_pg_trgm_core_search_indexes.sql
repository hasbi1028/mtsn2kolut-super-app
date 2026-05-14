-- mtsn2kolut:migration non-transactional
-- Sprint 4 DB lifecycle: online trigram indexes for the highest-cardinality
-- user-facing search surfaces. These statements must run outside BEGIN because
-- PostgreSQL forbids CREATE/DROP INDEX CONCURRENTLY inside a transaction block.

DROP INDEX CONCURRENTLY IF EXISTS idx_cbt_questions_code_trgm;
CREATE INDEX CONCURRENTLY idx_cbt_questions_code_trgm
  ON cbt_questions USING gin (code gin_trgm_ops);

DROP INDEX CONCURRENTLY IF EXISTS idx_cbt_questions_question_text_trgm;
CREATE INDEX CONCURRENTLY idx_cbt_questions_question_text_trgm
  ON cbt_questions USING gin (question_text gin_trgm_ops);

DROP INDEX CONCURRENTLY IF EXISTS idx_cbt_questions_material_topic_trgm;
CREATE INDEX CONCURRENTLY idx_cbt_questions_material_topic_trgm
  ON cbt_questions USING gin (material_topic gin_trgm_ops);

DROP INDEX CONCURRENTLY IF EXISTS idx_cbt_questions_cp_ref_trgm;
CREATE INDEX CONCURRENTLY idx_cbt_questions_cp_ref_trgm
  ON cbt_questions USING gin (cp_ref gin_trgm_ops);

DROP INDEX CONCURRENTLY IF EXISTS idx_cbt_questions_kd_ref_trgm;
CREATE INDEX CONCURRENTLY idx_cbt_questions_kd_ref_trgm
  ON cbt_questions USING gin (kd_ref gin_trgm_ops);

DROP INDEX CONCURRENTLY IF EXISTS idx_students_nis_trgm;
CREATE INDEX CONCURRENTLY idx_students_nis_trgm
  ON students USING gin (nis gin_trgm_ops);

DROP INDEX CONCURRENTLY IF EXISTS idx_students_nisn_trgm;
CREATE INDEX CONCURRENTLY idx_students_nisn_trgm
  ON students USING gin (nisn gin_trgm_ops);

DROP INDEX CONCURRENTLY IF EXISTS idx_students_nama_trgm;
CREATE INDEX CONCURRENTLY idx_students_nama_trgm
  ON students USING gin (nama gin_trgm_ops);

DROP INDEX CONCURRENTLY IF EXISTS idx_students_nik_trgm;
CREATE INDEX CONCURRENTLY idx_students_nik_trgm
  ON students USING gin (nik gin_trgm_ops);

DROP INDEX CONCURRENTLY IF EXISTS idx_employees_nama_trgm;
CREATE INDEX CONCURRENTLY idx_employees_nama_trgm
  ON employees USING gin (nama gin_trgm_ops);

DROP INDEX CONCURRENTLY IF EXISTS idx_employees_nip_trgm;
CREATE INDEX CONCURRENTLY idx_employees_nip_trgm
  ON employees USING gin ((COALESCE(nip, '')) gin_trgm_ops);

DROP INDEX CONCURRENTLY IF EXISTS idx_employees_pegawai_uid_trgm;
CREATE INDEX CONCURRENTLY idx_employees_pegawai_uid_trgm
  ON employees USING gin ((COALESCE(pegawai_uid, '')) gin_trgm_ops);
