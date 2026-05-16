const PUBLIC_SITE_EXACT_PATHS = ['/', '/ppdb', '/profil', '/berita', '/pengumuman', '/kontak'] as const;
const PUBLIC_SITE_PREFIXES = ['/berita/', '/pengumuman/'] as const;

const PUBLIC_API_EXACT_PATHS = [
	'/api/auth/logout',
	'/api/system/maintenance/status',
	'/api/public/analytics/events',
	'/api/public/branding',
	'/api/public/register-student',
	'/api/public/site/posts',
	'/api/public/site/announcements'
] as const;

const PUBLIC_API_PREFIXES = [
	'/api/exam/',
	'/api/branding/file/',
	'/releases/mobile/',
	'/api/public/site/pages/',
	'/api/public/site/posts/',
	'/api/public/site/announcements/'
] as const;

const PUBLIC_AUTH_EXACT_PATHS = ['/login', '/maintenance', '/manifest.webmanifest'] as const;

export function matchesPathSegment(pathname: string, prefix: string) {
	const normalizedPrefix = prefix === '/' ? '/' : prefix.replace(/\/$/, '');
	return pathname === normalizedPrefix || pathname.startsWith(`${normalizedPrefix}/`);
}

export function isPublicSitePath(pathname: string, authenticated = false) {
	if (authenticated && pathname === '/') return false;
	if (PUBLIC_SITE_EXACT_PATHS.includes(pathname as (typeof PUBLIC_SITE_EXACT_PATHS)[number])) {
		return true;
	}
	return PUBLIC_SITE_PREFIXES.some((prefix) => matchesPathSegment(pathname, prefix));
}

export function isPublicPath(pathname: string) {
	if (isPublicSitePath(pathname)) return true;
	if (PUBLIC_AUTH_EXACT_PATHS.includes(pathname as (typeof PUBLIC_AUTH_EXACT_PATHS)[number])) return true;
	if (PUBLIC_API_EXACT_PATHS.includes(pathname as (typeof PUBLIC_API_EXACT_PATHS)[number])) return true;
	return PUBLIC_API_PREFIXES.some((prefix) => matchesPathSegment(pathname, prefix));
}
