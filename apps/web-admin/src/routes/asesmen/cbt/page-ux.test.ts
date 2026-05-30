import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';

const root = process.cwd();
const layoutSource = readFileSync(join(root, 'src/routes/asesmen/cbt/+layout.svelte'), 'utf8');
const dashboardSource = readFileSync(join(root, 'src/routes/asesmen/cbt/+page.svelte'), 'utf8');
const bankSoalSource = readFileSync(join(root, 'src/routes/asesmen/cbt/bank-soal/+page.svelte'), 'utf8');
const ruangSource = readFileSync(join(root, 'src/routes/asesmen/cbt/ruang/+page.svelte'), 'utf8');
const paketSource = readFileSync(join(root, 'src/routes/asesmen/cbt/paket/+page.svelte'), 'utf8');
const sesiSource = readFileSync(join(root, 'src/routes/asesmen/cbt/sesi/+page.svelte'), 'utf8');
const proctoringSource = readFileSync(join(root, 'src/routes/asesmen/cbt/proctoring/+page.svelte'), 'utf8');
const rekapSource = readFileSync(join(root, 'src/routes/asesmen/cbt/rekap/+page.svelte'), 'utf8');
const docsSource = readFileSync(join(root, 'src/routes/asesmen/cbt/dokumen/+page.svelte'), 'utf8');
const componentSource = readFileSync(join(root, 'src/routes/asesmen/cbt/CbtFeatureSurface.svelte'), 'utf8');
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

	it('adds internal operational feature panels, not only static page shells', () => {
		expect(componentSource).toContain('Alur kerja');
		expect(componentSource).toContain('Fitur operasional');
		expect(dashboardSource).toContain('Quick Start Ujian');
		expect(ruangSource).toContain('assignment-preview');
		expect(ruangSource).toContain('assignment-apply');
		expect(ruangSource).toContain('participants/seat');
		expect(ruangSource).toContain('Simpan Pindahan');
		expect(paketSource).toContain('/api/asesmen/packages/options');
		expect(sesiSource).toContain('Kontrol buka/tutup');
		expect(proctoringSource).toContain('Antrian bantuan admin');
		expect(rekapSource).toContain('Koreksi essay');
		expect(docsSource).toContain('Lembar pengawas ruang');
	});

	it('keeps the dashboard production-facing and linked to the simplified command center', () => {
		expect(dashboardSource).toContain('aksi backend yang sudah ada');
		expect(dashboardSource).toContain('/asesmen');
		expect(dashboardSource).toContain('/pengawas-ujian?demo=1');
	});

	it('adjusts Bank Soal instead of copying the legacy SQLite Bank Soal module', () => {
		expect(bankSoalSource).toContain('Bank Soal Super App');
		expect(bankSoalSource).toContain('/bank-soal');
		expect(bankSoalSource).toContain('/bank-soal/tambah');
		expect(normalizedBankSoalSource).toContain('Bukan SQLite legacy');
	});

	it('keeps proctoring and documents safe for production operations', () => {
		expect(proctoringSource).toContain('Unlock/reset tidak di portal');
		expect(docsSource).toContain('QR+PIN tidak dibuat otomatis');
		expect(docsSource).toContain('/asesmen/dokumen');
	});
});
