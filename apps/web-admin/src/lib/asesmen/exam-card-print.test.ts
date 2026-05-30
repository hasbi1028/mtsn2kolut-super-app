import { describe, expect, it } from 'vitest';
import { deriveCardReadiness, filterAndSortExamCards, summarizeCardsByRoom, type ExamCardPrintItem } from './exam-card-print';

const cards: ExamCardPrintItem[] = [
	{
		participant_id: 'p-blank',
		student_name: 'Belum Ruang',
		class_code: 'VII-A',
		room_code: '',
		seat_no: 0,
		status: 'not_issued'
	},
	{
		participant_id: 'p-2',
		student_name: 'Budi',
		class_code: 'VIII-B',
		room_code: 'R02',
		seat_no: 1,
		status: 'active',
		token: 'tok-budi',
		pin: '123456'
	},
	{
		participant_id: 'p-1',
		student_name: 'Ani',
		class_code: 'VII-A',
		room_code: 'R01',
		seat_no: 2,
		status: 'active',
		token: 'tok-ani',
		pin: '654321'
	},
	{
		participant_id: 'p-3',
		student_name: 'Cici',
		class_code: 'VII-A',
		room_code: 'R01',
		seat_no: 1,
		status: 'active'
	}
];

describe('exam card print helpers', () => {
	it('sorts printable cards by room and seat while keeping blank rooms last', () => {
		const sorted = filterAndSortExamCards(cards, { search: '', room: 'all', status: 'all' });
		expect(sorted.map((card) => `${card.room_code || 'blank'}:${card.seat_no || '-'}`)).toEqual([
			'R01:1',
			'R01:2',
			'R02:1',
			'blank:-'
		]);
	});

	it('filters by search, room, and readiness status', () => {
		expect(filterAndSortExamCards(cards, { search: 'ani', room: 'R01', status: 'ready' }).map((card) => card.student_name)).toEqual(['Ani']);
		expect(filterAndSortExamCards(cards, { search: '', room: 'R01', status: 'missing-secret' }).map((card) => card.student_name)).toEqual(['Cici']);
		expect(filterAndSortExamCards(cards, { search: 'vii-a', room: 'all', status: 'all' })).toHaveLength(3);
	});

	it('derives readiness counts for issue and print actions', () => {
		expect(deriveCardReadiness({ participantCount: 0, roomCount: 0, cardCount: 0 })).toMatchObject({ canIssue: false, canPrint: false, actionLabel: 'Lengkapi peserta' });
		expect(deriveCardReadiness({ participantCount: 10, roomCount: 0, cardCount: 0 })).toMatchObject({ canIssue: false, actionLabel: 'Simpan ruang dulu' });
		expect(deriveCardReadiness({ participantCount: 10, roomCount: 2, cardCount: 0 })).toMatchObject({ canIssue: true, canPrint: false, actionLabel: 'Terbitkan QR+PIN' });
		expect(deriveCardReadiness({ participantCount: 10, roomCount: 2, cardCount: 10 })).toMatchObject({ canIssue: false, canPrint: true, actionLabel: 'Cetak Kartu' });
	});

	it('summarizes room sheets without exposing participant PINs', () => {
		const summary = summarizeCardsByRoom(cards);
		expect(summary).toEqual([
			{ roomCode: 'R01', total: 2, ready: 1, missingSecret: 1, firstSeat: 1, lastSeat: 2 },
			{ roomCode: 'R02', total: 1, ready: 1, missingSecret: 0, firstSeat: 1, lastSeat: 1 },
			{ roomCode: 'Belum ruang', total: 1, ready: 0, missingSecret: 1, firstSeat: 0, lastSeat: 0 }
		]);
	});
});
