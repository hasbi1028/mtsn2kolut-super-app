export type ExamCardLike = {
	event_id?: string;
	event_title?: string;
	exam_type?: string;
	event_scope?: string;
	academic_year_name?: string;
	session_id?: string;
	session_title?: string;
	scheduled_start?: string;
	participant_id: string;
	token?: string;
	card_token?: string;
	exam_card_token?: string;
	qr_token?: string;
	portal_url?: string;
	pin?: string;
	card_pin?: string;
	exam_card_pin?: string;
	seat_no?: number | null;
	nis?: string;
	student_nama?: string;
	student_name?: string;
	gender?: string;
	class_code?: string;
	class_name?: string;
	room_name?: string;
	status?: string;
	card_id?: string;
	qr_path?: string;
	package_title?: string;
};

export type ExamCardFilters = {
	sessionId?: string;
	classCode?: string;
	roomName?: string;
	status?: 'all' | 'ready' | 'needs_check';
	search?: string;
};

export type ExamCardSortMode = 'room-seat-name' | 'class-name' | 'session-room-seat' | 'name';

export function cardPortalToken(card: ExamCardLike) {
	return card.card_token || card.exam_card_token || card.qr_token || card.token || '';
}

export function cardPin(card: ExamCardLike) {
	const explicit = card.pin || card.card_pin || card.exam_card_pin;
	if (explicit) return explicit;
	const token = cardPortalToken(card);
	return token.length >= 4 ? token.slice(-4).toUpperCase() : '';
}

export function cardReady(card: ExamCardLike) {
	return Boolean(cardPortalToken(card) && cardPin(card) && card.room_name && card.seat_no !== null && card.seat_no !== undefined);
}

export function studentName(card: ExamCardLike) {
	return card.student_nama || card.student_name || 'Peserta';
}

export function displayClass(card: ExamCardLike) {
	return card.class_code || card.class_name || 'Kelas belum ada';
}

export function maskedToken(token: string) {
	if (!token) return 'Belum ada';
	if (token.length <= 4) return '••••';
	return `${token.slice(0, 2)}••••${token.slice(-2)}`;
}

export function cardReadinessIssues(cards: ExamCardLike[]) {
	const missingToken = cards.filter((card) => !cardPortalToken(card)).length;
	const missingPin = cards.filter((card) => !cardPin(card)).length;
	const missingRoom = cards.filter((card) => !card.room_name).length;
	const missingSeat = cards.filter((card) => card.seat_no === null || card.seat_no === undefined).length;
	const issues: string[] = [];
	if (missingToken > 0) issues.push(`${missingToken} kartu belum punya QR/kode kartu`);
	if (missingPin > 0) issues.push(`${missingPin} kartu belum punya PIN`);
	if (missingRoom > 0) issues.push(`${missingRoom} peserta belum punya ruangan`);
	if (missingSeat > 0) issues.push(`${missingSeat} peserta belum punya nomor meja`);
	return issues;
}

export function summarizeExamCards(cards: ExamCardLike[]) {
	const total = cards.length;
	const ready = cards.filter(cardReady).length;
	const missingToken = cards.filter((card) => !cardPortalToken(card)).length;
	const missingPin = cards.filter((card) => !cardPin(card)).length;
	const missingRoom = cards.filter((card) => !card.room_name).length;
	const missingSeat = cards.filter((card) => card.seat_no === null || card.seat_no === undefined).length;
	return { total, ready, needsCheck: total - ready, missingToken, missingPin, missingRoom, missingSeat };
}

function normalized(value: string | undefined | null) {
	return (value ?? '').toLocaleLowerCase('id-ID').trim();
}

function compareText(a: string | undefined | null, b: string | undefined | null) {
	return normalized(a).localeCompare(normalized(b), 'id-ID', { numeric: true, sensitivity: 'base' });
}

function compareTextBlankLast(a: string | undefined | null, b: string | undefined | null) {
	const aa = normalized(a);
	const bb = normalized(b);
	if (!aa && bb) return 1;
	if (aa && !bb) return -1;
	return aa.localeCompare(bb, 'id-ID', { numeric: true, sensitivity: 'base' });
}

export function sortExamCards<T extends ExamCardLike>(cards: T[], mode: ExamCardSortMode = 'room-seat-name'): T[] {
	return [...cards].sort((a, b) => {
		if (mode === 'class-name') {
			return compareText(displayClass(a), displayClass(b)) || compareText(studentName(a), studentName(b)) || compareText(a.nis, b.nis);
		}
		if (mode === 'session-room-seat') {
			return compareText(a.session_title, b.session_title) || compareTextBlankLast(a.room_name, b.room_name) || Number(a.seat_no ?? 9999) - Number(b.seat_no ?? 9999) || compareText(studentName(a), studentName(b));
		}
		if (mode === 'name') {
			return compareText(studentName(a), studentName(b)) || compareText(a.nis, b.nis);
		}
		return compareTextBlankLast(a.room_name, b.room_name) || Number(a.seat_no ?? 9999) - Number(b.seat_no ?? 9999) || compareText(studentName(a), studentName(b));
	});
}

export function filterExamCards<T extends ExamCardLike>(cards: T[], filters: ExamCardFilters = {}): T[] {
	const search = normalized(filters.search);
	return cards.filter((card) => {
		if (filters.sessionId && card.session_id !== filters.sessionId) return false;
		if (filters.classCode && displayClass(card) !== filters.classCode) return false;
		if (filters.roomName && (card.room_name || '') !== filters.roomName) return false;
		if (filters.status === 'ready' && !cardReady(card)) return false;
		if (filters.status === 'needs_check' && cardReady(card)) return false;
		if (search) {
			const haystack = [studentName(card), card.nis, displayClass(card), card.room_name, card.session_title].map(normalized).join(' ');
			if (!haystack.includes(search)) return false;
		}
		return true;
	});
}

export function uniqueSorted(values: Array<string | undefined | null>) {
	return [...new Set(values.map((value) => (value ?? '').trim()).filter(Boolean))].sort((a, b) => compareText(a, b));
}
