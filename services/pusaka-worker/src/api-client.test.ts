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

test('fetchRuntimeConfig ignores out-of-bounds backend concurrency', async () => {
	globalThis.fetch = (async (input: RequestInfo | URL, init?: RequestInit) => {
		fetchCalls.push({ input, init });
		return new Response(
			JSON.stringify({ data: { max_concurrent: '999', headless: 'true' } }),
			{ status: 200, headers: { 'content-type': 'application/json' } }
		);
	}) as typeof fetch;

	const { fetchRuntimeConfig } = await import('./api-client.js');
	const config = await fetchRuntimeConfig();

	assert.deepEqual(config, { headless: true });
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

test('claimJob attaches timeout signal and classifies timeout errors', async () => {
	globalThis.fetch = (async (input: RequestInfo | URL, init?: RequestInit) => {
		fetchCalls.push({ input, init });
		const error = new Error('operation timed out');
		error.name = 'TimeoutError';
		throw error;
	}) as typeof fetch;

	const { claimJob, WorkerApiTimeoutError } = await import('./api-client.js');
	await assert.rejects(() => claimJob(), WorkerApiTimeoutError);
	assert.equal(fetchCalls[0]?.init?.signal instanceof AbortSignal, true);
});

test('failJob retries bounded non-2xx responses and returns true on success', async () => {
	let attempts = 0;
	globalThis.fetch = (async (input: RequestInfo | URL, init?: RequestInit) => {
		fetchCalls.push({ input, init });
		attempts += 1;
		if (attempts < 3) {
			return new Response('temporary backend failure', { status: 503 });
		}
		return new Response(JSON.stringify({ data: { status: 'failed' } }), {
			status: 200,
			headers: { 'content-type': 'application/json' }
		});
	}) as typeof fetch;

	const { failJob } = await import('./api-client.js');
	assert.equal(await failJob('job-1', 'runner failed'), true);
	assert.equal(fetchCalls.length, 3);
	assert.equal(String(fetchCalls[0]?.input).includes('/api/pusaka/worker/jobs/job-1/fail'), true);
});

test('failJob throws after bounded report retries fail', async () => {
	globalThis.fetch = (async (input: RequestInfo | URL, init?: RequestInit) => {
		fetchCalls.push({ input, init });
		return new Response('still down', { status: 500 });
	}) as typeof fetch;

	const { failJob } = await import('./api-client.js');
	await assert.rejects(() => failJob('job-2', 'runner failed'), /fail report failed: 500/);
	assert.equal(fetchCalls.length, 3);
});

test('worker API errors cap response bodies', async () => {
	globalThis.fetch = (async (input: RequestInfo | URL, init?: RequestInit) => {
		fetchCalls.push({ input, init });
		return new Response('x'.repeat(10_000), { status: 500 });
	}) as typeof fetch;

	const { sendHeartbeat } = await import('./api-client.js');
	await assert.rejects(
		() => sendHeartbeat({ consumerCount: 1, targetConcurrency: 1, headless: true, lastSyncAt: '' }),
		(error: Error) => {
			assert.match(error.message, /heartbeat failed: 500/);
			assert.match(error.message, /\[truncated\]/);
			assert.ok(error.message.length < 4300);
			return true;
		}
	);
});
