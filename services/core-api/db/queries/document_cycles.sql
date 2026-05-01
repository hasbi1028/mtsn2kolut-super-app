-- name: GetDocumentCycleStats :one
SELECT
    (SELECT COUNT(*) FROM document_cycle_catalogs WHERE is_active)::BIGINT AS active_catalogs,
    COUNT(*)::BIGINT AS total_obligations,
    COUNT(*) FILTER (WHERE status = 'not_started')::BIGINT AS not_started_obligations,
    COUNT(*) FILTER (WHERE status = 'draft')::BIGINT AS draft_obligations,
    COUNT(*) FILTER (WHERE status = 'waiting_verification')::BIGINT AS waiting_verification_obligations,
    COUNT(*) FILTER (WHERE status = 'completed')::BIGINT AS completed_obligations,
    COUNT(*) FILTER (WHERE status <> 'completed' AND due_date < CURRENT_DATE)::BIGINT AS overdue_obligations,
    COUNT(*) FILTER (
        WHERE status <> 'completed'
          AND reminder_date <= CURRENT_DATE
          AND due_date >= CURRENT_DATE
    )::BIGINT AS due_soon_obligations,
    COUNT(*) FILTER (WHERE responsible_employee_id IS NULL)::BIGINT AS no_pic_obligations,
    COUNT(*) FILTER (WHERE archive_document_id IS NOT NULL)::BIGINT AS linked_archive_obligations,
    COUNT(*) FILTER (WHERE evidence_item_id IS NOT NULL)::BIGINT AS linked_evidence_obligations,
    COUNT(*) FILTER (WHERE governance_document_id IS NOT NULL)::BIGINT AS linked_governance_document_obligations,
    COUNT(*) FILTER (WHERE compliance_action_id IS NOT NULL)::BIGINT AS linked_compliance_action_obligations,
    COUNT(*) FILTER (WHERE external_system <> '')::BIGINT AS external_tracker_obligations
FROM document_cycle_obligations
WHERE sqlc.arg(period_year)::INT = 0 OR period_year = sqlc.arg(period_year)::INT;

-- name: ListDocumentCycleCatalogs :many
SELECT
    c.id,
    c.code,
    c.title,
    c.frequency,
    c.domain_area,
    c.external_system,
    c.snp_standard,
    c.regulation_ref,
    c.default_owner_unit_id,
    COALESCE(ou.name, '')::TEXT AS default_owner_unit_name,
    c.default_responsible_employee_id,
    COALESCE(re.nama, '')::TEXT AS default_responsible_employee_name,
    COALESCE(re.nip, '')::TEXT AS default_responsible_employee_nip,
    c.default_verifier_employee_id,
    COALESCE(ve.nama, '')::TEXT AS default_verifier_employee_name,
    COALESCE(ve.nip, '')::TEXT AS default_verifier_employee_nip,
    c.deadline_days_after_period,
    c.reminder_days_before_due,
    c.description,
    c.is_active,
    c.sort_order,
    c.created_at,
    c.updated_at
FROM document_cycle_catalogs c
LEFT JOIN governance_units ou ON ou.id = c.default_owner_unit_id
LEFT JOIN employees re ON re.id = c.default_responsible_employee_id
LEFT JOIN employees ve ON ve.id = c.default_verifier_employee_id
WHERE (
    sqlc.arg(search)::TEXT = '' OR
    c.code ILIKE '%' || sqlc.arg(search) || '%' OR
    c.title ILIKE '%' || sqlc.arg(search) || '%' OR
    c.description ILIKE '%' || sqlc.arg(search) || '%' OR
    c.regulation_ref ILIKE '%' || sqlc.arg(search) || '%'
) AND (
    sqlc.arg(frequency)::TEXT = '' OR c.frequency = sqlc.arg(frequency)
) AND (
    sqlc.arg(active_only)::BOOLEAN = FALSE OR c.is_active = TRUE
)
ORDER BY c.is_active DESC, c.sort_order, c.code;

