export type UIAccessSubject = {
	role?: string;
	roles?: readonly string[];
	permissions?: readonly string[];
};

export type UIPermissionPolicy = {
	permissions: readonly string[];
	roleFallbacks?: readonly string[];
	allowAuthenticatedFallback?: boolean;
	adminBypass?: boolean;
};

export type UIPolicyAccessReason =
	| 'admin'
	| 'permission'
	| 'role_fallback'
	| 'authenticated_fallback'
	| 'missing_permission';

export type UIPolicyAccessEvaluation = {
	allowed: boolean;
	reason: UIPolicyAccessReason;
	requiredPermissions: string[];
	matchedPermissions: string[];
	roleFallbacks: string[];
	matchedRoles: string[];
};

export function normalizeAccessValues(values: readonly string[] | undefined) {
	return Array.from(new Set((values ?? []).map((value) => value.trim()).filter(Boolean)));
}

export function rolesForAccess(subject: UIAccessSubject | undefined) {
	if (!subject) return [];
	const roles = normalizeAccessValues(subject.roles);
	return roles.length > 0 ? roles : normalizeAccessValues(subject.role ? [subject.role] : []);
}

export function permissionsForAccess(subject: UIAccessSubject | undefined) {
	return normalizeAccessValues(subject?.permissions);
}

export function hasAdminRole(subject: UIAccessSubject | undefined) {
	return rolesForAccess(subject).includes('admin');
}

export function evaluateUIPolicyAccess(
	policy: UIPermissionPolicy,
	subject: UIAccessSubject | undefined
): UIPolicyAccessEvaluation {
	const requiredPermissions = normalizeAccessValues(policy.permissions);
	const roleFallbacks = normalizeAccessValues(policy.roleFallbacks);
	const roles = rolesForAccess(subject);
	const permissions = permissionsForAccess(subject);
	const matchedPermissions = requiredPermissions.filter((permission) => permissions.includes(permission));
	const matchedRoles = roleFallbacks.filter((role) => roles.includes(role));
	const adminBypass = policy.adminBypass ?? true;

	if (adminBypass && roles.includes('admin')) {
		return { allowed: true, reason: 'admin', requiredPermissions, matchedPermissions: [], roleFallbacks, matchedRoles: ['admin'] };
	}
	if (matchedPermissions.length > 0) {
		return { allowed: true, reason: 'permission', requiredPermissions, matchedPermissions, roleFallbacks, matchedRoles: [] };
	}
	if (policy.allowAuthenticatedFallback && subject) {
		return { allowed: true, reason: 'authenticated_fallback', requiredPermissions, matchedPermissions: [], roleFallbacks, matchedRoles: [] };
	}
	if (matchedRoles.length > 0) {
		return { allowed: true, reason: 'role_fallback', requiredPermissions, matchedPermissions: [], roleFallbacks, matchedRoles };
	}
	return { allowed: false, reason: 'missing_permission', requiredPermissions, matchedPermissions: [], roleFallbacks, matchedRoles: [] };
}

export function hasAnyPermission(subject: UIAccessSubject | undefined, required: readonly string[]) {
	return evaluateUIPolicyAccess({ permissions: required, adminBypass: false }, subject).allowed;
}
