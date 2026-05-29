import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';

const pageSource = readFileSync('src/routes/asesmen/ruang-saya/+page.svelte', 'utf8');

describe('Ruang Saya page UX', () => {
	it('keeps the proctor screen focused on room work instead of committee setup', () => {
		expect(pageSource).toContain('Halaman kerja pengawas ruang.');
		expect(pageSource).toContain('Cara pakai singkat');
		expect(pageSource).toContain('Pengawas cukup buka satu kartu ruang.');
		expect(pageSource).toContain('Pengaturan sesi, peserta, dan pembagian ruang tetap dikerjakan panitia dari halaman Sesi.');
		expect(pageSource).toContain('Buka Panel Ruang');
		expect(pageSource).not.toContain('Mulai Ujian');
	});
});
