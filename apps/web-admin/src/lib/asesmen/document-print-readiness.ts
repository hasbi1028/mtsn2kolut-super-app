export type DocumentStatusState = 'blocked' | 'action' | 'ready' | 'warning';

export type AssessmentDocumentCounts = {
	participantCount: number;
	roomCount: number;
	cardCount: number;
};

export type DocumentPrintStatus = {
	state: DocumentStatusState;
	label: string;
	description: string;
};

export type DocumentPrintSummary = {
	participantCards: DocumentPrintStatus;
	supervisorSheets: DocumentPrintStatus;
	archiveChecklist: DocumentPrintStatus;
};

export function summarizeDocumentPrintStatus(counts: AssessmentDocumentCounts): DocumentPrintSummary {
	const participantCount = Number(counts.participantCount || 0);
	const roomCount = Number(counts.roomCount || 0);
	const cardCount = Number(counts.cardCount || 0);

	let participantCards: DocumentPrintStatus;
	if (participantCount <= 0) {
		participantCards = {
			state: 'blocked',
			label: 'Lengkapi Peserta',
			description: 'Pilih rombel dan simpan penempatan sebelum kartu peserta diterbitkan.'
		};
	} else if (roomCount <= 0) {
		participantCards = {
			state: 'blocked',
			label: 'Simpan ruang dulu',
			description: 'Kartu peserta menunggu ruang dan nomor kursi final.'
		};
	} else if (cardCount <= 0) {
		participantCards = {
			state: 'action',
			label: 'Buat QR+PIN',
			description: `${participantCount} peserta siap dibuatkan QR dan PIN.`
		};
	} else {
		participantCards = {
			state: 'ready',
			label: 'Cetak Kartu',
			description: `${cardCount}/${participantCount} kartu sudah tersedia. Jika PIN tidak muncul, buat PIN baru lalu cetak.`
		};
	}

	const supervisorSheets: DocumentPrintStatus = roomCount > 0
		? {
			state: 'ready',
			label: 'Siap dicetak',
			description: `${roomCount} ruang siap dibuat lembar pengawas berisi daftar peserta dan kursi.`
		}
		: {
			state: 'blocked',
			label: 'Simpan ruang dulu',
			description: 'Lembar pengawas dibuat setelah ruang tersedia.'
		};

	const archiveChecklist: DocumentPrintStatus = participantCards.state === 'ready' && supervisorSheets.state === 'ready'
		? {
			state: 'ready',
			label: 'Siap dilengkapi',
			description: 'Checklist arsip bisa disiapkan bersama kartu peserta dan lembar pengawas.'
		}
		: {
			state: 'warning',
			label: 'Menunggu dokumen utama',
			description: 'Lengkapi kartu peserta dan lembar pengawas sebelum finalisasi arsip.'
		};

	return { participantCards, supervisorSheets, archiveChecklist };
}
