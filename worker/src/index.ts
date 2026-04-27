import 'dotenv/config';
import fs from 'fs';
import path from 'path';
import os from 'os';
import { chromium, type Page } from 'playwright';

// ── Config ────────────────────────────────────────────────────────────────────

const BACKEND_URL    = (process.env.BACKEND_URL ?? 'http://localhost:8080').replace(/\/$/, '');
const WORKER_API_KEY = process.env.WORKER_API_KEY ?? '';
const WORKER_ID      = process.env.WORKER_ID    ?? `worker-${os.hostname()}-${process.pid}`;
const MAX_CONCURRENT = Math.max(1, Number(process.env.WORKER_CONCURRENCY ?? 5));
const HEADLESS       = ['1', 'true'].includes(process.env.HEADLESS ?? 'true');
const POLL_MS        = Number(process.env.POLL_MS        ?? 8000);
const SCRAPE_RETRIES = Math.max(1, Number(process.env.SCRAPE_RETRIES ?? 3));
const SCRAPE_RETRY_MS= Math.max(3000, Number(process.env.SCRAPE_RETRY_MS ?? 5000));
const ACTION_TIMEOUT = Math.max(5000, Number(process.env.ACTION_TIMEOUT  ?? 20000));
const LOG_PATH       = process.env.WORKER_LOG_PATH ?? path.resolve('../logs/worker.log');
const SCREENSHOT_DIR = process.env.SCREENSHOT_DIR  ?? path.resolve('../logs/screenshots');

const BASE_URL = 'https://pusaka-v3.kemenag.go.id';
const BASE_LAT = -3.2163111;
const BASE_LNG = 121.0428659;

const BROWSER_ARGS = [
	'--no-sandbox', '--disable-dev-shm-usage', '--disable-gpu',
	'--disable-extensions', '--disable-background-networking',
];
const USER_AGENT = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36';

// ── Types ─────────────────────────────────────────────────────────────────────

type RunType = 'morning' | 'afternoon' | 'checkin' | 'checkout';

interface ClaimedJob {
	id:              string;
	employee_id:     string;
	run_type:        RunType;
	attempts:        number;
	max_attempts:    number;
	pusaka_username: string;
	pusaka_password: string;
}

interface AttendanceRecord {
	tanggal:    string;
	jam_masuk:  string;
	jam_pulang: string;
}

interface GeoCoords { latitude: number; longitude: number; }

// ── HTTP client ───────────────────────────────────────────────────────────────

function workerHeaders(): Record<string, string> {
	return { 'content-type': 'application/json', 'x-worker-key': WORKER_API_KEY };
}

async function claimJob(): Promise<ClaimedJob | null> {
	const res = await fetch(`${BACKEND_URL}/api/worker/claim`, {
		method: 'POST',
		headers: workerHeaders(),
		body: JSON.stringify({ worker_id: WORKER_ID }),
	});
	if (res.status === 204) return null;
	if (!res.ok) throw new Error(`claim failed: ${res.status} ${await res.text()}`);
	const data = await res.json() as { data: ClaimedJob };
	return data.data ?? null;
}

async function completeJob(jobId: string, record: AttendanceRecord): Promise<void> {
	const res = await fetch(`${BACKEND_URL}/api/worker/jobs/${jobId}/complete`, {
		method: 'POST',
		headers: workerHeaders(),
		body: JSON.stringify(record),
	});
	if (!res.ok) throw new Error(`complete failed: ${res.status} ${await res.text()}`);
}

async function failJob(jobId: string, error: string): Promise<void> {
	const res = await fetch(`${BACKEND_URL}/api/worker/jobs/${jobId}/fail`, {
		method: 'POST',
		headers: workerHeaders(),
		body: JSON.stringify({ error }),
	});
	if (!res.ok) log('WARN', 'fail report error', { jobId, status: res.status });
}

// ── Bootstrap ─────────────────────────────────────────────────────────────────

