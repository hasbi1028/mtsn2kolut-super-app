import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';

const pageSource = readFileSync('src/routes/asesmen/kegiatan/[id]/+page.svelte', 'utf8');

describe('Detail kegiatan asesmen UX', () => {
	it('keeps one ringkasan surface and removes redundant blocker panel shortcuts', () => {
		expect(pageSource).toContain('Langkah berikutnya');
		expect(pageSource).toContain('Kesiapan kegiatan');
		expect(pageSource).toContain('Timeline SOP Kegiatan');
		expect(pageSource).toContain('EntityTabs');
		expect(pageSource).not.toContain('BlockerPanel');
		expect(pageSource).not.toContain('onclick={() => activeSection = group.id}');
	});
});
