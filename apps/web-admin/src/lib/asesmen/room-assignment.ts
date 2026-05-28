export type RoomAssignmentMixPolicy = 'same_class' | 'same_grade' | 'mixed_scope';
export type RoomAssignmentMode = 'random_balanced' | 'manual';

export type RoomAssignmentSeat = {
	participant_id: string;
	participant_name?: string;
	participant_nis?: string;
	participant_class?: string;
	room_id: string;
	room_name?: string;
	seat_no: number;
};

export type RoomAssignmentPayload = {
	mix_policy: RoomAssignmentMixPolicy;
	assignment_mode: RoomAssignmentMode;
	allow_cross_grade: boolean;
	is_special_event: boolean;
	assignments?: RoomAssignmentSeat[];
};

export type RoomAssignmentSessionRow = {
	id: string;
	title?: string;
	status?: string;
	session_status?: string;
	scheduled_start?: string;
	package_title?: string;
	unassigned_participant_count?: number;
};

export type RoomAssignmentPreview = {
	summary: {
		participant_count: number;
		room_count: number;
		capacity_total: number;
		assigned_count: number;
		unassigned_count: number;
		mix_policy: string;
		assignment_mode: string;
		allow_cross_grade?: boolean;
		is_special_event?: boolean;
	};
	rooms: Array<{ room_id: string; room_name: string; capacity: number; participant_count: number; levels: Record<string, number> }>;
	assignments?: RoomAssignmentSeat[];
	warnings?: string[];
};

export const roomAssignmentOptions: Array<{ value: RoomAssignmentMixPolicy; title: string; desc: string }> = [
	{ value: 'same_class', title: 'Tetap per Kelas', desc: 'Peserta tetap bersama rombel asal; satu ruang tidak dicampur rombel lain.' },
	{ value: 'same_grade', title: 'Campur per Siswa Satu Tingkat', desc: 'Siswa dari beberapa rombel pada tingkat yang sama dibagi merata per ruang, bukan dipindah sebagai rombel utuh.' },
	{ value: 'mixed_scope', title: 'Campur per Siswa Lintas Tingkat', desc: 'Siswa lintas tingkat/rombel dibagi merata per ruang. Gunakan hanya bila panitia memang memutuskan lintas tingkat.' }
];

export const roomAssignmentModeOptions: Array<{ value: RoomAssignmentMode; title: string; desc: string }> = [
	{ value: 'random_balanced', title: 'Otomatis seimbang', desc: 'Sistem menyusun pembagian ruang secara otomatis lalu menulis nomor kursi per ruang.' },
	{ value: 'manual', title: 'Manual per peserta', desc: 'Panitia boleh mengubah ruang/kursi per peserta sebelum menyimpan.' }
];

export function buildRoomAssignmentPayload(
	mixPolicy: RoomAssignmentMixPolicy,
	assignmentMode: RoomAssignmentMode = 'random_balanced',
	assignments: RoomAssignmentSeat[] = []
): RoomAssignmentPayload {
	return {
		mix_policy: mixPolicy,
		assignment_mode: assignmentMode,
		allow_cross_grade: mixPolicy === 'mixed_scope',
		is_special_event: mixPolicy === 'mixed_scope',
		assignments: assignmentMode === 'manual' && assignments.length > 0 ? assignments : undefined
	};
}

export function roomAssignmentModeLabel(value: string) {
	if (value === 'manual') return 'Manual per peserta';
	if (value === 'same_class') return 'Per Kelas';
	if (value === 'same_grade') return 'Campur per Siswa Satu Tingkat';
	return 'Campur per Siswa Lintas Tingkat';
}

export function roomAssignmentModeHint(value: RoomAssignmentMode) {
	return roomAssignmentModeOptions.find((option) => option.value === value)?.desc ?? '';
}

export function canSaveRoomAssignment(preview: RoomAssignmentPreview | null) {
	return Boolean(preview && (preview.summary.unassigned_count ?? 0) === 0);
}

export function cloneRoomAssignments(assignments: RoomAssignmentSeat[]) {
	return assignments.map((assignment) => ({ ...assignment }));
}

export function sortRoomAssignments(assignments: RoomAssignmentSeat[]) {
	return [...assignments].sort((a, b) => {
		const room = (a.room_name ?? a.room_id).localeCompare(b.room_name ?? b.room_id, 'id-ID');
		if (room !== 0) return room;
		const seat = a.seat_no - b.seat_no;
		if (seat !== 0) return seat;
		return (a.participant_name ?? a.participant_nis ?? a.participant_id).localeCompare(b.participant_name ?? b.participant_nis ?? b.participant_id, 'id-ID');
	});
}
