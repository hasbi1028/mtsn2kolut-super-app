import { rawDb, DB_PATH } from '$lib/server/db';
import { json } from '@sveltejs/kit';
import BetterSqlite3 from 'better-sqlite3';
import fs from 'fs';
import path from 'path';
import os from 'os';
import type { RequestHandler } from './$types';

const REQUIRED_TABLES = ['employees', 'schedules', 'jobs', 'attendance_records', 'app_settings'];

export const POST: RequestHandler = async ({ request }) => {
	let tmpPath = '';
	try {
		const formData = await request.formData();
		const file = formData.get('file');
		if (!(file instanceof File)) return json({ error: 'File tidak ditemukan' }, { status: 400 });
		if (file.size === 0) return json({ error: 'File kosong' }, { status: 400 });
		if (file.size > 100 * 1024 * 1024) return json({ error: 'File terlalu besar (maks 100MB)' }, { status: 400 });

		tmpPath = path.join(os.tmpdir(), `pusaka-restore-${Date.now()}.sqlite`);
		fs.writeFileSync(tmpPath, Buffer.from(await file.arrayBuffer()));

		// Validate SQLite file
		let tmpDb: InstanceType<typeof BetterSqlite3> | null = null;
		try {
			tmpDb = new BetterSqlite3(tmpPath, { readonly: true });
			const tables = (tmpDb.prepare("SELECT name FROM sqlite_master WHERE type='table'").all() as { name: string }[]).map((r) => r.name);
			const missing = REQUIRED_TABLES.filter((t) => !tables.includes(t));
			if (missing.length > 0) {
				return json({ error: `File tidak valid: tabel ${missing.join(', ')} tidak ditemukan` }, { status: 400 });
			}
		} catch {
			return json({ error: 'File bukan database SQLite yang valid' }, { status: 400 });
		} finally {
			tmpDb?.close();
		}

		// Resolve absolute DB path
		const absDbPath = path.isAbsolute(DB_PATH) ? DB_PATH : path.resolve(DB_PATH);

		// Checkpoint current WAL to consolidate all data
		rawDb.pragma('wal_checkpoint(TRUNCATE)');

		// Stage the new file next to the DB, then atomically rename
		const stagingPath = absDbPath + '.restore';
		fs.copyFileSync(tmpPath, stagingPath);

		// Remove WAL/SHM so the new DB starts clean after restart
		try { fs.unlinkSync(absDbPath + '-wal'); } catch { /* ok if not exists */ }
		try { fs.unlinkSync(absDbPath + '-shm'); } catch { /* ok if not exists */ }

		// Atomic rename (same filesystem)
		fs.renameSync(stagingPath, absDbPath);

		// Restart the process so the new DB is picked up
		setTimeout(() => process.exit(0), 1500);

		return json({ ok: true, message: 'Restore berhasil. App sedang restart...' });
	} finally {
		if (tmpPath) try { fs.unlinkSync(tmpPath); } catch { /* best-effort */ }
	}
};
