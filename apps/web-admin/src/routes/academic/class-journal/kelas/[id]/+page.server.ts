import type { PageServerLoad } from './$types.js';

type Assignment = {
	id: string; class_id: string; class_code: string; class_name: string;
	subject_name: string; teacher_name: string;
};

export const load: PageServerLoad = async ({ fetch, params }) => {
	const id = params.id;

	const res = await fetch('/api/academic/timetable/weekly');
	let kelas: { id: string; code: string; name: string; level: string } | null = null;
	let mapels: Assignment[] = [];

	if (res.ok) {
		const payload = await res.json();
		// BFF proxy already unwraps {data: {...}} → just {...}
		const d = payload?.data ?? payload ?? {};
		const classes: any[] = d.classes ?? [];
		const assignments: Assignment[] = d.assignments ?? [];
		kelas = classes.find((c: any) => c.id === id) ?? null;
		mapels = assignments.filter((a: Assignment) => a.class_id === id);
	}

	return { kelas, mapels };
};
