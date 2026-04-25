import { json } from '@sveltejs/kit';
import type { RequestHandler } from '@sveltejs/kit';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

export const POST: RequestHandler = async () => {
	try {
		await execAsync('pm2 restart pusaka-worker');
		return json({ ok: true });
	} catch (err) {
		const message = err instanceof Error ? err.message : 'Gagal restart worker';
		return json({ error: message }, { status: 500 });
	}
};
