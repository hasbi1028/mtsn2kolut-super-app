#!/usr/bin/env node

const baseURL = (process.env.WEB_ADMIN_BANK_SOAL_E2E_BASE_URL ?? process.env.WEB_ADMIN_SMOKE_BASE_URL ?? '').replace(/\/$/, '');
const headless = process.env.WEB_ADMIN_BANK_SOAL_E2E_HEADLESS !== 'false';
const timeout = Number(process.env.WEB_ADMIN_BANK_SOAL_E2E_TIMEOUT_MS ?? process.env.WEB_ADMIN_SMOKE_TIMEOUT_MS ?? 20000);
const allowSkip = process.env.WEB_ADMIN_BANK_SOAL_E2E_ALLOW_SKIP !== 'false';
const useSeededDefaults = process.env.WEB_ADMIN_BANK_SOAL_E2E_USE_SEEDED_DEFAULTS === 'true';
const seededUsernamePrefix = process.env.BANK_SOAL_E2E_USERNAME_PREFIX ?? 'e2e_bank_soal_';
const seededDefaultPassword = process.env.BANK_SOAL_E2E_PASSWORD;
const seededCredentialEnvNames = {
	'admin': { USERNAME: 'BANK_SOAL_E2E_ADMIN_USERNAME', PASSWORD: 'BANK_SOAL_E2E_ADMIN_PASSWORD' },
	'creator': { USERNAME: 'BANK_SOAL_E2E_CREATOR_USERNAME', PASSWORD: 'BANK_SOAL_E2E_CREATOR_PASSWORD' },
	'reviewer': { USERNAME: 'BANK_SOAL_E2E_REVIEWER_USERNAME', PASSWORD: 'BANK_SOAL_E2E_REVIEWER_PASSWORD' },
	'importer': { USERNAME: 'BANK_SOAL_E2E_IMPORTER_USERNAME', PASSWORD: 'BANK_SOAL_E2E_IMPORTER_PASSWORD' },
	'readonly': { USERNAME: 'BANK_SOAL_E2E_READONLY_USERNAME', PASSWORD: 'BANK_SOAL_E2E_READONLY_PASSWORD' },
	'no-access': { USERNAME: 'BANK_SOAL_E2E_NO_ACCESS_USERNAME', PASSWORD: 'BANK_SOAL_E2E_NO_ACCESS_PASSWORD' },
};

const finalRoutes = [
	'/bank-soal',
	'/bank-soal/daftar',
	'/bank-soal/tambah',
	'/bank-soal/verifikasi',
	'/bank-soal/impor',
	'/bank-soal/analisis-butir',
	'/bank-soal/mapel-kd',
	'/bank-soal/pengaturan',
];

const publicBoundaryChecks = [
	['/api/bank-soal/summary', 401],
	['/api/bank-soal/questions', 401],
	['/api/bank-soal/assets', 401],
];

const personas = {
	'admin': {
		envPrefix: 'ADMIN',
		reachable: finalRoutes,
		forbidden: [],
		visibleCreateControlsOn: ['/bank-soal', '/bank-soal/daftar', '/bank-soal/mapel-kd', '/bank-soal/pengaturan'],
		hiddenCreateControlsOn: [],
	},
	'creator': {
		envPrefix: 'CREATOR',
		reachable: ['/bank-soal', '/bank-soal/daftar', '/bank-soal/tambah', '/bank-soal/analisis-butir', '/bank-soal/mapel-kd'],
		forbidden: ['/bank-soal/verifikasi', '/bank-soal/impor', '/bank-soal/pengaturan'],
		visibleCreateControlsOn: ['/bank-soal', '/bank-soal/daftar', '/bank-soal/mapel-kd'],
		hiddenCreateControlsOn: [],
	},
	'reviewer': {
		envPrefix: 'REVIEWER',
		reachable: ['/bank-soal', '/bank-soal/daftar', '/bank-soal/verifikasi', '/bank-soal/analisis-butir', '/bank-soal/mapel-kd'],
		forbidden: ['/bank-soal/tambah', '/bank-soal/impor', '/bank-soal/pengaturan'],
		visibleCreateControlsOn: [],
		hiddenCreateControlsOn: ['/bank-soal', '/bank-soal/daftar', '/bank-soal/mapel-kd'],
	},
	'importer': {
		envPrefix: 'IMPORTER',
		reachable: ['/bank-soal', '/bank-soal/daftar', '/bank-soal/impor', '/bank-soal/analisis-butir', '/bank-soal/mapel-kd'],
		forbidden: ['/bank-soal/tambah', '/bank-soal/verifikasi', '/bank-soal/pengaturan'],
		visibleCreateControlsOn: [],
		hiddenCreateControlsOn: ['/bank-soal', '/bank-soal/daftar', '/bank-soal/mapel-kd'],
	},
	'readonly': {
		envPrefix: 'READONLY',
		reachable: ['/bank-soal', '/bank-soal/daftar', '/bank-soal/analisis-butir', '/bank-soal/mapel-kd'],
		forbidden: ['/bank-soal/tambah', '/bank-soal/verifikasi', '/bank-soal/impor', '/bank-soal/pengaturan'],
		visibleCreateControlsOn: [],
		hiddenCreateControlsOn: ['/bank-soal', '/bank-soal/daftar', '/bank-soal/mapel-kd'],
	},
	'no-access': {
		envPrefix: 'NO_ACCESS',
		reachable: [],
		forbidden: finalRoutes,
		visibleCreateControlsOn: [],
		hiddenCreateControlsOn: [],
	},
};

