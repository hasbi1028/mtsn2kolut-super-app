export type StatusTone = 'blue' | 'green' | 'amber' | 'red' | 'slate';

export type CommandMetric = {
	label: string;
	value: string;
	note: string;
	tone: StatusTone;
};

export type WorkArea = {
	code: string;
	label: string;
	description: string;
	status: string;
	cta: string;
	items: string[];
};

export type LegacyMenuGroup = {
	label: string;
	items: string[];
};

export type LegacyTableRow = {
	name: string;
	meta: string;
	status: string;
	actions: string[];
	tone: StatusTone;
};

export type PortalPreview = {
	label: string;
	route: string;
	title: string;
	description: string;
	steps: string[];
	primaryActions: string[];
	guardrails: string[];
};

export type ReadinessCard = {
	label: string;
	percent: number;
	status: string;
	detail: string;
	tone: StatusTone;
};

export type TodayAgenda = {
	time: string;
	title: string;
	detail: string;
	status: string;
	tone: StatusTone;
};

export type RoomBlueprint = {
	room: string;
	proctor: string;
	capacity: string;
	students: string;
	status: string;
	tone: StatusTone;
};

export type CardPrintStep = {
	label: string;
	detail: string;
	badge: string;
};

export type MobileScreen = {
	label: string;
	mode: 'student' | 'proctor';
	title: string;
	sections: string[];
	bottomActions: string[];
};

export type ResultSummary = {
	label: string;
	value: string;
	detail: string;
	tone: StatusTone;
};

export const sourceNotes = [
	'Rujukan: domain cbt.mtsn2kolut.sch.id dan repo lokal /home/servermtsn2kolut/cbt-ujian',
	'Bank Soal sengaja tidak dijadikan fokus prototype ini',
	'Prototype tetap statis/frontend-only: tidak login, tidak fetch, tidak menyimpan data'
];

export const commandMetrics: CommandMetric[] = [
	{ label: 'Siswa terdaftar', value: '42', note: 'contoh angka dari pola dashboard CBT lama', tone: 'blue' },
	{ label: 'Ruang ujian', value: '4', note: 'ruang, kapasitas, monitor, denah', tone: 'green' },
	{ label: 'Sesi siap', value: '23', note: 'jadwal sesi + paket aktif', tone: 'amber' },
	{ label: 'Perlu tindak', value: '3', note: 'kekurangan operasional terkelompok', tone: 'red' }
];

export const readinessCards: ReadinessCard[] = [
	{ label: 'Kesiapan Peserta', percent: 92, status: 'Hampir siap', detail: '3 peserta belum punya kartu QR+PIN cetak', tone: 'amber' },
	{ label: 'Kesiapan Ruang', percent: 100, status: 'Siap', detail: '4 ruang punya lembar pengawas ruang', tone: 'green' },
	{ label: 'Kesiapan Sesi', percent: 84, status: 'Perlu final', detail: '1 jadwal sesi masih draft operator', tone: 'amber' },
	{ label: 'Kesiapan Hasil', percent: 68, status: 'Menunggu ujian', detail: 'rekap nilai muncul setelah submit', tone: 'blue' }
];

export const todayAgenda: TodayAgenda[] = [
	{ time: '07.15', title: 'Briefing pengawas', detail: 'Bagikan lembar pengawas ruang + PIN tertutup', status: 'siap', tone: 'green' },
	{ time: '07.45', title: 'Sesi Informatika kelas IX', detail: 'Ruang 01–04, portal ujian web-mobile', status: 'aktif nanti', tone: 'blue' },
	{ time: '09.30', title: 'Cek peserta belum submit', detail: 'Pengawas pakai tab Peringatan dan Peserta', status: 'hari-H', tone: 'amber' },
	{ time: '10.15', title: 'Export rekap nilai', detail: 'Nilai PG + daftar essay belum dikoreksi', status: 'penutupan', tone: 'slate' }
];

