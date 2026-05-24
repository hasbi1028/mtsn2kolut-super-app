export type Supervisor = {
	code: number;
	name: string;
};

export type RoomCode = 'R.I' | 'R.II' | 'R.III' | 'R.IV' | 'R.V' | 'R.VI' | 'R.VII' | 'R.VIII';

export type ExamSlot = {
	date: string;
	day: string;
	start: string;
	end: string;
	subject: string;
	supervisors: Record<RoomCode, number>;
	status?: 'ready' | 'needs_review';
	notes?: string[];
};

export const roomCodes = ['R.I', 'R.II', 'R.III', 'R.IV', 'R.V', 'R.VI', 'R.VII', 'R.VIII'] as const satisfies readonly RoomCode[];

export const supervisors: Supervisor[] = [
	{ code: 1, name: 'Taufik S.Ag M.Pd' },
	{ code: 2, name: 'Kardi S.Pd' },
	{ code: 3, name: 'Irmawati Nur S.Ag M.Pd' },
	{ code: 4, name: 'Supriadi S.Pd' },
	{ code: 5, name: 'Dra Ratnawati' },
	{ code: 6, name: 'Drs H Chaeruddin' },
	{ code: 7, name: 'Andi Rasnia S.Ag' },
	{ code: 8, name: 'Sitti Rafidah S.Pd.I' },
	{ code: 9, name: 'Herniati S.Pd' },
	{ code: 10, name: 'Muh Saing S.E' },
	{ code: 11, name: 'Sutra S.Ag' },
	{ code: 12, name: 'Saimang S.Ag' },
	{ code: 13, name: 'Mida S.Ag' },
	{ code: 14, name: 'Rusnawati S.Pd' },
	{ code: 15, name: 'Wati Zaelani Sidi Ellyanando S.Pd' },
	{ code: 16, name: 'Susianti S.Pd' },
	{ code: 17, name: 'Rustiani S.Pd' },
	{ code: 18, name: 'Nirmawati Tahir S.Pd' },
	{ code: 19, name: 'Nurul Fitri Usman S.Pd' },
	{ code: 20, name: 'Kaharuddin S.Pd' },
	{ code: 21, name: 'Abdillah S.Pd' },
	{ code: 22, name: 'Yurnianti S.Pd' },
	{ code: 23, name: 'Mutmainnah S.Pd' },
	{ code: 24, name: 'Andi Mirna S.Pd' },
	{ code: 25, name: 'Nur Ilmi S.Pd' },
	{ code: 26, name: 'Mirnawati S.Pd' },
	{ code: 27, name: 'Nurunnisa Alimah A S.Pd' }
];

const slot = (date: string, day: string, start: string, end: string, subject: string, codes: number[], notes: string[] = []): ExamSlot => ({
	date,
	day,
	start,
	end,
	subject,
	supervisors: Object.fromEntries(roomCodes.map((room, index) => [room, codes[index]])) as Record<RoomCode, number>,
	status: notes.length ? 'needs_review' : 'ready',
	notes
});