-- name: CreateDocumentCycleCatalog :one
INSERT INTO document_cycle_catalogs (
    code,
    title,
    frequency,
    domain_area,
    external_system,
    snp_standard,
    regulation_ref,
    default_owner_unit_id,
    default_responsible_employee_id,
    default_verifier_employee_id,
    deadline_days_after_period,
    reminder_days_before_due,
    description,
    is_active,
    sort_order
) VALUES (
    sqlc.arg(code),
    sqlc.arg(title),
    sqlc.arg(frequency),
    sqlc.arg(domain_area),
    sqlc.arg(external_system),
    sqlc.arg(snp_standard),
    sqlc.arg(regulation_ref),
    sqlc.arg(default_owner_unit_id),
    sqlc.arg(default_responsible_employee_id),
    sqlc.arg(default_verifier_employee_id),
    sqlc.arg(deadline_days_after_period),
    sqlc.arg(reminder_days_before_due),
    sqlc.arg(description),
    sqlc.arg(is_active),
    sqlc.arg(sort_order)
)
RETURNING *;

-- name: UpdateDocumentCycleCatalog :one
UPDATE document_cycle_catalogs
SET code = sqlc.arg(code),
    title = sqlc.arg(title),
    frequency = sqlc.arg(frequency),
    domain_area = sqlc.arg(domain_area),
    external_system = sqlc.arg(external_system),
    snp_standard = sqlc.arg(snp_standard),
    regulation_ref = sqlc.arg(regulation_ref),
    default_owner_unit_id = sqlc.arg(default_owner_unit_id),
    default_responsible_employee_id = sqlc.arg(default_responsible_employee_id),
    default_verifier_employee_id = sqlc.arg(default_verifier_employee_id),
    deadline_days_after_period = sqlc.arg(deadline_days_after_period),
    reminder_days_before_due = sqlc.arg(reminder_days_before_due),
    description = sqlc.arg(description),
    is_active = sqlc.arg(is_active),
    sort_order = sqlc.arg(sort_order),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteDocumentCycleCatalog :exec
DELETE FROM document_cycle_catalogs WHERE id = $1;

-- name: ListDocumentCycleObligations :many
SELECT
    o.id,
    o.catalog_id,
    c.code AS catalog_code,
    c.title AS catalog_title,
    c.frequency,
    o.domain_area,
    o.external_system,
    c.snp_standard,
    c.regulation_ref,
    o.period_year,
    o.period_label,
    o.period_start,
    o.period_end,
    o.due_date,
    o.reminder_date,
    o.owner_unit_id,
    COALESCE(ou.name, '')::TEXT AS owner_unit_name,
    o.responsible_employee_id,
    COALESCE(re.nama, '')::TEXT AS responsible_employee_name,
    COALESCE(re.nip, '')::TEXT AS responsible_employee_nip,
    o.verifier_employee_id,
    COALESCE(ve.nama, '')::TEXT AS verifier_employee_name,
    COALESCE(ve.nip, '')::TEXT AS verifier_employee_nip,
    o.status,
    o.governance_document_id,
    COALESCE(gd.title, '')::TEXT AS governance_document_title,
    o.work_plan_item_id,
    COALESCE(gwpi.activity_code, '')::TEXT AS work_plan_item_code,
    COALESCE(gwpi.activity_name, '')::TEXT AS work_plan_item_name,
    o.performance_target_id,
    COALESCE(gpt.title, '')::TEXT AS performance_target_title,
    o.evidence_item_id,
    COALESCE(ge.title, '')::TEXT AS evidence_item_title,
    o.compliance_action_id,
    COALESCE(gca.title, '')::TEXT AS compliance_action_title,
    o.archive_document_id,
    COALESCE(ad.title, '')::TEXT AS archive_document_title,
    o.notes,
    o.verification_notes,
    o.completed_at,
    o.created_by_user_id,
    o.created_at,
    o.updated_at,
    (o.status <> 'completed' AND o.due_date < CURRENT_DATE)::BOOLEAN AS is_overdue,
    (o.status <> 'completed' AND o.reminder_date <= CURRENT_DATE AND o.due_date >= CURRENT_DATE)::BOOLEAN AS is_due_soon
