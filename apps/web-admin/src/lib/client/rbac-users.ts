import { readClientApiData, readClientJson } from '$lib/client/api';

export type RBACRole = {
	code: string;
	name?: string;
	description?: string;
	is_system?: boolean;
	is_active?: boolean;
};

export type RBACPermission = {
	code: string;
	name?: string;
	description?: string;
	module?: string;
	action?: string;
	is_active?: boolean;
};

export type RBACRolePermission = {
	role_code: string;
	permission_code: string;
};

export type RBACMatrix = {
	roles: RBACRole[];
	permissions: RBACPermission[];
	role_permissions: RBACRolePermission[] | Record<string, string[]>;
};

export type ProfileLinkPayloadInput = {
	employee_id?: string | null;
	student_id?: string | null;
	parent_id?: string | null;
};

export type RBACRoleInput = {
	code?: string;
	name?: string;
	description?: string;
};

export type RBACPermissionInput = {
	code?: string;
	module?: string;
	action?: string;
	description?: string;
};


export type EmployeeAccountGenerationItem = {
	employee_id: string;
	nip: string;
	nama: string;
	tanggal_lahir?: string;
	nomor_urut: number;
	username: string;
	password?: string;
	role: string;
	status: 'ready' | 'created' | 'skipped' | 'failed';
	message: string;
	existing_user_id?: string;
	username_user_id?: string;
};

export type EmployeeAccountGenerationResult = {
	npsn: string;
	default_role: string;
	password_same_as_username: boolean;
	total: number;
	ready: number;
	created: number;
	skipped: number;
	failed: number;
	items: EmployeeAccountGenerationItem[];
};

export type UserProfileCandidateChild = {
	id: string;
	nama: string;
	class_name?: string;
};

export type UserProfileCandidate = {
	id: string;
	profile_type: 'employee' | 'student' | 'parent' | string;
	nama: string;
	identifier?: string;
	class_id?: string;
	class_name?: string;
	linked_user_id?: string;
	linked_username?: string;
	is_linked?: boolean;
	children?: UserProfileCandidateChild[];
};

export type UserProfileCandidatesResponse = {
	role: string;
	profile: 'employee' | 'student' | 'parent' | string;
	class_id?: string;
	candidates: UserProfileCandidate[];
};

export type UserProfileCandidateParams = {
	role: string;
	class_id?: string | null;
	q?: string | null;
	include_linked?: boolean;
	limit?: number;
};

export type FetchLike = typeof fetch;

function cleanString(value: string | null | undefined) {
	const next = typeof value === 'string' ? value.trim() : '';
	return next || null;
}

function uniqueNonEmpty(values: string[]) {
	return Array.from(new Set(values.map((value) => value.trim()).filter(Boolean)));
}

export function buildUserRolePayload(roles: string[]) {
	return { roles: uniqueNonEmpty(roles) };
}

export function buildRolePermissionPayload(permissions: string[]) {
	return { permissions: uniqueNonEmpty(permissions) };
}

export function buildProfileLinkPayload(input: ProfileLinkPayloadInput) {
	return {
		employee_id: cleanString(input.employee_id),
		student_id: cleanString(input.student_id),
		parent_id: cleanString(input.parent_id)
	};
}

async function writeRBAC<T>(url: string, method: string, payload: unknown, fetcher: FetchLike) {
	const res = await fetcher(url, {
		method,
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(payload)
	});
	return readClientApiData<T>(res, 'Gagal menyimpan data RBAC.');
}

export async function fetchRBACMatrix(fetcher: FetchLike = fetch) {
	const res = await fetcher('/api/rbac/matrix');
	return (
		(await readClientApiData<RBACMatrix>(res, 'Gagal memuat matrix RBAC.')) ?? {
			roles: [],
			permissions: [],
			role_permissions: {}
		}
	);
}

