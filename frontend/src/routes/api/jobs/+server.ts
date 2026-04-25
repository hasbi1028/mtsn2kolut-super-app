import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { rawDb } from '$lib/server/db';
import type { JobStatus } from '$lib/server/schema';

const VALID_STATUSES = new Set<JobStatus>(['queued', 'running', 'success', 'failed']);

const stmtAll = rawDb.prepare(`
	SELECT j.id, j.run_type, j.status, j.error_message, j.attempts, j.max_attempts,
		j.next_retry_at, j.claimed_by, j.created_at,
		strftime('%Y-%m-%d %H:%M:%S', datetime(j.created_at, '+8 hours')) || ' WITA' AS created_at_wita,
		e.nip, e.nama
	FROM jobs j JOIN employees e ON e.id = j.employee_id
	ORDER BY j.created_at DESC LIMIT ?`);

const stmtByStatus = rawDb.prepare(`
	SELECT j.id, j.run_type, j.status, j.error_message, j.attempts, j.max_attempts,
		j.next_retry_at, j.claimed_by, j.created_at,
		strftime('%Y-%m-%d %H:%M:%S', datetime(j.created_at, '+8 hours')) || ' WITA' AS created_at_wita,
		e.nip, e.nama
	FROM jobs j JOIN employees e ON e.id = j.employee_id
	WHERE j.status = ?
	ORDER BY j.created_at DESC LIMIT ?`);

export const GET: RequestHandler = ({ url }) => {
	const limit  = Math.min(500, Math.max(1, Number(url.searchParams.get('limit') ?? 50)));
	const status = (url.searchParams.get('status') ?? '') as JobStatus;

	const items = VALID_STATUSES.has(status)
		? stmtByStatus.all(status, limit)
		: stmtAll.all(limit);

	return json({ items });
};