FROM document_cycle_obligations o
JOIN document_cycle_catalogs c ON c.id = o.catalog_id
LEFT JOIN governance_units ou ON ou.id = o.owner_unit_id
LEFT JOIN employees re ON re.id = o.responsible_employee_id
LEFT JOIN employees ve ON ve.id = o.verifier_employee_id
LEFT JOIN governance_documents gd ON gd.id = o.governance_document_id
LEFT JOIN governance_work_plan_items gwpi ON gwpi.id = o.work_plan_item_id
LEFT JOIN governance_performance_targets gpt ON gpt.id = o.performance_target_id
LEFT JOIN governance_evidence_items ge ON ge.id = o.evidence_item_id
LEFT JOIN governance_compliance_actions gca ON gca.id = o.compliance_action_id
LEFT JOIN archive_documents ad ON ad.id = o.archive_document_id
WHERE (
    sqlc.arg(search)::TEXT = '' OR
    c.code ILIKE '%' || sqlc.arg(search) || '%' OR
    c.title ILIKE '%' || sqlc.arg(search) || '%' OR
    o.period_label ILIKE '%' || sqlc.arg(search) || '%' OR
    o.notes ILIKE '%' || sqlc.arg(search) || '%'
) AND (
    sqlc.arg(status)::TEXT = '' OR o.status = sqlc.arg(status)
) AND (
    sqlc.arg(frequency)::TEXT = '' OR c.frequency = sqlc.arg(frequency)
) AND (
    sqlc.arg(domain_area)::TEXT = '' OR o.domain_area = sqlc.arg(domain_area)
) AND (
    sqlc.arg(external_system)::TEXT = '' OR o.external_system = sqlc.arg(external_system)
) AND (
    sqlc.arg(period_year)::INT = 0 OR o.period_year = sqlc.arg(period_year)::INT
) AND (
    sqlc.arg(owner_unit_id)::UUID IS NULL OR o.owner_unit_id = sqlc.arg(owner_unit_id)::UUID
) AND (
    sqlc.arg(responsible_employee_id)::UUID IS NULL OR o.responsible_employee_id = sqlc.arg(responsible_employee_id)::UUID
) AND (
    sqlc.arg(reminder_only)::BOOLEAN = FALSE OR (
        o.status <> 'completed'
        AND o.reminder_date <= CURRENT_DATE
    )
)
ORDER BY
    CASE
        WHEN o.status <> 'completed' AND o.due_date < CURRENT_DATE THEN 0
        WHEN o.status <> 'completed' AND o.reminder_date <= CURRENT_DATE THEN 1
        WHEN o.status = 'waiting_verification' THEN 2
        WHEN o.status = 'draft' THEN 3
        WHEN o.status = 'not_started' THEN 4
        ELSE 5
    END,
    o.due_date ASC,
    c.sort_order,
    c.code;

-- name: EnsureDocumentCycleObligation :one
INSERT INTO document_cycle_obligations (
    catalog_id,
    period_year,
    period_label,
    period_start,
    period_end,
    due_date,
    reminder_date,
    owner_unit_id,
    responsible_employee_id,
    verifier_employee_id,
    status,
    notes,
    created_by_user_id
) VALUES (
    sqlc.arg(catalog_id),
    sqlc.arg(period_year),
    sqlc.arg(period_label),
    sqlc.arg(period_start),
    sqlc.arg(period_end),
    sqlc.arg(due_date),
    sqlc.arg(reminder_date),
    sqlc.arg(owner_unit_id),
    sqlc.arg(responsible_employee_id),
    sqlc.arg(verifier_employee_id),
    sqlc.arg(status),
    sqlc.arg(notes),
    sqlc.arg(created_by_user_id)
)
ON CONFLICT (catalog_id, period_year, period_label) DO UPDATE
SET updated_at = document_cycle_obligations.updated_at
RETURNING *;

