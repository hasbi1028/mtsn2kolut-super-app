import { rawDb, DB_PATH } from '$lib/server/db';
import fs from 'fs';
import path from 'path';
import os from 'os';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async () => {
	const tmpPath = path.join(os.tmpdir(), `pusaka-backup-${Date.now()}.sqlite`);
	try {
		await rawDb.backup(tmpPath);
		const data = fs.readFileSync(tmpPath);
		const filename = `pusaka-backup-${new Date().toISOString().slice(0, 10)}.sqlite`;
		return new Response(data, {
			headers: {
				'Content-Type': 'application/octet-stream',
				'Content-Disposition': `attachment; filename="${filename}"`,
				'Content-Length': String(data.length),
			},
		});
	} finally {
		try { fs.unlinkSync(tmpPath); } catch { /* best-effort */ }
	}
};