export const examSlots: ExamSlot[] = [
	slot('2026-06-04', 'Kamis', '07:30', '09:00', 'Matematika', [1, 2, 11, 20, 5, 6, 7, 8]),
	slot('2026-06-04', 'Kamis', '09:30', '11:00', 'SKI', [9, 10, 3, 12, 13, 14, 15, 16]),
	slot('2026-06-05', 'Jumat', '07:30', '09:00', 'IPS', [17, 18, 19, 20, 21, 27, 5, 24]),
	slot('2026-06-05', 'Jumat', '09:30', '11:00', "Al-Qur'an Hadis", [25, 26, 23, 27, 3, 22, 23, 6], ['Nomor pengawas 23 tertulis dobel pada R.III dan R.VII; sesuai arahan panitia dibiarkan sebagai pending koreksi.']),
	slot('2026-06-06', 'Sabtu', '07:30', '09:00', 'IPA', [7, 8, 17, 10, 11, 12, 13, 14]),
	slot('2026-06-06', 'Sabtu', '09:00', '10:30', 'Informatika', [15, 16, 17, 18, 19, 20, 21, 22]),
	slot('2026-06-06', 'Sabtu', '10:40', '12:00', 'Akidah Akhlak', [23, 24, 25, 26, 1, 2, 19, 4]),
	slot('2026-06-08', 'Senin', '07:30', '09:00', 'Bahasa Indonesia', [21, 22, 25, 8, 9, 10, 11, 20]),
	slot('2026-06-08', 'Senin', '09:00', '10:30', 'Penjas', [13, 14, 15, 16, 17, 18, 19, 12]),
	slot('2026-06-08', 'Senin', '10:40', '12:00', 'Fiqih', [21, 22, 23, 24, 25, 26, 1, 2]),
	slot('2026-06-09', 'Selasa', '07:30', '09:00', 'Bahasa Inggris', [3, 4, 5, 6, 7, 8, 9, 10]),
	slot('2026-06-09', 'Selasa', '09:00', '10:30', 'Kewirausahaan', [24, 12, 13, 14, 15, 16, 25, 18]),
	slot('2026-06-09', 'Selasa', '10:40', '12:00', 'Seni Budaya', [19, 20, 21, 22, 23, 24, 17, 26]),
	slot('2026-06-10', 'Rabu', '07:30', '09:00', 'Bahasa Arab', [1, 10, 3, 27, 5, 6, 7, 8]),
	slot('2026-06-10', 'Rabu', '09:30', '11:00', 'PKN', [9, 27, 11, 12, 13, 14, 15, 16])
];

export const supervisorByCode = new Map(supervisors.map((supervisor) => [supervisor.code, supervisor]));

export function getSupervisorName(code: number): string {
	return supervisorByCode.get(code)?.name ?? `Pengawas ${code}`;
}

export function getFlatAssignments() {
	return examSlots.flatMap((item) =>
		roomCodes.map((room) => ({
			date: item.date,
			day: item.day,
			start: item.start,
			end: item.end,
			subject: item.subject,
			room,
			code: item.supervisors[room],
			name: getSupervisorName(item.supervisors[room]),
			needsReview: item.status === 'needs_review'
		}))
	);
}

export function getSupervisorLoad() {
	const counts = new Map<number, number>();
	for (const assignment of getFlatAssignments()) {
		counts.set(assignment.code, (counts.get(assignment.code) ?? 0) + 1);
	}
	return supervisors.map((supervisor) => ({
		...supervisor,
		count: counts.get(supervisor.code) ?? 0
	}));
}

export function getWarnings(): string[] {
	const warnings: string[] = [];
	for (const item of examSlots) {
		const seen = new Map<number, RoomCode[]>();
		for (const room of roomCodes) {
			const code = item.supervisors[room];
			seen.set(code, [...(seen.get(code) ?? []), room]);
		}
		for (const [code, rooms] of seen) {
			if (rooms.length > 1) {
				warnings.push(`${item.day}, ${item.date} ${item.start}-${item.end} ${item.subject}: pengawas ${code} (${getSupervisorName(code)}) muncul di ${rooms.join(', ')}.`);
			}
		}
	}
	return warnings;
}

export function toCsv(): string {
	const header = ['tanggal', 'hari', 'jam_mulai', 'jam_selesai', 'mata_pelajaran', 'ruang', 'kode_pengawas', 'nama_pengawas', 'status'];
	const rows = getFlatAssignments().map((row) => [
		row.date,
		row.day,
		row.start,
		row.end,
		row.subject,
		row.room,
		String(row.code),
		row.name,
		row.needsReview ? 'pending_review' : 'ready'
	]);
	return [header, ...rows]
		.map((columns) => columns.map((value) => `"${value.replace(/"/g, '""')}"`).join(','))
		.join('\n');
}
