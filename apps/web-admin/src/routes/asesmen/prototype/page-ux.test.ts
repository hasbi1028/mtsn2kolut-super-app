import { describe, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';

const pageSource = readFileSync('src/routes/asesmen/prototype/+page.svelte', 'utf8');
const modelSource = readFileSync('src/routes/asesmen/prototype/asesmen-prototype.model.ts', 'utf8');

describe('asesmen CBT reference prototype', () => {
	it('is clearly marked as a frontend-only prototype', () => {
		expect(pageSource).toContain('PROTOTYPE UI');
		expect(pageSource).toContain('belum terhubung backend');
		expect(pageSource).toContain('Belum menjalankan aksi data');
		expect(modelSource).toContain('frontend-only');
	});

	it('does not call backend, BFF, or mutate data', () => {
		for (const source of [pageSource, modelSource]) {
			expect(source).not.toContain('fetch(');
			expect(source).not.toContain('/api/asesmen');
			expect(source).not.toContain('/api/cbt');
			expect(source).not.toContain('/api/exam');
			expect(source).not.toContain('/api/cbt-portal');
			expect(source).not.toContain('method:');
		}
	});

	it('was rebuilt from the legacy CBT domain and cbt-ujian repository references', () => {
		expect(pageSource).toContain('cbt-ujian');
		expect(modelSource).toContain('cbt.mtsn2kolut.sch.id');
		expect(modelSource).toContain('/home/servermtsn2kolut/cbt-ujian');
	});

	it('uses the old CBT command-center/sidebar/table vocabulary instead of the previous micro workflow', () => {
		expect(pageSource).toContain('Blueprint Final CBT Web-First');
		expect(pageSource).toContain('Struktur menu CBT prototype');
		expect(pageSource).toContain('Tabel operasional');
		for (const label of ['Pusat Data', 'Manajemen Ujian', 'Pelaksanaan', 'Laporan']) {
			expect(modelSource).toContain(label);
		}
	});

	it('turns the prototype into a concrete final blueprint surface map', () => {
		for (const label of ['01 · Dashboard CBT', '02 · Ruang Ujian & Kartu QR', '03 · Portal mobile-web', '04 · Proctoring Live & Rekap Nilai']) {
			expect(pageSource).toContain(label);
		}
		for (const modelAnchor of ['readinessCards', 'todayAgenda', 'roomBlueprints', 'cardPrintSteps', 'mobileScreens', 'resultSummaries']) {
			expect(modelSource).toContain(modelAnchor);
		}
	});

	it('keeps familiar CBT operational surfaces excluding Bank Soal as the main focus', () => {
		for (const label of ['Ruang Ujian', 'Jadwal Sesi', 'Proctoring Live', 'Rekap Nilai', 'Kartu QR']) {
			expect(modelSource).toContain(label);
		}
		expect(modelSource).toContain('Bank Soal sengaja tidak dijadikan fokus prototype ini');
		expect(modelSource).toContain('Bank Soal lama tidak ditiru di prototype ini');
	});

	it('includes the standalone QR portal concepts from cbt-ujian', () => {
		expect(pageSource).toContain('Portal QR dari repo cbt-ujian');
		for (const label of ['Portal depan', 'Portal siswa', 'Portal pengawas', '/ujian/[token]', '/pengawas/[token]']) {
			expect(modelSource).toContain(label);
		}
		expect(modelSource).toContain('Masuk ujian cukup scan kartu QR');
		expect(modelSource).toContain('Monitoring peserta per ruang');
	});

	it('adds web-first QR+PIN and room-supervision blueprint details', () => {
		for (const label of ['Kartu Peserta Ujian', 'Lembar Pengawas Ruang', 'Portal Ujian Peserta', 'Portal Pengawasan Ruang', 'Ruang · Peringatan · Peserta', 'Hubungi Admin']) {
			expect(modelSource).toContain(label);
		}
		expect(pageSource).toContain('Siswa dan pengawas terasa seperti aplikasi HP');
	});

	it('does not keep the old A0-A9 prototype anchors or demo CTA', () => {
		expect(pageSource).not.toContain('A0 · Asesmen / Ujian Digital');
		expect(pageSource).not.toContain('A9 · Mode Lengkap Panitia');
		expect(modelSource).not.toContain('advancedPrototypeLinks');
		expect(modelSource).not.toContain('hiddenFromMainFlow');
		expect(pageSource).not.toContain('demo=1');
		expect(pageSource).not.toContain('Latihan Lokal');
		expect(pageSource).not.toContain('MODE DEMO');
	});
});
