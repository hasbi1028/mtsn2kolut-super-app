import { describe, expect, it } from 'vitest';

import { summarizeAssessmentPrintDocuments } from './document-print-readiness';

describe('summarizeAssessmentPrintDocuments', () => {
	it('keeps participant cards disabled until rooms and participants are ready', () => {
		const [participantCards] = summarizeAssessmentPrintDocuments({
			room_count: 8,
			participant_count: 0,
			card_count: 0
		});

		expect(participantCards.id).toBe('participant-cards');
		expect(participantCards.ready).toBe(false);
		expect(participantCards.primaryAction).toBe('Lengkapi Peserta');
		expect(participantCards.detail).toContain('belum ada peserta');
	});

	it('allows guarded participant card issuing only after participants are placed', () => {
		const [participantCards] = summarizeAssessmentPrintDocuments({
			room_count: 8,
			participant_count: 224,
			card_count: 0
		});

		expect(participantCards.ready).toBe(true);
		expect(participantCards.status).toBe('Siap diterbitkan');
		expect(participantCards.primaryAction).toBe('Terbitkan QR+PIN');
		expect(participantCards.dangerous).toBe(true);
	});

	it('marks participant cards printable after cards already exist', () => {
		const [participantCards, supervisorSheets, archiveChecklist] = summarizeAssessmentPrintDocuments({
			room_count: 8,
			participant_count: 224,
			card_count: 224
		});

		expect(participantCards.status).toBe('Siap cetak');
		expect(participantCards.primaryAction).toBe('Cetak Kartu');
		expect(participantCards.dangerous).toBe(false);
		expect(supervisorSheets.ready).toBe(true);
		expect(supervisorSheets.countLabel).toBe('8 ruang');
		expect(archiveChecklist.primaryAction).toBe('Unduh Checklist');
	});
});
