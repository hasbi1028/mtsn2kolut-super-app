import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';

const pageSource = readFileSync(join(process.cwd(), 'src/routes/asesmen/+page.svelte'), 'utf8');
const normalizedPageSource = pageSource.replace(/\s+/g, ' ');

describe('/asesmen production shell', () => {
	it('uses the simple CBT command-center copy with production-backed exam actions', () => {
		expect(pageSource).toContain('Command Center CBT');
		expect(pageSource).toContain('CBT Web');
		expect(pageSource).toContain('Alur sederhana untuk panitia');
		expect(normalizedPageSource).toContain('siapkan ujian, jalankan ruang, buka portal peserta, lalu tutup hasil');
		expect(pageSource).toContain('/api/asesmen/exams');
		expect(pageSource).toContain('Promise.allSettled');
	});

	it('shows the primary preparation lanes and safe document print lane', () => {
		for (const label of ['Persiapan', 'Paket & Jadwal', 'Ruang & Dokumen', 'Pelaksanaan & Hasil']) {
			expect(pageSource).toContain(label);
		}
		expect(pageSource).toContain('Dokumen & Cetak');
		expect(pageSource).toContain('Kartu peserta dan lembar pengawas');
		expect(normalizedPageSource).toContain('QR+PIN hanya dicoba diterbitkan saat panitia menekan tombol khusus');
		expect(pageSource).toContain('window.confirm');
	});

	it('keeps Bank Soal separate and exposes familiar/prototype review links', () => {
		expect(pageSource).toContain('Bank Soal tetap modul terpisah');
		expect(pageSource).toContain('/asesmen/cbt');
		expect(pageSource).toContain('Mode CBT Familiar');
		expect(pageSource).toContain('/asesmen/prototype');
		expect(pageSource).toContain('Lihat Prototype');
	});
});