-- name: GenerateDocumentCycleYearObligations :one
WITH params AS (
    SELECT
        sqlc.arg(period_year)::INT AS period_year,
        sqlc.arg(created_by_user_id)::UUID AS created_by_user_id
),
month_names(month_no, month_name) AS (
    VALUES
        (1, 'Januari'),
        (2, 'Februari'),
        (3, 'Maret'),
        (4, 'April'),
        (5, 'Mei'),
        (6, 'Juni'),
        (7, 'Juli'),
        (8, 'Agustus'),
        (9, 'September'),
        (10, 'Oktober'),
        (11, 'November'),
        (12, 'Desember')
),
quarter_periods(label, start_month, end_month) AS (
    VALUES
        ('Triwulan I', 1, 3),
        ('Triwulan II', 4, 6),
        ('Triwulan III', 7, 9),
        ('Triwulan IV', 10, 12)
),
semester_periods(label, start_month, end_month) AS (
    VALUES
        ('Semester I', 1, 6),
        ('Semester II', 7, 12)
),
raw_periods AS (
    SELECT
        c.id AS catalog_id,
        p.period_year,
        FORMAT('Bundel Harian %s %s', mn.month_name, p.period_year)::TEXT AS period_label,
        MAKE_DATE(p.period_year, mn.month_no, 1)::DATE AS period_start,
        (MAKE_DATE(p.period_year, mn.month_no, 1) + INTERVAL '1 month - 1 day')::DATE AS period_end,
        c.domain_area,
        c.external_system,
        c.default_owner_unit_id,
        c.default_responsible_employee_id,
        c.default_verifier_employee_id,
        c.deadline_days_after_period,
        c.reminder_days_before_due,
        p.created_by_user_id
    FROM document_cycle_catalogs c
    CROSS JOIN params p
    CROSS JOIN month_names mn
    WHERE c.is_active = TRUE AND c.frequency = 'daily'

    UNION ALL

    SELECT
        c.id AS catalog_id,
        p.period_year,
        FORMAT('Minggu %s %s', LPAD(w.week_no::TEXT, 2, '0'), p.period_year)::TEXT AS period_label,
        w.period_start,
        w.period_end,
        c.domain_area,
        c.external_system,
        c.default_owner_unit_id,
        c.default_responsible_employee_id,
        c.default_verifier_employee_id,
        c.deadline_days_after_period,
        c.reminder_days_before_due,
        p.created_by_user_id
    FROM document_cycle_catalogs c
    CROSS JOIN params p
    CROSS JOIN LATERAL (
        SELECT
            ROW_NUMBER() OVER (ORDER BY gs.period_start)::INT AS week_no,
            gs.period_start::DATE AS period_start,
            LEAST((gs.period_start + INTERVAL '6 days')::DATE, MAKE_DATE(p.period_year, 12, 31))::DATE AS period_end
        FROM GENERATE_SERIES(
            MAKE_DATE(p.period_year, 1, 1)::TIMESTAMP,
            MAKE_DATE(p.period_year, 12, 31)::TIMESTAMP,
            INTERVAL '7 days'
        ) AS gs(period_start)
    ) w
    WHERE c.is_active = TRUE AND c.frequency = 'weekly'

    UNION ALL

    SELECT
        c.id AS catalog_id,
        p.period_year,
        FORMAT('Bulanan %s %s', mn.month_name, p.period_year)::TEXT AS period_label,
        MAKE_DATE(p.period_year, mn.month_no, 1)::DATE AS period_start,
        (MAKE_DATE(p.period_year, mn.month_no, 1) + INTERVAL '1 month - 1 day')::DATE AS period_end,
        c.domain_area,
        c.external_system,
        c.default_owner_unit_id,
        c.default_responsible_employee_id,
        c.default_verifier_employee_id,
        c.deadline_days_after_period,
        c.reminder_days_before_due,
        p.created_by_user_id
    FROM document_cycle_catalogs c
    CROSS JOIN params p
    CROSS JOIN month_names mn
    WHERE c.is_active = TRUE AND c.frequency = 'monthly'

    UNION ALL

    SELECT
        c.id AS catalog_id,
        p.period_year,
        FORMAT('%s %s', qp.label, p.period_year)::TEXT AS period_label,
        MAKE_DATE(p.period_year, qp.start_month, 1)::DATE AS period_start,
        (MAKE_DATE(p.period_year, qp.end_month, 1) + INTERVAL '1 month - 1 day')::DATE AS period_end,
        c.domain_area,
        c.external_system,
        c.default_owner_unit_id,
        c.default_responsible_employee_id,
        c.default_verifier_employee_id,
        c.deadline_days_after_period,
        c.reminder_days_before_due,
        p.created_by_user_id
    FROM document_cycle_catalogs c
    CROSS JOIN params p
    CROSS JOIN quarter_periods qp
    WHERE c.is_active = TRUE AND c.frequency = 'quarterly'

    UNION ALL

    SELECT
        c.id AS catalog_id,
        p.period_year,
        FORMAT('%s %s', sp.label, p.period_year)::TEXT AS period_label,
        MAKE_DATE(p.period_year, sp.start_month, 1)::DATE AS period_start,
        (MAKE_DATE(p.period_year, sp.end_month, 1) + INTERVAL '1 month - 1 day')::DATE AS period_end,
        c.domain_area,
        c.external_system,
        c.default_owner_unit_id,
        c.default_responsible_employee_id,
        c.default_verifier_employee_id,
        c.deadline_days_after_period,
        c.reminder_days_before_due,
        p.created_by_user_id
    FROM document_cycle_catalogs c
    CROSS JOIN params p
    CROSS JOIN semester_periods sp
    WHERE c.is_active = TRUE AND c.frequency = 'semester'

    UNION ALL

    SELECT
        c.id AS catalog_id,
        p.period_year,
        FORMAT('Tahunan %s', p.period_year)::TEXT AS period_label,
        MAKE_DATE(p.period_year, 1, 1)::DATE AS period_start,
        MAKE_DATE(p.period_year, 12, 31)::DATE AS period_end,
        c.domain_area,
        c.external_system,
        c.default_owner_unit_id,
        c.default_responsible_employee_id,
        c.default_verifier_employee_id,
        c.deadline_days_after_period,
        c.reminder_days_before_due,
        p.created_by_user_id
    FROM document_cycle_catalogs c
    CROSS JOIN params p
    WHERE c.is_active = TRUE AND c.frequency = 'annual'

    UNION ALL

    SELECT
        c.id AS catalog_id,
        p.period_year,
        FORMAT('Periode %s-%s', p.period_year, p.period_year + 3)::TEXT AS period_label,
        MAKE_DATE(p.period_year, 1, 1)::DATE AS period_start,
        MAKE_DATE(p.period_year + 3, 12, 31)::DATE AS period_end,
        c.domain_area,
        c.external_system,
        c.default_owner_unit_id,
        c.default_responsible_employee_id,
        c.default_verifier_employee_id,
        c.deadline_days_after_period,
        c.reminder_days_before_due,
        p.created_by_user_id
    FROM document_cycle_catalogs c
    CROSS JOIN params p
    WHERE c.is_active = TRUE AND c.frequency = 'four_year'

    UNION ALL

    SELECT
        c.id AS catalog_id,
        p.period_year,
        FORMAT('Periode %s-%s', p.period_year, p.period_year + 4)::TEXT AS period_label,
        MAKE_DATE(p.period_year, 1, 1)::DATE AS period_start,
        MAKE_DATE(p.period_year + 4, 12, 31)::DATE AS period_end,
        c.domain_area,
        c.external_system,
        c.default_owner_unit_id,
        c.default_responsible_employee_id,
        c.default_verifier_employee_id,
        c.deadline_days_after_period,
        c.reminder_days_before_due,
        p.created_by_user_id
    FROM document_cycle_catalogs c
    CROSS JOIN params p
    WHERE c.is_active = TRUE AND c.frequency = 'five_year'
),
periods AS (
    SELECT
        catalog_id,
        period_year,
        period_label,
        period_start,
        period_end,
        domain_area,
        external_system,
        (period_end + deadline_days_after_period::INT)::DATE AS due_date,
        ((period_end + deadline_days_after_period::INT)::DATE - reminder_days_before_due::INT)::DATE AS reminder_date,
        default_owner_unit_id AS owner_unit_id,
        default_responsible_employee_id AS responsible_employee_id,
        default_verifier_employee_id AS verifier_employee_id,
        'not_started'::TEXT AS status,
        ''::TEXT AS notes,
        created_by_user_id
    FROM raw_periods
),
inserted AS (
    INSERT INTO document_cycle_obligations (
        catalog_id,
        period_year,
        period_label,
        period_start,
        period_end,
        domain_area,
        external_system,
        due_date,
        reminder_date,
        owner_unit_id,
        responsible_employee_id,
        verifier_employee_id,
        status,
        notes,
        created_by_user_id
    )
    SELECT
        catalog_id,
        period_year,
        period_label,
        period_start,
        period_end,
        domain_area,
        external_system,
        due_date,
        reminder_date,
        owner_unit_id,
        responsible_employee_id,
        verifier_employee_id,
        status,
        notes,
        created_by_user_id
    FROM periods
    ON CONFLICT (catalog_id, period_year, period_label) DO NOTHING
    RETURNING id, status, created_by_user_id
),
events AS (
    INSERT INTO document_cycle_events (
        obligation_id,
        event_type,
        from_status,
        to_status,
        notes,
        actor_user_id
    )
    SELECT
        id,
        'generated',
        '',
        status,
        'Jadwal dokumen dibuat dari generator tahunan.',
        created_by_user_id
    FROM inserted
    RETURNING id
)
SELECT
    (SELECT COUNT(*) FROM periods)::INT AS generated,
    (SELECT COUNT(*) FROM inserted)::INT AS inserted,
    (SELECT COUNT(*) FROM events)::INT AS event_count;

