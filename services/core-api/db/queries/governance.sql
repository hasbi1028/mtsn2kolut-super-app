-- name: GetGovernanceStats :one
SELECT
    (SELECT COUNT(*) FROM governance_units WHERE is_active) AS total_units,
    (SELECT COUNT(*) FROM governance_positions WHERE is_active) AS total_positions,
    (SELECT COUNT(*) FROM governance_assignments WHERE end_date IS NULL) AS active_assignments,
    (SELECT COUNT(*) FROM governance_documents WHERE status = 'final') AS final_documents,
    (SELECT COUNT(*) FROM governance_programs WHERE status IN ('planned', 'in_progress')) AS active_programs,
    (SELECT COUNT(*) FROM governance_programs WHERE status = 'blocked') AS blocked_programs,
    (SELECT COUNT(*) FROM governance_programs WHERE status = 'done') AS completed_programs,
    (SELECT COUNT(*) FROM governance_programs WHERE evidence_url <> '') AS programs_with_evidence,
    (SELECT COUNT(*) FROM governance_performance_targets) AS performance_targets,
    (SELECT COUNT(*) FROM governance_performance_targets WHERE status = 'done') AS completed_performance_targets,
    (SELECT COUNT(*) FROM governance_performance_targets WHERE status = 'blocked') AS blocked_performance_targets,
    (SELECT COUNT(*) FROM governance_evidence_items) AS evidence_items,
    (SELECT COUNT(*) FROM governance_evidence_items WHERE status = 'verified') AS verified_evidence_items,
    (SELECT COUNT(*) FROM governance_evidence_items WHERE status = 'gap') AS gap_evidence_items,
    (SELECT COUNT(*) FROM governance_work_plan_items) AS work_plan_items,
    (SELECT COUNT(*) FROM governance_work_plan_items WHERE status = 'done') AS completed_work_plan_items,
    (SELECT COUNT(*) FROM governance_work_plan_items WHERE status = 'blocked') AS blocked_work_plan_items,
    (SELECT COALESCE(SUM(budget_amount), 0)::BIGINT FROM governance_work_plan_items) AS work_plan_budget_amount,
    (SELECT COALESCE(SUM(realization_amount), 0)::BIGINT FROM governance_work_plan_items) AS work_plan_realization_amount,
    (SELECT COUNT(*) FROM governance_compliance_actions) AS compliance_actions,
    (SELECT COUNT(*) FROM governance_compliance_actions WHERE status NOT IN ('done', 'cancelled')) AS open_compliance_actions,
    (SELECT COUNT(*) FROM governance_compliance_actions WHERE status = 'done') AS completed_compliance_actions,
    (SELECT COUNT(*) FROM governance_compliance_actions WHERE priority IN ('high', 'urgent') AND status NOT IN ('done', 'cancelled')) AS critical_compliance_actions;

-- name: ListGovernanceEmployeeOptions :many
SELECT id, COALESCE(nip, '')::text AS nip, nama, unit_kerja
FROM employees
WHERE is_active = TRUE
ORDER BY nama ASC;

-- name: ListGovernanceUnits :many
SELECT
    u.id, u.code, u.name, u.unit_type, u.parent_id, u.description, u.is_active,
    u.sort_order, u.created_at, u.updated_at,
    pu.name AS parent_unit_name
FROM governance_units u
LEFT JOIN governance_units pu ON pu.id = u.parent_id
WHERE (
    sqlc.arg(search)::TEXT = '' OR
    u.code ILIKE '%' || sqlc.arg(search) || '%' OR
    u.name ILIKE '%' || sqlc.arg(search) || '%' OR
    u.unit_type ILIKE '%' || sqlc.arg(search) || '%'
)
ORDER BY COALESCE(pu.sort_order, -1), u.sort_order, u.name;

-- name: CreateGovernanceUnit :one
INSERT INTO governance_units (code, name, unit_type, parent_id, description, is_active, sort_order)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateGovernanceUnit :one
UPDATE governance_units
SET code = $2,
    name = $3,
    unit_type = $4,
    parent_id = $5,
    description = $6,
    is_active = $7,
    sort_order = $8,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteGovernanceUnit :exec
DELETE FROM governance_units WHERE id = $1;

