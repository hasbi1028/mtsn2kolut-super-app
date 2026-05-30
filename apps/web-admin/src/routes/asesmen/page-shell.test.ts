import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';

const pageSource = readFileSync(join(process.cwd(), 'src/routes/asesmen/+page.svelte'), 'utf8');

describe('/asesmen production shell', () => {
	it('uses the simple CBT command-center copy without backend coupling', () => {
		expect(pageSource).toContain('Command Center CBT');
		expect(pageSource).toContain('CBT Web');
		expect(pageSource).toContain('Alur sederhana untuk panitia');
		expect(pageSource).not.toContain('fetch(');
		expect(pageSource).not.toContain('/api/');
	});

	it('shows only the four primary preparation lanes and marks future actions inactive', () => {
		for (const label of ['Siapkan Ujian', 'Atur Peserta & Ruang', 'Cetak Kartu & Lembar Pengawas', 'Pelaksanaan & Hasil']) {
			expect(pageSource).toContain(label);
		}
		expect(pageSource).toContain('Tahap berikutnya');
		expect(pageSource).toContain('Belum aktif');
	});

	it('keeps Bank Soal separate and exposes the prototype as a review-only link', () => {
		expect(pageSource).toContain('Bank Soal tetap modul terpisah');
		expect(pageSource).toContain('/asesmen/prototype');
		expect(pageSource).toContain('Lihat Prototype');
	});
});
