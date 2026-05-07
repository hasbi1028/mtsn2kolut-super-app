const CACHE_VERSION = 'v1';
const CACHE_NAME = `mtsn2-web-admin-shell-${CACHE_VERSION}`;
const APP_SHELL_PATHS = ['/manifest.webmanifest', '/offline.html', '/pwa-icon.svg', '/robots.txt'];

function isApiPath(pathname) {
	return pathname === '/api' || pathname.startsWith('/api/');
}

function isDocumentRequest(request) {
	return request.mode === 'navigate' || request.destination === 'document';
}

function shouldBypassCache(request, url) {
	if (request.method !== 'GET') return true;
	if (url.origin !== self.location.origin) return true;
	if (isApiPath(url.pathname)) return true;
	if (isDocumentRequest(request)) return true;
	return false;
}

function shouldCacheAppShellAsset(request, url) {
	if (shouldBypassCache(request, url)) return false;
	if (url.pathname.startsWith('/_app/immutable/')) return true;
	return APP_SHELL_PATHS.includes(url.pathname);
}

function noStoreRequest(request) {
	try {
		return new Request(request, { cache: 'no-store' });
	} catch {
		return request;
	}
}

self.addEventListener('install', (event) => {
	event.waitUntil(
		caches
			.open(CACHE_NAME)
			.then((cache) => cache.addAll(APP_SHELL_PATHS))
			.then(() => self.skipWaiting())
	);
});

self.addEventListener('activate', (event) => {
	event.waitUntil(
		caches
			.keys()
			.then((keys) => Promise.all(keys.filter((key) => key.startsWith('mtsn2-web-admin-shell-') && key !== CACHE_NAME).map((key) => caches.delete(key))))
			.then(() => self.clients.claim())
	);
});

self.addEventListener('fetch', (event) => {
	const request = event.request;
	const url = new URL(request.url);

	if (isApiPath(url.pathname)) {
		event.respondWith(fetch(noStoreRequest(request)));
		return;
	}

	if (isDocumentRequest(request)) {
		event.respondWith(
			fetch(noStoreRequest(request)).catch(() => caches.match('/offline.html'))
		);
		return;
	}

	if (!shouldCacheAppShellAsset(request, url)) return;

	event.respondWith(
		caches.match(request).then((cached) => {
			if (cached) return cached;
			return fetch(request).then((response) => {
				if (response.ok && response.type === 'basic') {
					const copy = response.clone();
					caches.open(CACHE_NAME).then((cache) => cache.put(request, copy));
				}
				return response;
			});
		})
	);
});
