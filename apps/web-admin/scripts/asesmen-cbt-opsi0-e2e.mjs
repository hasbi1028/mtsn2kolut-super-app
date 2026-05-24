#!/usr/bin/env node

const ENV_PREFIX = 'WEB_ADMIN_ASESMEN_CBT_OPSI0_';
const REQUIRED_ENV_HINTS = Object.freeze({
	baseUrl: 'WEB_ADMIN_ASESMEN_CBT_OPSI0_BASE_URL',
	destructive: 'WEB_ADMIN_ASESMEN_CBT_OPSI0_DESTRUCTIVE',
});
const baseURL = env('BASE_URL').replace(/\/$/, '');
const headless = env('HEADLESS', 'true') !== 'false';
const timeout = Number(env('TIMEOUT_MS', '20000'));
const allowSkip = env('ALLOW_SKIP', 'true') !== 'false';
const destructive = env('DESTRUCTIVE', 'false') === 'true';
const confirmTestDb = env('CONFIRM_TEST_DB', 'false') === 'true';
const ignoreHTTPSErrors = env('IGNORE_HTTPS_ERRORS', 'false') === 'true';
const fingerprint = env('DEVICE_FINGERPRINT', 'opsi0-smoke-device');
const badFingerprint = env('BAD_DEVICE_FINGERPRINT', 'opsi0-smoke-bad-device');
const badRoomToken = env('BAD_ROOM_TOKEN', 'SALAH-ROOM-TOKEN');

const routePersonas = {
	admin: {
		envPrefix: 'ADMIN',
		fromPath: '/asesmen',
		reachable: ['/asesmen', '/asesmen/persiapan', '/asesmen/pelaksanaan', '/asesmen/pengawasan', '/ujian/command-center'],
		forbidden: [],
	},
	proctor: {
		envPrefix: 'PROCTOR',
		fromPath: '/asesmen/pelaksanaan',
		reachable: ['/asesmen', '/asesmen/persiapan', '/asesmen/pelaksanaan', '/asesmen/pengawasan', '/ujian/command-center'],
		forbidden: [],
	},
	result: {
		envPrefix: 'RESULT',
		fromPath: '/asesmen/hasil',
		reachable: ['/asesmen/hasil'],
		forbidden: [],
	},
	guru: {
		envPrefix: 'GURU',
		fromPath: '/asesmen',
		reachable: ['/asesmen'],
		forbidden: ['/asesmen/pelaksanaan', '/asesmen/pengawasan', '/asesmen/hasil'],
	},
};

function env(name, fallback = '') {
	return process.env[`${ENV_PREFIX}${name}`] ?? fallback;
}

function skipOrFail(reason) {
	if (allowSkip) {
		console.log(`SKIP Asesmen CBT Opsi 0 E2E: ${reason}`);
		console.log(`Set ${ENV_PREFIX}ALLOW_SKIP=false in staging/CI to require this smoke.`);
		process.exit(0);
	}
	throw new Error(`Asesmen CBT Opsi 0 E2E prerequisites missing: ${reason}`);
}

function requireDestructiveGuard() {
	if (!destructive) return;
	if (!confirmTestDb) {
		throw new Error(
			`Refusing destructive mode: ${REQUIRED_ENV_HINTS.destructive}=true requires ${ENV_PREFIX}CONFIRM_TEST_DB=true before /api/exam/answer or /api/exam/submit can run.`,
		);
	}
	console.log('Asesmen CBT Opsi 0 destructive mode confirmed for resettable test DB; guarded destructive scaffold is enabled.');
}

async function loadChromium() {
	try {
		return (await import('playwright')).chromium;
	} catch {
		try {
			return (await import('@playwright/test')).chromium;
		} catch {
			skipOrFail('Playwright is not installed');
		}
	}
}

function absolutePath(path) {
	return `${baseURL}${path}`;
}

function credential(persona, suffix) {
	return env(`${routePersonas[persona].envPrefix}_${suffix}`);
}

function personaHasCredentials(persona) {
	return Boolean(credential(persona, 'USERNAME') && credential(persona, 'PASSWORD'));
}

