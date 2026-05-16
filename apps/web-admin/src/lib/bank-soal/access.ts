export type BankSoalAccessUser = {
	role?: string;
	roles?: string[];
	permissions?: string[];
};

function rolesOf(user?: BankSoalAccessUser): string[] {
	if (!user) return [];
	const roles = Array.isArray(user.roles) ? user.roles : [];
	return roles.length > 0 ? roles : user.role ? [user.role] : [];
}

function permissionsOf(user?: BankSoalAccessUser): string[] {
	return (user?.permissions ?? []).map((permission) => permission.trim()).filter(Boolean);
}

function hasRole(user: BankSoalAccessUser | undefined, role: string): boolean {
	return rolesOf(user).includes(role);
}

function hasPermission(user: BankSoalAccessUser | undefined, permission: string): boolean {
	return permissionsOf(user).includes(permission);
}

export function canCreateBankSoal(user?: BankSoalAccessUser): boolean {
	return hasRole(user, 'admin') || hasPermission(user, 'bank_soal.create');
}

export function canReviewBankSoal(user?: BankSoalAccessUser): boolean {
	return hasRole(user, 'admin') || hasPermission(user, 'bank_soal.review') || hasPermission(user, 'bank_soal.publish');
}

export function canImportBankSoal(user?: BankSoalAccessUser): boolean {
	return hasRole(user, 'admin') || hasPermission(user, 'bank_soal.import');
}

export function canManageBankSoalSettings(user?: BankSoalAccessUser): boolean {
	return hasRole(user, 'admin') || hasPermission(user, 'bank_soal.settings');
}

export function canAssignBankSoalReviewer(user?: BankSoalAccessUser): boolean {
	return hasRole(user, 'admin') || hasPermission(user, 'bank_soal.assign_reviewer') || hasPermission(user, 'bank_soal.settings');
}

export function canPublishBankSoal(user?: BankSoalAccessUser): boolean {
	return hasRole(user, 'admin') || hasPermission(user, 'bank_soal.publish') || hasPermission(user, 'bank_soal.approve');
}

export function canDeleteBankSoal(user?: BankSoalAccessUser): boolean {
	return hasRole(user, 'admin') || hasPermission(user, 'bank_soal.delete');
}