export const legacyMenuGroups: LegacyMenuGroup[] = [
	{ label: 'Pusat Data', items: ['Dashboard', 'Mata Pelajaran', 'Data Kelas', 'Daftar Siswa', 'Staf & Pengawas', 'Ruang Ujian'] },
	{ label: 'Manajemen Ujian', items: ['Paket Ujian', 'Jadwal Sesi', 'Kartu QR', 'Penempatan'] },
	{ label: 'Pelaksanaan', items: ['Portal Siswa', 'Portal Pengawas', 'Proctoring Live', 'Anti-cheat'] },
	{ label: 'Laporan', items: ['Rekap Nilai', 'Koreksi Essay', 'Export Excel', 'Arsip'] }
];

export const workAreas: WorkArea[] = [
	{
		code: '01',
		label: 'Dashboard Utama',
		description: 'Command center terang seperti CBT lama: angka besar, readiness, agenda hari ini, dan kekurangan operasional.',
		status: 'Arah utama',
		cta: 'Buka command center',
		items: ['Kesiapan peserta', 'Kesiapan ruang', 'Kesiapan sesi', 'Kesiapan paket']
	},
	{
		code: '02',
		label: 'Ruang & Kartu QR',
		description: 'Mengikuti repo cbt-ujian: siswa scan kartu QR, pengawas scan lembar pengawasan, panitia cetak dari satu tempat.',
		status: 'Dibuat familiar',
		cta: 'Cek kartu & ruang',
		items: ['Cetak kartu peserta', 'Lembar pengawas', 'Denah ruang', 'Token sesi']
	},
	{
		code: '03',
		label: 'Jadwal Sesi',
		description: 'Sesi ujian ditampilkan sebagai daftar operasional dengan status aktif/draft/final dan filter sederhana.',
		status: 'Mirip CBT lama',
		cta: 'Kelola jadwal',
		items: ['Tanggal sesi', 'Mata pelajaran', 'Target kelas', 'Status sesi']
	},
	{
		code: '04',
		label: 'Proctoring Live',
		description: 'Pengawas melihat peserta, progress jawaban, status submit, dan log anti-cheat terbaru tanpa masuk ke menu rumit.',
		status: 'Hari-H',
		cta: 'Pantau ruang',
		items: ['Monitoring peserta', 'Log anti-cheat', 'Terkunci', 'Sudah submit']
	},
	{
		code: '05',
		label: 'Rekap Nilai',
		description: 'Rekap dibuat seperti CBT lama: filter kelas/mapel, ringkasan peserta tampil, lengkap/perlu tindak, dan export.',
		status: 'Penutupan',
		cta: 'Lihat rekap',
		items: ['Rata-rata', 'Progress', 'Koreksi essay', 'Export CSV/Excel']
	}
];

export const tableRows: LegacyTableRow[] = [
	{ name: 'Ruang 01', meta: '15 siswa · Gedung Kelas 9A', status: 'SIAP', actions: ['DENAH', 'MONITOR'], tone: 'green' },
	{ name: 'Jadwal Sesi UAS', meta: 'Informatika · Tingkat 9', status: 'AKTIF', actions: ['EDIT', 'FINAL'], tone: 'blue' },
	{ name: 'Peserta belum submit', meta: '3 siswa perlu dicek pengawas', status: 'PERLU TINDAK', actions: ['PANTAU', 'HUBUNGI'], tone: 'amber' },
	{ name: 'Koreksi Essay', meta: 'Jawaban uraian menunggu guru mapel', status: 'MENUNGGU', actions: ['KOREKSI', 'EXPORT'], tone: 'slate' }
];

export const roomBlueprints: RoomBlueprint[] = [
	{ room: 'Ruang 01', proctor: 'Pengawas Ruang A', capacity: '15 kursi', students: '14 hadir / 1 belum', status: 'Siap mulai', tone: 'green' },
	{ room: 'Ruang 02', proctor: 'Pengawas Ruang B', capacity: '15 kursi', students: '15 hadir', status: 'Sedang ujian', tone: 'blue' },
	{ room: 'Ruang 03', proctor: 'Pengawas Ruang C', capacity: '14 kursi', students: '13 hadir / 1 izin', status: 'Perlu cek', tone: 'amber' },
	{ room: 'Ruang 04', proctor: 'Pengawas pengganti', capacity: '12 kursi', students: '12 hadir', status: 'Lembar siap', tone: 'slate' }
];

