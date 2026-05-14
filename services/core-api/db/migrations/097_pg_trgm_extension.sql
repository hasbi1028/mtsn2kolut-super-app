-- Sprint 4 DB lifecycle: enable trigram operator classes for indexed ILIKE search.
-- Keep this extension setup transactional; online index builds live in explicit
-- non-transactional migrations after this file.

CREATE EXTENSION IF NOT EXISTS pg_trgm;
