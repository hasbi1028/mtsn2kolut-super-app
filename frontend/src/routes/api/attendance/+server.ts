import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { rawDb } from '$lib/server/db';

const stmtAll = rawDb.prepare(`
	SELECT a.*,
		strftime('%Y-%m-%d %H:%M:%S', datetime(a.updated_at, '+8 hours')) || ' WITA' AS updated_at_wita,
		e.nip, e.nama, e.unit_kerja
	FROM attendance_records a JOIN employees e ON e.id = a.employee_id
	ORDER BY e.nip ASC LIMIT ?`);

const stmtByDate = rawDb.prepare(`
	SELECT a.*,
		strftime('%Y-%m-%d %H:%M:%S', datetime(a.updated_at, '+8 hours')) || ' WITA' AS updated_at_wita,
		e.nip, e.nama, e.unit_kerja
	FROM attendance_records a JOIN employees e ON e.id = a.employee_id
	WHERE a.tanggal = ?
	ORDER BY e.nip ASC LIMIT ?`);

export const GET: RequestHandler = ({ url }) => {
	const limit = Math.min(500, Math.max(1, Number(url.searchParams.get('limit') ?? 50)));
	const date  = url.searchParams.get('date') ?? '';

	const items = /^\d{4}-\d{2}-\d{2}$/.test(date)
		? stmtByDate.all(date, limit)
		: stmtAll.all(limit);

	return json({ items });
};
