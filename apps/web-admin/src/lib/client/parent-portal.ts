import { clientApiPath, readClientApiData } from '$lib/client/api';
import type { StudentPortalResultItem } from '$lib/client/student-portal';

export type FetchLike = typeof fetch;

export type ParentPortalChild = {
	id: string;
	nis: string;
	nama: string;
	class_id: string;
	class_name: string;
	relationship: string;
	is_primary_contact: boolean;
	notes: string;
};

export type ParentPortalChildProfile = {
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
	is_active: boolean;
	status: string;
};

export type ParentPortalChildScheduleItem = {
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

export type ParentPortalChildrenPayload = {
	children: ParentPortalChild[];
};

export type ParentPortalChildProfilePayload = {
	student: ParentPortalChildProfile;
};

export type ParentPortalChildSchedulePayload = {
	schedule: ParentPortalChildScheduleItem[];
};

export type ParentPortalChildResultsPayload = {
	results: StudentPortalResultItem[];
};

export async function fetchParentPortalChildren(fetcher: FetchLike = fetch) {
	const res = await fetcher('/api/portal/orang-tua/children');
	return readClientApiData<ParentPortalChildrenPayload>(res, 'Gagal memuat daftar anak portal orang tua.');
}

export async function fetchParentPortalChildProfile(studentID: string, fetcher: FetchLike = fetch) {
	const res = await fetcher(clientApiPath`/api/portal/orang-tua/children/${studentID}/profile`);
	return readClientApiData<ParentPortalChildProfilePayload>(res, 'Gagal memuat profil anak.');
}

export async function fetchParentPortalChildSchedule(studentID: string, fetcher: FetchLike = fetch) {
	const res = await fetcher(clientApiPath`/api/portal/orang-tua/children/${studentID}/schedule`);
	return readClientApiData<ParentPortalChildSchedulePayload>(res, 'Gagal memuat jadwal anak.');
}

export async function fetchParentPortalChildResults(studentID: string, fetcher: FetchLike = fetch) {
	const res = await fetcher(clientApiPath`/api/portal/orang-tua/children/${studentID}/results`);
	return readClientApiData<ParentPortalChildResultsPayload>(res, 'Gagal memuat hasil anak.');
}
