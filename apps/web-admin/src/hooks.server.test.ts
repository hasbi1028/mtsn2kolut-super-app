import { afterEach, describe, expect, it, vi } from 'vitest';

function jwt(payload: Record<string, unknown>) {
	const encoded = btoa(JSON.stringify(payload)).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
	return `header.${encoded}.signature`;
}

function token(type: 'access' | 'refresh', extra: Record<string, unknown> = {}) {
	return jwt({
		type,
		exp: Math.floor(Date.now() / 1000) + 3600,
		...extra,
	});
}

function makeEvent(refreshToken = token('refresh')) {
	return {
		request: new Request('http://localhost/dashboard', {
			headers: { 'user-agent': 'vitest-agent' }
		}),
		getClientAddress: () => '127.0.0.1',
		fetch: vi.fn<typeof fetch>(),
		locals: {
			accessToken: undefined as string | undefined,
			user: undefined as { id: string; username: string; role: string; roles: string[] } | undefined
		},
		cookies: {
			get: vi.fn((name: string) => (name === 'refresh_token' ? refreshToken : undefined)),
			set: vi.fn(),
			delete: vi.fn()
		}
	};
}

async function loadHandleFetch() {
	const mod = await loadHooks();
	return mod.handleFetch;
}

async function loadHooks() {
	vi.resetModules();
	vi.doUnmock('$lib/server/api');
	return await import('./hooks.server');
}

function makeHandleEvent(accessToken?: string, refreshToken?: string, url = 'http://localhost/dashboard', method = 'GET', headers: Record<string, string> = {}) {
	return {
		request: new Request(url, {
			method,
			headers: { 'user-agent': 'vitest-agent', ...headers }
		}),
		url: new URL(url),
		getClientAddress: () => '127.0.0.1',
		fetch: vi.fn<typeof fetch>(),
		locals: {
			accessToken: undefined as string | undefined,
			user: undefined as { id: string; username: string; role: string; roles: string[] } | undefined
		},
		cookies: {
			get: vi.fn((name: string) => {
				if (name === 'access_token') return accessToken;
				if (name === 'refresh_token') return refreshToken;
				return undefined;
			}),
			set: vi.fn(),
			delete: vi.fn()
		}
	};
}

function validationOkResponse() {
	return new Response(JSON.stringify({ data: [] }), {
		status: 200,
		headers: { 'Content-Type': 'application/json' }
	});
}

function callUrl(input: RequestInfo | URL | undefined): string {
	if (typeof input === 'string') return input;
	if (input instanceof URL) return input.toString();
	if (input instanceof Request) return input.url;
	return '';
}

