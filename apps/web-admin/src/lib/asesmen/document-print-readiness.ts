import { cardReadinessIssues, summarizeExamCards, type ExamCardLike } from './exam-card-print';

export type SupervisorCardLike = {
	id?: string;
	room_id?: string;
	roomId?: string;
	room_name?: string;
	roomName?: string;
	session_id?: string;
	sessionId?: string;
	token?: string;
	pin?: string;
	status?: string;
};

export type DocumentPrintOverview = {
	member_count?: number;
	target_count?: number;
	review_count?: number;
	question_count?: number;
	published_question_count?: number;
	package_count?: number;
	session_count?: number;
	room_count?: number;
	token_count?: number;
	card_count?: number;
	result_count?: number;
};

export type DocumentReadinessState = 'ready' | 'warning' | 'empty';

export type DocumentReadinessSummary = {
	state: DocumentReadinessState;
	label: string;
	count: number;
	blockers: string[];
	nextAction: string;
};

export function supervisorCardReady(card: SupervisorCardLike) {
	return Boolean((card.token || '').trim() && (card.pin || '').trim() && ((card.room_name || card.roomName || '').trim()));
}

export function summarizeSupervisorCards(cards: SupervisorCardLike[]): DocumentReadinessSummary {
	if (cards.length === 0) {
		return { state: 'empty', label: 'Belum ada lembar', count: 0, blockers: ['Belum ada lembar pengawas ruang yang termuat'], nextAction: 'Terbitkan QR+PIN Ruang' };
	}
	const missingToken = cards.filter((card) => !(card.token || '').trim()).length;
	const missingPin = cards.filter((card) => !(card.pin || '').trim()).length;
	const missingRoom = cards.filter((card) => !(card.room_name || card.roomName || '').trim()).length;
	const blockers: string[] = [];
	if (missingToken > 0) blockers.push(`${missingToken} lembar belum punya QR`);
	if (missingPin > 0) blockers.push(`${missingPin} lembar belum punya PIN`);
	if (missingRoom > 0) blockers.push(`${missingRoom} lembar belum punya ruang`);
	return {
		state: blockers.length > 0 ? 'warning' : 'ready',
		label: blockers.length > 0 ? 'Perlu cek' : 'Siap cetak',
		count: cards.length,
		blockers,
		nextAction: blockers.length > 0 ? 'Terbitkan QR+PIN Ruang' : 'Cetak Semua Lembar'
	};
}

export function summarizeParticipantCards(cards: ExamCardLike[]): DocumentReadinessSummary {
	if (cards.length === 0) {
		return { state: 'empty', label: 'Belum ada kartu', count: 0, blockers: ['Belum ada kartu peserta yang termuat'], nextAction: 'Terbitkan QR+PIN Peserta' };
	}
	const summary = summarizeExamCards(cards);
	const blockers = cardReadinessIssues(cards);
	return {
		state: blockers.length > 0 ? 'warning' : 'ready',
		label: blockers.length > 0 ? `${summary.ready}/${summary.total} siap` : 'Siap cetak',
		count: cards.length,
		blockers,
		nextAction: blockers.length > 0 ? 'Perbaiki data kartu' : 'Cetak Massal Kartu'
	};
}

export function summarizeDocumentPrintStatus(input: { examCards: ExamCardLike[]; supervisorCards: SupervisorCardLike[]; overview?: DocumentPrintOverview | null }) {
	const participantCards = summarizeParticipantCards(input.examCards);
	const supervisorCards = summarizeSupervisorCards(input.supervisorCards);
	const blockers = [...participantCards.blockers, ...supervisorCards.blockers];
	return {
		participantCards,
		supervisorCards,
		overview: input.overview ?? null,
		state: blockers.length > 0 ? 'warning' as const : participantCards.state === 'empty' && supervisorCards.state === 'empty' ? 'empty' as const : 'ready' as const,
		blockers,
		recommendedNextAction: participantCards.state !== 'ready' ? participantCards.nextAction : supervisorCards.state !== 'ready' ? supervisorCards.nextAction : 'Cetak dokumen kegiatan'
	};
}
