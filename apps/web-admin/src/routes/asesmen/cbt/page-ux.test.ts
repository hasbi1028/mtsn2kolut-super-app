import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';

const root = process.cwd();
const layoutSource = readFileSync(join(root, 'src/routes/asesmen/cbt/+layout.svelte'), 'utf8');
const dashboardSource = readFileSync(join(root, 'src/routes/asesmen/cbt/+page.svelte'), 'utf8');
const bankSoalSource = readFileSync(join(root, 'src/routes/asesmen/cbt/bank-soal/+page.svelte'), 'utf8');
const proctoringSource = readFileSync(join(root, 'src/routes/asesmen/cbt/proctoring/+page.svelte'), 'utf8');
const docsSource = readFileSync(join(root, 'src/routes/asesmen/cbt/dokumen/+page.svelte'), 'utf8');
const normalizedBankSoalSource = bankSoalSource.replace(/\s+/g, ' ');

describe('Mode CBT Familiar production module', () => {
	it('copies the familiar CBT menu vocabulary into a Super App module', () => {
		for (const label of ['Pusat Data', 'Manajemen Soal', 'Pelaksanaan', 'Laporan', 'Pengaturan']) {
			expect(layoutSource).toContain(label);
		}
		for (const label of ['Dashboard Utama', 'Ruang Ujian', 'Paket Ujian', 'Jadwal Sesi', 'Proctoring Live', 'Rekap Nilai']) {
			expect(layoutSource).toContain(label);
		}
	});

	it('keeps the dashboard production-facing and linked to the simplified command center', () => {
		expect(dashboardSource).toContain('Panel familiar dari CBT lama');
		expect(dashboardSource).toContain('/asesmen/persiapan');
		expect(dashboardSource).toContain('/asesmen');
	});

	it('adjusts Bank Soal instead of copying the legacy SQLite Bank Soal module', () => {
		expect(bankSoalSource).toContain('Bank Soal Super App');
		expect(bankSoalSource).toContain('/bank-soal');
		expect(bankSoalSource).toContain('/bank-soal/tambah');
		expect(normalizedBankSoalSource).toContain('data dan workflow tetap milik Super App');
	});

	it('keeps proctoring and documents safe for production operations', () => {
		expect(proctoringSource).toContain('Reset login dan finalisasi tetap di mode panitia/admin');
		expect(docsSource).toContain('QR+PIN diterbitkan hanya saat panitia menekan tombol khusus');
		expect(docsSource).toContain('/asesmen/dokumen');
	});
});
