import crypto from 'crypto';
import { db } from '$lib/server/db';

const insertJobStmt = db.prepare(
  `INSERT INTO jobs (id, employee_id, run_type, status, error_message, attempts, max_attempts)
   VALUES (?, ?, ?, 'queued', '', 0, ?)`
);

export function enqueueOne(employeeId, runType = 'morning', maxAttempts = 3) {
  const exists = db
    .prepare(
      `SELECT 1
       FROM jobs
       WHERE employee_id = ?
         AND run_type = ?
         AND status IN ('queued','running')
       LIMIT 1`
    )
    .get(employeeId, runType);

  if (exists) return { inserted: 0, skipped: 1 };
  insertJobStmt.run(crypto.randomUUID(), employeeId, runType, maxAttempts);
  return { inserted: 1, skipped: 0 };
}

export function enqueueAllActive(runType = 'morning', maxAttempts = 3) {
  const employees = db.prepare('SELECT id FROM employees WHERE is_active = 1').all();

  const tx = db.transaction((rows) => {
    let inserted = 0;
    let skipped = 0;
    for (const row of rows) {
      const result = enqueueOne(row.id, runType, maxAttempts);
      inserted += result.inserted;
      skipped += result.skipped;
    }
    return { inserted, skipped, total_active: rows.length };
  });

  return tx(employees);
}

export function getQueueStats() {
  const rows = db
    .prepare(
      `SELECT status, COUNT(*) AS n
       FROM jobs
       GROUP BY status`
    )
    .all();

  const stats = { queued: 0, running: 0, success: 0, failed: 0 };
  for (const row of rows) stats[row.status] = row.n;

  const retryDue = db
    .prepare(
      `SELECT COUNT(*) AS n
       FROM jobs
       WHERE status = 'queued'
         AND next_retry_at IS NOT NULL
         AND datetime(next_retry_at) <= CURRENT_TIMESTAMP`
    )
    .get().n;

  return {
    ...stats,
    retry_due: retryDue,
    total: stats.queued + stats.running + stats.success + stats.failed
  };
}
