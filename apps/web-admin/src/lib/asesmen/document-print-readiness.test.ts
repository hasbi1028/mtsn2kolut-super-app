import { describe, expect, it } from 'vitest';

import { summarizeDocumentPrintStatus } from './document-print-readiness';

describe('summarizeDocumentPrintStatus', () => {
	it('blocks participant cards until peserta and room placement are ready', () => {
		expect(summarizeDocumentPrintStatus({ participantCount: 0, roomCount: 0, cardCount: 0 }).participantCards).toMatchObject({
			state: 'blocked',
			label: 'Lengkapi Peserta'
		});
		expect(summarizeDocumentPrintStatus({ participantCount: 12, roomCount: 0, cardCount: 0 }).participantCards).toMatchObject({
			state: 'blocked',
			label: 'Simpan ruang dulu'
		});
	});

	it('marks participant cards as issuable then printable after QR and PIN exist', () => {
		expect(summarizeDocumentPrintStatus({ participantCount: 24, roomCount: 4, cardCount: 0 }).participantCards).toMatchObject({
			state: 'action',
			label: 'Terbitkan QR+PIN'
		});
		expect(summarizeDocumentPrintStatus({ participantCount: 24, roomCount: 4, cardCount: 24 }).participantCards).toMatchObject({
			state: 'ready',
			label: 'Cetak Kartu'
		});
	});

	it('keeps supervisor sheets room-level without participant PIN exposure', () => {
		const summary = summarizeDocumentPrintStatus({ participantCount: 24, roomCount: 4, cardCount: 0 });
		expect(summary.supervisorSheets).toMatchObject({
			state: 'ready',
			label: 'Siap dicetak'
		});
		expect(summary.supervisorSheets.description.toLowerCase()).not.toContain('pin');
	});
});
