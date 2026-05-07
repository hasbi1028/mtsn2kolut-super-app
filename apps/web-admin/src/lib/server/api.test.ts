import { afterEach, describe, expect, it, vi } from 'vitest';
import type { RequestEvent } from '@sveltejs/kit';
import * as apiModule from './api';
import { ApiError, AuthValidationUnavailableError, RequestPayloadError, apiLoginWithFetch, apiPath, apiPathWithQuery, apiPublicGetWithFetch, apiPublicPostWithFetch, apiValidateAuthWithFetch, handleRouteError, jsonProxyResponse, proxy, readOptionalRequestJson, readProxyJson, readRequestJson, requireAuthorizationHeader, requiredRouteParam, streamProxyResponse } from './api';

function okResponse<T>(data: T, init?: ResponseInit) {
	return new Response(JSON.stringify({ data }), {
		status: 200,
		headers: { 'Content-Type': 'application/json' },
		...init,
	});
}

function fakeEvent(fetchMock: typeof fetch, accessToken = 'access-token-1') {
	return {
		fetch: fetchMock,
		locals: { accessToken },
		cookies: {
			get: vi.fn()
		}
	} as unknown as RequestEvent;
}

describe('server api helpers', () => {
	afterEach(() => {
		vi.restoreAllMocks();
	});

	it('requireAuthorizationHeader rejects missing token', () => {
		expect(() => requireAuthorizationHeader()).toThrowError(ApiError);
	});

	it('does not expose legacy helpers that bypass SvelteKit event.fetch', () => {
		expect('apiGet' in apiModule).toBe(false);
		expect('apiPost' in apiModule).toBe(false);
		expect('apiPublicGet' in apiModule).toBe(false);
		expect('apiLogin' in apiModule).toBe(false);
		expect('apiRefresh' in apiModule).toBe(false);
	});

	it('apiPublicGetWithFetch uses the provided SvelteKit fetch boundary', async () => {
		const eventFetch = vi.fn<typeof fetch>().mockResolvedValue(okResponse({ ok: true }));

		const result = await apiPublicGetWithFetch<{ ok: boolean }>(
			eventFetch,
			'/api/public/site/pages/profil'
		);

		expect(result).toEqual({ ok: true });
		expect(eventFetch).toHaveBeenCalledWith(
			expect.stringContaining('/api/public/site/pages/profil'),
			expect.objectContaining({
				headers: { 'Content-Type': 'application/json' }
			})
		);
	});

	it('apiPublicPostWithFetch uses the provided SvelteKit fetch boundary', async () => {
		const eventFetch = vi.fn<typeof fetch>().mockResolvedValue(okResponse({ ok: true }));

		const result = await apiPublicPostWithFetch<{ ok: boolean }>(
			eventFetch,
			'/api/auth/logout',
			{ refresh_token: 'refresh-1' }
		);

		expect(result).toEqual({ ok: true });
		expect(eventFetch).toHaveBeenCalledWith(
			expect.stringContaining('/api/auth/logout'),
			expect.objectContaining({
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ refresh_token: 'refresh-1' })
			})
		);
	});

	it('apiPath encodes interpolated route params as path segments', () => {
		const path = apiPath`/api/asesmen/sessions/${'session 1/2026'}/participants/${'siswa?1'}/seat`;

		expect(path).toBe('/api/asesmen/sessions/session%201%2F2026/participants/siswa%3F1/seat');
	});

	it('apiPathWithQuery appends query strings only when present', () => {
		expect(apiPathWithQuery('/api/document-cycles/stats', new URLSearchParams())).toBe(
			'/api/document-cycles/stats'
		);
		expect(apiPathWithQuery('/api/document-cycles/stats', new URLSearchParams({
			period_year: '2026',
			status: 'waiting_verification'
		}))).toBe('/api/document-cycles/stats?period_year=2026&status=waiting_verification');
		expect(apiPathWithQuery('/api/notifications', '?limit=60')).toBe('/api/notifications?limit=60');
	});

	it('requiredRouteParam rejects missing dynamic path params', () => {
		expect(requiredRouteParam('asset-1', 'id')).toBe('asset-1');
		expect(() => requiredRouteParam('', 'id')).toThrowError(ApiError);
		expect(() => requiredRouteParam(undefined, 'id')).toThrowError(ApiError);
	});

	it('apiLoginWithFetch forwards login metadata through the provided fetch boundary', async () => {
		const eventFetch = vi.fn<typeof fetch>().mockResolvedValue(okResponse({
			access_token: 'access-1',
			refresh_token: 'refresh-1'
		}));

		const result = await apiLoginWithFetch(eventFetch, 'admin', 'secret', {
			userAgent: 'vitest-agent',
			ipAddress: '127.0.0.1'
		});

		expect(result).toEqual({ access_token: 'access-1', refresh_token: 'refresh-1' });
		expect(eventFetch).toHaveBeenCalledWith(
			expect.stringContaining('/api/auth/login'),
			expect.objectContaining({
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'X-Client-User-Agent': 'vitest-agent',
					'X-Client-IP': '127.0.0.1'
				},
				body: JSON.stringify({ username: 'admin', password: 'secret' })
			})
		);
	});

	it('apiValidateAuthWithFetch probes the protected auth sessions endpoint', async () => {
		const eventFetch = vi.fn<typeof fetch>().mockResolvedValue(okResponse([]));

		await expect(apiValidateAuthWithFetch(eventFetch, 'access-1')).resolves.toBe(true);
		expect(eventFetch).toHaveBeenCalledWith(
			expect.stringContaining('/api/auth/sessions'),
			expect.objectContaining({
				headers: { Authorization: 'Bearer access-1' }
			})
		);
	});

	it('apiValidateAuthWithFetch treats explicit auth rejection as an invalid session', async () => {
		const eventFetch = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify({ error: 'forbidden' }), { status: 403 }));

		await expect(apiValidateAuthWithFetch(eventFetch, 'access-1')).resolves.toBe(false);
	});

	it('apiValidateAuthWithFetch preserves validation outage semantics for protected UX', async () => {
		const eventFetch = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify({ error: 'down' }), { status: 503 }));

		await expect(apiValidateAuthWithFetch(eventFetch, 'access-1')).rejects.toBeInstanceOf(AuthValidationUnavailableError);
	});

	it('apiValidateAuthWithFetch maps network failures to validation outage errors', async () => {
		const eventFetch = vi.fn<typeof fetch>().mockRejectedValue(new TypeError('connect ECONNREFUSED'));

		await expect(apiValidateAuthWithFetch(eventFetch, 'access-1')).rejects.toBeInstanceOf(AuthValidationUnavailableError);
	});

	it('readRequestJson returns typed request bodies and maps malformed JSON to RequestPayloadError', async () => {
		await expect(readRequestJson<{ title: string }>(new Request('http://localhost/api/test', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ title: 'Dokumen' })
		}))).resolves.toEqual({ title: 'Dokumen' });

		await expect(readRequestJson(new Request('http://localhost/api/test', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: '{'
		}))).rejects.toBeInstanceOf(RequestPayloadError);
	});

	it('limits JSON request bodies before parsing in BFF helpers', async () => {
		await expect(readRequestJson(new Request('http://localhost/api/test', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json', 'Content-Length': '20' },
			body: JSON.stringify({ title: 'Dokumen' })
		}), 8)).rejects.toBeInstanceOf(RequestPayloadError);

		await expect(readOptionalRequestJson(new Request('http://localhost/api/test', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ title: 'Dokumen' })
		}), {}, 8)).rejects.toBeInstanceOf(RequestPayloadError);
	});

	it('readOptionalRequestJson accepts empty bodies but rejects malformed non-empty JSON', async () => {
		await expect(readOptionalRequestJson(new Request('http://localhost/api/test', {
			method: 'POST'
		}), {})).resolves.toEqual({});

		await expect(readOptionalRequestJson<{ reason: string }>(new Request('http://localhost/api/test', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ reason: 'Dibatalkan' })
		}), { reason: '' })).resolves.toEqual({ reason: 'Dibatalkan' });

		await expect(readOptionalRequestJson(new Request('http://localhost/api/test', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: '{'
		}), {})).rejects.toBeInstanceOf(RequestPayloadError);
	});

	it('returns null for 204 upstream responses', async () => {
		const eventFetch = vi.fn<typeof fetch>().mockResolvedValue(new Response(null, { status: 204 }));

		await expect(apiPublicGetWithFetch<null>(eventFetch, '/api/public/site/pages/profil')).resolves.toBeNull();
	});

	it('preserves upstream status when the upstream error is not JSON', async () => {
		const eventFetch = vi.fn<typeof fetch>().mockResolvedValue(new Response('service unavailable', { status: 503 }));

		await expect(apiPublicGetWithFetch(eventFetch, '/api/public/site/pages/profil')).rejects.toMatchObject({
			status: 503,
			message: 'Layanan backend sedang bermasalah. Silakan coba beberapa saat lagi.',
			upstreamMessage: 'HTTP 503',
		});
	});

	it('rejects successful upstream responses that are empty or not JSON', async () => {
		const eventFetch = vi.fn<typeof fetch>().mockResolvedValue(new Response('ok', { status: 200 }));

		await expect(apiPublicGetWithFetch(eventFetch, '/api/public/site/pages/profil')).rejects.toMatchObject({
			status: 502,
			message: 'Layanan backend sedang bermasalah. Silakan coba beberapa saat lagi.',
			upstreamMessage: 'Respons backend kosong atau bukan JSON',
		});
	});

	it('rejects successful API helper envelopes without data', async () => {
		const eventFetch = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify({ ok: true }), {
			status: 200,
			headers: { 'Content-Type': 'application/json' }
		}));

		await expect(apiPublicGetWithFetch(eventFetch, '/api/public/site/pages/profil')).rejects.toMatchObject({
			status: 502,
			upstreamMessage: 'Respons backend kosong atau bukan JSON',
		});
	});

	it('masks 5xx upstream messages in route responses', async () => {
		const response = handleRouteError(new ApiError(500, 'raw postgres panic', 'raw postgres panic'), '/api/test');

		expect(response.status).toBe(503);
		await expect(response.json()).resolves.toEqual({
			error: 'Layanan backend sedang bermasalah. Silakan coba beberapa saat lagi.'
		});
	});

	it('maps malformed request JSON errors to a controlled 400 response', async () => {
		const warn = vi.spyOn(console, 'warn').mockImplementation(() => {});

		const response = handleRouteError(new RequestPayloadError(), '/api/test');

		expect(response.status).toBe(400);
		await expect(response.json()).resolves.toEqual({ error: 'Payload JSON tidak valid' });
		expect(warn).toHaveBeenCalledWith(
			'[/api/test] malformed request JSON:',
			'Payload JSON tidak valid'
		);
	});

	it('keeps legacy malformed SyntaxError mapping for routes that still call request.json directly', async () => {
		const warn = vi.spyOn(console, 'warn').mockImplementation(() => {});

		const response = handleRouteError(new SyntaxError('Unexpected token } in JSON at position 0'), '/api/test');

		expect(response.status).toBe(400);
		await expect(response.json()).resolves.toEqual({ error: 'Payload JSON tidak valid' });
		expect(warn).toHaveBeenCalledWith(
			'[/api/test] malformed request JSON:',
			'Unexpected token } in JSON at position 0'
		);
	});

	it('readProxyJson returns raw JSON payloads and maps upstream errors through ApiError', async () => {
		await expect(readProxyJson<{ data: string[] }>(
			new Response(JSON.stringify({ data: ['a'] }), {
				status: 200,
				headers: { 'Content-Type': 'application/json' }
			})
		)).resolves.toEqual({ data: ['a'] });

		await expect(readProxyJson(new Response(JSON.stringify({ error: 'input salah' }), { status: 400 })))
			.rejects.toMatchObject({ status: 400, message: 'input salah' });
	});

	it('readProxyJson rejects successful upstream responses that are empty or not JSON', async () => {
		await expect(readProxyJson(new Response('not-json', { status: 200 }), 'Gagal membaca response.'))
			.rejects.toMatchObject({
				status: 502,
				message: 'Layanan backend sedang bermasalah. Silakan coba beberapa saat lagi.',
				upstreamMessage: 'Gagal membaca response.',
			});
	});

	it('jsonProxyResponse returns mapped JSON responses with controlled success status', async () => {
		const response = await jsonProxyResponse<{ data: string[] }, string[]>(
			new Response(JSON.stringify({ data: ['asset-1'] }), {
				status: 201,
				headers: { 'Content-Type': 'application/json' }
			}),
			{
				status: 200,
				map: (payload) => payload.data
			}
		);

		expect(response.status).toBe(200);
		await expect(response.json()).resolves.toEqual(['asset-1']);
	});

	it('jsonProxyResponse preserves 204 responses without writing a JSON body', async () => {
		const response = await jsonProxyResponse<null>(new Response(null, { status: 204 }));

		expect(response.status).toBe(204);
		await expect(response.text()).resolves.toBe('');
	});

	it('streamProxyResponse forwards safe file headers and applies defaults', async () => {
		const response = await streamProxyResponse(new Response(new Uint8Array([102, 105, 108, 101]), {
			status: 200,
			headers: {
				'content-disposition': 'attachment; filename="bukti.pdf"'
			}
		}), {
			defaultCacheControl: 'private, max-age=300'
		});

		expect(response.status).toBe(200);
		expect(response.headers.get('content-type')).toBe('application/octet-stream');
		expect(response.headers.get('cache-control')).toBe('private, max-age=300');
		expect(response.headers.get('content-disposition')).toBe('attachment; filename="bukti.pdf"');
		expect(response.headers.get('x-content-type-options')).toBe('nosniff');
		await expect(response.text()).resolves.toBe('file');
	});

	it('streamProxyResponse preserves upstream X-Content-Type-Options for file streams', async () => {
		const response = await streamProxyResponse(new Response(new Uint8Array([102, 105, 108, 101]), {
			status: 200,
			headers: {
				'content-type': 'application/pdf',
				'x-content-type-options': 'nosniff'
			}
		}));

		expect(response.headers.get('content-type')).toBe('application/pdf');
		expect(response.headers.get('x-content-type-options')).toBe('nosniff');
	});

	it('streamProxyResponse maps upstream file errors through ApiError', async () => {
		await expect(streamProxyResponse(new Response(JSON.stringify({ error: 'aset tidak boleh dibaca' }), {
			status: 403,
			headers: { 'Content-Type': 'application/json' }
		}), {
			fallbackMessage: 'Gagal mengambil aset.'
		})).rejects.toMatchObject({ status: 403, message: 'aset tidak boleh dibaca' });
	});

	it('proxy uses SvelteKit event.fetch so handleFetch can refresh retryable 401 responses', async () => {
		const globalFetch = vi.spyOn(globalThis, 'fetch').mockResolvedValue(okResponse({ global: true }));
		const eventFetch = vi.fn<typeof fetch>().mockResolvedValue(okResponse({ ok: true }));

		const result = await proxy(fakeEvent(eventFetch)).get<{ ok: boolean }>('/api/auth/sessions');

		expect(result).toEqual({ ok: true });
		expect(eventFetch).toHaveBeenCalledOnce();
		expect(globalFetch).not.toHaveBeenCalled();
		expect(eventFetch).toHaveBeenCalledWith(
			expect.stringContaining('/api/auth/sessions'),
			expect.objectContaining({
				headers: {
					'Content-Type': 'application/json',
					Authorization: 'Bearer access-token-1',
				},
			})
		);
	});

	it('proxy reads the latest event access token for each request', async () => {
		const eventFetch = vi.fn<typeof fetch>().mockImplementation(async () => okResponse({ ok: true }));
		const event = fakeEvent(eventFetch, 'initial-token');
		const client = proxy(event);

		await client.get('/api/auth/sessions');
		event.locals.accessToken = 'rotated-token';
		await client.get('/api/auth/sessions');

		expect(eventFetch).toHaveBeenNthCalledWith(
			1,
			expect.stringContaining('/api/auth/sessions'),
			expect.objectContaining({
				headers: {
					'Content-Type': 'application/json',
					Authorization: 'Bearer initial-token',
				},
			})
		);
		expect(eventFetch).toHaveBeenNthCalledWith(
			2,
			expect.stringContaining('/api/auth/sessions'),
			expect.objectContaining({
				headers: {
					'Content-Type': 'application/json',
					Authorization: 'Bearer rotated-token',
				},
			})
		);
	});

	it('proxy.fetch forces the real access token and never falls back to an internal key', async () => {
		const eventFetch = vi.fn<typeof fetch>().mockResolvedValue(okResponse({ ok: true }));

		await proxy(fakeEvent(eventFetch, 'real-user-token')).fetch('/api/bank-soal/questions', {
			method: 'POST',
			headers: {
				Authorization: 'Bearer caller-override',
				'X-Internal-Key': 'not-allowed'
			}
		});

		expect(eventFetch).toHaveBeenCalledWith(
			expect.stringContaining('/api/bank-soal/questions'),
			expect.objectContaining({
				method: 'POST',
				headers: {
					Authorization: 'Bearer real-user-token'
				}
			})
		);
	});

	it('proxy.fetch also uses the latest event access token for streamed or form-data requests', async () => {
		const eventFetch = vi.fn<typeof fetch>().mockResolvedValue(okResponse({ ok: true }));
		const event = fakeEvent(eventFetch, 'initial-token');
		const client = proxy(event);

		await client.fetch('/api/bank-soal/assets');
		event.locals.accessToken = 'rotated-token';
		await client.fetch('/api/bank-soal/assets', {
			headers: { Authorization: 'Bearer caller-override' }
		});

		expect(eventFetch).toHaveBeenNthCalledWith(
			1,
			expect.stringContaining('/api/bank-soal/assets'),
			expect.objectContaining({
				headers: { Authorization: 'Bearer initial-token' }
			})
		);
		expect(eventFetch).toHaveBeenNthCalledWith(
			2,
			expect.stringContaining('/api/bank-soal/assets'),
			expect.objectContaining({
				headers: { Authorization: 'Bearer rotated-token' }
			})
		);
	});
});
