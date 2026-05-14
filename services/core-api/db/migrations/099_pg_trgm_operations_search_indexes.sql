-- mtsn2kolut:migration non-transactional
-- Sprint 4 DB lifecycle: second online trigram batch for operational catalogs
-- that can grow over time. Keep lower-cardinality lookups on existing btree
-- indexes until query evidence shows more trigram coverage is needed.

DROP INDEX CONCURRENTLY IF EXISTS idx_library_books_judul_trgm;
CREATE INDEX CONCURRENTLY idx_library_books_judul_trgm
  ON library_books USING gin (judul gin_trgm_ops);

DROP INDEX CONCURRENTLY IF EXISTS idx_library_books_pengarang_trgm;
CREATE INDEX CONCURRENTLY idx_library_books_pengarang_trgm
  ON library_books USING gin (pengarang gin_trgm_ops);

DROP INDEX CONCURRENTLY IF EXISTS idx_library_books_kode_trgm;
CREATE INDEX CONCURRENTLY idx_library_books_kode_trgm
  ON library_books USING gin (kode gin_trgm_ops);

DROP INDEX CONCURRENTLY IF EXISTS idx_archive_documents_title_trgm;
CREATE INDEX CONCURRENTLY idx_archive_documents_title_trgm
  ON archive_documents USING gin (title gin_trgm_ops);

DROP INDEX CONCURRENTLY IF EXISTS idx_archive_documents_number_trgm;
CREATE INDEX CONCURRENTLY idx_archive_documents_number_trgm
  ON archive_documents USING gin (archive_number gin_trgm_ops);

DROP INDEX CONCURRENTLY IF EXISTS idx_archive_documents_original_name_trgm;
CREATE INDEX CONCURRENTLY idx_archive_documents_original_name_trgm
  ON archive_documents USING gin (original_name gin_trgm_ops);

DROP INDEX CONCURRENTLY IF EXISTS idx_governance_documents_title_trgm;
CREATE INDEX CONCURRENTLY idx_governance_documents_title_trgm
  ON governance_documents USING gin (title gin_trgm_ops);

DROP INDEX CONCURRENTLY IF EXISTS idx_governance_programs_name_trgm;
CREATE INDEX CONCURRENTLY idx_governance_programs_name_trgm
  ON governance_programs USING gin (name gin_trgm_ops);

DROP INDEX CONCURRENTLY IF EXISTS idx_governance_evidence_title_trgm;
CREATE INDEX CONCURRENTLY idx_governance_evidence_title_trgm
  ON governance_evidence_items USING gin (title gin_trgm_ops);
