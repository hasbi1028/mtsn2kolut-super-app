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
		expect(pageSource).not.toContain('fetch(');
		expect(pageSource).not.toContain('/api/asesmen');
		expect(pageSource).not.toContain('/api/cbt');
		expect(pageSource).not.toContain('method:');
		expect(modelSource).not.toContain('/api/asesmen');
		expect(modelSource).not.toContain('/api/cbt');
	});

	it('uses the numbered A0-A5 review flow', () => {
		expect(pageSource).toContain('A0 · Asesmen / Ujian Digital');
		for (const label of ['A1', 'A2', 'A3', 'A4', 'A5', 'Persiapan Ujian', 'Dokumen & Cetak', 'Pelaksanaan Hari-H', 'Ruang Saya', 'Hasil & Penutupan']) {
			expect(modelSource).toContain(label);
		}
	});

	it('keeps the prototype compact and avoids demo/local CTA', () => {
		expect(pageSource).toContain('Micro');
		expect(pageSource).not.toContain('demo=1');
		expect(pageSource).not.toContain('Latihan Lokal');
		expect(pageSource).not.toContain('MODE DEMO');
	});
});
