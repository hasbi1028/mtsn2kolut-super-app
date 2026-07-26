import type { PageLoad } from './$types.js';

export const load: PageLoad = async ({ fetch, params }) => {
	const res = await fetch(`/api/academic/timetable/weekly`);
	let kelas: Record<string, unknown> | null = null;
	let mapels: any[] = [];

	if (res.ok) {
		const payload = await res.json();
		const d = payload?.data ?? payload ?? {};
		kelas = (d.classes ?? []).find((c: any) => c.id === params.id) ?? null;
		mapels = (d.assignments ?? []).filter((a: any) => a.class_id === params.id);
	}

	return { kelas, mapels };
};
