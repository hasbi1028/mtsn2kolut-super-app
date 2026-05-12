import { clientApiPath, readClientApiData } from '$lib/client/api';
import type { StudentPortalResultItem } from '$lib/client/student-portal';

export type FetchLike = typeof fetch;

export type ParentPortalPreviewParent = {
	id: string;
	nama: string;
	phone: string;
	linked_student_count?: number;
};

export type ParentPortalPreviewParentsPayload = {
	parents: ParentPortalPreviewParent[];
};

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

export async function fetchParentPortalPreviewParents(fetcher: FetchLike = fetch) {
	const res = await fetcher('/api/portal/preview/parents');
	const payload = await readClientApiData<ParentPortalPreviewParent[] | { parents?: ParentPortalPreviewParent[]; data?: ParentPortalPreviewParent[] }>(res, 'Gagal memuat daftar orang tua untuk preview portal.');
	return { parents: Array.isArray(payload) ? payload : (payload.parents ?? payload.data ?? []) } satisfies ParentPortalPreviewParentsPayload;
}

export async function fetchParentPortalChildren(fetcher: FetchLike = fetch, parentID = '') {
	const url = parentID ? clientApiPath`/api/portal/preview/parents/${parentID}/children` : '/api/portal/orang-tua/children';
	const res = await fetcher(url);
	return readClientApiData<ParentPortalChildrenPayload>(res, 'Gagal memuat daftar anak portal orang tua.');
}

export async function fetchParentPortalChildProfile(studentID: string, fetcher: FetchLike = fetch, parentID = '') {
	const url = parentID
		? clientApiPath`/api/portal/preview/parents/${parentID}/children/${studentID}/profile`
		: clientApiPath`/api/portal/orang-tua/children/${studentID}/profile`;
	const res = await fetcher(url);
	return readClientApiData<ParentPortalChildProfilePayload>(res, 'Gagal memuat profil anak.');
}

export async function fetchParentPortalChildSchedule(studentID: string, fetcher: FetchLike = fetch, parentID = '') {
	const url = parentID
		? clientApiPath`/api/portal/preview/parents/${parentID}/children/${studentID}/schedule`
		: clientApiPath`/api/portal/orang-tua/children/${studentID}/schedule`;
	const res = await fetcher(url);
	return readClientApiData<ParentPortalChildSchedulePayload>(res, 'Gagal memuat jadwal anak.');
}

export async function fetchParentPortalChildResults(studentID: string, fetcher: FetchLike = fetch, parentID = '') {
	const url = parentID
		? clientApiPath`/api/portal/preview/parents/${parentID}/children/${studentID}/results`
		: clientApiPath`/api/portal/orang-tua/children/${studentID}/results`;
	const res = await fetcher(url);
	return readClientApiData<ParentPortalChildResultsPayload>(res, 'Gagal memuat hasil anak.');
}
