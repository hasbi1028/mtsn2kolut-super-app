import { beforeEach, describe, expect, it, vi } from 'vitest';

const proxyGetMock = vi.fn();
const proxyPostMock = vi.fn();
const proxyPutMock = vi.fn();
const proxyPatchMock = vi.fn();
const proxyDeleteMock = vi.fn();
const proxyFetchMock = vi.fn();
const apiPublicPostWithFetchMock = vi.fn();
const proxyMock = vi.fn(() => ({
	get: proxyGetMock,
	post: proxyPostMock,
	put: proxyPutMock,
	patch: proxyPatchMock,
	del: proxyDeleteMock,
	fetch: proxyFetchMock
}));
const readRequestJsonMock = vi.fn(async <T>(request: Request) => request.json() as Promise<T>);
const readOptionalRequestJsonMock = vi.fn(async <T>(request: Request, fallback: T) => {
	const text = await request.text();
	return text ? JSON.parse(text) as T : fallback;
});
const readProxyJsonMock = vi.fn(async <T>(response: Response) => response.json() as Promise<T>);
const jsonProxyResponseMock = vi.fn(async (response: Response) => response);
const streamProxyResponseMock = vi.fn(async (response: Response) => response);
const handleRouteErrorMock = vi.fn((error: unknown, route: string) => new Response(JSON.stringify({
	error: (error as Error)?.message ?? 'failed',
	route
}), { status: 503 }));
const apiPathMock = (strings: TemplateStringsArray, ...values: Array<string | number | boolean>) => {
	let path = strings[0] ?? '';
	for (let i = 0; i < values.length; i += 1) {
		path += encodeURIComponent(String(values[i]));
		path += strings[i + 1] ?? '';
	}
	return path;
};
const apiPathWithQueryMock = (path: string, params: URLSearchParams | string) => {
	const query = typeof params === 'string' ? params.replace(/^\?/, '') : params.toString();
	return query ? `${path}?${query}` : path;
};
const backendCbtPrefix = ['/api', 'cbt'].join('/');
const backendCbtPath = (path: string) => `${backendCbtPrefix}${path.startsWith('/') ? path : `/${path}`}`;
const backendCbtApiPath = (strings: TemplateStringsArray, ...values: Array<string | number | boolean>) =>
	`${backendCbtPrefix}${apiPathMock(strings, ...values)}`;
const requiredRouteParamMock = (value: string | undefined, name: string) => {
	if (!value) throw new MockApiError(400, `${name} tidak valid`);
	return value;
};

class MockApiError extends Error {
	constructor(public status: number, message: string) {
		super(message);
	}
}

vi.mock('$lib/server/api', () => ({
	proxy: proxyMock,
	apiPublicPostWithFetch: apiPublicPostWithFetchMock,
	readRequestJson: readRequestJsonMock,
	readOptionalRequestJson: readOptionalRequestJsonMock,
	readProxyJson: readProxyJsonMock,
	jsonProxyResponse: jsonProxyResponseMock,
	streamProxyResponse: streamProxyResponseMock,
	handleRouteError: handleRouteErrorMock,
	apiPath: apiPathMock,
	apiPathWithQuery: apiPathWithQueryMock,
	requiredRouteParam: requiredRouteParamMock,
	ApiError: MockApiError
}));

function createEvent(overrides: Record<string, unknown> = {}) {
	const deleted: Array<{ key: string; opts: { path: string } }> = [];
	const cookieValues = (overrides.cookieValues ?? {}) as Record<string, string | undefined>;
	const eventOverrides = { ...overrides };
	delete eventOverrides.cookieValues;
	return {
		locals: {},
		params: {},
		request: new Request('http://localhost/test'),
		fetch: vi.fn(),
		cookies: {
			get: (key: string) => cookieValues[key],
			delete: (key: string, opts: { path: string }) => {
				deleted.push({ key, opts });
			}
		},
		deletedCookies: deleted,
		...eventOverrides
	};
}

