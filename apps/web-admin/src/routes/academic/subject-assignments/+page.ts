import type { PageLoad } from './$types.js';
import { redirect } from '@sveltejs/kit';

export const load: PageLoad = async ({ fetch, parent }) => {
	const { user } = await parent();
	const isAdmin = user?.role === 'admin' || user?.roles?.includes('admin');
	if (!isAdmin) throw redirect(302, '/');

	const res = await fetch('/api/academic/subject-assignments');

	let classes: any[] = [];
	let subjects: any[] = [];
	let teachers: any[] = [];
	let cells: any[] = [];

	if (res.ok) {
		const payload = await res.json();
		classes = payload.classes ?? [];
		subjects = payload.subjects ?? [];
		teachers = payload.teachers ?? [];
		cells = payload.cells ?? [];
	}

	return { classes, subjects, teachers, cells };
};