fs.mkdirSync(path.dirname(LOG_PATH), { recursive: true });
log('INFO', 'worker starting', { WORKER_ID, BACKEND_URL, MAX_CONCURRENT, HEADLESS, POLL_MS });

// Spawn consumers
for (let i = 1; i <= MAX_CONCURRENT; i++) {
	const consumerId = `${WORKER_ID}-c${i}`;
	consumerLoop(consumerId).catch((e) => {
		log('ERROR', 'consumer crashed', { consumerId, error: (e as Error)?.message ?? String(e) });
	});
}

// ── Consumer loop ─────────────────────────────────────────────────────────────

async function consumerLoop(consumerId: string): Promise<void> {
	log('INFO', 'consumer started', { consumerId });
	while (true) {
		try {
			const job = await claimJob();
			if (!job) { await sleep(POLL_MS); continue; }

			log('INFO', 'job claimed', { consumerId, job_id: job.id, run_type: job.run_type, employee_id: job.employee_id });
			await processJob(job, consumerId);
		} catch (e) {
			log('ERROR', 'consumer loop error', { consumerId, error: (e as Error)?.message ?? String(e) });
			await sleep(Math.max(1000, Math.floor(POLL_MS / 2)));
		}
	}
}

// ── Job processing ────────────────────────────────────────────────────────────

async function processJob(job: ClaimedJob, consumerId: string): Promise<void> {
	try {
		let record: AttendanceRecord;

		if (job.run_type === 'checkin') {
			await withRetry((a) => checkin(job.pusaka_username, job.pusaka_password, a), 'checkin');
			record = { tanggal: toISODateMakassar(), jam_masuk: toTimeWITA(), jam_pulang: '' };
		} else if (job.run_type === 'checkout') {
			await withRetry((a) => checkout(job.pusaka_username, job.pusaka_password, a), 'checkout');
			record = { tanggal: toISODateMakassar(), jam_masuk: '', jam_pulang: toTimeWITA() };
		} else {
			record = await withRetry<AttendanceRecord>(
				(a) => scrapeOnce(job.pusaka_username, job.pusaka_password, a), 'scrape'
			);
		}

		await completeJob(job.id, record);
		log('INFO', 'job success', { consumerId, job_id: job.id, record });
	} catch (e) {
		const errMsg = String((e as Error)?.message ?? e);
		await failJob(job.id, errMsg);
		log('WARN', 'job failed — reported to frontend', { consumerId, job_id: job.id, error: errMsg });
	}
}

// ── Retry wrapper ─────────────────────────────────────────────────────────────

async function withRetry<T>(fn: (attempt: number) => Promise<T>, label: string): Promise<T> {
	let lastError: unknown;
	for (let attempt = 1; attempt <= SCRAPE_RETRIES; attempt++) {
		try { return await fn(attempt); }
		catch (e) {
			lastError = e;
			log('WARN', `${label} attempt ${attempt}/${SCRAPE_RETRIES} failed`, { error: (e as Error)?.message ?? String(e) });
			if (attempt < SCRAPE_RETRIES) await sleep(SCRAPE_RETRY_MS * attempt);
		}
	}
	throw lastError;
}

// ── Shared browser helpers ────────────────────────────────────────────────────

function getBrowserOpts() {
	return { headless: HEADLESS, args: BROWSER_ARGS };
}

function getContextOpts(extra: Record<string, unknown> = {}) {
	return { timezoneId: 'Asia/Makassar', locale: 'id-ID', viewport: { width: 1280, height: 800 }, userAgent: USER_AGENT, ...extra };
}

async function blockAssets(page: Page): Promise<void> {
	await page.route('**/*', (route) => {
		if (['image', 'font', 'media', 'stylesheet'].includes(route.request().resourceType()))
			return route.abort();
		route.continue();
	});
}

