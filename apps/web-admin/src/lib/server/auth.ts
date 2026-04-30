type JwtPayload = {
	exp?: number;
	type?: string;
	sub?: string;
	uid?: string;
	usr?: string;
	role?: string;
	roles?: string[];
	ssid?: string;
	eid?: string;
	sid?: string;
	pid?: string;
};

function decodePayload(token: string): JwtPayload | null {
	const parts = token.split('.');
	if (parts.length !== 3) return null;
	try {
		const b64 = parts[1].replace(/-/g, '+').replace(/_/g, '/');
		const padded = b64 + '='.repeat((4 - (b64.length % 4 || 4)) % 4);
		const json = globalThis.atob(padded);
		return JSON.parse(json) as JwtPayload;
	} catch {
		return null;
	}
}

export function isAccessTokenValid(token: string | undefined): boolean {
	if (!token) return false;
	const payload = decodePayload(token);
	if (!payload || payload.type !== 'access' || !payload.exp) return false;
	return Date.now() < payload.exp * 1000;
}

export function hasRefreshToken(token: string | undefined): boolean {
	if (!token) return false;
	const payload = decodePayload(token);
	return !!payload && payload.type === 'refresh' && !!payload.exp && Date.now() < payload.exp * 1000;
}

export function getUserFromToken(token: string | undefined) {
	if (!token) return null;
	const p = decodePayload(token);
	if (!p || p.type !== 'access') return null;
	return {
		id: p.uid ?? p.sub,
		username: p.usr,
		role: p.role,
		roles: p.roles || [p.role],
		session_id: p.ssid,
		employee_id: p.eid,
		student_id: p.sid,
		parent_id: p.pid,
	};
}
