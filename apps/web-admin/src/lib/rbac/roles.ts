import type { RBACRole, RBACRoleInput } from '$lib/client/rbac-users';

export type RoleMetadataDraft = Required<Pick<RBACRoleInput, 'code' | 'name' | 'description'>>;

export function normalizeRoleCode(value: string | null | undefined) {
	return (value ?? '')
		.trim()
		.toLowerCase()
		.replace(/[^a-z0-9]+/g, '_')
		.replace(/^_+|_+$/g, '')
		.replace(/_+/g, '_');
}

export function buildRoleMetadataDraft(role: RBACRole | null | undefined): RoleMetadataDraft {
	return {
		code: role?.code ?? '',
		name: role?.name ?? '',
		description: role?.description ?? ''
	};
}

export function sanitizeRoleMetadataPayload(input: RBACRoleInput): RoleMetadataDraft {
	return {
		code: normalizeRoleCode(input.code),
		name: (input.name ?? '').trim(),
		description: (input.description ?? '').trim()
	};
}

export function canEditRoleMetadata(role: RBACRole | null | undefined) {
	return Boolean(role && !role.is_system);
}

export function roleMetadataChanged(role: RBACRole | null | undefined, draft: RBACRoleInput) {
	if (!role) return false;
	const normalized = sanitizeRoleMetadataPayload({ ...draft, code: role.code });
	return (role.name ?? '') !== normalized.name || (role.description ?? '') !== (normalized.description ?? '');
}
