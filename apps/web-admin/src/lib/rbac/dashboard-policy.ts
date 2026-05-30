import {
	evaluateUIPolicyAccess,
	hasAdminRole,
	permissionsForAccess,
	rolesForAccess,
	type UIAccessSubject
} from './access-policy';

export type DashboardWidgetDefinition = {
	id: string;
	label: string;
	description: string;
	permissions: readonly string[];
	roleFallbacks?: readonly string[];
	audienceRoles?: readonly string[];
	adminDefault?: boolean;
};

export type DashboardDataAccess = {
	academicStats: boolean;
	assessmentSessions: boolean;
	bankSoal: boolean;
	parentPortal: boolean;
	studentPortal: boolean;
	studentSummary: boolean;
	teacherTimetable: boolean;
	ungradedEssays: boolean;
};

export const DASHBOARD_WIDGETS = [
	{
		id: 'academic-overview',
		label: 'Ringkasan Akademik',
		description: 'Statistik siswa, kelas, mapel, dan tahun ajaran.',
		permissions: ['academic.read'],
		adminDefault: true
	},
	{
		id: 'bank-soal-overview',
		label: 'Bank Soal',
		description: 'Akses cepat daftar soal dan pemetaan Bank Soal.',
		permissions: ['bank_soal.read'],
		adminDefault: true
	},
	{
		id: 'bank-soal-authoring',
		label: 'Authoring Bank Soal',
		description: 'Shortcut tambah soal untuk penyusun soal.',
		permissions: ['bank_soal.create'],
		adminDefault: true
	},
	{
		id: 'teacher-students',
		label: 'Siswa Diampu',
		description: 'Ringkasan siswa yang dapat dipantau guru.',
		permissions: ['students.read'],
		audienceRoles: ['guru']
	},
	{
		id: 'teacher-timetable',
		label: 'Jadwal Mengajar',
		description: 'Slot kelas-mapel guru pada minggu berjalan.',
		permissions: ['academic.read', 'journal.read', 'journal.manage', 'journal.read_all', 'journal.manage_all'],
		audienceRoles: ['guru']
	},
	{
		id: 'student-portal',
		label: 'Portal Siswa',
		description: 'Profil, jadwal, dan asesmen siswa.',
		permissions: ['student_portal.read'],
		roleFallbacks: ['siswa']
	},
	{
		id: 'parent-portal',
		label: 'Portal Orang Tua',
		description: 'Data anak dan jadwal yang terhubung ke wali.',
		permissions: ['parent_portal.read'],
		roleFallbacks: ['ortu']
	},
	{
		id: 'staff-operations',
		label: 'Operasional Staf',
		description: 'Shortcut tata usaha, perpustakaan, inventaris, dan dokumen.',
		permissions: [
			'letters.read',
			'archives.read',
			'library.read',
			'inventory.read',
			'document_cycles.read',
			'governance.read'
		]
	}
] as const satisfies readonly DashboardWidgetDefinition[];

function audienceMatches(widget: DashboardWidgetDefinition, user: UIAccessSubject | undefined) {
	if (!widget.audienceRoles || widget.audienceRoles.length === 0) return true;
	const roles = rolesForAccess(user);
	return widget.audienceRoles.some((role) => roles.includes(role));
}

export function dashboardWidgetEvaluation(widget: DashboardWidgetDefinition, user: UIAccessSubject | undefined) {
	if (!audienceMatches(widget, user)) {
		return {
			...evaluateUIPolicyAccess({ permissions: widget.permissions, adminBypass: false }, user),
			allowed: false
		};
	}
	if (widget.adminDefault && hasAdminRole(user)) {
		return evaluateUIPolicyAccess({ permissions: widget.permissions }, user);
	}
	return evaluateUIPolicyAccess({
		permissions: widget.permissions,
		roleFallbacks: widget.roleFallbacks,
		adminBypass: false
	}, user);
}

export function visibleDashboardWidgetsForUser(user: UIAccessSubject | undefined) {
	return DASHBOARD_WIDGETS.filter((widget) => dashboardWidgetEvaluation(widget, user).allowed);
}

function hasPermission(user: UIAccessSubject | undefined, permissions: readonly string[]) {
	const owned = permissionsForAccess(user);
	return permissions.some((permission) => owned.includes(permission));
}

export function dashboardDataAccessForUser(user: UIAccessSubject | undefined): DashboardDataAccess {
	const admin = hasAdminRole(user);
	const roles = rolesForAccess(user);
	const isGuru = roles.includes('guru');
	const isSiswa = roles.includes('siswa');
	const isParent = roles.includes('ortu');
	const canUseStudentPortal = isSiswa && dashboardWidgetEvaluation(DASHBOARD_WIDGETS.find((widget) => widget.id === 'student-portal')!, user).allowed;
	const canUseParentPortal = isParent && dashboardWidgetEvaluation(DASHBOARD_WIDGETS.find((widget) => widget.id === 'parent-portal')!, user).allowed;
	const hasStudents = isGuru && hasPermission(user, ['students.read']);
	const hasTimetable = isGuru && hasPermission(user, ['academic.read', 'journal.read', 'journal.manage', 'journal.read_all', 'journal.manage_all']);
	const hasBankSoal = admin || hasPermission(user, ['bank_soal.read', 'bank_soal.create', 'bank_soal.review', 'bank_soal.import', 'bank_soal.settings', 'bank_soal.analytics']);

	return {
		academicStats: admin || hasPermission(user, ['academic.read']),
		assessmentSessions: false,
		bankSoal: hasBankSoal,
		parentPortal: canUseParentPortal,
		studentPortal: canUseStudentPortal,
		studentSummary: hasStudents,
		teacherTimetable: hasTimetable,
		ungradedEssays: false
	};
}
