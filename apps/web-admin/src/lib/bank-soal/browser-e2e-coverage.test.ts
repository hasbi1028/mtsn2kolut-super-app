/// <reference types="node" />

import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { describe, expect, it } from 'vitest';

const scriptPath = resolve(process.cwd(), 'scripts/bank-soal-role-e2e.mjs');
const script = () => readFileSync(scriptPath, 'utf8');

describe('Bank Soal browser E2E smoke coverage contract', () => {
	it('covers all final Bank Soal route groups and role personas', () => {
		const source = script();
		for (const route of [
			'/bank-soal',
			'/bank-soal/daftar',
			'/bank-soal/tambah',
			'/bank-soal/verifikasi',
			'/bank-soal/impor',
			'/bank-soal/analisis-butir',
			'/bank-soal/mapel-kd',
			'/bank-soal/pengaturan',
		]) {
			expect(source, `missing route coverage for ${route}`).toContain(route);
		}
		for (const persona of ['admin', 'creator', 'reviewer', 'importer', 'readonly', 'no-access']) {
			expect(source, `missing persona coverage for ${persona}`).toContain(`'${persona}'`);
		}
	});

	it('asserts permission-aware controls and unauthorized boundaries', () => {
		const source = script();
		expect(source).toContain('assertVisibleCreateControls');
		expect(source).toContain('assertHiddenCreateControls');
		expect(source).toContain('assertReachable');
		expect(source).toContain('assertForbidden');
		expect(source).toContain('/api/bank-soal/summary');
		expect(source).toContain('/api/bank-soal/questions');
		expect(source).toContain('/api/bank-soal/assets');
		expect(source).toContain('/api/cbt/questions');
	});

	it('can opt into the controlled DB seed defaults for manual smoke runs', () => {
		const source = script();
		expect(source).toContain('WEB_ADMIN_BANK_SOAL_E2E_USE_SEEDED_DEFAULTS');
		expect(source).toContain('BANK_SOAL_E2E_PASSWORD');
		expect(source).toContain('seededDefaultPassword = process.env.BANK_SOAL_E2E_PASSWORD');
		expect(source).not.toMatch(/seededDefaultPassword\s*=\s*process\.env\.BANK_SOAL_E2E_PASSWORD\s*\?\?/);
		for (const persona of ['ADMIN', 'CREATOR', 'REVIEWER', 'IMPORTER', 'READONLY', 'NO_ACCESS']) {
			expect(source).toContain(`BANK_SOAL_E2E_${persona}_USERNAME`);
			expect(source).toContain(`BANK_SOAL_E2E_${persona}_PASSWORD`);
		}
	});
});