async function loginToPusaka(page: Page, username: string, password: string, label: string): Promise<void> {
	await page.goto(BASE_URL, { waitUntil: 'domcontentloaded', timeout: 45000 });
	await page.getByRole('link', { name: /login/i }).first().waitFor({ state: 'visible' });
	await page.getByRole('link', { name: /login/i }).first().click();
	await page.getByPlaceholder('Username').fill(username);
	await page.getByPlaceholder('Password').fill(password);

	const urlBeforeSubmit = page.url();
	await page.getByRole('button', { name: 'Masuk', exact: true }).click();

	// Option A: wait for URL to change away from the login page
	try {
		await page.waitForURL(
			(url) => url.href !== urlBeforeSubmit && !/login/i.test(url.pathname),
			{ timeout: 25000 },
		);

		// Landed on a new page — double-check body for in-page error (some SPAs redirect then show error)
		const bodyAfter = await page.locator('body').innerText().catch(() => '');
		if (/username atau password salah|invalid credentials|login gagal/i.test(bodyAfter)) {
			throw new Error('Login gagal: username/password ditolak oleh server');
		}

		log('INFO', `${label}: login ok`, { username, url: page.url() });
		return;
	} catch (urlErr) {
		if ((urlErr as Error)?.message?.includes('username/password')) throw urlErr;
	}

	// Option C: URL did not change — inspect body to determine real state
	const body = await page.locator('body').innerText().catch(() => '');

	if (/username atau password salah|invalid credentials|login gagal/i.test(body)) {
		throw new Error('Login gagal: username/password ditolak oleh server');
	}

	// Heuristic: page already shows post-login content despite URL not matching expectation
	if (/absensi|dashboard|beranda|profil|selamat\s+datang/i.test(body)) {
		log('INFO', `${label}: login ok (via body fallback)`, { username, url: page.url() });
		return;
	}

	throw new Error(`Login timeout: halaman tidak merespons setelah submit (URL: ${page.url()})`);
}

async function triggerGeo(page: Page): Promise<void> {
	await page.evaluate(() => new Promise<void>((resolve) => {
		navigator.geolocation.getCurrentPosition(() => resolve(), () => resolve(), { enableHighAccuracy: true, timeout: 5000 });
	})).catch(() => {});
}

async function assertPresenceResult(page: Page, label: string, username: string): Promise<void> {
	const text = await page.locator('body').innerText().catch(() => '');
	if (text.includes('PRESENSI GAGAL'))
		throw new Error('PRESENSI GAGAL — Bad Request (mungkin GPS tidak valid)');
	if (/berhasil/i.test(text)) log('INFO', `${label}: BERHASIL`, { username });
	else if (/sudah presensi/i.test(text)) log('WARN', `${label}: sudah absen sebelumnya`, { username });
	else log('WARN', `${label}: status tidak jelas`, { username, snippet: text.substring(0, 300) });
}

async function saveFailScreenshot(page: Page, name: string): Promise<void> {
	try {
		fs.mkdirSync(SCREENSHOT_DIR, { recursive: true });
		const ts   = new Date().toISOString().replace(/[:.]/g, '-');
		const file = path.join(SCREENSHOT_DIR, `${name}-${ts}.png`);
		await page.screenshot({ path: file, fullPage: false });
		log('WARN', 'screenshot saved', { file });
	} catch { /* best-effort */ }
}

// ── Scrape ────────────────────────────────────────────────────────────────────

async function scrapeOnce(username: string, password: string, attempt: number): Promise<AttendanceRecord> {
	const browser = await chromium.launch(getBrowserOpts());
	const context = await browser.newContext(getContextOpts());
	context.setDefaultTimeout(ACTION_TIMEOUT);
	const page = await context.newPage();

	try {
		await loginToPusaka(page, username, password, 'scrape');
		// Block heavy assets only after login so login page renders with full CSS
		await blockAssets(page);

		const absensiLink = page.getByRole('link', { name: /Absensi/i }).first();
		await absensiLink.waitFor({ state: 'visible' });
		await absensiLink.click();

		const riwayatBtn = page.getByRole('button', { name: 'Riwayat Presensi', exact: true });
		await riwayatBtn.waitFor({ state: 'visible' });
		await riwayatBtn.click();

		await page.getByText(/Jam Masuk|Sedang mengambil riwayat/i).first().waitFor().catch(() => {});
		await page.waitForTimeout(2000);

		const parsed = parseTodayFromText(await page.locator('body').innerText(), getTodayLabelID());
		if (!parsed) throw new Error(`Data hari ini tidak ditemukan: ${getTodayLabelID()}`);
		return parsed;
	} catch (e) {
		await saveFailScreenshot(page, `scrape-fail-a${attempt}`);
		throw e;
	} finally {
		await context.close();
		await browser.close();
	}
}