-- name: ListGovernancePositions :many
SELECT
    p.id, p.unit_id, p.title, p.position_type, p.parent_position_id, p.description,
    p.tupoksi, p.is_active, p.sort_order, p.created_at, p.updated_at,
    u.name AS unit_name,
    pp.title AS parent_position_title,
    ae.nama AS active_employee_name,
    COALESCE(ae.nip, '')::text AS active_employee_nip
FROM governance_positions p
JOIN governance_units u ON u.id = p.unit_id
LEFT JOIN governance_positions pp ON pp.id = p.parent_position_id
LEFT JOIN governance_assignments ga ON ga.position_id = p.id AND ga.end_date IS NULL
LEFT JOIN employees ae ON ae.id = ga.employee_id
WHERE (
    sqlc.arg(search)::TEXT = '' OR
    p.title ILIKE '%' || sqlc.arg(search) || '%' OR
    u.name ILIKE '%' || sqlc.arg(search) || '%' OR
    p.position_type ILIKE '%' || sqlc.arg(search) || '%'
)
ORDER BY u.sort_order, p.sort_order, p.title;

-- name: CreateGovernancePosition :one
INSERT INTO governance_positions (
    unit_id, title, position_type, parent_position_id, description, tupoksi, is_active, sort_order
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateGovernancePosition :one
UPDATE governance_positions
SET unit_id = $2,
    title = $3,
    position_type = $4,
    parent_position_id = $5,
    description = $6,
    tupoksi = $7,
    is_active = $8,
    sort_order = $9,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteGovernancePosition :exec
DELETE FROM governance_positions WHERE id = $1;

-- name: ListGovernanceAssignments :many
SELECT
    ga.id, ga.position_id, ga.employee_id, ga.start_date, ga.end_date,
    ga.decree_outgoing_letter_id, ga.notes, ga.created_at, ga.updated_at,
    p.title AS position_title,
    u.name AS unit_name,
    e.nama AS employee_name,
    COALESCE(e.nip, '')::text AS employee_nip,
    ol.nomor_surat AS decree_nomor_surat
FROM governance_assignments ga
JOIN governance_positions p ON p.id = ga.position_id
JOIN governance_units u ON u.id = p.unit_id
JOIN employees e ON e.id = ga.employee_id
LEFT JOIN outgoing_letters ol ON ol.id = ga.decree_outgoing_letter_id
WHERE (
    sqlc.arg(active_only)::BOOLEAN = FALSE OR ga.end_date IS NULL
) AND (
    sqlc.arg(search)::TEXT = '' OR
    p.title ILIKE '%' || sqlc.arg(search) || '%' OR
    u.name ILIKE '%' || sqlc.arg(search) || '%' OR
    e.nama ILIKE '%' || sqlc.arg(search) || '%' OR
    COALESCE(e.nip, '') ILIKE '%' || sqlc.arg(search) || '%'
)
ORDER BY ga.end_date NULLS FIRST, ga.start_date DESC, u.sort_order, p.sort_order;

-- name: CreateGovernanceAssignment :one
INSERT INTO governance_assignments (
    position_id, employee_id, start_date, end_date, decree_outgoing_letter_id, notes
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateGovernanceAssignment :one
UPDATE governance_assignments
SET position_id = $2,
    employee_id = $3,
    start_date = $4,
    end_date = $5,
    decree_outgoing_letter_id = $6,
    notes = $7,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteGovernanceAssignment :exec
DELETE FROM governance_assignments WHERE id = $1;

-- name: ListGovernanceDocuments :many
SELECT
    gd.id, gd.doc_type, gd.title, gd.period_year, gd.period_label, gd.owner_unit_id,
    gd.snp_standard, gd.status, gd.document_url, gd.outgoing_letter_id, gd.summary,
    gd.created_by_user_id, gd.created_at, gd.updated_at,
    u.name AS owner_unit_name,
    ol.nomor_surat AS outgoing_nomor_surat
FROM governance_documents gd
LEFT JOIN governance_units u ON u.id = gd.owner_unit_id
LEFT JOIN outgoing_letters ol ON ol.id = gd.outgoing_letter_id
WHERE (
    sqlc.arg(search)::TEXT = '' OR
    gd.title ILIKE '%' || sqlc.arg(search) || '%' OR
    gd.doc_type ILIKE '%' || sqlc.arg(search) || '%' OR
    gd.summary ILIKE '%' || sqlc.arg(search) || '%'
) AND (
    sqlc.arg(doc_type)::TEXT = '' OR gd.doc_type = sqlc.arg(doc_type)
) AND (
    sqlc.arg(period_year)::INT = 0 OR gd.period_year = sqlc.arg(period_year)
) AND (
    sqlc.arg(snp_standard)::TEXT = '' OR gd.snp_standard = sqlc.arg(snp_standard)
)
ORDER BY gd.period_year DESC, gd.doc_type, gd.updated_at DESC;

-- name: CreateGovernanceDocument :one
INSERT INTO governance_documents (
    doc_type, title, period_year, period_label, owner_unit_id, snp_standard, status,
    document_url, outgoing_letter_id, summary, created_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: UpdateGovernanceDocument :one
UPDATE governance_documents
SET doc_type = $2,
    title = $3,
    period_year = $4,
    period_label = $5,
    owner_unit_id = $6,
    snp_standard = $7,
    status = $8,
    document_url = $9,
    outgoing_letter_id = $10,
    summary = $11,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteGovernanceDocument :exec
DELETE FROM governance_documents WHERE id = $1;

-- name: ListGovernancePrograms :many
SELECT
    gp.id, gp.period_year, gp.code, gp.name, gp.source_document_id, gp.owner_unit_id,
    gp.responsible_position_id, gp.responsible_employee_id, gp.snp_standard, gp.iku_code,
    gp.indicator, gp.target_value, gp.target_unit, gp.status, gp.progress_percent,
    gp.realization_summary, gp.evidence_url, gp.due_date, gp.created_at, gp.updated_at,
    gd.title AS source_document_title,
    u.name AS owner_unit_name,
    p.title AS responsible_position_title,
    e.nama AS responsible_employee_name
FROM governance_programs gp
LEFT JOIN governance_documents gd ON gd.id = gp.source_document_id
LEFT JOIN governance_units u ON u.id = gp.owner_unit_id
LEFT JOIN governance_positions p ON p.id = gp.responsible_position_id
LEFT JOIN employees e ON e.id = gp.responsible_employee_id
WHERE (
    sqlc.arg(search)::TEXT = '' OR
    gp.code ILIKE '%' || sqlc.arg(search) || '%' OR
    gp.name ILIKE '%' || sqlc.arg(search) || '%' OR
    gp.indicator ILIKE '%' || sqlc.arg(search) || '%' OR
    gp.iku_code ILIKE '%' || sqlc.arg(search) || '%'
) AND (
    sqlc.arg(period_year)::INT = 0 OR gp.period_year = sqlc.arg(period_year)
) AND (
    sqlc.arg(status)::TEXT = '' OR gp.status = sqlc.arg(status)
) AND (
    sqlc.arg(snp_standard)::TEXT = '' OR gp.snp_standard = sqlc.arg(snp_standard)
)
ORDER BY gp.period_year DESC, gp.code;

-- name: CreateGovernanceProgram :one
INSERT INTO governance_programs (
    period_year, code, name, source_document_id, owner_unit_id, responsible_position_id,
    responsible_employee_id, snp_standard, iku_code, indicator, target_value, target_unit,
    status, progress_percent, realization_summary, evidence_url, due_date
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
RETURNING *;

-- name: UpdateGovernanceProgram :one
UPDATE governance_programs
SET period_year = $2,
    code = $3,
    name = $4,
    source_document_id = $5,
    owner_unit_id = $6,
    responsible_position_id = $7,
    responsible_employee_id = $8,
    snp_standard = $9,
    iku_code = $10,
    indicator = $11,
    target_value = $12,
    target_unit = $13,
    status = $14,
    progress_percent = $15,
    realization_summary = $16,
    evidence_url = $17,
    due_date = $18,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteGovernanceProgram :exec
DELETE FROM governance_programs WHERE id = $1;

-- name: ListGovernanceWorkPlanItems :many
SELECT
    gwpi.id,
    gwpi.period_year,
    gwpi.program_id,
    gwpi.source_document_id,
    gwpi.owner_unit_id,
    gwpi.responsible_employee_id,
    gwpi.evidence_item_id,
    gwpi.activity_code,
    gwpi.activity_name,
    gwpi.output_indicator,
    gwpi.target_volume,
    gwpi.target_unit,
    gwpi.budget_source,
    gwpi.budget_amount,
    gwpi.realization_amount,
    gwpi.status,
    gwpi.progress_percent,
    gwpi.start_date,
    gwpi.end_date,
    gwpi.evidence_url,
    gwpi.notes,
    gwpi.created_at,
    gwpi.updated_at,
    gp.code AS program_code,
    gp.name AS program_name,
    gp.snp_standard AS program_snp_standard,
    gd.title AS source_document_title,
    u.name AS owner_unit_name,
    e.nama AS responsible_employee_name,
    COALESCE(e.nip, '')::text AS responsible_employee_nip,
    gei.title AS evidence_item_title
FROM governance_work_plan_items gwpi
JOIN governance_programs gp ON gp.id = gwpi.program_id
LEFT JOIN governance_documents gd ON gd.id = gwpi.source_document_id
LEFT JOIN governance_units u ON u.id = gwpi.owner_unit_id
LEFT JOIN employees e ON e.id = gwpi.responsible_employee_id
LEFT JOIN governance_evidence_items gei ON gei.id = gwpi.evidence_item_id
WHERE (
    sqlc.arg(search)::TEXT = '' OR
    gwpi.activity_code ILIKE '%' || sqlc.arg(search) || '%' OR
    gwpi.activity_name ILIKE '%' || sqlc.arg(search) || '%' OR
    gwpi.output_indicator ILIKE '%' || sqlc.arg(search) || '%' OR
    gwpi.budget_source ILIKE '%' || sqlc.arg(search) || '%' OR
    gp.code ILIKE '%' || sqlc.arg(search) || '%' OR
    gp.name ILIKE '%' || sqlc.arg(search) || '%' OR
    u.name ILIKE '%' || sqlc.arg(search) || '%' OR
    e.nama ILIKE '%' || sqlc.arg(search) || '%'
) AND (
    sqlc.arg(period_year)::INT = 0 OR gwpi.period_year = sqlc.arg(period_year)
) AND (
    sqlc.arg(program_id)::UUID IS NULL OR gwpi.program_id = sqlc.arg(program_id)::UUID
) AND (
    sqlc.arg(status)::TEXT = '' OR gwpi.status = sqlc.arg(status)
)
ORDER BY gwpi.period_year DESC, gp.code, gwpi.activity_code;

-- name: CreateGovernanceWorkPlanItem :one
INSERT INTO governance_work_plan_items (
    period_year, program_id, source_document_id, owner_unit_id, responsible_employee_id,
    evidence_item_id, activity_code, activity_name, output_indicator, target_volume,
    target_unit, budget_source, budget_amount, realization_amount, status,
    progress_percent, start_date, end_date, evidence_url, notes
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
)
RETURNING *;

-- name: UpdateGovernanceWorkPlanItem :one
UPDATE governance_work_plan_items
SET period_year = $2,
    program_id = $3,
    source_document_id = $4,
    owner_unit_id = $5,
    responsible_employee_id = $6,
    evidence_item_id = $7,
    activity_code = $8,
    activity_name = $9,
    output_indicator = $10,
    target_volume = $11,
    target_unit = $12,
    budget_source = $13,
    budget_amount = $14,
    realization_amount = $15,
    status = $16,
    progress_percent = $17,
    start_date = $18,
    end_date = $19,
    evidence_url = $20,
    notes = $21,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteGovernanceWorkPlanItem :exec
DELETE FROM governance_work_plan_items WHERE id = $1;

-- name: ListGovernancePerformanceTargets :many
SELECT
    gpt.id, gpt.period_year, gpt.employee_id, gpt.position_id, gpt.program_id,
    gpt.parent_target_id, gpt.aspect, gpt.title, gpt.indicator, gpt.target_value,
    gpt.target_unit, gpt.status, gpt.progress_percent, gpt.evidence_url,
    gpt.review_notes, gpt.due_date, gpt.created_by_user_id, gpt.created_at, gpt.updated_at,
    e.nama AS employee_name,
    COALESCE(e.nip, '')::text AS employee_nip,
    p.title AS position_title,
    gp.code AS program_code,
    gp.name AS program_name,
    parent.title AS parent_target_title
FROM governance_performance_targets gpt
JOIN employees e ON e.id = gpt.employee_id
LEFT JOIN governance_positions p ON p.id = gpt.position_id
LEFT JOIN governance_programs gp ON gp.id = gpt.program_id
LEFT JOIN governance_performance_targets parent ON parent.id = gpt.parent_target_id
WHERE (
    sqlc.arg(search)::TEXT = '' OR
    gpt.title ILIKE '%' || sqlc.arg(search) || '%' OR
    gpt.indicator ILIKE '%' || sqlc.arg(search) || '%' OR
    e.nama ILIKE '%' || sqlc.arg(search) || '%' OR
    COALESCE(e.nip, '') ILIKE '%' || sqlc.arg(search) || '%' OR
    gp.code ILIKE '%' || sqlc.arg(search) || '%' OR
    gp.name ILIKE '%' || sqlc.arg(search) || '%'
) AND (
    sqlc.arg(period_year)::INT = 0 OR gpt.period_year = sqlc.arg(period_year)
) AND (
    sqlc.arg(status)::TEXT = '' OR gpt.status = sqlc.arg(status)
) AND (
    sqlc.arg(employee_id)::UUID IS NULL OR gpt.employee_id = sqlc.arg(employee_id)::UUID
)
ORDER BY gpt.period_year DESC, e.nama ASC, gpt.created_at DESC;

-- name: CreateGovernancePerformanceTarget :one
INSERT INTO governance_performance_targets (
    period_year, employee_id, position_id, program_id, parent_target_id, aspect,
    title, indicator, target_value, target_unit, status, progress_percent,
    evidence_url, review_notes, due_date, created_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
RETURNING *;

-- name: UpdateGovernancePerformanceTarget :one
UPDATE governance_performance_targets
SET period_year = $2,
    employee_id = $3,
    position_id = $4,
    program_id = $5,
    parent_target_id = $6,
    aspect = $7,
    title = $8,
    indicator = $9,
    target_value = $10,
    target_unit = $11,
    status = $12,
    progress_percent = $13,
    evidence_url = $14,
    review_notes = $15,
    due_date = $16,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteGovernancePerformanceTarget :exec
DELETE FROM governance_performance_targets WHERE id = $1;

-- name: ListGovernanceEvidenceItems :many
SELECT
    gei.id, gei.period_year, gei.title, gei.evidence_type, gei.snp_standard,
    gei.owner_unit_id, gei.document_id, gei.program_id, gei.performance_target_id,
    gei.source_module, gei.evidence_url, gei.status, gei.notes, gei.verified_by_user_id,
    gei.verified_at, gei.created_by_user_id, gei.created_at, gei.updated_at,
    u.name AS owner_unit_name,
    gd.title AS document_title,
    gp.code AS program_code,
    gp.name AS program_name,
    gpt.title AS performance_target_title
FROM governance_evidence_items gei
LEFT JOIN governance_units u ON u.id = gei.owner_unit_id
LEFT JOIN governance_documents gd ON gd.id = gei.document_id
LEFT JOIN governance_programs gp ON gp.id = gei.program_id
LEFT JOIN governance_performance_targets gpt ON gpt.id = gei.performance_target_id
WHERE (
    sqlc.arg(search)::TEXT = '' OR
    gei.title ILIKE '%' || sqlc.arg(search) || '%' OR
    gei.evidence_type ILIKE '%' || sqlc.arg(search) || '%' OR
    gei.source_module ILIKE '%' || sqlc.arg(search) || '%' OR
    gei.notes ILIKE '%' || sqlc.arg(search) || '%' OR
    u.name ILIKE '%' || sqlc.arg(search) || '%' OR
    gd.title ILIKE '%' || sqlc.arg(search) || '%' OR
    gp.code ILIKE '%' || sqlc.arg(search) || '%' OR
    gp.name ILIKE '%' || sqlc.arg(search) || '%' OR
    gpt.title ILIKE '%' || sqlc.arg(search) || '%'
) AND (
    sqlc.arg(period_year)::INT = 0 OR gei.period_year = sqlc.arg(period_year)
) AND (
    sqlc.arg(status)::TEXT = '' OR gei.status = sqlc.arg(status)
) AND (
    sqlc.arg(snp_standard)::TEXT = '' OR gei.snp_standard = sqlc.arg(snp_standard)
)
ORDER BY gei.period_year DESC, gei.status, gei.updated_at DESC;

-- name: CreateGovernanceEvidenceItem :one
INSERT INTO governance_evidence_items (
    period_year, title, evidence_type, snp_standard, owner_unit_id, document_id,
    program_id, performance_target_id, source_module, evidence_url, status, notes,
    created_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: UpdateGovernanceEvidenceItem :one
UPDATE governance_evidence_items
SET period_year = $2,
    title = $3,
    evidence_type = $4,
    snp_standard = $5,
    owner_unit_id = $6,
    document_id = $7,
    program_id = $8,
    performance_target_id = $9,
    source_module = $10,
    evidence_url = $11,
    status = $12,
    notes = $13,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteGovernanceEvidenceItem :exec
DELETE FROM governance_evidence_items WHERE id = $1;

-- name: ListGovernanceComplianceActions :many
SELECT
    gca.id,
    gca.period_year,
    gca.source_type,
    gca.source_ref_id,
    gca.snp_standard,
    gca.program_id,
    gca.document_id,
    gca.performance_target_id,
    gca.evidence_item_id,
    gca.owner_unit_id,
    gca.responsible_employee_id,
    gca.title,
    gca.description,
    gca.priority,
    gca.status,
    gca.due_date,
    gca.completed_at,
    gca.follow_up_notes,
    gca.evidence_url,
    gca.created_by_user_id,
    gca.created_at,
    gca.updated_at,
    gp.code AS program_code,
    gp.name AS program_name,
    gd.title AS document_title,
    gpt.title AS performance_target_title,
    gei.title AS evidence_item_title,
    u.name AS owner_unit_name,
    e.nama AS responsible_employee_name,
    COALESCE(e.nip, '')::text AS responsible_employee_nip
FROM governance_compliance_actions gca
LEFT JOIN governance_programs gp ON gp.id = gca.program_id
LEFT JOIN governance_documents gd ON gd.id = gca.document_id
LEFT JOIN governance_performance_targets gpt ON gpt.id = gca.performance_target_id
LEFT JOIN governance_evidence_items gei ON gei.id = gca.evidence_item_id
LEFT JOIN governance_units u ON u.id = gca.owner_unit_id
LEFT JOIN employees e ON e.id = gca.responsible_employee_id
WHERE (
    sqlc.arg(search)::TEXT = '' OR
    gca.title ILIKE '%' || sqlc.arg(search) || '%' OR
    gca.description ILIKE '%' || sqlc.arg(search) || '%' OR
    gca.follow_up_notes ILIKE '%' || sqlc.arg(search) || '%' OR
    gp.code ILIKE '%' || sqlc.arg(search) || '%' OR
    gp.name ILIKE '%' || sqlc.arg(search) || '%' OR
    gd.title ILIKE '%' || sqlc.arg(search) || '%' OR
    gpt.title ILIKE '%' || sqlc.arg(search) || '%' OR
    gei.title ILIKE '%' || sqlc.arg(search) || '%' OR
    u.name ILIKE '%' || sqlc.arg(search) || '%' OR
    e.nama ILIKE '%' || sqlc.arg(search) || '%'
) AND (
    sqlc.arg(period_year)::INT = 0 OR gca.period_year = sqlc.arg(period_year)
) AND (
    sqlc.arg(status)::TEXT = '' OR gca.status = sqlc.arg(status)
) AND (
    sqlc.arg(priority)::TEXT = '' OR gca.priority = sqlc.arg(priority)
) AND (
    sqlc.arg(source_type)::TEXT = '' OR gca.source_type = sqlc.arg(source_type)
) AND (
    sqlc.arg(snp_standard)::TEXT = '' OR gca.snp_standard = sqlc.arg(snp_standard)
) AND (
    sqlc.arg(responsible_employee_id)::UUID IS NULL OR gca.responsible_employee_id = sqlc.arg(responsible_employee_id)::UUID
)
ORDER BY
    gca.period_year DESC,
    CASE gca.status
        WHEN 'open' THEN 1
        WHEN 'in_progress' THEN 2
        WHEN 'waiting_evidence' THEN 3
        WHEN 'done' THEN 4
        ELSE 5
    END,
    CASE gca.priority
        WHEN 'urgent' THEN 1
        WHEN 'high' THEN 2
        WHEN 'medium' THEN 3
        ELSE 4
    END,
    gca.due_date NULLS LAST,
    gca.updated_at DESC;

-- name: CreateGovernanceComplianceAction :one
INSERT INTO governance_compliance_actions (
    period_year, source_type, source_ref_id, snp_standard, program_id, document_id,
    performance_target_id, evidence_item_id, owner_unit_id, responsible_employee_id,
    title, description, priority, status, due_date, completed_at, follow_up_notes,
    evidence_url, created_by_user_id
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19
)
RETURNING *;

-- name: UpdateGovernanceComplianceAction :one
UPDATE governance_compliance_actions
SET period_year = $2,
    source_type = $3,
    source_ref_id = $4,
    snp_standard = $5,
    program_id = $6,
    document_id = $7,
    performance_target_id = $8,
    evidence_item_id = $9,
    owner_unit_id = $10,
    responsible_employee_id = $11,
    title = $12,
    description = $13,
    priority = $14,
    status = $15,
    due_date = $16,
    completed_at = $17,
    follow_up_notes = $18,
    evidence_url = $19,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteGovernanceComplianceAction :exec
DELETE FROM governance_compliance_actions WHERE id = $1;

-- name: ListGovernanceSNPMatrix :many
WITH standards AS (
    SELECT 'skl'::TEXT AS code, 'Standar Kompetensi Lulusan'::TEXT AS name
    UNION ALL SELECT 'isi', 'Standar Isi'
    UNION ALL SELECT 'proses', 'Standar Proses'
    UNION ALL SELECT 'penilaian', 'Standar Penilaian'
    UNION ALL SELECT 'ptk', 'Standar Pendidik dan Tenaga Kependidikan'
    UNION ALL SELECT 'sarpras', 'Standar Sarana dan Prasarana'
    UNION ALL SELECT 'pengelolaan', 'Standar Pengelolaan'
    UNION ALL SELECT 'pembiayaan', 'Standar Pembiayaan'
),
doc_counts AS (
    SELECT snp_standard, COUNT(*) AS total_documents
    FROM governance_documents
    WHERE snp_standard <> ''
    GROUP BY snp_standard
),
program_counts AS (
    SELECT
        snp_standard,
        COUNT(*) AS total_programs,
        COUNT(*) FILTER (WHERE status = 'done') AS completed_programs,
        COUNT(*) FILTER (WHERE evidence_url <> '') AS evidence_programs
    FROM governance_programs
    WHERE snp_standard <> ''
    GROUP BY snp_standard
),
evidence_counts AS (
    SELECT
        snp_standard,
        COUNT(*) AS total_evidence_items,
        COUNT(*) FILTER (WHERE status = 'verified') AS verified_evidence_items,
        COUNT(*) FILTER (WHERE status = 'gap') AS gap_evidence_items
    FROM governance_evidence_items
    WHERE snp_standard <> ''
    GROUP BY snp_standard
),
last_updates AS (
    SELECT snp_standard, MAX(updated_at) AS last_updated_at
    FROM (
        SELECT snp_standard, updated_at FROM governance_documents WHERE snp_standard <> ''
        UNION ALL
        SELECT snp_standard, updated_at FROM governance_programs WHERE snp_standard <> ''
        UNION ALL
        SELECT snp_standard, updated_at FROM governance_evidence_items WHERE snp_standard <> ''
    ) items
    GROUP BY snp_standard
)
SELECT
    s.code,
    s.name,
    COALESCE(dc.total_documents, 0)::BIGINT AS total_documents,
    COALESCE(pc.total_programs, 0)::BIGINT AS total_programs,
    COALESCE(pc.completed_programs, 0)::BIGINT AS completed_programs,
    COALESCE(pc.evidence_programs, 0)::BIGINT AS evidence_programs,
    COALESCE(ec.total_evidence_items, 0)::BIGINT AS total_evidence_items,
    COALESCE(ec.verified_evidence_items, 0)::BIGINT AS verified_evidence_items,
    COALESCE(ec.gap_evidence_items, 0)::BIGINT AS gap_evidence_items,
    lu.last_updated_at::TIMESTAMPTZ AS last_updated_at
FROM standards s
LEFT JOIN doc_counts dc ON dc.snp_standard = s.code
LEFT JOIN program_counts pc ON pc.snp_standard = s.code
LEFT JOIN evidence_counts ec ON ec.snp_standard = s.code
LEFT JOIN last_updates lu ON lu.snp_standard = s.code
ORDER BY CASE s.code
    WHEN 'skl' THEN 1
    WHEN 'isi' THEN 2
    WHEN 'proses' THEN 3
    WHEN 'penilaian' THEN 4
    WHEN 'ptk' THEN 5
    WHEN 'sarpras' THEN 6
    WHEN 'pengelolaan' THEN 7
    WHEN 'pembiayaan' THEN 8
    ELSE 99
END;
