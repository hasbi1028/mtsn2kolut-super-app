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

	it('offers operator-friendly randomization choices with a recommended default', () => {
		const source = pageSource();
		expect(source).toContain('Acak merata ke seluruh ruang');
		expect(source).toContain('Acak campur rombel');
		expect(source).toContain('Kelompok per rombel');
		expect(source).toContain('Urut nomor peserta');
		expect(source).toContain('Manual dari CSV');
		expect(source).toContain('Seimbangkan jumlah peserta per ruang');
		expect(source).toContain('Usahakan rombel tidak berkumpul');
	});

	it('supports CSV template export and import through preview-first manual placement', () => {
		const source = pageSource();
		expect(source).toContain('downloadPlacementTemplateCsv');
		expect(source).toContain('handlePlacementCsvImport');
		expect(source).toContain('Download Template CSV');
		expect(source).toContain('Import CSV');
		expect(source).toContain('student_id,nomor_peserta,nama,rombel,ruang,urutan');
		expect(source).toContain('Import → Validasi → Preview → Simpan');
	});

	it('does not send UI-only toggle fields that the backend JSON decoder rejects', () => {
		const source = pageSource();
		const payloadStart = source.indexOf('function assignmentPayload()');
		const payloadEnd = source.indexOf('function csvCell', payloadStart);
		const payloadSource = source.slice(payloadStart, payloadEnd);
		expect(payloadSource).toContain('mix_policy: backendMixPolicy()');
		expect(payloadSource).not.toContain('balance_rooms');
		expect(payloadSource).not.toContain('spread_rombel');
	});
	describe('/asesmen Step 7 document and print workflow', () => {
		it('exposes a compact Dokumen & Cetak panel with guarded QR+PIN issuance', () => {
			const source = pageSource();
			expect(source).toContain('Step 7 · Dokumen & Cetak');
			expect(source).toContain('Kartu Peserta');
			expect(source).toContain('Lembar Pengawas Ruang');
			expect(source).toContain('Checklist Arsip');
			expect(source).toContain('/cards');
			expect(source).toContain('/issue-cards');
			expect(source).toContain('Terbitkan QR+PIN');
			expect(source).toContain('PIN hanya tampil pada hasil terbitkan');
		});
	});

});