// ── Checkin ───────────────────────────────────────────────────────────────────

async function checkin(username: string, password: string, attempt: number): Promise<void> {
	const label = 'checkin';
	const geo   = randomGeo(BASE_LAT, BASE_LNG, 28);
	const browser = await chromium.launch(getBrowserOpts());
	const context = await browser.newContext(getContextOpts({ geolocation: geo, permissions: ['geolocation'] }));
	context.setDefaultTimeout(ACTION_TIMEOUT);
	const page = await context.newPage();

	try {
		log('INFO', `${label}: starting`, { username, lat: geo.latitude.toFixed(7), lng: geo.longitude.toFixed(7) });
		await loginToPusaka(page, username, password, label);
		await page.goto(`${BASE_URL}/profile/presence`, { waitUntil: 'domcontentloaded' });
		await page.waitForTimeout(3000);
		await triggerGeo(page);
		await page.waitForTimeout(2000);

		const btn = page.locator('button:has-text("Presensi masuk")');
		if (await btn.count() === 0) {
			if (/hadir|sudah/i.test(await page.locator('body').innerText().catch(() => ''))) {
				log('WARN', `${label}: sudah absen`, { username }); return;
			}
			throw new Error('Tombol Presensi masuk tidak ditemukan');
		}
		if (await btn.first().isDisabled().catch(() => false)) {
			log('WARN', `${label}: tombol disabled`, { username }); return;
		}

		await triggerGeo(page);
		await page.waitForTimeout(1000);
		await btn.first().click();
		log('INFO', `${label}: tombol diklik`, { username });
		await page.waitForTimeout(3000);
		await assertPresenceResult(page, label, username);
	} catch (e) {
		await saveFailScreenshot(page, `${label}-fail-a${attempt}-${username}`);
		throw e;
	} finally {
		await context.close();
		await browser.close();
	}
}

// ── Checkout ──────────────────────────────────────────────────────────────────

