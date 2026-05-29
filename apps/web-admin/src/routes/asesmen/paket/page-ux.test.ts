import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';

const pageSource = readFileSync('src/routes/asesmen/paket/+page.svelte', 'utf8');

describe('Paket page UX', () => {
	it('keeps the package hub short and task oriented', () => {
		expect(pageSource).toContain('Pilih paket siap pakai, lalu tautkan ke kegiatan bila perlu.');
		expect(pageSource).toContain('Ringkasan');
		expect(pageSource).toContain('Buat paket');
		expect(pageSource).toContain('Pakai di kegiatan');
		expect(pageSource).toContain('Catatan singkat');
		expect(pageSource).toContain('Halaman ini hanya memilih paket siap pakai dan paket khusus kegiatan ini.');
	});
});
