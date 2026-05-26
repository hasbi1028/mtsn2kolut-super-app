import { describe, expect, it } from 'vitest';

import { summarizeDocumentPrintStatus, summarizeSupervisorCards } from './document-print-readiness';

describe('document print readiness', () => {
	it('marks supervisor cards ready only when QR PIN and room exist', () => {
		expect(summarizeSupervisorCards([{ room_name: 'Ruang 1', token: 'abc', pin: '1234' }])).toMatchObject({ state: 'ready', label: 'Siap cetak' });
		expect(summarizeSupervisorCards([{ room_name: 'Ruang 1', token: '', pin: '' }])).toMatchObject({ state: 'warning', nextAction: 'Terbitkan QR+PIN Ruang' });
	});

	it('combines participant and supervisor blockers', () => {
		const result = summarizeDocumentPrintStatus({
			examCards: [{ participant_id: 'p1', student_name: 'A', room_name: '', seat_no: null, token: '' }],
			supervisorCards: [{ room_name: 'Ruang 1', token: 'abc', pin: '1234' }],
			overview: { room_count: 1 }
		});
		expect(result.state).toBe('warning');
		expect(result.participantCards.label).toBe('0/1 siap');
		expect(result.blockers.length).toBeGreaterThan(0);
	});
});
