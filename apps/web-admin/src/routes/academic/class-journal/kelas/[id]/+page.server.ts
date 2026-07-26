import type { PageServerLoad } from './$types.js';

type Assignment = {
	id: string; class_id: string; class_code: string; class_name: string;
	subject_name: string; teacher_name: string;
	teacher_employee_id?: string;
};

export const load: PageServerLoad = async ({ fetch, params, locals }) => {
	const id = params.id;

	const res = await fetch('/api/academic/timetable/weekly');
	let kelas: { id: string; code: string; name: string; level: string } | null = null;
	let mapels: Assignment[] = [];

	if (res.ok) {
		const payload = await res.json();
		const d = payload?.data ?? payload ?? {};
		const classes: any[] = d.classes ?? [];
		const assignments: Assignment[] = d.assignments ?? [];

		kelas = classes.find((c: any) => c.id === id) ?? null;

		// Filter by class
		mapels = assignments.filter((a: Assignment) => a.class_id === id);

		// Filter for non-admin: only show this teacher's subjects
		const user = locals.user;
		const isAdmin = user?.role === 'admin' || user?.roles?.includes('admin');
		if (!isAdmin && user?.employee_id) {
			mapels = mapels.filter(a => a.teacher_employee_id === user.employee_id);
		}
	}

	return { kelas, mapels };
};
