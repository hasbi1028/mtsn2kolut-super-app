import test, { beforeEach, afterEach } from 'node:test';
import assert from 'node:assert/strict';

const originalFetch = globalThis.fetch;
let fetchCalls: Array<{ input: RequestInfo | URL; init?: RequestInit }> = [];

beforeEach(() => {
	fetchCalls = [];
});

afterEach(() => {
	globalThis.fetch = originalFetch;
});

test('fetchRuntimeConfig normalizes backend strings into runtime config', async () => {
	globalThis.fetch = (async (input: RequestInfo | URL, init?: RequestInit) => {
		fetchCalls.push({ input, init });
		return new Response(
			JSON.stringify({ data: { max_concurrent: '7', headless: 'false' } }),
			{ status: 200, headers: { 'content-type': 'application/json' } }
		);
	}) as typeof fetch;

	const { fetchRuntimeConfig } = await import('./api-client.js');
	const config = await fetchRuntimeConfig();

	assert.deepEqual(config, { maxConcurrent: 7, headless: false });
	assert.equal(String(fetchCalls[0]?.input).includes('/api/pusaka/worker/config'), true);
	assert.equal(fetchCalls[0]?.init?.method, 'GET');
});

test('sendHeartbeat posts worker status payload', async () => {
	globalThis.fetch = (async (input: RequestInfo | URL, init?: RequestInit) => {
		fetchCalls.push({ input, init });
		return new Response(JSON.stringify({ data: { status: 'ok' } }), {
			status: 200,
			headers: { 'content-type': 'application/json' }
		});
	}) as typeof fetch;

	const { sendHeartbeat } = await import('./api-client.js');
	await sendHeartbeat({
		consumerCount: 3,
		targetConcurrency: 5,
		headless: true,
		lastSyncAt: '2026-05-01T10:00:00Z'
	});

	assert.equal(String(fetchCalls[0]?.input).includes('/api/pusaka/worker/heartbeat'), true);
	assert.equal(fetchCalls[0]?.init?.method, 'POST');
	assert.match(String(fetchCalls[0]?.init?.body), /"active_consumers":3/);
	assert.match(String(fetchCalls[0]?.init?.body), /"target_concurrency":5/);
});
