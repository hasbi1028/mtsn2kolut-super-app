-- Backfill academic display codes for CBT questions that still have blank/missing codes.
-- Format: SUBJECTCODE-TARGETLEVEL-TYPECODE-0001, e.g. MTK-VII-PG-0001.
-- This script is idempotent for already-coded rows: it only updates NULL/blank code values.

WITH missing AS (
  SELECT
    q.id,
    UPPER(COALESCE(NULLIF(regexp_replace(btrim(s.code), '[^A-Za-z0-9]+', '', 'g'), ''), 'MAPEL')) AS subject_code,
    UPPER(COALESCE(NULLIF(regexp_replace(btrim(q.target_level), '[^A-Za-z0-9]+', '', 'g'), ''), 'NA')) AS target_level,
    CASE q.question_type::text
      WHEN 'multiple_choice' THEN 'PG'
      WHEN 'multiple_answer' THEN 'PGK'
      WHEN 'true_false' THEN 'TF'
      WHEN 'agree_disagree' THEN 'BS'
      WHEN 'matching' THEN 'JD'
      WHEN 'ordering' THEN 'UR'
      WHEN 'short_answer' THEN 'IS'
      WHEN 'essay' THEN 'ES'
      ELSE 'SOAL'
    END AS type_code,
    q.created_at,
    q.id::text AS id_text
  FROM cbt_questions q
  JOIN subjects s ON s.id = q.subject_id
  WHERE btrim(COALESCE(q.code, '')) = ''
), with_prefix AS (
  SELECT
    id,
    subject_code || '-' || target_level || '-' || type_code AS prefix,
    created_at,
    id_text
  FROM missing
), existing_max AS (
  SELECT
    p.prefix,
    COALESCE(MAX(substring(q.code from '[0-9]{4}$')::int), 0) AS max_number
  FROM (SELECT DISTINCT prefix FROM with_prefix) p
  LEFT JOIN cbt_questions q
    ON q.code ~ ('^' || p.prefix || '-[0-9]{4}$')
  GROUP BY p.prefix
), numbered AS (
  SELECT
    w.id,
    w.prefix || '-' || LPAD((e.max_number + ROW_NUMBER() OVER (PARTITION BY w.prefix ORDER BY w.created_at, w.id_text))::text, 4, '0') AS new_code
  FROM with_prefix w
  JOIN existing_max e ON e.prefix = w.prefix
)
UPDATE cbt_questions q
SET code = numbered.new_code,
    updated_at = NOW()
FROM numbered
WHERE q.id = numbered.id
RETURNING q.id, q.code;
