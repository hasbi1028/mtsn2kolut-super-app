export type SopStageKey =
	| 'draft'
	| 'question_authoring'
	| 'question_verification'
	| 'package_ready'
	| 'participants_rooms_ready'
	| 'tokens_cards_ready'
	| 'execution'
	| 'grading'
	| 'result_verification'
	| 'final_archive'
	| 'cancelled';

export type SopStageStatus = 'ready' | 'warning' | 'blocked' | 'running';

export type SopNextAction = {
	label: string;
	href: string;
};

export type SopStageDefinition = {
	key: SopStageKey;
	label: string;
	owner: string;
	description: string;
};

export type SopStageReadiness = {
	key: SopStageKey;
	label: string;
	owner?: string;
	description?: string;
	status: SopStageStatus;
	blocking_count: number;
	warning_count: number;
	next_actions: SopNextAction[];
};

export type SopReadinessResponse = {
	event_id: string;
	stages: SopStageReadiness[];
};

export const sopStages: SopStageDefinition[] = [
	{ key: 'draft', label: 'Draft', owner: 'Panitia', description: 'Identitas kegiatan dan tahun ajaran sudah dicatat.' },
	{ key: 'question_authoring', label: 'Penyusunan Soal', owner: 'Guru', description: 'Soal dilengkapi sesuai kebutuhan kegiatan.' },
	{ key: 'question_verification', label: 'Verifikasi Soal', owner: 'Reviewer', description: 'Soal ditelaah sampai siap terbit.' },
	{ key: 'package_ready', label: 'Paket Siap', owner: 'Operator', description: 'Paket soal aktif dan berisi soal.' },
	{ key: 'participants_rooms_ready', label: 'Peserta & Ruang Siap', owner: 'Operator', description: 'Peserta, ruang, kursi, dan pengawas sudah siap.' },
	{ key: 'tokens_cards_ready', label: 'Token & Kartu Siap', owner: 'Operator', description: 'Token ujian/ruang dan kartu ujian siap digunakan.' },
	{ key: 'execution', label: 'Pelaksanaan', owner: 'Pengawas', description: 'Ujian berjalan dan dipantau dari ruang.' },
	{ key: 'grading', label: 'Pengolahan Nilai', owner: 'Guru', description: 'Nilai otomatis/manual diselesaikan.' },
	{ key: 'result_verification', label: 'Verifikasi Hasil', owner: 'Panitia', description: 'Hasil diperiksa sebelum final.' },
	{ key: 'final_archive', label: 'Arsip Final', owner: 'Panitia', description: 'Kegiatan ditutup dan dokumen arsip disiapkan.' },
	{ key: 'cancelled', label: 'Dibatalkan', owner: 'Panitia', description: 'Kegiatan dibatalkan dan catatan pembatalan diarsipkan.' },
];

export const sopStatusLabels: Record<SopStageStatus, string> = {
	ready: 'Siap',
	warning: 'Perlu Perhatian',
	blocked: 'Belum Siap',
	running: 'Sedang Berjalan',
};
