import { readFileSync } from 'node:fs';
import path from 'node:path';
import { describe, expect, it } from 'vitest';

const pageSource = () => readFileSync(path.resolve(process.cwd(), 'src/routes/asesmen/+page.svelte'), 'utf8');

describe('/asesmen Step 6 room and participant workflow', () => {
	it('exposes room planning controls and calls preview/apply endpoints from the detail drawer', () => {
		const source = pageSource();
		expect(source).toContain('Step 6 · Ruang & Peserta');
		expect(source).toContain('/api/academic/rombel');
		expect(source).toContain('/assignment-preview');
		expect(source).toContain('/assignment-apply');
		expect(source).toContain('Kartu/QR+PIN belum diterbitkan');
	});

	it('keeps a granular manual mode for operator-controlled seat shuffling', () => {
		const source = pageSource();
		expect(source).toContain('Mode Manual · Acak Sendiri');
		expect(source).toContain('/participants');
		expect(source).toContain('/participants/seat');
		expect(source).toContain('Acak Tampilan');
		expect(source).toContain('Pindah');
	});

	it('uses the recommended 4-stage wizard with room visual review', () => {
		const source = pageSource();
		expect(source).toContain('① Pilih Rombel');
		expect(source).toContain('② Atur Acak');
		expect(source).toContain('③ Review Ruang');
		expect(source).toContain('④ Manual');
		expect(source).toContain('Peta Ruang Visual');
		expect(source).toContain('Simpan Penempatan');
	});
});
