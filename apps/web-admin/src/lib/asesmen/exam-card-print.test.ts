import { describe, expect, it } from 'vitest';

import { cardPin, cardReady, filterExamCards, sortExamCards, summarizeExamCards, type ExamCardLike } from './exam-card-print';

const cards: ExamCardLike[] = [
	{ participant_id: '1', student_name: 'Budi', nis: '002', class_code: 'VII A', room_name: 'Ruang 2', seat_no: 2, session_id: 's1', session_title: 'Sesi 1', token: 'abcd1234' },
	{ participant_id: '2', student_name: 'Ani', nis: '001', class_code: 'VII A', room_name: 'Ruang 1', seat_no: 1, session_id: 's1', session_title: 'Sesi 1', token: 'efgh5678', pin: '2468' },
	{ participant_id: '3', student_name: 'Cici', nis: '003', class_code: 'VIII B', room_name: '', seat_no: null, session_id: 's2', session_title: 'Sesi 2', token: '' }
];

describe('exam-card-print utilities', () => {
	it('derives PIN and readiness from card fields', () => {
		expect(cardPin(cards[0])).toBe('1234');
		expect(cardPin(cards[1])).toBe('2468');
		expect(cardReady(cards[0])).toBe(true);
		expect(cardReady(cards[2])).toBe(false);
	});

	it('filters cards by class, room, session, status, and search', () => {
		expect(filterExamCards(cards, { classCode: 'VII A' })).toHaveLength(2);
		expect(filterExamCards(cards, { roomName: 'Ruang 1' }).map((card) => card.participant_id)).toEqual(['2']);
		expect(filterExamCards(cards, { sessionId: 's2' }).map((card) => card.participant_id)).toEqual(['3']);
		expect(filterExamCards(cards, { status: 'ready' }).map((card) => card.participant_id)).toEqual(['1', '2']);
		expect(filterExamCards(cards, { status: 'needs_check' }).map((card) => card.participant_id)).toEqual(['3']);
		expect(filterExamCards(cards, { search: 'ani' }).map((card) => card.participant_id)).toEqual(['2']);
	});

	it('sorts by default operational distribution order', () => {
		expect(sortExamCards(cards).map((card) => card.participant_id)).toEqual(['2', '1', '3']);
		expect(sortExamCards(cards, 'class-name').map((card) => card.participant_id)).toEqual(['2', '1', '3']);
	});

	it('summarizes readiness counts', () => {
		expect(summarizeExamCards(cards)).toMatchObject({ total: 3, ready: 2, needsCheck: 1, missingToken: 1, missingRoom: 1, missingSeat: 1 });
	});
});