function assertNoPartialPersonaCredentials() {
	for (const persona of Object.keys(routePersonas)) {
		const username = credential(persona, 'USERNAME');
		const password = credential(persona, 'PASSWORD');
		if (Boolean(username) !== Boolean(password)) {
			skipOrFail(`partial credentials for ${persona}; both ${ENV_PREFIX}${routePersonas[persona].envPrefix}_USERNAME and ${ENV_PREFIX}${routePersonas[persona].envPrefix}_PASSWORD are required`);
		}
	}
}

function requireStrictEnv() {
	if (allowSkip) return;
	const required = [
		'BASE_URL',
		'ADMIN_USERNAME',
		'ADMIN_PASSWORD',
		'PROCTOR_USERNAME',
		'PROCTOR_PASSWORD',
		'GURU_USERNAME',
		'GURU_PASSWORD',
		'EXAM_TOKEN',
		'ROOM_TOKEN',
	];
	const missing = required.filter((name) => !env(name));
	if (missing.length > 0) {
		throw new Error(`Asesmen CBT Opsi 0 strict mode missing env: ${missing.map((name) => `${ENV_PREFIX}${name}`).join(', ')}`);
	}
}

function collectRunnablePhases() {
	assertNoPartialPersonaCredentials();
	requireStrictEnv();
	const personas = Object.keys(routePersonas).filter(personaHasCredentials);
	const exam = Boolean(env('EXAM_TOKEN') && env('ROOM_TOKEN'));
	return { personas, exam };
}

async function loginAs(page, username, password, fromPath) {
	await page.goto(absolutePath(`/login?from=${encodeURIComponent(fromPath)}`), { waitUntil: 'domcontentloaded' });
	await page.locator('input[name="username"]').fill(username);
	await page.locator('input[name="password"]').fill(password);
	await page.getByRole('button', { name: /^Masuk$/i }).click();
	await page.waitForLoadState('networkidle').catch(() => undefined);
	if (new URL(page.url()).pathname === '/login') {
		const errorText = await page.locator('#login-error').innerText().catch(() => 'no login error rendered');
		throw new Error(`login failed for configured ${fromPath} persona: ${errorText}`);
	}
}

async function expectReachable(page, path, label, expectedTexts = []) {
	const response = await page.goto(absolutePath(path), { waitUntil: 'domcontentloaded' });
	await page.waitForLoadState('networkidle').catch(() => undefined);
	const status = response?.status() ?? 0;
	const finalPath = new URL(page.url()).pathname;
	if (status >= 400) throw new Error(`${label} returned HTTP ${status}`);
	if (finalPath === '/login') throw new Error(`${label} redirected to login`);
	if (finalPath !== path && !finalPath.startsWith(`${path.replace(/\/$/, '')}/`)) {
		throw new Error(`${label} landed on unexpected path ${finalPath}`);
	}
	if (expectedTexts.length > 0) {
		const body = await page.locator('body').innerText().catch(() => '');
		for (const expected of expectedTexts) {
			if (!new RegExp(expected, 'i').test(body)) throw new Error(`${label} did not render expected text ${expected}`);
		}
	}
}

async function expectForbidden(page, path, label) {
	const response = await page.goto(absolutePath(path), { waitUntil: 'domcontentloaded' });
	await page.waitForLoadState('networkidle').catch(() => undefined);
	const status = response?.status() ?? 0;
	const body = await page.locator('body').innerText().catch(() => '');
	const finalPath = new URL(page.url()).pathname;
	const forbiddenText = /forbidden|tidak diizinkan|tidak memiliki akses|admin role required|403/i.test(body);
	if (status === 403 || forbiddenText || finalPath === '/' || finalPath === '/login') return;
	throw new Error(`${label} was not forbidden; final URL ${page.url()} status ${status}`);
}

