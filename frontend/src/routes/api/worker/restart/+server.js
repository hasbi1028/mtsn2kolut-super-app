import { json } from '@sveltejs/kit';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

export async function POST() {
  try {
    await execAsync('pm2 restart pusaka-worker');
    return json({ ok: true });
  } catch (err) {
    return json({ error: err.message || 'Gagal restart worker' }, { status: 500 });
  }
}