function envName(persona, suffix) {
	return `WEB_ADMIN_BANK_SOAL_E2E_${personas[persona].envPrefix}_${suffix}`;
}

function seedEnvName(persona, suffix) {
	return seededCredentialEnvNames[persona][suffix];
}

function seededCredential(persona, suffix) {
	if (!useSeededDefaults) return undefined;
	if (suffix === 'USERNAME') {
		return process.env[seedEnvName(persona, suffix)] ?? `${seededUsernamePrefix}${persona.replaceAll('-', '_')}`;
	}
	if (suffix === 'PASSWORD') {
		return process.env[seedEnvName(persona, suffix)] ?? seededDefaultPassword;
	}
	return undefined;
}

function credential(persona, suffix) {
	return process.env[envName(persona, suffix)] ?? seededCredential(persona, suffix);
}

function collectMissingEnv() {
	const missing = [];
	if (!baseURL) missing.push('WEB_ADMIN_BANK_SOAL_E2E_BASE_URL or WEB_ADMIN_SMOKE_BASE_URL');
	for (const persona of Object.keys(personas)) {
		for (const suffix of ['USERNAME', 'PASSWORD']) {
			const name = envName(persona, suffix);
			if (!credential(persona, suffix)) missing.push(useSeededDefaults ? `${name} or ${seedEnvName(persona, suffix)}` : name);
		}
	}
	return missing;
}

