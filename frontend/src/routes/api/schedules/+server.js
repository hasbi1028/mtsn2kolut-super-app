import { json } from '@sveltejs/kit';
import { db } from '$lib/server/db';

export function GET() {
  const items = db.prepare('SELECT * FROM schedules ORDER BY run_time ASC').all();
  return json({ items });
}

export async function PUT({ request }) {
  const { schedules } = await request.json();
  if (!Array.isArray(schedules)) return json({ error: 'schedules harus array' }, { status: 400 });

  const tx = db.transaction(() => {
    for (const s of schedules) {
      db.prepare(
        `UPDATE schedules SET label=?, run_time=?, run_type=?, is_enabled=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`
      ).run(s.label, s.run_time, s.run_type, s.is_enabled ? 1 : 0, s.id);
    }
  });
  tx();

  return json({ ok: true });
}
