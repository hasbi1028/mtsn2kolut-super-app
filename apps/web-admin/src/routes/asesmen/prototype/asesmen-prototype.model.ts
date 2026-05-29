export type PrototypeStatus = 'siap' | 'perlu-dicek' | 'menunggu';

export type PrototypeStepState = 'selesai' | 'lanjut' | 'cek' | 'opsional';

export type PrototypeChecklist = {
	label: string;
	state: PrototypeStepState;
};

export type PrototypeMetric = {
	label: string;
	value: string;
	note: string;
};

export type PrototypeRoom = {
	code: string;
	name: string;
	students: string;
	status: string;
	tone: 'green' | 'amber' | 'blue';
};

export type PrototypeLane = {
	id: 'a1-persiapan' | 'a2-dokumen' | 'a3-hari-h' | 'a4-ruang-saya' | 'a5-hasil';
	code: 'A1' | 'A2' | 'A3' | 'A4' | 'A5';
	number: string;
	title: string;
	audience: string;
	description: string;
	primaryAction: string;
	secondaryAction?: string;
	status: PrototypeStatus;
	items: string[];
	checklist: PrototypeChecklist[];
	operatorNote: string;
	modeLengkap?: string[];
};

export const prototypeMetrics: PrototypeMetric[] = [
	{
		label: 'Kegiatan aktif',
		value: 'UAS Genap 2025/2026',
		note: 'Contoh data statis untuk review alur'
	},
	{
		label: 'Ruang ujian',
		value: '8 ruang',
		note: 'Dibayangkan campur kelas/rombel otomatis'
	},
	{
		label: 'Mode utama',
		value: '5 langkah',
		note: 'Persiapan → Dokumen → Hari-H → Ruang → Hasil'
	}
];

export const prototypeRooms: PrototypeRoom[] = [
	{
		code: 'R.I',
		name: 'Ruang 1',
		students: '32 peserta',
		status: 'Siap cetak kartu',
		tone: 'green'
	},
	{
		code: 'R.II',
		name: 'Ruang 2',
		students: '31 peserta',
		status: 'Butuh cek pengawas',
		tone: 'amber'
	},
	{
		code: 'R.III',
		name: 'Ruang 3',
		students: '32 peserta',
		status: 'Menunggu jadwal sesi',
		tone: 'blue'
	}
];