function skipOrFail(reason) {
	if (allowSkip) {
		console.log(`Bank Soal role E2E skipped: ${reason}`);
		console.log('Set WEB_ADMIN_BANK_SOAL_E2E_ALLOW_SKIP=false in staging/CI to require this browser smoke.');
		process.exit(0);
	}
	throw new Error(`Bank Soal role E2E prerequisites missing: ${reason}`);
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

async function login(page, persona) {
	const username = credential(persona, 'USERNAME');
	const password = credential(persona, 'PASSWORD');
	await page.goto(absolutePath(`/login?from=${encodeURIComponent('/bank-soal')}`), { waitUntil: 'domcontentloaded' });
	await page.locator('input[name="username"]').fill(username);
	await page.locator('input[name="password"]').fill(password);
	await page.getByRole('button', { name: /^Masuk$/i }).click();
	await page.waitForLoadState('networkidle').catch(() => undefined);
	if (new URL(page.url()).pathname === '/login') {
		const errorText = await page.locator('#login-error').innerText().catch(() => 'no login error rendered');
		throw new Error(`login failed for ${persona}/${username}: ${errorText}`);
	}
}

async function assertReachable(page, path, label) {
	const response = await page.goto(absolutePath(path), { waitUntil: 'domcontentloaded' });
	await page.waitForLoadState('networkidle').catch(() => undefined);
	const status = response?.status() ?? 0;
	const finalPath = new URL(page.url()).pathname;
	if (status >= 400) throw new Error(`${label} returned HTTP ${status}`);
	if (finalPath === '/login') throw new Error(`${label} redirected to login`);
	if (finalPath !== path && !finalPath.startsWith(path.replace(/\/$/, '') + '/')) {
		throw new Error(`${label} landed on unexpected path ${finalPath}`);
	}
}

async function assertForbidden(page, path, label) {
	const response = await page.goto(absolutePath(path), { waitUntil: 'domcontentloaded' });
	await page.waitForLoadState('networkidle').catch(() => undefined);
	const status = response?.status() ?? 0;
	const body = await page.locator('body').innerText().catch(() => '');
	const finalPath = new URL(page.url()).pathname;
	const forbiddenText = /forbidden|tidak diizinkan|tidak memiliki akses|admin role required|403/i.test(body);
	if (status === 403 || forbiddenText || finalPath === '/' || finalPath === '/login') return;
	throw new Error(`${label} was not forbidden; final URL ${page.url()} status ${status}`);
}

async function assertVisibleCreateControls(page, path, label) {
	await assertReachable(page, path, label);
	const link = page.locator('a[href="/bank-soal/tambah"], a[href$="/bank-soal/tambah"]').first();
	if (!(await link.isVisible().catch(() => false))) {
		throw new Error(`${label} did not show a visible create shortcut/link`);
	}
}

async function assertHiddenCreateControls(page, path, label) {
	await assertReachable(page, path, label);
	const visibleCreateLinks = await page.locator('a[href="/bank-soal/tambah"], a[href$="/bank-soal/tambah"]').evaluateAll((nodes) =>
		nodes.filter((node) => {
			const el = node;
			const style = window.getComputedStyle(el);
			const rect = el.getBoundingClientRect();
			return style.visibility !== 'hidden' && style.display !== 'none' && rect.width > 0 && rect.height > 0;
		}).length,
	);
	if (visibleCreateLinks > 0) {
		throw new Error(`${label} exposed ${visibleCreateLinks} visible create shortcut(s) without bank_soal.create`);
	}
}

async function assertApiBoundaries(context) {
	for (const [path, expected] of publicBoundaryChecks) {
		const response = await context.request.get(absolutePath(path), { maxRedirects: 0 });
		if (response.status() !== expected) {
			throw new Error(`${path} returned ${response.status()}, expected ${expected}`);
		}
	}
	const cbtResponse = await context.request.get(absolutePath('/api/cbt/questions'), { maxRedirects: 0 });
	if (cbtResponse.status() === 200) {
		throw new Error('/api/cbt/questions unexpectedly returned public 200');
	}
}

async function runPersona(browser, persona, config) {
	const context = await browser.newContext({ baseURL, ignoreHTTPSErrors: process.env.WEB_ADMIN_BANK_SOAL_E2E_IGNORE_HTTPS_ERRORS === 'true' });
	const page = await context.newPage();
	page.setDefaultTimeout(timeout);
	try {
		await login(page, persona);
		for (const path of config.reachable) await assertReachable(page, path, `${persona} ${path}`);
		for (const path of config.forbidden) await assertForbidden(page, path, `${persona} ${path}`);
		for (const path of config.visibleCreateControlsOn) await assertVisibleCreateControls(page, path, `${persona} ${path}`);
		for (const path of config.hiddenCreateControlsOn) await assertHiddenCreateControls(page, path, `${persona} ${path}`);
		console.log(`Bank Soal role E2E passed for ${persona}`);
	} finally {
		await context.close();
	}
}

const missingEnv = collectMissingEnv();
if (missingEnv.length > 0) skipOrFail(`missing ${missingEnv.join(', ')}`);

const chromium = await loadChromium();
const browser = await chromium.launch({ headless, args: ['--no-sandbox', '--disable-setuid-sandbox'] });
try {
	const unauthContext = await browser.newContext({ baseURL, ignoreHTTPSErrors: process.env.WEB_ADMIN_BANK_SOAL_E2E_IGNORE_HTTPS_ERRORS === 'true' });
	try {
		await assertApiBoundaries(unauthContext);
	} finally {
		await unauthContext.close();
	}

	for (const [persona, config] of Object.entries(personas)) {
		await runPersona(browser, persona, config);
	}
} finally {
	await browser.close();
}

console.log('Bank Soal role E2E completed successfully.');