async function checkout(username: string, password: string, attempt: number): Promise<void> {
	const label = 'checkout';
	const geo   = randomGeo(BASE_LAT, BASE_LNG, 28);
	const browser = await chromium.launch(getBrowserOpts());
	const context = await browser.newContext(getContextOpts({ geolocation: geo, permissions: ['geolocation'] }));
	context.setDefaultTimeout(ACTION_TIMEOUT);
	const page = await context.newPage();

	try {
		log('INFO', `${label}: starting`, { username, lat: geo.latitude.toFixed(7), lng: geo.longitude.toFixed(7) });
		await loginToPusaka(page, username, password, label);
		await page.goto(`${BASE_URL}/profile/presence`, { waitUntil: 'domcontentloaded' });
		await page.waitForTimeout(3000);
		await triggerGeo(page);
		await page.waitForTimeout(2000);

		const btn = page.locator('button:has-text("Presensi pulang")');
		if (await btn.count() === 0) throw new Error('Tombol Presensi pulang tidak ditemukan');
		if (await btn.first().isDisabled().catch(() => false)) {
			log('WARN', `${label}: tombol disabled`, { username }); return;
		}

		await btn.first().click();
		log('INFO', `${label}: tombol diklik`, { username });
		await page.waitForTimeout(2000);

		const yaBtn = page.getByRole('button', { name: /^ya$/i });
		if (await yaBtn.count() > 0 && !await yaBtn.first().isDisabled().catch(() => false)) {
			await yaBtn.first().click();
			log('INFO', `${label}: konfirmasi Ya diklik`, { username });
		}

		await page.waitForTimeout(3000);
		await assertPresenceResult(page, label, username);
	} catch (e) {
		await saveFailScreenshot(page, `${label}-fail-a${attempt}-${username}`);
		throw e;
	} finally {
		await context.close();
		await browser.close();
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

function randomGeo(baseLat: number, baseLng: number, radiusMeters = 10): GeoCoords {
	const R = 6371000;
	const latOff = (Math.random() - 0.5) * 2 * (radiusMeters / R) * (180 / Math.PI);
	const lngOff = ((Math.random() - 0.5) * 2 * (radiusMeters / R) * (180 / Math.PI)) / Math.cos((baseLat * Math.PI) / 180);
	return { latitude: baseLat + latOff, longitude: baseLng + lngOff };
}

function parseTodayFromText(text: string, todayLabel: string): AttendanceRecord | null {
	const lines = text.replace(/\r/g, '').split('\n').map((l) => l.trim()).filter(Boolean);
	const candidates = lines.reduce<number[]>((acc, l, i) => { if (l === todayLabel) acc.push(i); return acc; }, []);
	if (!candidates.length) return null;

	const idx = candidates[candidates.length - 1];
	const blockLines: string[] = [];
	for (let i = idx + 1; i < lines.length; i++) {
		if (/^(Senin|Selasa|Rabu|Kamis|Jumat|Sabtu|Minggu),\s+\d{2}\s+\w+\s+\d{4}$/.test(lines[i])) break;
		blockLines.push(lines[i]);
	}
	const block = blockLines.join('\n');
	const jamMasuk  = extractJam(block, 'Jam Masuk');
	const jamPulang = extractJam(block, 'Jam Pulang');
	return { tanggal: toISODateMakassar(), jam_masuk: jamMasuk === '-' ? '' : jamMasuk, jam_pulang: jamPulang === '-' ? '' : jamPulang };
}

function extractJam(block: string, label: string): string {
	const lines = block.split('\n').map((l) => l.trim()).filter(Boolean);
	for (let i = 0; i < lines.length; i++) {
		if (!lines[i].toLowerCase().includes(label.toLowerCase())) continue;
		const inline = lines[i].match(/([0-9]{2}:[0-9]{2}(?::[0-9]{2})?(?:\s*(?:WITA|WIB|WIT))?|-)/i);
		if (inline) return inline[1].replace(/\s+/g, ' ').trim();
		for (let j = i + 1; j < Math.min(lines.length, i + 4); j++) {
			const next = lines[j].match(/([0-9]{2}:[0-9]{2}(?::[0-9]{2})?(?:\s*(?:WITA|WIB|WIT))?|-)/i);
			if (next) return next[1].replace(/\s+/g, ' ').trim();
		}
	}
	return '';
}

function getTodayLabelID(): string {
	return new Intl.DateTimeFormat('id-ID', {
		weekday: 'long', day: '2-digit', month: 'long', year: 'numeric', timeZone: 'Asia/Makassar',
	}).format(new Date());
}

function toISODateMakassar(): string {
	return new Intl.DateTimeFormat('en-CA', {
		timeZone: 'Asia/Makassar', year: 'numeric', month: '2-digit', day: '2-digit',
	}).format(new Date());
}

function toTimeWITA(): string {
	return new Intl.DateTimeFormat('id-ID', {
		timeZone: 'Asia/Makassar', hour: '2-digit', minute: '2-digit', hour12: false,
	}).format(new Date());
}

function log(level: 'INFO' | 'WARN' | 'ERROR', message: string, context: Record<string, unknown> = {}): void {
	const line = `${new Date().toISOString()} [${level}] ${message} ${JSON.stringify(context)}`;
	try { fs.appendFileSync(LOG_PATH, `${line}\n`); } catch { /* no-op */ }
	if (level === 'ERROR')     console.error(line);
	else if (level === 'WARN') console.warn(line);
	else                       console.log(line);
}

function sleep(ms: number): Promise<void> {
	return new Promise((resolve) => setTimeout(resolve, ms));
}
