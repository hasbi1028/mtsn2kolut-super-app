import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';

const pageSource = readFileSync(join(process.cwd(), 'src/routes/asesmen/prototype/+page.svelte'), 'utf8');

describe('/asesmen/prototype UI-only contract', () => {
	it('stays frontend-only without backend or API coupling', () => {
		expect(pageSource).not.toContain('fetch(');
		expect(pageSource).not.toContain('/api/');
		expect(pageSource).toContain('PROTOTYPE UI');
		expect(pageSource).toContain('belum terhubung backend');
	});

	it('shows familiar CBT review anchors and separates Bank Soal from Asesmen', () => {
		for (const label of ['Mengikuti CBT lama', 'Ruang Ujian', 'Jadwal Sesi', 'Paket Ujian', 'Proctoring Live', 'Rekap Nilai']) {
			expect(pageSource).toContain(label);
		}
		expect(pageSource).toContain('Bank Soal tetap modul terpisah');
	});

	it('orders A0-A5 first and keeps A9 as a secondary complete-admin lane', () => {
		const anchors = ['A0', 'A1', 'A2', 'A3', 'A4', 'A5', 'A9'];
		const positions = anchors.map((anchor) => pageSource.indexOf(anchor));
		for (const position of positions) {
			expect(position).toBeGreaterThanOrEqual(0);
		}
		for (let i = 1; i < positions.length; i += 1) {
			expect(positions[i]).toBeGreaterThan(positions[i - 1]);
		}
		expect(pageSource).toContain('A9 — Mode Lengkap Admin');
		expect(pageSource).toContain('sekunder');
	});

	it('previews admin, student, and proctor flows', () => {
		for (const label of ['Preview Panitia', 'Preview Peserta', 'Preview Pengawas']) {
			expect(pageSource).toContain(label);
		}
	});
});
