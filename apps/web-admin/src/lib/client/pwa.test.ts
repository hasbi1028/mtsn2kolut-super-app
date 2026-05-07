import { describe, expect, it } from 'vitest';
import {
	isApiPath,
	isDocumentRequest,
	shouldBypassServiceWorkerCache,
	shouldCacheAppShellAsset,
} from './pwa-cache-policy';
import { shouldRegisterWebAdminPwa } from './pwa';

describe('web admin PWA cache policy', () => {
	const origin = 'https://mtsn2kolut.sch.id';
	const getRequest = { method: 'GET', mode: 'same-origin', destination: 'script' } as Pick<Request, 'method' | 'mode' | 'destination'>;

	it('never caches API or document navigations', () => {
		expect(isApiPath('/api/bank-soal/questions')).toBe(true);
		expect(isDocumentRequest({ method: 'GET', mode: 'navigate', destination: '' } as Pick<Request, 'method' | 'mode' | 'destination'>)).toBe(true);
		expect(shouldBypassServiceWorkerCache(getRequest, new URL('/api/auth/account', origin), origin)).toBe(true);
		expect(shouldBypassServiceWorkerCache({ ...getRequest, mode: 'navigate', destination: 'document' }, new URL('/bank-soal/tambah', origin), origin)).toBe(true);
		expect(shouldCacheAppShellAsset(getRequest, new URL('/api/bank-soal/questions', origin), origin)).toBe(false);
	});

	it('only caches static app shell and immutable assets from same origin', () => {
		expect(shouldCacheAppShellAsset(getRequest, new URL('/manifest.webmanifest', origin), origin)).toBe(true);
		expect(shouldCacheAppShellAsset(getRequest, new URL('/_app/immutable/chunks/a.js', origin), origin)).toBe(true);
		expect(shouldCacheAppShellAsset(getRequest, new URL('https://cdn.example.test/a.js'), origin)).toBe(false);
		expect(shouldCacheAppShellAsset({ ...getRequest, method: 'POST' }, new URL('/manifest.webmanifest', origin), origin)).toBe(false);
	});
});

describe('web admin PWA registration guard', () => {
	const sw = { register: async () => ({}) } as unknown as ServiceWorkerContainer;
	const location = { protocol: 'https:' } as Pick<Location, 'protocol'>;

	it('registers only for authenticated internal pages with service worker support', () => {
		expect(shouldRegisterWebAdminPwa({ isAuthenticated: true, isPublicSite: false, isLogin: false, serviceWorker: sw, location })).toBe(true);
		expect(shouldRegisterWebAdminPwa({ isAuthenticated: false, isPublicSite: false, isLogin: false, serviceWorker: sw, location })).toBe(false);
		expect(shouldRegisterWebAdminPwa({ isAuthenticated: true, isPublicSite: true, isLogin: false, serviceWorker: sw, location })).toBe(false);
		expect(shouldRegisterWebAdminPwa({ isAuthenticated: true, isPublicSite: false, isLogin: true, serviceWorker: sw, location })).toBe(false);
		expect(shouldRegisterWebAdminPwa({ isAuthenticated: true, isPublicSite: false, isLogin: false, serviceWorker: undefined, location })).toBe(false);
	});
});
