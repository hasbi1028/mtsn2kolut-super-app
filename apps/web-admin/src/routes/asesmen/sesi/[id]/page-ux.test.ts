import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';

const pageSource = readFileSync('src/routes/asesmen/sesi/[id]/+page.svelte', 'utf8');

describe('Detail sesi UX', () => {
	it('makes manual room assignment clearer and less noisy', () => {
		expect(pageSource).toContain('Ruangan Manual & Pengawas');
		expect(pageSource).toContain('Atur Ruangan Manual');
		expect(pageSource).toContain('Pilih ruang peserta di kolom kanan, isi nomor meja, lalu simpan.');
		expect(pageSource).toContain('Simpan penempatan');
		expect(pageSource).toContain('Token');
		expect(pageSource).toContain('Bagi Peserta Otomatis');
		expect(pageSource).toContain('Atur Meja Otomatis');
	});
});
