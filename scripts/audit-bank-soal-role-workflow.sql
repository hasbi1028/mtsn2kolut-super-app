-- Sprint 0 Bank Soal role workflow production audit.
-- Read-only report queries. Do not run with credentials echoed.
\set ON_ERROR_STOP on
\pset pager off
\pset null '(null)'
\timing off

\echo '## 1. Question workflow/status distribution'
SELECT
  q.workflow_status,
  q.status,
  COUNT(*) AS total,
  COUNT(*) FILTER (WHERE COALESCE(q.author_username, '') = '') AS missing_author_username,
  COUNT(*) FILTER (WHERE COALESCE(q.reviewer_username, '') = '') AS missing_reviewer_username,
  COUNT(*) FILTER (WHERE COALESCE(q.approver_username, '') = '') AS missing_approver_username,
  COUNT(*) FILTER (WHERE q.reviewed_at IS NULL) AS missing_reviewed_at,
  COUNT(*) FILTER (WHERE q.approved_at IS NULL) AS missing_approved_at
FROM cbt_questions q
GROUP BY q.workflow_status, q.status
ORDER BY q.workflow_status, q.status;

\echo '## 2. Question workflow statuses outside target compatibility set'
SELECT
  q.workflow_status,
  COUNT(*) AS total
FROM cbt_questions q
WHERE q.workflow_status NOT IN (
  'draft', 'review', 'submitted', 'revision_needed', 'reviewed', 'approved', 'published', 'rejected', 'archived'
)
GROUP BY q.workflow_status
ORDER BY q.workflow_status;

\echo '## 3. Package questions by question workflow/status'
SELECT
  q.workflow_status,
  q.status,
  COUNT(*) AS total_package_question_links,
  COUNT(DISTINCT pq.package_id) AS affected_packages,
  COUNT(DISTINCT pq.question_id) AS unique_questions
FROM cbt_package_questions pq
JOIN cbt_questions q ON q.id = pq.question_id
GROUP BY q.workflow_status, q.status
ORDER BY q.workflow_status, q.status;

\echo '## 4. Package questions that are not approved/published-compatible'
SELECT
  q.workflow_status,
  q.status,
  COUNT(*) AS total_package_question_links,
  COUNT(DISTINCT pq.package_id) AS affected_packages,
  COUNT(DISTINCT pq.question_id) AS unique_questions
FROM cbt_package_questions pq
JOIN cbt_questions q ON q.id = pq.question_id
WHERE NOT (
  q.workflow_status IN ('approved', 'published')
  OR q.status = 'published'
)
GROUP BY q.workflow_status, q.status
ORDER BY q.workflow_status, q.status;

\echo '## 5. Bank Soal RBAC role permissions'
SELECT
  r.code AS role_code,
  p.code AS permission_code,
  COUNT(*) AS grants
FROM rbac_role_permissions rp
JOIN rbac_roles r ON r.id = rp.role_id
JOIN rbac_permissions p ON p.id = rp.permission_id
WHERE p.code LIKE 'bank_soal.%'
GROUP BY r.code, p.code
ORDER BY r.code, p.code;

\echo '## 6. Bank Soal user counts by dynamic RBAC role and permission'
SELECT
  r.code AS role_code,
  p.code AS permission_code,
  COUNT(DISTINCT ur.user_id) AS users_with_permission
FROM rbac_user_roles ur
JOIN rbac_roles r ON r.id = ur.role_id
JOIN rbac_role_permissions rp ON rp.role_id = r.id
JOIN rbac_permissions p ON p.id = rp.permission_id
WHERE p.code LIKE 'bank_soal.%'
GROUP BY r.code, p.code
ORDER BY r.code, p.code;

\echo '## 7. User account role distribution'
SELECT
  uar.role::text AS account_role,
  COUNT(*) AS total_users,
  COUNT(*) FILTER (WHERE u.employee_id IS NOT NULL) AS linked_to_employee,
  COUNT(*) FILTER (WHERE u.employee_id IS NULL) AS not_linked_to_employee
FROM user_account_roles uar
JOIN users u ON u.id = uar.user_id
GROUP BY uar.role::text
ORDER BY uar.role::text;

\echo '## 8. Teacher assignment reviewer candidates by subject/level'
SELECT
  s.code AS subject_code,
  s.name AS subject_name,
  sc.level,
  COUNT(DISTINCT csa.teacher_employee_id) AS teacher_count,
  STRING_AGG(DISTINCT e.nama, ', ' ORDER BY e.nama) AS teacher_names
FROM class_subject_assignments csa
JOIN school_classes sc ON sc.id = csa.class_id
JOIN subjects s ON s.id = csa.subject_id
JOIN employees e ON e.id = csa.teacher_employee_id
WHERE COALESCE(sc.is_active, TRUE) = TRUE
GROUP BY s.code, s.name, sc.level
ORDER BY s.name, sc.level;

\echo '## 9. Teacher assignment candidates with user linkage'
SELECT
  s.code AS subject_code,
  s.name AS subject_name,
  sc.level,
  e.nama AS teacher_name,
  COALESCE(u.username, '') AS username,
  CASE WHEN u.id IS NULL THEN 'missing_user_link' ELSE 'linked' END AS user_link_status,
  COUNT(DISTINCT sc.id) AS class_count
FROM class_subject_assignments csa
JOIN school_classes sc ON sc.id = csa.class_id
JOIN subjects s ON s.id = csa.subject_id
JOIN employees e ON e.id = csa.teacher_employee_id
LEFT JOIN users u ON u.employee_id = e.id
WHERE COALESCE(sc.is_active, TRUE) = TRUE
GROUP BY s.code, s.name, sc.level, e.nama, u.username, u.id
ORDER BY s.name, sc.level, e.nama;

\echo '## 10. Questions by author username'
SELECT
  COALESCE(NULLIF(q.author_username, ''), '(blank)') AS author_username,
  q.workflow_status,
  q.status,
  COUNT(*) AS total
FROM cbt_questions q
GROUP BY COALESCE(NULLIF(q.author_username, ''), '(blank)'), q.workflow_status, q.status
ORDER BY total DESC, author_username, q.workflow_status, q.status
LIMIT 100;

\echo '## 11. Existing workflow constraints'
SELECT
  conname AS constraint_name,
  pg_get_constraintdef(oid) AS constraint_definition
FROM pg_constraint
WHERE conrelid = 'cbt_questions'::regclass
  AND conname ILIKE '%workflow%'
ORDER BY conname;

\echo '## 12. Existing Bank Soal/versioning related tables'
SELECT
  table_name
FROM information_schema.tables
WHERE table_schema = 'public'
  AND (
    table_name LIKE 'bank_soal%'
    OR table_name LIKE 'cbt_question%'
  )
ORDER BY table_name;
