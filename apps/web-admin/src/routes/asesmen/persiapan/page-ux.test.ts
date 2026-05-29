import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';

const pageSource = readFileSync('src/routes/asesmen/persiapan/+page.svelte', 'utf8');

describe('Persiapan page UX', () => {
	it('keeps the prep flow numbered, short, and aligned with the sidebar work codes', () => {
		expect(pageSource).toContain('code="7.1"');
		expect(pageSource).toContain('Kerjakan berurutan: kegiatan, paket, sesi, ruang dan peserta, lalu pengawas.');
		expect(pageSource).toContain('Langkah persiapan');
		expect(pageSource).toContain('7.1.1');
		expect(pageSource).toContain('7.1.2');
		expect(pageSource).toContain('7.1.3');
		expect(pageSource).toContain('7.1.4');
		expect(pageSource).toContain('7.1.5');
		expect(pageSource).toContain('Buka kegiatan');
		expect(pageSource).toContain('Buka paket');
		expect(pageSource).toContain('Buka sesi');
		expect(pageSource).toContain('Atur ruang');
		expect(pageSource).toContain('Atur pengawas');
		expect(pageSource).toContain('Catatan teknis persiapan');
	});
});
