import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';

const pageSource = readFileSync('src/routes/asesmen/kegiatan/[id]/+page.svelte', 'utf8');

describe('Detail kegiatan asesmen UX', () => {
	it('uses a simple launcher layout instead of confusing tabs', () => {
		expect(pageSource).toContain('Langkah utama kegiatan asesmen');
		expect(pageSource).toContain('Langkah berikutnya');
		expect(pageSource).toContain('Mode Lengkap');
		expect(pageSource).toContain('Dokumen & Cetak');
		expect(pageSource).not.toContain('EntityTabs');
		expect(pageSource).not.toContain('sectionTabs');
		expect(pageSource).not.toContain('activeSection');
	});
});
