-- Employee identity refactor and honorer baseline import.
-- Human-facing employee identity is pegawai_uid; employees.id remains the UUID PK.

ALTER TABLE employees
  ADD COLUMN IF NOT EXISTS pegawai_uid TEXT,
  ADD COLUMN IF NOT EXISTS jenis_kelamin TEXT,
  ADD COLUMN IF NOT EXISTS tempat_lahir TEXT;

ALTER TABLE employees
  ALTER COLUMN nip DROP NOT NULL;

ALTER TABLE employees
  DROP CONSTRAINT IF EXISTS employees_nip_key;

DROP INDEX IF EXISTS employees_nip_key;

UPDATE employees
SET nip = NULL
WHERE nip IS NOT NULL
  AND btrim(nip) = '';

UPDATE employees
SET jenis_kelamin = NULL
WHERE jenis_kelamin IS NOT NULL
  AND btrim(jenis_kelamin) = '';

UPDATE employees
SET tempat_lahir = NULL
WHERE tempat_lahir IS NOT NULL
  AND btrim(tempat_lahir) = '';

ALTER TABLE employees
  DROP CONSTRAINT IF EXISTS employees_jenis_kelamin_check;

ALTER TABLE employees
  ADD CONSTRAINT employees_jenis_kelamin_check
  CHECK (jenis_kelamin IS NULL OR jenis_kelamin IN ('L', 'P'));

WITH missing AS (
  SELECT
    id,
    ('40406031' || COALESCE(to_char(tanggal_lahir, 'YY'), '00'))::text AS uid_prefix,
    row_number() OVER (
      PARTITION BY ('40406031' || COALESCE(to_char(tanggal_lahir, 'YY'), '00'))
      ORDER BY created_at ASC, nama ASC, id ASC
    )::int AS sequence_no
  FROM employees
  WHERE pegawai_uid IS NULL
     OR btrim(pegawai_uid) = ''
)
UPDATE employees e
SET pegawai_uid = missing.uid_prefix || lpad(missing.sequence_no::text, 3, '0')
FROM missing
WHERE e.id = missing.id;

ALTER TABLE employees
  ALTER COLUMN pegawai_uid SET NOT NULL;

ALTER TABLE employees
  DROP CONSTRAINT IF EXISTS employees_pegawai_uid_format_check;

ALTER TABLE employees
  ADD CONSTRAINT employees_pegawai_uid_format_check
  CHECK (pegawai_uid ~ '^[0-9]{13}$');

CREATE UNIQUE INDEX IF NOT EXISTS uq_employees_pegawai_uid
  ON employees (pegawai_uid);

CREATE UNIQUE INDEX IF NOT EXISTS uq_employees_nip_not_empty
  ON employees (nip)
  WHERE nip IS NOT NULL AND btrim(nip) <> '';

WITH seed(nama, jenis_kelamin, tempat_lahir, tanggal_lahir, sort_order) AS (
  VALUES
    ('KM.Muh.Tang, S.Ag.', 'L', 'Jerae', DATE '1968-04-24', 1),
    ('Nurilmi, S.Pd', 'P', 'Olo-oloho', DATE '2000-10-03', 2),
    ('Jayanti Jufri, S.Pd', 'P', 'Olo-oloho', DATE '1988-08-15', 3),
    ('Nurunnisaa Alimah A, S.Pd', 'P', 'Majene', DATE '2000-01-03', 4),
    ('Mirnawati, S.Pd', 'P', 'Olo-oloho', DATE '2001-04-14', 5),
    ('Asniar, S.Pd', 'P', 'Olo-oloho', DATE '2000-12-18', 6),
    ('KM. Muhammad Yani, S.Pd', 'L', 'Mallengngeng', DATE '1992-11-14', 7),
    ('Sri Hardianti, S.Or', 'P', 'Mikuasi', DATE '1997-05-06', 8)
),
updated AS (
  UPDATE employees e
  SET nip = NULL,
      employment_type = 'honorer',
      is_active = TRUE,
      jenis_kelamin = seed.jenis_kelamin,
      tempat_lahir = seed.tempat_lahir,
      tanggal_lahir = seed.tanggal_lahir,
      updated_at = NOW()
  FROM seed
  WHERE lower(btrim(e.nama)) = lower(btrim(seed.nama))
    AND e.tanggal_lahir = seed.tanggal_lahir
  RETURNING e.id
),
to_insert AS (
  SELECT seed.*
  FROM seed
  WHERE NOT EXISTS (
    SELECT 1
    FROM employees e
    WHERE lower(btrim(e.nama)) = lower(btrim(seed.nama))
      AND e.tanggal_lahir = seed.tanggal_lahir
  )
),
sequenced AS (
  SELECT
    to_insert.*,
    ('40406031' || to_char(to_insert.tanggal_lahir, 'YY'))::text AS uid_prefix,
    row_number() OVER (
      PARTITION BY ('40406031' || to_char(to_insert.tanggal_lahir, 'YY'))
      ORDER BY to_insert.sort_order ASC, to_insert.nama ASC
    )::int AS row_no
  FROM to_insert
),
prefix_max AS (
  SELECT
    sequenced.uid_prefix,
    COALESCE(max(substring(e.pegawai_uid FROM length(sequenced.uid_prefix) + 1 FOR 3)::int), 0)::int AS max_suffix
  FROM sequenced
  LEFT JOIN employees e
    ON e.pegawai_uid ~ ('^' || sequenced.uid_prefix || '[0-9]{3}$')
  GROUP BY sequenced.uid_prefix
),
inserted AS (
  INSERT INTO employees (
    id,
    pegawai_uid,
    nip,
    nama,
    unit_kerja,
    employment_type,
    tanggal_lahir,
    jenis_kelamin,
    tempat_lahir,
    is_active
  )
  SELECT
    gen_random_uuid(),
    sequenced.uid_prefix || lpad((prefix_max.max_suffix + sequenced.row_no)::text, 3, '0'),
    NULL,
    sequenced.nama,
    '',
    'honorer',
    sequenced.tanggal_lahir,
    sequenced.jenis_kelamin,
    sequenced.tempat_lahir,
    TRUE
  FROM sequenced
  JOIN prefix_max ON prefix_max.uid_prefix = sequenced.uid_prefix
  RETURNING id
)
DELETE FROM pusaka_accounts pa
WHERE pa.employee_id IN (
  SELECT e.id
  FROM employees e
  JOIN seed
    ON lower(btrim(e.nama)) = lower(btrim(seed.nama))
   AND e.tanggal_lahir = seed.tanggal_lahir
);
