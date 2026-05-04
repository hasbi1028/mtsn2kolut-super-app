export function safeSameOriginRedirectPath(value: string | null | undefined, fallback = '/'): string {
	if (!value) return fallback;
	let decoded = value;
	try {
		decoded = decodeURIComponent(value);
	} catch {
		return fallback;
	}
	if (!decoded.startsWith('/')) return fallback;
	if (decoded.startsWith('//')) return fallback;
	if (decoded.includes('\\')) return fallback;
	if (/[\u0000-\u001F\u007F]/.test(decoded)) return fallback;
	if (/^[a-z][a-z0-9+.-]*:/i.test(decoded)) return fallback;
	if (decoded === '/login' || decoded.startsWith('/login?') || decoded.startsWith('/login#')) return fallback;
	return decoded;
}