-- name: UpdateDocumentCycleObligation :one
UPDATE document_cycle_obligations
SET due_date = sqlc.arg(due_date),
    reminder_date = sqlc.arg(reminder_date),
    domain_area = CASE
        WHEN sqlc.arg(domain_area)::TEXT = '' THEN domain_area
        ELSE sqlc.arg(domain_area)
    END,
    external_system = sqlc.arg(external_system),
    owner_unit_id = sqlc.arg(owner_unit_id),
    responsible_employee_id = sqlc.arg(responsible_employee_id),
    verifier_employee_id = sqlc.arg(verifier_employee_id),
    governance_document_id = sqlc.arg(governance_document_id),
    work_plan_item_id = sqlc.arg(work_plan_item_id),
    performance_target_id = sqlc.arg(performance_target_id),
    evidence_item_id = sqlc.arg(evidence_item_id),
    compliance_action_id = sqlc.arg(compliance_action_id),
    archive_document_id = sqlc.arg(archive_document_id),
    notes = sqlc.arg(notes),
    verification_notes = sqlc.arg(verification_notes),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: UpdateDocumentCycleObligationStatus :one
UPDATE document_cycle_obligations
SET status = sqlc.arg(status),
    notes = CASE
        WHEN sqlc.arg(notes)::TEXT = '' THEN notes
        ELSE sqlc.arg(notes)
    END,
    verification_notes = CASE
        WHEN sqlc.arg(status)::TEXT IN ('waiting_verification', 'completed') AND sqlc.arg(notes)::TEXT <> '' THEN sqlc.arg(notes)
        ELSE verification_notes
    END,
    completed_at = CASE
        WHEN sqlc.arg(status)::TEXT = 'completed' THEN NOW()
        ELSE NULL
    END,
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: GetDocumentCycleObligationStatus :one
SELECT status FROM document_cycle_obligations WHERE id = $1;

