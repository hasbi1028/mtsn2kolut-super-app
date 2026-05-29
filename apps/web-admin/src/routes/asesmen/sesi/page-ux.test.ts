import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';

const pageSource = readFileSync('src/routes/asesmen/sesi/+page.svelte', 'utf8');

describe('Sesi page UX', () => {
	it('keeps the session list focused on one clear detail path', () => {
		expect(pageSource).toContain('Untuk edit sesi, atur jadwal, dan manual ruangan, buka detail sesi.');
		expect(pageSource).toContain('Lanjutkan setup');
		expect(pageSource).toContain('Siap jadwal');
		expect(pageSource).toContain('Pantau sesi');
		expect(pageSource).toContain('BA Sesi');
	});
});