describe('api proxy route handlers', () => {
	beforeEach(() => {
		proxyMock.mockClear();
		proxyGetMock.mockReset();
		proxyPostMock.mockReset();
		proxyPutMock.mockReset();
		proxyPatchMock.mockReset();
		proxyDeleteMock.mockReset();
		proxyFetchMock.mockReset();
		apiPublicPostWithFetchMock.mockReset();
		readRequestJsonMock.mockClear();
		readOptionalRequestJsonMock.mockClear();
		readProxyJsonMock.mockClear();
		jsonProxyResponseMock.mockClear();
		streamProxyResponseMock.mockClear();
		handleRouteErrorMock.mockClear();
	});

	it('rejects unauthenticated auth sessions requests', async () => {
		const mod = await import('../../routes/api/auth/sessions/+server');
		const event = createEvent();

		const res = await mod.GET(event as never);

		expect(res.status).toBe(401);
		await expect(res.json()).resolves.toEqual({ error: 'Unauthorized' });
		expect(proxyGetMock).not.toHaveBeenCalled();
	});

	it('forwards auth sessions requests through authenticated proxy', async () => {
		const mod = await import('../../routes/api/auth/sessions/+server');
		const event = createEvent({
			locals: {
				user: { id: '1', username: 'admin', role: 'admin', roles: ['admin'] }
			}
		});
		proxyGetMock.mockResolvedValueOnce([{ id: 'sess-1' }]);

		const res = await mod.GET(event as never);

		expect(proxyGetMock).toHaveBeenCalledWith('/api/auth/sessions');
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual([{ id: 'sess-1' }]);
	});

	it('rejects unauthenticated account summary requests', async () => {
		const mod = await import('../../routes/api/auth/account/+server');
		const event = createEvent();

		const res = await mod.GET(event as never);

		expect(res.status).toBe(401);
		await expect(res.json()).resolves.toEqual({ error: 'Unauthorized' });
		expect(proxyGetMock).not.toHaveBeenCalled();
	});

	it('forwards account summary requests through authenticated proxy', async () => {
		const mod = await import('../../routes/api/auth/account/+server');
		const event = createEvent({
			locals: {
				user: { id: '1', username: 'guru.ipa', role: 'guru', roles: ['guru'] }
			}
		});
		proxyGetMock.mockResolvedValueOnce({ username: 'guru.ipa', roles: ['guru'] });

		const res = await mod.GET(event as never);

		expect(proxyGetMock).toHaveBeenCalledWith('/api/auth/account');
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ username: 'guru.ipa', roles: ['guru'] });
	});

	it('forwards account contact updates without a user id parameter', async () => {
		const mod = await import('../../routes/api/auth/account/+server');
		const request = new Request('http://localhost/api/auth/account', {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ phone: '0812', email: 'guru@example.id', address: 'Kolaka Utara' })
		});
		const event = createEvent({
			locals: {
				user: { id: '1', username: 'guru.ipa', role: 'guru', roles: ['guru'] }
			},
			request
		});
		proxyPatchMock.mockResolvedValueOnce({ username: 'guru.ipa', contact: { phone: '0812' } });

		const res = await mod.PATCH(event as never);

		expect(readRequestJsonMock).toHaveBeenCalledWith(request);
		expect(proxyPatchMock).toHaveBeenCalledWith('/api/auth/account/contact', {
			phone: '0812',
			email: 'guru@example.id',
			address: 'Kolaka Utara'
		});
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ username: 'guru.ipa', contact: { phone: '0812' } });
	});

	it('rejects account contact updates that try to edit official fields', async () => {
		const mod = await import('../../routes/api/auth/account/+server');
		const request = new Request('http://localhost/api/auth/account', {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ username: 'baru', role: 'admin' })
		});
		const event = createEvent({
			locals: {
				user: { id: '1', username: 'guru.ipa', role: 'guru', roles: ['guru'] }
			},
			request
		});

		const res = await mod.PATCH(event as never);

		expect(proxyPatchMock).not.toHaveBeenCalled();
		expect(res.status).toBe(400);
		await expect(res.json()).resolves.toEqual({ error: 'Field username tidak dapat diubah dari akun saya' });
	});

	it('renames an auth session through the authenticated proxy', async () => {
		const mod = await import('../../routes/api/auth/sessions/[id]/+server');
		const request = new Request('http://localhost/api/auth/sessions/sess%201%2F2026', {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ device_label: 'Laptop Operator' })
		});
		const event = createEvent({
			params: { id: 'sess 1/2026' },
			locals: {
				user: { id: '1', username: 'admin', role: 'admin', roles: ['admin'] }
			},
			request
		});

		const res = await mod.PATCH(event as never);

		expect(readRequestJsonMock).toHaveBeenCalledWith(request);
		expect(proxyPatchMock).toHaveBeenCalledWith('/api/auth/sessions/sess%201%2F2026', {
			device_label: 'Laptop Operator'
		});
		expect(proxyFetchMock).not.toHaveBeenCalled();
		expect(proxyPostMock).not.toHaveBeenCalled();
		expect(proxyGetMock).not.toHaveBeenCalled();
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ ok: true });
	});

	it('revokes an auth session through the authenticated proxy', async () => {
		const mod = await import('../../routes/api/auth/sessions/[id]/+server');
		proxyDeleteMock.mockResolvedValueOnce(undefined);
		const event = createEvent({
			params: { id: 'sess-2' },
			locals: {
				user: { id: '1', username: 'admin', role: 'admin', roles: ['admin'] }
			}
		});

		const res = await mod.DELETE(event as never);

		expect(proxyDeleteMock).toHaveBeenCalledWith('/api/auth/sessions/sess-2');
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ ok: true });
	});

	it('logout revokes the refresh session through the SvelteKit fetch boundary and clears cookies', async () => {
		const mod = await import('../../routes/api/auth/logout/+server');
		const eventFetch = vi.fn<typeof fetch>();
		const event = createEvent({
			cookieValues: { refresh_token: 'refresh-1' },
			fetch: eventFetch
		});
		apiPublicPostWithFetchMock.mockResolvedValueOnce(undefined);

		const res = await mod.POST(event as never);

		expect(apiPublicPostWithFetchMock).toHaveBeenCalledWith(eventFetch, '/api/auth/logout', {
			refresh_token: 'refresh-1'
		});
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ ok: true });
		expect((event as ReturnType<typeof createEvent>).deletedCookies).toEqual([
			{ key: 'access_token', opts: { path: '/' } },
			{ key: 'refresh_token', opts: { path: '/' } }
		]);
	});

	it('logout-all revokes backend session and clears auth cookies', async () => {
		const mod = await import('../../routes/api/auth/logout-all/+server');
		const event = createEvent({
			locals: {
				user: { id: '1', username: 'admin', role: 'admin', roles: ['admin'] }
			}
		});
		proxyPostMock.mockResolvedValueOnce(undefined);

		const res = await mod.POST(event as never);

		expect(proxyPostMock).toHaveBeenCalledWith('/api/auth/logout-all', {});
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ ok: true });
		expect((event as ReturnType<typeof createEvent>).deletedCookies).toEqual([
			{ key: 'access_token', opts: { path: '/' } },
			{ key: 'refresh_token', opts: { path: '/' } }
		]);
	});

	it('maps upstream unauthorized logout-all responses back to 401', async () => {
		const mod = await import('../../routes/api/auth/logout-all/+server');
		const event = createEvent({
			locals: {
				user: { id: '1', username: 'admin', role: 'admin', roles: ['admin'] }
			}
		});
		proxyPostMock.mockRejectedValueOnce(new MockApiError(401, 'unauthorized'));

		const res = await mod.POST(event as never);

		expect(res.status).toBe(401);
		await expect(res.json()).resolves.toEqual({ error: 'Unauthorized' });
	});

	it('validates change-password payload before proxying', async () => {
		const mod = await import('../../routes/api/auth/change-password/+server');
		const request = new Request('http://localhost/api/auth/change-password', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ current_password: '', new_password: 'short' })
		});
		const event = createEvent({
			locals: {
				user: { id: '1', username: 'admin', role: 'admin', roles: ['admin'] }
			},
			request
		});

		const res = await mod.POST(event as never);

		expect(readRequestJsonMock).toHaveBeenCalledWith(request);
		expect(proxyPostMock).not.toHaveBeenCalled();
		expect(res.status).toBe(400);
		await expect(res.json()).resolves.toEqual({ error: 'Field tidak boleh kosong' });
	});

	it('forwards change-password requests with authenticated username context', async () => {
		const mod = await import('../../routes/api/auth/change-password/+server');
		const request = new Request('http://localhost/api/auth/change-password', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ current_password: 'lama-sekali', new_password: 'baru-kuat-123' })
		});
		const event = createEvent({
			locals: {
				user: { id: '1', username: 'admin', role: 'admin', roles: ['admin'] }
			},
			request
		});
		proxyPostMock.mockResolvedValueOnce(undefined);

		const res = await mod.POST(event as never);

		expect(readRequestJsonMock).toHaveBeenCalledWith(request);
		expect(proxyPostMock).toHaveBeenCalledWith('/api/auth/change-password', {
			username: 'admin',
			old_password: 'lama-sekali',
			new_password: 'baru-kuat-123'
		});
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ ok: true });
	});

	it('maps suspended change-password responses to 403', async () => {
		const mod = await import('../../routes/api/auth/change-password/+server');
		const event = createEvent({
			locals: {
				user: { id: '1', username: 'admin', role: 'admin', roles: ['admin'] }
			},
			request: new Request('http://localhost/api/auth/change-password', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ current_password: 'lama-sekali', new_password: 'baru-kuat-123' })
			})
		});
		proxyPostMock.mockRejectedValueOnce(new MockApiError(403, 'suspended'));

		const res = await mod.POST(event as never);

		expect(res.status).toBe(403);
		await expect(res.json()).resolves.toEqual({ error: 'Akun sedang dinonaktifkan' });
	});

	it('forwards sidebar preference reads for authenticated users', async () => {
		const mod = await import('../../routes/api/auth/preferences/sidebar/+server');
		const event = createEvent({
			locals: {
				user: { id: '1', username: 'admin', role: 'admin', roles: ['admin'] }
			}
		});
		proxyGetMock.mockResolvedValueOnce({ collapsed: true });

		const res = await mod.GET(event as never);

		expect(proxyGetMock).toHaveBeenCalledWith('/api/auth/preferences/sidebar');
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ collapsed: true });
	});

	it('forwards sidebar preference updates for authenticated users', async () => {
		const mod = await import('../../routes/api/auth/preferences/sidebar/+server');
		proxyPatchMock.mockResolvedValueOnce({ collapsed: false });
		const event = createEvent({
			locals: {
				user: { id: '1', username: 'admin', role: 'admin', roles: ['admin'] }
			},
			request: new Request('http://localhost/api/auth/preferences/sidebar', {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ collapsed: false })
			})
		});

		const res = await mod.PATCH(event as never);

		expect(proxyPatchMock).toHaveBeenCalledWith('/api/auth/preferences/sidebar', { collapsed: false });
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ collapsed: false });
	});

	it('sanitizes blocked PUSAKA settings from backend responses', async () => {
		const mod = await import('../../routes/api/pusaka/settings/+server');
		const event = createEvent();
		proxyGetMock.mockResolvedValueOnce([
			{ key: 'max_concurrent', value: '9' },
			{ key: 'headless', value: 'true' },
			{ key: 'pusaka_geo_base_lat', value: '-3.2163111' },
			{ key: 'pusaka_geo_checkin_radius_m', value: '55' },
			{ key: 'admin_password', value: 'secret' },
			{ key: 'browser_path', value: '/usr/bin/chromium' }
		]);

		const res = await mod.GET(event as never);

		expect(proxyGetMock).toHaveBeenCalledWith('/api/pusaka/settings');
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({
			max_concurrent: 9,
			headless: true,
			pusaka_geo_base_lat: -3.2163111,
			pusaka_geo_checkin_radius_m: 55,
			browser_path: '/usr/bin/chromium'
		});
	});

	it('fans out PUSAKA setting updates per key', async () => {
		const mod = await import('../../routes/api/pusaka/settings/+server');
		proxyPutMock.mockResolvedValue(undefined);
		const event = createEvent({
			request: new Request('http://localhost/api/pusaka/settings', {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ max_concurrent: 7, headless: false, pusaka_geo_checkin_radius_m: 60 })
			})
		});

		const res = await mod.PUT(event as never);

		expect(proxyPutMock).toHaveBeenNthCalledWith(1, '/api/pusaka/settings/max_concurrent', {
			value: '7'
		});
		expect(proxyPutMock).toHaveBeenNthCalledWith(2, '/api/pusaka/settings/headless', {
			value: 'false'
		});
		expect(proxyPutMock).toHaveBeenNthCalledWith(3, '/api/pusaka/settings/pusaka_geo_checkin_radius_m', {
			value: '60'
		});
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ ok: true, max_concurrent: 7, headless: false, pusaka_geo_checkin_radius_m: 60 });
	});

	it('rejects blocked and unknown PUSAKA setting updates', async () => {
		const mod = await import('../../routes/api/pusaka/settings/+server');
		const event = createEvent({
			request: new Request('http://localhost/api/pusaka/settings', {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ admin_password: 'secret', 'bad/key': 'value' })
			})
		});

		const res = await mod.PUT(event as never);

		expect(proxyPutMock).not.toHaveBeenCalled();
		expect(res.status).toBe(400);
		await expect(res.json()).resolves.toMatchObject({
			error: expect.stringContaining('tidak dapat diubah')
		});
	});

	it('creates document cycle catalogs through the typed JSON body helper', async () => {
		const mod = await import('../../routes/api/document-cycles/catalogs/+server');
		proxyPostMock.mockResolvedValueOnce({ id: 'catalog-1', code: 'RKT' });
		const request = new Request('http://localhost/api/document-cycles/catalogs', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ code: 'RKT', title: 'RKT Tahunan' })
		});
		const event = createEvent({
			request
		});

		const res = await mod.POST(event as never);

		expect(readRequestJsonMock).toHaveBeenCalledWith(request);
		expect(proxyPostMock).toHaveBeenCalledWith('/api/document-cycles/catalogs', {
			code: 'RKT',
			title: 'RKT Tahunan'
		});
		expect(res.status).toBe(201);
		await expect(res.json()).resolves.toEqual({ id: 'catalog-1', code: 'RKT' });
	});

	it('forwards document verification queue filters through the authenticated proxy', async () => {
		const mod = await import('../../routes/api/document-cycles/verification-queue/+server');
		const event = createEvent({
			url: new URL('http://localhost/api/document-cycles/verification-queue?period_year=2026&all=true')
		});
		proxyGetMock.mockResolvedValueOnce([{ id: 'obligation-1', status: 'waiting_verification' }]);

		const res = await mod.GET(event as never);

		expect(proxyGetMock).toHaveBeenCalledWith('/api/document-cycles/verification-queue?period_year=2026&all=true');
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual([{ id: 'obligation-1', status: 'waiting_verification' }]);
	});

	it('does not append a blank query marker when list filters are empty', async () => {
		const mod = await import('../../routes/api/document-cycles/stats/+server');
		proxyGetMock.mockResolvedValueOnce({ overdue: 0 });

		const res = await mod.GET(createEvent({
			url: new URL('http://localhost/api/document-cycles/stats')
		}) as never);

		expect(proxyGetMock).toHaveBeenCalledWith('/api/document-cycles/stats');
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ overdue: 0 });
	});

	it('encodes document cycle path params while preserving event query strings', async () => {
		const mod = await import('../../routes/api/document-cycles/obligations/[id]/events/+server');
		const event = createEvent({
			params: { id: 'obligation 1/2026' },
			url: new URL('http://localhost/api/document-cycles/obligations/obligation%201%2F2026/events?limit=10')
		});
		proxyGetMock.mockResolvedValueOnce([{ id: 'event-1' }]);

		const res = await mod.GET(event as never);

		expect(proxyGetMock).toHaveBeenCalledWith('/api/document-cycles/obligations/obligation%201%2F2026/events?limit=10');
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual([{ id: 'event-1' }]);
	});

	it('forwards notification list filters through the authenticated proxy', async () => {
		const mod = await import('../../routes/api/notifications/+server');
		const event = createEvent({
			url: new URL('http://localhost/api/notifications?limit=60&unread_only=true')
		});
		proxyGetMock.mockResolvedValueOnce([{ id: 'notification-1', read_at: null }]);

		const res = await mod.GET(event as never);

		expect(proxyGetMock).toHaveBeenCalledWith('/api/notifications?limit=60&unread_only=true');
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual([{ id: 'notification-1', read_at: null }]);
	});

	it('forwards notification unread count through the authenticated proxy', async () => {
		const mod = await import('../../routes/api/notifications/unread-count/+server');
		proxyGetMock.mockResolvedValueOnce({ unread: 5 });

		const res = await mod.GET(createEvent() as never);

		expect(proxyGetMock).toHaveBeenCalledWith('/api/notifications/unread-count');
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ unread: 5 });
	});

	it('forwards notification mark-read mutations through the authenticated proxy', async () => {
		const mod = await import('../../routes/api/notifications/[id]/read/+server');
		const event = createEvent({ params: { id: 'notification 1/2026' } });
		proxyPostMock.mockResolvedValueOnce({ id: 'notification 1/2026', read_at: '2026-05-01T12:00:00Z' });

		const res = await mod.POST(event as never);

		expect(proxyPostMock).toHaveBeenCalledWith('/api/notifications/notification%201%2F2026/read');
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ id: 'notification 1/2026', read_at: '2026-05-01T12:00:00Z' });
	});

	it('forwards notification read-all mutations through the authenticated proxy', async () => {
		const mod = await import('../../routes/api/notifications/read-all/+server');
		proxyPostMock.mockResolvedValueOnce({ marked: 3 });

		const res = await mod.POST(createEvent() as never);

		expect(proxyPostMock).toHaveBeenCalledWith('/api/notifications/read-all');
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ marked: 3 });
	});

	it('forwards public student registration through the SvelteKit fetch boundary', async () => {
		const mod = await import('../../routes/api/public/register-student/+server');
		const upstream = new Response(JSON.stringify({ data: { id: 'registration-1' } }), {
			status: 201,
			headers: { 'Content-Type': 'application/json' }
		});
		const eventFetch = vi.fn<typeof fetch>().mockResolvedValueOnce(upstream);
		const request = new Request('http://localhost/api/public/register-student', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ nama: 'Siswa Baru', nis: '1234567890', gender: 'P', parent_name: 'Ortu' })
		});
		const event = createEvent({ request, fetch: eventFetch });

		const res = await mod.POST(event as never);

		expect(readRequestJsonMock).toHaveBeenCalledWith(request);
		expect(eventFetch).toHaveBeenCalledWith('http://127.0.0.1:8080/api/public/register-student', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({
				nis: '1234567890',
				nama: 'Siswa Baru',
				gender: 'P',
				parent_name: 'Ortu',
				parent_phone: ''
			})
		});
		expect(jsonProxyResponseMock).toHaveBeenCalledWith(upstream, { status: 201 });
		expect(res).toBe(upstream);
	}, 10000);

	it('rejects malformed public student registration before proxying', async () => {
		const mod = await import('../../routes/api/public/register-student/+server');
		const eventFetch = vi.fn<typeof fetch>();
		const request = new Request('http://localhost/api/public/register-student', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ nama: 'Siswa Baru', nis: 123 })
		});

		const res = await mod.POST(createEvent({ request, fetch: eventFetch }) as never);

		expect(res.status).toBe(400);
		expect(eventFetch).not.toHaveBeenCalled();
	});

	it('rejects public student registration without gender before proxying', async () => {
		const mod = await import('../../routes/api/public/register-student/+server');
		const eventFetch = vi.fn<typeof fetch>();
		const request = new Request('http://localhost/api/public/register-student', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ nama: 'Siswa Baru', nis: '1234567890' })
		});

		const res = await mod.POST(createEvent({ request, fetch: eventFetch }) as never);

		expect(res.status).toBe(400);
		await expect(res.json()).resolves.toEqual({ error: 'gender tidak valid' });
		expect(eventFetch).not.toHaveBeenCalled();
	});

	it('normalizes parent create payload before proxying', async () => {
		const mod = await import('../../routes/api/parents/+server');
		proxyPostMock.mockResolvedValueOnce({ id: 'parent-1', nama: 'Orang Tua' });
		const request = new Request('http://localhost/api/parents', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ nama: ' Orang Tua ', phone: ' 0812 ', address: null })
		});

		const res = await mod.POST(createEvent({ request }) as never);

		expect(proxyPostMock).toHaveBeenCalledWith('/api/parents', {
			nama: 'Orang Tua',
			phone: '0812',
			address: ''
		});
		expect(res.status).toBe(201);
	});

	it('rejects invalid parent link student_id before proxying', async () => {
		const mod = await import('../../routes/api/parents/[id]/link/+server');
		const request = new Request('http://localhost/api/parents/parent-1/link', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ student_id: 'not-a-uuid' })
		});

		const res = await mod.POST(createEvent({ request, params: { id: 'parent-1' } }) as never);

		expect(res.status).toBe(400);
		expect(proxyPostMock).not.toHaveBeenCalled();
	});

	it('creates users through the typed JSON body helper', async () => {
		const mod = await import('../../routes/api/users/+server');
		proxyPostMock.mockResolvedValueOnce({ id: 'user-1', username: 'operator' });
		const request = new Request('http://localhost/api/users', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ username: 'operator', role: 'staf' })
		});
		const event = createEvent({ request });

		const res = await mod.POST(event as never);

		expect(readRequestJsonMock).toHaveBeenCalledWith(request);
		expect(proxyPostMock).toHaveBeenCalledWith('/api/users', {
			username: 'operator',
			role: 'staf'
		});
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ id: 'user-1', username: 'operator' });
	}, 10000);

	it('creates inventory items through the typed JSON body helper', async () => {
		const mod = await import('../../routes/api/inventory/items/+server');
		proxyPostMock.mockResolvedValueOnce({ id: 'item-1', nama: 'Proyektor' });
		const request = new Request('http://localhost/api/inventory/items', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ nama: 'Proyektor', kategori: 'Elektronik' })
		});
		const event = createEvent({ request });

		const res = await mod.POST(event as never);

		expect(readRequestJsonMock).toHaveBeenCalledWith(request);
		expect(proxyPostMock).toHaveBeenCalledWith('/api/inventory/items', {
			nama: 'Proyektor',
			kategori: 'Elektronik'
		});
		expect(res.status).toBe(201);
		await expect(res.json()).resolves.toEqual({ id: 'item-1', nama: 'Proyektor' });
	});

	it('creates governance compliance actions through the typed JSON body helper', async () => {
		const mod = await import('../../routes/api/governance/compliance-actions/+server');
		proxyPostMock.mockResolvedValueOnce({ id: 'action-1', title: 'Upload SIPKA' });
		const request = new Request('http://localhost/api/governance/compliance-actions', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ title: 'Upload SIPKA', external_system: 'SIPKA' })
		});
		const event = createEvent({ request });

		const res = await mod.POST(event as never);

		expect(readRequestJsonMock).toHaveBeenCalledWith(request);
		expect(proxyPostMock).toHaveBeenCalledWith('/api/governance/compliance-actions', {
			title: 'Upload SIPKA',
			external_system: 'SIPKA'
		});
		expect(res.status).toBe(201);
		await expect(res.json()).resolves.toEqual({ id: 'action-1', title: 'Upload SIPKA' });
	});

	it('updates governance units through the typed JSON body helper', async () => {
		const mod = await import('../../routes/api/governance/units/[id]/+server');
		proxyPutMock.mockResolvedValueOnce({ id: 'unit 1/2026', name: 'Tata Usaha' });
		const request = new Request('http://localhost/api/governance/units/unit%201%2F2026', {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ name: 'Tata Usaha', is_active: true })
		});
		const event = createEvent({
			params: { id: 'unit 1/2026' },
			request
		});

		const res = await mod.PUT(event as never);

		expect(readRequestJsonMock).toHaveBeenCalledWith(request);
		expect(proxyPutMock).toHaveBeenCalledWith('/api/governance/units/unit%201%2F2026', {
			name: 'Tata Usaha',
			is_active: true
		});
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ id: 'unit 1/2026', name: 'Tata Usaha' });
	});

	it('creates TU incoming letters through the typed JSON body helper', async () => {
		const mod = await import('../../routes/api/tu/surat/incoming/+server');
		proxyPostMock.mockResolvedValueOnce({ id: 'letter-1', nomor_surat: '001/MTsN2/V/2026' });
		const request = new Request('http://localhost/api/tu/surat/incoming', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ nomor_surat: '001/MTsN2/V/2026', perihal: 'Undangan' })
		});
		const event = createEvent({ request });

		const res = await mod.POST(event as never);

		expect(readRequestJsonMock).toHaveBeenCalledWith(request);
		expect(proxyPostMock).toHaveBeenCalledWith('/api/tu/surat/incoming', {
			nomor_surat: '001/MTsN2/V/2026',
			perihal: 'Undangan'
		});
		expect(res.status).toBe(201);
		await expect(res.json()).resolves.toEqual({ id: 'letter-1', nomor_surat: '001/MTsN2/V/2026' });
	});

	it('allows empty TU certificate cancellation bodies through the optional JSON helper', async () => {
		const mod = await import('../../routes/api/tu/surat-keterangan/[id]/cancel/+server');
		proxyPostMock.mockResolvedValueOnce({ id: 'cert-1', status: 'cancelled' });
		const request = new Request('http://localhost/api/tu/surat-keterangan/cert-1/cancel', {
			method: 'POST'
		});
		const event = createEvent({
			params: { id: 'cert-1' },
			request
		});

		const res = await mod.POST(event as never);

		expect(readOptionalRequestJsonMock).toHaveBeenCalledWith(request, {});
		expect(proxyPostMock).toHaveBeenCalledWith('/api/tu/surat-keterangan/cert-1/cancel', {});
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ id: 'cert-1', status: 'cancelled' });
	});

	it('encodes TU dynamic path params before forwarding status mutations', async () => {
		const mod = await import('../../routes/api/tu/surat/incoming/[id]/status/+server');
		proxyPatchMock.mockResolvedValueOnce({ id: 'letter 1/2026', status: 'processed' });
		const request = new Request('http://localhost/api/tu/surat/incoming/letter%201%2F2026/status', {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ status: 'processed' })
		});
		const event = createEvent({
			params: { id: 'letter 1/2026' },
			request
		});

		const res = await mod.PATCH(event as never);

		expect(readRequestJsonMock).toHaveBeenCalledWith(request);
		expect(proxyPatchMock).toHaveBeenCalledWith('/api/tu/surat/incoming/letter%201%2F2026/status', {
			status: 'processed'
		});
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ id: 'letter 1/2026', status: 'processed' });
	});

	it('creates kesiswaan counseling sessions through the typed JSON body helper', async () => {
		const mod = await import('../../routes/api/kesiswaan/counseling-sessions/+server');
		proxyPostMock.mockResolvedValueOnce({ id: 'bk-1', student_id: 'student-1' });
		const request = new Request('http://localhost/api/kesiswaan/counseling-sessions', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ student_id: 'student-1', topic: 'Pendampingan' })
		});
		const event = createEvent({ request });

		const res = await mod.POST(event as never);

		expect(readRequestJsonMock).toHaveBeenCalledWith(request);
		expect(proxyPostMock).toHaveBeenCalledWith('/api/kesiswaan/counseling-sessions', {
			student_id: 'student-1',
			topic: 'Pendampingan'
		});
		expect(res.status).toBe(201);
		await expect(res.json()).resolves.toEqual({ id: 'bk-1', student_id: 'student-1' });
	});

	it('encodes Kesiswaan dynamic path params before forwarding profile updates', async () => {
		const mod = await import('../../routes/api/kesiswaan/students/[id]/profile/+server');
		proxyPutMock.mockResolvedValueOnce({ id: 'student 1/2026', name: 'Siswa' });
		const request = new Request('http://localhost/api/kesiswaan/students/student%201%2F2026/profile', {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ name: 'Siswa' })
		});
		const event = createEvent({
			params: { id: 'student 1/2026' },
			request
		});

		const res = await mod.PUT(event as never);

		expect(readRequestJsonMock).toHaveBeenCalledWith(request);
		expect(proxyPutMock).toHaveBeenCalledWith('/api/kesiswaan/students/student%201%2F2026/profile', {
			name: 'Siswa'
		});
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ id: 'student 1/2026', name: 'Siswa' });
	});

	it('encodes academic query-derived path params before forwarding updates', async () => {
		const mod = await import('../../routes/api/academic/+server');
		proxyPutMock.mockResolvedValueOnce({ id: 'slot 1/2026', subject: 'Matematika' });
		const request = new Request(
			'http://localhost/api/academic?entity=timetable%20slots%2F2026&id=slot%201%2F2026',
			{
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ subject: 'Matematika' })
			}
		);
		const event = createEvent({
			url: new URL(request.url),
			request
		});

		const res = await mod.PUT(event as never);

		expect(readRequestJsonMock).toHaveBeenCalledWith(request);
		expect(proxyPutMock).toHaveBeenCalledWith(
			'/api/academic/timetable%20slots%2F2026/slot%201%2F2026',
			{ subject: 'Matematika' }
		);
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ id: 'slot 1/2026', subject: 'Matematika' });
	});

	it('encodes student lifecycle ids read from query params before forwarding', async () => {
		const mod = await import('../../routes/api/students/+server');
		proxyPatchMock.mockResolvedValueOnce({ id: 'student 1/2026', status: 'inactive' });
		const request = new Request('http://localhost/api/students?id=student%201%2F2026', {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ status: 'inactive' })
		});
		const event = createEvent({
			url: new URL(request.url),
			request
		});

		const res = await mod.PATCH(event as never);

		expect(readRequestJsonMock).toHaveBeenCalledWith(request);
		expect(proxyPatchMock).toHaveBeenCalledWith('/api/students/student%201%2F2026/lifecycle', {
			status: 'inactive'
		});
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ id: 'student 1/2026', status: 'inactive' });
	});

	it('forwards student profile updates from the legacy query route with encoded ids', async () => {
		const mod = await import('../../routes/api/students/+server');
		proxyPutMock.mockResolvedValueOnce({ id: 'student 1/2026', nama: 'Siswa' });
		const request = new Request('http://localhost/api/students?id=student%201%2F2026', {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ nama: 'Siswa' })
		});
		const event = createEvent({
			url: new URL(request.url),
			request
		});

		const res = await mod.PUT(event as never);

		expect(readRequestJsonMock).toHaveBeenCalledWith(request);
		expect(proxyPutMock).toHaveBeenCalledWith('/api/students/student%201%2F2026', {
			nama: 'Siswa'
		});
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ id: 'student 1/2026', nama: 'Siswa' });
	});

	it('encodes employee status path params before forwarding mutations', async () => {
		const mod = await import('../../routes/api/employees/[id]/status/+server');
		proxyPatchMock.mockResolvedValueOnce({ id: 'employee 1/2026', is_active: false });
		const request = new Request('http://localhost/api/employees/employee%201%2F2026/status', {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ is_active: false })
		});
		const event = createEvent({
			params: { id: 'employee 1/2026' },
			request
		});

		const res = await mod.PATCH(event as never);

		expect(readRequestJsonMock).toHaveBeenCalledWith(request);
		expect(proxyPatchMock).toHaveBeenCalledWith('/api/employees/employee%201%2F2026/status', {
			is_active: false
		});
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ id: 'employee 1/2026', is_active: false });
	});

	it('encodes grade component path params before forwarding publish mutations', async () => {
		const mod = await import('../../routes/api/grades/components/[id]/publish/+server');
		proxyPatchMock.mockResolvedValueOnce({ id: 'component 1/2026', is_published: true });
		const request = new Request('http://localhost/api/grades/components/component%201%2F2026/publish', {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ publish: true })
		});
		const event = createEvent({
			locals: { user: { role: 'guru', roles: ['guru'] } },
			params: { id: 'component 1/2026' },
			request
		});

		const res = await mod.PATCH(event as never);

		expect(readRequestJsonMock).toHaveBeenCalledWith(request);
		expect(proxyPatchMock).toHaveBeenCalledWith('/api/grades/components/component%201%2F2026/publish', {
			publish: true
		});
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ id: 'component 1/2026', is_published: true });
	});

	it('encodes journal session path params before forwarding attendance saves', async () => {
		const mod = await import('../../routes/api/journal/sessions/[id]/attendances/+server');
		proxyPostMock.mockResolvedValueOnce({ id: 'journal 1/2026', recorded: 2 });
		const request = new Request('http://localhost/api/journal/sessions/journal%201%2F2026/attendances', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ attendances: [{ student_id: 's1', status: 'present' }] })
		});
		const event = createEvent({
			locals: { user: { role: 'guru', roles: ['guru'] } },
			params: { id: 'journal 1/2026' },
			request
		});

		const res = await mod.POST(event as never);

		expect(readRequestJsonMock).toHaveBeenCalledWith(request);
		expect(proxyPostMock).toHaveBeenCalledWith('/api/journal/sessions/journal%201%2F2026/attendances', {
			attendances: [{ student_id: 's1', status: 'present' }]
		});
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ id: 'journal 1/2026', recorded: 2 });
	}, 10000);

	it('normalizes pusaka worker status queue payload', async () => {
		const mod = await import('../../routes/api/pusaka/worker/status/+server');
		const event = createEvent();
		proxyGetMock.mockResolvedValueOnce({
			active_workers: [{ worker_id: 'worker-1' }],
			total: 1,
			last_checked: '2026-05-01T12:00:00Z'
		});

		const res = await mod.GET(event as never);

		expect(proxyGetMock).toHaveBeenCalledWith('/api/pusaka/worker/status');
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({
			data: {
				active_workers: [{ worker_id: 'worker-1' }],
				total: 1,
				queue: { queued: 0, running: 0, success: 0, failed: 0 },
				last_checked: '2026-05-01T12:00:00Z'
			}
		});
	});

	it('normalizes PUSAKA job stats through the canonical PUSAKA route', async () => {
		const mod = await import('../../routes/api/pusaka/jobs/stats/+server');
		const event = createEvent();
		proxyGetMock.mockResolvedValueOnce({
			queued: 2,
			running: 1,
			success: 7,
			failed: 3
		});

		const res = await mod.GET(event as never);

		expect(proxyGetMock).toHaveBeenCalledWith('/api/pusaka/jobs/stats');
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({
			queued: 2,
			running: 1,
			success: 7,
			failed: 3,
			retry_due: 0,
			total: 13
		});
	});

	it('encodes nested PUSAKA schedule params before forwarding deletes', async () => {
		const mod = await import('../../routes/api/pusaka/employees/[id]/schedules/[scheduleId]/+server');
		const event = createEvent({
			params: { id: 'employee 1/2026', scheduleId: 'schedule?1' }
		});
		proxyDeleteMock.mockResolvedValueOnce({ ok: true });

		const res = await mod.DELETE(event as never);

		expect(proxyDeleteMock).toHaveBeenCalledWith(
			'/api/pusaka/employees/employee%201%2F2026/schedules/schedule%3F1'
		);
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ ok: true });
	});

	it('forwards PUSAKA attendance through authenticated proxy fetch', async () => {
		const mod = await import('../../routes/api/pusaka/attendance/+server');
		const event = createEvent({
			url: new URL('http://localhost/api/pusaka/attendance?employee_id=e1')
		});
		proxyFetchMock.mockResolvedValueOnce(new Response(JSON.stringify({ data: [{ id: 'row-1' }] }), {
			status: 200,
			headers: { 'Content-Type': 'application/json' }
		}));

		const res = await mod.GET(event as never);

		expect(proxyFetchMock).toHaveBeenCalledWith('/api/pusaka/attendance?employee_id=e1');
		expect(jsonProxyResponseMock).toHaveBeenCalledWith(expect.any(Response));
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ data: [{ id: 'row-1' }] });
	});

	it('uploads website media through shared JSON proxy helper', async () => {
		const mod = await import('../../routes/api/website/media/+server');
		const upstream = new Response(JSON.stringify({ data: { url: '/media/a.jpg' } }), {
			status: 201,
			headers: { 'Content-Type': 'application/json' }
		});
		const event = createEvent({
			request: new Request('http://localhost/api/website/media', {
				method: 'POST',
				body: new FormData()
			})
		});
		proxyFetchMock.mockResolvedValueOnce(upstream);

		const res = await mod.POST(event as never);

		expect(proxyFetchMock).toHaveBeenCalledWith('/api/website/media', {
			method: 'POST',
			body: expect.any(Object)
		});
		expect(jsonProxyResponseMock).toHaveBeenCalledWith(upstream, {
			fallbackMessage: 'Gagal upload gambar.',
			status: 201
		});
		expect(res).toBe(upstream);
	});

	it('exposes Bank Soal questions through the backend question handler semantics', async () => {
		const mod = await import('../../routes/api/bank-soal/questions/+server');
		const listEvent = createEvent({
			url: new URL('http://localhost/api/bank-soal/questions?subject_id=ipa&q=energi&workflow_status=draft&event_id=event-1')
		});
		proxyGetMock.mockResolvedValueOnce({ items: [{ id: 'q-1' }], total: 1 });

		const listRes = await mod.GET(listEvent as never);

		expect(proxyGetMock).toHaveBeenCalledWith(
			backendCbtPath('/questions?subject_id=ipa&q=energi&workflow_status=draft&event_id=event-1')
		);
		expect(listRes.status).toBe(200);
		await expect(listRes.json()).resolves.toEqual({ items: [{ id: 'q-1' }], total: 1 });

		const invalidRequest = new Request('http://localhost/api/bank-soal/questions', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ stem_html: '<p>Soal tanpa mapel</p>' })
		});

		const invalidRes = await mod.POST(createEvent({ request: invalidRequest }) as never);

		expect(invalidRes.status).toBe(400);
		await expect(invalidRes.json()).resolves.toEqual({ error: 'subject_id wajib diisi' });
		expect(proxyPostMock).not.toHaveBeenCalled();

		const validRequest = new Request('http://localhost/api/bank-soal/questions', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ event_id: 'event-1', subject_id: 'ipa', stem_html: '<p>Energi</p>' })
		});
		proxyPostMock.mockResolvedValueOnce({ id: 'q-2' });

		const validRes = await mod.POST(createEvent({ request: validRequest }) as never);

		expect(proxyPostMock).toHaveBeenCalledWith(backendCbtPath('/questions'), {
			event_id: 'event-1',
			subject_id: 'ipa',
			stem_html: '<p>Energi</p>'
		});
		expect(validRes.status).toBe(201);
		await expect(validRes.json()).resolves.toEqual({ id: 'q-2' });
	});


	it('falls back to aggregated Bank Soal summary when legacy backend treats summary as an id', async () => {
		const mod = await import('../../routes/api/bank-soal/summary/+server');
		const event = createEvent({ url: new URL('http://localhost/api/bank-soal/summary') });
		proxyGetMock.mockRejectedValueOnce(new MockApiError(400, 'invalid id'));
		proxyGetMock.mockResolvedValueOnce({
			items: [
				{ id: 'q-1', subject_id: 'ipa', subject_name: 'IPA', workflow_status: 'approved', status: 'published', cognitive_level: 'C4', package_count: 1 },
				{ id: 'q-2', subject_id: 'mtk', subject_name: 'Matematika', workflow_status: 'review', status: 'draft', cognitive_level: 'C2', package_count: 0 }
			],
			meta: { total: 2 }
		});

		const res = await mod.GET(event as never);

		expect(proxyGetMock).toHaveBeenNthCalledWith(1, backendCbtPath('/questions/summary'));
		expect(proxyGetMock).toHaveBeenNthCalledWith(2, backendCbtPath('/questions?limit=500&page=1'));
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toMatchObject({
			counts: { all: 2, total: 2, review: 1, approved: 1, published: 1, package_usage: 1 },
			by_subject: expect.arrayContaining([expect.objectContaining({ subject_name: 'IPA', total: 1 })]),
			by_cognitive_level: expect.arrayContaining([expect.objectContaining({ cognitive_level: 'C4', total: 1 })])
		});
	});

	it('streams Bank Soal asset files through the backend asset proxy handler', async () => {
		const mod = await import('../../routes/api/bank-soal/assets/[id]/file/+server');
		const upstream = new Response(new Uint8Array([1, 2, 3]), {
			status: 200,
			headers: { 'content-type': 'application/pdf' }
		});
		const streamed = new Response(new Uint8Array([1, 2, 3]), {
			status: 200,
			headers: { 'content-type': 'application/pdf' }
		});
		const event = createEvent({
			params: { id: 'asset 1/scan' },
			url: new URL('http://localhost/api/bank-soal/assets/asset%201%2Fscan/file?download=1')
		});
		proxyFetchMock.mockResolvedValueOnce(upstream);
		streamProxyResponseMock.mockResolvedValueOnce(streamed);

		const res = await mod.GET(event as never);

		expect(proxyFetchMock).toHaveBeenCalledWith(
			apiPathWithQueryMock(backendCbtApiPath`/assets/${'asset 1/scan'}/file`, 'download=1')
		);
		expect(streamProxyResponseMock).toHaveBeenCalledWith(upstream, {
			fallbackMessage: 'Gagal mengambil aset CBT.',
			defaultCacheControl: 'private, max-age=300'
		});
		expect(res).toBe(streamed);
	});

	it('exposes Asesmen event, package, and session aliases through backend assessment paths', async () => {
		const eventsMod = await import('../../routes/api/asesmen/events/+server');
		const eventSessionsMod = await import('../../routes/api/asesmen/events/[id]/sessions/+server');
		const packagesMod = await import('../../routes/api/asesmen/packages/+server');
		const sessionsMod = await import('../../routes/api/asesmen/sessions/+server');
		proxyGetMock.mockResolvedValueOnce([{ id: 'event-1' }]);
		proxyGetMock.mockResolvedValueOnce([{ id: 'session-1' }]);
		proxyGetMock.mockResolvedValueOnce({ packages: [] });
		proxyGetMock.mockResolvedValueOnce([]);

		await eventsMod.GET(createEvent({ url: new URL('http://localhost/api/asesmen/events') }) as never);
		await eventSessionsMod.GET(createEvent({ params: { id: 'event 1/2026' } }) as never);
		await packagesMod.GET(createEvent({ url: new URL('http://localhost/api/asesmen/packages?event_id=event-1') }) as never);
		await sessionsMod.GET(createEvent({ url: new URL('http://localhost/api/asesmen/sessions?event_id=event-1') }) as never);

		expect(proxyGetMock).toHaveBeenNthCalledWith(1, backendCbtPath('/events'));
		expect(proxyGetMock).toHaveBeenNthCalledWith(2, backendCbtApiPath`/events/${'event 1/2026'}/sessions`);
		expect(proxyGetMock).toHaveBeenNthCalledWith(3, backendCbtPath('/packages?event_id=event-1'));
		expect(proxyGetMock).toHaveBeenNthCalledWith(4, backendCbtPath('/sessions?event_id=event-1'));

		const packageRequest = new Request('http://localhost/api/asesmen/packages', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ event_id: 'event-1', subject_id: 'ipa', title: 'Paket IPA', duration_minutes: 90 })
		});
		proxyPostMock.mockResolvedValueOnce({ id: 'package-1' });
		await packagesMod.POST(createEvent({ request: packageRequest }) as never);

		expect(proxyPostMock).toHaveBeenCalledWith(backendCbtPath('/packages'), expect.objectContaining({ event_id: 'event-1' }));

		const sessionRequest = new Request('http://localhost/api/asesmen/sessions', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ event_id: 'event-1', package_id: 'package-1', class_id: 'class-1', scope_type: 'class', title: 'Sesi IPA', scheduled_start: '2026-05-04T01:00:00Z', scheduled_end: '2026-05-04T02:00:00Z' })
		});
		proxyPostMock.mockResolvedValueOnce({ id: 'session-1' });
		await sessionsMod.POST(createEvent({ request: sessionRequest }) as never);

		expect(proxyPostMock).toHaveBeenCalledWith(backendCbtPath('/sessions'), expect.objectContaining({ event_id: 'event-1' }));
	});

	it('does not append blank query markers for audit and website list proxies', async () => {
		const auditMod = await import('../../routes/api/users/audit-logs/+server');
		const websiteMod = await import('../../routes/api/website/content/+server');
		const event = createEvent({ url: new URL('http://localhost/api/users/audit-logs') });
		proxyGetMock.mockResolvedValueOnce({ items: [] });
		proxyGetMock.mockResolvedValueOnce([]);

		const auditRes = await auditMod.GET(event as never);
		const websiteRes = await websiteMod.GET(createEvent({
			url: new URL('http://localhost/api/website/content')
		}) as never);

		expect(proxyGetMock).toHaveBeenNthCalledWith(1, '/api/users/audit-logs');
		expect(proxyGetMock).toHaveBeenNthCalledWith(2, '/api/website/content');
		expect(auditRes.status).toBe(200);
		expect(websiteRes.status).toBe(200);
	});

	it('forwards RBAC matrix and CRUD management routes through the authenticated proxy', async () => {
		const matrixMod = await import('../../routes/api/rbac/matrix/+server');
		const rolesMod = await import('../../routes/api/rbac/roles/+server');
		const roleMod = await import('../../routes/api/rbac/roles/[code]/+server');
		const roleStatusMod = await import('../../routes/api/rbac/roles/[code]/status/+server');
		const permissionsMod = await import('../../routes/api/rbac/permissions/+server');
		const permissionMod = await import('../../routes/api/rbac/permissions/[code]/+server');
		const permissionStatusMod = await import('../../routes/api/rbac/permissions/[code]/status/+server');

		proxyGetMock.mockResolvedValueOnce({ roles: [], permissions: [], role_permissions: {} });
		await matrixMod.GET(createEvent() as never);
		expect(proxyGetMock).toHaveBeenLastCalledWith('/api/rbac/matrix');

		const roleBody = { code: 'operator', name: 'Operator' };
		const roleRequest = new Request('http://localhost/api/rbac/roles', { method: 'POST', body: JSON.stringify(roleBody) });
		await rolesMod.POST(createEvent({ request: roleRequest }) as never);
		expect(proxyPostMock).toHaveBeenLastCalledWith('/api/rbac/roles', roleBody);

		const updateRoleBody = { name: 'Operator Baru' };
		const updateRoleRequest = new Request('http://localhost/api/rbac/roles/operator', { method: 'PUT', body: JSON.stringify(updateRoleBody) });
		await roleMod.PUT(createEvent({ params: { code: 'operator' }, request: updateRoleRequest }) as never);
		expect(proxyPutMock).toHaveBeenLastCalledWith('/api/rbac/roles/operator', updateRoleBody);

		const statusBody = { is_active: false };
		const statusRequest = new Request('http://localhost/api/rbac/roles/operator/status', { method: 'PATCH', body: JSON.stringify(statusBody) });
		await roleStatusMod.PATCH(createEvent({ params: { code: 'operator' }, request: statusRequest }) as never);
		expect(proxyPatchMock).toHaveBeenLastCalledWith('/api/rbac/roles/operator/status', statusBody);

		const permissionBody = { code: 'reports.view', module: 'reports', action: 'view', description: 'Lihat laporan' };
		const permissionRequest = new Request('http://localhost/api/rbac/permissions', { method: 'POST', body: JSON.stringify(permissionBody) });
		await permissionsMod.POST(createEvent({ request: permissionRequest }) as never);
		expect(proxyPostMock).toHaveBeenLastCalledWith('/api/rbac/permissions', permissionBody);

		const updatePermissionBody = { description: 'Lihat laporan madrasah' };
		const updatePermissionRequest = new Request('http://localhost/api/rbac/permissions/reports.view', { method: 'PUT', body: JSON.stringify(updatePermissionBody) });
		await permissionMod.PUT(createEvent({ params: { code: 'reports.view' }, request: updatePermissionRequest }) as never);
		expect(proxyPutMock).toHaveBeenLastCalledWith('/api/rbac/permissions/reports.view', updatePermissionBody);

		const permissionStatusBody = { is_active: true };
		const permissionStatusRequest = new Request('http://localhost/api/rbac/permissions/reports.view/status', { method: 'PATCH', body: JSON.stringify(permissionStatusBody) });
		await permissionStatusMod.PATCH(createEvent({ params: { code: 'reports.view' }, request: permissionStatusRequest }) as never);
		expect(proxyPatchMock).toHaveBeenLastCalledWith('/api/rbac/permissions/reports.view/status', permissionStatusBody);
	});

	it('forwards user lifecycle RBAC routes required by user management UI', async () => {
		const rolesMod = await import('../../routes/api/users/[id]/roles/+server');
		const resetMod = await import('../../routes/api/users/[id]/reset-password/+server');
		const profileMod = await import('../../routes/api/users/[id]/profile-link/+server');

		const rolesBody = { roles: ['guru'] };
		const rolesRequest = new Request('http://localhost/api/users/user-1/roles', { method: 'PATCH', body: JSON.stringify(rolesBody) });
		await rolesMod.PATCH(createEvent({ params: { id: 'user-1' }, request: rolesRequest }) as never);
		expect(proxyPatchMock).toHaveBeenLastCalledWith('/api/users/user-1/roles', rolesBody);

		const resetBody = { password: 'secret12345' };
		const resetRequest = new Request('http://localhost/api/users/user-1/reset-password', { method: 'POST', body: JSON.stringify(resetBody) });
		await resetMod.POST(createEvent({ params: { id: 'user-1' }, request: resetRequest }) as never);
		expect(proxyPostMock).toHaveBeenLastCalledWith('/api/users/user-1/reset-password', resetBody);

		const profileBody = { employee_id: 'employee-1', student_id: null, parent_id: null };
		const profileRequest = new Request('http://localhost/api/users/user-1/profile-link', { method: 'PATCH', body: JSON.stringify(profileBody) });
		await profileMod.PATCH(createEvent({ params: { id: 'user-1' }, request: profileRequest }) as never);
		expect(proxyPatchMock).toHaveBeenLastCalledWith('/api/users/user-1/profile-link', profileBody);
	});

});