async function apiRequest(context, method, path, body, headers = {}) {
	const options = {
		headers: {
			accept: 'application/json',
			'user-agent': 'Asesmen-CBT-Opsi0-Smoke/1.0',
			...headers,
		},
		maxRedirects: 0,
	};
	if (body !== undefined) {
		options.headers['content-type'] = options.headers['content-type'] ?? 'application/json';
		options.data = typeof body === 'string' ? body : JSON.stringify(body);
	}
	const response = await context.request[method.toLowerCase()](absolutePath(path), options);
	const text = await response.text();
	let json;
	try {
		json = text ? JSON.parse(text) : undefined;
	} catch {
		json = undefined;
	}
	return { response, status: response.status(), headers: response.headers(), text, json };
}

const sensitiveExamKeyPattern = /answer[-_ ]?key|correct[-_ ]?answers?|kunci[-_ ]?jawaban|is_correct|correct/i;

function assertNoSensitiveExamKeys(value, path = '$', seen = new WeakSet()) {
	if (!value || typeof value !== 'object') return;
	if (seen.has(value)) return;
	seen.add(value);
	if (Array.isArray(value)) {
		value.forEach((child, index) => assertNoSensitiveExamKeys(child, `${path}[${index}]`, seen));
		return;
	}
	for (const [key, child] of Object.entries(value)) {
		if (sensitiveExamKeyPattern.test(key)) {
			throw new Error(`Sensitive exam key leaked at ${path}.${key}`);
		}
		assertNoSensitiveExamKeys(child, `${path}.${key}`, seen);
	}
}

function assertNoRawTokenInText(text, token, label) {
	if (!token || !text) return;
	if (text.includes(token)) throw new Error(`${label} leaked a raw token in response body`);
}

function assertNoStore(headers, label) {
	const cacheControl = headers['cache-control'] ?? '';
	if (!/no-store/i.test(cacheControl)) throw new Error(`${label} did not include cache-control: no-store`);
}

function expectNonSuccess(result, label) {
	if (result.status >= 200 && result.status < 300) throw new Error(`${label} unexpectedly returned HTTP ${result.status}`);
}

function expectSuccess(result, label) {
	if (result.status < 200 || result.status >= 300) throw new Error(`${label} returned HTTP ${result.status}`);
}

function examSessionHeaders(examToken, activeFingerprint = fingerprint) {
	return {
		'x-exam-token': examToken,
		'x-device-fingerprint': activeFingerprint,
	};
}

function extractExamToken(loginResult) {
	const headerToken = loginResult.headers['x-exam-token'];
	if (headerToken) return headerToken;
	const body = loginResult.json;
	if (!body || typeof body !== 'object') return env('EXAM_TOKEN');
	for (const key of ['exam_token', 'examToken', 'session_token', 'sessionToken', 'token']) {
		const value = body[key];
		if (typeof value === 'string' && value) return value;
	}
	return env('EXAM_TOKEN');
}

async function runRouteRoleSmoke(browser, personas) {
	if (personas.length === 0) {
		skipOrFail('no role credentials configured for route role smoke');
	}
	for (const persona of personas) {
		const config = routePersonas[persona];
		const context = await browser.newContext({ baseURL, ignoreHTTPSErrors });
		const page = await context.newPage();
		page.setDefaultTimeout(timeout);
		try {
			await loginAs(page, credential(persona, 'USERNAME'), credential(persona, 'PASSWORD'), config.fromPath);
			for (const path of config.reachable) await expectReachable(page, path, `${persona} ${path}`);
			for (const path of config.forbidden) await expectForbidden(page, path, `${persona} ${path}`);
			console.log(`Asesmen CBT Opsi 0 route role smoke passed for ${persona}.`);
		} finally {
			await context.close();
		}
	}
}

