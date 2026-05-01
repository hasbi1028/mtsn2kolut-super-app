import { json } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import { execFile } from 'node:child_process';
import { promisify } from 'node:util';
import type { RequestHandler } from './$types';
import { hasAnyRole } from '$lib/server/route-access';

const execFileAsync = promisify(execFile);
const DEFAULT_RESTART_FILE = 'pm2';
const DEFAULT_RESTART_ARGS = ['restart', 'mtsn2kolut-pusaka-worker'] as const;
const UNSUPPORTED_SHELL_CHARS = /[;&|`$<>]/;

export function _parseRestartCommand(command = env.WORKER_RESTART_CMD ?? '') {
	const trimmed = command.trim();
	if (!trimmed) {
		return { file: DEFAULT_RESTART_FILE, args: [...DEFAULT_RESTART_ARGS] };
	}
	const parts = trimmed.split(/\s+/);
	if (parts.some((part) => UNSUPPORTED_SHELL_CHARS.test(part))) {
		throw new Error('restart command contains unsupported shell syntax');
	}
	return { file: parts[0], args: parts.slice(1) };
}

export const POST: RequestHandler = async (event) => {
	if (!hasAnyRole(event.locals.user, ['admin'])) {
		return json({ error: 'forbidden: admin role required' }, { status: 403 });
	}
	const user = event.locals.user;
	if (!user) {
		return json({ error: 'forbidden: admin role required' }, { status: 403 });
	}

	try {
		const { file, args } = _parseRestartCommand();
		await execFileAsync(file, args, { timeout: 15_000 });
		console.info('[pusaka/worker/restart] restart requested', {
			user_id: user.id,
			command: [file, ...args].join(' ')
		});
		return json({ ok: true, message: 'Worker restart dijalankan' });
	} catch (err) {
		console.error('[pusaka/worker/restart] restart failed', err);
		return json({ error: 'Restart worker gagal diproses' }, { status: 500 });
	}
};
