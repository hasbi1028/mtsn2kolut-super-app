import { createHmac, scryptSync, randomBytes, timingSafeEqual } from 'crypto';
import { rawDb } from './db.js';

const SESSION_SECRET = process.env.SESSION_SECRET ?? 'dev-insecure-change-me';
const SESSION_MS     = 12 * 60 * 60 * 1000; // 12 jam

// ── Password (scrypt) ─────────────────────────────────────────────────────────

export function hashPassword(password: string): string {
	const salt = randomBytes(16).toString('hex');
	const hash = scryptSync(password, salt, 64).toString('hex');
	return `${salt}:${hash}`;
}

export function verifyPassword(password: string, stored: string): boolean {
	const [salt, hash] = stored.split(':');
	if (!salt || !hash) return false;
	try {
		const derived = scryptSync(password, salt, 64);
		return timingSafeEqual(Buffer.from(hash, 'hex'), derived);
	} catch {
		return false;
	}
}

export function getPasswordHash(): string {
	const row = rawDb
		.prepare('SELECT value FROM app_settings WHERE key = ?')
		.get('admin_password_hash') as { value: string } | undefined;

	if (row?.value) return row.value;

	// First run: seed from ADMIN_PASSWORD env var or default 'admin'
	const initial = process.env.ADMIN_PASSWORD ?? 'admin';
	const hash    = hashPassword(initial);
	rawDb
		.prepare('INSERT OR REPLACE INTO app_settings (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)')
		.run('admin_password_hash', hash);
	return hash;
}

export function setPasswordHash(hash: string): void {
	rawDb
		.prepare('INSERT OR REPLACE INTO app_settings (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)')
		.run('admin_password_hash', hash);
}

// ── Session (stateless signed cookie) ────────────────────────────────────────

export function createSessionCookie(): string {
	const expires = Date.now() + SESSION_MS;
	const mac = createHmac('sha256', SESSION_SECRET).update(String(expires)).digest('hex');
	return `${expires}.${mac}`;
}

export function verifySession(cookie: string | undefined): boolean {
	if (!cookie) return false;
	const dot = cookie.lastIndexOf('.');
	if (dot < 1) return false;
	const expiresStr = cookie.slice(0, dot);
	const mac        = cookie.slice(dot + 1);
	if (mac.length !== 64) return false;
	try {
		const expected = createHmac('sha256', SESSION_SECRET).update(expiresStr).digest('hex');
		if (!timingSafeEqual(Buffer.from(mac, 'hex'), Buffer.from(expected, 'hex'))) return false;
	} catch {
		return false;
	}
	return Date.now() < Number(expiresStr);
}
