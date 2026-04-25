import crypto from 'crypto';
import { eq, inArray, and, sql } from 'drizzle-orm';
import { db, rawDb } from './db.js';
import { jobs, employees } from './schema.js';
import type { RunType } from './schema.js';

export function enqueueOne(employeeId: string, runType: RunType = 'morning', maxAttempts = 3) {
	const exists = db.select({ id: jobs.id })
		.from(jobs)
		.where(and(eq(jobs.employee_id, employeeId), eq(jobs.run_type, runType), inArray(jobs.status, ['queued', 'running'])))
		.limit(1)
		.get();

	if (exists) return { inserted: 0, skipped: 1 };

	db.insert(jobs).values({
		id:            crypto.randomUUID(),
		employee_id:   employeeId,
		run_type:      runType,
		status:        'queued',
		error_message: '',
		attempts:      0,
		max_attempts:  maxAttempts,
	}).run();

	return { inserted: 1, skipped: 0 };
}

export function enqueueAllActive(runType: RunType = 'morning', maxAttempts = 3) {
	const activeEmployees = db.select({ id: employees.id })
		.from(employees)
		.where(eq(employees.is_active, 1))
		.all();

	return db.transaction((tx) => {
		let inserted = 0;
		let skipped  = 0;
		for (const emp of activeEmployees) {
			// Check inside transaction with rawDb for performance
			const exists = rawDb.prepare(
				`SELECT 1 FROM jobs WHERE employee_id=? AND run_type=? AND status IN ('queued','running') LIMIT 1`
			).get(emp.id, runType);

			if (exists) { skipped++; continue; }

			tx.insert(jobs).values({
				id:            crypto.randomUUID(),
				employee_id:   emp.id,
				run_type:      runType,
				status:        'queued',
				error_message: '',
				attempts:      0,
				max_attempts:  maxAttempts,
			}).run();
			inserted++;
		}
		return { inserted, skipped, total_active: activeEmployees.length };
	});
}

export interface QueueStats {
	queued:    number;
	running:   number;
	success:   number;
	failed:    number;
	retry_due: number;
	total:     number;
}

export function getQueueStats(): QueueStats {
	const rows = rawDb.prepare(`SELECT status, COUNT(*) AS n FROM jobs GROUP BY status`).all() as Array<{ status: string; n: number }>;
	const stats: Record<string, number> = { queued: 0, running: 0, success: 0, failed: 0 };
	for (const row of rows) stats[row.status] = row.n;

	const retryDue = (rawDb.prepare(
		`SELECT COUNT(*) AS n FROM jobs WHERE status='queued' AND next_retry_at IS NOT NULL AND datetime(next_retry_at) <= CURRENT_TIMESTAMP`
	).get() as { n: number }).n;

	const total = stats.queued + stats.running + stats.success + stats.failed;
	return { queued: stats.queued, running: stats.running, success: stats.success, failed: stats.failed, retry_due: retryDue, total };
}
