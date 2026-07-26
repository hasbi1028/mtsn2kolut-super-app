import type { PageLoad } from './$types.js';
import { redirect } from '@sveltejs/kit';

type Semester = {
	id: string;
	academic_year_id: string;
	name: string;
	academic_year_name: string;
	semester_type: string;
	is_active: boolean;
	is_current: boolean;
};

export const load: PageLoad = async ({ fetch, url, parent }) => {
	const { user } = await parent();
	const isAdmin = user?.role === 'admin' || user?.roles?.includes('admin');
	if (!isAdmin) throw redirect(302, '/');

	const res = await fetch(`${url.origin}/api/academic/semesters`);
	if (!res.ok) {
		const items: Semester[] = [];
		return { items };
	}
	const payload = await res.json();
	const data = payload?.data ?? payload ?? {};
	const items: Semester[] = data.items ?? data ?? [];
	return { items };
};
