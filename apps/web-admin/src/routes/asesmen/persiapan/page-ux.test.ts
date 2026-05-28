import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';

const pageSource = readFileSync('src/routes/asesmen/persiapan/+page.svelte', 'utf8');

describe('/asesmen/persiapan UX', () => {
	it('keeps the page as a workflow step and not a hub duplicate', () => {
		expect(pageSource).toContain('Persiapan Ujian');
		expect(pageSource).toContain('Langkah Persiapan');
		expect(pageSource).toContain('Siapkan tiga hal inti sebelum hari ujian');
		expect(pageSource).toContain('Pembagian ruang bersifat dinamis per sesi');
		expect(pageSource).not.toContain('Kelola Sesi & Ruang');
		expect(pageSource).not.toContain('Masuk Pelaksanaan');
		expect(pageSource).not.toContain('Dashboard Asesmen');
		expect(pageSource).not.toContain('Pusat Persiapan');
	});
});