export const prototypeLanes: PrototypeLane[] = [
	{
		id: 'a1-persiapan',
		code: 'A1',
		number: '1',
		title: 'Persiapan Ujian',
		audience: 'Panitia / Operator',
		description: 'Satu tempat untuk memastikan kegiatan, paket, sesi, ruang, peserta, dan pengawas sudah siap sebelum dicetak.',
		primaryAction: 'Mulai Persiapan',
		secondaryAction: 'Lihat kendala',
		status: 'perlu-dicek',
		items: ['Kegiatan', 'Paket', 'Sesi', 'Ruang & Peserta', 'Pengawas'],
		checklist: [
			{ label: 'Pilih kegiatan ujian aktif', state: 'selesai' },
			{ label: 'Cek paket dan sesi', state: 'lanjut' },
			{ label: 'Susun 8 ruang otomatis', state: 'cek' },
			{ label: 'Tetapkan pengawas ruang', state: 'cek' }
		],
		operatorNote: 'Operator cukup mengikuti daftar cek; pengaturan teknis tetap tersedia di Mode Lengkap bila perlu.',
		modeLengkap: ['Edit kegiatan', 'Builder paket', 'Detail sesi', 'Assignment ruang']
	},
	{
		id: 'a2-dokumen',
		code: 'A2',
		number: '2',
		title: 'Dokumen & Cetak',
		audience: 'Panitia / Operator',
		description: 'Pusat cetak yang menggabungkan kartu peserta, lembar pengawas ruang, berita acara, dan arsip tanpa membuka banyak halaman.',
		primaryAction: 'Buka Dokumen',
		secondaryAction: 'Cek kelengkapan',
		status: 'menunggu',
		items: ['Kartu Peserta', 'Lembar Pengawas', 'Berita Acara', 'Arsip'],
		checklist: [
			{ label: 'Cetak kartu peserta per ruang', state: 'lanjut' },
			{ label: 'Cetak lembar pengawas ruang', state: 'lanjut' },
			{ label: 'Siapkan berita acara', state: 'opsional' },
			{ label: 'Simpan arsip final', state: 'opsional' }
		],
		operatorNote: 'Cetak dibuat berbasis ruang agar panitia tidak mencari peserta satu per satu.',
		modeLengkap: ['Reset QR/PIN', 'Cetak per sesi', 'Arsip lama']
	},
	{
		id: 'a3-hari-h',
		code: 'A3',
		number: '3',
		title: 'Pelaksanaan Hari-H',
		audience: 'Panitia / Operator',
		description: 'Command center sederhana untuk melihat sesi berjalan, kondisi ruang, kendala peserta, dan status pengumpulan.',
		primaryAction: 'Buka Hari-H',
		secondaryAction: 'Hubungi pengawas',
		status: 'siap',
		items: ['Sesi Hari Ini', 'Pantau Ruang', 'Bantuan Peserta', 'Status Submit'],
		checklist: [
			{ label: 'Buka sesi sesuai jadwal', state: 'lanjut' },
			{ label: 'Lihat ruang bermasalah', state: 'cek' },
			{ label: 'Bantu peserta yang terkunci', state: 'opsional' },
			{ label: 'Pantau submit akhir', state: 'lanjut' }
		],
		operatorNote: 'Hari-H hanya menampilkan aksi yang sering dipakai; tindakan sensitif tidak ditaruh di layar awal.',
		modeLengkap: ['Monitoring detail', 'Log proctoring', 'Override admin']
	},
	{
		id: 'a4-ruang-saya',
		code: 'A4',
		number: '4',
		title: 'Ruang Saya',
		audience: 'Pengawas / Guru',
		description: 'Tampilan pengawas difokuskan ke ruang tugas: daftar peserta, peringatan, bantuan admin, dan berita acara singkat.',
		primaryAction: 'Buka Ruang Saya',
		secondaryAction: 'Portal Pengawas Ruang',
		status: 'siap',
		items: ['Ruang Tugas', 'Peserta', 'Peringatan', 'Hubungi Admin'],
		checklist: [
			{ label: 'Scan/masukkan kode ruang', state: 'lanjut' },
			{ label: 'Cek peserta hadir', state: 'lanjut' },
			{ label: 'Kirim bantuan admin bila perlu', state: 'opsional' },
			{ label: 'Isi catatan pengawas', state: 'opsional' }
		],
		operatorNote: 'Pengawas tidak perlu melihat menu panitia; cukup ruangnya sendiri dan tombol bantuan.',
		modeLengkap: ['Rekap ruang', 'Berita acara lengkap', 'Riwayat peringatan']
	},
	{
		id: 'a5-hasil',
		code: 'A5',
		number: '5',
		title: 'Hasil & Penutupan',
		audience: 'Panitia / Guru Mapel',
		description: 'Menutup ujian dengan cek submit, koreksi uraian, rekap nilai, analisis, dan arsip akhir.',
		primaryAction: 'Buka Hasil',
		secondaryAction: 'Cek yang belum submit',
		status: 'menunggu',
		items: ['Status Submit', 'Koreksi', 'Rekap Nilai', 'Analisis', 'Arsip'],
		checklist: [
			{ label: 'Pastikan semua submit', state: 'cek' },
			{ label: 'Koreksi uraian', state: 'lanjut' },
			{ label: 'Rekap nilai per mapel', state: 'lanjut' },
			{ label: 'Kunci arsip akhir', state: 'opsional' }
		],
		operatorNote: 'Hasil dipisah dari Hari-H agar panitia tidak mencampur monitoring dengan penutupan.',
		modeLengkap: ['Analisis butir', 'Ekspor nilai', 'Sinkron penilaian']
	}
];

export const advancedPrototypeLinks = ['Kegiatan', 'Paket', 'Sesi', 'Perangkat Siswa', 'Non-Tes', 'Proctoring detail', 'Log teknis'];

export const hiddenFromMainFlow = [
	'Pengaturan teknis jarang dipakai',
	'Override admin sensitif',
	'Log proctoring rinci',
	'Halaman lama yang masih dibutuhkan panitia'
];
