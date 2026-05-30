export type AssessmentPrintReadinessInput = {
	room_count?: number | null;
	participant_count?: number | null;
	card_count?: number | null;
};

export type AssessmentPrintDocument = {
	id: 'participant-cards' | 'supervisor-sheets' | 'archive-checklist';
	title: string;
	description: string;
	status: string;
	detail: string;
	countLabel: string;
	ready: boolean;
	primaryAction: string;
	dangerous?: boolean;
};

function count(value: number | null | undefined) {
	return Math.max(0, Number(value ?? 0));
}

export function summarizeAssessmentPrintDocuments(input: AssessmentPrintReadinessInput): AssessmentPrintDocument[] {
	const rooms = count(input.room_count);
	const participants = count(input.participant_count);
	const cards = count(input.card_count);
	const hasRooms = rooms > 0;
	const hasParticipants = participants > 0;
	const hasCards = cards > 0;

	let participantStatus = 'Belum siap';
	let participantDetail = 'Atur ruang dan peserta dulu sebelum kartu diterbitkan.';
	let participantAction = 'Lengkapi Ruang';
	let participantReady = false;
	let participantDangerous = false;

	if (hasRooms && !hasParticipants) {
		participantDetail = 'Ruang sudah ada, tetapi belum ada peserta tersimpan untuk dicetak.';
		participantAction = 'Lengkapi Peserta';
	} else if (hasParticipants && !hasCards) {
		participantStatus = 'Siap diterbitkan';
		participantDetail = 'Peserta sudah masuk ruang. QR+PIN hanya dibuat setelah tombol terbitkan ditekan.';
		participantAction = 'Terbitkan QR+PIN';
		participantReady = true;
		participantDangerous = true;
	} else if (hasCards) {
		participantStatus = 'Siap cetak';
		participantDetail = 'Kartu peserta sudah tersedia untuk dicetak.';
		participantAction = 'Cetak Kartu';
		participantReady = true;
	}

	const supervisorReady = hasRooms;

	return [
		{
			id: 'participant-cards',
			title: 'Kartu Peserta',
			description: 'Kartu masuk siswa berisi identitas, ruang, kursi, QR dan PIN.',
			status: participantStatus,
			detail: participantDetail,
			countLabel: hasCards ? `${cards} kartu` : `${participants} peserta`,
			ready: participantReady,
			primaryAction: participantAction,
			dangerous: participantDangerous
		},
		{
			id: 'supervisor-sheets',
			title: 'Lembar Pengawas Ruang',
			description: 'Lembar per ruang untuk pengawas atau guru pengganti saat hari ujian.',
			status: supervisorReady ? 'Siap dicetak' : 'Belum siap',
			detail: supervisorReady ? 'Ruang ujian sudah tersedia. Lembar pengawas tidak membuat PIN peserta.' : 'Simpan ruang ujian dulu agar lembar pengawas punya daftar ruang.',
			countLabel: `${rooms} ruang`,
			ready: supervisorReady,
			primaryAction: supervisorReady ? 'Cetak Lembar' : 'Atur Ruang'
		},
		{
			id: 'archive-checklist',
			title: 'Checklist Arsip',
			description: 'Daftar ringkas dokumen yang perlu dibawa panitia setelah ujian.',
			status: hasRooms || hasParticipants ? 'Siap dibaca' : 'Template awal',
			detail: 'Checklist aman dibuka kapan saja karena tidak menerbitkan token atau PIN.',
			countLabel: `${rooms} ruang · ${participants} peserta`,
			ready: true,
			primaryAction: 'Unduh Checklist'
		}
	];
}
