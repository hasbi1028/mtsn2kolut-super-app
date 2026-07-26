import type { PageServerLoad } from './$types.js';

export const load: PageServerLoad = async ({ fetch, locals }) => {
	const res = await fetch('/api/academic/timetable/weekly');
	let classes: any[] = [];

	if (res.ok) {
		const payload = await res.json();
		const d = payload?.data ?? payload ?? {};
		classes = d.classes ?? [];
		const assignments: any[] = d.assignments ?? [];

		// Filter for non-admin: only show classes where the teacher has assignments
		const user = locals.user;
		const isAdmin = user?.role === 'admin' || user?.roles?.includes('admin');
		if (!isAdmin && user?.employee_id) {
			const teacherClassIds = new Set(
				assignments
					.filter((a: any) => a.teacher_employee_id === user.employee_id)
					.map((a: any) => a.class_id)
			);
			classes = classes.filter((c: any) => teacherClassIds.has(c.id));
		}
	}
	return { classes };
};
