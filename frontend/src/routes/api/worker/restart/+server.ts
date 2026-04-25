import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

// Set WORKER_RESTART_CMD in .env, e.g.:
//   WORKER_RESTART_CMD=/home/deploy/pusaka/scripts/restart-worker.sh
const RESTART_CMD = process.env.WORKER_RESTART_CMD ?? 'pm2 restart pusaka-worker';

export const POST: RequestHandler = async () => {
	try {
		await execAsync(RESTART_CMD);
		return json({ ok: true });
	} catch (err) {
		const message = err instanceof Error ? err.message : 'Gagal restart worker';
		return json({ error: message }, { status: 500 });
	}
};
