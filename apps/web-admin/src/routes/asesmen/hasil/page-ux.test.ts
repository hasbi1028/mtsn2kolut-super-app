import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';

const pageSource = readFileSync('src/routes/asesmen/hasil/+page.svelte', 'utf8');

describe('Hasil page UX', () => {
	it('uses numbered task cards instead of a dense action table', () => {
		expect(pageSource).toContain('code="7.3"');
		expect(pageSource).toContain('7.3.1');
		expect(pageSource).toContain('7.3.2');
		expect(pageSource).toContain('7.3.3');
		expect(pageSource).toContain('7.3.4');
		expect(pageSource).toContain('7.3.5');
		expect(pageSource).toContain('Catatan teknis hasil');
		expect(pageSource).not.toContain('MicroActionTable');
	});
});
