import { describe, expect, it } from 'vitest';
import { existsSync, readFileSync } from 'node:fs';
import { resolve } from 'node:path';

const scriptPath = resolve(process.cwd(), 'scripts/asesmen-cbt-opsi0-e2e.mjs');
const docsPath = resolve(process.cwd(), 'docs/asesmen-cbt-opsi0-e2e.md');
const pkgPath = resolve(process.cwd(), 'package.json');

describe('Asesmen CBT Opsi 0 E2E smoke coverage contract', () => {
	it('keeps the Opsi 0 smoke script, docs, and npm script discoverable', () => {
		expect(existsSync(scriptPath)).toBe(true);
		expect(existsSync(docsPath)).toBe(true);

		const script = readFileSync(scriptPath, 'utf8');
		expect(script).toContain('WEB_ADMIN_ASESMEN_CBT_OPSI0_BASE_URL');
		expect(script).toContain('/api/exam/login');
		expect(script).toContain('assertNoSensitiveExamKeys');
		expect(script).toContain('WEB_ADMIN_ASESMEN_CBT_OPSI0_DESTRUCTIVE');

		const docs = readFileSync(docsPath, 'utf8');
		expect(docs).toContain('Opsi 0');
		expect(docs).toContain('TEST_DATABASE_URL');
		expect(docs).toContain('tidak memakai data produksi');

		const pkg = JSON.parse(readFileSync(pkgPath, 'utf8'));
		expect(pkg.scripts['smoke:asesmen-cbt:opsi0']).toBe(
			'node scripts/asesmen-cbt-opsi0-e2e.mjs'
		);
	});
});
