import { describe, expect, it } from 'vitest';
import asesmenHubSource from './+page.svelte?raw';
import asesmenWorkflowHubSource from '$lib/asesmen/workflow-hub?raw';
import pelaksanaanSource from './pelaksanaan/+page.svelte?raw';
import aplikasiSiswaSource from './aplikasi-siswa/+page.svelte?raw';

describe('simplified Asesmen terminology', () => {
	it('keeps the main CBT hub free of day-of jargon', () => {
		expect(asesmenHubSource).toContain('Command Center CBT');
		expect(`${asesmenHubSource}
${asesmenWorkflowHubSource}`).toContain('sebelum pelaksanaan');
		expect(asesmenHubSource).not.toContain('hari-H');
		// Keep internal route names stable, but field-facing title/copy should be simple.
		expect(asesmenHubSource).not.toContain('Browser Darurat');
	});

	it('uses Portal Peserta copy in the pelaksanaan lane', () => {
		expect(pelaksanaanSource).toContain('Pilih jalur pelaksanaan');
		expect(pelaksanaanSource).toContain("title: 'Portal Peserta'");
		expect(pelaksanaanSource).toContain('kesulitan masuk Portal Ujian');
		expect(pelaksanaanSource).not.toContain('Pilih jalur kerja hari-H');
		expect(pelaksanaanSource).not.toContain("title: 'Perangkat Siswa'");
	});

	it('uses QR+PIN and app-archive wording in the participant portal guide', () => {
		expect(aplikasiSiswaSource).toContain('Panduan Portal Peserta');
		expect(aplikasiSiswaSource).toContain('Scan QR Kartu Peserta Ujian lalu masukkan PIN');
		expect(aplikasiSiswaSource).toContain('Arsip Aplikasi Lama');
		expect(aplikasiSiswaSource).not.toContain('token peserta dan kode ruang');
		expect(aplikasiSiswaSource).not.toContain('Panduan Perangkat Siswa');
	});
});
