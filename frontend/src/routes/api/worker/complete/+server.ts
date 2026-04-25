import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { rawDb } from '$lib/server/db';
import { checkWorkerAuth } from '$lib/server/workerAuth';

interface CompleteBody {
	job_id:    string;
	tanggal:   string;
	jam_masuk:  string;
	jam_pulang: string;
}

export const POST: RequestHandler = async ({ request }) => {
	if (!checkWorkerAuth(request)) return json({ error: 'Unauthorized' }, { status: 401 });

	const { job_id, tanggal, jam_masuk, jam_pulang } = await request.json() as CompleteBody;
	if (!job_id || !tanggal) return json({ error: 'job_id dan tanggal wajib' }, { status: 400 });

	rawDb.transaction(() => {
		// Upsert attendance — hanya overwrite field yang non-empty
		rawDb.prepare(`
			INSERT INTO attendance_records (id, employee_id, tanggal, jam_masuk, jam_pulang, source_job_id)
			SELECT lower(hex(randomblob(16))), employee_id, ?, ?, ?, id
			FROM jobs WHERE id = ?
			ON CONFLICT(employee_id, tanggal) DO UPDATE SET
			  jam_masuk     = CASE WHEN excluded.jam_masuk  != '' THEN excluded.jam_masuk  ELSE jam_masuk  END,
			  jam_pulang    = CASE WHEN excluded.jam_pulang != '' THEN excluded.jam_pulang ELSE jam_pulang END,
			  source_job_id = excluded.source_job_id,
			  updated_at    = CURRENT_TIMESTAMP
		`).run(tanggal, jam_masuk, jam_pulang, job_id);

		rawDb.prepare(`
			UPDATE jobs
			SET status='success', error_message='', next_retry_at=NULL, updated_at=CURRENT_TIMESTAMP
			WHERE id = ?
		`).run(job_id);
	})();

	return json({ ok: true });
};
