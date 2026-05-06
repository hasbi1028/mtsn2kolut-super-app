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

export type RBACMatrix = {
	roles: RBACRole[];
	permissions: RBACPermission[];
	role_permissions: Record<string, string[]>;
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

export async function createRBACRole(input: RBACRoleInput, fetcher: FetchLike = fetch) {
	return writeRBAC<RBACRole>('/api/rbac/roles', 'POST', input, fetcher);
}

export async function updateRBACRole(code: string, input: RBACRoleInput, fetcher: FetchLike = fetch) {
	return writeRBAC<RBACRole>(`/api/rbac/roles/${code}`, 'PUT', input, fetcher);
}

export async function setRBACRoleActive(code: string, isActive: boolean, fetcher: FetchLike = fetch) {
	return writeRBAC<unknown>(`/api/rbac/roles/${code}/status`, 'PATCH', { is_active: isActive }, fetcher);
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
