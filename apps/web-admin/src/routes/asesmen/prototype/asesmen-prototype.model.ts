export type PrototypeStatus = 'siap' | 'perlu-dicek' | 'menunggu';

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
};

export const prototypeLanes: PrototypeLane[] = [
	{
		id: 'a1-persiapan',
		code: 'A1',
		number: '1',
		title: 'Persiapan Ujian',
		audience: 'Panitia / Operator',
		description: 'Siapkan kegiatan, paket ujian, sesi, ruang, peserta, dan pengawas.',
		primaryAction: 'Mulai Persiapan',
		secondaryAction: 'Mode Lengkap',
		status: 'perlu-dicek',
		items: ['Kegiatan', 'Paket', 'Sesi', 'Ruang & Peserta', 'Pengawas']
	},
	{
		id: 'a2-dokumen',
		code: 'A2',
		number: '2',
		title: 'Dokumen & Cetak',
		audience: 'Panitia / Operator',
		description: 'Cetak kartu peserta, lembar pengawas ruang, dan berita acara dari satu tempat.',
		primaryAction: 'Buka Dokumen',
		status: 'menunggu',
		items: ['Kartu Peserta', 'Lembar Pengawas', 'Berita Acara', 'Arsip']
	},
	{
		id: 'a3-hari-h',
		code: 'A3',
		number: '3',
		title: 'Pelaksanaan Hari-H',
		audience: 'Panitia / Operator',
		description: 'Pantau sesi berjalan, kondisi ruang, kendala peserta, dan status pengumpulan.',
		primaryAction: 'Buka Hari-H',
		status: 'siap',
		items: ['Sesi Hari Ini', 'Pantau Ruang', 'Bantuan Peserta', 'Status Submit']
	},
	{
		id: 'a4-ruang-saya',
		code: 'A4',
		number: '4',
		title: 'Ruang Saya',
		audience: 'Pengawas / Guru',
		description: 'Tampilan sederhana untuk pengawas melihat ruang tugas dan peserta yang perlu dicek.',
		primaryAction: 'Buka Ruang Saya',
		secondaryAction: 'Portal Pengawas Ruang',
		status: 'siap',
		items: ['Ruang Tugas', 'Peserta', 'Peringatan', 'Hubungi Admin']
	},
	{
		id: 'a5-hasil',
		code: 'A5',
		number: '5',
		title: 'Hasil & Penutupan',
		audience: 'Panitia / Guru Mapel',
		description: 'Cek submit, koreksi uraian, rekap nilai, analisis, dan arsip akhir.',
		primaryAction: 'Buka Hasil',
		status: 'menunggu',
		items: ['Status Submit', 'Koreksi', 'Rekap Nilai', 'Analisis', 'Arsip']
	}
];

export const advancedPrototypeLinks = ['Kegiatan', 'Paket', 'Sesi', 'Perangkat Siswa', 'Non-Tes'];
