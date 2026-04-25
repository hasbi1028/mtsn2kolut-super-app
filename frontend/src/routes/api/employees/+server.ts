import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import crypto from 'crypto';
import { eq } from 'drizzle-orm';
import { db, rawDb } from '$lib/server/db';
import { employees } from '$lib/server/schema';

export const DELETE: RequestHandler = async ({ request }) => {
	const { id } = await request.json() as { id?: string };
	if (!id) return json({ error: 'id wajib diisi' }, { status: 400 });

	const emp = db.select({ id: employees.id }).from(employees).where(eq(employees.id, id)).get();
	if (!emp) return json({ error: 'Pegawai tidak ditemukan' }, { status: 404 });

	rawDb.transaction(() => {
		rawDb.prepare('DELETE FROM attendance_records WHERE employee_id = ?').run(id);
		rawDb.prepare('DELETE FROM jobs WHERE employee_id = ?').run(id);
		rawDb.prepare('DELETE FROM employees WHERE id = ?').run(id);
	})();

	return json({ ok: true });
};

export const GET: RequestHandler = () => {
	const items = rawDb.prepare(`
		SELECT
			e.id, e.nip, e.nama, e.unit_kerja, e.is_active, e.created_at,
			(SELECT j.status   FROM jobs j WHERE j.employee_id = e.id AND j.status IN ('running','queued')
			 ORDER BY CASE j.status WHEN 'running' THEN 0 ELSE 1 END, j.created_at DESC LIMIT 1) AS active_status,
			(SELECT j.run_type FROM jobs j WHERE j.employee_id = e.id AND j.status IN ('running','queued')
			 ORDER BY CASE j.status WHEN 'running' THEN 0 ELSE 1 END, j.created_at DESC LIMIT 1) AS active_run_type,
			(SELECT j.status   FROM jobs j WHERE j.employee_id = e.id ORDER BY j.created_at DESC LIMIT 1) AS last_status,
			(SELECT j.run_type FROM jobs j WHERE j.employee_id = e.id ORDER BY j.created_at DESC LIMIT 1) AS last_run_type
		FROM employees e
		ORDER BY e.created_at DESC
	`).all();
	return json({ items });
};

export const POST: RequestHandler = async ({ request }) => {
	const body = await request.json() as Record<string, string>;
	const { nip, nama, unit_kerja = '', pusaka_username, pusaka_password } = body;

	if (!nip || !nama || !pusaka_username || !pusaka_password) {
		return json({ error: 'nip, nama, pusaka_username, pusaka_password wajib diisi' }, { status: 400 });
	}

	try {
		db.insert(employees).values({
			id: crypto.randomUUID(), nip, nama, unit_kerja, pusaka_username, pusaka_password, is_active: 1,
		}).run();
		return json({ ok: true }, { status: 201 });
	} catch {
		return json({ error: 'Gagal menambah pegawai (mungkin NIP duplikat)' }, { status: 409 });
	}
};
