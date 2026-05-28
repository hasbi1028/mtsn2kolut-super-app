import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';

const ringkasPage = readFileSync('src/routes/asesmen/ringkas/+page.svelte', 'utf8');
const panitiaPage = readFileSync('src/routes/asesmen/panitia/+page.svelte', 'utf8');
const pelaksanaanPage = readFileSync('src/routes/asesmen/pelaksanaan/+page.svelte', 'utf8');

describe('asesmen alias and access-copy UX', () => {
	it('uses Ringkasan Asesmen wording instead of Dashboard Asesmen on aliases', () => {
		expect(ringkasPage).toContain('Mengalihkan ke Ringkasan Asesmen');
		expect(ringkasPage).not.toContain('Dashboard Asesmen');
		expect(panitiaPage).toContain('Mengalihkan ke Ringkasan Asesmen');
		expect(panitiaPage).not.toContain('Dashboard Asesmen');
		expect(pelaksanaanPage).toContain('Silakan kembali ke Ringkasan Asesmen');
		expect(pelaksanaanPage).toContain('Ringkasan Asesmen');
		expect(pelaksanaanPage).not.toContain('Dashboard Asesmen');
	});
});
