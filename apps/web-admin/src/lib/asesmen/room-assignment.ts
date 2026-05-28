export type RoomAssignmentMixPolicy = 'same_class' | 'same_grade' | 'mixed_scope';

export type RoomAssignmentPayload = {
	mix_policy: RoomAssignmentMixPolicy;
	assignment_mode: 'random_balanced';
	allow_cross_grade: boolean;
	is_special_event: boolean;
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
	};
	rooms: Array<{ room_id: string; room_name: string; capacity: number; participant_count: number; levels: Record<string, number> }>;
	warnings?: string[];
};

export const roomAssignmentOptions: Array<{ value: RoomAssignmentMixPolicy; title: string; desc: string }> = [
	{ value: 'same_class', title: 'Tetap per Kelas', desc: 'Peserta tetap mengikuti rombel asal.' },
	{ value: 'same_grade', title: 'Campur Satu Tingkat', desc: 'Rombel boleh bercampur, tetapi tingkat tetap dipisah.' },
	{ value: 'mixed_scope', title: 'Campur Lintas Tingkat', desc: 'Gunakan hanya bila panitia memang memutuskan lintas tingkat.' }
];

export function buildRoomAssignmentPayload(mixPolicy: RoomAssignmentMixPolicy): RoomAssignmentPayload {
	return {
		mix_policy: mixPolicy,
		assignment_mode: 'random_balanced',
		allow_cross_grade: mixPolicy === 'mixed_scope',
		is_special_event: mixPolicy === 'mixed_scope'
	};
}

export function roomAssignmentModeLabel(value: string) {
	if (value === 'same_class') return 'Per Kelas';
	if (value === 'same_grade') return 'Campur Satu Tingkat';
	return 'Campur Lintas Tingkat';
}

export function canSaveRoomAssignment(preview: RoomAssignmentPreview | null) {
	return Boolean(preview && (preview.summary.unassigned_count ?? 0) === 0);
}
