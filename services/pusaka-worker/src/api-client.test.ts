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
	assert.equal((fetchCalls[0]?.init?.headers as Record<string, string>)?.['x-worker-id']?.length > 0, true);
	assert.match(String(fetchCalls[0]?.init?.body), /"active_consumers":3/);
	assert.match(String(fetchCalls[0]?.init?.body), /"target_concurrency":5/);
});

test('worker API client uses only canonical pusaka worker routes', async () => {
	globalThis.fetch = (async (input: RequestInfo | URL, init?: RequestInit) => {
		fetchCalls.push({ input, init });
		const path = new URL(String(input)).pathname;
		if (path.endsWith('/claim')) {
			return new Response(JSON.stringify({ data: null }), {
				status: 200,
				headers: { 'content-type': 'application/json' }
			});
		}
		if (path.endsWith('/config')) {
			return new Response(JSON.stringify({ data: {} }), {
				status: 200,
				headers: { 'content-type': 'application/json' }
			});
		}
		return new Response(JSON.stringify({ data: { status: 'ok' } }), {
			status: 200,
			headers: { 'content-type': 'application/json' }
		});
	}) as typeof fetch;

	const { claimJob, fetchRuntimeConfig, sendHeartbeat, completeJob, failJob } = await import('./api-client.js');
	await claimJob();
	await fetchRuntimeConfig();
	await sendHeartbeat({ consumerCount: 1, targetConcurrency: 5, headless: true, lastSyncAt: '2026-05-01T10:00:00Z' });
	await completeJob('job-canonical', { tanggal: '2026-05-01', jam_masuk: '07:00', jam_pulang: '' });
	assert.equal(await failJob('job-canonical', 'runner failed'), true);

	const paths = fetchCalls.map((call) => new URL(String(call.input)).pathname);
	assert.deepEqual(paths, [
		'/api/pusaka/worker/claim',
		'/api/pusaka/worker/config',
		'/api/pusaka/worker/heartbeat',
		'/api/pusaka/worker/jobs/job-canonical/complete',
		'/api/pusaka/worker/jobs/job-canonical/fail'
	]);
	for (const path of paths) {
		assert.match(path, /^\/api\/pusaka\/worker(?:\/|$)/);
		assert.doesNotMatch(path, /^\/api\/(jobs|attendance|schedules|settings|worker)(?:\/|$)/);
	}
});

test('completeJob sends worker ownership in header and body', async () => {
	globalThis.fetch = (async (input: RequestInfo | URL, init?: RequestInit) => {
		fetchCalls.push({ input, init });
		return new Response(JSON.stringify({ data: { status: 'completed' } }), {
			status: 200,
			headers: { 'content-type': 'application/json' }
		});
	}) as typeof fetch;

	const { completeJob } = await import('./api-client.js');
	await completeJob('job-complete', { tanggal: '2026-05-01', jam_masuk: '07:00', jam_pulang: '' });

	const headers = fetchCalls[0]?.init?.headers as Record<string, string>;
	const body = JSON.parse(String(fetchCalls[0]?.init?.body)) as Record<string, unknown>;
	assert.equal(String(fetchCalls[0]?.input).includes('/api/pusaka/worker/jobs/job-complete/complete'), true);
	assert.equal(fetchCalls[0]?.init?.method, 'POST');
	assert.equal(typeof headers['x-worker-id'], 'string');
	assert.equal(body.worker_id, headers['x-worker-id']);
	assert.equal(body.tanggal, '2026-05-01');
});

test('completeJob classifies timeout or 5xx as uncertain completion', async () => {
	globalThis.fetch = (async (input: RequestInfo | URL, init?: RequestInit) => {
		fetchCalls.push({ input, init });
		return new Response('backend may have processed completion', { status: 500 });
	}) as typeof fetch;

	const { completeJob, WorkerApiUncertainCompletionError } = await import('./api-client.js');
	await assert.rejects(
		() => completeJob('job-complete-uncertain', { tanggal: '2026-05-01', jam_masuk: '07:00', jam_pulang: '' }),
		WorkerApiUncertainCompletionError
	);
});

test('completeJob classifies 409/404 final-state responses separately', async () => {
	globalThis.fetch = (async (input: RequestInfo | URL, init?: RequestInit) => {
		fetchCalls.push({ input, init });
		return new Response('already final', { status: 409 });
	}) as typeof fetch;

	const { completeJob, WorkerApiFinalStateError } = await import('./api-client.js');
	await assert.rejects(
		() => completeJob('job-complete-final', { tanggal: '2026-05-01', jam_masuk: '07:00', jam_pulang: '' }),
		(error: Error) => error instanceof WorkerApiFinalStateError && (error as InstanceType<typeof WorkerApiFinalStateError>).status === 409
	);
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
	assert.equal((fetchCalls[0]?.init?.headers as Record<string, string>)?.['x-worker-id']?.length > 0, true);
	assert.match(String(fetchCalls[0]?.init?.body), /"worker_id":"/);
	assert.match(String(fetchCalls[0]?.init?.body), /"retry_after_secs":0/);
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
