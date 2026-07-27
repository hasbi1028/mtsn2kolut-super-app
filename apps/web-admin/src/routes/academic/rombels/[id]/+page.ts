import type { PageLoad } from './$types.js';

export const load: PageLoad = async ({ fetch, params }) => {
	const id = params.id;
	if (!id) return { detail: null, students: [] };

	const [detailRes, studentsRes] = await Promise.all([
		fetch(`/api/academic/rombels/${id}`),
		fetch(`/api/academic/rombels/${id}/students`),
	]);

	let detail: any = null;
	if (detailRes.ok) {
		const p = await detailRes.json();
		detail = p?.data ?? p ?? null;
	}

	let students: any[] = [];
	if (studentsRes.ok) {
		const p = await studentsRes.json();
		students = p?.items ?? [];
	}

	return { detail, students };
};
