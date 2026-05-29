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

export const excludedSurfaces = [
	'Bank Soal lama tidak ditiru di prototype ini',
	'Form import soal CSV tidak jadi alur utama operator',
	'Pengaturan sistem dan backup tetap di Mode Lengkap',
	'Data sensitif siswa/staf tidak dipakai sebagai data mock'
];