export const cardPrintSteps: CardPrintStep[] = [
	{ label: 'Kartu Peserta Ujian', detail: 'QR menuju Portal Ujian Peserta, PIN disimpan/ditutup saat cetak.', badge: 'peserta' },
	{ label: 'Lembar Pengawas Ruang', detail: 'Satu QR+PIN per ruang/sesi agar pengawas pengganti tetap bisa bertugas.', badge: 'ruang' },
	{ label: 'Daftar hadir & denah', detail: 'Operator mencetak daftar peserta, denah kursi, dan catatan serah terima.', badge: 'dokumen' }
];

export const mobileScreens: MobileScreen[] = [
	{
		label: 'Portal Ujian Peserta',
		mode: 'student',
		title: 'Scan QR → PIN → konfirmasi identitas → ujian',
		sections: ['Identitas siswa', 'Status sesi', 'Soal satu layar', 'Timer besar', 'Kirim final'],
		bottomActions: ['Sebelumnya', 'Nomor Soal', 'Berikutnya']
	},
	{
		label: 'Portal Pengawasan Ruang',
		mode: 'proctor',
		title: 'Ruang · Peringatan · Peserta',
		sections: ['Mulai Ujian', 'Peringatan anti-cheat ramah', 'Daftar peserta', 'Hubungi Admin'],
		bottomActions: ['Ruang', 'Peringatan', 'Peserta']
	}
];

export const portalPreviews: PortalPreview[] = [
	{
		label: 'Portal depan',
		route: '/',
		title: 'Masuk ujian cukup scan kartu QR',
		description: 'Satu panel masuk untuk siswa dan pengawas seperti repo cbt-ujian.',
		steps: ['Scan QR', 'Kenali jenis token', 'Masuk portal sesuai peran'],
		primaryActions: ['Masuk portal', 'Cetak kartu & lembar QR'],
		guardrails: ['Token QR tidak ditampilkan mentah', 'Bahasa sederhana', 'Tidak ada sidebar admin']
	},
	{
		label: 'Portal siswa',
		route: '/ujian/[token]',
		title: 'Token sesi, timer, soal, autosave',
		description: 'Siswa fokus mengerjakan; PG dan essay didukung; anti-cheat diberi bahasa manusiawi.',
		steps: ['Masukkan token sesi', 'Mulai ujian', 'Jawab soal', 'Kirim final'],
		primaryActions: ['Mulai ujian', 'Kirim jawaban final'],
		guardrails: ['Timer jelas', 'Autosave', 'Peringatan anti-cheat ramah']
	},
	{
		label: 'Portal pengawas',
		route: '/pengawas/[token]',
		title: 'Monitoring peserta per ruang',
		description: 'Pengawas melihat submit, jawaban, status terkunci, dan log anti-cheat terbaru.',
		steps: ['Scan lembar pengawas', 'Lihat peserta', 'Cek log terbaru', 'Hubungi admin bila perlu'],
		primaryActions: ['Kartu QR', 'Admin', 'Pantau peserta'],
		guardrails: ['Tanpa unlock publik', 'Aksi sensitif tetap admin', 'Status peserta ringkas']
	}
];

export const resultSummaries: ResultSummary[] = [
	{ label: 'Submit selesai', value: '39/42', detail: '3 peserta perlu dicek pengawas', tone: 'amber' },
	{ label: 'Essay menunggu koreksi', value: '12', detail: 'dikirim ke guru mapel', tone: 'blue' },
	{ label: 'Export siap', value: 'PG', detail: 'rekap PG bisa diunduh setelah sesi tutup', tone: 'green' },
	{ label: 'Sinkron nilai', value: 'nanti', detail: 'masuk tahap produksi setelah blueprint disetujui', tone: 'slate' }
];

export const excludedSurfaces = [
	'Bank Soal lama tidak ditiru di prototype ini',
	'Form import soal CSV tidak jadi alur utama operator',
	'Pengaturan sistem dan backup tetap di Mode Lengkap',
	'Data sensitif siswa/staf tidak dipakai sebagai data mock'
];