-- name: GetDocumentCycleObligationCompletionReadiness :one
SELECT
    id,
    responsible_employee_id,
    verifier_employee_id,
    governance_document_id,
    work_plan_item_id,
    performance_target_id,
    evidence_item_id,
    compliance_action_id,
    archive_document_id
FROM document_cycle_obligations
WHERE id = $1;

-- name: DeleteDocumentCycleObligation :exec
DELETE FROM document_cycle_obligations WHERE id = $1;

-- name: CreateDocumentCycleEvent :one
INSERT INTO document_cycle_events (
    obligation_id,
    event_type,
    from_status,
    to_status,
    notes,
    actor_user_id
) VALUES (
    sqlc.arg(obligation_id),
    sqlc.arg(event_type),
    sqlc.arg(from_status),
    sqlc.arg(to_status),
    sqlc.arg(notes),
    sqlc.arg(actor_user_id)
)
RETURNING *;

-- name: ListDocumentCycleEventsByObligation :many
SELECT
    e.id,
    e.obligation_id,
    e.event_type,
    e.from_status,
    e.to_status,
    e.notes,
    e.actor_user_id,
    e.created_at,
    COALESCE(u.username, '')::TEXT AS actor_username
FROM document_cycle_events e
LEFT JOIN users u ON u.id = e.actor_user_id
WHERE e.obligation_id = $1
ORDER BY e.created_at DESC;
