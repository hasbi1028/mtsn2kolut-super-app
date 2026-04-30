import { json } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import { execFile } from 'node:child_process';
import { promisify } from 'node:util';

const execFileAsync = promisify(execFile);
const RESTART_CMD = env.WORKER_RESTART_CMD ?? 'pm2 restart mtsn2kolut-pusaka-worker';

export const POST = async () => {
	try {
		await execFileAsync('/bin/bash', ['-lc', RESTART_CMD]);
		return json({ ok: true, message: 'Worker restart dijalankan' });
	} catch (err) {
		const message = err instanceof Error ? err.message : 'Gagal restart worker';
		return json({ error: message }, { status: 500 });
	}
};
