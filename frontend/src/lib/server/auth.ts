import { createHmac, timingSafeEqual } from 'crypto';

const SESSION_SECRET = process.env.SESSION_SECRET ?? 'dev-insecure-change-me';
const SESSION_MS     = 12 * 60 * 60 * 1000;

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
