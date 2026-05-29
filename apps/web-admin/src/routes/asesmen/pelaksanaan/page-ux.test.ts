import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';

const pageSource = readFileSync('src/routes/asesmen/pelaksanaan/+page.svelte', 'utf8');

describe('Pelaksanaan page UX', () => {
	it('keeps day-of assessment navigation split by operator and proctor jobs', () => {
		expect(pageSource).toContain('Sesi Panitia');
		expect(pageSource).toContain('Untuk operator: cek jadwal hari ini, status sesi, ruang, peserta, pengawas, dan tindakan teknis panitia.');
		expect(pageSource).toContain('Ruang Saya');
		expect(pageSource).toContain('Untuk pengawas: pilih ruang yang ditugaskan, lihat kode ruang, pantau peserta, dan tangani atensi.');
		expect(pageSource).toContain('kendali panitia/operator');
		expect(pageSource).toContain('layar kerja pengawas');
		expect(pageSource).not.toContain('Pekerjaan Panitia');
		expect(pageSource).not.toContain('Cetak Kartu Peserta');
	});
});
