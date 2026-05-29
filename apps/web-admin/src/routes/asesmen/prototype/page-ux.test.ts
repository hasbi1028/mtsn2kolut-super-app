import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';

const pageSource = readFileSync('src/routes/asesmen/prototype/+page.svelte', 'utf8');
const modelSource = readFileSync('src/routes/asesmen/prototype/asesmen-prototype.model.ts', 'utf8');

describe('asesmen frontend-only prototype', () => {
	it('is clearly marked as a frontend prototype', () => {
		expect(pageSource).toContain('PROTOTYPE UI');
		expect(pageSource).toContain('belum terhubung backend');
		expect(pageSource).toContain('Belum menjalankan aksi data');
	});

	it('does not call backend, BFF, or mutate data', () => {
		for (const source of [pageSource, modelSource]) {
			expect(source).not.toContain('fetch(');
			expect(source).not.toContain('/api/asesmen');
			expect(source).not.toContain('/api/cbt');
			expect(source).not.toContain('/api/exam');
			expect(source).not.toContain('/api/cbt-portal');
			expect(source).not.toContain('method:');
		}
	});

	it('uses the numbered A0-A5 review flow', () => {
		expect(pageSource).toContain('A0 · Asesmen / Ujian Digital');
		for (const label of ['A1', 'A2', 'A3', 'A4', 'A5', 'Persiapan Ujian', 'Dokumen & Cetak', 'Pelaksanaan Hari-H', 'Ruang Saya', 'Hasil & Penutupan']) {
			expect(modelSource).toContain(label);
		}
	});

	it('keeps A9 as a secondary complete-mode lane', () => {
		expect(pageSource).toContain('A9 · Mode Lengkap Panitia');
		expect(modelSource).toContain('advancedPrototypeLinks');
		expect(modelSource).toContain('hiddenFromMainFlow');
		expect(pageSource).toContain('Disembunyikan dari alur utama');
	});

	it('adds compact operational details without becoming a production console', () => {
		expect(pageSource).toContain('Daftar kerja ringkas');
		expect(pageSource).toContain('Preview Ruang');
		expect(modelSource).toContain('8 ruang');
		expect(modelSource).toContain('Susun 8 ruang otomatis');
		expect(modelSource).toContain('Hubungi Admin');
	});

	it('follows the old CBT UI vocabulary while excluding Bank Soal as the main reference', () => {
		expect(pageSource).toContain('Mengikuti CBT lama');
		expect(pageSource).toContain('Struktur menu familiar');
		expect(pageSource).toContain('Prinsip yang dibawa ke Super App');
		expect(modelSource).toContain('Pusat Data');
		expect(modelSource).toContain('Ruang Ujian');
		expect(modelSource).toContain('Proctoring Live');
		expect(modelSource).toContain('Rekap Nilai');
		expect(modelSource).toContain('Bank Soal tidak dijadikan contoh utama prototype');
	});

	it('keeps the prototype compact and avoids demo/local CTA', () => {
		expect(pageSource).toContain('Micro');
		expect(pageSource).not.toContain('demo=1');
		expect(pageSource).not.toContain('Latihan Lokal');
		expect(pageSource).not.toContain('MODE DEMO');
	});
});