export async function fetchUserProfileCandidates(params: UserProfileCandidateParams, fetcher: FetchLike = fetch) {
	const query = new URLSearchParams();
	query.set('role', params.role.trim());
	const classID = cleanString(params.class_id);
	if (classID) query.set('class_id', classID);
	const q = cleanString(params.q);
	if (q) query.set('q', q);
	if (typeof params.include_linked === 'boolean') query.set('include_linked', String(params.include_linked));
	if (typeof params.limit === 'number' && Number.isFinite(params.limit)) query.set('limit', String(Math.trunc(params.limit)));
	const res = await fetcher(`/api/users/profile-candidates?${query.toString()}`);
	return readClientApiData<UserProfileCandidatesResponse>(res, 'Gagal memuat kandidat profil pengguna.');
}

export async function createRBACRole(input: RBACRoleInput, fetcher: FetchLike = fetch) {
	return writeRBAC<RBACRole>('/api/rbac/roles', 'POST', input, fetcher);
}

export async function updateRBACRole(code: string, input: RBACRoleInput, fetcher: FetchLike = fetch) {
	return writeRBAC<RBACRole>(`/api/rbac/roles/${code}`, 'PUT', input, fetcher);
}

export async function setRBACRoleActive(code: string, isActive: boolean, fetcher: FetchLike = fetch) {
	return writeRBAC<unknown>(`/api/rbac/roles/${code}/status`, 'PATCH', { is_active: isActive }, fetcher);
}

export async function updateRBACRolePermissions(code: string, permissions: string[], fetcher: FetchLike = fetch) {
	return writeRBAC<unknown>(`/api/rbac/roles/${code}/permissions`, 'PUT', buildRolePermissionPayload(permissions), fetcher);
}

export async function createRBACPermission(input: RBACPermissionInput, fetcher: FetchLike = fetch) {
	return writeRBAC<RBACPermission>('/api/rbac/permissions', 'POST', input, fetcher);
}

export async function updateRBACPermission(code: string, input: RBACPermissionInput, fetcher: FetchLike = fetch) {
	return writeRBAC<RBACPermission>(`/api/rbac/permissions/${code}`, 'PUT', input, fetcher);
}

export async function setRBACPermissionActive(code: string, isActive: boolean, fetcher: FetchLike = fetch) {
	return writeRBAC<unknown>(`/api/rbac/permissions/${code}/status`, 'PATCH', { is_active: isActive }, fetcher);
}

export async function updateUserRoles(userID: string, roles: string[], fetcher: FetchLike = fetch) {
	const res = await fetcher(`/api/users/${userID}/roles`, {
		method: 'PATCH',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(buildUserRolePayload(roles))
	});
	await readClientJson<unknown>(res);
}

export async function resetUserPassword(userID: string, password: string, fetcher: FetchLike = fetch) {
	const res = await fetcher(`/api/users/${userID}/reset-password`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ password })
	});
	await readClientJson<unknown>(res);
}

export async function updateUserProfileLink(
	userID: string,
	payload: ProfileLinkPayloadInput,
	fetcher: FetchLike = fetch
) {
	const res = await fetcher(`/api/users/${userID}/profile-link`, {
		method: 'PATCH',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(buildProfileLinkPayload(payload))
	});
	await readClientJson<unknown>(res);
}

export async function previewEmployeeAccountGeneration(fetcher: FetchLike = fetch) {
	const res = await fetcher('/api/users/generate-from-employees/preview');
	return readClientApiData<EmployeeAccountGenerationResult>(res, 'Gagal memuat preview generate akun pegawai.');
}

export async function generateEmployeeAccounts(fetcher: FetchLike = fetch) {
	const res = await fetcher('/api/users/generate-from-employees', { method: 'POST' });
	return readClientApiData<EmployeeAccountGenerationResult>(res, 'Gagal generate akun pegawai.');
}

export function employeeAccountGenerationCSV(result: EmployeeAccountGenerationResult) {
	const headers = ['nama', 'nip', 'tanggal_lahir', 'username', 'password_awal', 'role', 'status', 'keterangan'];
	const escape = (value: unknown) => `"${String(value ?? '').replaceAll('"', '""')}"`;
	const rows = result.items.map((item) => [
		item.nama,
		item.nip,
		item.tanggal_lahir ?? '',
		item.username,
		item.password ?? '',
		item.role,
		item.status,
		item.message
	]);
	return [headers, ...rows].map((row) => row.map(escape).join(',')).join('\n');
}
