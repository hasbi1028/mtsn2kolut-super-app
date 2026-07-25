import type { PageServerLoad } from './$types.js';

type Assignment = {
	id: string; class_id: string; class_code: string; class_name: string;
	subject_name: string; teacher_name: string;
};

export const load: PageServerLoad = async ({ fetch, params }) => {
	const assignmentId = params.assignmentId;

	const res = await fetch('/api/academic/timetable/weekly');
	let assignment: Assignment | null = null;

	if (res.ok) {
		const payload = await res.json();
		const d = payload.data ?? {};
		const assignments: Assignment[] = d.assignments ?? [];
		assignment = assignments.find((a: Assignment) => a.id === assignmentId) ?? null;
	}

	return { assignment };
};
