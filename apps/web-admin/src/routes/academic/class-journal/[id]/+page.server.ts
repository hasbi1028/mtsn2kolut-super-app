import type { PageServerLoad } from './$types.js';

type Assignment = {
	id: string;
	class_id: string;
	class_code: string;
	class_name: string;
	subject_name: string;
	teacher_name: string;
};

type Class = { id: string; code: string; name: string; level: string; };

export const load: PageServerLoad = async ({ fetch, params }) => {
	const id = params.id;

	// Load classes + assignments
	const res = await fetch('/api/academic/timetable/weekly');

	let assignment: Assignment | null = null;

	if (res.ok) {
		const payload = await res.json();
		const d = payload.data ?? {};
		const classes: Class[] = d.classes ?? [];
		const assignments: Assignment[] = d.assignments ?? [];
		assignment = assignments.find((a: Assignment) => a.id === id) ?? null;
	}

	return { assignment };
};
