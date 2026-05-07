import { describe, expect, it } from 'vitest';
import foundationMigration from '../../../../../services/core-api/db/migrations/081_student_parent_account_portal.sql?raw';
import usersSql from '../../../../../services/core-api/db/queries/users.sql?raw';

const studentPermissions = [
	'student_portal.read',
	'student_portal.profile_read',
	'student_portal.schedule_read',
	'student_portal.grades_read',
	'student_portal.assessment_take',
	'student_portal.profile_change_request',
];

const parentPermissions = [
	'parent_portal.read',
	'parent_portal.children_read',
	'parent_portal.child_profile_read',
	'parent_portal.child_schedule_read',
	'parent_portal.child_attendance_read',
	'parent_portal.child_grades_read',
	'parent_portal.profile_change_request',
];

const accountManagementPermissions = ['student_accounts.manage', 'parent_accounts.manage'];

describe('student and parent account portal foundation', () => {
	it('seeds portal/account permissions and default role grants additively', () => {
		const requiredPhrases = [
			'ADD COLUMN IF NOT EXISTS must_change_password',
			'ADD COLUMN IF NOT EXISTS password_changed_at',
			"'siswa'",
			"'ortu'",
			"'admin'",
			'ON CONFLICT',
			...studentPermissions,
			...parentPermissions,
			...accountManagementPermissions,
		];

		for (const phrase of requiredPhrases) {
			expect(foundationMigration).toContain(phrase);
		}
	});

	it('exposes password-change metadata through user queries for auth and account flows', () => {
		expect(usersSql).toContain('u.must_change_password');
		expect(usersSql).toContain('u.password_changed_at');
		expect(usersSql).toContain('must_change_password');
	});
});
