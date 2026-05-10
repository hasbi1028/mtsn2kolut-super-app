import { clientApiPath, readClientApiData } from '$lib/client/api';

export type FetchLike = typeof fetch;

export type StudentPortalProfile = {
	id: string;
	nis: string;
	nisn: string;
	nama: string;
	gender: string;
	parent_name: string;
	parent_phone: string;
	class_id: string;
	class_name: string;
	class_code: string;
	linked_parent_names: string;
	linked_parent_count: number;
	is_active: boolean;
	status: string;
};

export type StudentPortalScheduleItem = {
	id: string;
	day_of_week: number;
	start_time: string;
	end_time: string;
	room_label: string;
	notes: string;
	class_name: string;
	class_code: string;
	subject_name: string;
	subject_code: string;
	teacher_name: string;
};

export type StudentPortalResultItem = {
	participant_id: string;
	session_id: string;
	room_id: string;
	seat_no: unknown;
	joined_at: unknown;
	submitted_at: unknown;
	score: unknown;
	session_title: string;
	session_status: string;
	scheduled_start: unknown;
	scheduled_end: unknown;
	package_title: string;
	duration_minutes: number;
	room_name: string;
};

export type StudentPortalCbtScheduleStatus = 'upcoming' | 'token_window' | 'active' | 'submitted' | 'closed' | 'locked';

export type StudentPortalCbtScheduleItem = {
	participant_id: string;
	session_id: string;
	session_title: string;
	package_title: string;
	scheduled_start: string;
	scheduled_end: string;
	duration_minutes: number;
	room_id: string | null;
	room_name: string | null;
	seat_no: number | null;
	status: StudentPortalCbtScheduleStatus;
	can_reveal_token: boolean;
	requires_room_token: boolean;
	token_masked: string | null;
};

export type StudentPortalProfilePayload = {
	student: StudentPortalProfile;
};

export type StudentPortalSchedulePayload = {
	schedule: StudentPortalScheduleItem[];
};

export type StudentPortalResultsPayload = {
	results: StudentPortalResultItem[];
};

export type StudentPortalCbtSchedulePayload = {
	schedule: StudentPortalCbtScheduleItem[];
};

export type StudentPortalCbtRevealPayload = {
	token: string;
	expires_at: string;
};

export async function fetchStudentPortalProfile(fetcher: FetchLike = fetch) {
	const res = await fetcher('/api/portal/siswa/profile');
	return readClientApiData<StudentPortalProfilePayload>(res, 'Gagal memuat profil portal siswa.');
}

export async function fetchStudentPortalSchedule(fetcher: FetchLike = fetch) {
	const res = await fetcher('/api/portal/siswa/schedule');
	return readClientApiData<StudentPortalSchedulePayload>(res, 'Gagal memuat jadwal portal siswa.');
}

export async function fetchStudentPortalResults(fetcher: FetchLike = fetch) {
	const res = await fetcher('/api/portal/siswa/results');
	return readClientApiData<StudentPortalResultsPayload>(res, 'Gagal memuat hasil portal siswa.');
}

export async function fetchStudentPortalCbtSchedule(fetcher: FetchLike = fetch) {
	const res = await fetcher('/api/portal/siswa/cbt');
	return readClientApiData<StudentPortalCbtSchedulePayload>(res, 'Gagal memuat jadwal CBT portal siswa.');
}

export async function revealStudentPortalCbtToken(
	participantID: string,
	roomToken: string,
	fetcher: FetchLike = fetch
) {
	const res = await fetcher(clientApiPath`/api/portal/siswa/cbt/${participantID}/reveal-token`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ room_token: roomToken })
	});
	return readClientApiData<StudentPortalCbtRevealPayload>(res, 'Token ujian belum dapat dibuka.');
}
