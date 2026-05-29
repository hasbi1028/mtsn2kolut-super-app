import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';

const pageSource = readFileSync('src/routes/asesmen/persiapan/+page.svelte', 'utf8');

describe('Persiapan page UX', () => {
	it('keeps the prep flow vertical, short, and step based', () => {
		expect(pageSource).toContain('Kerjakan berurutan: kegiatan, paket, lalu sesi dan ruang.');
		expect(pageSource).toContain('Langkah persiapan');
		expect(pageSource).toContain('Buka kegiatan');
		expect(pageSource).toContain('Buka paket');
		expect(pageSource).toContain('Buka sesi');
		expect(pageSource).toContain('Pembagian ruang mengikuti data sesi.');
		expect(pageSource).toContain('Simulasi, gladi, dan ujian nyata tetap memakai data kegiatan, paket, dan sesi server.');
	});
});
