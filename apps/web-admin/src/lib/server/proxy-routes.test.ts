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
			{ key: 'admin_password', value: 'secret' },
			{ key: 'browser_path', value: '/usr/bin/chromium' }
		]);

		const res = await mod.GET(event as never);

		expect(proxyGetMock).toHaveBeenCalledWith('/api/pusaka/settings');
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({
			max_concurrent: 9,
			headless: true,
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
				body: JSON.stringify({ max_concurrent: 7, headless: false })
			})
		});

		const res = await mod.PUT(event as never);

		expect(proxyPutMock).toHaveBeenNthCalledWith(1, '/api/pusaka/settings/max_concurrent', {
			value: '7'
		});
		expect(proxyPutMock).toHaveBeenNthCalledWith(2, '/api/pusaka/settings/headless', {
			value: 'false'
		});
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ ok: true, max_concurrent: 7, headless: false });
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
			body: JSON.stringify({ nama: 'Siswa Baru', nisn: '1234567890' })
		});
		const event = createEvent({ request, fetch: eventFetch });

		const res = await mod.POST(event as never);

		expect(readRequestJsonMock).toHaveBeenCalledWith(request);
		expect(eventFetch).toHaveBeenCalledWith('http://localhost:8080/api/public/register-student', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ nama: 'Siswa Baru', nisn: '1234567890' })
		});
		expect(jsonProxyResponseMock).toHaveBeenCalledWith(upstream, { status: 201 });
		expect(res).toBe(upstream);
	}, 10000);

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

	it('streams CBT asset files through the shared stream proxy helper', async () => {
		const mod = await import('../../routes/api/cbt/assets/[id]/file/+server');
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
			url: new URL('http://localhost/api/cbt/assets/asset%201%2Fscan/file?download=1')
		});
		proxyFetchMock.mockResolvedValueOnce(upstream);
		streamProxyResponseMock.mockResolvedValueOnce(streamed);

		const res = await mod.GET(event as never);

		expect(proxyFetchMock).toHaveBeenCalledWith('/api/cbt/assets/asset%201%2Fscan/file?download=1');
		expect(streamProxyResponseMock).toHaveBeenCalledWith(upstream, {
			fallbackMessage: 'Gagal mengambil aset CBT.',
			defaultCacheControl: 'private, max-age=300'
		});
		expect(res).toBe(streamed);
	});

	it('encodes CBT participant path params before forwarding mutations', async () => {
		const mod = await import('../../routes/api/cbt/sessions/[id]/participants/[pid]/seat/+server');
		const request = new Request('http://localhost/api/cbt/sessions/session%201/participants/student%3F1/seat', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ seat_no: 'A-01' })
		});
		const event = createEvent({
			params: { id: 'session 1/2026', pid: 'student?1' },
			request
		});
		proxyPostMock.mockResolvedValueOnce({ ok: true });

		const res = await mod.POST(event as never);

		expect(readRequestJsonMock).toHaveBeenCalledWith(request);
		expect(proxyPostMock).toHaveBeenCalledWith(
			'/api/cbt/sessions/session%201%2F2026/participants/student%3F1/seat',
			{ seat_no: 'A-01' }
		);
		expect(res.status).toBe(200);
		await expect(res.json()).resolves.toEqual({ ok: true });
	});

	it('encodes CBT delete ids read from query params before forwarding', async () => {
		const mod = await import('../../routes/api/cbt/packages/+server');
		const event = createEvent({
			url: new URL('http://localhost/api/cbt/packages?id=package%201%2F2026')
		});

		const res = await mod.DELETE(event as never);

		expect(proxyDeleteMock).toHaveBeenCalledWith('/api/cbt/packages/package%201%2F2026');
		expect(res.status).toBe(204);
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
});