async function runExamProxySafeSmoke(browser) {
	const examToken = env('EXAM_TOKEN');
	const roomToken = env('ROOM_TOKEN');
	if (!examToken || !roomToken) {
		skipOrFail(`missing ${ENV_PREFIX}EXAM_TOKEN or ${ENV_PREFIX}ROOM_TOKEN for exam proxy smoke`);
	}

	const context = await browser.newContext({ baseURL, ignoreHTTPSErrors });
	try {
		const statusWithoutToken = await apiRequest(context, 'GET', '/api/exam/status');
		expectNonSuccess(statusWithoutToken, 'GET /api/exam/status without token/fingerprint');
		assertNoStore(statusWithoutToken.headers, 'GET /api/exam/status without token/fingerprint');

		const badRoom = await apiRequest(context, 'POST', '/api/exam/login', { token: examToken, room_token: badRoomToken }, { 'x-device-fingerprint': fingerprint });
		expectNonSuccess(badRoom, 'POST /api/exam/login with bad room token');
		assertNoRawTokenInText(badRoom.text, examToken, 'bad room login response');
		assertNoRawTokenInText(badRoom.text, roomToken, 'bad room login response');
		if (badRoom.json) assertNoSensitiveExamKeys(badRoom.json);

		const loginResult = await apiRequest(context, 'POST', '/api/exam/login', { token: examToken, room_token: roomToken }, { 'x-device-fingerprint': fingerprint });
		expectSuccess(loginResult, 'POST /api/exam/login');
		assertNoStore(loginResult.headers, 'POST /api/exam/login');
		if (loginResult.json) assertNoSensitiveExamKeys(loginResult.json);
		assertNoRawTokenInText(loginResult.text, examToken, 'exam login response');

		const activeExamToken = extractExamToken(loginResult);
		const statusResult = await apiRequest(context, 'GET', '/api/exam/status', undefined, examSessionHeaders(activeExamToken));
		expectSuccess(statusResult, 'GET /api/exam/status with token');
		assertNoStore(statusResult.headers, 'GET /api/exam/status with token');
		if (statusResult.json) assertNoSensitiveExamKeys(statusResult.json);

		const heartbeat = await apiRequest(context, 'POST', '/api/exam/heartbeat', { at: new Date().toISOString() }, examSessionHeaders(activeExamToken));
		expectSuccess(heartbeat, 'POST /api/exam/heartbeat');
		assertNoStore(heartbeat.headers, 'POST /api/exam/heartbeat');
		if (heartbeat.json) assertNoSensitiveExamKeys(heartbeat.json);

		const badFingerprintStatus = await apiRequest(context, 'GET', '/api/exam/status', undefined, examSessionHeaders(activeExamToken, badFingerprint));
		expectNonSuccess(badFingerprintStatus, 'GET /api/exam/status with bad fingerprint');

		const oversizedPayload = JSON.stringify({ event: 'oversized-smoke', data: 'x'.repeat((96 * 1024) + 1) });
		const oversizedEvent = await apiRequest(context, 'POST', '/api/exam/event', oversizedPayload, {
			...examSessionHeaders(activeExamToken),
			'content-type': 'application/json',
		});
		expectNonSuccess(oversizedEvent, 'POST /api/exam/event oversized payload');
		if (oversizedEvent.status !== 413) console.log(`Asesmen CBT Opsi 0 oversized /api/exam/event returned non-success HTTP ${oversizedEvent.status}.`);

		if (destructive) {
			console.log('Asesmen CBT Opsi 0 destructive answer/submit checks are scaffolded but not executed by this safe smoke script.');
		}
		console.log('Asesmen CBT Opsi 0 exam proxy safe smoke passed.');
	} finally {
		await context.close();
	}
}

requireDestructiveGuard();
if (!baseURL) skipOrFail(`missing ${REQUIRED_ENV_HINTS.baseUrl}`);

const phases = collectRunnablePhases();
if (phases.personas.length === 0 && !phases.exam) {
	skipOrFail(
		`no runnable phase configured; provide role credentials such as ${ENV_PREFIX}ADMIN_USERNAME/${ENV_PREFIX}ADMIN_PASSWORD and/or ${ENV_PREFIX}EXAM_TOKEN/${ENV_PREFIX}ROOM_TOKEN`,
	);
}

const chromium = await loadChromium();
const browser = await chromium.launch({ headless, args: ['--no-sandbox', '--disable-setuid-sandbox'] });
try {
	if (phases.personas.length > 0) await runRouteRoleSmoke(browser, phases.personas);
	if (phases.exam) await runExamProxySafeSmoke(browser);
} finally {
	await browser.close();
}

console.log('Asesmen CBT Opsi 0 E2E smoke completed successfully.');
