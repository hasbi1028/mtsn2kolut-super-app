export type ExamCardPrintItem = {
	card_id?: string;
	participant_id: string;
	session_id?: string;
	room_id?: string;
	student_id?: string;
	student_name: string;
	nis?: string;
	nisn?: string;
	class_code?: string;
	class_name?: string;
	grade_level?: number;
	room_code?: string;
	room_name?: string;
	seat_no?: number;
	status?: string;
	failed_attempts?: number;
	card_created_at?: string;
	card_expires_at?: string;
	token?: string;
	pin?: string;
	qr_path?: string;
};

export type ExamCardFilter = {
	search: string;
	room: string;
	status: 'all' | 'ready' | 'missing-secret' | 'not-issued';
};

export type CardReadinessInput = {
	participantCount: number;
	roomCount: number;
	cardCount: number;
};

export type CardReadiness = {
	canIssue: boolean;
	canPrint: boolean;
	actionLabel: string;
	message: string;
};

export type RoomSheetSummary = {
	roomCode: string;
	total: number;
	ready: number;
	missingSecret: number;
	firstSeat: number;
	lastSeat: number;
};

export function cardHasPrintableSecret(card: ExamCardPrintItem): boolean {
	return Boolean((card.token ?? '').trim() && (card.pin ?? '').trim());
}

export function deriveCardReadiness(input: CardReadinessInput): CardReadiness {
	const participantCount = Math.max(0, input.participantCount || 0);
	const roomCount = Math.max(0, input.roomCount || 0);
	const cardCount = Math.max(0, input.cardCount || 0);
	if (participantCount === 0) {
		return {
			canIssue: false,
			canPrint: false,
			actionLabel: 'Lengkapi peserta',
			message: 'Belum ada peserta. Lengkapi peserta dan ruang sebelum menerbitkan QR+PIN.'
		};
	}
	if (roomCount === 0) {
		return {
			canIssue: false,
			canPrint: false,
			actionLabel: 'Simpan ruang dulu',
			message: 'Peserta sudah ada, tetapi ruang/kursi belum disimpan.'
		};
	}
	if (cardCount === 0) {
		return {
			canIssue: true,
			canPrint: false,
			actionLabel: 'Terbitkan QR+PIN',
			message: 'Ruang dan peserta siap. Terbitkan QR+PIN hanya saat siap cetak.'
		};
	}
	return {
		canIssue: false,
		canPrint: true,
		actionLabel: 'Cetak Kartu',
		message: 'Kartu sudah pernah diterbitkan. Cetak dari daftar, atau regenerasi dengan sengaja bila perlu.'
	};
}

export function filterAndSortExamCards(cards: ExamCardPrintItem[], filter: ExamCardFilter): ExamCardPrintItem[] {
	const search = normalize(filter.search);
	return [...cards]
		.filter((card) => {
			if (filter.room !== 'all' && roomLabel(card) !== filter.room) return false;
			if (filter.status === 'ready' && !cardHasPrintableSecret(card)) return false;
			if (filter.status === 'missing-secret' && cardHasPrintableSecret(card)) return false;
			if (filter.status === 'not-issued' && normalize(card.status) !== 'not_issued') return false;
			if (!search) return true;
			return normalize([
				card.student_name,
				card.nis,
				card.nisn,
				card.class_code,
				card.class_name,
				card.room_code,
				String(card.seat_no || '')
			].join(' ')).includes(search);
		})
		.sort(compareExamCards);
}

export function summarizeCardsByRoom(cards: ExamCardPrintItem[]): RoomSheetSummary[] {
	const sorted = filterAndSortExamCards(cards, { search: '', room: 'all', status: 'all' });
	const byRoom = new Map<string, RoomSheetSummary>();
	for (const card of sorted) {
		const key = roomLabel(card);
		const seat = card.seat_no || 0;
		const current = byRoom.get(key) ?? { roomCode: key, total: 0, ready: 0, missingSecret: 0, firstSeat: seat, lastSeat: seat };
		current.total += 1;
		if (cardHasPrintableSecret(card)) current.ready += 1;
		else current.missingSecret += 1;
		if (seat > 0) {
			current.firstSeat = current.firstSeat > 0 ? Math.min(current.firstSeat, seat) : seat;
			current.lastSeat = Math.max(current.lastSeat, seat);
		}
		byRoom.set(key, current);
	}
	return Array.from(byRoom.values()).sort(compareRoomSummary);
}

export function availableRooms(cards: ExamCardPrintItem[]): string[] {
	return Array.from(new Set(cards.map(roomLabel))).sort(compareRoomCode);
}

function compareExamCards(a: ExamCardPrintItem, b: ExamCardPrintItem): number {
	const roomCompare = compareRoomCode(roomLabel(a), roomLabel(b));
	if (roomCompare !== 0) return roomCompare;
	const seatCompare = (a.seat_no || Number.MAX_SAFE_INTEGER) - (b.seat_no || Number.MAX_SAFE_INTEGER);
	if (seatCompare !== 0) return seatCompare;
	return (a.student_name || '').localeCompare(b.student_name || '', 'id');
}

function compareRoomSummary(a: RoomSheetSummary, b: RoomSheetSummary): number {
	return compareRoomCode(a.roomCode, b.roomCode);
}

function compareRoomCode(a: string, b: string): number {
	if (a === 'Belum ruang' && b !== 'Belum ruang') return 1;
	if (b === 'Belum ruang' && a !== 'Belum ruang') return -1;
	return a.localeCompare(b, 'id', { numeric: true });
}

function roomLabel(card: ExamCardPrintItem): string {
	return (card.room_code ?? '').trim() || 'Belum ruang';
}

function normalize(value: unknown): string {
	return String(value ?? '').trim().toLowerCase();
}
