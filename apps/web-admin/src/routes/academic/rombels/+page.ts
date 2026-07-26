import type { PageLoad } from './$types.js';

type Rombel = {
	id: string;
	code: string;
	name: string;
	level: string;
	is_active: boolean;
	academic_year_id: string;
	academic_year_name: string;
};

export const load: PageLoad = async ({ fetch, url }) => {
	const [semRes, rombelRes] = await Promise.all([
		fetch('/api/academic/semesters/active'),
		fetch('/api/academic/rombels'),
	]);

	let activeSemester: { id: string; label: string; academic_year_id: string } | null = null;
	if (semRes.ok) {
		const sem = await semRes.json();
		activeSemester = sem?.data ?? sem ?? null;
	}

	let items: Rombel[] = [];
	if (rombelRes.ok) {
		const payload = await rombelRes.json();
		items = payload.items ?? [];
	}

	return { items, activeSemester };
};
