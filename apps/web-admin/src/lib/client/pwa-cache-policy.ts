export const WEB_ADMIN_STATIC_CACHE_PREFIX = 'mtsn2-web-admin-shell';

const APP_SHELL_PATHS = new Set([
	'/manifest.webmanifest',
	'/offline.html',
	'/pwa-icon.svg',
	'/brand/m2k-mark.svg',
	'/favicon.ico',
	'/favicon-32x32.png',
	'/pwa-icon-192.png',
	'/pwa-icon-512.png',
	'/robots.txt',
	'/sw.js',
]);

export function isSameOriginUrl(url: URL, origin: string): boolean {
	return url.origin === origin;
}

export function isApiPath(pathname: string): boolean {
	return pathname === '/api' || pathname.startsWith('/api/');
}

export function isDocumentRequest(request: Pick<Request, 'mode' | 'destination'>): boolean {
	return request.mode === 'navigate' || request.destination === 'document';
}

export function shouldBypassServiceWorkerCache(
	request: Pick<Request, 'method' | 'mode' | 'destination'>,
	url: URL,
	origin: string
): boolean {
	if (request.method !== 'GET') return true;
	if (!isSameOriginUrl(url, origin)) return true;
	if (isApiPath(url.pathname)) return true;
	if (isDocumentRequest(request)) return true;
	return false;
}

export function shouldCacheAppShellAsset(
	request: Pick<Request, 'method' | 'mode' | 'destination'>,
	url: URL,
	origin: string
): boolean {
	if (shouldBypassServiceWorkerCache(request, url, origin)) return false;
	if (url.pathname.startsWith('/_app/immutable/')) return true;
	return APP_SHELL_PATHS.has(url.pathname);
}
