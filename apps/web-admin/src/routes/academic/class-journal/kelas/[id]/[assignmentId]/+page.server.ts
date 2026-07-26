import type { PageServerLoad } from './$types.js';

type Assignment = {
	id: string; class_id: string; class_code: string; class_name: string;
	subject_name: string; teacher_name: string;
	teacher_employee_id?: string;
};

export const load: PageServerLoad = async ({ fetch, params, locals }) => {
	const assignmentId = params.assignmentId;

	const res = await fetch('/api/academic/timetable/weekly');
	let assignment: Assignment | null = null;

	if (res.ok) {
		const payload = await res.json();
		const d = payload?.data ?? payload ?? {};
		const assignments: Assignment[] = d.assignments ?? [];

		const found = assignments.find((a: Assignment) => a.id === assignmentId) ?? null;
		if (found) {
			// Verify access: admin can see all, teacher only sees their own
			const user = locals.user;
			const isAdmin = user?.role === 'admin' || user?.roles?.includes('admin');
			if (isAdmin || (user?.employee_id && found.teacher_employee_id === user.employee_id)) {
				assignment = found;
			}
		}
	}

	return { assignment };
};
