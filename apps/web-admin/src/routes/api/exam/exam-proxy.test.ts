import { describe, expect, it, vi } from 'vitest';

vi.mock('$env/dynamic/private', () => ({
	env: { API_BASE_URL: 'https://core-api.test/' }
}));

function createEvent(overrides: Record<string, unknown> = {}) {
	const request = (overrides.request as Request | undefined) ?? new Request('http://web.test/api/exam/login', {
		method: 'POST',
		headers: {
			'content-type': 'application/json',
			accept: 'application/json',
			'user-agent': 'ExamClient/1.0',
			'x-device-fingerprint': 'device-1',
			'x-forwarded-for': '203.0.113.200',
			'x-real-ip': '203.0.113.201'
		},
		body: JSON.stringify({ token: 'tok' })
	});
	return {
		params: { path: 'login' },
		url: new URL(request.url),
		request,
		fetch: vi.fn(async () => new Response(JSON.stringify({ ok: true }), {
			status: 200,
			headers: { 'content-type': 'application/json' }
		})),
		getClientAddress: () => '198.51.100.10',
		...overrides
	};
}

describe('student exam proxy hardening', () => {
	it('derives forwarded IP headers from SvelteKit client address instead of trusting client-supplied values', async () => {
		const mod = await import('./[...path]/+server');
		const fetchSpy = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) => new Response(JSON.stringify({ ok: true }), {
			status: 200,
			headers: { 'content-type': 'application/json' }
		}));
		const event = createEvent({ fetch: fetchSpy });

		const res = await mod.POST(event as never);

		expect(res.status).toBe(200);
		expect(fetchSpy).toHaveBeenCalledOnce();
		const [url, init] = fetchSpy.mock.calls[0];
		expect(url).toBe('https://core-api.test/api/exam/login');
		const headers = new Headers((init as RequestInit).headers);
		expect(headers.get('x-forwarded-for')).toBe('198.51.100.10');
		expect(headers.get('x-real-ip')).toBe('198.51.100.10');
		expect(headers.get('x-device-fingerprint')).toBe('device-1');
		expect(headers.get('user-agent')).toBe('ExamClient/1.0');
	});

	it('redacts obvious answer keys from JSON exam responses', async () => {
		const mod = await import('./[...path]/+server');
		const event = createEvent({
			fetch: vi.fn(async () => new Response(JSON.stringify({
				questions: [
					{
						id: 'q-1',
						answer_key: 'A',
						answerKey: 'B',
						correct_answer: 'C',
						options: [{ label: 'A', text: 'Pilihan A', kunci_jawaban: true }]
					}
				]
			}), { headers: { 'content-type': 'application/json; charset=utf-8' } }))
		});

		const res = await mod.POST(event as never);
		const payload = await res.json();

		expect(payload).toEqual({
			questions: [
				{
					id: 'q-1',
					options: [{ label: 'A', text: 'Pilihan A' }]
				}
			]
		});
	});

	it('keeps the exam proxy body limit in bytes', async () => {
		const mod = await import('./[...path]/+server');
		const event = createEvent({
			request: new Request('http://web.test/api/exam/answers', {
				method: 'POST',
				body: 'x'.repeat((96 * 1024) + 1)
			}),
			params: { path: 'answers' }
		});

		const res = await mod.POST(event as never);

		expect(res.status).toBe(413);
		expect(event.fetch).not.toHaveBeenCalled();
		await expect(res.json()).resolves.toEqual({ error: 'Payload ujian terlalu besar' });
	});
});