describe('SvelteKit handleFetch auth refresh', () => {
	afterEach(() => {
		vi.restoreAllMocks();
	});

	it('retries an authenticated backend request with a rotated access token after a 401', async () => {
		const handleFetch = await loadHandleFetch();
		const nextAccess = token('access', { uid: 'u1', role: 'admin' });
		const nextRefresh = token('refresh');
		const globalFetch = vi.spyOn(globalThis, 'fetch');
		const routeFetch = vi.fn<typeof fetch>()
			.mockResolvedValueOnce(new Response(JSON.stringify({ error: 'unauthorized' }), { status: 401 }))
			.mockResolvedValueOnce(new Response(JSON.stringify({ data: { access_token: nextAccess, refresh_token: nextRefresh } }), {
				status: 200,
				headers: { 'Content-Type': 'application/json' }
			}))
			.mockResolvedValueOnce(new Response(JSON.stringify({ data: { ok: true } }), { status: 200 }));
		const request = new Request('http://127.0.0.1:8080/api/document-cycles/stats', {
			headers: { Authorization: `Bearer ${token('access')}` }
		});
		const event = makeEvent();

		const response = await handleFetch({ event, request, fetch: routeFetch } as never);

		expect(response.status).toBe(200);
		expect(globalFetch).not.toHaveBeenCalled();
		expect(routeFetch).toHaveBeenCalledTimes(3);
		const refreshCallInput = routeFetch.mock.calls[1]?.[0] as RequestInfo | URL | undefined;
		expect(callUrl(refreshCallInput)).toContain('/api/auth/refresh');
		const retryRequest = routeFetch.mock.calls[2]?.[0] as Request;
		expect(retryRequest.headers.get('Authorization')).toBe(`Bearer ${nextAccess}`);
		expect(event.locals.accessToken).toBe(nextAccess);
		expect(event.locals.user).toBeUndefined();
		expect(event.cookies.set).toHaveBeenCalledWith('access_token', nextAccess, expect.objectContaining({ httpOnly: true }));
		expect(event.cookies.set).toHaveBeenCalledWith('refresh_token', nextRefresh, expect.objectContaining({ httpOnly: true }));
	}, 10000);

	it('retries a consumed POST request body from a pre-fetch clone', async () => {
		const handleFetch = await loadHandleFetch();
		const nextAccess = token('access', { uid: 'u1', role: 'admin' });
		const nextRefresh = token('refresh');
		const globalFetch = vi.spyOn(globalThis, 'fetch');
		const routeFetch = vi.fn<typeof fetch>()
			.mockImplementationOnce(async (input) => {
				await (input as Request).text();
				return new Response(JSON.stringify({ error: 'unauthorized' }), { status: 401 });
			})
			.mockResolvedValueOnce(new Response(JSON.stringify({ data: { access_token: nextAccess, refresh_token: nextRefresh } }), {
				status: 200,
				headers: { 'Content-Type': 'application/json' }
			}))
			.mockResolvedValueOnce(new Response(JSON.stringify({ data: { ok: true } }), { status: 200 }));
		const request = new Request('http://127.0.0.1:8080/api/document-cycles/catalogs', {
			method: 'POST',
			headers: {
				Authorization: `Bearer ${token('access')}`,
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ title: 'Dokumen' })
		});
		const event = makeEvent();

		const response = await handleFetch({ event, request, fetch: routeFetch } as never);

		expect(response.status).toBe(200);
		expect(globalFetch).not.toHaveBeenCalled();
		expect(routeFetch).toHaveBeenCalledTimes(3);
		const retryRequest = routeFetch.mock.calls[2]?.[0] as Request;
		expect(retryRequest.headers.get('Authorization')).toBe(`Bearer ${nextAccess}`);
		await expect(retryRequest.text()).resolves.toBe('{"title":"Dokumen"}');
	});

	it('does not refresh or clone unauthenticated 401 requests', async () => {
		const handleFetch = await loadHandleFetch();
		const globalFetch = vi.spyOn(globalThis, 'fetch');
		const cloneSpy = vi.spyOn(Request.prototype, 'clone');
		const routeFetch = vi.fn<typeof fetch>()
			.mockImplementationOnce(async (input) => {
				await (input as Request).text();
				return new Response(JSON.stringify({ error: 'unauthorized' }), { status: 401 });
			});
		const request = new Request('http://localhost/api/public/register-student', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ nama: 'Siswa' })
		});
		const event = makeEvent();

		const response = await handleFetch({ event, request, fetch: routeFetch } as never);

		expect(response.status).toBe(401);
		expect(routeFetch).toHaveBeenCalledTimes(1);
		expect(globalFetch).not.toHaveBeenCalled();
		expect(cloneSpy).not.toHaveBeenCalled();
	});

	it('does not refresh or retry authorized 401 requests outside the core API origin', async () => {
		const handleFetch = await loadHandleFetch();
		const globalFetch = vi.spyOn(globalThis, 'fetch');
		const routeFetch = vi.fn<typeof fetch>()
			.mockResolvedValueOnce(new Response(JSON.stringify({ error: 'unauthorized' }), { status: 401 }));
		const request = new Request('https://external.example.test/private-api', {
			headers: { Authorization: 'Bearer third-party-token' }
		});
		const event = makeEvent();

		const response = await handleFetch({ event, request, fetch: routeFetch } as never);

		expect(response.status).toBe(401);
		expect(routeFetch).toHaveBeenCalledTimes(1);
		expect(globalFetch).not.toHaveBeenCalled();
		expect(event.cookies.set).not.toHaveBeenCalled();
	});

	it('does not refresh non-bearer authorization failures from the core API', async () => {
		const handleFetch = await loadHandleFetch();
		const globalFetch = vi.spyOn(globalThis, 'fetch');
		const routeFetch = vi.fn<typeof fetch>()
			.mockResolvedValueOnce(new Response(JSON.stringify({ error: 'unauthorized' }), { status: 401 }));
		const request = new Request('http://127.0.0.1:8080/api/document-cycles/stats', {
			headers: { Authorization: 'Basic operator-secret' }
		});
		const event = makeEvent();

		const response = await handleFetch({ event, request, fetch: routeFetch } as never);

		expect(response.status).toBe(401);
		expect(routeFetch).toHaveBeenCalledTimes(1);
		expect(globalFetch).not.toHaveBeenCalled();
		expect(event.cookies.set).not.toHaveBeenCalled();
	});

	it('keeps the original 401 and clears cookies when refresh fails', async () => {
		const handleFetch = await loadHandleFetch();
		const globalFetch = vi.spyOn(globalThis, 'fetch');
		const routeFetch = vi.fn<typeof fetch>()
			.mockResolvedValue(new Response(JSON.stringify({ error: 'unauthorized' }), { status: 401 }));
		const request = new Request('http://127.0.0.1:8080/api/document-cycles/stats', {
			headers: { Authorization: `Bearer ${token('access')}` }
		});
		const event = makeEvent();

		const response = await handleFetch({ event, request, fetch: routeFetch } as never);

		expect(response.status).toBe(401);
		expect(globalFetch).not.toHaveBeenCalled();
		expect(routeFetch).toHaveBeenCalledTimes(2);
		expect(event.cookies.delete).toHaveBeenCalledWith('access_token', { path: '/' });
		expect(event.cookies.delete).toHaveBeenCalledWith('refresh_token', { path: '/' });
	});

	it('shares a single refresh across concurrent 401 retries for one request event', async () => {
		const handleFetch = await loadHandleFetch();
		const nextAccess = token('access', { uid: 'u1', role: 'admin' });
		const nextRefresh = token('refresh');
		const globalFetch = vi.spyOn(globalThis, 'fetch');
		const routeFetch = vi.fn<typeof fetch>(async (input) => {
			if (callUrl(input).includes('/api/auth/refresh')) {
				return new Response(JSON.stringify({ data: { access_token: nextAccess, refresh_token: nextRefresh } }), {
					status: 200,
					headers: { 'Content-Type': 'application/json' }
				});
			}
			const request = input as Request;
			if (request.headers.get('Authorization') === `Bearer ${nextAccess}`) {
				return new Response(JSON.stringify({ data: { ok: true } }), { status: 200 });
			}
			return new Response(JSON.stringify({ error: 'unauthorized' }), { status: 401 });
		});
		const event = makeEvent();

		const [first, second] = await Promise.all([
			handleFetch({
				event,
				request: new Request('http://127.0.0.1:8080/api/document-cycles/stats', {
					headers: { Authorization: `Bearer ${token('access')}` }
				}),
				fetch: routeFetch
			} as never),
			handleFetch({
				event,
				request: new Request('http://127.0.0.1:8080/api/document-cycles/catalogs', {
					headers: { Authorization: `Bearer ${token('access')}` }
				}),
				fetch: routeFetch
			} as never)
		]);

		expect(first.status).toBe(200);
		expect(second.status).toBe(200);
		expect(globalFetch).not.toHaveBeenCalled();
		expect(routeFetch).toHaveBeenCalledTimes(5);
		expect(event.locals.accessToken).toBe(nextAccess);
	});
});

