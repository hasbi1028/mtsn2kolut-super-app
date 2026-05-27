import { describe, expect, it } from 'vitest';

import { summarizeDocumentPrintStatus, summarizeSupervisorCards } from './document-print-readiness';
import type { DocumentPrintOverview } from './document-print-readiness';

describe('document print readiness', () => {
	it('marks supervisor cards ready only when QR PIN and room exist', () => {
		expect(summarizeSupervisorCards([{ room_name: 'Ruang 1', token: 'abc', pin: '1234' }])).toMatchObject({ state: 'ready', label: 'Siap cetak' });
		expect(summarizeSupervisorCards([{ room_name: 'Ruang 1', token: '', pin: '' }])).toMatchObject({ state: 'warning', nextAction: 'Terbitkan QR+PIN Ruang' });
	});

	it('combines participant and supervisor blockers', () => {
		const overview: DocumentPrintOverview = { room_count: 1 };
		const result = summarizeDocumentPrintStatus({
			examCards: [{ participant_id: 'p1', student_name: 'A', room_name: '', seat_no: null, token: '' }],
			supervisorCards: [{ room_name: 'Ruang 1', token: 'abc', pin: '1234' }],
			overview
		});
		expect(result.state).toBe('warning');
		expect(result.participantCards.label).toBe('0/1 siap');
		expect(result.blockers.length).toBeGreaterThan(0);
		expect(result.recommendedNextAction).toBe('Perbaiki data kartu');
	});

	it('recommends final printing only when participant and supervisor documents are ready', () => {
		const result = summarizeDocumentPrintStatus({
			examCards: [{ participant_id: 'p1', student_name: 'A', room_name: 'Ruang 1', seat_no: 1, token: 'abcd1234', pin: '1234' }],
			supervisorCards: [{ room_name: 'Ruang 1', token: 'room1234', pin: '5678' }],
			overview: { room_count: 1, card_count: 1 }
		});
		expect(result.state).toBe('ready');
		expect(result.blockers).toEqual([]);
		expect(result.recommendedNextAction).toBe('Cetak dokumen kegiatan');
	});
});
