import type { PageLoad } from './$types.js';

export const load: PageLoad = async ({ fetch, params }) => {
	const res = await fetch(`/api/academic/timetable/weekly`);
	let assignment: any = null;

	if (res.ok) {
		const payload = await res.json();
		const d = payload?.data ?? payload ?? {};
		assignment = (d.assignments ?? []).find((a: any) => a.id === params.assignmentId) ?? null;
	}

	return { assignment };
};