describe('SvelteKit handle auth gate', () => {
	afterEach(() => {
		vi.restoreAllMocks();
	});

	it('does not authenticate an expired access token without a refresh token', async () => {
		const { handle } = await loadHooks();
		const expiredAccess = token('access', {
			uid: 'u1',
			role: 'admin',
			exp: Math.floor(Date.now() / 1000) - 60
		});
		const event = makeHandleEvent(expiredAccess);
		const resolve = vi.fn(async () => new Response('ok'));

		await expect(handle({ event, resolve } as never)).rejects.toMatchObject({
			status: 302,
			location: '/login?from=%2Fdashboard'
		});
		expect(resolve).not.toHaveBeenCalled();
		expect(event.locals.accessToken).toBeUndefined();
		expect(event.locals.user).toBeUndefined();
	});

	it('returns JSON 401 for unauthenticated protected API requests', async () => {
		const { handle } = await loadHooks();
		const event = makeHandleEvent(undefined, undefined, 'http://localhost/api/school-profile');
		const resolve = vi.fn(async () => new Response('ok'));

		const response = await handle({ event, resolve } as never);

		expect(response.status).toBe(401);
		expect(response.headers.get('Content-Type')).toBe('application/json');
		await expect(response.json()).resolves.toEqual({ error: 'unauthorized' });
		expect(resolve).not.toHaveBeenCalled();
	});

	it('does not mark Bank Soal or Asesmen API aliases as deprecated', async () => {
		const { handle } = await loadHooks();
		const aliases = ['/api/bank-soal/questions', '/api/asesmen/events'];

		for (const path of aliases) {
			const access = token('access', { uid: 'u1', role: 'admin', roles: ['admin'] });
			const event = makeHandleEvent(access, undefined, `http://localhost${path}`);
			event.fetch.mockResolvedValueOnce(validationOkResponse());
			const resolve = vi.fn(async () => new Response(JSON.stringify({ ok: true }), {
				status: 200,
				headers: { 'Content-Type': 'application/json' }
			}));

			const response = await handle({ event, resolve } as never);

			expect(response.status).toBe(200);
			expect(response.headers.get('Deprecation')).toBeNull();
			expect(response.headers.get('Warning')).toBeNull();
			expect(response.headers.get('Link')).toBeNull();
			await expect(response.json()).resolves.toEqual({ ok: true });
		}
	});

	it('refreshes an expired access token through the event fetch boundary before resolving', async () => {
		const { handle } = await loadHooks();
		const expiredAccess = token('access', {
			uid: 'u1',
			role: 'admin',
			exp: Math.floor(Date.now() / 1000) - 60
		});
		const nextAccess = token('access', { uid: 'u1', role: 'admin' });
		const nextRefresh = token('refresh');
		const event = makeHandleEvent(expiredAccess, token('refresh'));
		const eventFetch = event.fetch;
		eventFetch.mockResolvedValueOnce(new Response(JSON.stringify({
			data: { access_token: nextAccess, refresh_token: nextRefresh }
		}), {
			status: 200,
			headers: { 'Content-Type': 'application/json' }
		})).mockResolvedValueOnce(validationOkResponse());
		const resolve = vi.fn(async () => new Response('ok'));

		const response = await handle({ event, resolve } as never);

		expect(response.status).toBe(200);
		expect(eventFetch).toHaveBeenCalledTimes(2);
		const refreshUrl = String(eventFetch.mock.calls[0]?.[0]);
		expect(refreshUrl).toContain('/api/auth/refresh');
		expect(String(eventFetch.mock.calls[1]?.[0])).toContain('/api/auth/sessions');
		expect(resolve).toHaveBeenCalled();
		expect(event.locals.accessToken).toBe(nextAccess);
		expect(event.locals.user).toMatchObject({ id: 'u1', role: 'admin', roles: ['admin'] });
	});

	it('does not trust an unsigned access-token payload when backend validation rejects it', async () => {
		const { handle } = await loadHooks();
		const spoofedAccess = token('access', { uid: 'u1', role: 'admin' });
		const event = makeHandleEvent(spoofedAccess, undefined, 'http://localhost/settings/users');
		event.fetch.mockResolvedValueOnce(new Response(JSON.stringify({ error: 'unauthorized' }), { status: 401 }));
		const resolve = vi.fn(async () => new Response('ok'));

		await expect(handle({ event, resolve } as never)).rejects.toMatchObject({
			status: 302,
			location: '/login?from=%2Fsettings%2Fusers'
		});
		expect(resolve).not.toHaveBeenCalled();
		expect(event.locals.user).toBeUndefined();
		expect(event.cookies.delete).toHaveBeenCalledWith('access_token', { path: '/' });
	});

	it('surfaces validation outages on protected routes without clearing cookies', async () => {
		const { handle } = await loadHooks();
		const access = token('access', { uid: 'u1', role: 'admin' });
		const refresh = token('refresh');
		const event = makeHandleEvent(access, refresh, 'http://localhost/settings/users');
		event.fetch.mockResolvedValueOnce(new Response(JSON.stringify({ error: 'temporarily unavailable' }), { status: 503 }));
		const resolve = vi.fn(async () => new Response('ok'));

		await expect(handle({ event, resolve } as never)).rejects.toMatchObject({
			status: 503,
			body: { message: 'Layanan validasi sesi sedang bermasalah. Silakan coba beberapa saat lagi.' }
		});
		expect(resolve).not.toHaveBeenCalled();
		expect(event.cookies.delete).not.toHaveBeenCalled();
	});

	it('does not clear cookies when auth validation cannot reach the backend', async () => {
		const { handle } = await loadHooks();
		const access = token('access', { uid: 'u1', role: 'admin' });
		const event = makeHandleEvent(access, undefined, 'http://localhost/dashboard');
		event.fetch.mockRejectedValueOnce(new TypeError('connect ECONNREFUSED'));
		const resolve = vi.fn(async () => new Response('ok'));

		await expect(handle({ event, resolve } as never)).rejects.toMatchObject({
			status: 503,
			body: { message: 'Layanan validasi sesi sedang bermasalah. Silakan coba beberapa saat lagi.' }
		});
		expect(resolve).not.toHaveBeenCalled();
		expect(event.cookies.delete).not.toHaveBeenCalled();
	});

	it('redirects first-login users away from portal pages until password change', async () => {
		const { handle } = await loadHooks();
		const access = token('access', {
			uid: 'u1',
			role: 'siswa',
			roles: ['siswa'],
			permissions: ['student_portal.read'],
			must_change_password: true
		});
		const event = makeHandleEvent(access, undefined, 'http://localhost/portal/siswa');
		event.fetch.mockResolvedValueOnce(validationOkResponse());
		const resolve = vi.fn(async () => new Response('ok'));

		await expect(handle({ event, resolve } as never)).rejects.toMatchObject({
			status: 302,
			location: '/settings/account'
		});
		expect(resolve).not.toHaveBeenCalled();
	});

	it('allows first-login users to open account settings and blocks other protected APIs', async () => {
		const { handle } = await loadHooks();
		const access = token('access', {
			uid: 'u1',
			role: 'siswa',
			roles: ['siswa'],
			permissions: ['student_portal.read'],
			must_change_password: true
		});
		const settingsEvent = makeHandleEvent(access, undefined, 'http://localhost/settings/account');
		settingsEvent.fetch.mockResolvedValueOnce(validationOkResponse());
		const resolve = vi.fn(async () => new Response('ok'));

		const response = await handle({ event: settingsEvent, resolve } as never);
		expect(response.status).toBe(200);
		expect(resolve).toHaveBeenCalled();

		const apiEvent = makeHandleEvent(access, undefined, 'http://localhost/api/portal/siswa/profile');
		apiEvent.fetch.mockResolvedValueOnce(validationOkResponse());
		await expect(handle({ event: apiEvent, resolve: vi.fn(async () => new Response('ok')) } as never)).rejects.toMatchObject({
			status: 403
		});
	});

	it('blocks protected API mutations without a same-origin Origin or Referer', async () => {
		const { handle } = await loadHooks();
		const access = token('access', { uid: 'u1', role: 'admin', roles: ['admin'] });
		const event = makeHandleEvent(access, undefined, 'http://localhost/api/school-profile', 'POST');
		event.fetch.mockResolvedValueOnce(validationOkResponse());
		const resolve = vi.fn(async () => new Response('ok'));

		await expect(handle({ event, resolve } as never)).rejects.toMatchObject({
			status: 403,
			body: { message: 'csrf validation failed' }
		});
		expect(resolve).not.toHaveBeenCalled();
	});

	it('allows protected API mutations with a same-origin Origin header', async () => {
		const { handle } = await loadHooks();
		const access = token('access', { uid: 'u1', role: 'admin', roles: ['admin'] });
		const event = makeHandleEvent(access, undefined, 'http://localhost/api/school-profile', 'POST', { origin: 'http://localhost' });
		event.fetch.mockResolvedValueOnce(validationOkResponse());
		const resolve = vi.fn(async () => new Response('ok'));

		const response = await handle({ event, resolve } as never);

		expect(response.status).toBe(200);
		expect(resolve).toHaveBeenCalled();
	});

	const roleCases = [
		{ path: '/settings/account', roles: ['guru'], allowed: true, api: false },
		{ path: '/api/auth/account', roles: ['guru'], allowed: true, api: true },
		{ path: '/settings/users', roles: ['admin'], allowed: true, api: false },
		{ path: '/settings/users', roles: ['staf'], allowed: false, api: false },
		{ path: '/api/school-profile', roles: ['admin'], allowed: true, api: true },
		{ path: '/api/school-profile', roles: ['staf'], allowed: false, api: true },
		{ path: '/asesmen/kegiatan', roles: ['admin'], allowed: true, api: false },
		{ path: '/asesmen/kegiatan/event-1', roles: ['guru'], allowed: false, api: false },
		{ path: '/api/asesmen/events', roles: ['admin'], allowed: true, api: true },
		{ path: '/api/asesmen/events', roles: ['guru'], allowed: false, api: true, method: 'GET' },
		{ path: '/api/asesmen/events', roles: ['guru'], allowed: false, api: true, method: 'POST' },
		{ path: '/api/asesmen/events/event-1', roles: ['guru'], allowed: false, api: true },
		{ path: '/api/asesmen/events/event-1/question-targets', roles: ['guru'], allowed: false, api: true, method: 'GET' },
		{ path: '/api/asesmen/events/event-1/question-targets', roles: ['guru'], allowed: false, api: true, method: 'PUT' },
		{ path: '/asesmen/kegiatan/event-1/members', roles: ['guru'], allowed: false, api: false },
		{ path: '/bank-soal', roles: ['guru'], allowed: false, api: false },
		{ path: '/bank-soal/verifikasi', roles: ['guru'], allowed: false, api: false },
		{ path: '/bank-soal/impor', roles: ['guru'], allowed: false, api: false },
		{ path: '/api/bank-soal/soal-support/subjects', roles: ['guru'], allowed: false, api: true, method: 'GET' },
		{ path: '/asesmen/paket', roles: ['guru'], allowed: false, api: false },
		{ path: '/asesmen/sesi', roles: ['guru'], allowed: false, api: false },
		{ path: '/api/asesmen/packages', roles: ['guru'], allowed: false, api: true, method: 'GET' },
		{ path: '/api/asesmen/packages', roles: ['guru'], allowed: false, api: true, method: 'POST' },
		{ path: '/api/asesmen/sessions', roles: ['guru'], allowed: false, api: true, method: 'GET' },
		{ path: '/api/asesmen/sessions', roles: ['guru'], allowed: false, api: true, method: 'POST' },
		{ path: '/bank-soal/tambah', roles: ['guru'], allowed: false, api: false },
		{ path: '/api/bank-soal/questions', roles: ['guru'], allowed: false, api: true, method: 'GET' },
		{ path: '/api/bank-soal/questions', roles: ['guru'], allowed: false, api: true, method: 'POST' },
		{ path: '/parents', roles: ['admin'], allowed: true, api: false },
		{ path: '/parents', roles: ['staf'], allowed: false, api: false },
		{ path: '/api/parents', roles: ['admin'], allowed: true, api: true },
		{ path: '/api/parents', roles: ['guru'], allowed: false, api: true },
		{ path: '/kesiswaan', roles: ['kesiswaan'], allowed: false, api: false },
		{ path: '/kesiswaan', roles: ['guru'], allowed: false, api: false },
		{ path: '/kesiswaan', roles: ['staf'], allowed: false, api: false },
		{ path: '/api/kesiswaan/students', roles: ['guru'], allowed: false, api: true, method: 'GET' },
		{ path: '/api/kesiswaan/students', roles: ['guru'], allowed: false, api: true, method: 'POST' },
		{ path: '/api/kesiswaan/students', roles: ['kesiswaan'], allowed: false, api: true, method: 'POST' },
		{ path: '/students', roles: ['kesiswaan'], allowed: false, api: false },
		{ path: '/students', roles: ['guru'], allowed: false, api: false },
		{ path: '/students', roles: ['staf'], allowed: false, api: false },
		{ path: '/api/students', roles: ['guru'], allowed: false, api: true, method: 'GET' },
		{ path: '/api/students', roles: ['siswa'], allowed: false, api: true, method: 'GET' },
		{ path: '/api/students', roles: ['staf'], allowed: false, api: true, method: 'GET' },
		{ path: '/api/students', roles: ['guru'], allowed: false, api: true, method: 'POST' },
		{ path: '/api/students', roles: ['kesiswaan'], allowed: false, api: true, method: 'POST' },
		{ path: '/portal/siswa', roles: ['siswa'], allowed: true, api: false },
		{ path: '/portal/siswa', roles: ['ortu'], allowed: false, api: false },
		{ path: '/portal/orang-tua', roles: ['ortu'], allowed: true, api: false },
		{ path: '/portal/orang-tua', roles: ['siswa'], allowed: false, api: false },
		{ path: '/api/portal/siswa/profile', roles: ['siswa'], allowed: true, api: true },
		{ path: '/api/portal/siswa/results', roles: ['ortu'], allowed: false, api: true },
		{ path: '/api/scheduler/tick', roles: ['staf'], allowed: false, api: true },
		{ path: '/api/scheduler/tick', roles: ['admin'], allowed: true, api: true },
	] as const;

	for (const item of roleCases) {
		it(`enforces hook role access for ${item.path} as ${item.roles.join(',')}`, async () => {
			const { handle } = await loadHooks();
			const access = token('access', { uid: 'u1', role: item.roles[0], roles: item.roles });
			const method = 'method' in item ? item.method : undefined;
			const event = makeHandleEvent(access, undefined, `http://localhost${item.path}`, method);
			event.fetch.mockResolvedValueOnce(validationOkResponse());
			const resolve = vi.fn(async () => new Response('ok'));

			if (item.allowed) {
				const response = await handle({ event, resolve } as never);
				expect(response.status).toBe(200);
				expect(resolve).toHaveBeenCalled();
				return;
			}

			await expect(handle({ event, resolve } as never)).rejects.toMatchObject(item.api
				? { status: 403 }
				: { status: 302, location: '/' });
			expect(resolve).not.toHaveBeenCalled();
		});
	}
});
