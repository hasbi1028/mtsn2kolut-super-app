import { existsSync } from 'node:fs';
import path from 'node:path';
import { describe, expect, it } from 'vitest';
import { sidebarNavGroups } from '$lib/components/sidebar/sidebar-config';

const SRC_ROOT = path.resolve(process.cwd(), 'src');
const ROUTES_ROOT = path.join(SRC_ROOT, 'routes');

const finalBankSoalRoutes = [
	'/bank-soal',
	'/bank-soal/daftar',
	'/bank-soal/tambah',
	'/bank-soal/verifikasi',
	'/bank-soal/penerbitan',
	'/bank-soal/impor',
	'/bank-soal/analisis-butir',
	'/bank-soal/mapel-kd',
	'/bank-soal/pengaturan'
];

function routeToPageFile(route: string): string {
	const relative = route.replace(/^\//, '');
	return path.join(ROUTES_ROOT, relative, '+page.svelte');
}

describe('Bank Soal final route map', () => {
	it('keeps every final Bank Soal page route backed by a Svelte page', () => {
		expect(finalBankSoalRoutes.map((route) => [route, existsSync(routeToPageFile(route))])).toEqual(
			finalBankSoalRoutes.map((route) => [route, true])
		);
	});

	it('keeps sidebar navigation aligned with the final Bank Soal route map', () => {
		const bankSoalItems = sidebarNavGroups.find((group) => group.group === 'Bank Soal')?.items ?? [];
		expect(bankSoalItems.map((item) => item.href)).toEqual(finalBankSoalRoutes);
		expect(new Set(bankSoalItems.map((item) => item.href)).size).toBe(finalBankSoalRoutes.length);
		expect(bankSoalItems.find((item) => item.href === '/bank-soal/verifikasi')?.permissions).toEqual(['bank_soal.review']);
		expect(bankSoalItems.find((item) => item.href === '/bank-soal/penerbitan')?.permissions).toEqual(['bank_soal.publish']);
		expect(bankSoalItems.find((item) => item.href === '/bank-soal/impor')?.permissions).toEqual(['bank_soal.import']);
		expect(bankSoalItems.find((item) => item.href === '/bank-soal/pengaturan')?.permissions).toEqual(['bank_soal.settings']);
	});

	it('keeps final routes free from legacy CBT/mode naming', () => {
		expect(finalBankSoalRoutes.every((route) => !route.startsWith('/cbt'))).toBe(true);
		expect(finalBankSoalRoutes.every((route) => !route.includes('?mode='))).toBe(true);
		expect(finalBankSoalRoutes).not.toContain('/bank-soal/komposer');
		expect(finalBankSoalRoutes).not.toContain('/bank-soal/import');
		expect(finalBankSoalRoutes).not.toContain('/bank-soal/review');
	});
});
