import { json } from '@sveltejs/kit';
import crypto from 'crypto';
import { db } from '$lib/server/db';

export async function DELETE({ request }) {
  const { id } = await request.json();
  if (!id) return json({ error: 'id wajib diisi' }, { status: 400 });

  const employee = db.prepare('SELECT id FROM employees WHERE id = ?').get(id);
  if (!employee) return json({ error: 'Pegawai tidak ditemukan' }, { status: 404 });

  db.transaction(() => {
    db.prepare('DELETE FROM attendance_records WHERE employee_id = ?').run(id);
    db.prepare('DELETE FROM jobs WHERE employee_id = ?').run(id);
    db.prepare('DELETE FROM employees WHERE id = ?').run(id);
  })();

  return json({ ok: true });
}

export function GET() {
  const items = db.prepare(`
    SELECT
      e.id, e.nip, e.nama, e.unit_kerja, e.is_active, e.created_at,
      -- job aktif: running > queued (prioritas running)
      (SELECT j.status   FROM jobs j WHERE j.employee_id = e.id AND j.status IN ('running','queued')
       ORDER BY CASE j.status WHEN 'running' THEN 0 ELSE 1 END, j.created_at DESC LIMIT 1) AS active_status,
      (SELECT j.run_type FROM jobs j WHERE j.employee_id = e.id AND j.status IN ('running','queued')
       ORDER BY CASE j.status WHEN 'running' THEN 0 ELSE 1 END, j.created_at DESC LIMIT 1) AS active_run_type,
      -- job terakhir (apapun statusnya)
      (SELECT j.status   FROM jobs j WHERE j.employee_id = e.id ORDER BY j.created_at DESC LIMIT 1) AS last_status,
      (SELECT j.run_type FROM jobs j WHERE j.employee_id = e.id ORDER BY j.created_at DESC LIMIT 1) AS last_run_type
    FROM employees e
    ORDER BY e.created_at DESC
  `).all();
  return json({ items });
}

export async function POST({ request }) {
  const body = await request.json();
  const { nip, nama, unit_kerja = '', pusaka_username, pusaka_password } = body;

  if (!nip || !nama || !pusaka_username || !pusaka_password) {
    return json({ error: 'nip, nama, pusaka_username, pusaka_password wajib diisi' }, { status: 400 });
  }

  try {
    db.prepare(
      `INSERT INTO employees (id, nip, nama, unit_kerja, pusaka_username, pusaka_password, is_active)
       VALUES (?, ?, ?, ?, ?, ?, 1)`
    ).run(crypto.randomUUID(), nip, nama, unit_kerja, pusaka_username, pusaka_password);
    return json({ ok: true }, { status: 201 });
  } catch (err) {
    return json({ error: 'Gagal menambah pegawai (mungkin NIP duplikat)' }, { status: 409 });
  }
}
