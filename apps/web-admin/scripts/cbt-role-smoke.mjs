#!/usr/bin/env node

const requiredEnv = [
	'WEB_ADMIN_SMOKE_BASE_URL',
	'WEB_ADMIN_SMOKE_ADMIN_USERNAME',
	'WEB_ADMIN_SMOKE_ADMIN_PASSWORD',
	'WEB_ADMIN_SMOKE_GURU_USERNAME',
	'WEB_ADMIN_SMOKE_GURU_PASSWORD',
];

const missingEnv = requiredEnv.filter((name) => !process.env[name]);
if (missingEnv.length > 0) {
	console.log(`CBT role smoke skipped: missing ${missingEnv.join(', ')}`);
	process.exit(0);
}

const baseURL = process.env.WEB_ADMIN_SMOKE_BASE_URL.replace(/\/$/, '');
const headless = process.env.WEB_ADMIN_SMOKE_HEADLESS !== 'false';
const timeout = Number(process.env.WEB_ADMIN_SMOKE_TIMEOUT_MS ?? 15000);

async function loadChromium() {
	try {
		return (await import('playwright')).chromium;
	} catch (playwrightError) {
		try {
			return (await import('@playwright/test')).chromium;
		} catch {
			console.log('CBT role smoke skipped: Playwright is not installed in apps/web-admin.');
			console.log('Install Playwright only in staging/local operator environments when browser smoke tests are needed.');
			process.exit(0);
		}
	}
}

function absolutePath(path) {
	return `${baseURL}${path}`;
}

async function assertReachable(page, path, label) {
	const response = await page.goto(absolutePath(path), { waitUntil: 'networkidle' });
	const status = response?.status() ?? 0;
	if (status >= 400) throw new Error(`${label} returned HTTP ${status}`);
	if (new URL(page.url()).pathname === '/login') throw new Error(`${label} redirected to login`);
}

async function assertForbidden(page, path, label) {
	const response = await page.goto(absolutePath(path), { waitUntil: 'networkidle' });
	const status = response?.status() ?? 0;
	const body = await page.locator('body').innerText().catch(() => '');
	const forbiddenText = /forbidden|tidak diizinkan|admin role required|403/i.test(body);
	if (status !== 403 && !forbiddenText) {
		throw new Error(`${label} was not forbidden; final URL ${page.url()} status ${status}`);
	}
}

async function assertQuestionsRedirect(page) {
	await page.goto(absolutePath('/cbt/questions?question_id=abc-123&mode=review&unsafe=ignored'), { waitUntil: 'networkidle' });
	const redirected = new URL(page.url());
	if (redirected.pathname !== '/cbt/soal') throw new Error(`/cbt/questions did not redirect to /cbt/soal: ${page.url()}`);
	if (redirected.searchParams.get('question_id') !== 'abc-123') throw new Error('redirect did not preserve question_id');
	if (redirected.searchParams.get('mode') !== 'review') throw new Error('redirect did not preserve retained mode');
	if (redirected.searchParams.has('unsafe')) throw new Error('redirect preserved unexpected query parameter');

	await page.goto(absolutePath('/cbt/questions?question_id=abc-123&mode=studio'), { waitUntil: 'networkidle' });
	const retiredMode = new URL(page.url());
	if (retiredMode.pathname !== '/cbt/soal') throw new Error('retired mode redirect did not land on /cbt/soal');
	if (retiredMode.searchParams.has('mode')) throw new Error('retired experiment mode was preserved');
}

async function login(page, username, password, fromPath) {
	await page.goto(absolutePath(`/login?from=${encodeURIComponent(fromPath)}`), { waitUntil: 'networkidle' });
	await page.locator('input[name="username"]').fill(username);
	await page.locator('input[name="password"]').fill(password);
	await page.getByRole('button', { name: /^Masuk$/i }).click();
	await page.waitForLoadState('networkidle').catch(() => undefined);
	if (new URL(page.url()).pathname === '/login') {
		const errorText = await page.locator('#login-error').innerText().catch(() => 'no login error rendered');
		throw new Error(`login failed for ${username}: ${errorText}`);
	}
}

async function runForRole(browser, role, credentials, checks) {
	const context = await browser.newContext({ baseURL, ignoreHTTPSErrors: process.env.WEB_ADMIN_SMOKE_IGNORE_HTTPS_ERRORS === 'true' });
	const page = await context.newPage();
	page.setDefaultTimeout(timeout);
	try {
		await login(page, credentials.username, credentials.password, '/cbt/soal');
		for (const check of checks) await check(page);
	} finally {
		await context.close();
	}
	console.log(`CBT role smoke passed for ${role}`);
}

const chromium = await loadChromium();
const browser = await chromium.launch({ headless });

try {
	await runForRole(browser, 'admin', {
		username: process.env.WEB_ADMIN_SMOKE_ADMIN_USERNAME,
		password: process.env.WEB_ADMIN_SMOKE_ADMIN_PASSWORD,
	}, [
		(page) => assertReachable(page, '/cbt/soal', 'admin /cbt/soal'),
		(page) => assertReachable(page, '/cbt/events', 'admin /cbt/events'),
		(page) => assertQuestionsRedirect(page),
	]);

	await runForRole(browser, 'guru', {
		username: process.env.WEB_ADMIN_SMOKE_GURU_USERNAME,
		password: process.env.WEB_ADMIN_SMOKE_GURU_PASSWORD,
	}, [
		(page) => assertReachable(page, '/cbt/soal', 'guru /cbt/soal'),
		(page) => assertForbidden(page, '/cbt/events', 'guru /cbt/events'),
	]);
} finally {
	await browser.close();
}

console.log('CBT role smoke completed successfully.');
